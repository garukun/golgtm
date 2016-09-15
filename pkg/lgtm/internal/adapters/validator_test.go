package adapters_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/garukun/golgtm/pkg/lgtm/internal/adapters"
)

func TestValidateSignature(t *testing.T) {
	const customSuccessStatus = 529 // Some unique non-standard HTTP status code

	tests := []struct {
		body      string
		secret    string
		signature string
		status    int
	}{
		// Successful validation
		{
			body:      "There is no spoon",
			secret:    "matrix",
			signature: "sha1:041711d156ab84e80e9ef409de159d64b6a7b04d", // SHA1 of "There is no spoon".
			status:    customSuccessStatus,
		},
		// Wrong secret
		{
			body:      "There is no spoon",
			secret:    "MATRIX",
			signature: "sha1:041711d156ab84e80e9ef409de159d64b6a7b04d", // SHA1 of "There is no spoon".
			status:    http.StatusBadRequest,
		},
	}

	// Test http.Handler implementation to verify the actual HTTP handling.
	testHandler := func(expectedBody string) http.Handler {
		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			body := bytes.NewBuffer(nil)
			io.Copy(body, req.Body)

			if body.String() != expectedBody {
				t.Errorf("Expected test handler to still process original body of `%s`\ninstead of `%s`.", expectedBody, body.String())
			}

			resp.WriteHeader(customSuccessStatus)
		})
	}

	for i, test := range tests {
		t.Logf("Testing %d...", i)
		bodyReader := bytes.NewReader([]byte(test.body))

		req := httptest.NewRequest(http.MethodPost, "http://localhost", bodyReader)
		req.Header.Set(adapters.GithubSigHeader, test.signature)
		resp := httptest.NewRecorder()

		v := &adapters.Validator{Secret: []byte(test.secret)}
		h := v.Adapt(testHandler(test.body))
		h.ServeHTTP(resp, req)

		if resp.Code != test.status {
			t.Log("Response header:", resp.HeaderMap)
			t.Errorf("Expected response status code of %d instead of %d.", test.status, resp.Code)
		}
	}
}
