/*
Package main provides entrypoint for the service.
*/
package main

import (
	_ "net/http/pprof"

	"expvar"
	"flag"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sync"

	"github.com/garukun/golgtm/http/certs"
	"github.com/garukun/golgtm/lgtm"
)

var (
	port             = flag.Int("port", 8080, "Port on which the service will run")
	debugPort        = flag.Int("debugport", -1, "Port to which the service will expose debug information")
	blockProfileRate = flag.Int("blockprofilerate", 0, "Rate at which the profiler profiles for blocking contentions; see 'go doc runtime.SetBlockProfileRate'.")
)

var (
	// revision denotes the git commit revision at which the binary is built. The default value
	// provided here is father of all git revisions, e.g., an empty git tree. It will be replaced to
	// HEAD commit during build time.
	revision = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"
)

func init() {
	flag.Parse()

	exposeBuildInfo()
}

func main() {
	servers := map[string]*http.Server{
		"main": {
			Addr:    fmt.Sprintf(":%d", *port),
			Handler: lgtmHandler(),
		},

		// Add other servers here.
	}

	// Expose debug info via debugPort.
	if *debugPort != -1 {
		runtime.SetBlockProfileRate(*blockProfileRate)

		servers["debug"] = &http.Server{
			Addr:    fmt.Sprintf(":%d", *debugPort),
			Handler: http.DefaultServeMux,
		}
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(servers))

	for n, s := range servers {
		go func(name string, server *http.Server) {
			log.Printf("Starting server %s (rev:%s) on %s...", name, revision, server.Addr)
			log.Fatal(server.ListenAndServe())
			wg.Done()
		}(n, s)
	}

	wg.Wait()
	log.Print("Bye!")
}

// exposeBuildInfo method exposes the build information such as revision via the expvar package.
func exposeBuildInfo() {
	expvar.NewString("rev").Set(revision)
}

func lgtmHandler() http.Handler {
	conf, err := lgtm.ConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	return lgtm.New(certs.DefaultHTTPClient, conf)
}
