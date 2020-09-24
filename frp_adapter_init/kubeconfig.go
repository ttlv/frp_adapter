package frp_adapter_init

import (
	"flag"
	"github.com/ttlv/frp_adapter/config"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

var (
	Kubeconfig       *string
	FrpsConfig       *config.FrpsConfig
	FrpAdapterConfig *config.FrpAdapterConfig
)

func init() {
	// init kubeconfig
	if home := homedir.HomeDir(); home != "" {
		Kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		Kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	// init FrpsConfig
	frpsConfig := config.MustGetFrpsConfig()
	frpAdapterConfig := config.MustGetFrpAdapterConfig()
	FrpsConfig = &frpsConfig
	FrpAdapterConfig = &frpAdapterConfig
}

func NewDynamicClient() (client dynamic.Interface, err error) {
	config, err := clientcmd.BuildConfigFromFlags("", *Kubeconfig)
	if err != nil {
		return nil, err
	}
	if client, err = dynamic.NewForConfig(config); err != nil {
		return nil, err
	}
	return
}
