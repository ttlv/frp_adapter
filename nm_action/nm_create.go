package nm_action

import (
	"fmt"
	"github.com/ttlv/frp_adapter/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"log"
)

func NmCreate(dynamicClient dynamic.Interface, gvr schema.GroupVersionResource, frpServers []model.FrpServer) (err error) {
	for _, frpServer := range frpServers {
		if frpServer.UniqueID != "" {
			result, getErr := dynamicClient.Resource(gvr).Get(fmt.Sprintf("nodemaintenances-%v", frpServer.UniqueID), metav1.GetOptions{})
			if getErr != nil {
				// 捕捉到错误说明当前frpc的unique的nm对象不存在需要重新创建
				nodeMaintenance := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "edge.harmonycloud.cn/v1alpha1",
						"kind":       "NodeMaintenance",
						"metadata": map[string]interface{}{
							"name":       fmt.Sprintf("nodemaintenances-%v", frpServer.UniqueID),
							"labels":     map[string]interface{}{},
							"annotation": map[string]interface{}{},
						},
						"spec": map[string]interface{}{
							"nodeName": fmt.Sprintf("node-%v", frpServer.UniqueID),
							"proxy": map[string]interface{}{
								"type":     "FRP",
								"endpoint": "",
							},
							"services": []map[string]interface{}{
								{
									"name":               fmt.Sprintf("ssh-%v", frpServer.UniqueID),
									"type":               "ssh",
									"proxyPort":          frpServer.Port,
									"frpServerIpAddress": frpServer.PublicIpAddress,
									"uniqueID":           frpServer.UniqueID,
								},
							},
						},
						"macAddress": frpServer.MacAddress,
						"hostName":   frpServer.HostName,
					},
				}
				// Create Deployment
				log.Println("Creating NodeMaintenance...")
				_, err = dynamicClient.Resource(gvr).Create(nodeMaintenance, metav1.CreateOptions{})
				if err != nil {
					return err
				}
				log.Printf("nodemaintenances-%v is created successfully", frpServer.UniqueID)
			}
			if result != nil {
				log.Println(fmt.Sprintf("%v is already exist and can't be created now", fmt.Sprintf("nodemaintenances-%v", frpServer.UniqueID)))
			}
			//初始化status对象,使用NMNormalUpdate方法，因为frpc的状态可能是false
			// 初始化 model.FrpServer对象
			if err := InitNMUpdate(dynamicClient, gvr, frpServer.UniqueID); err != nil {
				return fmt.Errorf(fmt.Sprintf("Init status object failed,err is: %v", err))
			}
			log.Println(fmt.Sprintf("Init status object Successfully and init %v object successfully", fmt.Sprintf("nodemaintenances-%v", frpServer.UniqueID)))
		}
	}
	return
}
