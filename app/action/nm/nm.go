package nm

import (
	"github.com/ttlv/frp_adapter/app/helpers"
	"github.com/ttlv/frp_adapter/model"
	"github.com/ttlv/frp_adapter/nm_action"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"net/http"
)

type Handlers struct {
	DynamicClient dynamic.Interface
	GVR           schema.GroupVersionResource
}

func NewHandlers(dynamicClient dynamic.Interface, gvr schema.GroupVersionResource) Handlers {
	return Handlers{DynamicClient: dynamicClient, GVR: gvr}
}

// 当frp server服务停止服务时应该让所有的nm对象全部变成unmaintainable
func (handler *Handlers) NmUseless(w http.ResponseWriter, r *http.Request) {
	var (
		uselessUniqueIDs []string
		err              error
		uselessNMs       []model.FrpServer
	)
	if uselessUniqueIDs, err = nm_action.NMFetch(handler.DynamicClient, handler.GVR); err != nil {
		helpers.RenderFailureJSON(w, 400, "can't fetch nodemaintenances objects from k8s cluster")
		return
	}
	for _, uniqueID := range uselessUniqueIDs {
		uselessNMs = append(uselessNMs, model.FrpServer{
			UniqueID: uniqueID,
			Status:   model.FrpOffline,
		})
	}
	if err = nm_action.NMNormalUpdate(handler.DynamicClient, handler.GVR, uselessNMs); err != nil {
		helpers.RenderFailureJSON(w, 400, "can't make all nodemaintenances objects become useless")
		return
	}
	helpers.RenderSuccessJSON(w, 200, "make all nodemaintenances objects become useless successfully")
	return
}
