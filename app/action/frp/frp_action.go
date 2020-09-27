package frp

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/ttlv/frp_adapter/app/entries"
	"github.com/ttlv/frp_adapter/app/helpers"
	"github.com/ttlv/frp_adapter/model"
	"github.com/ttlv/frp_adapter/nm_action"
	"k8s.io/client-go/dynamic"
	"net/http"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Handlers struct {
	DynamicClient dynamic.Interface
	GVR           schema.GroupVersionResource
}

func NewHandlers(dynamicClient dynamic.Interface, gvr schema.GroupVersionResource) Handlers {
	return Handlers{DynamicClient: dynamicClient, GVR: gvr}
}

// 当有新的frpc节点注册到frps时，创建NM对象
func (handler *Handlers) FrpCreate(w http.ResponseWriter, r *http.Request) {
	var (
		nms []model.FrpServer
	)
	nms = append(nms, model.FrpServer{
		PublicIpAddress: r.FormValue("frp_server_ip_address"),
		Status:          model.FrpOnline,
		UniqueID:        r.FormValue("unique_id"),
		Port:            r.FormValue("port"),
	})
	if err := nm_action.NmCreate(handler.DynamicClient, handler.GVR, nms); err != nil {
		helpers.RenderFailureJSON(w, http.StatusBadRequest, fmt.Sprintf("can't create nodemaintenances crd resource in k8s cluster,err is:%v", err))
		return
	}
	helpers.RenderSuccessJSON(w, 200, fmt.Sprintf("create nodemaintenances-%v crd resource in k8s cluster successfully", r.FormValue("unique_id")))
	return
}

// 当frpc的状态更新时需要立即更新nodemaintenances资源
func (handler *Handlers) FrpUpdate(w http.ResponseWriter, r *http.Request) {
	// 更新前判断nm资源是否存在，避免frpc已经接入frps但是没有nm对象的情况，如果不存在应该先创建
	frpServers := []model.FrpServer{}
	frpServers = append(frpServers, model.FrpServer{
		PublicIpAddress: r.FormValue("frp_server_ip_address"),
		Status:          r.FormValue("status"),
		UniqueID:        r.FormValue("unique_id"),
		Port:            r.FormValue("port"),
	})
	err := nm_action.NmCreate(handler.DynamicClient, handler.GVR, frpServers)
	if err != nil {
		helpers.RenderFailureJSON(w, 400, fmt.Sprintf("created failed: %v", err))
		return
	}
	err = nm_action.NMNormalUpdate(handler.DynamicClient, handler.GVR, frpServers)
	if err != nil {
		helpers.RenderFailureJSON(w, 400, fmt.Sprintf("update failed: %v", err))
		return
	}
	helpers.RenderSuccessJSON(w, 200, "Update Successfully")
	return
}

// Frps请求Frp Adapter获取nodemaintenances资源数据
func (handler *Handlers) FrpFetch(w http.ResponseWriter, r *http.Request) {
	var (
		nodeMaintenanceName = chi.URLParam(r, "name")
		coreFrp             = entries.CoreFrp{}
	)
	if nodeMaintenanceName == "" {
		helpers.RenderFailureJSON(w, 400, "nodemaintenances name为空")
		return
	}
	result, getErr := handler.DynamicClient.Resource(handler.GVR).Get(nodeMaintenanceName, metav1.GetOptions{})
	if getErr != nil {
		helpers.RenderFailureJSON(w, 401, fmt.Sprintf("failed to get latest version of nodeMaintenance: %v", getErr))
		return
	}
	specServices, found, err := unstructured.NestedSlice(result.Object, "spec", "services")
	if err != nil || !found || specServices == nil {
		helpers.RenderFailureJSON(w, 400, fmt.Sprintf("nodemaintenance services not found or error in sepc.service: %v", err))
		return
	}
	statusServices, found, err := unstructured.NestedSlice(result.Object, "status", "services")
	if err != nil || !found || statusServices == nil {
		helpers.RenderFailureJSON(w, 400, fmt.Sprintf("nodemaintenance services not found or error in status.service: %v", err))
		return
	}
	// frpServerIpAddress
	for _, ss := range specServices {
		if ss.(map[string]interface{})["name"] == fmt.Sprintf("ssh-%v", strings.Split(nodeMaintenanceName, "-")[0]) {
			coreFrp.FrpServerIpAddress = ss.(map[string]interface{})["frpServerIpAddress"].(string)
			coreFrp.ProxyPort = ss.(map[string]interface{})["proxyPort"].(string)
		}
	}

	for _, sss := range statusServices {
		if sss.(map[string]interface{})["name"] == fmt.Sprintf("ssh-%v", strings.Split(nodeMaintenanceName, "-")[0]) {
			coreFrp.Status = sss.(map[string]interface{})["status"].(string)
		}
	}

	helpers.RenderSuccessJSON(w, 200, coreFrp)
	return
}
