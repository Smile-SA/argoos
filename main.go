package main

import (
	"argoos/apiutils"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
)

func sig() {
	c := make(chan os.Signal, 0)
	signal.Notify(c, os.Interrupt)

	// Block until a signal is received.
	s := <-c
	apiutils.StopRollout()
	log.Println("Got signal", s)
	os.Exit(0)
}

// Action is sent each time the registry sends an event.
func Action(w http.ResponseWriter, r *http.Request) {
	c, _ := ioutil.ReadAll(r.Body)
	events := apiutils.GetEvents(c)
	for _, e := range events.Events {
		apiutils.ImpactedDeployments(e)
	}
}

func Health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func main() {
	if v := os.Getenv("KUBE_MASTER_URL"); len(v) > 0 {
		apiutils.KubeMasterURL = v
	}

	if v := os.Getenv("SKIP_SSL_VERIFICATION"); strings.ToUpper(v) == "TRUE" {
		apiutils.SkipSSLVerification = true
	}

	flag.StringVar(&apiutils.KubeMasterURL,
		"master",
		apiutils.KubeMasterURL,
		"Kube master host:port")
	flag.BoolVar(&apiutils.SkipSSLVerification,
		"skip-ssl-verification",
		apiutils.SkipSSLVerification,
		"Skip SSL verification for kubernetes api")
	flag.Parse()

	go sig()
	apiutils.StartRollout()

	log.Println("Starting")
	http.HandleFunc("/healthz", Health)
	http.HandleFunc("/event", Action)
	log.Fatal(http.ListenAndServe(":3000", nil))

}
