package dnssd

import (
	"bytes"
	"unsafe"
)

// ResolveCallbackFunc is called when a service is resolved or an error occurs.
type ResolveCallbackFunc func(op *ResolveOp, err error, host string, port int, txt map[string]string)

// ResolveOp represents an operation that resolves a service instance to a host, port and TXT map containing meta data.
type ResolveOp struct {
	baseOp
	name     string
	stype    string
	domain   string
	callback ResolveCallbackFunc
}

// NewResolveOp creates a new ResolveOp with the associated parameters set.
// It should be called with the parameters supplied to the callback of a browse operation.
func NewResolveOp(interfaceIndex int, name, serviceType, domain string, f ResolveCallbackFunc) *ResolveOp {
	op := &ResolveOp{}
	op.SetInterfaceIndex(interfaceIndex)
	op.SetName(name)
	op.SetType(serviceType)
	op.SetDomain(domain)
	op.SetCallback(f)
	return op
}

// StartResolveOp returns the equivalent of calling NewResolveOp and Start.
func StartResolveOp(interfaceIndex int, name, serviceType, domain string, f ResolveCallbackFunc) (*ResolveOp, error) {
	op := NewResolveOp(interfaceIndex, name, serviceType, domain, f)
	return op, op.Start()
}

// Name returns the name of the service.
func (o *ResolveOp) Name() string {
	o.m.Lock()
	defer o.m.Unlock()
	return o.name
}

// SetName set's the name of the service.
func (o *ResolveOp) SetName(n string) error {
	o.m.Lock()
	defer o.m.Unlock()
	if o.started {
		return ErrStarted
	}
	o.name = n
	return nil
}

// Type returns the service type associated with the op.
func (o *ResolveOp) Type() string {
	o.m.Lock()
	defer o.m.Unlock()
	return o.stype
}

// SetType sets the service type associated with the op.
func (o *ResolveOp) SetType(s string) error {
	o.m.Lock()
	defer o.m.Unlock()
	if o.started {
		return ErrStarted
	}
	o.stype = s
	return nil
}

// Domain returns the domain associated with the op.
func (o *ResolveOp) Domain() string {
	o.m.Lock()
	defer o.m.Unlock()
	return o.domain
}

// SetDomain sets the domain associated with the op.
func (o *ResolveOp) SetDomain(s string) error {
	o.m.Lock()
	defer o.m.Unlock()
	if o.started {
		return ErrStarted
	}
	o.domain = s
	return nil
}

// SetCallback sets the function to call when a service is resolved or an error occurs.
func (o *ResolveOp) SetCallback(f ResolveCallbackFunc) error {
	o.m.Lock()
	defer o.m.Unlock()
	if o.started {
		return ErrStarted
	}
	o.callback = f
	return nil
}

// Start begins the resolve operation. Resolve operations should be stopped as soon as they are no longer needed.
func (o *ResolveOp) Start() error {
	o.m.Lock()
	defer o.m.Unlock()
	if o.started {
		return ErrStarted
	}
	if o.callback == nil {
		return ErrMissingCallback
	}
	err := pollServer.startOp(o)
	o.started = err == nil || err == ErrStarted
	return err
}

func (o *ResolveOp) init(sharedref uintptr) (ref uintptr, err error) {
	ref = sharedref
	o.setFlag(_FlagsShareConnection, ref != 0)
	if err = resolveStart(&ref, o.flags, o.interfaceIndexC(), o.name, o.stype, o.domain, unsafe.Pointer(o)); err != nil {
		ref = 0
	}
	return
}

// Stop stops the operation.
func (o *ResolveOp) Stop() {
	o.m.Lock()
	defer o.m.Unlock()
	if !o.started {
		return
	}
	o.started = false
	pollServer.stopOp(o)
}

func (o *ResolveOp) handleError(e error) {
	if !o.started {
		return
	}
	o.started = false
	pollServer.removePollOp(o)
	queueCallback(func() { o.callback(o, e, "", 0, nil) })
}

func dnssdResolveCallback(sdRef unsafe.Pointer, flags, interfaceIndex uint32, err int32, fullname, hosttarget unsafe.Pointer, port uint16, txtLen uint16, txtRecord, ctx unsafe.Pointer) {
	o := (*ResolveOp)(ctx)
	if e := getError(err); e != nil {
		o.handleError(e)
	} else {
		h := cStringToString(hosttarget)
		p := int(port)
		var txtBytes []byte
		if txtLen > 0 && txtRecord != nil {
			txtBytes = (*[65535]byte)(txtRecord)[:txtLen]
		}
		txt := decodeTxt(txtBytes)
		queueCallback(func() { o.callback(o, e, h, p, txt) })
	}
}

func decodeTxt(txt []byte) map[string]string {
	m := make(map[string]string)
	for offset := 0; offset < len(txt); {
		start, end := offset+1, offset+1+int(txt[offset])
		if end <= len(txt) && start != end {
			s := txt[start:end]
			if i := bytes.IndexByte(s, '='); i > 0 {
				m[string(s[:i])] = string(s[i+1:])
			} else {
				m[string(s)] = ""
			}
		}
		offset = end
	}
	return m
}
