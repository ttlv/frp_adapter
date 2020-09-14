package helpers

import (
	"encoding/json"
	"github.com/ttlv/frp_adapter/app/entries"
	"net/http"
)

func RenderFailureJSON(w http.ResponseWriter, code int, message string) {
	result, _ := json.Marshal(entries.Error{
		Error: entries.ErrorDetail{
			Message: message,
			Code:    code,
		},
	})
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func RenderSuccessJSON(w http.ResponseWriter, code int, data interface{}) {
	result, _ := json.Marshal(entries.Success{
		Code: code,
		Data: data,
	})
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}
