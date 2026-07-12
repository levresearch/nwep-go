#ifndef NWEP_H
#define NWEP_H

#include <stddef.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

/*
 * threading contract
 *
 * all nwep_* functions are non-reentrant and not thread-safe. drive every
 * handle from a single thread. tick functions advance state machines  -
 * call them from your event loop on every I/O event or timer expiry.
 *
 * reentrancy from a dispatch handler: a handler runs synchronously inside
 * nwep_server_tick. from within it you may call the response builders
 * (nwep_response_*) and nwep_server_notify on the same connection. you must
 * not, from within a handler, call nwep_server_tick (reentrant tick corrupts
 * QUIC/connection state) or nwep_server_close / nwep_dht_close (these free the
 * very connection/server being dispatched  -  a use-after-free on return).
 * defer teardown to after tick returns.
 *
 * pointers borrowed from a message handle (header values, body) are valid
 * only while the message handle itself is alive.
 *
 * memory ownership is documented per-function; the doc-comments are
 * authoritative.
 */

/* version */
#define NWEP_VERSION_MAJOR 0
#define NWEP_VERSION_MINOR 1
#define NWEP_VERSION_PATCH 0

/* returns the library version string, e.g. "0.1.0". static storage; do not
 * free. */
const char *nwep_version(void);

/* protocol constants NW140000 */
/* transport */
#define NWEP_DEFAULT_PORT                6937

/* message limits */
#define NWEP_MAX_HEADERS                 128
#define NWEP_MAX_HEADER_NAME             256
#define NWEP_MAX_HEADER_VALUE            8192
#define NWEP_MAX_PATH                    2048
#define NWEP_MAX_MESSAGE_SIZE            25165824  /* 24 MiB */
#define NWEP_MAX_HANDSHAKE_MSG_SIZE      8192

/* connection */
#define NWEP_MAX_STREAMS                 65535
#define NWEP_DEFAULT_MAX_STREAMS         100
#define NWEP_HANDSHAKE_TIMEOUT_MS        10000
/* NWEP_DEFAULT_STREAM_TIMEOUT_MS is the WEB/1 application-layer stream
 * inactivity timeout NW140000. NWEP_QUIC_IDLE_TIMEOUT_MS is the QUIC
 * transport idle timeout  -  a different layer. do NOT use the QUIC timeout
 * to implement stream timeouts. */
#define NWEP_DEFAULT_STREAM_TIMEOUT_MS   30000
#define NWEP_QUIC_IDLE_TIMEOUT_MS        120000

/* identity */
#define NWEP_NODEID_SIZE                 32
#define NWEP_PUBKEY_SIZE                 32
#define NWEP_PRIVKEY_SIZE                32
#define NWEP_CHALLENGE_NONCE_SIZE        32

/* trust */
#define NWEP_CHECKPOINT_EPOCH_SECS       3600
#define NWEP_CHECKPOINT_WARN_SECS        86400    /* 24 hours */
#define NWEP_MAX_ANCHORS                 64

/* DHT NW180000 NW140000 */
#define NWEP_DHT_BUCKET_SIZE             20       /* k */
#define NWEP_DHT_CONCURRENCY             3        /* alpha */
#define NWEP_DHT_RECORD_TTL_SECS         3600
#define NWEP_DHT_REPUBLISH_INTERVAL_SECS 1800
#define NWEP_DHT_MAX_RECORD_AGE_SECS     3600
#define NWEP_DHT_MAX_DATAGRAM            1200

/* error codes NW130000 */
#define NWEP_OK                                  0

/* config (1xx) */
#define NWEP_ERR_CONFIG_INVALID                (-101)
#define NWEP_ERR_CONFIG_MISSING                (-102)

/* network (2xx) */
#define NWEP_ERR_NETWORK_CONNECT               (-201)
#define NWEP_ERR_NETWORK_TIMEOUT               (-202)
#define NWEP_ERR_NETWORK_CLOSED                (-203)
#define NWEP_ERR_NETWORK_QUIC                  (-204)
#define NWEP_ERR_NETWORK_TLS                   (-205)

/* crypto non-fatal (3xx) */
#define NWEP_ERR_CRYPTO_KEYGEN                 (-301)
#define NWEP_ERR_CRYPTO_RAND                   (-302)
#define NWEP_ERR_CRYPTO_SIGN                   (-303)
#define NWEP_ERR_CRYPTO_VERIFY                 (-304)

/* crypto fatal (3xx)  -  silent connection close.
 * CRYPTO_FATAL_CERT (-381): TLS certificate parse/decode failure during the
 *   QUIC handshake. */
#define NWEP_ERR_CRYPTO_FATAL_CERT             (-381)
#define NWEP_ERR_CRYPTO_FATAL_NODEID_MISMATCH  (-382)
#define NWEP_ERR_CRYPTO_FATAL_CHALLENGE        (-383)
#define NWEP_ERR_CRYPTO_FATAL_SERVER_SIG       (-384)
#define NWEP_ERR_CRYPTO_FATAL_CLIENT_SIG       (-385)

/* protocol (4xx) */
#define NWEP_ERR_PROTO_INVALID_MESSAGE         (-401)
#define NWEP_ERR_PROTO_INVALID_METHOD          (-402)
#define NWEP_ERR_PROTO_INVALID_HEADER          (-403)
#define NWEP_ERR_PROTO_CONNECT_REQUIRED        (-404)
#define NWEP_ERR_PROTO_STREAM_REUSE            (-405)
#define NWEP_ERR_PROTO_MAX_STREAMS             (-406)
#define NWEP_ERR_PROTO_FLOW_CONTROL            (-407)
#define NWEP_ERR_PROTO_MESSAGE_TOO_LARGE       (-408)
#define NWEP_ERR_PROTO_FATAL_VERSION           (-481)

/* identity (5xx) */
#define NWEP_ERR_IDENTITY_GENERATE             (-501)
#define NWEP_ERR_IDENTITY_MISMATCH             (-502)
#define NWEP_ERR_IDENTITY_NOT_FOUND            (-503)
#define NWEP_ERR_IDENTITY_REVOKED              (-504)

/* application (6xx) */
#define NWEP_ERR_APP_NOT_FOUND                 (-601)
#define NWEP_ERR_APP_CONFLICT                  (-602)
#define NWEP_ERR_APP_RATE_LIMITED              (-603)
#define NWEP_ERR_APP_FORBIDDEN                 (-604)

/* trust non-fatal (7xx) */
#define NWEP_ERR_TRUST_INVALID_ENTRY           (-701)
#define NWEP_ERR_TRUST_INVALID_ANCHOR          (-702)
#define NWEP_ERR_TRUST_STALE_CHECKPOINT        (-703)
#define NWEP_ERR_TRUST_THRESHOLD               (-704)
#define NWEP_ERR_TRUST_REVOKED                 (-705)
#define NWEP_ERR_TRUST_NO_CHECKPOINT           (-706)

/* trust fatal (7xx) */
#define NWEP_ERR_TRUST_FATAL_EQUIVOCATION      (-781)
#define NWEP_ERR_TRUST_FATAL_LOG_CORRUPT       (-782)

/* internal (8xx) */
#define NWEP_ERR_INTERNAL                      (-801)
#define NWEP_ERR_INTERNAL_ALLOC                (-802)
#define NWEP_ERR_WOULD_BLOCK                   (-803)  /* async op not ready yet */

/* true if `err` is a fatal protocol error requiring silent connection close. */
static inline int nwep_is_fatal(int err) {
    return err == NWEP_ERR_CRYPTO_FATAL_CERT
        || err == NWEP_ERR_CRYPTO_FATAL_NODEID_MISMATCH
        || err == NWEP_ERR_CRYPTO_FATAL_CHALLENGE
        || err == NWEP_ERR_CRYPTO_FATAL_SERVER_SIG
        || err == NWEP_ERR_CRYPTO_FATAL_CLIENT_SIG
        || err == NWEP_ERR_PROTO_FATAL_VERSION
        || err == NWEP_ERR_TRUST_FATAL_EQUIVOCATION
        || err == NWEP_ERR_TRUST_FATAL_LOG_CORRUPT;
}

/* returns a static human-readable string for the given error code. never
 * NULL; unknown codes return "unknown". the returned pointer must not be
 * freed. */
const char *nwep_strerror(int err);

/* zeroizes `len` bytes at `ptr` so the compiler cannot optimize the write
 * away. must be called on secret-key memory (Ed25519 privkey, BLS secret
 * key) before passing the buffer to free(). NULL `ptr` is a safe no-op. */
void nwep_zeroize(void *ptr, size_t len);

/* identity */
typedef struct {
    uint8_t bytes[NWEP_NODEID_SIZE];
} nwep_node_id;

typedef struct {
    uint8_t pub_[NWEP_PUBKEY_SIZE];
    uint8_t priv_[NWEP_PRIVKEY_SIZE];
} nwep_keypair;

/* generates a fresh Ed25519 identity. returns 0 on success or a negative
 * NWEP_ERR_*. the caller must zeroize out_kp->priv_ via nwep_zeroize
 * before discarding it. */
int nwep_identity_generate(nwep_node_id *out_id, nwep_keypair *out_kp);

/* verifies a NodeID is the SHA-256(pubkey || "WEB/1") binding of the given
 * pubkey. constant-time comparison. returns 0 on match, negative NWEP_ERR_*
 * on mismatch or NULL args. */
int nwep_nodeid_verify(const nwep_node_id *id, const uint8_t pubkey[NWEP_PUBKEY_SIZE]);

/* encodes a NodeID as a Base58 string. *outlen is the buffer capacity on
 * input and the byte count actually written on output. returns 0 or negative
 * NWEP_ERR_* (NWEP_ERR_CONFIG_INVALID if out or outlen is NULL). */
int nwep_nodeid_to_base58(char *out, size_t *outlen, const nwep_node_id *id);

/* decodes a Base58 string into a NodeID. */
int nwep_nodeid_from_base58(nwep_node_id *out, const char *str, size_t len);

/* derives a NodeID from an Ed25519 public key: SHA-256(pubkey || "WEB/1").
 * use this to recover the NodeID of a keypair loaded via nwep_keypair_load_pem,
 * whose bytes alone do not carry it. */
