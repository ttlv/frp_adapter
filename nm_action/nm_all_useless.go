package nm_action

import (
	"fmt"
	"github.com/ttlv/frp_adapter/model"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func MakeAllNMUseless(dynamicClient dynamic.Interface, gvr schema.GroupVersionResource) (err error) {
	var (
		uselessUniqueIDs []string
		uselessNMs       []model.FrpServer
	)
	if uselessUniqueIDs, err = NMFetch(dynamicClient, gvr); err != nil {
		err = fmt.Errorf("can't fetch nodemaintenances objects from k8s cluster")
		return
	}
	for _, uniqueID := range uselessUniqueIDs {
		uselessNMs = append(uselessNMs, model.FrpServer{
			UniqueID: uniqueID,
			Status:   model.FrpOffline,
		})
	}
	if err = NMNormalUpdate(dynamicClient, gvr, uselessNMs); err != nil {
		err = fmt.Errorf("can't make all nodemaintenances objects become useless")
		return
	}
}
