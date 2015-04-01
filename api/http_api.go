package api

import (
	"encoding/json"
	"fmt"
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
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Printf("%#v\n", requestStruct)
	w.Write([]byte("Test"))
}

//TODO: https instead of http
func StartHTTPApi(requestChan, responseChan chan APIMsg) {
	http.HandleFunc("/", makeChanHandler(handler, requestChan, responseChan))
	http.ListenAndServe("localhost:62307", nil)
}
