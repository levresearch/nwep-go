// the layer 0 stable error, method, and status codes from the header NW130000 NW040400.

package sys

/*
#include <nwep.h>
*/
import "C"

// the error codes, stable negative integers grouped by family NW130000.
const (
	ErrConfigInvalid = int(C.NWEP_ERR_CONFIG_INVALID)
	ErrConfigMissing = int(C.NWEP_ERR_CONFIG_MISSING)

	ErrNetworkConnect = int(C.NWEP_ERR_NETWORK_CONNECT)
	ErrNetworkTimeout = int(C.NWEP_ERR_NETWORK_TIMEOUT)
	ErrNetworkClosed  = int(C.NWEP_ERR_NETWORK_CLOSED)
	ErrNetworkQUIC    = int(C.NWEP_ERR_NETWORK_QUIC)
	ErrNetworkTLS     = int(C.NWEP_ERR_NETWORK_TLS)

	ErrCryptoKeygen = int(C.NWEP_ERR_CRYPTO_KEYGEN)
	ErrCryptoRand   = int(C.NWEP_ERR_CRYPTO_RAND)
	ErrCryptoSign   = int(C.NWEP_ERR_CRYPTO_SIGN)
	ErrCryptoVerify = int(C.NWEP_ERR_CRYPTO_VERIFY)

	ErrCryptoFatalCert           = int(C.NWEP_ERR_CRYPTO_FATAL_CERT)
	ErrCryptoFatalNodeidMismatch = int(C.NWEP_ERR_CRYPTO_FATAL_NODEID_MISMATCH)
	ErrCryptoFatalChallenge      = int(C.NWEP_ERR_CRYPTO_FATAL_CHALLENGE)
	ErrCryptoFatalServerSig      = int(C.NWEP_ERR_CRYPTO_FATAL_SERVER_SIG)
	ErrCryptoFatalClientSig      = int(C.NWEP_ERR_CRYPTO_FATAL_CLIENT_SIG)

	ErrProtoInvalidMessage  = int(C.NWEP_ERR_PROTO_INVALID_MESSAGE)
	ErrProtoInvalidMethod   = int(C.NWEP_ERR_PROTO_INVALID_METHOD)
	ErrProtoInvalidHeader   = int(C.NWEP_ERR_PROTO_INVALID_HEADER)
	ErrProtoConnectRequired = int(C.NWEP_ERR_PROTO_CONNECT_REQUIRED)
	ErrProtoStreamReuse     = int(C.NWEP_ERR_PROTO_STREAM_REUSE)
	ErrProtoMaxStreams      = int(C.NWEP_ERR_PROTO_MAX_STREAMS)
	ErrProtoFlowControl     = int(C.NWEP_ERR_PROTO_FLOW_CONTROL)
	ErrProtoMessageTooLarge = int(C.NWEP_ERR_PROTO_MESSAGE_TOO_LARGE)

	ErrProtoFatalVersion = int(C.NWEP_ERR_PROTO_FATAL_VERSION)

	ErrIdentityGenerate = int(C.NWEP_ERR_IDENTITY_GENERATE)
	ErrIdentityMismatch = int(C.NWEP_ERR_IDENTITY_MISMATCH)
	ErrIdentityNotFound = int(C.NWEP_ERR_IDENTITY_NOT_FOUND)
	ErrIdentityRevoked  = int(C.NWEP_ERR_IDENTITY_REVOKED)

	ErrAppNotFound    = int(C.NWEP_ERR_APP_NOT_FOUND)
	ErrAppConflict    = int(C.NWEP_ERR_APP_CONFLICT)
	ErrAppRateLimited = int(C.NWEP_ERR_APP_RATE_LIMITED)
	ErrAppForbidden   = int(C.NWEP_ERR_APP_FORBIDDEN)

	ErrTrustInvalidEntry    = int(C.NWEP_ERR_TRUST_INVALID_ENTRY)
	ErrTrustInvalidAnchor   = int(C.NWEP_ERR_TRUST_INVALID_ANCHOR)
	ErrTrustStaleCheckpoint = int(C.NWEP_ERR_TRUST_STALE_CHECKPOINT)
	ErrTrustThreshold       = int(C.NWEP_ERR_TRUST_THRESHOLD)
	ErrTrustRevoked         = int(C.NWEP_ERR_TRUST_REVOKED)
	ErrTrustNoCheckpoint    = int(C.NWEP_ERR_TRUST_NO_CHECKPOINT)

	ErrTrustFatalEquivocation = int(C.NWEP_ERR_TRUST_FATAL_EQUIVOCATION)
	ErrTrustFatalLogCorrupt   = int(C.NWEP_ERR_TRUST_FATAL_LOG_CORRUPT)

	ErrInternal      = int(C.NWEP_ERR_INTERNAL)
	ErrInternalAlloc = int(C.NWEP_ERR_INTERNAL_ALLOC)
	ErrWouldBlock    = int(C.NWEP_ERR_WOULD_BLOCK)
)

