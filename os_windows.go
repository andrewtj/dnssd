package dnssd

import (
	"sync"
	"syscall"
	"unsafe"
)

var lib struct {
	dll  map[string]*syscall.DLL
	proc map[string]map[string]*syscall.Proc
}

func getDLL(name string) (*syscall.DLL, error) {
	if dll, present := lib.dll[name]; present {
		return dll, nil
	}
	dll, err := syscall.LoadDLL(name)
	if err != nil {
		return nil, err
	}
	if lib.dll == nil {
		lib.dll = make(map[string]*syscall.DLL)
	}
	lib.dll[name] = dll
	if lib.proc == nil {
		lib.proc = make(map[string]map[string]*syscall.Proc)
	}
	lib.proc[name] = make(map[string]*syscall.Proc)
	return dll, nil
}

func getProc(dllName, procName string) (*syscall.Proc, error) {
	if proc, present := lib.proc[dllName][procName]; present {
		return proc, nil
	}
	dll, err := getDLL(dllName)
	if err != nil {
		return nil, err
	}
	proc, err := dll.FindProc(procName)
	if err != nil {
		return nil, err
	}
	lib.proc[dllName][procName] = proc
	return proc, nil
}

func mustGetProc(dllName, procName string) *syscall.Proc {
	proc, err := getProc(dllName, procName)
	if err != nil {
		panic(err)
	}
	return proc
}

func browseStart(ref *uintptr, flags, ifIndex uint32, typ, domain string, ctx unsafe.Pointer) error {
	proc, err := getProc("dnssd.dll", "DNSServiceBrowse")
	if err != nil {
		return err
	}
	btyp, err := syscall.BytePtrFromString(typ)
	if err != nil {
		return err
	}
	bdomain, err := syscall.BytePtrFromString(domain)
	if err != nil {
		return err
	}
	r, _, _ := proc.Call(
		(uintptr)(unsafe.Pointer(ref)),
		uintptr(flags),
		uintptr(ifIndex),
		(uintptr)(unsafe.Pointer(btyp)),
		(uintptr)(unsafe.Pointer(bdomain)),
		syscall.NewCallback(browseCallbackWrapper),
		(uintptr)(ctx),
	)
	return getError(int32(r))
}

func browseCallbackWrapper(sdRef unsafe.Pointer, flags, interfaceIndex uint32, err int32, name, stype, domain unsafe.Pointer, ctx unsafe.Pointer) int32 {
	dnssdBrowseCallback(sdRef, flags, interfaceIndex, err, name, stype, domain, ctx)
	return 0
}

func resolveStart(ref *uintptr, flags, ifIndex uint32, name, typ, domain string, ctx unsafe.Pointer) error {
	proc, err := getProc("dnssd.dll", "DNSServiceResolve")
	if err != nil {
		return err
	}
	bname, err := syscall.BytePtrFromString(name)
	if err != nil {
		return err
	}
	btyp, err := syscall.BytePtrFromString(typ)
	if err != nil {
		return err
	}
	bdomain, err := syscall.BytePtrFromString(domain)
	if err != nil {
		return err
	}
	r, _, _ := proc.Call(
		(uintptr)(unsafe.Pointer(ref)),
		uintptr(flags),
		uintptr(ifIndex),
		uintptr(unsafe.Pointer(bname)),
		uintptr(unsafe.Pointer(btyp)),
		uintptr(unsafe.Pointer(bdomain)),
		syscall.NewCallback(dnssdResolveCallbackWrapper),
		uintptr(ctx),
	)
	return getError(int32(r))
}

func dnssdResolveCallbackWrapper(sdRef unsafe.Pointer, flags, interfaceIndex uint32, err int32, fullname, hosttarget unsafe.Pointer, port uint16, txtLen uint32 /* docs say uint16 but it seems to be a uint32 */, txtRecord, ctx unsafe.Pointer) int32 {
	dnssdResolveCallback(sdRef, flags, interfaceIndex, err, fullname, hosttarget, port, uint16(txtLen), txtRecord, ctx)
	return 0
}

func registerStart(ref *uintptr, flags, ifIndex uint32, name, typ, domain, host string, port int, txt []byte, ctx unsafe.Pointer) error {
	proc, err := getProc("dnssd.dll", "DNSServiceRegister")
	if err != nil {
		return err
	}
	bname, err := syscall.BytePtrFromString(name)
	if err != nil {
		return err
	}
	btyp, err := syscall.BytePtrFromString(typ)
	if err != nil {
		return err
	}
	bdomain, err := syscall.BytePtrFromString(domain)
	if err != nil {
		return err
	}
	bhost, err := syscall.BytePtrFromString(host)
	if err != nil {
		return err
	}

	txtLen := uintptr(len(txt))
	txtPtr := unsafe.Pointer(nil)
	if txtLen > 0 {
		txtPtr = unsafe.Pointer(&txt[0])
	}

	r, _, _ := proc.Call(
		(uintptr)(unsafe.Pointer(ref)),
		uintptr(flags),
		uintptr(ifIndex),
		(uintptr)(unsafe.Pointer(bname)),
		(uintptr)(unsafe.Pointer(btyp)),
		(uintptr)(unsafe.Pointer(bdomain)),
		(uintptr)(unsafe.Pointer(bhost)),
		uintptr(port),
		txtLen,
		(uintptr)(txtPtr),
		syscall.NewCallback(registerCallbackWrapper),
		uintptr(ctx),
	)
	return getError(int32(r))

}

