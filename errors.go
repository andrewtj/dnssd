package dnssd

import (
	"errors"
	"fmt"
)

// ErrStarted is returned when trying to mutate an active operation or when starting a started operation.
var ErrStarted = errors.New("already started")

// ErrMissingCallback is returned when an operation is started without setting a callback.
var ErrMissingCallback = errors.New("no callback set")

// ErrTXTStringLen is returned when setting a TXT pair that would exceed the 255 byte string limit.
var ErrTXTStringLen = errors.New("TXT string may not exceed 255 bytes")

// ErrTXTLen is returned when setting a TXT pair that would exceed the 65,535 byte TXT record limit.
var ErrTXTLen = errors.New("TXT size may not exceed 65535 bytes")

// Error structs meet the error interface and are returned when errors occur in the underlying C API.
type Error struct {
	n int32
	d string
}

func (e Error) Error() string { return fmt.Sprintf("%s (%d)", e.Desc(), e.Num()) }

// Num returns an error number.
func (e Error) Num() int32 { return e.n }

// Desc returns a string describing the error.
func (e Error) Desc() string { return e.d }

// Errors returned by the underlying C API.
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

func getError(n int32) error {
	var m = map[int32]error{
		0:      nil,
		-65537: ErrUnknown,
		-65538: ErrNoSuchName,
		-65539: ErrNoMemory,
		-65540: ErrBadParam,
		-65541: ErrBadReference,
		-65542: ErrBadState,
		-65543: ErrBadFlags,
		-65544: ErrUnsupported,
		-65545: ErrNotInitialized,
		-65547: ErrAlreadyRegistered,
		-65548: ErrNameConflict,
		-65549: ErrInvalid,
		-65550: ErrFirewall,
		-65551: ErrIncompatible,
		-65552: ErrBadInterfaceIndex,
		-65553: ErrRefused,
		-65554: ErrNoSuchRecord,
		-65555: ErrNoAuth,
		-65556: ErrNoSuchKey,
		-65557: ErrNATTraversal,
		-65558: ErrDoubleNAT,
		-65559: ErrBadTime,
		-65560: ErrBadSig,
		-65561: ErrBadKey,
		-65562: ErrTransient,
		-65563: ErrServiceNotRunning,
		-65564: ErrNATPortMappingUnsupported,
		-65565: ErrNATPortMappingDisabled,
		-65566: ErrNoRouter,
		-65567: ErrPollingMode,
		-65568: ErrTimeout,
	}
	if err, ok := m[n]; ok {
		return err
	}
	return Error{n, "Unknown"}
}
