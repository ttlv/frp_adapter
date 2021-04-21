package nm_action

import (
	"encoding/json"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"strings"
)

func NMFetchAll(dynamicClient dynamic.Interface, gvr schema.GroupVersionResource) (nms []string, err error) {
	lists, err := dynamicClient.Resource(gvr).List(metav1.ListOptions{})
	if err != nil {
		err = fmt.Errorf(fmt.Sprintf("NM fetch failed, err is: %v"))
		return
	}
	for _, list := range lists.Items {
		nmNmae, found, getErr := unstructured.NestedString(list.Object, "metadata", "name")
		if nmNmae == "" || !found || getErr != nil {
			err = fmt.Errorf(fmt.Sprintf("NM fetch failed, err is: %v", getErr))
			return
		}
		nms = append(nms, strings.Split(nmNmae, "-")[1])
	}
	return
}

//获取status信息
func NMGet(dynamicClient dynamic.Interface, gvr schema.GroupVersionResource, uniqueID string) (status string, err error){
	nm, err := dynamicClient.Resource(gvr).Get(fmt.Sprintf("nodemaintenances-%v", uniqueID), metav1.GetOptions{})
	if err != nil {
		err = fmt.Errorf(fmt.Sprintf("NM get failed,err is: %v"))
		return
	}
	nmNmae, found1, getErr := unstructured.NestedSlice(nm.Object, "status","services")
	if  !found1 || getErr != nil {
		err = fmt.Errorf(fmt.Sprintf("NM fetch failed, err is: %v", getErr))
	}
	//nmNmae为[]interface{}，转换为string
	str1, _ :=json.Marshal(nmNmae)
	if strings.Contains(string(str1),"offline"){
		status ="offline"
	}else if strings.Contains(string(str1),"online"){
		status ="online"
	}else {
		err = fmt.Errorf(fmt.Sprintf("NM get failed,err is: %v"))
	}

	return
}

func NMExist(dynamicClient dynamic.Interface, gvr schema.GroupVersionResource, uniqueID string) bool {
	if _, getErr := dynamicClient.Resource(gvr).Get(fmt.Sprintf("nodemaintenances-%v", uniqueID), metav1.GetOptions{}); getErr == nil {
		return true
	}
	return false
}
