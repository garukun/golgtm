package adapters

import (
	"log"
	"net/http"

	"github.com/garukun/golgtm/http/httpadapter"
)

// EventRouter routes a GitHub webhook event to a matching httpadapter.Adapter.
//
// See all possible events: https://developer.github.com/webhooks/#events
type EventRouter struct {
	Events map[string]httpadapter.Adapter // key: GitHub event.
}

func (r *EventRouter) Adapt(h http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		eventType := req.Header.Get(GithubEventHeader)
		a, ok := r.Events[eventType]
		log.Printf("Event: %s, %t", eventType, ok)

		if ok {
			h = a.Adapt(h)
		}

		h.ServeHTTP(resp, req)
	})
}
