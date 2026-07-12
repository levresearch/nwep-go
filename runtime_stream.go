// the managed stream, a streamed response body pulled chunk by chunk NW060200.
//
// a streamed body is too large for one message, and the c stream recv blocks (it
// busy ticks until a chunk or the end), so it cannot share the concurrent request
// loop without stalling it. a RunningStream therefore runs on its own owner
// goroutine over its own dedicated connection (the streaming-response model): the
// owner opens the stream, reads the metadata frame, then loops recv pushing each
// chunk over a channel Recv awaits, and verifies the trailer at the end NW060900.

package nwep

import "runtime"

// streamChunk is the buffer size the owner reads each stream recv into.
const streamChunk = 64 * 1024

// streamItem is one delivered chunk or the error that ended the stream.
type streamItem struct {
	chunk []byte
	err   error
}

// RunningStream is a streamed response body delivered chunk by chunk NW060200.
//
// obtained from ClientBuilder.Stream, with its status and headers already read.
// pull the body with Recv until it returns a nil chunk, then close with Close.
type RunningStream struct {
	status  string
	headers [][2]string
	items   chan streamItem
	stop    chan struct{}
	done    chan struct{}
}

// Status returns the streamed response status, from its leading frame NW080000.
func (s *RunningStream) Status() string { return s.status }

// Header returns a response header value, or empty when absent NW060300.
func (s *RunningStream) Header(name string) string {
	for _, h := range s.headers {
		if h[0] == name {
			return h[1]
		}
	}
	return ""
}

// Headers returns the streamed response headers in wire order NW060300.
func (s *RunningStream) Headers() [][2]string { return s.headers }

// Recv returns the next body chunk, or a nil chunk at the verified end NW060200.
//
// the owner goroutine reads the body and delivers each chunk here. once Recv
// returns a nil chunk and a nil error the body is complete and its trailer
// signature has verified against the peer NW060900, an authentic, whole body.
//
// returns a chunk while streaming, or nil at the verified end.
// errors a transport error mid-stream, or a crypto-verify error when the trailer
// signature does not verify.
func (s *RunningStream) Recv() ([]byte, error) {
	select {
	case item, ok := <-s.items:
		if !ok {
			return nil, nil // the owner ended without an explicit error.
		}
		return item.chunk, item.err
	case <-s.done:
		return nil, nil
	}
}

// Close stops the owner goroutine and tears down the dedicated connection.
//
// safe to call more than once, and safe to call before the body is fully read (it
// abandons the rest).
func (s *RunningStream) Close() {
	select {
	case <-s.done:
		return
	default:
	}
	close(s.stop)
	<-s.done
}

// Stream opens a streamed response over a dedicated connection, the managed
// streaming terminal NW060200.
//
// the counterpart of OpenStream on the driven client, for receiving a body too
// large for one message. it connects on its own owner goroutine, opens the stream,
// reads the metadata frame, and returns a RunningStream whose Recv yields the body
// chunk by chunk. the connection is dedicated to this stream and closes on Close.
//
// returns the open RunningStream, its status and headers already read.
// errors with whatever Connect would, and any transport error opening the stream.
func (b *ClientBuilder) Stream(method Method, path string) (*RunningStream, error) {
	s := &RunningStream{
		items: make(chan streamItem, 8),
		stop:  make(chan struct{}),
		done:  make(chan struct{}),
	}
	type meta struct {
		status  string
		headers [][2]string
		err     error
	}
	ready := make(chan meta, 1)

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		defer close(s.done)
		defer close(s.items)

		client, err := b.Connect()
		if err != nil {
			ready <- meta{err: err}
			return
		}
		defer client.Close()

		streamID, err := client.OpenStream(method, path, nil)
		if err != nil {
			ready <- meta{err: err}
			return
		}
		resp, err := client.StreamResponse(streamID)
		if err != nil {
			ready <- meta{err: err}
			return
		}
		md := meta{status: resp.Status(), headers: resp.Headers()}
		resp.Close()
		ready <- md

		peer, _ := client.PeerPubkey()
		buf := make([]byte, streamChunk)
		for {
			select {
			case <-s.stop:
				return
			default:
			}
			n, ended, err := client.StreamRecv(streamID, buf)
			if err != nil {
				s.deliver(streamItem{err: err})
				return
			}
			if n > 0 {
				chunk := make([]byte, n)
				copy(chunk, buf[:n])
				if !s.deliver(streamItem{chunk: chunk}) {
					return // the consumer closed the stream.
				}
			}
			if ended {
				// the trailer must verify against the peer for the body to count as
				// authentic and whole NW060900. a closing channel signals success.
				if err := client.StreamVerify(streamID, peer); err != nil {
					s.deliver(streamItem{err: err})
				}
				return
			}
		}
	}()

	m := <-ready
	if m.err != nil {
		<-s.done
		return nil, m.err
	}
	s.status, s.headers = m.status, m.headers
	return s, nil
}

// deliver pushes an item to the consumer, returning false if it has closed.
func (s *RunningStream) deliver(item streamItem) bool {
	select {
	case s.items <- item:
		return true
	case <-s.stop:
		return false
	}
}
