package frp

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ttlv/frp_adapter/app/entries"
	"github.com/ttlv/frp_adapter/app/helpers"
	"github.com/ttlv/frp_adapter/model"
	"github.com/ttlv/frp_adapter/nm_action"
	"github.com/ttlv/frp_adapter/notification"
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
func (handler *Handlers) FrpCreate(c *gin.Context) {
	var (
		nms []model.FrpServer
	)
	nms = append(nms, model.FrpServer{
		PublicIpAddress: c.PostForm("frp_server_ip_address"),
		Status:          model.FrpOnline,
		UniqueID:        c.PostForm("unique_id"),
		Port:            c.PostForm("port"),
		MacAddress:      c.PostForm("mac_address"),
	})
	if err := nm_action.NmCreate(handler.DynamicClient, handler.GVR, nms); err != nil {
		helpers.RenderFailureJSON(c, http.StatusBadRequest, fmt.Sprintf("can't create nodemaintenances crd resource in k8s cluster,err is:%v", err))
		return
	}
	//nm创建成功后，发送通知
	notification.Notice("创建","Not created",c)
	helpers.RenderSuccessJSON(c, http.StatusOK, fmt.Sprintf("create nodemaintenances-%v crd resource in k8s cluster successfully", c.PostForm("unique_id")))
	return
}

// 当frpc的状态更新时需要立即更新nodemaintenances资源
func (handler *Handlers) FrpUpdate(c *gin.Context) {
	var (
		err        error
		frpServers = []model.FrpServer{}
	)
	frpServers = append(frpServers, model.FrpServer{
		PublicIpAddress: c.Request.FormValue("frp_server_ip_address"),
		Status:          c.Request.FormValue("status"),
		UniqueID:        c.Request.FormValue("unique_id"),
		Port:            c.Request.FormValue("port"),
		HostName:        c.Request.FormValue("host_name"),
	})
	// 更新前判断nm资源是否存在，避免frpc已经接入frps但是没有nm对象的情况，如果不存在应该先创建
	if !nm_action.NMExist(handler.DynamicClient, handler.GVR, c.Request.FormValue("unique_id")) {
		err = nm_action.NmCreate(handler.DynamicClient, handler.GVR, frpServers)
		if err != nil {
			helpers.RenderFailureJSON(c, http.StatusBadRequest, fmt.Sprintf("created failed: %v", err))
			return
		}
	}
	//在nm更新前获取status
	nm ,err:=nm_action.NMFetchOne(handler.DynamicClient, handler.GVR, c.Request.FormValue("unique_id"))
	if err != nil {
		helpers.RenderFailureJSON(c, http.StatusBadRequest, fmt.Sprintf("created failed: %v", err))
		return
	}

	err = nm_action.NMNormalUpdate(handler.DynamicClient, handler.GVR, frpServers)
	if err != nil {
		helpers.RenderFailureJSON(c, http.StatusBadRequest, fmt.Sprintf("update failed: %v", err))
		return
	}
	//更新成果后发送通知
	notification.Notice("更新",nm.Status.Services[0].Status,c)
	helpers.RenderSuccessJSON(c, http.StatusOK, "Update Successfully")
	return
}

// Frps请求Frp Adapter获取nodemaintenances资源数据
func (handler *Handlers) FrpFetch(c *gin.Context) {
	var (
		nodeMaintenanceName = c.Param("nm_name")
		coreFrp             = entries.CoreFrp{}
	)
	if nodeMaintenanceName == "" {
		helpers.RenderFailureJSON(c, http.StatusBadRequest, "nodemaintenances name为空")
		return
	}
	result, getErr := handler.DynamicClient.Resource(handler.GVR).Get(nodeMaintenanceName, metav1.GetOptions{})
	if getErr != nil {
		helpers.RenderFailureJSON(c, http.StatusNotFound, fmt.Sprintf("failed to get latest version of nodeMaintenance: %v", getErr))
		return
	}
	specServices, found, err := unstructured.NestedSlice(result.Object, "spec", "services")
	if err != nil || !found || specServices == nil {
		helpers.RenderFailureJSON(c, http.StatusBadRequest, fmt.Sprintf("nodemaintenance services not found or error in sepc.service: %v", err))
		return
	}
	statusServices, found, err := unstructured.NestedSlice(result.Object, "status", "services")
	if err != nil || !found || statusServices == nil {
		helpers.RenderFailureJSON(c, http.StatusBadRequest, fmt.Sprintf("nodemaintenance services not found or error in status.service: %v", err))
		return
	}
	// frpServerIpAddress
	for _, ss := range specServices {
		if ss.(map[string]interface{})["name"] == fmt.Sprintf("ssh-%v", strings.Split(nodeMaintenanceName, "-")[1]) {
			coreFrp.FrpServerIpAddress = ss.(map[string]interface{})["frpServerIpAddress"].(string)
			coreFrp.ProxyPort = ss.(map[string]interface{})["proxyPort"].(string)
		}
	}
	for _, sss := range statusServices {
		if sss.(map[string]interface{})["name"] == fmt.Sprintf("ssh-%v", strings.Split(nodeMaintenanceName, "-")[1]) {
			coreFrp.Status = sss.(map[string]interface{})["status"].(string)
		}
	}
	helpers.RenderSuccessJSON(c, http.StatusOK, coreFrp)
	return
}
