package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ttlv/frp_adapter/frp_adapter_init"
	"github.com/ttlv/frp_adapter/frps_action/frps_fetch"
	"github.com/ttlv/frp_adapter/http_server"
	"github.com/ttlv/frp_adapter/model"
	"github.com/ttlv/frp_adapter/nm_action"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method               //请求方法
		origin := c.Request.Header.Get("Origin") //请求头部
		var headerKeys []string                  // 声明请求头keys
		for k, _ := range c.Request.Header {
			headerKeys = append(headerKeys, k)
		}
		headerStr := strings.Join(headerKeys, ", ")
		if headerStr != "" {
			headerStr = fmt.Sprintf("access-control-allow-origin, access-control-allow-headers, %s", headerStr)
		} else {
			headerStr = "access-control-allow-origin, access-control-allow-headers"
		}
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Origin", "*")                                       // 这是允许访问所有域
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE") //服务器支持的所有跨域请求的方法,为了避免浏览次请求的多次'预检'请求
			//  header的类型
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session,X_Requested_With,Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language,DNT, X-CustomHeader, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma")
			//              允许跨域设置                                                                                                      可以返回其他子段
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers,Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma,FooBar") // 跨域关键设置 让浏览器可以解析
			c.Header("Access-Control-Max-Age", "172800")                                                                                                                                                           // 缓存请求信息 单位为秒
			c.Header("Access-Control-Allow-Credentials", "false")                                                                                                                                                  //  跨域请求是否需要带cookie信息 默认设置为true
			c.Set("content-type", "application/json")                                                                                                                                                              // 设置返回格式是json
		}

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "Options Request!")
		}
		// 处理请求
		c.Next() //  处理请求
	}
}

func main() {
	var (
		dynamicClient dynamic.Interface
		err           error
		results       []model.FrpServer
		gvr           = schema.GroupVersionResource{Group: "edge.harmonycloud.cn", Version: "v1alpha1", Resource: "nodemaintenances"}
	)
	dynamicClient, err = frp_adapter_init.NewDynamicClient()
	if err != nil {
		panic(err)
	}
	defer func() {
		if r := recover(); r != nil {
			log.Println("Frp Adapter has been recovered")
		}
	}()
	// frp_adapter初始化获取frps的数据并更新到k8s集群
	if results, err = frps_fetch.FetchFromFrps(); err != nil {
		if err.Error() == "当前无Frpc节点" {
			// Frpa 连不上或者是当前的Frps无任何Frpc节点，需要把k8s中所有的NM对象设置成useless
			if err := nm_action.MakeAllNMUseless(dynamicClient, gvr); err != nil {
				log.Println(err)
			}
		}
		log.Print(err)
	}
	// frpa重启或者是程序初始化时做的一些必要的操作
	if len(results) != 0 {
		frp_adapter_init.FrpAdapterCheck(dynamicClient, gvr, results)
	}
	router := gin.Default()
	router.Use(Cors())
	http_server.New(dynamicClient, frp_adapter_init.FrpsConfig, gvr, router).Run()
}
