package k8s_action

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
		err = unstructured.SetNestedField(result.Object, model.FrpOnline, "status", "services", "status")
		if err != nil {
			return fmt.Errorf("SetNestedField status.services.status error: %v", err)
		}
		err = unstructured.SetNestedField(result.Object, fmt.Sprintf("ssh-%v", uniqueID), "status", "services", "name")
		if err != nil {
			return fmt.Errorf("SetNestedField status.services.name error: %v", err)
		}
		err = unstructured.SetNestedField(result.Object, time.Now().UTC().Format("2006-01-02 15:04:05"), "status", "services", "lastUpdate")
		if err != nil {
			return fmt.Errorf("SetNestedField status.services.lastUpdate error: %v", err)
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

		//update Status.condition
		err = unstructured.SetNestedField(result.Object, "maintainable", "status", "conditions", "name")
		if err != nil {
			return fmt.Errorf("SetNestedField status.conditions.name error: %v", err)
		}
		err = unstructured.SetNestedField(result.Object, true, "status", "conditions", "status")
		if err != nil {
			return fmt.Errorf("SetNestedField status.conditions.status error: %v", err)
		}
		err = unstructured.SetNestedField(result.Object, time.Now().UTC().Format("2006-01-02 15:04:05"), "status", "conditions", "timeStamp")
		if err != nil {
			return fmt.Errorf("SetNestedField status.conditions.timeStamp error: %v", err)
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
				return fmt.Errorf("failed to get latest version of NodeMaintenance: %v", getErr)
			}
			if frpServer.PublicIpAddress != "" {
				err = unstructured.SetNestedField(result.Object, frpServer.PublicIpAddress, "spec", "services", "frpServerIpAddress")
				if err != nil {
					return fmt.Errorf("SetNestedField spec.services.frpServerIpAddress error: %v", err)
				}
			}
			if frpServer.Port != "" {
				err = unstructured.SetNestedField(result.Object, frpServer.Port, "spec", "services", "proxyPort")
				if err != nil {
					return fmt.Errorf("SetNestedField spec.services.proxyPort error: %v", err)
				}
			}
			if frpServer.Status != "" {
				err = unstructured.SetNestedField(result.Object, frpServer.Status, "status", "services", "status")
				if err != nil {
					return fmt.Errorf("SetNestedField status.services.status error: %v", err)
				}
			}
			_, updateErr := dynamicClient.Resource(gvr).Update(result, metav1.UpdateOptions{}, "status")
			return updateErr
		})
		if retryErr != nil {
			stringErr += fmt.Sprintf("[%v]%v \n", fmt.Sprintf("nodemaintenances-%v", frpServer.UniqueID), retryErr)
		}
	}
	if stringErr != "" {
		return fmt.Errorf(stringErr)
	}
	return
}