int nwep_nodeid_from_pubkey(nwep_node_id *out, const uint8_t pubkey[NWEP_PUBKEY_SIZE]);

/* signs msg with an Ed25519 private key; out_sig receives the 64-byte
 * signature. for C embedders authenticating their own server payloads (e.g.
 * a log server signing no-revocation assertions). returns 0 on success or
 * NWEP_ERR_CONFIG_INVALID if msg is NULL. */
int nwep_ed25519_sign(uint8_t out_sig[64], const uint8_t *msg, size_t msg_len,
                      const uint8_t privkey[NWEP_PRIVKEY_SIZE]);

/* verifies a 64-byte Ed25519 signature over msg under pubkey. 0 if valid,
 * negative NWEP_ERR_* otherwise. */
int nwep_ed25519_verify(const uint8_t sig[64], const uint8_t *msg, size_t msg_len,
                        const uint8_t pubkey[NWEP_PUBKEY_SIZE]);

/* key I/O (in-memory PEM only  -  the library never touches the FS) */

/* encodes a keypair to unencrypted PKCS#8 PEM. *outlen is the buffer
 * capacity on input and the byte count actually written on output. returns 0
 * or negative NWEP_ERR_* (NWEP_ERR_CONFIG_INVALID if out or outlen is NULL). */
int nwep_keypair_save_pem(uint8_t *out, size_t *outlen, const nwep_keypair *kp);

/* decodes PEM bytes into a keypair. the decoded privkey is secret material;
 * zero it via nwep_zeroize before disposal. */
int nwep_keypair_load_pem(nwep_keypair *out_kp, const uint8_t *pem, size_t len);

/* shamir secret sharing NW150400.
 *
 * split a secret (intended use: the offline recovery private key) into n
 * shares so any t reconstruct it and t-1 reveal nothing. each share is
 * 1 + secret_len bytes (a 1-based index byte + data). shares and reconstructed
 * secrets are key material  -  nwep_zeroize() them before freeing. */

/* splits `secret` into n shares (2 <= t <= n <= 255), written contiguously to
 * `out` as n blobs of (1 + secret_len) bytes. two-call sizing: out == NULL
 * writes the required size (n * (1 + secret_len)) to *outlen and returns 0; a
 * too-small buffer reports the size and returns NWEP_ERR_INTERNAL. */
int nwep_shamir_split(const uint8_t *secret, size_t secret_len,
                      size_t t, size_t n, uint8_t *out, size_t *outlen);

/* reconstructs a secret from n_shares shares (each share_len bytes, packed
 * contiguously). writes share_len - 1 bytes to out_secret; *out_secret_len
 * must be >= that and receives the count written. returns 0 or a negative
 * NWEP_ERR_* (NWEP_ERR_CONFIG_INVALID for duplicate indices or length
 * mismatch; NWEP_ERR_INTERNAL if *out_secret_len is too small, in which case
 * *out_secret_len is updated to the required size). at least t shares are
 * required for a correct result. */
int nwep_shamir_combine(const uint8_t *shares, size_t n_shares, size_t share_len,
                        uint8_t *out_secret, size_t *out_secret_len);

/* uri */
typedef struct {
    nwep_node_id node_id;
    uint16_t     port;
    /* borrowed slice into the caller's input buffer  -  do not free. valid
     * only while the input buffer is alive. */
    const char  *path;
    size_t       path_len;
} nwep_uri;

/* parses a "web://nodeid_base58[:port]/path" URI. */
int nwep_uri_parse(nwep_uri *out, const char *input, size_t len);

/* method codes (for nwep_client_send) */
#define NWEP_METHOD_READ         0
#define NWEP_METHOD_WRITE        1
#define NWEP_METHOD_UPDATE       2
#define NWEP_METHOD_DELETE       3
#define NWEP_METHOD_HEARTBEAT    6
#define NWEP_METHOD_HEAD         7

/* status codes NW080000.
 * integer handles for the status token set. pass to nwep_status_str to get
 * the wire token string, or compare against nwep_message_get_status_code. */
#define NWEP_STATUS_OK                    0
#define NWEP_STATUS_CREATED               1
#define NWEP_STATUS_ACCEPTED              2
#define NWEP_STATUS_NO_CONTENT            3
#define NWEP_STATUS_PARTIAL_CONTENT       4
#define NWEP_STATUS_MOVED                 5
#define NWEP_STATUS_NOT_MODIFIED          6
#define NWEP_STATUS_BAD_REQUEST           7
#define NWEP_STATUS_UNAUTHORIZED          8
#define NWEP_STATUS_FORBIDDEN             9
#define NWEP_STATUS_NOT_FOUND             10
#define NWEP_STATUS_NOT_ALLOWED           11
#define NWEP_STATUS_CONFLICT              12
#define NWEP_STATUS_GONE                  13
#define NWEP_STATUS_TOO_LARGE             14
#define NWEP_STATUS_PRECONDITION_FAILED   15
#define NWEP_STATUS_RANGE_NOT_SATISFIABLE 16
#define NWEP_STATUS_RATE_LIMITED          17
#define NWEP_STATUS_ERROR                 18
#define NWEP_STATUS_UNAVAILABLE           19
#define NWEP_STATUS_TIMEOUT               20
#define NWEP_STATUS_NOT_IMPLEMENTED       21

/* address */

/* opaque IPv6 socket address. WEB/1 is IPv6-only; IPv4 callers use
 * nwep_address_ipv4_mapped to embed a v4 address into the ::ffff:a.b.c.d
 * IPv4-mapped IPv6 form (RFC 4291 2.5.5.2). */
typedef struct {
    /* opaque storage  -  never inspect these bytes directly. sized to hold
     * a sockaddr_in6 (28 bytes on every supported platform). */
    uint8_t opaque[32];
} nwep_address;

/* `::1` loopback at the given port. */
void nwep_address_loopback(nwep_address *out, uint16_t port);

/* `::` wildcard (all interfaces) at the given port. */
void nwep_address_wildcard(nwep_address *out, uint16_t port);

/* `::ffff:a.b.c.d` IPv4-mapped IPv6 helper. */
void nwep_address_ipv4_mapped(nwep_address *out, uint8_t a, uint8_t b,
                              uint8_t c, uint8_t d, uint16_t port);

/* builds an address from a 16-byte raw IPv6 address (network order) + port. */
void nwep_address_from_bytes(nwep_address *out,
                             const uint8_t addr[16], uint16_t port);

/* returns the host-order port. */
uint16_t nwep_address_get_port(const nwep_address *addr);

/* client */

/* opaque client handle. one handle drives one outbound connection. */
typedef struct nwep_client nwep_client;

/* opaque decoded response message. */
typedef struct nwep_message nwep_message;

/* forward declaration so nwep_client_connect_by_nodeid below can take a
 * DHT handle. the full typedef is in the DHT section. */
typedef struct nwep_dht nwep_dht;

/* opens a connection to `target_addr` and pumps `tick()` until either the
 * handshake completes or NWEP_HANDSHAKE_TIMEOUT_MS expires (blocking).
 * on success `*out` owns the connection; caller frees via
 * nwep_client_close. */
int nwep_client_connect(nwep_client **out,
                        const nwep_keypair *identity,
                        const nwep_node_id *target_node_id,
                        const nwep_address *target_addr);

/* like nwep_client_connect, but adopts a caller-created UDP socket instead of
 * binding an ephemeral one (NW000017 caller-owned I/O  -  lets the client
 * share a poll set, or a SO_REUSEPORT port table, with the caller's other
 * sockets). ownership of `fd` transfers to the client: it is closed by
 * nwep_client_close and on a failure here. the socket must be AF_INET6 + UDP;
 * bind it first for a fixed local port, otherwise the OS assigns one on first
 * send. `fd` is uintptr_t to hold both a POSIX int fd and a Windows SOCKET.
 * still blocks pumping the handshake; the caller-owned socket is for unifying
 * the poll set after connect (via nwep_client_fd). */
int nwep_client_connect_fd(nwep_client **out,
                           const nwep_keypair *identity,
                           const nwep_node_id *target_node_id,
                           const nwep_address *target_addr,
                           uintptr_t fd);

/* non-blocking connect (NW000017). returns immediately (rc 0) with the
 * client in the handshaking state  -  it never blocks pumping the handshake. add
 * nwep_client_fd to your poller, drive nwep_client_tick on each readiness event
 * / timer (timeout from nwep_client_next_timeout_ms), and call
 * nwep_client_connect_poll to learn when the handshake finishes. the blocking
 * nwep_client_connect is this start plus an internal pump. on a non-zero rc
 * *out is untouched. */
int nwep_client_connect_async(nwep_client **out,
                              const nwep_keypair *identity,
                              const nwep_node_id *target_node_id,
                              const nwep_address *target_addr);

/* non-blocking nwep_client_connect_fd: adopts the caller-owned socket and
 * returns a handshaking client (poll via nwep_client_connect_poll). ownership
 * of `fd` transfers to the client. */
int nwep_client_connect_fd_async(nwep_client **out,
                                 const nwep_keypair *identity,
                                 const nwep_node_id *target_node_id,
                                 const nwep_address *target_addr,
                                 uintptr_t fd);

/* polls an async connect started by the two functions above. tick first, then:
 *    1  handshake complete  -  the client is ready for nwep_client_send/_submit
 *    0  still handshaking  -  keep ticking
 *   <0  handshake failed (NWEP_ERR_*)  -  the client is closed; free with
 *       nwep_client_close. the specific reason is surfaced when known, else
 *       NWEP_ERR_NETWORK_CLOSED. returns -801 on a NULL handle. */
int nwep_client_connect_poll(nwep_client *client);

