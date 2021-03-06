package nm

import (
	"github.com/gin-gonic/gin"
	"github.com/ttlv/frp_adapter/app/helpers"
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
func (handler *Handlers) NmUseless(c *gin.Context) {
	if err := nm_action.MakeAllNMUseless(handler.DynamicClient, handler.GVR); err != nil {
		helpers.RenderFailureJSON(c, http.StatusBadRequest, err.Error())
		return
	}
	helpers.RenderSuccessJSON(c, 200, "make all nodemaintenances objects become useless successfully")
	return
}
