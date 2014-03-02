// +build darwin freebsd linux netbsd openbsd

package dnssd

/*

#cgo !darwin LDFLAGS: -ldns_sd

#include <stdlib.h>
#include <sys/select.h>
#include <arpa/inet.h>
#include <dns_sd.h>

extern void browseCallbackWrapper(
	void                  *sdRef,
	uint32_t              flags,
	uint32_t              ifIndex,
	DNSServiceErrorType   errorCode,
	void                  *serviceName,
	void                  *regtype,
	void                  *replyDomain,
	void                  *context
);

static int32_t dnssdBrowse(
	void                  *sdRef,
	DNSServiceFlags       flags,
	uint32_t              ifIndex,
	const char            *regtype,
	const char            *domain,
	void                  *context
	) {
	DNSServiceBrowseReply callback = (DNSServiceBrowseReply) browseCallbackWrapper;
	return DNSServiceBrowse(sdRef, flags, ifIndex, regtype, domain, callback, context);
}

extern void registerCallbackWrapper(
	void                  *sdRef,
	uint32_t              flags,
	int32_t               errorCode,
	void                  *name,
	void                  *regtype,
	void                  *domain,
	void                  *context
);

static int32_t dnssdRegister(
	void                  *sdRef,
	DNSServiceFlags       flags,
	uint32_t              ifIndex,
	const char            *name,
	const char            *regtype,
	const char            *domain,
	const char            *host,
	uint16_t              port,
	uint16_t              txtLen,
	const void            *txtRecord,
	void                  *context
	) {
	port = htons(port);
	DNSServiceRegisterReply callback = (DNSServiceRegisterReply) registerCallbackWrapper;
	return DNSServiceRegister(sdRef, flags, ifIndex, name, regtype, domain, host, port, txtLen, txtRecord, callback, context);
}

extern void resolveCallbackWrapper(
	void                  *sdRef,
	uint32_t              flags,
	uint32_t              ifIndex,
	int32_t               errorCode,
	void                  *fullname,
	void                  *hosttarget,
	uint16_t              port,
	uint16_t              txtLen,
	void                  *txtRecord,
	void                  *context
	);

static int32_t dnssdResolve(
	void                  *sdRef,
	DNSServiceFlags       flags,
	uint32_t              ifIndex,
	const char            *name,
	const char            *regtype,
	const char            *domain,
	void                  *context
	) {
	DNSServiceResolveReply callback = (DNSServiceResolveReply) resolveCallbackWrapper;
	return DNSServiceResolve(sdRef, flags, ifIndex, name, regtype, domain, callback, context);
}

extern void queryCallbackWrapper(
    void                  *sdRef,
    uint32_t              flags,
    uint32_t              ifIndex,
    int32_t               errorCode,
    void                  *fullname,
    uint16_t              rrtype,
    uint16_t              rrclass,
    uint16_t              rdlen,
    void                  *rdata,
    uint32_t              ttl,
    void                  *context
    );

static int32_t dnssdQuery(
    void                  *sdRef,
    DNSServiceFlags       flags,
    uint32_t              ifIndex,
    const char            *name,
    uint16_t              rrtype,
    uint16_t              rrclass,
    void                  *context
    ) {
    DNSServiceQueryRecordReply callback = (DNSServiceQueryRecordReply) queryCallbackWrapper;
    return DNSServiceQueryRecord(sdRef, flags, ifIndex, name, rrtype, rrclass, callback, context);
}

static uint16_t dnssdNtohs(uint16_t n) {
	return ntohs(n);
}

static int dnssdSelect(int nfds, fd_set *readfds, fd_set *writefds, fd_set *errorfds) {
	struct timeval tv;
	tv.tv_sec = 100000000;
	tv.tv_usec = 0;
	return select(nfds, readfds, writefds, errorfds, &tv);
}

static void dnssdFdZero(fd_set *fdset) {
	FD_ZERO(fdset);
}

static void dnssdFdSet(int fd, fd_set *fdset) {
	FD_SET(fd, fdset);
}

static int dnssdFdIsSet(int fd, fd_set *fdset) {
	return FD_ISSET(fd, fdset);
}

*/
import "C"
import (
	"os"
	"unsafe"
)