/* resolves `target_node_id` to a transport address via the attached DHT,
 * then opens a connection (blocking until handshake completes or the
 * timeout expires).
 *
 * the DHT must be attached to a server (nwep_dht_attach). this call drives
 * that server's tick and the DHT timers itself while it waits, so  -  unlike
 * earlier versions  -  it does not require a separate thread to be ticking the
 * server. for the same reason, do not tick this DHT's server concurrently from
 * another thread for the duration of the call (the DHT keeps no internal lock;
 * single-threaded use is the contract).
 *
 * resolution proceeds in two stages:
 *
 *   1. fast path  -  read the DHT's local store. if a record for
 *      `target_node_id` is already cached (e.g. because we announced
 *      that node locally), we use its address immediately.
 *
 *   2. slow path  -  run an iterative Kademlia FIND_VALUE lookup NW110800
 *      for up to `lookup_timeout_ms`, pumping the shared UDP socket and the
 *      DHT timers on the calling thread until the record resolves.
 *
 * returns NWEP_ERR_IDENTITY_NOT_FOUND if the lookup timed out without a
 * hit. connection-establishment errors after the address is known are
 * surfaced exactly as nwep_client_connect would surface them.
 *
 * this call blocks (it pumps the lookup internally). for an event loop, compose
 * the async pieces instead  -  nwep_dht_start_lookup + nwep_dht_lookup_result to
 * resolve, then nwep_client_connect_async(.., &record.addr) to connect  -  all
 * driven from your own loop. see tests/c_async_dht_connect.c for the pattern. */
int nwep_client_connect_by_nodeid(nwep_client **out,
                                  const nwep_keypair *identity,
                                  const nwep_node_id *target_node_id,
                                  nwep_dht *dht,
                                  uint32_t lookup_timeout_ms);

/* sends a single request and waits (blocking) for the response. method is
 * one of NWEP_METHOD_*. path must be NUL-terminated and obey NW040400
 * (begins with `/`, no `..`, no encoded `/`).
 *
 * `headers` points to a null-terminated array of {name, value} pairs (the
 * sentinel is a struct whose name and value are both NULL). pass NULL for
 * no extra headers.
 *
 * on success `*out_response` owns a decoded message; caller frees via
 * nwep_message_free. */
typedef struct {
    const char *name;
    const char *value;
} nwep_header;

int nwep_client_send(nwep_client *client,
                     int method,
                     const char *path,
                     const nwep_header *headers,
                     const uint8_t *body, size_t body_len,
                     nwep_message **out_response);

/* async event-loop integration NW000017.
 *
 * the blocking nwep_client_send above is a thin wrapper over the primitives
 * here. for an event-loop application (a proxy, a gateway, a fan-out client),
 * drive the client like the server: add nwep_client_fd to your poller, and on
 * every readiness event or timer expiry call nwep_client_tick.
 *
 *     int cfd = nwep_client_fd(cli);
 *     for (;;) {
 *         int t = nwep_client_next_timeout_ms(cli, now_ms());
 *         poll(&(struct pollfd){.fd=cfd, .events=POLLIN}, 1, t);
 *         nwep_client_tick(cli, now_ms());
 *         // ... drive request completions (poll or done-callback) ...
 *     }
 */

/* the client's UDP socket fd, for your poller (EPOLLIN). drive all I/O through
 * nwep_client_tick  -  never read/write it directly. -1 on a NULL handle. */
intptr_t nwep_client_fd(const nwep_client *client);

/* advance client state: read datagrams, run QUIC timers, complete in-flight
 * async requests (firing the done-callback), flush output. returns 0 or a
 * negative NWEP_ERR_*. now_ms is a monotonic clock (same as the server). */
int nwep_client_tick(nwep_client *client, int64_t now_ms);

/* milliseconds until nwep_client_tick must next run (QUIC timer or the soonest
 * request deadline). -1 = none; 0 = already due. fold into your poll timeout. */
int nwep_client_next_timeout_ms(nwep_client *client, int64_t now_ms);

/* whether the connection is still usable (1) or has terminally closed (0)  -  an
 * idle timeout, a peer CONNECTION_CLOSE, or a fatal QUIC error. a closed client
 * ticks as a cheap no-op and reports no timer (next_timeout_ms = -1), so an
 * event loop never busy-spins on a dead connection; any in-flight async requests
 * fail with NWEP_ERR_NETWORK_CLOSED. poll this to drive reconnection of a
 * persistent connection (e.g. a proxy's origin link). -1 on a NULL handle. */
int nwep_client_is_alive(const nwep_client *client);

/* observability snapshot for one client (NW000017). cumulative counters
 * plus two gauges (requests_inflight, alive) and the connection's smoothed RTT.
 * pull model: read whenever you scrape metrics; nwep keeps no exporter and no
 * history. */
typedef struct {
  uint64_t requests_inflight;  /* gauge: submitted but not yet terminal */
  uint64_t requests_completed; /* cumulative: finished with a response */
  uint64_t requests_failed;    /* cumulative: timed out / closed / errored */
  uint64_t smoothed_rtt_us;    /* ngtcp2 smoothed RTT, microseconds (0 if down) */
  int32_t alive;               /* 1 if the connection is usable, else 0 */
} nwep_client_metrics;

/* fills *out with the client's current metrics snapshot. returns 0, or
 * NWEP_ERR_INTERNAL on a NULL client or out. call from the client's owning
 * thread  -  the one that runs nwep_client_tick. the counters are non-atomic and
 * the gauges walk live state at read time, so a cross-thread scrape races the
 * tick. fold the scrape into your event loop. */
int nwep_client_metrics_get(const nwep_client *client, nwep_client_metrics *out);

/* opaque per-request token, unique within a client. 0 is never a valid id. */
typedef uint64_t nwep_request_id;

/* submit a request without blocking. opens a stream, encodes the REQUEST,
 * returns immediately with *out_id. the body is copied (caller may free it on
 * return). drive completion with nwep_client_tick + nwep_client_request_poll
 * (or the done-callback). returns 0, or a negative NWEP_ERR_*  -  notably
 * NWEP_ERR_PROTO_MAX_STREAMS at the connection's concurrent-stream limit
 * (apply backpressure and retry later). */
int nwep_client_request_submit(nwep_client *client,
                               int method, const char *path,
                               const nwep_header *headers,
                               const uint8_t *body, size_t body_len,
                               nwep_request_id *out_id);

/* non-blocking completion check. returns 0 (done; *out_response owns a message,
 * free with nwep_message_free, id retired), NWEP_ERR_WOULD_BLOCK (not ready  - 
 * keep ticking), or a negative NWEP_ERR_* (this request failed; id retired).
 * a :stream response is collected and folded into a unary-looking, verifiable
 * message exactly as nwep_client_send does. unknown id -> NWEP_ERR_INTERNAL. */
int nwep_client_request_poll(nwep_client *client, nwep_request_id id,
                             nwep_message **out_response);

/* abandon a submitted request: stop tracking it, free any buffered response.
 * safe whether pending or complete-but-unpolled; a no-op for an unknown id. */
void nwep_client_request_cancel(nwep_client *client, nwep_request_id id);

/* optional push completion, fired from inside nwep_client_tick the instant a
 * request finishes. status is 0 (resp owns a message  -  the callback must take
 * ownership and nwep_message_free it) or a negative NWEP_ERR_* (resp is NULL).
 * the id is retired before the callback runs, so request_poll won't also
 * deliver it. one callback per client; pass cb=NULL to clear. do NOT call
 * nwep_client_tick re-entrantly from the callback. */
typedef void (*nwep_request_done_fn)(nwep_client *client, nwep_request_id id,
                                     int status, nwep_message *resp, void *ud);
int nwep_client_set_request_done(nwep_client *client,
                                 nwep_request_done_fn cb, void *ud);

/* streamed responses (STREAM mode) NW060200.
 *
 * for receiving a response whose body is larger than a single message
 * (NWEP_MAX_MESSAGE_SIZE, 24 MiB) or of unknown length, sent as raw STREAM
 * frames ended by QUIC FIN. flow:
 *
 *     uint64_t sid;
 *     nwep_client_open_stream(cli, NWEP_METHOD_READ, "/big", NULL, &sid);
 *     nwep_message *meta;
 *     nwep_client_stream_response(cli, sid, &meta);   // status + headers
 *     // ... read meta (content-type, etc.), then nwep_message_free(meta);
 *     uint8_t buf[65536];
 *     for (;;) {
 *         size_t n; int ended;
 *         if (nwep_client_stream_recv(cli, sid, buf, sizeof buf, &n, &ended)) break;
 *         // ... consume buf[0..n] ...
 *         if (ended) break;
 *     }
 *     nwep_client_stream_close(cli, sid);
 *
 * these calls are blocking (they drive tick internally), like nwep_client_send.
 */

/* opens a client-bidi stream and sends a body-less request on it; the stream id
 * is written to *out_stream_id. returns 0 or negative NWEP_ERR_*. */
int nwep_client_open_stream(nwep_client *client, int method, const char *path,
                            const nwep_header *headers, uint64_t *out_stream_id);

/* reads the leading RESPONSE frame (status + headers) of the streamed response.
 * blocks until it arrives. caller owns *out_response (nwep_message_free). */
int nwep_client_stream_response(nwep_client *client, uint64_t stream_id,
                               nwep_message **out_response);

/* reads the next body chunk into out_buf (capacity cap). blocks until >= 1 byte
 * is available or the stream ends. writes the byte count to *out_len and 1/0 to
 * *out_ended (FIN seen and all bytes delivered). stop once *out_ended is 1. */
int nwep_client_stream_recv(nwep_client *client, uint64_t stream_id,
                            uint8_t *out_buf, size_t cap,
                            size_t *out_len, int *out_ended);

/* verifies a fully-received streamed response's signature NW060900 against
 * the origin's Ed25519 `pubkey`. call after nwep_client_stream_recv reports
 * ended. returns 0 valid / NWEP_ERR_CRYPTO_VERIFY (missing/invalid signature or
 * truncated stream). the path is the one passed to nwep_client_open_stream. */
int nwep_client_stream_verify(nwep_client *client, uint64_t stream_id,
                              const uint8_t *pubkey);

/* releases the stream's bookkeeping; call once stream_recv reports ended. */
void nwep_client_stream_close(nwep_client *client, uint64_t stream_id);

/* closes the client and frees its handle. */
void nwep_client_close(nwep_client *client);

