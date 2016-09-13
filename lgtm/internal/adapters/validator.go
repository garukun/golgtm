package adapters

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
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

	payload *bytes.Buffer
}

func (v *Validator) validate(body []byte, signature string) error {
	mac := hmac.New(sha1.New, v.Secret)
	mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))
	if sig != signature {
		return fmt.Errorf("Invalid signature from Github: %s", signature)
	}

	return nil
}

func (v *Validator) Adapt(h http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			log.Print("Must be HTTP POST method")
			resp.Header().Set(responseHeader, "not post")
			resp.WriteHeader(http.StatusNoContent)
			return
		}

		signature := req.Header.Get(githubSigHeader)
		v.payload = bytes.NewBuffer(nil)
		downstream := ioutil.NopCloser(io.TeeReader(req.Body, v.payload))

		// Skip the first 5 characters because it's used to indicate the hash mechanism, e.g. "sha1:".
		signature = signature[5:]
		if err := v.validate(v.payload.Bytes(), signature); err != nil {
			log.Print(err)
			resp.Header().Set(responseHeader, "naughty hacker")
			resp.WriteHeader(http.StatusBadRequest)
			return
		}

		req.Body = downstream

		h.ServeHTTP(resp, req)
	})
}
