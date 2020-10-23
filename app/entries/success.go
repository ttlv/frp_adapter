package entries

// http请求成功的统一数据结构
type Success struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}
