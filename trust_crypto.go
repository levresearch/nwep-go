// the trust crypto primitives, bls threshold signing and key-entry creation NW150500.

package nwep

import (
	"unsafe"

	"github.com/levresearch/nwep-go/sys"
)

// BLSKeygen generates a bls12-381 secret and public key pair (nwep_bls_keygen).
func BLSKeygen() (sk [32]byte, pk [48]byte, err error) {
	sk, pk, rc := sys.BLSKeygen()
	return sk, pk, check(rc)
}

// BLSSign signs msg with a bls secret key, returning the 96-byte signature (nwep_bls_sign).
func BLSSign(sk [32]byte, msg []byte) ([96]byte, error) {
	sig, rc := sys.BLSSign(sk, msg)
	return sig, check(rc)
}

// BLSVerify verifies a bls signature over msg under pk (nwep_bls_verify).
func BLSVerify(sig [96]byte, pk [48]byte, msg []byte) bool {
	return sys.BLSVerify(sig, pk, msg) == 0
}

// BLSAggregate aggregates n concatenated 96-byte signatures into one (nwep_bls_aggregate).
func BLSAggregate(sigs []byte, n int) ([96]byte, error) {
	sig, rc := sys.BLSAggregate(sigs, n)
	return sig, check(rc)
}

// BLSVerifyAggregate verifies an aggregate signature over msg under n keys (nwep_bls_verify_aggregate).
//
// pks is n concatenated 48-byte public keys, all signing the same msg.
func BLSVerifyAggregate(aggSig [96]byte, pks []byte, n int, msg []byte) bool {
	return sys.BLSVerifyAggregate(aggSig, pks, n, msg) == 0
}

// Checkpoint is a decoded, anchor-co-signed snapshot of the log at an epoch NW120000.
type Checkpoint struct {
	ptr unsafe.Pointer
}

// DecodeCheckpoint decodes checkpoint bytes into an owned handle (nwep_checkpoint_decode).
func DecodeCheckpoint(bytes []byte) (*Checkpoint, error) {
	ptr, rc := sys.CheckpointDecode(bytes)
	if err := check(rc); err != nil {
		return nil, err
	}
	return &Checkpoint{ptr: ptr}, nil
}

// Staleness returns how stale the checkpoint is at nowSecs, in seconds NW120000.
func (cp *Checkpoint) Staleness(nowSecs int64) int { return sys.CheckpointStaleness(cp.ptr, nowSecs) }

// Close frees the decoded checkpoint (nwep_checkpoint_free).
func (cp *Checkpoint) Close() {
	if cp.ptr != nil {
		sys.CheckpointFree(cp.ptr)
		cp.ptr = nil
	}
}

// Raw returns the underlying sys checkpoint pointer, the no-cliffs escape NWG0200.
func (cp *Checkpoint) Raw() unsafe.Pointer { return cp.ptr }

// CreateGenesisCheckpoint builds the founding checkpoint for a fresh network NW120000.
//
// blsSecrets is nFounders concatenated 32-byte secrets, blsPubkeys nFounders
// concatenated 48-byte keys, indices one byte per founder. returns the encoded
// genesis checkpoint bytes.
func CreateGenesisCheckpoint(blsSecrets, blsPubkeys, indices []byte, nFounders, threshold int) ([]byte, error) {
	b, rc := sys.GenesisCheckpointCreate(blsSecrets, blsPubkeys, indices, nFounders, threshold)
	if err := check(rc); err != nil {
		return nil, err
	}
	return b, nil
}

// CreateKeyBinding builds a signed key binding entry (nwep_keybinding_create).
func CreateKeyBinding(pubkey [32]byte, recoveryCommitment [32]byte, timestamp uint64, privkey [32]byte) ([]byte, error) {
	b, rc := sys.KeybindingCreate(pubkey, recoveryCommitment, timestamp, privkey)
	if err := check(rc); err != nil {
		return nil, err
	}
	return b, nil
}

// CreateKeyRotation builds a signed key rotation entry, signed by both keys (nwep_keyrotation_create).
func CreateKeyRotation(nodeID [32]byte, oldPubkey, newPubkey [32]byte, timestamp, overlapExpiry uint64, oldPrivkey, newPrivkey [32]byte) ([]byte, error) {
	b, rc := sys.KeyrotationCreate(nodeID, oldPubkey, newPubkey, timestamp, overlapExpiry, oldPrivkey, newPrivkey)
	if err := check(rc); err != nil {
		return nil, err
	}
	return b, nil
}

// CreateRevocation builds a signed revocation entry, signed by the recovery key (nwep_revocation_create).
func CreateRevocation(nodeID [32]byte, revokedPubkey, recoveryPubkey [32]byte, reason uint8, timestamp uint64, recoveryPrivkey [32]byte) ([]byte, error) {
	b, rc := sys.RevocationCreate(nodeID, revokedPubkey, recoveryPubkey, reason, timestamp, recoveryPrivkey)
	if err := check(rc); err != nil {
		return nil, err
	}
	return b, nil
}
