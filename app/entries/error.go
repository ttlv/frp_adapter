package entries

// http请求失败的统一数据结构
type Error struct {
	Code  int         `json:"code"`
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Message string `json:"message"`
}