// the request method codes NW040400.
const (
	MethodRead      = int(C.NWEP_METHOD_READ)
	MethodWrite     = int(C.NWEP_METHOD_WRITE)
	MethodUpdate    = int(C.NWEP_METHOD_UPDATE)
	MethodDelete    = int(C.NWEP_METHOD_DELETE)
	MethodHeartbeat = int(C.NWEP_METHOD_HEARTBEAT)
	MethodHead      = int(C.NWEP_METHOD_HEAD)
)

// the status codes NW080000. use with nwep_status_str or nwep_message_get_status_code.
const (
	StatusOk                  = int(C.NWEP_STATUS_OK)
	StatusCreated             = int(C.NWEP_STATUS_CREATED)
	StatusAccepted            = int(C.NWEP_STATUS_ACCEPTED)
	StatusNoContent           = int(C.NWEP_STATUS_NO_CONTENT)
	StatusPartialContent      = int(C.NWEP_STATUS_PARTIAL_CONTENT)
	StatusMoved               = int(C.NWEP_STATUS_MOVED)
	StatusNotModified         = int(C.NWEP_STATUS_NOT_MODIFIED)
	StatusBadRequest          = int(C.NWEP_STATUS_BAD_REQUEST)
	StatusUnauthorized        = int(C.NWEP_STATUS_UNAUTHORIZED)
	StatusForbidden           = int(C.NWEP_STATUS_FORBIDDEN)
	StatusNotFound            = int(C.NWEP_STATUS_NOT_FOUND)
	StatusNotAllowed          = int(C.NWEP_STATUS_NOT_ALLOWED)
	StatusConflict            = int(C.NWEP_STATUS_CONFLICT)
	StatusGone                = int(C.NWEP_STATUS_GONE)
	StatusTooLarge            = int(C.NWEP_STATUS_TOO_LARGE)
	StatusPreconditionFailed  = int(C.NWEP_STATUS_PRECONDITION_FAILED)
	StatusRangeNotSatisfiable = int(C.NWEP_STATUS_RANGE_NOT_SATISFIABLE)
	StatusRateLimited         = int(C.NWEP_STATUS_RATE_LIMITED)
	StatusError               = int(C.NWEP_STATUS_ERROR)
	StatusUnavailable         = int(C.NWEP_STATUS_UNAVAILABLE)
	StatusTimeout             = int(C.NWEP_STATUS_TIMEOUT)
	StatusNotImplemented      = int(C.NWEP_STATUS_NOT_IMPLEMENTED)
)

// IsFatal reports whether code kills the connection, not retryable (nwep_is_fatal).
func IsFatal(code int) bool {
	return C.nwep_is_fatal(C.int(code)) != 0
}

// MethodStr returns the human name of a method code (nwep_method_str).
func MethodStr(method int) string {
	return C.GoString(C.nwep_method_str(C.int(method)))
}

// StatusStr returns the human name of a status code (nwep_status_str).
func StatusStr(status int) string {
	return C.GoString(C.nwep_status_str(C.int(status)))
}
