// the trust layer, the bls anchor set and checkpoint chain NW120000 NW150500.
//
// these wrap the complete libnwep, the trust build, not libnwep_core. a
// TrustStore holds the verified anchor public keys and the latest checkpoint, and
// is what a client consults to verify a node's current key against the log. the
// bls primitives, checkpoint decoding, and anchor co-signing round out the layer.

package nwep

import (
	"unsafe"

	"github.com/levresearch/nwep-go/sys"
)

// TrustVersion returns the trust-layer version string (nwep_trust_version).
func TrustVersion() string { return sys.TrustVersion() }

// the bls12-381 key and signature sizes NW150500.
const (
	BLSPubkeySize    = 48
	BLSSeckeySize    = 32
	BLSSignatureSize = 96
)

// the key-management log entry types, re-exported so callers can branch on EntryType.
type (
	KeyBinding  = sys.KeyBinding
	KeyRotation = sys.KeyRotation
	Revocation  = sys.Revocation
)

// TrustStore is the verified anchor set and checkpoint chain a client trusts NW120000.
type TrustStore struct {
	ptr unsafe.Pointer
}

// NewTrustStore allocates an empty trust store (nwep_trust_store_create).
func NewTrustStore() *TrustStore { return &TrustStore{ptr: sys.TrustStoreCreate()} }

// LoadGenesisAnchors seeds the store with the founding anchor public keys NW120000.
//
// pubkeys is n concatenated 48-byte bls public keys.
func (ts *TrustStore) LoadGenesisAnchors(pubkeys []byte, n int) error {
	return check(sys.TrustStoreLoadGenesisAnchors(ts.ptr, pubkeys, n))
}

// UpdateCheckpoint verifies and adopts a newer checkpoint (nwep_trust_store_update_checkpoint).
//
// errors with a trust code on a stale, under-threshold, or invalid checkpoint, and
// the fatal equivocation code on a conflicting checkpoint for the same epoch.
func (ts *TrustStore) UpdateCheckpoint(cpBytes []byte, nowSecs int64) error {
	return check(sys.TrustStoreUpdateCheckpoint(ts.ptr, cpBytes, nowSecs))
}

// VerifyCheckpoint checks a checkpoint against the anchors without adopting it (nwep_checkpoint_verify).
func (ts *TrustStore) VerifyCheckpoint(cpBytes []byte, nowSecs int64) error {
	return check(sys.CheckpointVerify(ts.ptr, cpBytes, nowSecs))
}

// ApplyAnchorChange applies a signed anchor-set change at currentEpoch NW120000.
func (ts *TrustStore) ApplyAnchorChange(entryBytes []byte, currentEpoch uint64) error {
	return check(sys.TrustStoreApplyAnchorChange(ts.ptr, entryBytes, currentEpoch))
}

// ObserveLogSize records a seen log size, so a later rollback is detectable NW120000.
func (ts *TrustStore) ObserveLogSize(observed uint64) error {
	return check(sys.TrustStoreObserveLogSize(ts.ptr, observed))
}

// MaxLogSize returns the largest log size the store has seen.
func (ts *TrustStore) MaxLogSize() uint64 { return sys.TrustStoreMaxLogSize(ts.ptr) }

// Save serializes the store for persistence (nwep_trust_store_save).
func (ts *TrustStore) Save() ([]byte, error) {
	b, rc := sys.TrustStoreSave(ts.ptr)
	if err := check(rc); err != nil {
		return nil, err
	}
	return b, nil
}

// Load restores a store from saved bytes, re-verifying as it loads.
func (ts *TrustStore) Load(bytes []byte) error {
	return check(sys.TrustStoreLoad(ts.ptr, bytes))
}

// VerifyKey resolves and verifies a node's current key through a log client NW120000.
//
// drives client to fetch the node's key-management history from the log and checks
// it against the store's checkpoint. errors with a trust or identity code when the
// key cannot be verified or has been revoked.
func (ts *TrustStore) VerifyKey(client *Client, nodeID, recoveryCommitment [32]byte, nowSecs int64) error {
	return check(sys.TrustStoreVerifyKey(ts.ptr, client.ptr, nodeID, recoveryCommitment, nowSecs))
}

// VerifyKeyBinding verifies a key binding bundle offline against the store NW120000.
func (ts *TrustStore) VerifyKeyBinding(nodeID, expectedPubkey [32]byte, bundle []byte, nowSecs int64) error {
	return check(sys.TrustStoreVerifyKeyBinding(ts.ptr, nodeID, expectedPubkey, bundle, nowSecs))
}

// Close frees the trust store (nwep_trust_store_free).
func (ts *TrustStore) Close() {
	if ts.ptr != nil {
		sys.TrustStoreFree(ts.ptr)
		ts.ptr = nil
	}
}

// Raw returns the underlying sys trust store pointer, the no-cliffs escape NWG0200.
func (ts *TrustStore) Raw() unsafe.Pointer { return ts.ptr }

// EvaluateKeyRotation reports whether a rotation justifies a presented key NW150000.
func EvaluateKeyRotation(rotationBytes []byte, presentedPubkey [32]byte, nowSecs int64) error {
	return check(sys.TrustStoreEvaluateKeyRotation(rotationBytes, presentedPubkey, nowSecs))
}

// DecodeKeyBinding decodes a key binding log entry (nwep_keybinding_decode).
func DecodeKeyBinding(entry []byte) (KeyBinding, error) {
	b, rc := sys.KeybindingDecode(entry)
	return b, check(rc)
}

// DecodeKeyRotation decodes a key rotation log entry (nwep_keyrotation_decode).
func DecodeKeyRotation(entry []byte) (KeyRotation, error) {
	r, rc := sys.KeyrotationDecode(entry)
	return r, check(rc)
}

// DecodeRevocation decodes a revocation log entry (nwep_revocation_decode).
func DecodeRevocation(entry []byte) (Revocation, error) {
	r, rc := sys.RevocationDecode(entry)
	return r, check(rc)
}
