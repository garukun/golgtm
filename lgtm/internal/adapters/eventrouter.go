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
		eventType := req.Header.Get(githubEventHeader)
		if a, ok := r.Events[eventType]; ok {
			log.Printf("Event router recognized event: %s.", eventType)
			h = a.Adapt(h)
		} else {
			log.Printf("Event router unrecognized event: %s.", eventType)
		}

		h.ServeHTTP(resp, req)
	})
}
