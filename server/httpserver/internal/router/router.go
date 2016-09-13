package router

import (
	"net/http"

	"github.com/garukun/golgtm/certs"
	"github.com/garukun/golgtm/lgtm"
)

var DefaultRouter http.Handler = lgtm.New(certs.DefaultHTTPClient)
