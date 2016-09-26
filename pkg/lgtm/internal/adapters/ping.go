package adapters

import "net/http"

// Ping implements httpadapter.Adapter interface and respond to a ping event with
// HTTP 204 No Content.
type Ping struct{}

func (p Ping) Adapt(h http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set(ResponseHeader, "ping")
		resp.WriteHeader(http.StatusNoContent)
	})
}
