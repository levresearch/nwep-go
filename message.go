// the message types, the read view of a decoded message and the response writer NW050000 NW060000.
//
// Message is a decoded request or response. on the client side a response Message
// is owned and freed with Close. inside a server handler the request Message is
// borrowed for the call and must not be kept. Responder is the handler's output,
// the encoded response written into the library's buffer.

package nwep

import (
	"unsafe"

	"github.com/levresearch/nwep-go/sys"
)

// Status is a WEB/1 response status token NW080000.
//
// the token is the literal ASCII string on the wire; the typed constant exists
// for compile-time safety and autocomplete. use StatusFromString to parse a
// token received in a response.
type Status int

// the full status token set NW080000.
const (
	StatusOk                  Status = Status(sys.StatusOk)
	StatusCreated             Status = Status(sys.StatusCreated)
	StatusAccepted            Status = Status(sys.StatusAccepted)
	StatusNoContent           Status = Status(sys.StatusNoContent)
	StatusPartialContent      Status = Status(sys.StatusPartialContent)
	StatusMoved               Status = Status(sys.StatusMoved)
	StatusNotModified         Status = Status(sys.StatusNotModified)
	StatusBadRequest          Status = Status(sys.StatusBadRequest)
	StatusUnauthorized        Status = Status(sys.StatusUnauthorized)
	StatusForbidden           Status = Status(sys.StatusForbidden)
	StatusNotFound            Status = Status(sys.StatusNotFound)
	StatusNotAllowed          Status = Status(sys.StatusNotAllowed)
	StatusConflict            Status = Status(sys.StatusConflict)
	StatusGone                Status = Status(sys.StatusGone)
	StatusTooLarge            Status = Status(sys.StatusTooLarge)
	StatusPreconditionFailed  Status = Status(sys.StatusPreconditionFailed)
	StatusRangeNotSatisfiable Status = Status(sys.StatusRangeNotSatisfiable)
	StatusRateLimited         Status = Status(sys.StatusRateLimited)
	StatusError               Status = Status(sys.StatusError)
	StatusUnavailable         Status = Status(sys.StatusUnavailable)
	StatusTimeout             Status = Status(sys.StatusTimeout)
	StatusNotImplemented      Status = Status(sys.StatusNotImplemented)
)

// String returns the wire token string for this status (e.g. "not-found").
func (s Status) String() string { return sys.StatusStr(int(s)) }

// IsSuccess reports whether the status represents a successful outcome.
func (s Status) IsSuccess() bool {
	return s == StatusOk || s == StatusCreated || s == StatusAccepted ||
		s == StatusNoContent || s == StatusNotModified || s == StatusPartialContent
}

// StatusFromString parses a wire token into a Status. unknown tokens return StatusError
// per NW080000.
func StatusFromString(token string) Status {
	for code := 0; code <= sys.StatusNotImplemented; code++ {
		if sys.StatusStr(code) == token {
			return Status(code)
		}
	}
	return StatusError
}

// Message is a decoded web/1 message, headers, status, and body NW050000.
type Message struct {
	ptr   unsafe.Pointer
	owned bool
	// connID and streamID identify the connection and quic stream a server
	// request arrived on, for deferring or streaming the reply. zero on a client
	// response, where they have no meaning.
	connID   uint64
	streamID uint64
}

// ConnID returns the connection a server request arrived on, for defer or stream.
func (m *Message) ConnID() uint64 { return m.connID }

// StreamID returns the quic stream a server request arrived on NW060200.
func (m *Message) StreamID() uint64 { return m.streamID }

// Header returns the value of header name, or empty when absent.
func (m *Message) Header(name string) string { return sys.MessageGetHeader(m.ptr, name) }

// Headers returns every header as name/value pairs, pseudo-headers included.
func (m *Message) Headers() [][2]string {
	n := sys.MessageHeaderCount(m.ptr)
	out := make([][2]string, 0, n)
	for i := 0; i < n; i++ {
		name, value, rc := sys.MessageHeaderAt(m.ptr, i)
		if rc != 0 {
			break
		}
		out = append(out, [2]string{name, value})
	}
	return out
}

