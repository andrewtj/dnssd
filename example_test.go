package dnssd_test

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/andrewtj/dnssd"
)

func ExampleRegisterCallbackFunc(op *dnssd.RegisterOp, err error, add bool, name, serviceType, domain string) {
	if err != nil {
		// op is now inactive
		log.Printf("Service registration failed: %s", err)
		return
	}
	if add {
		log.Printf("Service registered as “%s“ in %s", name, domain)
	} else {
		log.Printf("Service “%s” removed from %s", name, domain)
	}
}
func ExampleRegisterOp() {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Printf("Listen failed: %s", err)
		return
	}
	port := listener.Addr().(*net.TCPAddr).Port

	op, err := dnssd.StartRegisterOp("", "_http._tcp", port, ExampleRegisterCallbackFunc)
	if err != nil {
		log.Printf("Failed to register service: %s", err)
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %s", r.RemoteAddr)
	})
	http.Serve(listener, nil)

	// later...
	op.Stop()
}

func ExampleRegisterOp_proxy() {
	op := dnssd.NewProxyRegisterOp("dnssd godoc", "_http._tcp", "godoc.org", 80, ExampleRegisterCallbackFunc)
	if err := op.SetTXTPair("path", "/github.com/andrewtj/dnssd"); err != nil {
		log.Printf("Failed to set key-value pair: %s", err)
		return
	}
	if err := op.Start(); err != nil {
		log.Printf("Failed to register service: %s", err)
		return
	}
	// later...
	op.Stop()
}

func ExampleBrowseCallbackFunc(op *dnssd.BrowseOp, err error, add bool, interfaceIndex int, name string, serviceType string, domain string) {
	if err != nil {
		// op is now inactive
		log.Printf("Browse operation failed: %s", err)
		return
	}
	change := "lost"
	if add {
		change = "found"
	}
	log.Printf("Browse operation %s %s service “%s” in %s on interface %d", change, serviceType, name, domain, interfaceIndex)
}

func ExampleBrowseOp() {
	op, err := dnssd.StartBrowseOp("_http._tcp", ExampleBrowseCallbackFunc)
	if err != nil {
		// op is now inactive
		log.Printf("Browse operation failed: %s", err)
		return
	}
	// later...
	op.Stop()
}

func ExampleBrowseOp_domain() {
	op := dnssd.NewBrowseOp("_http._tcp", ExampleBrowseCallbackFunc)
	op.SetDomain("dns-sd.org")
	if err := op.Start(); err != nil {
		log.Printf("Failed to start browse operation: %s", err)
		return
	}
	// later...
	op.Stop()
}

func ExampleResolveCallbackFunc(op *dnssd.ResolveOp, err error, host string, port int, txt map[string]string) {
	if err != nil {
		// op is now inactive
		log.Printf("Resolve operation failed: %s", err)
		return
	}
	log.Printf("Resolved service to host %s port %d with meta info: %v", host, port, txt)
}

func ExampleResolveOp() {
	op, err := dnssd.StartResolveOp(0, " * DNS Service Discovery", "_http._tcp", "dns-sd.org", ExampleResolveCallbackFunc)
	if err != nil {
		log.Printf("Failed to start resolve operation: %s", err)
		return
	}
	// later...
	op.Stop()
}
