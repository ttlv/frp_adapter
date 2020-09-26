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
		log.Print(err)
	}
	if len(results) != 0 {
		// 如果在frp_adapter重启的期间，frps正常运行，此时恰好有新的frpc发起了连接，那么当前frpc的nm对象不会创建
		// 因为frp_adapter正在启动中，会出现一种情况，frps有新的frpc注册，但是k8s中没有该frpc的nm对象，此情况需要
		// frp_adapter初始化后遇到这种情况应该立刻去创建nm对象
		// 第一步应该是先获取k8s中nm的资源对象个数，和frps中的frpc的个数做对比，如果发现k8s中不存在frps中存在的unique_id立刻创建
		// 如果存在k8s中的unique_id frps中没有的，则忽略，一切以frps为准
		nms, err := nm_action.NMFetch(dynamicClient, gvr)
		if err != nil {
			log.Println(err)
		} else {
			var (
				frpsUniqueIDs, shortUniqueIDs, needUpdateUniqueIDs, uselessUniqueIDs []string
				shortFrps, needUpdateFrps, uselessFrps                               []model.FrpServer
			)

			for _, result := range results {
				frpsUniqueIDs = append(frpsUniqueIDs, result.UniqueID)
			}
			for _, unique_id := range frpsUniqueIDs {
				for _, nm := range nms {
					count := 0
					if nm != unique_id {
						count += 1
					} else {
						needUpdateUniqueIDs = append(needUpdateUniqueIDs, nm)
					}
					if count == len(nms) {
						shortUniqueIDs = append(shortUniqueIDs, nm)
					}
				}
				// shortUniqueIDs数组为0说明k8s中不存在nm对象，要把results全部创建
				if len(shortUniqueIDs) == 0 {
					for _, result := range results {
						shortFrps = append(shortFrps, result)
					}
				} else {
					for _, unique_id := range shortUniqueIDs {
						for _, result := range results {
							if result.UniqueID == unique_id {
								shortFrps = append(shortFrps, result)
							}
						}
					}
				}
			}

			err = nm_action.NmCreate(dynamicClient, gvr, shortFrps)
			// nm对象无法创建可能是k8s集群出了问题，此时重试也毫无意义，直接在日志中打印，等集群恢复正常，重启frps或者是重启frpc即可恢复正常。
			if err != nil {
				log.Println(err)
			}
			for _, unique_id := range shortUniqueIDs {
				if err = nm_action.InitNMUpdate(dynamicClient, gvr, unique_id); err != nil {
					log.Println(err)
				}
			}

			if err = nm_action.NMNormalUpdate(dynamicClient, gvr, shortFrps); err != nil {
				log.Println(err)
			} else {
				for _, short := range shortFrps {
					log.Printf("update nodemaintenances-%v successfully", short.UniqueID)
				}
			}
			// 更新frps与k8s都有的unique_id,强制更新needUpdateUniqueIDs
			for _, result := range results {
				for _, uniqueID := range needUpdateUniqueIDs {
					if uniqueID == result.UniqueID {
						needUpdateFrps = append(needUpdateFrps, result)
					}
				}
			}
			if err = nm_action.NMNormalUpdate(dynamicClient, gvr, needUpdateFrps); err != nil {
				log.Println(err)
			} else {
				for _, update := range needUpdateFrps {
					log.Printf("update nodemaintenances-%v successfully", update.UniqueID)
				}
			}
			// nm的unique_id比frps多的情况，要把这些多余的全部设置成offline和unmaintainable
			if len(nms) != 0 {
				for _, nm := range nms {
					for _, unique_id := range frpsUniqueIDs {
						count := 0
						if unique_id != nm {
							count += 1
						}
						if count == len(frpsUniqueIDs) {
							uselessUniqueIDs = append(uselessUniqueIDs, unique_id)
						}
					}
				}
				for _, unique_id := range uselessUniqueIDs {
					for _, result := range results {
						if result.UniqueID == unique_id {
							uselessFrps = append(uselessFrps, result)
						}
					}
				}
				if err = nm_action.NMNormalUpdate(dynamicClient, gvr, uselessFrps); err != nil {
					log.Println(err)
				} else {
					for _, useless := range uselessFrps {
						log.Printf("update nodemaintenances-%v successfully", useless.UniqueID)
					}
				}
			}
		}
	}
	router := http_server.New(dynamicClient, frp_adapter_init.FrpsConfig, gvr)
	handler := cs.Handler(router)

	log.Printf("========== Visit http://%v ==========\n", frp_adapter_init.FrpAdapterConfig.Address)
	log.Fatal(http.ListenAndServe(frp_adapter_init.FrpAdapterConfig.Address, handler))
}
