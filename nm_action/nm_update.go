package nm_action

import (
	"fmt"
	"github.com/ttlv/frp_adapter/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/retry"
	"log"
	"time"
)

func InitNMUpdate(dynamicClient dynamic.Interface, gvr schema.GroupVersionResource, uniqueID string) (err error) {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// 初始化NodeMaintenanceStatus对象
		// update Status.services

		result, getErr := dynamicClient.Resource(gvr).Get(fmt.Sprintf("nodemaintenances-%v", uniqueID), metav1.GetOptions{})
		if getErr != nil {
			return fmt.Errorf("failed to get latest version of NodeMaintenance: %v", getErr)
		}
		servicesMap := make(map[string]interface{})
		servicesMap["name"] = fmt.Sprintf("ssh-%v", uniqueID)
		servicesMap["status"] = model.FrpOnline
		servicesMap["lastUpdate"] = time.Now().UTC().Format("2006-01-02 15:04:05")

		err = unstructured.SetNestedSlice(result.Object, []interface{}{servicesMap}, "status", "services")
		if err != nil {
			return fmt.Errorf("SetNestedField status.services error: %v", err)
		}

		//update Status.bindStatus
		err = unstructured.SetNestedField(result.Object, "Unbound", "status", "bindStatus", "phase")
		if err != nil {
			return fmt.Errorf("SetNestedField status.bindStatus.phase error: %v", err)
		}
		err = unstructured.SetNestedField(result.Object, "", "status", "bindStatus", "nodeDeploymentReference")
		if err != nil {
			return fmt.Errorf("SetNestedField status.bindStatus.nodeDeploymentReference error: %v", err)
		}
		err = unstructured.SetNestedField(result.Object, "", "status", "bindStatus", "timeStamp")
		if err != nil {
			return fmt.Errorf("SetNestedField status.bindStatus.status error: %v", err)
		}

		conditionMap := make(map[string]interface{})
		conditionMap["name"] = "maintainable"
		conditionMap["status"] = true
		conditionMap["timeStamp"] = time.Now().UTC().Format("2006-01-02 15:04:05")

		//update Status.condition
		err = unstructured.SetNestedSlice(result.Object, []interface{}{conditionMap}, "status", "conditions")
		if err != nil {
			return fmt.Errorf("SetNestedField status.conditions.name error: %v", err)
		}
		_, updateErr := dynamicClient.Resource(gvr).Update(result, metav1.UpdateOptions{}, "status")
		return updateErr
	})
	if retryErr != nil {
		return err
	}
	log.Println("init status Successfully")
	return
}

func NMNormalUpdate(dynamicClient dynamic.Interface, gvr schema.GroupVersionResource, frpServers []model.FrpServer) (err error) {
	var stringErr string
	for _, frpServer := range frpServers {
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			result, getErr := dynamicClient.Resource(gvr).Get(fmt.Sprintf("nodemaintenances-%v", frpServer.UniqueID), metav1.GetOptions{})
			if getErr != nil {
				return fmt.Errorf("failed to get latest version of NodeMaintenance-%v,err is: %v", frpServer.UniqueID, getErr)
			}
			// update sepc.services
			if frpServer.PublicIpAddress != "" && frpServer.Port != "" {
				objs, _, err := unstructured.NestedSlice(result.Object, "spec", "services")
				if err != nil {
					return fmt.Errorf("NestedSlice spec.services error: %v", err)
				}
				for _, obj := range objs {
					if obj.(map[string]interface{})["name"] == fmt.Sprintf("ssh-%v", frpServer.UniqueID) {
						obj.(map[string]interface{})["frpServerIpAddress"] = frpServer.PublicIpAddress
						obj.(map[string]interface{})["proxyPort"] = frpServer.Port
						obj.(map[string]interface{})["hostName"] = frpServer.HostName
						obj.(map[string]interface{})["macAddress"] = frpServer.MacAddress
					}
				}
				if err = unstructured.SetNestedSlice(result.Object, objs, "spec", "services"); err != nil {
					return fmt.Errorf("SetNestedSlice spec.services error: %v", err)
				}
			}
			// update status.services & update status.conditions
			if frpServer.Status != "" {
				if frpServer.Status == model.FrpOnline {
					// change status.conditions firstly
					objs, _, err := unstructured.NestedSlice(result.Object, "status", "conditions")
					if err != nil {
						return fmt.Errorf("NestedSlice spec.conditions error: %v", err)
					}
					for _, obj := range objs {
						obj.(map[string]interface{})["name"] = "maintainable"
						obj.(map[string]interface{})["status"] = true
						obj.(map[string]interface{})["timeStamp"] = time.Now().UTC().Format("2006-01-02 15:04:05")
					}
					if err = unstructured.SetNestedSlice(result.Object, objs, "status", "conditions"); err != nil {
						return fmt.Errorf("SetNestedSlice status.conditions error: %v", err)
					}
					// change status.services
					objs, _, err = unstructured.NestedSlice(result.Object, "status", "services")
					if err != nil {
						return fmt.Errorf("NestedSlice spec.services error: %v", err)
					}
					for _, obj := range objs {
						if obj.(map[string]interface{})["name"] == fmt.Sprintf("ssh-%v", frpServer.UniqueID) {
							obj.(map[string]interface{})["lastUpdate"] = time.Now().UTC().Format("2006-01-02 15:04:05")
							obj.(map[string]interface{})["status"] = model.FrpOnline
						}
					}
					if err = unstructured.SetNestedSlice(result.Object, objs, "status", "services"); err != nil {
						return fmt.Errorf("SetNestedSlice status.services error: %v", err)
					}
				} else {
					// change status.conditions firstly
					objs, _, err := unstructured.NestedSlice(result.Object, "status", "conditions")
					if err != nil {
						return fmt.Errorf("NestedSlice spec.conditions error: %v", err)
					}
					// TODO 目前conditions只会存在一条记录，如果存在多条记录需要修改status,需要加一个name予以区分,当前没有name,所以直接强制更新。
					for _, obj := range objs {
						obj.(map[string]interface{})["name"] = "unmaintainable"
						obj.(map[string]interface{})["status"] = false
						obj.(map[string]interface{})["timeStamp"] = time.Now().UTC().Format("2006-01-02 15:04:05")
					}
					if err = unstructured.SetNestedSlice(result.Object, objs, "status", "conditions"); err != nil {
						return fmt.Errorf("SetNestedSlice status.conditions error: %v", err)
					}
					// change status.services
					objs, _, err = unstructured.NestedSlice(result.Object, "status", "services")
					if err != nil {
						return fmt.Errorf("NestedSlice spec.services error: %v", err)
					}
					for _, obj := range objs {
						if obj.(map[string]interface{})["name"] == fmt.Sprintf("ssh-%v", frpServer.UniqueID) {
							obj.(map[string]interface{})["lastUpdate"] = time.Now().UTC().Format("2006-01-02 15:04:05")
							obj.(map[string]interface{})["status"] = model.FrpOffline
						}
					}
					if err = unstructured.SetNestedSlice(result.Object, objs, "status", "services"); err != nil {
						return fmt.Errorf("SetNestedSlice status.services error: %v", err)
					}
				}
			}
			_, updateErr := dynamicClient.Resource(gvr).Update(result, metav1.UpdateOptions{}, "status")
			return updateErr
		})
		if retryErr != nil {
			stringErr += fmt.Sprintf("[%v]update err: %v \n", fmt.Sprintf("nodemaintenances-%v", frpServer.UniqueID), retryErr)
		}
	}
	if stringErr != "" {
		return fmt.Errorf(stringErr)
	}
	return
}
