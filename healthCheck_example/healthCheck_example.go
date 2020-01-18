package main

import (
	"flag"
	"fmt"
	"net/http"
)

var (
	address    = flag.String("listen.address", "0.0.0.0", "assign an IP address for the service to listen on")
	healthcode = flag.Int("healthcode", 0, "pass a healthcode number to test [0 - 6]")
	port       = flag.String("listen.port", "8081", "set the port the service will listen on")
)

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, healthcodeResponse(*healthcode))
}

func main() {
	flag.Parse()
	http.HandleFunc("/healthCheck.html", healthCheckHandler)

	fmt.Println("Beginning to serve health code", *healthcode, "on", *address+":"+*port+"/healthCheck.html")
	http.ListenAndServe(*address+":"+*port, nil)
}

func healthcodeResponse(healthcode int) string {
	switch healthcode {
	case 1:
		return "[{\"healthCode\":1,\"httpCode\":503,\"description\":\"DBConnectionError\"}]\n"
	case 2:
		return "[{\"healthCode\":2,\"httpCode\":200,\"description\":\"SetupAssistant\"}]\n"
	case 3:
		return "[{\"healthCode\":3,\"httpCode\":503,\"description\":\"DBConnectionConfigError\"}]\n"
	case 4:
		return "[{\"healthCode\":4,\"httpCode\":503,\"description\":\"Initializing\"}]\n"
	case 5:
		return "[{\"healthCode\":5,\"httpCode\":503,\"description\":\"ChildNodeStartUpError\"}]\n"
	case 6:
		return "[{\"healthCode\":6,\"httpCode\":503,\"description\":\"InitializationError\"}]\n"
	default:
		return "[]\n"
	}
}
