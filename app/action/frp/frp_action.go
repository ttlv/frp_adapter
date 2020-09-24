package frp

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/ttlv/frp_adapter/app/entries"
	"github.com/ttlv/frp_adapter/app/helpers"
	"github.com/ttlv/frp_adapter/k8s_action"
	"github.com/ttlv/frp_adapter/model"
	"k8s.io/client-go/dynamic"
	"net/http"

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

// 当有新的frpc注册时立即创建新的nodemaintenances对象
func (handler *Handlers) FrpCreate(w http.ResponseWriter, r *http.Request) {
	result, getErr := handler.DynamicClient.Resource(handler.GVR).Get(fmt.Sprintf("nodemaintenances-%v", r.FormValue("unique_id")), metav1.GetOptions{})
	if getErr != nil {
		// 优先判断当前nodemaintenances对象是否存在，如果存在则不创建
		nodeMaintenance := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "edge.harmonycloud.cn/v1alpha1",
				"kind":       "NodeMaintenance",
				"metadata": map[string]interface{}{
					"name":       fmt.Sprintf("nodemaintenances-%v", r.FormValue("unique_id")),
					"labels":     map[string]interface{}{},
					"annotation": map[string]interface{}{},
				},
				"spec": map[string]interface{}{
					"nodeName": fmt.Sprintf("node-%v", r.FormValue("unique_id")),
					"proxy": map[string]interface{}{
						"type":     "FRP",
						"endpoint": "",
					},
					"services": map[string]interface{}{
						"name":               fmt.Sprintf("ssh-%v", r.FormValue("unique_id")),
						"type":               "ssh",
						"proxyPort":          r.FormValue("port"),
						"frpServerIpAddress": r.FormValue("frp_server_ip_address"),
						"uniqueID":           r.FormValue("unique_id"),
					},
				},
			},
		}
		// Create Deployment
		fmt.Println("Creating NodeMaintenance...")
		_, err := handler.DynamicClient.Resource(handler.GVR).Create(nodeMaintenance, metav1.CreateOptions{})
		if err != nil {
			helpers.RenderFailureJSON(w, 400, err.Error())
			return
		}
		helpers.RenderSuccessJSON(w, 200, "Frp client info is created into k8s successfully")
	}
	if result != nil {
		helpers.RenderFailureJSON(w, 400, fmt.Sprintf("%v is already exist and can't be created now", fmt.Sprintf("nodemaintenances-%v", r.FormValue("unique_id"))))
	}
	//初始化status对象
	if err := k8s_action.InitNMUpdate(handler.DynamicClient, handler.GVR, r.FormValue("unique_id")); err != nil {
		helpers.RenderFailureJSON(w, 400, fmt.Sprintf("Init status object failed,err is: %v", err))
		return
	}
	helpers.RenderSuccessJSON(w, 200, fmt.Sprintf("Init status object Successfully and init %v object successfully", fmt.Sprintf("nodemaintenances-%v", r.FormValue("unique_id"))))
	return
}

// 当frpc的状态更新时需要立即更新nodemaintenances资源
func (handler *Handlers) FrpUpdate(w http.ResponseWriter, r *http.Request) {
	frpServers := []model.FrpServer{}
	frpServers = append(frpServers, model.FrpServer{
		PublicIpAddress: r.FormValue("frp_server_ip_address"),
		Status:          r.FormValue("status"),
		UniqueID:        r.FormValue("unique_id"),
		Port:            r.FormValue("port"),
	})
	err := k8s_action.NMNormalUpdate(handler.DynamicClient, handler.GVR, frpServers)
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
		ok                  bool
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
	specServices, found, err := unstructured.NestedMap(result.Object, "spec", "services")
	if err != nil || !found || specServices == nil {
		helpers.RenderFailureJSON(w, 400, fmt.Sprintf("nodemaintenance services not found or error in sepc.service: %v", err))
		return
	}
	statusServices, found, err := unstructured.NestedMap(result.Object, "status", "services")
	if err != nil || !found || statusServices == nil {
		helpers.RenderFailureJSON(w, 400, fmt.Sprintf("nodemaintenance services not found or error in status.service: %v", err))
		return
	}
	if coreFrp.FrpServerIpAddress, ok = specServices["frpServerIpAddress"].(string); !ok {
		helpers.RenderFailureJSON(w, 400, "invalid value for FrpServerIpAddress")
		return
	}
	if coreFrp.ProxyPort, ok = specServices["proxyPort"].(string); !ok {
		helpers.RenderFailureJSON(w, 400, "invalid value for ProxyPort")
		return
	}
	if coreFrp.Status, ok = statusServices["status"].(string); !ok {
		helpers.RenderFailureJSON(w, 400, "invalid value for status")
		return
	}
	helpers.RenderSuccessJSON(w, 200, coreFrp)
	return
}