/* response cache NW060700.
 *
 * an opt-in, client-private cache of read/head responses. when attached, a
 * cacheable response (status ok, carrying `cache-control: max-age=N`, no
 * `no-store`) is stored; a later identical read/head is served from the cache
 * with NO network round-trip while fresh, and revalidated with `if-none-match`
 * once stale (a `not-modified` reuses the cached body). private to one client  -
 * WEB/1 has no shared cache (responses are authenticated by the connection, not
 * signed). default: no cache (the simple client stays simple).
 *
 *     nwep_cache *c = nwep_cache_create(8*1024*1024, 256);
 *     nwep_client_set_cache(cli, c);
 *     // ... nwep_client_send(cli, NWEP_METHOD_READ, ...) now caches ...
 *     nwep_client_close(cli);   // does NOT free the cache
 *     nwep_cache_free(c);
 */
typedef struct nwep_cache nwep_cache;

/* creates a cache bounded by total stored bytes and entry count. NULL on OOM. */
nwep_cache *nwep_cache_create(size_t max_bytes, size_t max_entries);

/* frees a cache. detach it from any client first (or close the client). */
void nwep_cache_free(nwep_cache *cache);

/* drops all stored entries; the cache stays usable. */
void nwep_cache_clear(nwep_cache *cache);

/* copies counters into the out-params (any may be NULL). for tests/demos. */
void nwep_cache_stats(const nwep_cache *cache, uint64_t *out_hits,
                      uint64_t *out_misses, uint64_t *out_stores,
                      uint64_t *out_evictions);

/* attaches `cache` to `client` (borrowed  -  must outlive the client; not freed
 * by nwep_client_close). pass NULL to detach. returns 0 or negative. */
int nwep_client_set_cache(nwep_client *client, nwep_cache *cache);

/* shared cache NW060700 NW060900.
 *
 * a shared cache stores `public`, signed responses received from origin servers
 * and serves them to other clients, who trust them via the response signature
 * (verified against the origin NodeID) rather than the connection. build the
 * cache with nwep_cache_create as usual; these two calls are the proxy surface.
 *
 *     // proxy received `resp` from origin <pubkey> for `path`:
 *     nwep_cache_put_signed(c, "read", path, resp, pubkey, now);
 *     // ... later, another client asks the proxy for the same resource:
 *     nwep_message *m;
 *     if (nwep_cache_get_signed(c, "read", path, pubkey, now, &m) == 0) { ... }
 */

/* verifies a `public`, signed RESPONSE against `origin_pubkey` for `path` and
 * stores it NW060900. returns 0 stored / NWEP_ERR_PROTO_INVALID_HEADER (not
 * `public`, missing signature, missing :status, or missing cache-control) /
 * NWEP_ERR_CRYPTO_VERIFY (bad signature). */
int nwep_cache_put_signed(nwep_cache *cache, const char *method,
                          const char *path, const nwep_message *resp,
                          const uint8_t *origin_pubkey, uint64_t now_secs);

/* looks up a shared entry, re-verifies its signature + freshness against
 * `origin_pubkey` at `now_secs`, and on success returns it as a Message (caller
 * frees via nwep_message_free). returns 0 hit / NWEP_ERR_APP_NOT_FOUND miss /
 * NWEP_ERR_CRYPTO_VERIFY (no longer verifies or stale). */
int nwep_cache_get_signed(nwep_cache *cache, const char *method,
                          const char *path, const uint8_t *origin_pubkey,
                          uint64_t now_secs, nwep_message **out);

/* returns the algorithm negotiated during the handshake:
 *   0  -  none (no compression on this connection)
 *   1  -  zstd
 *   -1  -  handle is NULL or not yet connected */
#define NWEP_COMPRESSION_NONE 0
#define NWEP_COMPRESSION_ZSTD 1
int nwep_client_compression(const nwep_client *client);

/* copies the connected server's Ed25519 public key (learned during the
 * handshake) into `out_pubkey` (32 bytes)  -  use it to verify signed responses
 * (nwep_response_verify / nwep_client_stream_verify) without a separate key
 * exchange. returns 0, or NWEP_ERR_IDENTITY_NOT_FOUND if not yet connected. */
int nwep_client_peer_pubkey(const nwep_client *client, uint8_t *out_pubkey);

/* verifies a response against the connected peer's own public key  -  the same
 * check as nwep_response_verify, but it pulls the pubkey from the connection so
 * the caller can't pass the wrong key. `path` is still required (the response
 * does not carry it; the signed form binds it, NW060900. `now_secs` follows
 * nwep_response_verify's freshness rule (pass 0 to skip). returns 0 if valid,
 * NWEP_ERR_IDENTITY_NOT_FOUND if not yet connected, -801 on NULL client, or a
 * negative NWEP_ERR_* from verification. */
int nwep_client_verify_response(const nwep_client *client,
                                const nwep_message *resp,
                                const char *path, uint64_t now_secs);

/* pumps the connection once and returns the next queued NOTIFY push NW060200,
 * or NULL if none is pending. the returned handle is owned by the
 * caller  -  free it with nwep_message_free. read the event with
 * nwep_message_get_header(msg, ":event"). call repeatedly to drain all
 * pending NOTIFYs. */
nwep_message *nwep_client_poll_notify(nwep_client *client);

/* message accessors.
 *
 * all pointers returned here are borrowed from the message handle and are
 * valid only until nwep_message_free is called. do not free them.
 *
 * header lookups are case-sensitive; the protocol's pseudo-headers
 * (":status", ":path", ":method", ":event") are lowercase. */

/* returns the value of the named header (NUL-terminated borrowed), or
 * NULL if the header is absent. */
const char *nwep_message_get_header(const nwep_message *msg, const char *name);

/* number of headers on the message, including pseudo-headers (":status", ...).
 * pair with nwep_message_header_at to walk every header by index  -  for a
 * consumer that must print or forward headers whose names it does not know in
 * advance (curl -i, a reverse proxy, a generic binding). */
size_t nwep_message_header_count(const nwep_message *msg);

/* borrows the i-th header's name and value (NUL-terminated, valid until
 * nwep_message_free, like nwep_message_get_header). headers keep wire order.
 * returns 0 on success, NWEP_ERR_INTERNAL if i is out of range, or
 * NWEP_ERR_INTERNAL_ALLOC on a copy failure. typical dump:
 *     for (size_t i = 0, n = nwep_message_header_count(msg); i < n; i++) {
 *         const char *k, *v;
 *         if (nwep_message_header_at(msg, i, &k, &v) == 0)
 *             fprintf(stderr, "%s: %s\n", k, v);
 *     }
 */
int nwep_message_header_at(const nwep_message *msg, size_t i,
                           const char **name, const char **value);

/* returns the WEB/1 status string (e.g. "ok", "not-found", "error"). for
 * non-response messages returns NULL. */
const char *nwep_message_get_status(const nwep_message *msg);

/* returns the response body and its length. sets `*out_len` to the byte
 * count. empty body returns NULL and *out_len = 0. NULL `out_len` returns
 * NULL without writing anything. */
const uint8_t *nwep_message_get_body(const nwep_message *msg, size_t *out_len);

/* conditional-request helper NW060700: returns 1 if the REQUEST's
 * `if-none-match` covers `etag`  -  i.e. the client's cached copy is current and
 * the handler should answer not-modified. matches `*` or any ETag in the
 * comma-separated list (compared as opaque strings). returns 0 when the header
 * is absent (unconditional request); -801 if `req` or `etag` is NULL. typical
 * handler:
 *     char etag[64]; ... compute etag for the resource ...
 *     if (nwep_request_is_fresh(req, etag))
 *         return nwep_response_not_modified(resp, etag);
 *     // else send the body with nwep_response_ok + content/etag headers
 */
int nwep_request_is_fresh(const nwep_message *req, const char *etag);

/* byte ranges NW060800 */

/* one inclusive byte range [start, end]. */
typedef struct { uint64_t start; uint64_t end; } nwep_range;

/* nwep_request_range outcome codes. */
#define NWEP_RANGE_OK            0 /* out holds *out_count satisfiable ranges */
#define NWEP_RANGE_NONE          1 /* no/ignored range (or if-range miss) -> full body */
#define NWEP_RANGE_UNSATISFIABLE 2 /* valid but out of bounds -> range-not-satisfiable */

/* parses a REQUEST's `range` header against a resource of `total_len` bytes
 * NW060800, resolving suffix (-N) and open (N-) forms and clamping to bounds.
 * if `etag` is non-NULL and the request carries `if-range` that does not match
 * it, the range is ignored (NWEP_RANGE_NONE) so a resumed transfer never mixes
 * versions. satisfiable ranges are written to `out` (capacity `max_out`) and
 * their count to *out_count. returns NWEP_RANGE_OK / _NONE / _UNSATISFIABLE, or
 * a negative NWEP_ERR_* on a NULL arg. typical handler:
 *     nwep_range rs[16]; size_t nr;
 *     switch (nwep_request_range(req, file_size, etag, rs, 16, &nr)) {
 *       case NWEP_RANGE_OK:            return nwep_response_partial(resp, body, file_size, rs, nr, ctype);
 *       case NWEP_RANGE_UNSATISFIABLE: return nwep_response_range_not_satisfiable(resp, file_size);
 *       default:                       return nwep_response_ok(resp, body, file_size); // full
 *     }
 */
int nwep_request_range(const nwep_message *req, uint64_t total_len,
                       const char *etag, nwep_range *out, size_t max_out,
                       size_t *out_count);

/* verifies a RESPONSE's signature NW060900 against the origin's Ed25519
 * public key. the C server signs every response automatically; a client (or a
 * shared cache) verifies with this. `path` is the request path the response
 * answers (part of the signed form, not carried in the response). `now_secs` is
 * unix seconds: if non-zero and the response carries cache-control max-age, the
 * signature is rejected once signature-ts + max-age has passed (shared-cache
 * freshness); pass 0 to skip the freshness check (point-to-point use). returns 0
 * if valid (and fresh), or negative NWEP_ERR_* (PROTO_INVALID_HEADER for a
 * missing/garbled signature, CRYPTO_VERIFY for a bad signature or stale). */
int nwep_response_verify(const nwep_message *resp, const uint8_t *pubkey,
                         const char *path, uint64_t now_secs);

