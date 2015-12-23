/*
Package main provides the entry point of the golgtm binary.

Binary golgtm inspects all the comments from a given PULL_REQUEST; if there are more than N number of LGTM comments in
the list of comments, golgtm will make sure the label APPROVED is attached to the pull request and IN_PROGRESS is
removed.

PULL_REQUEST, N, LGTM, APPROVED, IN_PROGRESS can be configured through the following ways:
*/
package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"log"
	"net/http"
)

var certClient *http.Client

func init() {
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(pemCerts)
	certClient = &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}}}
}

func main() {
	flag.Parse()

	if pr := *flagPR; pr != 0 {
		PR = pr
	}

	lgtm := NewLGTM(certClient)
	if !lgtm.IsApproved() {
		lgtm.Unapprove()
		log.Println("Not done yet!")
	} else {
		lgtm.Approve()
		log.Println("Approved!")
	}
}