func browseStart(ref *uintptr, flags, ifIndex uint32, typ, domain string, ctx unsafe.Pointer) error {
	cref := unsafe.Pointer(ref)
	cflags := C.DNSServiceFlags(flags)
	cifIndex := C.uint32_t(ifIndex)
	ctype := C.CString(typ)
	defer C.free(unsafe.Pointer(ctype))
	cdomain := C.CString(domain)
	defer C.free(unsafe.Pointer(cdomain))
	return getError(int32(C.dnssdBrowse(cref, cflags, cifIndex, ctype, cdomain, ctx)))
}

//export browseCallbackWrapper
func browseCallbackWrapper(sdRef unsafe.Pointer, flags, ifIndex uint32, err int32, name, stype, domain unsafe.Pointer, ctx unsafe.Pointer) {
	dnssdBrowseCallback(sdRef, flags, ifIndex, err, name, stype, domain, ctx)
}

func resolveStart(ref *uintptr, flags, ifIndex uint32, name, typ, domain string, ctx unsafe.Pointer) error {
	cref := unsafe.Pointer(ref)
	cflags := C.DNSServiceFlags(flags)
	cifIndex := C.uint32_t(ifIndex)
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	ctype := C.CString(typ)
	defer C.free(unsafe.Pointer(ctype))
	cdomain := C.CString(domain)
	defer C.free(unsafe.Pointer(cdomain))
	return getError(int32(C.dnssdResolve(cref, cflags, cifIndex, cname, ctype, cdomain, ctx)))
}

//export resolveCallbackWrapper
func resolveCallbackWrapper(sdRef unsafe.Pointer, flags, ifIndex uint32, err int32, fullname, hosttarget unsafe.Pointer, port, txtLen uint16, txtRecord, ctx unsafe.Pointer) {
	port = uint16(C.dnssdNtohs(C.uint16_t(port)))
	dnssdResolveCallback(sdRef, flags, ifIndex, err, fullname, hosttarget, port, txtLen, txtRecord, ctx)
}

func registerStart(ref *uintptr, flags, ifIndex uint32, name, typ, domain, host string, port int, txt []byte, ctx unsafe.Pointer) error {
	cref := unsafe.Pointer(ref)
	cflags := C.DNSServiceFlags(flags)
	cifIndex := C.uint32_t(ifIndex)
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	ctyp := C.CString(typ)
	defer C.free(unsafe.Pointer(ctyp))
	cdomain := C.CString(domain)
	defer C.free(unsafe.Pointer(cdomain))
	chost := C.CString(host)
	defer C.free(unsafe.Pointer(chost))
	cport := C.uint16_t(port)
	txtLen := C.uint16_t(len(txt))
	txtPtr := unsafe.Pointer(nil)
	if txtLen > 0 {
		txtPtr = unsafe.Pointer(&txt[0])
	}
	e := C.dnssdRegister(cref, cflags, cifIndex, cname, ctyp, cdomain, chost, cport, txtLen, txtPtr, ctx)
	return getError(int32(e))
}

//export registerCallbackWrapper
func registerCallbackWrapper(sdRef unsafe.Pointer, flags uint32, err int32, name, regtype, domain, ctx unsafe.Pointer) {
	dnssdRegisterCallback(sdRef, flags, err, name, regtype, domain, ctx)
}

func queryStart(ref *uintptr, flags, ifIndex uint32, name string, rrtype, rrclass uint16, ctx unsafe.Pointer) error {
	cref := unsafe.Pointer(ref)
	cflags := C.DNSServiceFlags(flags)
	cifIndex := C.uint32_t(ifIndex)
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	crrtype, crrclass := C.uint16_t(rrtype), C.uint16_t(rrclass)
	e := C.dnssdQuery(cref, cflags, cifIndex, cname, crrtype, crrclass, ctx)
	return getError(int32(e))
}

