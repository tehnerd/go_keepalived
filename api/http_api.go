package api

import (
	"encoding/json"
	"net/http"
)

func makeChanHandler(fn func(w http.ResponseWriter, r *http.Request, a, b chan APIMsg), a, b chan APIMsg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, a, b)
	}
}

func handler(w http.ResponseWriter, r *http.Request, requestChan,
	responseChan chan APIMsg) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	requestStruct := make(map[string]string)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&requestStruct)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if _, exists := requestStruct["Command"]; !exists {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Command argument is required"))
		return
	}
	requestChan <- APIMsg{Data: &requestStruct}
	responseStruct := <-responseChan
	resp, err := json.Marshal(responseStruct)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("cant marshal json response"))
		return
	}
	w.Write(resp)
}

//TODO: https instead of http
func StartHTTPApi(requestChan, responseChan chan APIMsg) {
	http.HandleFunc("/", makeChanHandler(handler, requestChan, responseChan))
	http.ListenAndServe("localhost:62307", nil)
}
