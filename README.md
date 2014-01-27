
# dnssd
    import "github.com/andrewtj/dnssd"

Package dnssd implements a wrapper for Apple's C DNS Service Discovery API.

The DNS Service Discovery API is part of the Apple Bonjour zero
configuration networking stack. The API allows for network services to be
registered, browsed and resolved without configuration via multicast DNS
in the ".local" domain and with additional configuration in unicast DNS
domains. A service consists of a name, type, host, port and a set of
key-value pairs containing meta information.

Bonjour is bundled with OS X and available for Windows via [Bonjour Print
Services for Windows](http://support.apple.com/kb/dl999), the [Bonjour SDK for
Windows](https://developer.apple.com/bonjour/) or bundled with iTunes.
For other POSIX platforms Apple offer mDNSResponder³ as open-source, however
the [Avahi](http://avahi.org/) project is the de facto choice on most Linux
and BSD systems. Although Avahi has a different API, it does offer a
compatibility shim which covers a subset of the DNS Service Discovery API, and
which this package largely sticks to.

The DNS Service Discovery API is wrapped as follows:


	DNSServiceRegister() -> RegisterOp
	DNSServiceBrowse()   -> BrowseOp
	DNSServiceResolve()  -> ResolveOp

All operations require a callback and service type be set. If an
InterfaceIndex is not set the default value of InterfaceIndexAny is used
which applies the operation to all network interfaces. If no domain is set or
the domain is set to an empty-string the operation applies to all applicable
DNS-SD domains.

If a service is registered with an empty string as it's name, the local
computer name (or hostname) will be substitued. If no host is specified a
hostname for the local machine will be used. By default services will be
renamed with a numeric suffix if a name collision occurs.

Callbacks are executed in serial. If an error is supplied to a callback
the operation will no longer be active and other arguments must be ignored.




## Constants
``` go
const InterfaceIndexAny = 0
```
InterfaceIndexAny is the default for all operations.

``` go
const InterfaceIndexLocalOnly = int(^uint(0) >> 1)
```
InterfaceIndexLocalOnly limits the scope of the operation to the local machine.


## Variables
``` go
var (
    ErrUnknown                   = Error{-65537, "Unknown"}
    ErrNoSuchName                = Error{-65538, "No Such Name"}
    ErrNoMemory                  = Error{-65539, "No Memory"}
    ErrBadParam                  = Error{-65540, "Bad Param"}
    ErrBadReference              = Error{-65541, "Bad Reference"}
    ErrBadState                  = Error{-65542, "Bad State"}
    ErrBadFlags                  = Error{-65543, "Bad Flags"}
    ErrUnsupported               = Error{-65544, "Unsupported"}
    ErrNotInitialized            = Error{-65545, "Not Initialized"}
    ErrAlreadyRegistered         = Error{-65547, "Already Registered"}
    ErrNameConflict              = Error{-65548, "Name Conflict"}
    ErrInvalid                   = Error{-65549, "Invalid"}
    ErrFirewall                  = Error{-65550, "Firewall"}
    ErrIncompatible              = Error{-65551, "Incompatible"}
    ErrBadInterfaceIndex         = Error{-65552, "Bad Interface Index"}
    ErrRefused                   = Error{-65553, "Refused"}
    ErrNoSuchRecord              = Error{-65554, "No Such Record"}
    ErrNoAuth                    = Error{-65555, "No Auth"}
    ErrNoSuchKey                 = Error{-65556, "No Such Key"}
    ErrNATTraversal              = Error{-65557, "NAT Traversal"}
    ErrDoubleNAT                 = Error{-65558, "Double NAT"}
    ErrBadTime                   = Error{-65559, "Bad Time"}
    ErrBadSig                    = Error{-65560, "Bad Sig"}
    ErrBadKey                    = Error{-65561, "Bad Key"}
    ErrTransient                 = Error{-65562, "Transient"}
    ErrServiceNotRunning         = Error{-65563, "Service Not Running"}
    ErrNATPortMappingUnsupported = Error{-65564, "NAT Port Mapping Unsupported"}
    ErrNATPortMappingDisabled    = Error{-65565, "NAT Port Mapping Disabled"}
    ErrNoRouter                  = Error{-65566, "No Router"}
    ErrPollingMode               = Error{-65567, "Polling Mode"}
    ErrTimeout                   = Error{-65568, "Timeout"}
)
```
Errors returned by the underlying C API.

``` go
var ErrMissingCallback = errors.New("no callback set")
```
ErrMissingCallback is returned when an operation is started without setting a callback.

``` go
var ErrStarted = errors.New("already started")
```
ErrStarted is returned when trying to mutate an active operation or when starting a started operation.

``` go
var ErrTXTLen = errors.New("TXT size may not exceed 65535 bytes")
```
ErrTXTLen is returned when setting a TXT pair that would exceed the 65,535 byte TXT record limit.

``` go
var ErrTXTStringLen = errors.New("TXT string may not exceed 255 bytes")
```
ErrTXTStringLen is returned when setting a TXT pair that would exceed the 255 byte string limit.



## type BrowseCallbackFunc
``` go
type BrowseCallbackFunc func(op *BrowseOp, err error, add bool, interfaceIndex int, name string, serviceType string, domain string)
```
BrowseCallbackFunc is called when an error occurs or a service is lost or found.











## type BrowseOp
``` go
type BrowseOp struct {
    // contains filtered or unexported fields
}
```
BrowseOp represents a query for services of a particular type.









### func NewBrowseOp
``` go
func NewBrowseOp(serviceType string, f BrowseCallbackFunc) *BrowseOp
```
NewBrowseOp creates a new BrowseOp with the given service type and call back set.


### func StartBrowseOp
``` go
func StartBrowseOp(serviceType string, f BrowseCallbackFunc) (*BrowseOp, error)
```
StartBrowseOp returns the equivalent of calling NewBrowseOp and Start().




### func (\*BrowseOp) Active
``` go
func (o *BrowseOp) Active() bool
```
Active indicates whether an operation is active



### func (\*BrowseOp) Domain
``` go
func (o *BrowseOp) Domain() string
```
Domain returns the domain associated with the op.



### func (\*BrowseOp) InterfaceIndex
``` go
func (o *BrowseOp) InterfaceIndex() int
```
InterfaceIndex returns the interface index the op is tied to.



### func (\*BrowseOp) SetCallback
``` go
func (o *BrowseOp) SetCallback(f BrowseCallbackFunc) error
```
SetCallback sets the function to call when an error occurs or a service is lost or found.



### func (\*BrowseOp) SetDomain
``` go
func (o *BrowseOp) SetDomain(s string) error
```
SetDomain sets the domain associated with the op.



### func (\*BrowseOp) SetInterfaceIndex
``` go
func (o *BrowseOp) SetInterfaceIndex(i int) error
```
SetInterfaceIndex sets the interface index the op is tied to.



### func (\*BrowseOp) SetType
``` go
func (o *BrowseOp) SetType(s string) error
```
SetType sets the service type associated with the op.



### func (\*BrowseOp) Start
``` go
func (o *BrowseOp) Start() error
```
Start begins the browse query.



### func (\*BrowseOp) Stop
``` go
func (o *BrowseOp) Stop()
```
Stop stops the operation.



### func (\*BrowseOp) Type
``` go
func (o *BrowseOp) Type() string
```
Type returns the service type associated with the op.



## type Error
``` go
type Error struct {
    // contains filtered or unexported fields
}
```
Error structs meet the error interface and are returned when errors occur in the underlying C API.











### func (Error) Desc
``` go
func (e Error) Desc() string
```
Desc returns a string describing the error.



### func (Error) Error
``` go
func (e Error) Error() string
```


### func (Error) Num
``` go
func (e Error) Num() int32
```
Num returns an error number.



## type RegisterCallbackFunc
``` go
type RegisterCallbackFunc func(op *RegisterOp, err error, add bool, name, serviceType, domain string)
```
RegisterCallbackFunc is called when a name is registered or deregistered in a given domain, or when an error occurs.











## type RegisterOp
``` go
type RegisterOp struct {
    // contains filtered or unexported fields
}
```
RegisterOp represents a service registration operation.









### func NewProxyRegisterOp
``` go
func NewProxyRegisterOp(name, serviceType, host string, port int, f RegisterCallbackFunc) *RegisterOp
```
NewProxyRegisterOp creates a new RegisterOp with the given parameters set.


### func NewRegisterOp
``` go
func NewRegisterOp(name, serviceType string, port int, f RegisterCallbackFunc) *RegisterOp
```
NewRegisterOp creates a new RegisterOp with the given parameters set.


### func StartProxyRegisterOp
``` go
func StartProxyRegisterOp(name, serviceType, host string, port int, f RegisterCallbackFunc) (*RegisterOp, error)
```
StartProxyRegisterOp returns the equivalent of calling NewProxyRegisterOp and Start().


### func StartRegisterOp
``` go
func StartRegisterOp(name, serviceType string, port int, f RegisterCallbackFunc) (*RegisterOp, error)
```
StartRegisterOp returns the equivalent of calling NewRegisterOp and Start().




### func (\*RegisterOp) Active
``` go
func (o *RegisterOp) Active() bool
```
Active indicates whether an operation is active



### func (\*RegisterOp) DeleteTXTPair
``` go
func (o *RegisterOp) DeleteTXTPair(key string) error
```
DeleteTXTPair deletes the TXT string with the provided key.



### func (\*RegisterOp) Domain
``` go
func (o *RegisterOp) Domain() string
```
Domain returns the domain associated with the op.



### func (\*RegisterOp) Host
``` go
func (o *RegisterOp) Host() string
```
Host returns the hostname of the service. An empty string will result in the local machine's hostname being used.



### func (\*RegisterOp) InterfaceIndex
``` go
func (o *RegisterOp) InterfaceIndex() int
```
InterfaceIndex returns the interface index the op is tied to.



### func (\*RegisterOp) Name
``` go
func (o *RegisterOp) Name() string
```
Name returns the name of the service.



### func (\*RegisterOp) NoAutoRename
``` go
func (o *RegisterOp) NoAutoRename() bool
```
NoAutoRename indicates how service-name conflicts will be handled.



### func (\*RegisterOp) Port
``` go
func (o *RegisterOp) Port() int
```
Port returns the port the service is available from.



### func (\*RegisterOp) SetCallback
``` go
func (o *RegisterOp) SetCallback(f RegisterCallbackFunc) error
```
SetCallback sets the function to call when a name is registered or deregistered in a given domain, or when an error occurs.



### func (\*RegisterOp) SetDomain
``` go
func (o *RegisterOp) SetDomain(s string) error
```
SetDomain sets the domain associated with the op.



### func (\*RegisterOp) SetHost
``` go
func (o *RegisterOp) SetHost(h string) error
```
SetHost sets the hostname of the service. An empty string will result in the local machine's hostname being used.



### func (\*RegisterOp) SetInterfaceIndex
``` go
func (o *RegisterOp) SetInterfaceIndex(i int) error
```
SetInterfaceIndex sets the interface index the op is tied to.



### func (\*RegisterOp) SetName
``` go
func (o *RegisterOp) SetName(n string) error
```
SetName sets the name of the service. A service name can not exceed 63 bytes.



### func (\*RegisterOp) SetNoAutoRename
``` go
func (o *RegisterOp) SetNoAutoRename(e bool) error
```
SetNoAutoRename sets how service-name conflicts will be handled.
If set to the default, false, conflicts will be handled automatically be renaming the service (eg: "My Service" will be become "My Service 2" or similar).
If set to true the operations callback will be invoked with an error.



### func (\*RegisterOp) SetPort
``` go
func (o *RegisterOp) SetPort(p int) error
```
SetPort sets the port the service is available from.



### func (\*RegisterOp) SetTXTPair
``` go
func (o *RegisterOp) SetTXTPair(key, value string) error
```
SetTXTPair creates or updates a TXT string with the provided value.



### func (\*RegisterOp) SetType
``` go
func (o *RegisterOp) SetType(s string) error
```
SetType sets the service type associated with the op.



### func (\*RegisterOp) Start
``` go
func (o *RegisterOp) Start() error
```
Start begins advertising the service.



### func (\*RegisterOp) Stop
``` go
func (o *RegisterOp) Stop()
```
Stop stops the operation.



### func (\*RegisterOp) Type
``` go
func (o *RegisterOp) Type() string
```
Type returns the service type associated with the op.



## type ResolveCallbackFunc
``` go
type ResolveCallbackFunc func(op *ResolveOp, err error, host string, port int, txt map[string]string)
```
ResolveCallbackFunc is called when a service is resolved or an error occurs.











## type ResolveOp
``` go
type ResolveOp struct {
    // contains filtered or unexported fields
}
```
ResolveOp represents an operation that resolves a service instance to a host, port and TXT map containing meta data.









### func NewResolveOp
``` go
func NewResolveOp(interfaceIndex int, name, serviceType, domain string, f ResolveCallbackFunc) *ResolveOp
```
NewResolveOp creates a new ResolveOp with the associated parameters set.
It should be called with the parameters supplied to the callback of a browse operation.


### func StartResolveOp
``` go
func StartResolveOp(interfaceIndex int, name, serviceType, domain string, f ResolveCallbackFunc) (*ResolveOp, error)
```
StartResolveOp returns the equivalent of calling NewResolveOp and Start.




### func (\*ResolveOp) Active
``` go
func (o *ResolveOp) Active() bool
```
Active indicates whether an operation is active



### func (\*ResolveOp) Domain
``` go
func (o *ResolveOp) Domain() string
```
Domain returns the domain associated with the op.



### func (\*ResolveOp) InterfaceIndex
``` go
func (o *ResolveOp) InterfaceIndex() int
```
InterfaceIndex returns the interface index the op is tied to.



### func (\*ResolveOp) Name
``` go
func (o *ResolveOp) Name() string
```
Name returns the name of the service.



### func (\*ResolveOp) SetCallback
``` go
func (o *ResolveOp) SetCallback(f ResolveCallbackFunc) error
```
SetCallback sets the function to call when a service is resolved or an error occurs.



### func (\*ResolveOp) SetDomain
``` go
func (o *ResolveOp) SetDomain(s string) error
```
SetDomain sets the domain associated with the op.



### func (\*ResolveOp) SetInterfaceIndex
``` go
func (o *ResolveOp) SetInterfaceIndex(i int) error
```
SetInterfaceIndex sets the interface index the op is tied to.



### func (\*ResolveOp) SetName
``` go
func (o *ResolveOp) SetName(n string) error
```
SetName set's the name of the service.



### func (\*ResolveOp) SetType
``` go
func (o *ResolveOp) SetType(s string) error
```
SetType sets the service type associated with the op.



### func (\*ResolveOp) Start
``` go
func (o *ResolveOp) Start() error
```
Start begins the resolve operation. Resolve operations should be stopped as soon as they are no longer needed.



### func (\*ResolveOp) Stop
``` go
func (o *ResolveOp) Stop()
```
Stop stops the operation.



### func (\*ResolveOp) Type
``` go
func (o *ResolveOp) Type() string
```
Type returns the service type associated with the op.









- - -
Generated by [godoc2md](http://godoc.org/github.com/davecheney/godoc2md)