/* frees the message handle and every pointer borrowed from it. */
void nwep_message_free(nwep_message *msg);

/* server */

/* opaque server handle. */
typedef struct nwep_server nwep_server;

/* mutable response buffer passed to a handler. the handler appends the
 * complete encoded response  -  typically via nwep_response_ok / _bytes /
 * _error helpers (declared below). */
typedef struct {
    void *opaque; /* internal pointer to a Buf; do not inspect */
} nwep_buf;

/* handler signature. called for each fully-decoded incoming request.
 *   server    -  the server handle (for nwep_server_get_peer_nodeid).
 *   conn_id   -  opaque identifier of the connection the request arrived on.
 *   stream_id  -  QUIC stream id the request rode in on.
 *   request   -  read-only decoded message; borrowed for the call duration.
 *   resp_buf  -  output buffer; the handler must append the encoded response
 *              before returning. use nwep_response_* helpers.
 *   userdata  -  value passed to nwep_server_set_handler.
 *
 * returning 0 is success. negative NWEP_ERR_* turns into a generic error
 * response (the handler should not also write to resp_buf in that case).
 *
 * returning NWEP_DEFER answers the request OUT OF BAND later (NW000017):
 * the handler does NOT write resp_buf; the server keeps (conn_id, stream_id)
 * open, and the application delivers the response from its event loop via
 * nwep_server_respond / nwep_server_relay. use it when the response depends on
 * a backend fetch you don't want to block the loop on (a proxy, a gateway). */
typedef int (*nwep_handler_fn)(
    nwep_server *server,
    uint64_t     conn_id,
    uint64_t     stream_id,
    const nwep_message *request,
    nwep_buf    *resp_buf,
    void        *userdata
);

/* handler return sentinel: answer this request later, out of band. distinct
 * from 0 (answered synchronously) and negative (error). */
#define NWEP_DEFER 1

/* builds an `:status = ok` response with the given body (may be NULL when
 * len is 0). returns 0 on success, negative on allocation failure. */
int nwep_response_ok(nwep_buf *resp, const uint8_t *body, size_t body_len);

/* builds a response with the given WEB/1 status token and body. `status` is
 * one of "ok", "created", "no-content", "bad-request", "not-found",
 * "conflict", "rate-limited", "error", ... NW080000. */
int nwep_response_status(nwep_buf *resp, const char *status,
                         const uint8_t *body, size_t body_len);

/* builds a `:status = not-modified` response with the given `etag` and an empty
 * body  -  the answer to a conditional read/head whose if-none-match matched
 * NW060700. pairs with nwep_request_is_fresh. */
int nwep_response_not_modified(nwep_buf *resp, const char *etag);

/* builds a `:status = partial-content` response for one or more byte ranges of
 * `body` (the full resource, `body_len` bytes), NW060800. one range sends the
 * sub-range with `content-range: bytes start-end/total` + `content_type`;
 * multiple ranges send a `multipart/byteranges` body. `ranges`/`count` come from
 * nwep_request_range. returns 0 or negative NWEP_ERR_*; count == 0 returns
 * NWEP_ERR_PROTO_INVALID_METHOD (-402). */
int nwep_response_partial(nwep_buf *resp, const uint8_t *body, size_t body_len,
                          const nwep_range *ranges, size_t count,
                          const char *content_type);

/* builds a `:status = range-not-satisfiable` response with
 * `content-range: bytes * /total` NW060800  -  when a well-formed `range`
 * selected no bytes. */
int nwep_response_range_not_satisfiable(nwep_buf *resp, uint64_t total_len);

/* attaches a custom response header to the next nwep_response_* call on
 * this buffer. the library copies both name and value, so the caller's
 * storage does not need to outlive the call. may be invoked any number
 * of times; the appended headers are flushed with the response. */
int nwep_response_header(nwep_buf *resp, const char *name, const char *value);

/* verbatim relay (NW000017): emits `origin`'s :status + headers (incl.
 * its signature / signature-ts) + body into this handler's response buffer
 * without re-signing  -  so a cache/proxy serving a stored origin response on a
 * synchronous cache hit preserves the origin's end-to-end signature (the
 * client verifies against the origin NodeID, not the proxy). for a deferred
 * (post-backend) relay, use nwep_server_relay instead. returns 0,
 * NWEP_ERR_PROTO_INVALID_HEADER (origin has no :status), or an encode error. */
int nwep_response_relay(nwep_buf *resp, const nwep_message *origin);

/* pre-signed cache blit NW000017.
 *
 * serve a hot resource with zero per-hit work: build (or relay) the response
 * once, capture its encoded wire frame, and on subsequent hits blit those bytes
 * back verbatim  -  no decode, re-encode, re-compress, or re-sign. the frame cache
 * is yours (as the response cache is); nwep just provides capture + blit +
 * a codec query so you key the cache correctly.
 *
 * a frame is codec-specific and time-bounded:
 *   - it encodes for one negotiated codec, so key your cache by
 *     (path x nwep_server_conn_compression(...)) and only blit a frame onto a
 *     connection reporting the same codec;
 *   - its validity is the response's signature-ts + max-age (the same cache-
 *     control TTL you already honor)  -  rebuild before that lapses or clients
 *     with a non-zero clock reject it as stale.
 * the signature covers path + body + headers + ts (not the connection or
 * stream), so one captured frame blits safely onto any stream/connection for the
 * same path.
 *
 *     // first hit  -  build + stash:
 *     nwep_response_ok(resp, body, len);
 *     size_t n; nwep_response_capture(resp, NULL, 0, &n);   // probe
 *     uint8_t *f = malloc(n); nwep_response_capture(resp, f, n, &n);
 *     cache_put(path, codec, f, n);
 *     // later hits  -  blit:
 *     nwep_response_blit(resp, f, n);
 */

/* copies the encoded frame just built into `resp` out to the caller for caching.
 * two-call probe: pass out=NULL to read the needed length into *out_len, then
 * call again with a buffer of at least that size. returns 0, or NWEP_ERR_INTERNAL
 * on a NULL handle or when out is non-NULL but cap is too small (*out_len is
 * still set, so resize and retry). */
int nwep_response_capture(nwep_buf *resp, uint8_t *out, size_t cap,
                          size_t *out_len);

/* writes a captured frame verbatim into `resp` as the response  -  no re-encode or
 * re-sign. the frame must be for this connection's codec (see
 * nwep_server_conn_compression) and still within its signature-ts + max-age.
 * synchronous in-handler counterpart of nwep_server_respond_blit. returns 0,
 * NWEP_ERR_INTERNAL on a NULL handle/frame, or NWEP_ERR_INTERNAL_ALLOC. */
int nwep_response_blit(nwep_buf *resp, const uint8_t *frame, size_t len);

/* deferred responses NW000017.
 *
 * after a handler returns NWEP_DEFER, deliver the response out of band from your
 * event loop with these. the stream stays open until you respond, the parked
 * deadline elapses (then the client gets a generic error), or the peer
 * disconnects. each call below returns NWEP_ERR_APP_NOT_FOUND if (conn_id,
 * stream_id) is no longer parked  -  treat that as "client gone, discard the
 * work," not an error to retry. delivery is exactly-once: the first respond /
 * relay wins and clears the parked entry.
 *
 *     // in the handler:
 *     submit_backend_fetch(...); remember (conn_id, stream_id);
 *     return NWEP_DEFER;
 *     // later, when the backend completes in the same loop:
 *     nwep_server_relay(srv, conn_id, stream_id, origin_resp);  // verbatim, or
 *     nwep_server_respond(srv, conn_id, stream_id, "ok", body, len); // re-signed
 */

/* attaches a header to the next nwep_server_respond on a parked (conn,stream)  -
 * the deferred analogue of nwep_response_header. the library copies name+value.
 * ignored by nwep_server_relay (which emits the origin message's own headers).
 * returns 0, NWEP_ERR_APP_NOT_FOUND (not parked), NWEP_ERR_INTERNAL_ALLOC, or
 * -801 if server, name, or value is NULL. */
int nwep_server_respond_header(nwep_server *server, uint64_t conn_id,
                               uint64_t stream_id, const char *name,
                               const char *value);

/* delivers a deferred response, signed with the server identity over the
 * request path NW060900  -  exactly like a synchronous nwep_response_*. returns
 * 0, NWEP_ERR_APP_NOT_FOUND (not parked), or a negative encode error. */
int nwep_server_respond(nwep_server *server, uint64_t conn_id, uint64_t stream_id,
                        const char *status,
                        const uint8_t *body, size_t body_len);

/* delivers an existing signed message verbatim onto a parked stream  -  no
 * re-sign. emits origin_resp's :status + headers (incl. its signature /
 * signature-ts) + body unchanged, so a cache/proxy preserves the origin's
 * end-to-end signature instead of re-signing with its own identity. the client
 * then verifies against the origin NodeID through the intermediary. returns 0,
 * NWEP_ERR_APP_NOT_FOUND, NWEP_ERR_PROTO_INVALID_HEADER (origin has no :status),
 * or a negative encode error. */
int nwep_server_relay(nwep_server *server, uint64_t conn_id, uint64_t stream_id,
                      const nwep_message *origin_resp);

/* delivers a captured frame (nwep_response_capture) verbatim onto a parked
 * stream  -  no re-encode or re-sign (NW000017 pre-signed blit). deferred
 * counterpart of nwep_response_blit: the cache-hit fast path for an out-of-band
 * responder. the frame must be for this connection's codec
 * (nwep_server_conn_compression) and within its signature-ts + max-age. clears
 * the parked entry (exactly-once). returns 0, NWEP_ERR_INTERNAL on a NULL
 * handle/frame, NWEP_ERR_APP_NOT_FOUND if not parked, or a delivery error. */
int nwep_server_respond_blit(nwep_server *server, uint64_t conn_id,
                             uint64_t stream_id, const uint8_t *frame,
                             size_t len);

/* the codec a connection negotiated: 0 = none, 1 = zstd, -1 = NULL handle or
 * unknown connection (NW000017). a captured pre-signed frame is codec-
 * specific, so key your frame cache by (path x this value) and only blit a frame
 * onto a connection reporting the same codec; on -1, build fresh rather than
 * risk a mismatch. */
