// Package dnssd implements a wrapper for Apple's C DNS Service Discovery API.
//
// The DNS Service Discovery API is part of the Apple Bonjour zero
// configuration networking stack. The API allows for network services to be
// registered, browsed and resolved without configuration via multicast DNS
// in the ".local" domain and with additional configuration in unicast DNS
// domains. A service consists of a name, type, host, port and a set of
// key-value pairs containing meta information.
//
// Bonjour is bundled with OS X and available for Windows via Bonjour Print
// Services for Windows¹, the Bonjour SDK for Windows² or bundled with iTunes.
// For other POSIX platforms Apple offer mDNSResponder³ as open-source, however
// the Avahi⁴ project is the de facto choice on most Linux and BSD systems.
// Although Avahi has a different API, it does offer a compatibility shim which
// covers a subset of the DNS Service Discovery API, and which this package
// largely sticks to.
//
//  1. http://support.apple.com/kb/dl999
//  2. https://developer.apple.com/bonjour/
//  3. http://opensource.apple.com/tarballs/mDNSResponder/
//  4. http://www.avahi.org/
//
// The DNS Service Discovery API is wrapped as follows:
//
//  DNSServiceRegister()    -> RegisterOp
//  DNSServiceBrowse()      -> BrowseOp
//  DNSServiceResolve()     -> ResolveOp
//  DNSServiceQueryRecord() -> QueryOp
//
// All operations require a callback be set. RegisterOp, BrowseOp and ResolveOp
// require a service type be set. QueryOp requires name, class and type be set.
// If an InterfaceIndex is not set the default value of InterfaceIndexAny is
// used which applies the operation to all network interfaces. For operations
// that take a domain, if no domain is set or the domain is set to an empty
// string the operation applies to all applicable DNS-SD domains.
//
// If a service is registered with an empty string as it's name, the local
// computer name (or hostname) will be substitued. If no host is specified a
// hostname for the local machine will be used. By default services will be
// renamed with a numeric suffix if a name collision occurs.
//
// Callbacks are executed in serial. If an error is supplied to a callback
// the operation will no longer be active and other arguments must be ignored.
//
package dnssd

import (
	"sync"
	"unsafe"
)

// InterfaceIndexAny is the default for all operations.
const InterfaceIndexAny = 0

// InterfaceIndexLocalOnly limits the scope of the operation to the local machine.
const InterfaceIndexLocalOnly = int(^uint(0) >> 1)

const (
	_FlagsAdd             uint32 = 0x2
	_FlagsNoAutoRename           = 0x8
	_FlagsShareConnection        = 0x4000
)

type baseOp struct {
	m              sync.Mutex
	shared         bool
	started        bool
	interfaceIndex int
	flags          uint32
}

var callbackQueueState struct {
	sync.Mutex
	c chan bool
	f []func()
}

func queueCallback(f func()) {
	callbackQueueState.Lock()
	defer callbackQueueState.Unlock()
	if callbackQueueState.c == nil {
		callbackQueueState.c = make(chan bool, 1)
		go callbackQueueLoop()
	}
	callbackQueueState.f = append(callbackQueueState.f, f)
	select {
	case callbackQueueState.c <- true:
	default:
	}
}

func callbackQueueLoop() {
	for {
		_ = <-callbackQueueState.c
		callbackQueueState.Lock()
		f := callbackQueueState.f
		callbackQueueState.f = nil
		callbackQueueState.Unlock()
		for i := range f {
			f[i]()
		}
	}
}

func (o *baseOp) setFlag(flag uint32, enabled bool) {
	set := o.flags&flag != 0
	if set != enabled {
		o.flags ^= flag
	}
}

// Active indicates whether an operation is active
func (o *baseOp) Active() bool {
	o.m.Lock()
	defer o.m.Unlock()
	return o.started
}

// InterfaceIndex returns the interface index the op is tied to.
func (o *baseOp) InterfaceIndex() int {
	o.m.Lock()
	defer o.m.Unlock()
	return o.interfaceIndex
}

// SetInterfaceIndex sets the interface index the op is tied to.
func (o *baseOp) SetInterfaceIndex(i int) error {
	o.m.Lock()
	defer o.m.Unlock()
	if o.started {
		return ErrStarted
	}
	o.interfaceIndex = i
	return nil
}

func (o *baseOp) interfaceIndexC() uint32 {
	if o.interfaceIndex == InterfaceIndexLocalOnly {
		return ^uint32(0)
	}
	return uint32(o.interfaceIndex)
}

func (o *baseOp) init(sharedref uintptr) (ref uintptr, err error) {
	panic("unreachable")
}

func cStringToString(c unsafe.Pointer) string {
	if c == nil {
		return ""
	}
	const maxlen = 1009 // See dns_sd.h's kDNSServiceMaxDomainName for commentary on this size
	s := (*[maxlen]byte)(c)
	for i := range s {
		if s[i] == 0 {
			return string(s[:i])
		}
	}
	panic("unreachable")
}

func deallocateRef(ref *uintptr) {
	if *ref != 0 {
		platformDeallocateRef(ref)
		*ref = 0
	}
}
