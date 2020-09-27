package entries

// http请求成功的统一数据结构
type Success struct {
	Data interface{} `json:"data"`
}