func registerCallbackWrapper(sdRef unsafe.Pointer, flags uint32, err int32, name, regtype, domain, ctx unsafe.Pointer) int32 {
	dnssdRegisterCallback(sdRef, flags, err, name, regtype, domain, ctx)
	return 0
}

func queryStart(ref *uintptr, flags, ifIndex uint32, name string, rrtype, rrclass uint16, ctx unsafe.Pointer) error {
	proc, err := getProc("dnssd.dll", "DNSServiceQueryRecord")
	if err != nil {
		return err
	}
	bname, err := syscall.BytePtrFromString(name)
	if err != nil {
		return err
	}
	r, _, _ := proc.Call(
		(uintptr)(unsafe.Pointer(ref)),
		uintptr(flags),
		uintptr(ifIndex),
		(uintptr)(unsafe.Pointer(bname)),
		uintptr(rrtype),
		uintptr(rrclass),
		syscall.NewCallback(queryCallbackWrapper),
		(uintptr)(ctx),
	)
	return getError(int32(r))
}

func queryCallbackWrapper(sdRef unsafe.Pointer, flags, ifIndex uint32, err int32, fullname unsafe.Pointer, rrtype, rrclass, rdlen uint32 /* docs say uint16 but seems to be a uint32 !*/, rdataptr unsafe.Pointer, ttl uint32, ctx unsafe.Pointer) int32 {
	dnssdQueryCallback(sdRef, flags, ifIndex, err, fullname, uint16(rrtype), uint16(rrclass), uint16(rdlen), rdataptr, ttl, ctx)
	return 0
}

func refSockFd(ref *uintptr) int {
	proc := mustGetProc("dnssd.dll", "DNSServiceRefSockFD")
	fd, _, _ := proc.Call(*ref)
	return int(fd)
}

func platformDeallocateRef(ref *uintptr) {
	proc := mustGetProc("dnssd.dll", "DNSServiceRefDeallocate")
	_, _, _ = proc.Call(*ref)
}

func createConnection(ref *uintptr) error {
	proc, err := getProc("dnssd.dll", "DNSServiceCreateConnection")
	if err != nil {
		return err
	}
	e, _, _ := proc.Call((uintptr)(unsafe.Pointer(ref)))
	return getError(int32(e))
}

type platformPollServerState struct {
	event uintptr
	once  sync.Once
}

func (s *pollServerState) stopPoll() {
	if s.event != 0 {
		mustGetProc("ws2_32.dll", "WSASetEvent").Call(s.event)
	}
}

func (s *pollServerState) startPoll() {
	s.m.internal.Lock()
	s.once.Do(func() {
		s.event = createEvent()
	})
	go pollLoop(s)
}

func pollLoop(s *pollServerState) {
	defer s.m.internal.Unlock()
	sharedPollables, uniquePollables := s.sharedAndUniquePollables()
	events := []uintptr{s.event}
	for i := range uniquePollables {
		events = append(events, createFdEvent(uniquePollables[i].fd))
	}
	if s.shared.fd > 0 {
		events = append(events, createFdEvent(s.shared.fd))
	}
	for {
		r, _, err := waitForObjects(events)
		switch r {
		case 0xFFFFFFFF: // WAIT_FAILED
			panic(err)
		case 0x00000102: // WAIT_OBJECT_TIMEOUT
		case 0: // WAIT_OBJECT_0
			resetEvent(events[r])
			return
		default: // WAIT_OBJECT_N
			resetEvent(events[r])
			var ref uintptr
			if s.shared.ref != 0 && len(events)-1 == int(r) {
				ref = s.shared.ref
			} else {
				ref = uniquePollables[r].ref
			}
			e, _, _ := mustGetProc("dnssd.dll", "DNSServiceProcessResult").Call(ref)
			err := getError(int32(e))
			if err != nil && ref == s.shared.ref {
				// ref is no longer valid. ops using callback should have had their
				// callback invoked. can call them anyway since we only pass on the first error.
				s.shared.ref = 0
				s.shared.fd = 0
				for i := range sharedPollables {
					sharedPollables[i].ref = 0
					sharedPollables[i].p.handleError(err)
				}
			} else if err != nil {
				uniquePollables[r].p.handleError(err)
			}
			if err != nil {
				events = append(events[:r], events[r+1:]...)
			}
		}
	}
}

func createEvent() uintptr {
	event, _, err := mustGetProc("ws2_32.dll", "WSACreateEvent").Call()
	if event == 0 {
		panic(err)
	}
	return event
}

func createFdEvent(fd int) uintptr {
	event := createEvent()
	proc := mustGetProc("ws2_32.dll", "WSAEventSelect")
	r, _, err := proc.Call(uintptr(fd), event, 1) //  1 == FD_READ
	if r == ^uintptr(0) {                         // -1 == SOCKET_ERROR
		panic(err)
	}
	return event
}

func resetEvent(event uintptr) {
	r, _, err := mustGetProc("ws2_32.dll", "WSAResetEvent").Call(event)
	if r == 0 {
		panic(err)
	}
}

func waitForObjects(events []uintptr) (r1, r2 uintptr, lastErr error) {
	proc := mustGetProc("kernel32.dll", "WaitForMultipleObjects")
	r1, r2, lastErr = proc.Call(uintptr(len(events)), uintptr(unsafe.Pointer(&events[0])), 0, 0xFFFFFFFF)
	return
}