// Status returns the response status string, empty for a request NW060100.
func (m *Message) Status() string { return sys.MessageGetStatus(m.ptr) }

// Body returns the decoded body bytes, nil for an empty body.
func (m *Message) Body() []byte { return sys.MessageGetBody(m.ptr) }

// Path returns the request path pseudo-header NW040400.
func (m *Message) Path() string { return m.Header(":path") }

// Method returns the request method pseudo-header NW040400.
func (m *Message) Method() string { return m.Header(":method") }

// Event returns the NOTIFY event pseudo-header, for a notify message NW060500.
func (m *Message) Event() string { return m.Header(":event") }

// Close frees an owned message, a no-op for a borrowed request (nwep_message_free).
func (m *Message) Close() {
	if m.owned && m.ptr != nil {
		sys.MessageFree(m.ptr)
		m.ptr = nil
	}
}

// Raw returns the underlying sys message pointer, the no-cliffs escape to L0 NWG0200.
func (m *Message) Raw() unsafe.Pointer { return m.ptr }

// Responder writes a handler's response into the library's output buffer NW060000.
//
// the helpers are chainable and record the first failure, retrievable with Err.
// call exactly one terminal helper (OK, Status, NotModified, and so on) unless
// the handler defers with Defer to answer out of band later.
type Responder struct {
	buf unsafe.Pointer
	err error
	// deferred is set by Defer (answer out of band later). streamed is set by
	// Stream (the reply is a server-pushed stream, already begun). server, connID,
	// and streamID let Stream call begin_stream for this request.
	deferred bool
	streamed bool
	server   unsafe.Pointer
	connID   uint64
	streamID uint64
}

// set records the first error from a builder call.
func (r *Responder) set(rc int) *Responder {
	if r.err == nil {
		r.err = check(rc)
	}
	return r
}

// OK writes an ok response carrying body (nwep_response_ok).
func (r *Responder) OK(body []byte) *Responder { return r.set(sys.ResponseOk(r.buf, body)) }

// Respond writes a response with an explicit Status and body (nwep_response_status).
func (r *Responder) Respond(status Status, body []byte) *Responder {
	return r.set(sys.ResponseStatus(r.buf, status.String(), body))
}

// Created writes a created response with body (write succeeded).
func (r *Responder) Created(body []byte) *Responder {
	return r.set(sys.ResponseStatus(r.buf, "created", body))
}

// Accepted writes an accepted response (request received, processing async).
func (r *Responder) Accepted(body []byte) *Responder {
	return r.set(sys.ResponseStatus(r.buf, "accepted", body))
}

// NoContent writes a no-content response (success, no body).
func (r *Responder) NoContent() *Responder {
	return r.set(sys.ResponseStatus(r.buf, "no-content", nil))
}

// Moved writes a moved response with a location header pointing to the new web:// URI.
func (r *Responder) Moved(location string) *Responder {
	return r.Header("location", location).set(sys.ResponseStatus(r.buf, "moved", nil))
}

// NotModified writes a not-modified response carrying etag (nwep_response_not_modified).
func (r *Responder) NotModified(etag string) *Responder {
	return r.set(sys.ResponseNotModified(r.buf, etag))
}

// BadRequest writes a bad-request response with an optional body.
func (r *Responder) BadRequest(body []byte) *Responder {
	return r.set(sys.ResponseStatus(r.buf, "bad-request", body))
}

// Unauthorized writes an unauthorized response.
func (r *Responder) Unauthorized(body []byte) *Responder {
	return r.set(sys.ResponseStatus(r.buf, "unauthorized", body))
}

// Forbidden writes a forbidden response.
func (r *Responder) Forbidden(body []byte) *Responder {
	return r.set(sys.ResponseStatus(r.buf, "forbidden", body))
}

// NotFound writes a not-found response.
func (r *Responder) NotFound() *Responder {
	return r.set(sys.ResponseStatus(r.buf, "not-found", nil))
}

