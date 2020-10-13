package http_server

import (
	"github.com/ttlv/frp_adapter/app/action/frp"
	"github.com/ttlv/frp_adapter/app/action/nm"
	"github.com/ttlv/frp_adapter/app/action/reverse_proxy"

	"github.com/gin-gonic/gin"
	"github.com/ttlv/frp_adapter/config"
	"github.com/ttlv/frp_adapter/home"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func New(dynamicClient dynamic.Interface, frpsConfig *config.FrpsConfig, gvr schema.GroupVersionResource, router *gin.Engine) *gin.Engine {

	homeHandlers := home.NewHandlers()
	frpHandlers := frp.NewHandlers(dynamicClient, gvr)
	nmUselessHandlers := nm.NewHandlers(dynamicClient, gvr)
	reverseProxyHandlers := reverse_proxy.NewHandlers(1024, 1024*1024*10)
	// home页面,主要用于发起请求验证frp adapter是否存活
	router.GET("/", homeHandlers.Home)

	router.GET("/frp_fetch/:nm_name", frpHandlers.FrpFetch) // GET /frp_adapter/fetch/xxxxxx
	router.POST("/frp_create", frpHandlers.FrpCreate)       // POST /frp_adapter/create
	router.PUT("/frp_update", frpHandlers.FrpUpdate)        // PUT  /frp_adapter/update

	router.PUT("/nm_useless", nmUselessHandlers.NmUseless) // PUT /nm_useless make all nodemaintenances objects become useless

	// reserve proxy
	router.GET("/reverse_proxy/:nm_name", reverseProxyHandlers.ReverseProxy) //反向代理建立ssh session并维持长连接

	// reserve shell command
	router.POST("/reverse_shell_command/:nm_name", reverseProxyHandlers.ReverseProxySshCommand)
	return router
}
