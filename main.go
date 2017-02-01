package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/Smile-SA/argoos/apiutils"
)

// signal handling, if server should stop, cleanup goroutines.
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
	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
	c, _ := ioutil.ReadAll(r.Body)
	events := apiutils.GetEvents(c)
	for _, e := range events.Events {
		apiutils.ImpactedDeployments(e)
	}
}

// Health return always "ok" with 200 OK. Usefull to check liveness.
func Health(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
	w.Write([]byte("ok"))
}

func main() {
	host := ":3000"
	var servercert, serverkey string

	if v := os.Getenv("KUBE_MASTER_URL"); len(v) > 0 {
		apiutils.KubeMasterURL = v
	}

	if v := os.Getenv("SKIP_SSL_VERIFICATION"); strings.ToUpper(v) == "TRUE" {
		apiutils.SkipSSLVerification = true
		// Certificates
		if v := os.Getenv("CA_FILE"); len(v) > 0 {
			apiutils.CAFile = v
		}
		if v := os.Getenv("CERT_FILE"); len(v) > 0 {
			apiutils.CertFile = v
		}
		if v := os.Getenv("KEY_FILE"); len(v) > 0 {
			apiutils.KeyFile = v
		}
	}

	if v := os.Getenv("SERVER_CERT"); len(v) > 0 {
		servercert = v
	}

	if v := os.Getenv("SERVER_KEY"); len(v) > 0 {
		serverkey = v
	}

	if v := os.Getenv("LISTEN"); len(v) > 0 {
		host = v
	}

	flag.StringVar(&apiutils.KubeMasterURL,
		"master",
		apiutils.KubeMasterURL,
		"Kube master scheme://host:port")
	flag.BoolVar(&apiutils.SkipSSLVerification,
		"skip-ssl-verification",
		apiutils.SkipSSLVerification,
		"Skip SSL verification for kubernetes api")
	flag.StringVar(&host,
		"listen",
		host,
		"Listen interface, could be host:port, or :port")

	// certs
	flag.StringVar(&apiutils.CAFile,
		"ca-file",
		apiutils.CAFile,
		"Certificate Authority certificate file path (only if using https to contact kubernetes api)")
	flag.StringVar(&apiutils.CertFile,
		"cert-file",
		apiutils.CertFile,
		"Client certificate file path (client authentication only)")
	flag.StringVar(&apiutils.KeyFile,
		"key-file",
		apiutils.KeyFile,
		"Client private key file path (client authentication only)")

	// argoos can serve https
	flag.StringVar(&servercert,
		"server-cert",
		servercert,
		"Server certificate to serve SSL")

	flag.StringVar(&serverkey,
		"server-key",
		serverkey,
		"Server key to server SSL")

	flag.Parse()

	go sig()
	apiutils.StartRollout()

	log.Println("Starting")

	http.HandleFunc("/healthz", Health)
	http.HandleFunc("/event", Action)

	if len(serverkey) > 0 && len(servercert) > 0 {
		log.Fatal(http.ListenAndServeTLS(host, servercert, serverkey, nil))
	} else {
		log.Fatal(http.ListenAndServe(host, nil))
	}

}
