package utils

import (
	"encoding/json"
	"net/http"
)

func JSONResponse(w http.ResponseWriter, obj interface{}) {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	encoder.Encode(obj)
}
