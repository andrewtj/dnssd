package dnssd_test

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/andrewtj/dnssd"
	"github.com/miekg/dns"
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

func ExampleQueryCallbackFunc(op *dnssd.QueryOp, err error, add bool, interfaceIndex int, fullname string, rrtype, rrclass uint16, rdata []byte, ttl uint32) {
	if err != nil {
		// op is now inactive
		log.Printf("Query operation failed: %s", err)
		return
	}
	change := "removed"
	if add {
		change = "added"
	}
	log.Printf("Query operation %s %s/%d/%d/%v (TTL: %d) on interface %d", change, fullname, rrtype, rrclass, rdata, ttl, interfaceIndex)
}

func ExampleQueryCallbackFunc_unpackRR(op *dnssd.QueryOp, err error, add bool, interfaceIndex int, fullname string, rrtype, rrclass uint16, rdata []byte, ttl uint32) {
	// Demonstrates constructing a resource record and unpacking it using
	// Miek Gieben's dns package (https://github.com/miekg/dns/).

	if err != nil {
		// op is now inactive
		log.Printf("Query operation failed: %s", err)
		return
	}

	buf := make([]byte, len(fullname)+1+2+2+4+2+len(rdata))
	off, err := dns.PackDomainName(fullname, buf, 0, nil, false)
	if err != nil {
		log.Fatalf("Error packing domain: %s", err)
	}
	buf = buf[:off]
	buf = append(buf, byte(rrtype>>8), byte(rrtype))
	buf = append(buf, byte(rrclass>>8), byte(rrclass))
	buf = append(buf, byte(ttl>>24), byte(ttl>>16), byte(ttl>>8), byte(ttl))
	buf = append(buf, byte(len(rdata)>>8), byte(len(rdata)))
	buf = append(buf, rdata...)
	rr, off, err := dns.UnpackRR(buf, 0)
	if err != nil {
		log.Fatalf("Error unpacking rr: %s", err)
	}

	change := "removed"
	if add {
		change = "added"
	}

	log.Printf("Query operation on interface %d %s:\n%s", interfaceIndex, change, rr.String())
}

func ExampleQueryOp() {
	op := dnssd.NewQueryOp(0, "golang.org.", 1, 1, ExampleQueryCallbackFunc)
	if err := op.Start(); err != nil {
		log.Printf("Failed to start query operation: %s", err)
		return
	}
	// later
	op.Stop()
}
