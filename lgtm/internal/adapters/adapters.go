package adapters

import (
	"net/http"

	"github.com/garukun/golgtm/http/httpadapter"
)

func Adapt(h http.Handler, a ...httpadapter.Adapter) http.Handler {
	for _, v := range a {
		h = v.Adapt(h)
	}

	return h
}
