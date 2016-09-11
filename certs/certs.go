package certs

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
)

var DefaultHTTPClient *http.Client

func init() {
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(pemCerts)
	DefaultHTTPClient = &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}}}
}
