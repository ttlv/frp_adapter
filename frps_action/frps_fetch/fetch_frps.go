package frps_fetch

import (
	"encoding/base64"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/ttlv/common_utils/utils"
	"github.com/ttlv/frp_adapter/frp_adapter_init"
	"github.com/ttlv/frp_adapter/model"
)

func FetchFromFrps() (frpServers []model.FrpServer, err error) {
	var (
		authorization = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v:%v", frp_adapter_init.FrpsConfig.HttpAuthUserName, frp_adapter_init.FrpsConfig.HttpAuthPassword)))
		resp          string
		headers       = make(map[string]string)
	)
	headers["Authorization"] = fmt.Sprintf("Basic %v", authorization)
	if resp, err = utils.Get(frp_adapter_init.FrpsConfig.Api, nil, headers); err != nil {
		return
	}
	// Frps重启或者是当前没有frpc连入frps
	if resp == `{"proxies":[]}` {
		err = fmt.Errorf("当前无Frpc节点")
		return
	}
	gjson.Get(resp, "proxies").ForEach(func(key, value gjson.Result) bool {
		var frpServer model.FrpServer
		if value.Get("public_ip_address").String() != "" {
			frpServer.PublicIpAddress = value.Get("public_ip_address").String()
		}
		if value.Get("status").String() != "" {
			frpServer.Status = value.Get("status").String()
		}
		if value.Get("unique_id").String() != "" {
			frpServer.UniqueID = value.Get("unique_id").String()
		}
		if value.Get("conf.remote_port").String() != "" {
			frpServer.Port = value.Get("conf.remote_port").String()
		}
		if value.Get("mac_address").String() != "" {
			frpServer.MacAddress = value.Get("mac_address").String()
		}
		frpServers = append(frpServers, frpServer)
		return true
	})
	return
}
