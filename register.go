package dnssd

import (
	"os"
	"unsafe"
)

// RegisterCallbackFunc is called when a name is registered or deregistered in a given domain, or when an error occurs.
type RegisterCallbackFunc func(op *RegisterOp, err error, add bool, name, serviceType, domain string)

// RegisterOp represents a service registration operation.
type RegisterOp struct {
	baseOp
	name   string
	stype  string
	domain string
	host   string
	port   int
	txt    struct {
		l int
		m map[string]string
	}
	callback RegisterCallbackFunc
	seenAdd  bool
}

// NewRegisterOp creates a new RegisterOp with the given parameters set.
func NewRegisterOp(name, serviceType string, port int, f RegisterCallbackFunc) *RegisterOp {
	op := &RegisterOp{}
	op.SetName(name)
	op.SetType(serviceType)
	op.SetPort(port)
	op.SetCallback(f)
	return op
}

// StartRegisterOp returns the equivalent of calling NewRegisterOp and Start().
func StartRegisterOp(name, serviceType string, port int, f RegisterCallbackFunc) (*RegisterOp, error) {
	op := NewRegisterOp(name, serviceType, port, f)
	return op, op.Start()
}

// NewProxyRegisterOp creates a new RegisterOp with the given parameters set.
func NewProxyRegisterOp(name, serviceType, host string, port int, f RegisterCallbackFunc) *RegisterOp {
	op := NewRegisterOp(name, serviceType, port, f)
	op.SetHost(host)
	return op
}

// StartProxyRegisterOp returns the equivalent of calling NewProxyRegisterOp and Start().
func StartProxyRegisterOp(name, serviceType, host string, port int, f RegisterCallbackFunc) (*RegisterOp, error) {
	op := NewProxyRegisterOp(name, serviceType, host, port, f)
	return op, op.Start()
}

// Name returns the name of the service.
func (o *RegisterOp) Name() string {
	o.m.Lock()
	defer o.m.Unlock()
	return o.name
}

// SetName sets the name of the service. A service name can not exceed 63 bytes.
func (o *RegisterOp) SetName(n string) error {
	o.m.Lock()
	defer o.m.Unlock()
	if o.started {
		return ErrStarted
	}
	o.name = n
	return nil
}

// Type returns the service type associated with the op.
func (o *RegisterOp) Type() string {
	o.m.Lock()
	defer o.m.Unlock()
	return o.stype
}

// SetType sets the service type associated with the op.
func (o *RegisterOp) SetType(s string) error {
	o.m.Lock()
	defer o.m.Unlock()
	if o.started {
		return ErrStarted
	}
	o.stype = s
	return nil
}

// Domain returns the domain associated with the op.
func (o *RegisterOp) Domain() string {
	o.m.Lock()
	defer o.m.Unlock()
	return o.domain
}

// SetDomain sets the domain associated with the op.
func (o *RegisterOp) SetDomain(s string) error {
	o.m.Lock()
	defer o.m.Unlock()
	if o.started {
		return ErrStarted
	}
	o.domain = s
	return nil
}

// Host returns the hostname of the service. An empty string will result in the local machine's hostname being used.
func (o *RegisterOp) Host() string {
	o.m.Lock()
	defer o.m.Unlock()
	return o.host
}

// SetHost sets the hostname of the service. An empty string will result in the local machine's hostname being used.
func (o *RegisterOp) SetHost(h string) error {
	o.m.Lock()
	defer o.m.Unlock()
	if o.started {
		return ErrStarted
	}
	o.host = h
	return nil
}

// Port returns the port the service is available from.
func (o *RegisterOp) Port() int {
	o.m.Lock()
	defer o.m.Unlock()
	return o.port
}

// SetPort sets the port the service is available from.
func (o *RegisterOp) SetPort(p int) error {
	o.m.Lock()
	defer o.m.Unlock()
	if o.started {
		return ErrStarted
	}
	o.port = p
	return nil
}

// SetTXTPair creates or updates a TXT string with the provided value.
func (o *RegisterOp) SetTXTPair(key, value string) error {
	o.m.Lock()
	defer o.m.Unlock()
	if o.started {
		return ErrStarted
	}
	if o.txt.m == nil {
		o.txt.m = make(map[string]string)
	}
	slen := len(key) + len(value) + 2
	if slen > 255 {
		return ErrTXTStringLen
	}
	oldslen := 0
	if oldvalue, exists := o.txt.m[key]; exists {
		oldslen = len(key) + len(oldvalue) + 2
	}
	newtlen := o.txt.l - oldslen + slen
	if newtlen > 65535 {
		return ErrTXTLen
	}
	o.txt.l = newtlen
	o.txt.m[key] = value
	return nil
}

