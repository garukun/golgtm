package adapters

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type Validator struct {
	// Secret against the Github signature on every request.
	Secret []byte
}

func (v *Validator) validate(body io.Reader, signature string) error {
	mac := hmac.New(sha1.New, v.Secret)
	io.Copy(mac, body)

	if sig := hex.EncodeToString(mac.Sum(nil)); sig != signature {
		return fmt.Errorf("Invalid request signature: %s, expecting %s", signature, sig)
	}

	return nil
}

func (v *Validator) Adapt(h http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			log.Print("Must be HTTP POST method")
			resp.Header().Set(ResponseHeader, "not post")
			resp.WriteHeader(http.StatusNoContent)
			return
		}

		signature := req.Header.Get(GithubSigHeader)
		downstream := bytes.NewBuffer(nil)
		payload := io.TeeReader(req.Body, downstream)

		// Skip the first 5 characters because it's used to indicate the hash mechanism, e.g. "sha1:".
		signature = signature[5:]
		if err := v.validate(payload, signature); err != nil {
			log.Print(err)
			log.Printf("HTTP request body: %s", base64.StdEncoding.EncodeToString(downstream.Bytes()))
			resp.Header().Set(ResponseHeader, "naughty hacker")
			resp.WriteHeader(http.StatusBadRequest)
			return
		}

		req.Body = ioutil.NopCloser(downstream)

		h.ServeHTTP(resp, req)
	})
}
