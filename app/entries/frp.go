package entries

type Frp struct {
	FrpServerIpAddress string `json:"frp_server_ip_address"`
	UniqueID           string `json:"unique_id"`
	Port               string `json:"port"`
	Status             string `json:"status"`
}
