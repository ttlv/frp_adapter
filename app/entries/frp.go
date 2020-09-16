package entries

type CoreFrp struct {
	FrpServerIpAddress string `json:"frp_server_ip_address"`
	ProxyPort          string `json:"proxy_port"`
	Status             string `json:"status"`
}
