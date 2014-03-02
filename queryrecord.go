package dnssd

import "unsafe"

// QueryCallbackFunc is called when an error occurs or a record is added or removed.
// Results may be cached for ttl seconds. After ttl seconds the result should be discarded.
// Alternatively the operation may be left running in which case the result can be considered valid
// until a callback indicates otherwise.
type QueryCallbackFunc func(op *QueryOp, err error, add bool, interfaceIndex int, fullname string, rrtype, rrclass uint16, rdata []byte, ttl uint32)

// QueryOp represents a query for a specific name, class and type.
type QueryOp struct {
	baseOp
	name            string
	rrtype, rrclass uint16
	callback        QueryCallbackFunc
}

// NewQueryOp creates a new QueryOp with the associated parameters set.
func NewQueryOp(interfaceIndex int, name string, rrtype, rrclass uint16, f QueryCallbackFunc) *QueryOp {
	op := &QueryOp{}
	op.SetInterfaceIndex(interfaceIndex)
	op.SetName(name)
	op.SetType(rrtype)
	op.SetClass(rrclass)
	op.SetCallback(f)
	return op
}

// StartQueryOp returns the equivalent of calling NewQueryOp and Start.
func StartQueryOp(interfaceIndex int, name string, rrtype, rrclass uint16, f QueryCallbackFunc) (*QueryOp, error) {
	op := NewQueryOp(interfaceIndex, name, rrtype, rrclass, f)
	return op, op.Start()
}

// Name returns the domain name for the operation.
func (o *QueryOp) Name() string {
	o.m.Lock()
	defer o.m.Unlock()
	return o.name
}

// SetName sets the domain name for the operation.
func (o *QueryOp) SetName(n string) error {
	o.m.Lock()
	defer o.m.Unlock()
	if o.started {
		return ErrStarted
	}
	o.name = n
	return nil
}

// Type returns the DNS Resource Record Type for the operation.
func (o *QueryOp) Type() uint16 {
	o.m.Lock()
	defer o.m.Unlock()
	return o.rrtype
}

// SetType sets the DNS Resource Record Type for the operation.
func (o *QueryOp) SetType(t uint16) error {
	o.m.Lock()
	defer o.m.Unlock()
	if o.started {
		return ErrStarted
	}
	o.rrtype = t
	return nil
}

// Class returns the DNS Resource Record Class for the operation.
func (o *QueryOp) Class() uint16 {
	o.m.Lock()
	defer o.m.Unlock()
	return o.rrclass
}

// SetClass sets the DNS Resource Record Class for the operation.
func (o *QueryOp) SetClass(c uint16) error {
	o.m.Lock()
	defer o.m.Unlock()
	if o.started {
		return ErrStarted
	}
	o.rrclass = c
	return nil
}

// SetCallback sets the function to call when an error occurs or a record is added or removed.
func (o *QueryOp) SetCallback(f QueryCallbackFunc) error {
	o.m.Lock()
	defer o.m.Unlock()
	if o.started {
		return ErrStarted
	}
	o.callback = f
	return nil
}

// Start begins the query operation.
func (o *QueryOp) Start() error {
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

func (o *QueryOp) init(sharedref uintptr) (ref uintptr, err error) {
	ref = sharedref
	o.setFlag(_FlagsShareConnection, ref != 0)
	if err = queryStart(&ref, o.flags, o.interfaceIndexC(), o.name, o.rrtype, o.rrclass, unsafe.Pointer(o)); err != nil {
		ref = 0
	}
	return
}

// Stop stops the operation.
func (o *QueryOp) Stop() {
	o.m.Lock()
	defer o.m.Unlock()
	if !o.started {
		return
	}
	o.started = false
}

func (o *QueryOp) handleError(e error) {
	if !o.started {
		return
	}
	o.started = false
	pollServer.removePollOp(o)
	queueCallback(func() { o.callback(o, e, false, 0, "", 0, 0, nil, 0) })
}

func dnssdQueryCallback(sdRef unsafe.Pointer, flags, interfaceIndex uint32, err int32, fullname unsafe.Pointer, rrtype, rrclass, rdlen uint16, rdataptr unsafe.Pointer, ttl uint32, ctx unsafe.Pointer) {
	o := (*QueryOp)(ctx)
	if e := getError(err); e != nil {
		o.handleError(e)
	} else {
		a := flags&_FlagsAdd != 0
		i := int(interfaceIndex)
		f := cStringToString(fullname)
		var rdata []byte
		if rdlen > 0 && rdataptr != nil {
			s := (*[65535]byte)(rdataptr)[:rdlen]
			rdata = make([]byte, rdlen)
			copy(rdata, s)
		}
		queueCallback(func() { o.callback(o, e, a, i, f, rrtype, rrclass, rdata, ttl) })
	}
}
