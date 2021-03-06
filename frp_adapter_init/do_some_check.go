package frp_adapter_init

import (
	"fmt"
	"github.com/ttlv/frp_adapter/model"
	"github.com/ttlv/frp_adapter/nm_action"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"log"
)

// 如果在frp_adapter重启的期间，frps正常运行，此时恰好有新的frpc发起了连接，那么当前frpc的nm对象不会创建
// 因为frp_adapter正在启动中，会出现一种情况，frps有新的frpc注册，但是k8s中没有该frpc的nm对象，此情况需要
// frp_adapter初始化后遇到这种情况应该立刻去创建nm对象
// 第一步应该是先获取k8s中nm的资源对象个数，和frps中的frpc的个数做对比，如果发现k8s中不存在frps中存在的unique_id立刻创建
// 如果存在k8s中的unique_id frps中没有的，则忽略，一切以frps为准
func FrpAdapterCheck(dynamicClient dynamic.Interface, gvr schema.GroupVersionResource, results []model.FrpServer) (err error) {
	var (
		nms       []string
		errString string
	)
	if nms, err = nm_action.NMFetchAll(dynamicClient, gvr); err != nil {
		errString += fmt.Sprintf("%v \n", err)
	} else {
		var (
			frpsUniqueIDs, shortUniqueIDs, needUpdateUniqueIDs, uselessUniqueIDs []string
			shortFrps, needUpdateFrps, uselessFrps                               []model.FrpServer
		)

		for _, result := range results {
			frpsUniqueIDs = append(frpsUniqueIDs, result.UniqueID)
		}
		if len(nms) == 0 {
			// shortUniqueIDs数组长度为0说明k8s中不存在nm对象，要把results全部创建
			log.Println("There is not any unique_id in k8s cluster and frp adapter will create all nodemaintenances in k8s cluster")
			if err = nm_action.NmCreate(dynamicClient, gvr, results); err != nil {
				// 如果出现无法创建的错误一般都是k8s集群存在问题，重试毫无意义，仅仅在日志中打印错误，不会继续后续的InitNMUpdate和NMNormalUpdate操作
				errString += fmt.Sprintf("There are some fatal errors in k8s cluster \n")
			} else {
				for _, result := range results {
					if err = nm_action.InitNMUpdate(dynamicClient, gvr, result.UniqueID); err != nil {
						errString += fmt.Sprintf("%v \n", err)
					}
				}
				if err = nm_action.NMNormalUpdate(dynamicClient, gvr, results); err != nil {
					errString += fmt.Sprintf("%v \n", err)
				} else {
					for _, short := range results {
						log.Printf("update nodemaintenances-%v successfully", short.UniqueID)
					}
				}
			}
		} else {
			for _, unique_id := range frpsUniqueIDs {
				count := 0
				for _, nm := range nms {
					if nm != unique_id {
						count += 1
					} else {
						needUpdateUniqueIDs = append(needUpdateUniqueIDs, nm)
					}
					if count == len(nms) {
						shortUniqueIDs = append(shortUniqueIDs, unique_id)
					}
				}
				// shortUniqueIDs数组长度为0说明frps与k8s中unique_id的个数相同
				if len(shortUniqueIDs) == 0 {
					log.Println("the number of unique_id is equal to the number of nodemaintenances which in k8s cluster")
				} else {
					for _, unique_id := range shortUniqueIDs {
						for _, result := range results {
							if result.UniqueID == unique_id {
								shortFrps = append(shortFrps, result)
							}
						}
					}
					err = nm_action.NmCreate(dynamicClient, gvr, shortFrps)
					// nm对象无法创建可能是k8s集群出了问题，此时重试也毫无意义，直接在日志中打印，等集群恢复正常，重启frps或者是重启frpc即可恢复正常。
					if err != nil {
						errString += fmt.Sprintf("%v \n", err)
					}
					for _, unique_id := range shortUniqueIDs {
						if err = nm_action.InitNMUpdate(dynamicClient, gvr, unique_id); err != nil {
							errString += fmt.Sprintf("%v \n", err)
						}
					}

					if err = nm_action.NMNormalUpdate(dynamicClient, gvr, shortFrps); err != nil {
						errString += fmt.Sprintf("%v \n", err)
					} else {
						for _, short := range shortFrps {
							log.Printf("update nodemaintenances-%v successfully", short.UniqueID)
						}
					}
				}
				// needUpdateFrps数组长度为0说明没有需要更新的unique_id
				if len(needUpdateUniqueIDs) == 0 {
					log.Println("There are no another unique_ids which are needed to be updated")
				} else {
					for _, result := range results {
						for _, uniqueID := range needUpdateUniqueIDs {
							if uniqueID == result.UniqueID {
								needUpdateFrps = append(needUpdateFrps, result)
							}
						}
					}
					if err = nm_action.NMNormalUpdate(dynamicClient, gvr, needUpdateFrps); err != nil {
						errString += fmt.Sprintf("%v \n", err)
					} else {
						for _, update := range needUpdateFrps {
							log.Printf("update nodemaintenances-%v successfully", update.UniqueID)
						}
					}
				}
				// k8s nm的unique_id比frps多的情况，要把这些多余的全部设置成offline和unmaintainable
				// 这种情况通常是废弃的unique_id,而frpa不会去删除这些无效的nm对象
				for _, nm := range nms {
					for _, unique_id := range frpsUniqueIDs {
						count := 0
						if unique_id != nm {
							count += 1
						}
						if count == len(frpsUniqueIDs) {
							uselessUniqueIDs = append(uselessUniqueIDs, nm)
						}
					}
				}
				// uselessUniqueIDs数组长度为0则说明k8s中不存在废弃的unique_id
				if len(uselessUniqueIDs) == 0 {
					log.Println("There are no any useless unique_id in k8s cluster")
				} else {
					for _, unique_id := range uselessUniqueIDs {
						uselessFrps = append(uselessFrps, model.FrpServer{
							UniqueID: unique_id,
							Status:   model.FrpOffline,
						})
					}
					if err = nm_action.NMNormalUpdate(dynamicClient, gvr, uselessFrps); err != nil {
						errString += fmt.Sprintf("%v \n", err)
					} else {
						for _, useless := range uselessFrps {
							log.Printf("update nodemaintenances-%v successfully", useless.UniqueID)
						}
					}
				}
			}
		}
	}
	return fmt.Errorf(errString)
}
