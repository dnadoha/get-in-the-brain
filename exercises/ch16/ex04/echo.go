package echo

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

type echoRequest struct {
	Message string `json:"message"`
}

func EchoHandler(w http.ResponseWriter, r *http.Request) {
	var req echoRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Message == "" {
		WriteError(w, http.StatusBadRequest, "message is required")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"echo": req.Message})
}