int nwep_server_conn_compression(const nwep_server *server, uint64_t conn_id);

/* binds to `bind_addr` and prepares an empty server. the handler must be
 * registered via nwep_server_set_handler before nwep_server_tick is called
 * for the first time. */
int nwep_server_listen(nwep_server **out,
                       const nwep_keypair *identity,
                       const nwep_address *bind_addr);

/* like nwep_server_listen, but binds with SO_REUSEPORT so N reactors can share
 * one port and the kernel fans connections across them (NW000017). the
 * single-binary convenience equivalent of each reactor creating its own
 * SO_REUSEPORT socket and calling nwep_server_listen_fd. Linux/Android only: on
 * unsupported platforms it returns NWEP_ERR_CONFIG_INVALID (-101)  -  query
 * nwep_reuse_port_supported() first to branch. */
int nwep_server_listen_reuseport(nwep_server **out,
                                 const nwep_keypair *identity,
                                 const nwep_address *bind_addr);

/* like nwep_server_listen, but adopts a caller-created, already-bound UDP
 * socket instead of binding one (NW000017 caller-owned I/O). ownership
 * of `fd` transfers to the server: it is closed by nwep_server_close and on a
 * failure here. the socket must be AF_INET6, UDP, and bound. this is the
 * portable primitive for multi-reactor scale-out: each reactor process creates
 * its own socket (on Linux, with SO_REUSEPORT so the kernel fans connections
 * across them) and hands the fd here. `fd` is uintptr_t to hold both a POSIX
 * int fd and a Windows SOCKET. */
int nwep_server_listen_fd(nwep_server **out,
                          const nwep_keypair *identity,
                          uintptr_t fd);

/* like nwep_server_listen_fd, but tags this reactor with `shard_id` NW000017:
 * every connection id the server issues carries the shard in a fixed
 * prefix, so a SO_REUSEPORT steering program (BPF/XDP) routes a packet to the
 * reactor that owns the connection  -  surviving peer 4-tuple migration. the
 * canonical multi-reactor entry point: each reactor adopts its own SO_REUSEPORT
 * socket and a distinct shard_id. see examples/reuseport_steering.bpf.c. */
int nwep_server_listen_fd_sharded(nwep_server **out,
                                  const nwep_keypair *identity,
                                  uintptr_t fd,
                                  uint16_t shard_id);

/* extracts the shard id a server stamped into a connection id, or -1 if `cid`
 * is not shard-encoded. defines the on-wire scheme a steering program reads;
 * also lets a userspace fallback / test recover the shard. the scheme: byte 0
 * is NWEP_CID_SHARD_MARKER, bytes 1..3 are the shard id (big-endian u16), the
 * rest is random. server SCIDs are NWEP_SERVER_CID_LEN bytes. */
int nwep_cid_shard_id(const uint8_t *cid, size_t cid_len);

/* on-wire CID shard scheme (NW000017)  -  shared with steering programs. */
#define NWEP_CID_SHARD_MARKER     0x5e
#define NWEP_CID_SHARD_PREFIX_LEN 3
#define NWEP_SERVER_CID_LEN       18

/* whether this build supports SO_REUSEPORT kernel load-balancing (Linux and
 * Android only; not Windows or macOS). returns 1 if supported, 0 otherwise.
 * query this before building a multi-reactor pool that shares one port  -  on
 * unsupported platforms the kernel-fanned model is unavailable and the
 * supported scale model is a single acceptor (see docs/LIMITATIONS.md). */
int nwep_reuse_port_supported(void);

/* backpressure & admission control NW000017 */

/* sets the caller-fed overload signal. nonzero `on` makes the server shed every
 * new request with `rate-limited` (+ retry-after) before the handler runs, and
 * nwep_server_load report 100. nwep has no global allocator, so it cannot see
 * the process's memory  -  raise this from your own OS-level pressure (RSS,
 * cgroup limit, queue depth) and clear it when pressure subsides. the library
 * still caps what it manages (connections, deferred/parked responses)
 * independently. */
void nwep_server_set_overloaded(nwep_server *server, int on);

/* tunes the deferred-response (parked) cap at runtime. past this many concurrent
 * deferred (NWEP_DEFER) responses the server sheds new requests with
 * `rate-limited`. default is a built-in cap. */
void nwep_server_set_max_parked(nwep_server *server, size_t max_parked);

/* a 0..100 per-reactor load factor for an L4 router or health check to steer
 * traffic away from a hot reactor (pairs with the phase 3 shard steering): the
 * max of the connection-cap and parked-cap utilisations, or 100 when overloaded.
 * returns -1 on a NULL handle. */
int nwep_server_load(const nwep_server *server);

/* observability NW000017 */

/* snapshot of one reactor's counters. cumulative counters plus three gauges
 * (connections_active, parked_active, load). pull model: read whenever you
 * scrape metrics; nwep keeps no exporter. there is no RTT or packet-loss rollup
 * and no memory-in-use figure  -  a single server-level RTT would hide the
 * per-connection distribution (a p50/p99 histogram is the real answer, deferred
 * not precluded) and nwep has no global allocator to watermark (see
 * docs/LIMITATIONS.md 13). */
typedef struct {
  uint64_t connections_active;   /* gauge: live connections right now */
  uint64_t connections_accepted; /* cumulative: handshakes admitted */
  uint64_t connections_refused;  /* cumulative: dropped at the connection cap */
  uint64_t connections_closed;   /* cumulative: connections torn down */
  uint64_t bytes_received;       /* cumulative: UDP payload bytes in */
  uint64_t bytes_sent;           /* cumulative: UDP payload bytes out */
  uint64_t datagrams_received;   /* cumulative: UDP datagrams in */
  uint64_t datagrams_sent;       /* cumulative: UDP datagrams out */
  uint64_t requests_dispatched;  /* cumulative: requests that reached a handler */
  uint64_t requests_shed;        /* cumulative: requests shed at the front door */
  uint64_t parked_active;        /* gauge: deferred responses outstanding */
  int32_t load;                  /* gauge: 0..100 (== nwep_server_load) */
} nwep_server_metrics;

/* fills *out with the server's current metrics snapshot. returns 0, or
 * NWEP_ERR_INTERNAL on a NULL server or out. call from the reactor's owning
 * thread  -  the one that runs nwep_server_tick. the counters are non-atomic and
 * the gauges walk live maps at read time, so a cross-thread scrape races the
 * tick (in a SO_REUSEPORT pool each reactor owns its own counters). fold the
 * scrape into your event loop. */
int nwep_server_metrics_get(const nwep_server *server, nwep_server_metrics *out);

/* graceful drain NW000017 */

/* begins a graceful drain: the reactor stops accepting new connections (their
 * initials are dropped so a fresh reactor in the SO_REUSEPORT group serves the
 * client's retransmit) while existing connections keep running and finishing
 * in-flight requests. idempotent. the typical zero-downtime sequence:
 *
 *     // bring up the replacement first so the port group keeps full capacity:
 *     spawn_new_reactor();                  // nwep_server_listen_fd(..., reuse)
 *     nwep_server_drain(old);               // old stops accepting
 *     while (!nwep_server_is_drained(old) && now() < deadline)
 *         nwep_server_tick(old, now());     // finish in-flight work
 *     nwep_server_close(old);               // graceful CONNECTION_CLOSE to any
 *                                           // stragglers, then teardown
 *
 * returns 0, or NWEP_ERR_INTERNAL on a NULL handle. see docs/LIMITATIONS.md for
 * the SO_REUSEPORT hand-off caveat (a new client whose initial hashes to the
 * draining socket retries until that socket finally closes). */
int nwep_server_drain(nwep_server *server);

/* whether a drain has completed: 1 once nwep_server_drain was called and no
 * connections remain (idle  -  safe to nwep_server_close and exit with no dropped
 * in-flight work), 0 while still draining or not draining at all, -1 on a NULL
 * handle. poll this each tick; pair with a wall-clock deadline so a stuck peer
 * can't pin the drain open forever (close() still CONNECTION_CLOSEs stragglers).
 * the connections_active metric (nwep_server_metrics_get) shows drain progress. */
int nwep_server_is_drained(const nwep_server *server);

/* registers the dispatch handler. replaces any previously registered
 * handler. userdata is opaque and is passed through to every call.
 * returns 0, or -801 if server is NULL. */
int nwep_server_set_handler(nwep_server *server,
                            nwep_handler_fn handler,
                            void *userdata);

/* drives all state machines: reads inbound datagrams, runs the QUIC
 * handshake, fires the application handler for completed requests, and
 * flushes outbound datagrams. call from your event loop on every I/O
 * event or timer expiry. `now_ms` is a monotonic millisecond clock. */
int nwep_server_tick(nwep_server *server, int64_t now_ms);

/* returns the port the server actually bound to (resolves the "let the
 * kernel pick" case where bind_addr was `*::*:0`). */
uint16_t nwep_server_local_port(const nwep_server *server);

/* returns the UDP socket descriptor backing the server  -  an int fd on POSIX,
 * a SOCKET (cast to this signed type) on Windows  -  or -1 if server is NULL.
 * register it with your poller (epoll/kqueue/poll/io_uring) for read
 * readiness. the socket is non-blocking and owned by the server; do not close
 * it. drive all I/O through nwep_server_tick. */
intptr_t nwep_server_fd(const nwep_server *server);

/* returns the number of milliseconds until nwep_server_tick must next run to
 * service a pending QUIC timer (retransmit / PTO / handshake / idle), or -1
 * when nothing is pending. the value is designed to be passed straight as the
 * timeout to poll()/epoll_wait(); after the poll returns (whether from a
 * datagram or the timeout), call nwep_server_tick. an already-expired timer
 * returns 0. `now_ms` is the same monotonic clock passed to nwep_server_tick.
 *
 * the canonical event loop:
 *     int fd = nwep_server_fd(srv);  // add fd to your poller for EPOLLIN
 *     for (;;) {
 *         int t = nwep_server_next_timeout_ms(srv, now_ms());
 *         poll(&pfd, 1, t);          // wakes on a datagram OR the timer
 *         nwep_server_tick(srv, now_ms());
 *     }
 */
