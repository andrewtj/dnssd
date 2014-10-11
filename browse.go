package dnssd

import "unsafe"

// BrowseCallbackFunc is called when an error occurs or a service is lost or found.
type BrowseCallbackFunc func(op *BrowseOp, err error, add bool, interfaceIndex int, name string, serviceType string, domain string)

// BrowseOp represents a query for services of a particular type.
type BrowseOp struct {
	baseOp
	stype    string
	domain   string
	callback BrowseCallbackFunc
}

// NewBrowseOp creates a new BrowseOp with the given service type and call back set.
func NewBrowseOp(serviceType string, f BrowseCallbackFunc) *BrowseOp {
	op := &BrowseOp{}
	op.SetType(serviceType)
	op.SetCallback(f)
	return op
}

// StartBrowseOp returns the equivalent of calling NewBrowseOp and Start().
func StartBrowseOp(serviceType string, f BrowseCallbackFunc) (*BrowseOp, error) {
	op := NewBrowseOp(serviceType, f)
	return op, op.Start()
}

// Type returns the service type associated with the op.
func (o *BrowseOp) Type() string {
	o.m.Lock()
	defer o.m.Unlock()
	return o.stype
}

// SetType sets the service type associated with the op.
func (o *BrowseOp) SetType(s string) error {
	o.m.Lock()
	defer o.m.Unlock()
	if o.started {
		return ErrStarted
	}
	o.stype = s
	return nil
}

// Domain returns the domain associated with the op.
func (o *BrowseOp) Domain() string {
	o.m.Lock()
	defer o.m.Unlock()
	return o.domain
}

// SetDomain sets the domain associated with the op.
func (o *BrowseOp) SetDomain(s string) error {
	o.m.Lock()
	defer o.m.Unlock()
	if o.started {
		return ErrStarted
	}
	o.domain = s
	return nil
}

// SetCallback sets the function to call when an error occurs or a service is lost or found.
func (o *BrowseOp) SetCallback(f BrowseCallbackFunc) error {
	o.m.Lock()
	defer o.m.Unlock()
	if o.started {
		return ErrStarted
	}
	o.callback = f
	return nil
}

// Start begins the browse query.
func (o *BrowseOp) Start() error {
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

func (o *BrowseOp) init(sharedref uintptr) (ref uintptr, err error) {
	ref = sharedref
	o.setFlag(_FlagsShareConnection, ref != 0)
	if err = browseStart(&ref, o.flags, o.interfaceIndexC(), o.stype, o.domain, unsafe.Pointer(o)); err != nil {
		ref = 0
	}
	return
}

// Stop stops the operation.
func (o *BrowseOp) Stop() {
	o.m.Lock()
	defer o.m.Unlock()
	if !o.started {
		return
	}
	o.started = false
	pollServer.stopOp(o)
}

func (o *BrowseOp) handleError(e error) {
	if !o.started {
		return
	}
	o.started = false
	pollServer.removePollOp(o)
	queueCallback(func() { o.callback(o, e, false, 0, "", "", "") })
}

func dnssdBrowseCallback(sdRef unsafe.Pointer, flags, interfaceIndex uint32, err int32, name, stype, domain unsafe.Pointer, ctx unsafe.Pointer) {
	o := (*BrowseOp)(ctx)
	if e := getError(err); e != nil {
		o.handleError(e)
	} else {
		a := flags&_FlagsAdd != 0
		i := int(interfaceIndex)
		n := cStringToString(name)
		t := cStringToString(stype)
		d := cStringToString(domain)
		queueCallback(func() { o.callback(o, nil, a, i, n, t, d) })
	}
}
