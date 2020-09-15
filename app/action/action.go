package action

import (
	"net/http"

	"fmt"
	"github.com/gorilla/sessions"
	"github.com/ttlv/frp_adapter/app/helpers"
	"github.com/ttlv/frp_adapter/kubeconfig_init"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

type Handlers struct {
	SessionStore *sessions.CookieStore
}

func NewHandlers(sessionStore *sessions.CookieStore) Handlers {
	return Handlers{SessionStore: sessionStore}
}

func (handler *Handlers) FrpCreate(w http.ResponseWriter, r *http.Request) {
	namespace := "default"

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig_init.Kubeconfig)
	if err != nil {
		helpers.RenderFailureJSON(w, 400, err.Error())
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		helpers.RenderFailureJSON(w, 400, err.Error())
	}

	nodeMaintenanceRes := schema.GroupVersionResource{Group: "ke.harmonycloud.io", Version: "v1", Resource: "nodemaintenances"}

	nodeMaintenance := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "ke.harmonycloud.io/v1",
			"kind":       "NodeMaintenance",
			"metadata": map[string]interface{}{
				"name":       fmt.Sprintf("nodemaintenances-%v", r.FormValue("unique_id")),
				"labels":     map[string]interface{}{},
				"annotation": map[string]interface{}{},
			},
			"spec": map[string]interface{}{
				"nodeName": fmt.Sprintf("node-%v", r.FormValue("unique_id")),
				"proxy": map[string]interface{}{
					"type": "FRP",
				},
				"services": map[string]interface{}{
					"name":               fmt.Sprintf("ssh-%v", r.FormValue("unique_id")),
					"type":               "ssh",
					"proxyPort":          r.FormValue("port"),
					"FrpServerIpAddress": r.FormValue("frp_server_ip_address"),
					"uniqueID":           r.FormValue("unique_id"),
				},
			},
			"status": map[string]interface{}{
				"services": map[string]interface{}{
					"name":       fmt.Sprintf("ssh-%v", r.FormValue("unique_id")),
					"status":     r.FormValue("status"),
					"lastUpdate": time.Now().UTC().Format("2006-01-02 15:04:05"),
				},
				"conditions": map[string]interface{}{
					"name":      "Maintenable",
					"status":    true,
					"timeStamp": fmt.Sprintf("%v", time.Now().Unix()),
				},
			},
		},
	}

	// Create Deployment
	fmt.Println("Creating NodeMaintenance...")
	result, err := client.Resource(nodeMaintenanceRes).Namespace(namespace).Create(nodeMaintenance, metav1.CreateOptions{})
	if err != nil {
		helpers.RenderFailureJSON(w, 400, err.Error())
	}
	fmt.Printf("Created NodeMaintenance %q.\n", result.GetName())
	helpers.RenderSuccessJSON(w, 200, "Frp client info is created into k8s successfully")
}