int nwep_server_next_timeout_ms(nwep_server *server, int64_t now_ms);

/* copies the peer's authenticated NodeID into `out_node_id`. returns 0 on
 * success or NWEP_ERR_IDENTITY_NOT_FOUND if conn_id is unknown / hasn't
 * completed the WEB/1 handshake yet. */
int nwep_server_get_peer_nodeid(const nwep_server *server,
                                uint64_t conn_id,
                                nwep_node_id *out_node_id);

/* copies the server's own NodeID (the one clients dial) into `out`. returns 0,
 * or NWEP_ERR_INTERNAL if `server` is NULL. */
int nwep_server_local_nodeid(const nwep_server *server, nwep_node_id *out);

/* returns the negative NWEP_ERR_* reason the most recent connection was
 * rejected for a fatal handshake failure, or 0 if none has occurred. a
 * rejected handshake is closed silently to the peer NW150200  -  no error is
 * sent on the wire; this is the server operator's only window into why
 * inbound dials are failing (e.g. NWEP_ERR_CRYPTO_FATAL_NODEID_MISMATCH).
 * for local diagnostics; the value is never transmitted. returns 0 if none
 * has occurred, the negative NWEP_ERR_* reason otherwise, or -801 on NULL. */
int nwep_server_last_handshake_error(const nwep_server *server);

/* sends a NOTIFY push NW060200 to connection `conn_id` on a fresh
 * server-initiated unidirectional stream. `event` is the :event value;
 * `headers` is an optional NULL-terminated nwep_header array of extra headers
 * (NULL for none); `body` may be NULL. the push is flushed on the next
 * nwep_server_tick. returns 0 or a negative NWEP_ERR_* (e.g.
 * NWEP_ERR_IDENTITY_NOT_FOUND for an unknown conn_id). */
int nwep_server_notify(nwep_server *server, uint64_t conn_id,
                       const char *event, const nwep_header *headers,
                       const uint8_t *body, size_t body_len);

/* streamed responses (STREAM mode) NW060200.
 *
 * serve a response body larger than NWEP_MAX_MESSAGE_SIZE (24 MiB) or of unknown
 * length, as raw STREAM frames ended by QUIC FIN, without buffering it all in
 * one message. typical use from the dispatch handler:
 *
 *     // in the handler, having decided to stream:
 *     nwep_header h[] = {{"content-type","video/mp4"},{NULL,NULL}};
 *     nwep_server_begin_stream(server, conn, stream, path, "ok", h);
 *     // stash (conn, stream); return 0 without writing resp.
 *     // then, from your event loop across ticks:
 *     int took = nwep_server_stream_send(server, conn, stream, chunk, len);
 *     //   took may be < len (back-pressure)  -  retry the tail after a tick.
 *     nwep_server_stream_end(server, conn, stream);   // when done
 *
 * the streamed body bypasses the single-message size cap and is never
 * compressed. the response is signed NW060900: the signature is emitted in a
 * trailer frame on nwep_server_stream_end, and the client verifies it with
 * nwep_client_stream_verify. the stream is retired once the peer acks it. */

/* emits the metadata RESPONSE frame (status + headers, no body) and puts the
 * stream into STREAM mode. `path` is the request's path (bound into the response
 * signature  -  pass the `:path` being answered). `status` is a NW080000 token;
 * `headers` is an optional NULL-terminated nwep_header array. returns 0 or
 * negative NWEP_ERR_*. */
int nwep_server_begin_stream(nwep_server *server, uint64_t conn_id,
                             uint64_t stream_id, const char *path,
                             const char *status, const nwep_header *headers);

/* queues body bytes on an out-streaming stream. returns the number of bytes
 * accepted (>= 0), which may be fewer than body_len (including 0) under
 * back-pressure  -  let nwep_server_tick drain/ack, then retry the unaccepted
 * tail. a negative return is an NWEP_ERR_*. */
int nwep_server_stream_send(nwep_server *server, uint64_t conn_id,
                            uint64_t stream_id, const uint8_t *body,
                            size_t body_len);

/* ends an out-streaming response (flushes remaining frames + writes QUIC FIN).
 * no further stream_send is permitted. returns 0 or negative NWEP_ERR_*. */
int nwep_server_stream_end(nwep_server *server, uint64_t conn_id,
                           uint64_t stream_id);

/* closes the listening socket and tears down all connections. each live
 * connection is first sent a graceful CONNECTION_CLOSE(NO_ERROR) (best-effort)
 * so peers learn the server is gone immediately rather than hanging until their
 * idle timeout (NW000017). also closes an attached DHT. */
void nwep_server_close(nwep_server *server);

/* log server NW120400 NW121000.
 *
 * an embedder running a log server keeps an in-memory append-only log and
 * routes the log endpoints through the real handlers from inside its server
 * dispatch callback:
 *
 *     int handler(nwep_server *s, uint64_t conn, uint64_t stream,
 *                 const nwep_message *req, nwep_buf *resp, void *ud) {
 *         nwep_log_server *ls = ud;
 *         int rc = nwep_log_server_dispatch(ls, conn, req, resp, time(NULL));
 *         if (rc == 1) return nwep_response_status(resp, "not-found", NULL, 0);
 *         return rc;  // 0 handled, or negative error
 *     }
 *
 * persistent storage (entry/index files, fsync ordering) is caller-owned;
 * rebuild the in-memory log at startup with nwep_log_append. */

typedef struct nwep_log nwep_log;
typedef struct nwep_log_server nwep_log_server;

/* creates an empty in-memory log, or NULL on allocation failure. */
nwep_log *nwep_log_create(void);
void nwep_log_free(nwep_log *log);

/* appends a raw entry (hashed as a Merkle leaf as-is  -  no structural
 * validation at this layer). returns the new entry index (>= 0) or a
 * negative NWEP_ERR_* (-801 on NULL log; -101 on NULL bytes). */
int64_t nwep_log_append(nwep_log *log, const uint8_t *bytes, size_t len);

/* number of entries appended. returns 0 for a NULL handle (indistinguishable
 * from an empty log  -  do not pass NULL). */
uint64_t nwep_log_size(const nwep_log *log);

/* writes the log's current 32-byte Merkle root into `out_root`. lets an
 * embedder produce a checkpoint over its own log without reading /log/root back
 * over the wire. returns 0 or a negative NWEP_ERR_*. `out_root` must not be
 * NULL (no null check  -  crashes). */
int nwep_log_root(const nwep_log *log, uint8_t *out_root);

/* creates a log server signing with `identity` over `log`. the server's
 * NodeID is what a no-revocation assertion's `server-id` binds to, so this
 * identity should match the QUIC server's identity. `log` is borrowed and
 * must outlive the returned handle. `identity` is copied  -  its lifetime is
 * not constrained. `identity` must not be NULL (crashes). NULL on failure. */
nwep_log_server *nwep_log_server_create(const nwep_keypair *identity,
                                        nwep_log *log);
void nwep_log_server_free(nwep_log_server *ls);

/* persistence hook: invoked with the raw bytes + log index of each entry the
 * server ACCEPTS via WRITE /log/entry (after the in-memory append, before the
 * `created` response). lets an embedder durably persist accepted entries
 * directly  -  no need to capture the request body and infer acceptance from a
 * log-size delta. `ctx` is the userdata passed to set_on_append; `entry`/`len`
 * are borrowed for the call only (copy if you keep them). */
typedef void (*nwep_log_append_fn)(void *ctx, const uint8_t *entry, size_t len,
                                   uint64_t index);

/* registers (or, with cb=NULL, clears) the accepted-entry persistence hook. */
void nwep_log_server_set_on_append(nwep_log_server *ls, nwep_log_append_fn cb,
                                   void *ctx);

/* routes `req` through the real log-endpoint handlers, writing the response
 * into `resp`. `conn_id` is the request's connection id (from the server
 * dispatch callback)  -  the write endpoints rate-limit per connection and
 * answer `rate-limited` with a `retry-after` header once a connection
 * exceeds its budget. `now_secs` is Unix seconds (used by the root and
 * checkpoint endpoints and the rate-limit window; ignored by the revocation
 * query). returns:
 *    0   -  handled; response written to `resp`.
 *    1   -  not a log-server route; fall through to your own handler.
 *   <0   -  negative NWEP_ERR_*. */
int nwep_log_server_dispatch(nwep_log_server *ls, uint64_t conn_id,
                             const nwep_message *req, nwep_buf *resp,
                             int64_t now_secs);

/* trust-log entry creation NW120300.
 *
 * builds a node's own signed log entries, ready to submit via
 * WRITE /log/entry. the producer side of the identity lifecycle; verifying
 * others' entries lives in libnwep_trust. Ed25519-only (no BLS), so these are
 * in core. all use two-call sizing: out == NULL writes the required size to
 * *outlen and returns 0; a too-small buffer reports the size and returns
 * NWEP_ERR_INTERNAL. timestamps are Unix seconds. */

/* registers `pubkey` (NodeID derived from it). recovery_commitment is
 * SHA-256(recovery_pubkey)  -  recovery key stays offline. signed by privkey.
 * produces the 169-byte KeyBinding entry. */
int nwep_keybinding_create(const uint8_t pubkey[NWEP_PUBKEY_SIZE],
                           const uint8_t recovery_commitment[32],
                           uint64_t timestamp,
                           const uint8_t privkey[NWEP_PRIVKEY_SIZE],
                           uint8_t *out, size_t *outlen);

/* rotates `node_id` from old_pubkey to new_pubkey. overlap_expiry is the
 * unix-seconds cutoff after which the old key is rejected. signed by both
 * keys. produces the 241-byte KeyRotation entry. */
int nwep_keyrotation_create(const uint8_t node_id[NWEP_NODEID_SIZE],
                            const uint8_t old_pubkey[NWEP_PUBKEY_SIZE],
                            const uint8_t new_pubkey[NWEP_PUBKEY_SIZE],
                            uint64_t timestamp, uint64_t overlap_expiry,
                            const uint8_t old_privkey[NWEP_PRIVKEY_SIZE],
                            const uint8_t new_privkey[NWEP_PRIVKEY_SIZE],
                            uint8_t *out, size_t *outlen);

