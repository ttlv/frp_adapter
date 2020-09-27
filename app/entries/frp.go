package entries

// 从k8s 获取到的有关NM的核心数据结构
type CoreFrp struct {
	FrpServerIpAddress string `json:"frp_server_ip_address"`
	ProxyPort          string `json:"proxy_port"`
	Status             string `json:"status"`
}
