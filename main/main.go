package main

import (
	"github.com/rs/cors"
	"github.com/ttlv/frp_adapter/frp_adapter_init"
	"github.com/ttlv/frp_adapter/frps_action/frps_fetch"
	"github.com/ttlv/frp_adapter/http_server"
	"github.com/ttlv/frp_adapter/k8s_action"
	"github.com/ttlv/frp_adapter/model"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"log"
	"net/http"
)

func main() {
	var (
		dynamicClient dynamic.Interface
		err           error
		results       []model.FrpServer
		gvr           = schema.GroupVersionResource{Group: "edge.harmonycloud.cn", Version: "v1alpha1", Resource: "nodemaintenances"}
	)
	cs := cors.New(cors.Options{
		//AllowedOrigins:   []string{"http://localhost:3002"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization"},
		Debug:            true,
	})
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
		log.Print(err)
	}
	// 如果在frp_adapter重启的期间，frps正常运行，此时恰好有新的frpc发起了连接，那么当前frpc的nm对象不会创建
	// 因为frp_adapter正在启动中，会出现一种情况，frps有新的frpc注册，但是k8s中没有该frpc的nm对象，此情况需要
	// frp_adapter初始化后遇到这种情况应该立刻去创建nm对象
	err = k8s_action.NmCreate(dynamicClient, gvr, results)
	// nm对象无法创建可能是k8s集群出了问题，此时重试也毫无意义，直接在日志中打印，等集群恢复正常，重启frps或者是重启frpc即可恢复正常。
	if err != nil {
		log.Println(err)
	}
	err = k8s_action.NMNormalUpdate(dynamicClient, gvr, results)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Update NM successfully")
	}

	router := http_server.New(dynamicClient, frp_adapter_init.FrpsConfig, gvr)
	handler := cs.Handler(router)

	log.Printf("========== Visit http://%v ==========\n", frp_adapter_init.FrpAdapterConfig.Address)
	log.Fatal(http.ListenAndServe(frp_adapter_init.FrpAdapterConfig.Address, handler))
}
