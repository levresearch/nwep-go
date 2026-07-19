// the shared error taxonomy, one Error type over the NW130000 code families NWG0800.
//
// the c abi returns stable negative integers. Error preserves the numeric code,
// carries the identical human message from nwep_strerror, classifies the family,
// and marks whether the failure is fatal (the connection is dead) versus
// retryable. match a specific code with errors.Is against the sentinels below.

package nwep

import "github.com/levresearch/nwep-go/sys"

// Family groups the error codes by NW130000 families.
type Family int

// the error families NW130000.
const (
	FamilyConfig Family = iota
	FamilyNetwork
	FamilyCrypto
	FamilyProtocol
	FamilyIdentity
	FamilyApp
	FamilyTrust
	FamilyInternal
)

// String returns the name of the family (e.g. "FamilyCrypto").
func (f Family) String() string {
	switch f {
	case FamilyConfig:
		return "FamilyConfig"
	case FamilyNetwork:
		return "FamilyNetwork"
	case FamilyCrypto:
		return "FamilyCrypto"
	case FamilyProtocol:
		return "FamilyProtocol"
	case FamilyIdentity:
		return "FamilyIdentity"
	case FamilyApp:
		return "FamilyApp"
	case FamilyTrust:
		return "FamilyTrust"
	default:
		return "FamilyInternal"
	}
}

// Error is a failure from the library, a stable code plus its human message.
type Error struct {
	// Code is the raw negative nwep code NW130000, preserved for matching.
	Code int
	msg  string
}

// Error returns the human message, identical across every binding (nwep_strerror).
func (e *Error) Error() string { return e.msg }

// Fatal reports whether the error killed the connection and is not retryable.
//
// the fatal codes NW130000 (equivocation, version, the handshake signature
// failures) mean the connection is dead, callers branch on this to decide
// between tearing down and retrying.
func (e *Error) Fatal() bool { return sys.IsFatal(e.Code) }

// Family returns which NW130000 family the code belongs to.
func (e *Error) Family() Family {
	switch c := -e.Code; {
	case c >= 100 && c < 200:
		return FamilyConfig
	case c >= 200 && c < 300:
		return FamilyNetwork
	case c >= 300 && c < 400:
		return FamilyCrypto
	case c >= 400 && c < 500:
		return FamilyProtocol
	case c >= 500 && c < 600:
		return FamilyIdentity
	case c >= 600 && c < 700:
		return FamilyApp
	case c >= 700 && c < 800:
		return FamilyTrust
	default:
		return FamilyInternal
	}
}

// Is matches by code, so errors.Is(err, nwep.ErrIdentityNotFound) works.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	return ok && t.Code == e.Code
}

// newError builds an Error for a negative code, reading its message from the lib.
func newError(code int) *Error {
	return &Error{Code: code, msg: sys.Strerror(code)}
}

// check turns a c return code into an error, nil for success (0 or positive). only a negative code is an error.
func check(rc int) error {
	if rc >= 0 {
		return nil
	}
	return newError(rc)
}

// the sentinel errors, one per NW130000 code, for matching with errors.Is.
var (
	ErrConfigInvalid = newError(sys.ErrConfigInvalid)
	ErrConfigMissing = newError(sys.ErrConfigMissing)

	ErrNetworkConnect = newError(sys.ErrNetworkConnect)
	ErrNetworkTimeout = newError(sys.ErrNetworkTimeout)
	ErrNetworkClosed  = newError(sys.ErrNetworkClosed)
	ErrNetworkQUIC    = newError(sys.ErrNetworkQUIC)
	ErrNetworkTLS     = newError(sys.ErrNetworkTLS)

	ErrCryptoKeygen = newError(sys.ErrCryptoKeygen)
	ErrCryptoRand   = newError(sys.ErrCryptoRand)
	ErrCryptoSign   = newError(sys.ErrCryptoSign)
	ErrCryptoVerify = newError(sys.ErrCryptoVerify)

	ErrCryptoFatalCert           = newError(sys.ErrCryptoFatalCert)
	ErrCryptoFatalNodeidMismatch = newError(sys.ErrCryptoFatalNodeidMismatch)
	ErrCryptoFatalChallenge      = newError(sys.ErrCryptoFatalChallenge)
	ErrCryptoFatalServerSig      = newError(sys.ErrCryptoFatalServerSig)
	ErrCryptoFatalClientSig      = newError(sys.ErrCryptoFatalClientSig)

	ErrProtoInvalidMessage  = newError(sys.ErrProtoInvalidMessage)
	ErrProtoInvalidMethod   = newError(sys.ErrProtoInvalidMethod)
	ErrProtoInvalidHeader   = newError(sys.ErrProtoInvalidHeader)
	ErrProtoConnectRequired = newError(sys.ErrProtoConnectRequired)
	ErrProtoStreamReuse     = newError(sys.ErrProtoStreamReuse)
	ErrProtoMaxStreams      = newError(sys.ErrProtoMaxStreams)
	ErrProtoFlowControl     = newError(sys.ErrProtoFlowControl)
	ErrProtoMessageTooLarge = newError(sys.ErrProtoMessageTooLarge)

	ErrProtoFatalVersion = newError(sys.ErrProtoFatalVersion)

	ErrIdentityGenerate = newError(sys.ErrIdentityGenerate)
	ErrIdentityMismatch = newError(sys.ErrIdentityMismatch)
	ErrIdentityNotFound = newError(sys.ErrIdentityNotFound)
	ErrIdentityRevoked  = newError(sys.ErrIdentityRevoked)

	ErrAppNotFound    = newError(sys.ErrAppNotFound)
	ErrAppConflict    = newError(sys.ErrAppConflict)
	ErrAppRateLimited = newError(sys.ErrAppRateLimited)
	ErrAppForbidden   = newError(sys.ErrAppForbidden)

	ErrTrustInvalidEntry    = newError(sys.ErrTrustInvalidEntry)
	ErrTrustInvalidAnchor   = newError(sys.ErrTrustInvalidAnchor)
	ErrTrustStaleCheckpoint = newError(sys.ErrTrustStaleCheckpoint)
	ErrTrustThreshold       = newError(sys.ErrTrustThreshold)
	ErrTrustRevoked         = newError(sys.ErrTrustRevoked)
	ErrTrustNoCheckpoint    = newError(sys.ErrTrustNoCheckpoint)

	ErrTrustFatalEquivocation = newError(sys.ErrTrustFatalEquivocation)
	ErrTrustFatalLogCorrupt   = newError(sys.ErrTrustFatalLogCorrupt)

	ErrInternal      = newError(sys.ErrInternal)
	ErrInternalAlloc = newError(sys.ErrInternalAlloc)
	ErrWouldBlock    = newError(sys.ErrWouldBlock)
)
