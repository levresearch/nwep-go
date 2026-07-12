#ifndef NWEP_TRUST_H
#define NWEP_TRUST_H

#include <stddef.h>
#include <stdint.h>
#include "nwep.h"

#ifdef __cplusplus
extern "C" {
#endif

/*
 * threading contract
 *
 * same as libnwep: single-threaded, non-reentrant. all time inputs are in
 * unix seconds (not milliseconds  -  the staleness thresholds in NW120700
 * are seconds, and a 1000x error makes every checkpoint look stale).
 *
 * persistence
 *
 * `epoch_roots` (the per-epoch root history used for equivocation
 * detection) and `max_log_size` (the rollback-attack guard) must survive
 * process restarts. persistence is caller-owned  -  the library does not
 * touch the filesystem. NW000014's log-server work will expose a
 * serialize/deserialize entry point; for now embedders persist by reading
 * the in-memory state directly.
 *
 * bls secret-key hygiene
 *
 * `nwep_bls_keygen` writes 32 bytes of secret material. zeroize with
 * `nwep_zeroize(sk, 32)` before passing the buffer to free(). this is a
 * must for production builds NW030300.
 */

#define NWEP_TRUST_VERSION_MAJOR 0
#define NWEP_TRUST_VERSION_MINOR 1
#define NWEP_TRUST_VERSION_PATCH 0

const char *nwep_trust_version(void);

/* constants */
#define NWEP_BLS_PUBKEY_SIZE        48
#define NWEP_BLS_SECKEY_SIZE        32
#define NWEP_BLS_SIGNATURE_SIZE     96
#define NWEP_CHECKPOINT_SIGN_DATA   56

/* opaque handles */
typedef struct nwep_trust_store nwep_trust_store;
typedef struct nwep_checkpoint  nwep_checkpoint;

/* BLS primitives */

/* generates a BLS12-381 keypair. returns 0 on success or negative NWEP_ERR_*.
 * the 32-byte secret in `out_sk` must be zeroized via
 * `nwep_zeroize(out_sk, 32)` before disposal. */
int nwep_bls_keygen(uint8_t out_sk[NWEP_BLS_SECKEY_SIZE],
                    uint8_t out_pk[NWEP_BLS_PUBKEY_SIZE]);

/* signs `msg` under `sk` with domain tag "WEB/1-CHECKPOINT". always returns 0
 * (no error path  -  the int return type is for ABI consistency). */
int nwep_bls_sign(uint8_t out_sig[NWEP_BLS_SIGNATURE_SIZE],
                  const uint8_t sk[NWEP_BLS_SECKEY_SIZE],
                  const uint8_t *msg, size_t msg_len);

/* verifies a single-signer BLS signature. */
int nwep_bls_verify(const uint8_t sig[NWEP_BLS_SIGNATURE_SIZE],
                    const uint8_t pk[NWEP_BLS_PUBKEY_SIZE],
                    const uint8_t *msg, size_t msg_len);

/* aggregates `n` BLS signatures (each 96 bytes, contiguous) into one. */
int nwep_bls_aggregate(uint8_t out_sig[NWEP_BLS_SIGNATURE_SIZE],
                       const uint8_t *sigs, size_t n);

/* verifies an aggregate signature against `n` pubkeys (each 48 bytes,
 * contiguous), all covering the same message. */
int nwep_bls_verify_aggregate(const uint8_t agg_sig[NWEP_BLS_SIGNATURE_SIZE],
                              const uint8_t *pks, size_t n,
                              const uint8_t *msg, size_t msg_len);

/* checkpoint */

typedef enum {
    NWEP_CHECKPOINT_FRESH   = 0,  /* age < CHECKPOINT_EPOCH_SECS */
    NWEP_CHECKPOINT_WARNING = 1,  /* EPOCH <= age <= WARN          */
    NWEP_CHECKPOINT_STALE   = 2   /* age > CHECKPOINT_WARN_SECS  */
} nwep_checkpoint_status;

/* decodes a checkpoint from wire bytes. allocates; caller frees with
 * nwep_checkpoint_free(). */
int nwep_checkpoint_decode(const uint8_t *bytes, size_t len, nwep_checkpoint **out_cp);

/* frees a checkpoint handle returned by nwep_checkpoint_decode. */
void nwep_checkpoint_free(nwep_checkpoint *cp);

/* returns the staleness band per NW120700, or negative NWEP_ERR_*. */
int nwep_checkpoint_staleness(const nwep_checkpoint *cp, int64_t now_secs);

/* runs the genesis ceremony NW121100 and encodes the network's epoch-0
 * checkpoint to wire bytes  -  the value committed as the hardcoded genesis to
 * bootstrap trust. every founding anchor signs; the aggregate is BLS-verified
 * before the bytes are produced.
 *
 * bls_secrets / bls_pubkeys / indices are n_founders contiguous 32-byte secret
 * keys, 48-byte public keys, and 1-byte 1-based share indices. threshold is
 * the quorum the genesis must satisfy. two-call sizing: out == NULL writes the
 * required size to *outlen and returns 0; a too-small buffer reports the size
 * and returns NWEP_ERR_INTERNAL.
 *
 * security: the founding secret keys are the root of all network trust. the
 * caller must nwep_zeroize(bls_secrets, ...) after this returns. */
int nwep_genesis_checkpoint_create(const uint8_t *bls_secrets,
                                   const uint8_t *bls_pubkeys,
                                   const uint8_t *indices,
                                   size_t n_founders, size_t threshold,
                                   uint8_t *out, size_t *outlen);

/* trust store */

/* creates an empty trust store. caller must seed the anchor set via
 * nwep_trust_store_load_genesis_anchors before any non-genesis
 * checkpoint will verify. */
nwep_trust_store *nwep_trust_store_create(void);

/* frees a trust store. */
void nwep_trust_store_free(nwep_trust_store *ts);

/* loads `n` genesis anchor BLS pubkeys (each 48 bytes, contiguous) into
 * the anchor set. returns 0 on success or negative NWEP_ERR_*. */
int nwep_trust_store_load_genesis_anchors(nwep_trust_store *ts,
                                          const uint8_t *pubkeys,
                                          size_t n);

/* installs a checkpoint. returns the staleness band on success
 * (NWEP_CHECKPOINT_FRESH/WARNING) or a negative NWEP_ERR_* on rejection.
 * stale checkpoints are rejected with NWEP_ERR_TRUST_STALE_CHECKPOINT  -
 * they are never installed.
 *
 * time is Unix seconds (not ms). */
int nwep_trust_store_update_checkpoint(nwep_trust_store *ts,
                                       const uint8_t *cp_bytes,
                                       size_t cp_len,
                                       int64_t now_secs);

/* verifies a checkpoint against the store's anchor set without installing it
 * NW120800 structural + threshold + BLS aggregate checks (including staleness:
 * a checkpoint whose signature-ts + max-age is in the past returns
 * NWEP_ERR_TRUST_STALE_CHECKPOINT). unlike nwep_trust_store_update_checkpoint,
 * it runs no equivocation guard and does not mutate the store. returns 0 if
 * valid, or a negative NWEP_ERR_*. time is unix seconds. */
int nwep_checkpoint_verify(const nwep_trust_store *ts,
                           const uint8_t *cp_bytes, size_t cp_len,
                           int64_t now_secs);

/* applies an AnchorChange log entry NW120300 to the store's anchor set:
 * decodes the raw variable-length entry (leading type byte 0x04, carrying its
 * signer-subset list), verifies that a quorum (>= threshold) of distinct
 * current members signed it, then adds/removes the anchor. current_epoch stamps
 * an added anchor's added_at_epoch. the caller must have already checked the
 * entry's node_id against a current KeyBinding NW120300  -  this verifies
 * quorum authorization, not the identity binding.
 * returns 0 or a negative NWEP_ERR_*. */
int nwep_trust_store_apply_anchor_change(nwep_trust_store *ts,
                                         const uint8_t *entry_bytes,
                                         size_t entry_len,
                                         uint64_t current_epoch);

/* bumps the rollback-protection counter from a non-checkpoint observation
 * (typically a /log/root response). refuses to go backwards. */
int nwep_trust_store_observe_log_size(nwep_trust_store *ts, uint64_t observed);

/* returns the current `max_log_size`. returns 0 for a fresh store or a NULL
 * handle  -  the two are indistinguishable, so do not pass NULL. */
uint64_t nwep_trust_store_max_log_size(const nwep_trust_store *ts);

/* serializes the rollback-critical state  -  max_log_size, the epoch-root
 * equivocation history, and the installed checkpoint  -  for persistence
 * across restarts NW120700 NW121000. two-call idiom:
 *   - out == NULL: *outlen receives the required size; returns 0.
 *   - out != NULL: writes the blob when *outlen >= required and sets *outlen
 *     to the byte count written. a too-small buffer sets *outlen to the
 *     required size and returns NWEP_ERR_INTERNAL.
 * the anchor set is not included  -  after load, reload genesis anchors and
 * replay AnchorChange entries as usual. */
int nwep_trust_store_save(const nwep_trust_store *ts, uint8_t *out, size_t *outlen);

/* restores state written by nwep_trust_store_save into an existing store,
 * replacing max_log_size, the epoch-root history, and the checkpoint. the
 * restored checkpoint is re-verified via a full BLS pairing check against the
 * current anchor set before being committed  -  load genesis anchors first.
 * on malformed input or failed verification the store is left unchanged. */
int nwep_trust_store_load(nwep_trust_store *ts, const uint8_t *bytes, size_t len);

/* networked verifyNodeKey NW120800 NW121000. issues
 * READ /log/revocation/<base58(node_id)> over `client` and validates the
 * server's answer against `ts`.
 *
 * `client` must be connected to the trusted log server: a no-revocation
 * assertion's `server-id` is checked against the connection's authenticated
 * peer NodeID. `recovery_commitment` is `node_id`'s 32-byte KeyBinding
 * commitment, needed to verify a revocation proof; pass NULL only when a
 * returned revocation should be treated as an error rather than verified.
 *
 * time is Unix seconds (not ms). returns:
 *    0   -  not revoked (assertion verified; rollback counter advanced).
 *    1   -  revoked (revocation proof + recovery-key signature verified).
 *   <0   -  negative NWEP_ERR_* (network, decode, signature, rollback, NULL). */
int nwep_trust_store_verify_key(nwep_trust_store *ts,
                                nwep_client *client,
                                const uint8_t node_id[32],
                                const uint8_t recovery_commitment[32],
                                int64_t now_secs);

/* verifies a node's KeyBinding against the installed checkpoint NW120800
 * step 1)  -  the foundational "this NodeID's key is in the trust log under a
 * checkpoint I trust" check. `bundle` is the 169-byte KeyBinding entry
 * followed by its encoded Merkle inclusion proof, assembled from two reads:
 * the entry from READ /log/entry/{idx} and the proof from READ
 * /log/proof/{idx} (which returns the proof only, not the entry). returns 0
 * if valid, or a negative
 * NWEP_ERR_* (TRUST_NO_CHECKPOINT, TRUST_STALE_CHECKPOINT, IDENTITY_MISMATCH,
 * CRYPTO_VERIFY, PROTO_INVALID_MESSAGE). time is Unix seconds. */
int nwep_trust_store_verify_key_binding(nwep_trust_store *ts,
                                        const uint8_t node_id[32],
                                        const uint8_t expected_pubkey[32],
                                        const uint8_t *bundle, size_t bundle_len,
                                        int64_t now_secs);

/* decides whether presented_pubkey is currently acceptable for a node that has
 * published a KeyRotation NW120800 step 4. rotation_bytes is the raw
 * 241-byte KeyRotation entry; this function verifies both self-contained
 * Ed25519 signatures (old-key and new-key) internally  -  the caller does not
 * need to pre-verify them. returns 0 (acceptable: the new key, or the old key within the
 * overlap window), NWEP_ERR_IDENTITY_REVOKED (old key past overlap),
 * NWEP_ERR_IDENTITY_MISMATCH (neither key), or a decode error. Unix seconds. */
int nwep_trust_store_evaluate_key_rotation(const uint8_t *rotation_bytes,
                                           size_t rotation_len,
                                           const uint8_t presented_pubkey[32],
                                           int64_t now_secs);

/* anchor / checkpoint production NW120900.
 *
 * an anchor both answers partial-signature requests from peers and, acting as
 * coordinator for an epoch, gathers partials from peers, aggregates them, and
 * publishes the checkpoint via WRITE /log/checkpoint.
 *
 * respond side: create an anchor node, feed it the epoch's verified log root
 * (nwep_anchor_node_collect_log_root), and route /anchor/partial-sig requests
 * through nwep_anchor_node_dispatch from your server handler.
 *
 * coordinate side: for each peer anchor, connect a client and call
 * nwep_anchor_request_partial_sig; collect your own partial the same way you
 * answer one; then nwep_anchor_finish_checkpoint to produce the bytes for a
 * WRITE /log/checkpoint (sent with nwep_client_send). */

typedef struct nwep_anchor_node nwep_anchor_node;

/* creates an anchor node from its WEB/1 keypair and BLS share. share_index is
 * 1-based (1..n). collection_window_ms is the partial-sig collection window
 * NW120900 55 min = 3300000. NULL on failure. */
nwep_anchor_node *nwep_anchor_node_create(const uint8_t pubkey[NWEP_PUBKEY_SIZE],
                                          const uint8_t privkey[NWEP_PRIVKEY_SIZE],
                                          const uint8_t bls_secret[32],
                                          const uint8_t bls_pubkey[48],
                                          uint64_t share_index,
                                          uint64_t collection_window_ms);
void nwep_anchor_node_free(nwep_anchor_node *node);

/* records a /log/root?epoch=N snapshot (server_root, server_log_size) the
 * embedder fetched, cross-checked against its own replica (local_root). the
 * anchor will not sign a partial for an epoch whose root it has not collected.
 * returns 0 or negative NWEP_ERR_* (TRUST_FATAL_LOG_CORRUPT on mismatch). */
int nwep_anchor_node_collect_log_root(nwep_anchor_node *node, uint64_t epoch,
                                      const uint8_t server_root[32],
                                      uint64_t server_log_size,
                                      const uint8_t local_root[32]);

/* answers a READ /anchor/partial-sig from a peer. requester_node_id is the
 * authenticated peer NodeID (nwep_server_get_peer_nodeid); anchor_ids is the
 * anchor set's NodeIDs (n_anchors * 32 contiguous bytes). returns 0 (handled),
 * 1 (not this route  -  fall through), or negative NWEP_ERR_*. */
int nwep_anchor_node_dispatch(nwep_anchor_node *node,
                              const uint8_t requester_node_id[32],
                              const uint8_t *anchor_ids, size_t n_anchors,
                              const nwep_message *req, nwep_buf *resp,
                              int64_t now_secs);

/* produces this anchor's own partial signature for `epoch` (the coordinator's
 * contribution NW120600. writes the 1-byte share index and 96-byte
 * signature. */
int nwep_anchor_node_produce_partial_sig(nwep_anchor_node *node, uint64_t epoch,
                                         const uint8_t merkle_root[32],
                                         uint64_t log_size,
                                         uint8_t *out_index, uint8_t out_sig[96]);

/* coordinator: requests one peer's partial signature over `client` and
 * verifies it against peer_bls_pubkey before returning. writes the 1-byte
 * share index and 96-byte signature. any non-"ok" response from the peer
 * (forbidden or conflict/wrong epoch) yields NWEP_ERR_APP_FORBIDDEN; a bad
 * signature yields NWEP_ERR_CRYPTO_VERIFY. */
int nwep_anchor_request_partial_sig(nwep_client *client, uint64_t epoch,
                                    const uint8_t merkle_root[32],
                                    uint64_t log_size,
                                    const uint8_t peer_bls_pubkey[48],
                                    uint8_t *out_index, uint8_t out_sig[96]);

/* aggregates gathered partials into a checkpoint and encodes it for a
 * WRITE /log/checkpoint. indices/sigs are the partials (n_partials * 1 and
 * * 96 bytes); anchor_bls_pks is the ordered anchor set (n_anchors * 48),
 * with each partial's index 1-based into it. two-call sizing (out == NULL
 * queries the size). NWEP_ERR_TRUST_THRESHOLD if too few partials. */
int nwep_anchor_finish_checkpoint(uint64_t epoch, const uint8_t merkle_root[32],
                                  uint64_t log_size,
                                  const uint8_t *indices, const uint8_t *sigs,
                                  size_t n_partials,
                                  const uint8_t *anchor_bls_pks, size_t n_anchors,
                                  uint8_t *out, size_t *outlen);

#ifdef __cplusplus
}
#endif

#endif /* NWEP_TRUST_H */
