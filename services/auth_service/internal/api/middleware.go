package api

import "net/http"

type ResponceWriter struct {
	http.ResponseWriter
	code int
}

func (rw *ResponceWriter) WriteHeader(code int) {
	rw.code = code
	rw.ResponseWriter.WriteHeader(code)
}