//export queryCallbackWrapper
func queryCallbackWrapper(sdRef unsafe.Pointer, flags, ifIndex uint32, err int32, f unsafe.Pointer, rrtype, rrclass, rdlen uint16, rdata unsafe.Pointer, ttl uint32, ctx unsafe.Pointer) {
	dnssdQueryCallback(sdRef, flags, ifIndex, err, f, rrtype, rrclass, rdlen, rdata, ttl, ctx)
}

func refSockFd(ref *uintptr) int {
	return int(C.DNSServiceRefSockFD(*(*C.DNSServiceRef)(unsafe.Pointer(ref))))
}

func platformDeallocateRef(ref *uintptr) {
	C.DNSServiceRefDeallocate(*(*C.DNSServiceRef)(unsafe.Pointer(ref)))
}

func createConnection(ref *uintptr) error {
	return getError(int32(C.DNSServiceCreateConnection((*C.DNSServiceRef)(unsafe.Pointer(ref)))))
}

func processResult(ref uintptr) error {
	return getError(int32(C.DNSServiceProcessResult(C.DNSServiceRef(unsafe.Pointer(ref)))))
}

func fdSet(fd, maxfd int, s *C.fd_set) int {
	C.dnssdFdSet(C.int(fd), s)
	if maxfd < fd {
		return fd
	}
	return maxfd
}

func fdIsSet(fd int, s *C.fd_set) bool {
	return C.dnssdFdIsSet(C.int(fd), s) != 0
}

type platformPollServerState struct{ pipe struct{ r, w *os.File } }

func (s *pollServerState) stopPoll() {
	if s.pipe.w == nil {
		return
	}
	_, err := s.pipe.w.WriteString("I")
	if err != nil {
		panic(err)
	}
}

func (s *pollServerState) startPoll() {
	s.m.internal.Lock()
	if s.pipe.w == nil {
		r, w, err := os.Pipe()
		if err != nil {
			panic(err)
		}
		s.pipe.r, s.pipe.w = r, w
	}
	go pollLoop(s)
}

func pollLoop(s *pollServerState) {
	defer s.m.internal.Unlock()
	pipefd := int(s.pipe.r.Fd())
	pipebuf := make([]byte, 1)
	var readset C.fd_set
	for {
		sharedPollables, uniquePollables := s.sharedAndUniquePollables()
		C.dnssdFdZero(&readset)
		maxfd := fdSet(pipefd, 0, &readset)
		if s.shared.fd > 0 {
			maxfd = fdSet(s.shared.fd, maxfd, &readset)
		}
		for i := range uniquePollables {
			maxfd = fdSet(uniquePollables[i].fd, maxfd, &readset)
		}
		if r := C.dnssdSelect(C.int(maxfd+1), &readset, nil, nil); r <= 0 {
			continue
		}
		if fd := s.shared.fd; fd > 0 && fdIsSet(fd, &readset) {
			if e := processResult(s.shared.ref); e != nil {
				// ref is no longer valid. ops using callback should have had their
				// callback invoked. can call them anyway since we only pass on the first error.
				s.shared.ref = 0
				s.shared.fd = 0
				for i := range sharedPollables {
					sharedPollables[i].ref = 0
					sharedPollables[i].p.handleError(e)
				}
			}
		}
		for i := range uniquePollables {
			op := uniquePollables[i]
			if fd := op.fd; fdIsSet(fd, &readset) {
				if e := processResult(op.ref); e != nil {
					// invalidate the ref. not clear if callback will have been invoked
					// but can call it anyway since only the first error gets passed on
					op.p.handleError(e)
				}
			}
		}
		if fdIsSet(pipefd, &readset) {
			_, err := s.pipe.r.Read(pipebuf)
			if err != nil {
				panic(err)
			}
			return
		}
	}
}
