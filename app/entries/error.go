package entries

type Error struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}