/* revokes revoked_pubkey under node_id. signed by the offline recovery key;
 * recovery_pubkey is carried in the entry. reason: 1=compromised,
 * 2=rotation, 3=decommission. produces the 170-byte Revocation entry. */
int nwep_revocation_create(const uint8_t node_id[NWEP_NODEID_SIZE],
                           const uint8_t revoked_pubkey[NWEP_PUBKEY_SIZE],
                           const uint8_t recovery_pubkey[NWEP_PUBKEY_SIZE],
                           uint8_t reason, uint64_t timestamp,
                           const uint8_t recovery_privkey[NWEP_PRIVKEY_SIZE],
                           uint8_t *out, size_t *outlen);

/* trust-log entry decoding NW120300.
 *
 * the inverse of the create calls above: turn an encoded entry read back from
 * the log (e.g. WEB/1 READ /log/entry/{idx}) into named fields, so a consumer
 * auditing the log never hand-slices the wire layout by offset. these parse
 * only  -  they do NOT verify signatures; use libnwep_trust's verify path for
 * that. the structs are a decoded view (no type byte, compiler-chosen padding),
 * NOT the wire layout  -  populate them only via the decode calls, never by
 * memcpy from the wire bytes. each decode returns 0, NWEP_ERR_INTERNAL on a
 * NULL buffer, or NWEP_ERR_PROTO_INVALID_MESSAGE if the bytes are too short or
 * the wrong type. timestamps are Unix seconds. */

/* entry-type codes returned by nwep_log_entry_type. */
#define NWEP_ENTRY_KEY_BINDING  1
#define NWEP_ENTRY_KEY_ROTATION 2
#define NWEP_ENTRY_REVOCATION   3
#define NWEP_ENTRY_ANCHOR_CHANGE 4

typedef struct {
    uint8_t node_id[NWEP_NODEID_SIZE];
    uint8_t pubkey[NWEP_PUBKEY_SIZE];
    uint8_t recovery_commitment[32];
    uint64_t timestamp;
    uint8_t signature[64];
} nwep_keybinding;

typedef struct {
    uint8_t node_id[NWEP_NODEID_SIZE];
    uint8_t old_pubkey[NWEP_PUBKEY_SIZE];
    uint8_t new_pubkey[NWEP_PUBKEY_SIZE];
    uint64_t timestamp;
    uint64_t overlap_expiry;
    uint8_t sig_old[64];
    uint8_t sig_new[64];
} nwep_keyrotation;

typedef struct {
    uint8_t node_id[NWEP_NODEID_SIZE];
    uint8_t revoked_pubkey[NWEP_PUBKEY_SIZE];
    uint8_t recovery_pubkey[NWEP_PUBKEY_SIZE];
    uint8_t reason; /* 1=compromised, 2=rotation, 3=decommission */
    uint64_t timestamp;
    uint8_t signature[64];
} nwep_revocation;

/* returns the entry's type code (NWEP_ENTRY_*, all positive) so a caller that
 * fetched an arbitrary entry can branch to the right decode below. returns
 * NWEP_ERR_INTERNAL on a NULL/empty buffer or NWEP_ERR_PROTO_INVALID_MESSAGE on
 * an unknown type byte. */
int nwep_log_entry_type(const uint8_t *bytes, size_t len);

/* decode a KeyBinding / KeyRotation / Revocation entry into the typed struct.
 * parse only (no signature verification). */
int nwep_keybinding_decode(const uint8_t *bytes, size_t len, nwep_keybinding *out);
int nwep_keyrotation_decode(const uint8_t *bytes, size_t len, nwep_keyrotation *out);
int nwep_revocation_decode(const uint8_t *bytes, size_t len, nwep_revocation *out);

/* DHT NW110000.
 *
 * the DHT shares a UDP socket with a running server NW180100
 * demuxes by first byte: 0x80/0x81 -> DHT, otherwise QUIC. create the
 * server first, then attach a DHT to it via nwep_dht_attach. the DHT
 * borrows the server's socket; closing the server closes the DHT too.
 *
 * bootstrap entries are 32B NodeID + IPv6 address. the text format from
 * NW110900 (`<NodeID_base58>@[<ipv6>]:<port>`) is parsed by
 * nwep_dht_parse_bootstrap. */

typedef struct nwep_dht nwep_dht;

typedef struct {
    nwep_node_id node_id;
    nwep_address addr;
} nwep_bootstrap_entry;

/* parses a "<NodeID_base58>@[<ipv6>]:<port>" entry per NW110900. returns 0 or
 * negative NWEP_ERR_* (NWEP_ERR_CONFIG_INVALID on NULL input or parse
 * failure). */
int nwep_dht_parse_bootstrap(nwep_bootstrap_entry *out,
                             const char *input, size_t len);

/* attaches a DHT to a running server, reusing its UDP socket and identity.
 * bootstrap_nodes must contain at least one entry NW110900.
 * initial_seq is the last successfully-announced record seq from a previous
 * run; pass 0 on first boot NW110600. calling this twice on the same server
 * returns NWEP_ERR_CONFIG_INVALID (-101)  -  one DHT per server. */
int nwep_dht_attach(nwep_dht **out,
                    nwep_server *server,
                    const nwep_bootstrap_entry *bootstrap_nodes,
                    size_t bootstrap_count,
                    uint64_t initial_seq);

/* sends PING + FIND_NODE to every bootstrap peer. responses arrive
 * asynchronously via nwep_server_tick; call nwep_dht_tick afterwards. */
int nwep_dht_bootstrap(nwep_dht *dht, uint64_t now_secs);

/* publishes a signed record binding this node to `service_addr`. caller
 * should re-call every NWEP_DHT_REPUBLISH_INTERVAL_SECS NW110700. */
int nwep_dht_announce(nwep_dht *dht,
                      const nwep_address *service_addr,
                      uint64_t now_secs);

/* begins a FIND_VALUE lookup for `target_node_id`. returns immediately;
 * poll nwep_dht_lookup_result for the cached record after responses are
 * absorbed via nwep_server_tick. */
int nwep_dht_start_lookup(nwep_dht *dht,
                          const nwep_node_id *target_node_id,
                          uint64_t now_secs);

/* returns the cached record for `target_node_id` if one has been observed.
 * returns 0 on hit (and fills out_record), NWEP_ERR_APP_NOT_FOUND on miss. */
typedef struct {
    nwep_node_id node_id;
    nwep_address addr;
    uint8_t      pubkey[NWEP_PUBKEY_SIZE];
    uint64_t     seq;
    uint64_t     timestamp;
} nwep_dht_record;

int nwep_dht_lookup_result(const nwep_dht *dht,
                           const nwep_node_id *target_node_id,
                           nwep_dht_record *out_record);

/* advances DHT timers (refresh, expiry, retransmit). call from your event
 * loop alongside nwep_server_tick. */
int nwep_dht_tick(nwep_dht *dht, uint64_t now_secs);

/* milliseconds until the DHT's next timer is due, for a poll()-style wait  - 
 * the DHT counterpart to nwep_server_next_timeout_ms. returns -1 when no
 * transaction is outstanding (block on socket readability alone; an idle DHT
 * node then costs ~0% CPU) or on a NULL handle, and 0 when a deadline has
 * already passed. now_secs is Unix seconds (the DHT clock). fold this into the
 * same poll timeout as the server's accessor (take the minimum of the two). */
int nwep_dht_next_timeout_ms(const nwep_dht *dht, uint64_t now_secs);

/* DHT traffic counters. the DHT shares the server's UDP socket, but its
 * datagrams never touch the server's send path, so nwep_server_metrics_get
 * cannot see DHT traffic  -  scrape this for the DHT half. (Inbound is symmetric
 * either way: the server counts every received datagram before demuxing it to
 * the DHT, and the DHT counts it again here.) Cumulative; pull model. */
typedef struct {
  uint64_t datagrams_sent;     /* cumulative: DHT datagrams out */
  uint64_t datagrams_received; /* cumulative: DHT datagrams in (incl. dropped) */
  uint64_t bytes_sent;         /* cumulative: DHT UDP payload bytes out */
  uint64_t bytes_received;     /* cumulative: DHT UDP payload bytes in */
} nwep_dht_metrics;

/* fills *out with the DHT's current traffic snapshot. returns 0, or -801 on a
 * NULL handle/out. */
int nwep_dht_metrics_get(const nwep_dht *dht, nwep_dht_metrics *out);

/* detaches and frees the DHT handle. does NOT close the server's socket. */
void nwep_dht_close(nwep_dht *dht);

/* wire-format helpers */

/* returns the lowercase ASCII name of a method enum value, or NULL on an
 * unknown index. the returned pointer is to a static literal  -  do not free.
 * indices match the Zig enum tag order:
 *   0=read, 1=write, 2=update, 3=delete, 4=connect, 5=authenticate, 6=heartbeat */
const char *nwep_method_str(int method);

/* returns the lowercase ASCII token for a NWEP_STATUS_* code, or NULL on an
 * unknown index. the returned pointer is to a static literal  -  do not free.
 * see NWEP_STATUS_* defines above for the full mapping NW080000. */
const char *nwep_status_str(int status);

/*
 * deferred  -  not part of v0.1.0
 *
 *   - `nw3-client` example: NodeID-only URI form (`web://NodeID/path`)
 *     would require the example to additionally bring up a server +
 *     DHT for resolution. the example deliberately keeps the explicit-
 *     address form (`web://NodeID@[ipv6]:port/path`). use
 *     nwep_client_connect_by_nodeid directly from your application.
 *
 * every API in this header is available on linux / android / windows
 * (x86 / x86_64 / arm / aarch64).
 *
 * example programs (built via `zig build examples`):
 *   - keygen <output.pem>                             -  generate identity
 *   - nw3-server <identity.pem> [port]                -  serves READ /hello
 *   - nw3-client <identity.pem> <web://NodeID@...>    -  READ + print body
 */

#ifdef __cplusplus
}
#endif

#endif /* NWEP_H */
