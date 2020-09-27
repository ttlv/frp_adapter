package main

import (
	"github.com/rs/cors"
	"github.com/ttlv/frp_adapter/frp_adapter_init"
	"github.com/ttlv/frp_adapter/frps_action/frps_fetch"
	"github.com/ttlv/frp_adapter/http_server"
	"github.com/ttlv/frp_adapter/model"
	"github.com/ttlv/frp_adapter/nm_action"
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
	router := http_server.New(dynamicClient, frp_adapter_init.FrpsConfig, gvr)
	handler := cs.Handler(router)

	log.Printf("========== Visit http://%v ==========\n", frp_adapter_init.FrpAdapterConfig.Address)
	log.Fatal(http.ListenAndServe(frp_adapter_init.FrpAdapterConfig.Address, handler))
}
