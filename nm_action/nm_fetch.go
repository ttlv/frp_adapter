package nm_action

import (
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

func NMExist(dynamicClient dynamic.Interface, gvr schema.GroupVersionResource, uniqueID string) bool {
	if _, getErr := dynamicClient.Resource(gvr).Get(fmt.Sprintf("nodemaintenances-%v", uniqueID), metav1.GetOptions{}); getErr == nil {
		return true
	}
	return false
}
