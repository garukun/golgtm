package httpadapter

import "net/http"

// Adapter takes an http.Handler as input adapts it and output another http.Handler which can be
// handled by an HTTP server.
//
// A common use case of the Adapter can be applying gzip compression, adding additional header, etc.
type Adapter interface {
	Adapt(h http.Handler) http.Handler
}
