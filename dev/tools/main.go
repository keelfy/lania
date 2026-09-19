// Command tools runs the small stand-ins the local environment needs instead of Docker services.
//
//	tools redis    <addr>  in-memory Redis
//	tools imgproxy <addr>  imgproxy replacement that redirects to the original image
package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/alicebob/miniredis/v2"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: %s redis|imgproxy <addr>", os.Args[0])
	}
	addr := os.Args[2]

	switch os.Args[1] {
	case "redis":
		runRedis(addr)
	case "imgproxy":
		runImgproxy(addr)
	default:
		log.Fatalf("unknown command %q", os.Args[1])
	}
}

func runRedis(addr string) {
	server := miniredis.NewMiniRedis()
	if err := server.StartAddr(addr); err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	log.Printf("redis listening on %s", addr)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
}

// runImgproxy answers /unsafe/<processing options>/plain/<source url>[@ext] with a redirect to the source.
func runImgproxy(addr string) {
	// A bare handler is used because ServeMux would collapse the "//" of the source url.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, source, found := strings.Cut(r.URL.EscapedPath(), "/plain/")
		if !found {
			http.Error(w, "only plain source urls are supported", http.StatusBadRequest)
			return
		}
		source, err := url.PathUnescape(source)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if i := strings.LastIndex(source, "@"); i > strings.LastIndex(source, "/") {
			source = source[:i]
		}
		if r.URL.RawQuery != "" {
			source += "?" + r.URL.RawQuery
		}
		if strings.HasPrefix(source, "/") {
			source = "http://localhost:3000" + source
		}
		http.Redirect(w, r, source, http.StatusFound)
	})

	log.Printf("imgproxy stub listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
