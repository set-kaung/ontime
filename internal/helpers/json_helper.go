package helpers

import (
	"encoding/json"
	"net/http"
)

type jsonResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, data jsonResponse, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}
	js = append(js, '\n')
	for key, value := range headers {
		w.Header()[key] = value
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

func WriteError(w http.ResponseWriter, status int, errorMessage string, headers http.Header) error {
	js := jsonResponse{}
	js.Status = "error"
	js.Message = errorMessage
	return writeJSON(w, status, js, headers)
}

func WriteData(w http.ResponseWriter, status int, data interface{}, headers http.Header) error {
	js := jsonResponse{}
	js.Status = "success"
	js.Data = data
	return writeJSON(w, status, js, headers)
}

func WriteSuccess(w http.ResponseWriter, status int, message string, headers http.Header) error {
	js := jsonResponse{}
	js.Status = "success"
	return writeJSON(w, status, js, headers)
}
