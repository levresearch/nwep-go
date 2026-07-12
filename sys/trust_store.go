// the layer 0 trust store, the verified anchor set and checkpoint chain NW120000.

package sys

/*
#include <stdlib.h>
#include <nwep.h>
#include <nwep_trust.h>
*/
import "C"

import "unsafe"

// TrustStoreCreate allocates an empty trust store (nwep_trust_store_create).
func TrustStoreCreate() unsafe.Pointer {
	return unsafe.Pointer(C.nwep_trust_store_create())
}

// TrustStoreFree frees a trust store (nwep_trust_store_free).
func TrustStoreFree(ts unsafe.Pointer) {
	C.nwep_trust_store_free((*C.nwep_trust_store)(ts))
}

// TrustStoreLoadGenesisAnchors seeds the store with n concatenated 48-byte anchor keys (nwep_trust_store_load_genesis_anchors).
func TrustStoreLoadGenesisAnchors(ts unsafe.Pointer, pubkeys []byte, n int) int {
	return int(C.nwep_trust_store_load_genesis_anchors((*C.nwep_trust_store)(ts), bytePtr(pubkeys), C.size_t(n)))
}

// TrustStoreUpdateCheckpoint verifies and adopts a newer checkpoint (nwep_trust_store_update_checkpoint).
func TrustStoreUpdateCheckpoint(ts unsafe.Pointer, cpBytes []byte, nowSecs int64) int {
	return int(C.nwep_trust_store_update_checkpoint((*C.nwep_trust_store)(ts), bytePtr(cpBytes), C.size_t(len(cpBytes)), C.int64_t(nowSecs)))
}

// CheckpointVerify checks a checkpoint against the store's anchors without adopting it (nwep_checkpoint_verify).
func CheckpointVerify(ts unsafe.Pointer, cpBytes []byte, nowSecs int64) int {
	return int(C.nwep_checkpoint_verify((*C.nwep_trust_store)(ts), bytePtr(cpBytes), C.size_t(len(cpBytes)), C.int64_t(nowSecs)))
}

// TrustStoreApplyAnchorChange applies a signed anchor-set change at an epoch (nwep_trust_store_apply_anchor_change).
func TrustStoreApplyAnchorChange(ts unsafe.Pointer, entryBytes []byte, currentEpoch uint64) int {
	return int(C.nwep_trust_store_apply_anchor_change((*C.nwep_trust_store)(ts), bytePtr(entryBytes), C.size_t(len(entryBytes)), C.uint64_t(currentEpoch)))
}

// TrustStoreObserveLogSize records a seen log size to detect rollback (nwep_trust_store_observe_log_size).
func TrustStoreObserveLogSize(ts unsafe.Pointer, observed uint64) int {
	return int(C.nwep_trust_store_observe_log_size((*C.nwep_trust_store)(ts), C.uint64_t(observed)))
}

// TrustStoreMaxLogSize returns the largest log size the store has seen (nwep_trust_store_max_log_size).
func TrustStoreMaxLogSize(ts unsafe.Pointer) uint64 {
	return uint64(C.nwep_trust_store_max_log_size((*C.nwep_trust_store)(ts)))
}

// TrustStoreSave serializes the store for persistence (nwep_trust_store_save).
func TrustStoreSave(ts unsafe.Pointer) ([]byte, int) {
	var outlen C.size_t
	rc := int(C.nwep_trust_store_save((*C.nwep_trust_store)(ts), nil, &outlen))
	if rc != 0 {
		return nil, rc
	}
	out := make([]byte, outlen)
	rc = int(C.nwep_trust_store_save((*C.nwep_trust_store)(ts), bytePtr(out), &outlen))
	if rc != 0 {
		return nil, rc
	}
	return out[:outlen], 0
}

// TrustStoreLoad restores a store from saved bytes, re-verifying as it loads (nwep_trust_store_load).
func TrustStoreLoad(ts unsafe.Pointer, bytes []byte) int {
	return int(C.nwep_trust_store_load((*C.nwep_trust_store)(ts), bytePtr(bytes), C.size_t(len(bytes))))
}

// TrustStoreVerifyKey resolves and verifies a node's current key via a log client (nwep_trust_store_verify_key).
func TrustStoreVerifyKey(ts, client unsafe.Pointer, nodeID, recoveryCommitment [32]byte, nowSecs int64) int {
	return int(C.nwep_trust_store_verify_key((*C.nwep_trust_store)(ts), (*C.nwep_client)(client), arrPtr(&nodeID), arrPtr(&recoveryCommitment), C.int64_t(nowSecs)))
}

// TrustStoreVerifyKeyBinding verifies a key binding bundle offline against the store (nwep_trust_store_verify_key_binding).
func TrustStoreVerifyKeyBinding(ts unsafe.Pointer, nodeID, expectedPubkey [32]byte, bundle []byte, nowSecs int64) int {
	return int(C.nwep_trust_store_verify_key_binding((*C.nwep_trust_store)(ts), arrPtr(&nodeID), arrPtr(&expectedPubkey), bytePtr(bundle), C.size_t(len(bundle)), C.int64_t(nowSecs)))
}

// TrustStoreEvaluateKeyRotation checks whether a rotation justifies a presented key (nwep_trust_store_evaluate_key_rotation).
func TrustStoreEvaluateKeyRotation(rotationBytes []byte, presentedPubkey [32]byte, nowSecs int64) int {
	return int(C.nwep_trust_store_evaluate_key_rotation(bytePtr(rotationBytes), C.size_t(len(rotationBytes)), arrPtr(&presentedPubkey), C.int64_t(nowSecs)))
}
