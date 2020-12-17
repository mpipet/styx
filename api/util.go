package api

import (
	"encoding/json"
	"io"
	"net/http"
)

func WriteResponse(w http.ResponseWriter, statusCode int, v interface{}) {

	if v == nil {
		w.WriteHeader(statusCode)
		return
	}

	bytes, err := MarshalJson(v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")

	w.WriteHeader(statusCode)
	w.Write(bytes)
}

func ReadResponse(r io.Reader, v interface{}) {

	dec := json.NewDecoder(r)

	err := dec.Decode(v)
	if err != nil {
		return
	}
}

func WriteError(w http.ResponseWriter, statusCode int, v interface{}) {

	WriteResponse(w, statusCode, v)
}

func MarshalJson(v interface{}) (bytes []byte, err error) {

	bytes, err = json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}

	bytes = append(bytes, byte("\n"[0]))

	return bytes, nil
}
