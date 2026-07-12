// the layer 0 key-management log entries, binding, rotation, and revocation NW150000.

package sys

/*
#include <nwep.h>
*/
import "C"

import "unsafe"

// arrPtr returns a c uint8 pointer to the first element of a fixed byte array.
func arrPtr[T ~[1]byte | ~[16]byte | ~[32]byte | ~[48]byte | ~[64]byte | ~[96]byte](a *T) *C.uint8_t {
	return (*C.uint8_t)(unsafe.Pointer(a))
}

// KeyBinding mirrors nwep_keybinding, a node binding its pubkey and recovery commitment.
type KeyBinding struct {
	NodeID             [NodeIDSize]byte
	Pubkey             [PubKeySize]byte
	RecoveryCommitment [32]byte
	Timestamp          uint64
	Signature          [64]byte
}

// KeyRotation mirrors nwep_keyrotation, a signed handover from an old to a new key.
type KeyRotation struct {
	NodeID        [NodeIDSize]byte
	OldPubkey     [PubKeySize]byte
	NewPubkey     [PubKeySize]byte
	Timestamp     uint64
	OverlapExpiry uint64
	SigOld        [64]byte
	SigNew        [64]byte
}

// Revocation mirrors nwep_revocation, a signed retirement of a compromised key.
type Revocation struct {
	NodeID         [NodeIDSize]byte
	RevokedPubkey  [PubKeySize]byte
	RecoveryPubkey [PubKeySize]byte
	Reason         uint8
	Timestamp      uint64
	Signature      [64]byte
}

// KeybindingCreate builds a signed key binding entry (nwep_keybinding_create).
func KeybindingCreate(pubkey [PubKeySize]byte, recoveryCommitment [32]byte, timestamp uint64, privkey [PrivKeySize]byte) ([]byte, int) {
	var outlen C.size_t
	probe := func(out *C.uint8_t) C.int {
		return C.nwep_keybinding_create(arrPtr(&pubkey), arrPtr(&recoveryCommitment), C.uint64_t(timestamp), arrPtr(&privkey), out, &outlen)
	}
	if rc := int(probe(nil)); rc != 0 {
		return nil, rc
	}
	out := make([]byte, outlen)
	if rc := int(probe(bytePtr(out))); rc != 0 {
		return nil, rc
	}
	return out[:outlen], 0
}

// KeyrotationCreate builds a signed key rotation entry, signed by both keys (nwep_keyrotation_create).
func KeyrotationCreate(nodeID [NodeIDSize]byte, oldPubkey, newPubkey [PubKeySize]byte, timestamp, overlapExpiry uint64, oldPrivkey, newPrivkey [PrivKeySize]byte) ([]byte, int) {
	var outlen C.size_t
	probe := func(out *C.uint8_t) C.int {
		return C.nwep_keyrotation_create(arrPtr(&nodeID), arrPtr(&oldPubkey), arrPtr(&newPubkey), C.uint64_t(timestamp), C.uint64_t(overlapExpiry), arrPtr(&oldPrivkey), arrPtr(&newPrivkey), out, &outlen)
	}
	if rc := int(probe(nil)); rc != 0 {
		return nil, rc
	}
	out := make([]byte, outlen)
	if rc := int(probe(bytePtr(out))); rc != 0 {
		return nil, rc
	}
	return out[:outlen], 0
}

// RevocationCreate builds a signed revocation entry, signed by the recovery key (nwep_revocation_create).
func RevocationCreate(nodeID [NodeIDSize]byte, revokedPubkey, recoveryPubkey [PubKeySize]byte, reason uint8, timestamp uint64, recoveryPrivkey [PrivKeySize]byte) ([]byte, int) {
	var outlen C.size_t
	probe := func(out *C.uint8_t) C.int {
		return C.nwep_revocation_create(arrPtr(&nodeID), arrPtr(&revokedPubkey), arrPtr(&recoveryPubkey), C.uint8_t(reason), C.uint64_t(timestamp), arrPtr(&recoveryPrivkey), out, &outlen)
	}
	if rc := int(probe(nil)); rc != 0 {
		return nil, rc
	}
	out := make([]byte, outlen)
	if rc := int(probe(bytePtr(out))); rc != 0 {
		return nil, rc
	}
	return out[:outlen], 0
}

// KeybindingDecode decodes a key binding entry (nwep_keybinding_decode).
func KeybindingDecode(bytes []byte) (KeyBinding, int) {
	var c C.nwep_keybinding
	rc := int(C.nwep_keybinding_decode(bytePtr(bytes), C.size_t(len(bytes)), &c))
	var b KeyBinding
	if rc == 0 {
		b.NodeID = *(*[NodeIDSize]byte)(unsafe.Pointer(&c.node_id[0]))
		b.Pubkey = *(*[PubKeySize]byte)(unsafe.Pointer(&c.pubkey[0]))
		b.RecoveryCommitment = *(*[32]byte)(unsafe.Pointer(&c.recovery_commitment[0]))
		b.Timestamp = uint64(c.timestamp)
		b.Signature = *(*[64]byte)(unsafe.Pointer(&c.signature[0]))
	}
	return b, rc
}

// KeyrotationDecode decodes a key rotation entry (nwep_keyrotation_decode).
func KeyrotationDecode(bytes []byte) (KeyRotation, int) {
	var c C.nwep_keyrotation
	rc := int(C.nwep_keyrotation_decode(bytePtr(bytes), C.size_t(len(bytes)), &c))
	var r KeyRotation
	if rc == 0 {
		r.NodeID = *(*[NodeIDSize]byte)(unsafe.Pointer(&c.node_id[0]))
		r.OldPubkey = *(*[PubKeySize]byte)(unsafe.Pointer(&c.old_pubkey[0]))
		r.NewPubkey = *(*[PubKeySize]byte)(unsafe.Pointer(&c.new_pubkey[0]))
		r.Timestamp = uint64(c.timestamp)
		r.OverlapExpiry = uint64(c.overlap_expiry)
		r.SigOld = *(*[64]byte)(unsafe.Pointer(&c.sig_old[0]))
		r.SigNew = *(*[64]byte)(unsafe.Pointer(&c.sig_new[0]))
	}
	return r, rc
}

// RevocationDecode decodes a revocation entry (nwep_revocation_decode).
func RevocationDecode(bytes []byte) (Revocation, int) {
	var c C.nwep_revocation
	rc := int(C.nwep_revocation_decode(bytePtr(bytes), C.size_t(len(bytes)), &c))
	var r Revocation
	if rc == 0 {
		r.NodeID = *(*[NodeIDSize]byte)(unsafe.Pointer(&c.node_id[0]))
		r.RevokedPubkey = *(*[PubKeySize]byte)(unsafe.Pointer(&c.revoked_pubkey[0]))
		r.RecoveryPubkey = *(*[PubKeySize]byte)(unsafe.Pointer(&c.recovery_pubkey[0]))
		r.Reason = uint8(c.reason)
		r.Timestamp = uint64(c.timestamp)
		r.Signature = *(*[64]byte)(unsafe.Pointer(&c.signature[0]))
	}
	return r, rc
}