// NotAllowed writes a not-allowed response (method not permitted on this resource).
func (r *Responder) NotAllowed() *Responder {
	return r.set(sys.ResponseStatus(r.buf, "not-allowed", nil))
}

// Conflict writes a conflict response.
func (r *Responder) Conflict(body []byte) *Responder {
	return r.set(sys.ResponseStatus(r.buf, "conflict", body))
}

// Gone writes a gone response (resource permanently removed).
func (r *Responder) Gone() *Responder {
	return r.set(sys.ResponseStatus(r.buf, "gone", nil))
}

// TooLarge writes a too-large response (request body exceeded server limit).
func (r *Responder) TooLarge() *Responder {
	return r.set(sys.ResponseStatus(r.buf, "too-large", nil))
}

// PreconditionFailed writes a precondition-failed response.
func (r *Responder) PreconditionFailed() *Responder {
	return r.set(sys.ResponseStatus(r.buf, "precondition-failed", nil))
}

// Partial writes a partial-content response for the given ranges (nwep_response_partial).
func (r *Responder) Partial(body []byte, ranges []sys.Range, contentType string) *Responder {
	return r.set(sys.ResponsePartial(r.buf, body, ranges, contentType))
}

// RangeNotSatisfiable writes a range-not-satisfiable response with the resource length.
func (r *Responder) RangeNotSatisfiable(totalLen uint64) *Responder {
	return r.set(sys.ResponseRangeNotSatisfiable(r.buf, totalLen))
}

// RateLimited writes a rate-limited response; retryAfter is seconds until the client may retry.
func (r *Responder) RateLimited(retryAfter string) *Responder {
	return r.Header("retry-after", retryAfter).set(sys.ResponseStatus(r.buf, "rate-limited", nil))
}

// ServerError writes an error response (internal server error) with an optional body.
func (r *Responder) ServerError(body []byte) *Responder {
	return r.set(sys.ResponseStatus(r.buf, "error", body))
}

// Unavailable writes an unavailable response.
func (r *Responder) Unavailable() *Responder {
	return r.set(sys.ResponseStatus(r.buf, "unavailable", nil))
}

// Timeout writes a timeout response (server took too long).
func (r *Responder) Timeout() *Responder {
	return r.set(sys.ResponseStatus(r.buf, "timeout", nil))
}

// NotImplemented writes a not-implemented response.
func (r *Responder) NotImplemented() *Responder {
	return r.set(sys.ResponseStatus(r.buf, "not-implemented", nil))
}

// Header appends one header to the response being built (nwep_response_header).
func (r *Responder) Header(name, value string) *Responder {
	return r.set(sys.ResponseHeader(r.buf, name, value))
}

// Relay writes an upstream response back out verbatim (nwep_response_relay).
func (r *Responder) Relay(origin *Message) *Responder {
	return r.set(sys.ResponseRelay(r.buf, origin.ptr))
}

// Defer answers this request out of band later instead of now (NWEP_DEFER).
//
// the handler writes nothing, and the application delivers the response from its
// event loop with Server.Respond or Server.Relay NW000017. use it when the
// answer depends on a backend fetch you do not want to block the tick on.
func (r *Responder) Defer() { r.deferred = true }

// Stream begins a server-pushed streamed response and returns it open NW060200.
//
// it emits the leading metadata frame (status and headers) now, then the
// application pushes the body across later ticks with Server.StreamSend and ends
// it with Server.StreamEnd, addressing the stream by the request's ConnID and
// StreamID. use it for a body too large for one message. requires a server-side
// responder (a handler's res), not a captured buffer.
func (r *Responder) Stream(path, status string, headers [][2]string) *Responder {
	if r.server == nil {
		r.err = ErrConfigInvalid
		return r
	}
	r.streamed = true
	return r.set(sys.ServerBeginStream(r.server, r.connID, r.streamID, path, status, headers))
}

// Err returns the first error a builder helper hit, or nil.
func (r *Responder) Err() error { return r.err }

// Raw returns the underlying sys buf pointer, the no-cliffs escape to L0 NWG0200.
func (r *Responder) Raw() unsafe.Pointer { return r.buf }