// DeleteTXTPair deletes the TXT string with the provided key.
func (o *RegisterOp) DeleteTXTPair(key string) error {
	o.m.Lock()
	defer o.m.Unlock()
	if o.started {
		return ErrStarted
	}
	if s, e := o.txt.m[key]; e {
		o.txt.l = o.txt.l - len(key) - len(s) - 2
		delete(o.txt.m, key)
	}
	return nil
}

// SetCallback sets the function to call when a name is registered or deregistered in a given domain, or when an error occurs.
func (o *RegisterOp) SetCallback(f RegisterCallbackFunc) error {
	o.m.Lock()
	defer o.m.Unlock()
	if o.started {
		return ErrStarted
	}
	o.callback = f
	return nil
}

// NoAutoRename indicates how service-name conflicts will be handled.
func (o *RegisterOp) NoAutoRename() bool {
	o.m.Lock()
	defer o.m.Unlock()
	return o.flags&_FlagsNoAutoRename != 0
}

// SetNoAutoRename sets how service-name conflicts will be handled.
// If set to the default, false, conflicts will be handled automatically be renaming the service (eg: "My Service" will be become "My Service 2" or similar).
// If set to true the operations callback will be invoked with an error.
func (o *RegisterOp) SetNoAutoRename(e bool) error {
	o.m.Lock()
	defer o.m.Unlock()
	if o.started {
		return ErrStarted
	}
	o.setFlag(_FlagsNoAutoRename, e)
	return nil
}

// Start begins advertising the service.
func (o *RegisterOp) Start() error {
	o.m.Lock()
	defer o.m.Unlock()
	if o.started {
		return ErrStarted
	}
	o.seenAdd = false
	if o.callback == nil {
		return ErrMissingCallback
	}
	err := pollServer.startOp(o)
	o.started = err == nil || err == ErrStarted
	return err
}

func (o *RegisterOp) init(sharedref uintptr) (ref uintptr, err error) {
	ref = sharedref
	o.setFlag(_FlagsShareConnection, ref != 0)
	txt := make([]byte, 0, o.txt.l)
	for k, v := range o.txt.m {
		s := k + "=" + v
		txt = append(txt, byte(len(s)))
		txt = append(txt, s...)
	}
	err = registerStart(&ref, o.flags, o.interfaceIndexC(), o.name, o.stype, o.domain, o.host, o.port, txt, unsafe.Pointer(o))
	// Avahi's Bonjour compatibility layer doesn't substitute the system's
	// name in place of an empty service name string.
	if err == ErrBadParam && o.name == "" {
		ref = sharedref
		hostname, _ := os.Hostname()
		err = registerStart(&ref, o.flags, o.interfaceIndexC(), hostname, o.stype, o.domain, o.host, o.port, txt, unsafe.Pointer(o))
	}
	if err != nil {
		ref = 0
	}
	return
}

// Stop stops the operation.
func (o *RegisterOp) Stop() {
	o.m.Lock()
	defer o.m.Unlock()
	if !o.started {
		return
	}
	o.started = false
	pollServer.stopOp(o)
}

func (o *RegisterOp) handleError(e error) {
	if !o.started {
		return
	}
	o.started = false
	pollServer.removePollOp(o)
	queueCallback(func() { o.callback(o, e, false, "", "", "") })
}

func dnssdRegisterCallback(sdRef unsafe.Pointer, flags uint32, err int32, name, regtype, domain, ctx unsafe.Pointer) {
	o := (*RegisterOp)(ctx)
	if e := getError(err); e != nil {
		o.handleError(e)
	} else {
		a := flags&_FlagsAdd != 0
		// Avahi's Bonjour compatibility layer doesn't set kDNSServiceFlagsAdd,
		// so if a remove callback occurs before an add has been seen, pretend
		// it's an add. This should do the right-thing since Avahi only supports
		// registration in ".local".
		if !a && !o.seenAdd {
			a = true
		}
		if a && !o.seenAdd {
			o.seenAdd = a
		}
		n := cStringToString(name)
		r := cStringToString(regtype)
		d := cStringToString(domain)
		queueCallback(func() { o.callback(o, e, a, n, r, d) })
	}
}
