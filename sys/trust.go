// the layer 0 trust primitives, bls threshold signing and checkpoints NW120000 NW150500.
//
// these live in the complete libnwep, the trust build, not libnwep_core. linking
// libnwep makes them present.

package sys

/*
#include <stdlib.h>
#include <nwep.h>
#include <nwep_trust.h>
*/
import "C"

import "unsafe"

// the bls12-381 key and signature sizes from the trust header NW150500.
const (
	BLSPubkeySize    = int(C.NWEP_BLS_PUBKEY_SIZE)
	BLSSeckeySize    = int(C.NWEP_BLS_SECKEY_SIZE)
	BLSSignatureSize = int(C.NWEP_BLS_SIGNATURE_SIZE)
)

// TrustVersion returns the trust-layer version string (nwep_trust_version).
func TrustVersion() string {
	return C.GoString(C.nwep_trust_version())
}

// BLSKeygen generates a bls secret and public key pair (nwep_bls_keygen).
func BLSKeygen() (sk [32]byte, pk [48]byte, rc int) {
	rc = int(C.nwep_bls_keygen(arrPtr(&sk), arrPtr(&pk)))
	return
}

// BLSSign signs msg with a bls secret key, returning the 96-byte signature (nwep_bls_sign).
func BLSSign(sk [32]byte, msg []byte) (sig [96]byte, rc int) {
	rc = int(C.nwep_bls_sign(arrPtr(&sig), arrPtr(&sk), bytePtr(msg), C.size_t(len(msg))))
	return
}

// BLSVerify verifies a bls signature over msg under pk (nwep_bls_verify).
func BLSVerify(sig [96]byte, pk [48]byte, msg []byte) int {
	return int(C.nwep_bls_verify(arrPtr(&sig), arrPtr(&pk), bytePtr(msg), C.size_t(len(msg))))
}

// BLSAggregate aggregates n concatenated signatures into one (nwep_bls_aggregate).
func BLSAggregate(sigs []byte, n int) (sig [96]byte, rc int) {
	rc = int(C.nwep_bls_aggregate(arrPtr(&sig), bytePtr(sigs), C.size_t(n)))
	return
}

// BLSVerifyAggregate verifies an aggregate signature over msg under n public keys (nwep_bls_verify_aggregate).
func BLSVerifyAggregate(aggSig [96]byte, pks []byte, n int, msg []byte) int {
	return int(C.nwep_bls_verify_aggregate(arrPtr(&aggSig), bytePtr(pks), C.size_t(n), bytePtr(msg), C.size_t(len(msg))))
}

// CheckpointDecode decodes a checkpoint into an owned handle (nwep_checkpoint_decode).
func CheckpointDecode(bytes []byte) (unsafe.Pointer, int) {
	var out *C.nwep_checkpoint
	rc := int(C.nwep_checkpoint_decode(bytePtr(bytes), C.size_t(len(bytes)), &out))
	return unsafe.Pointer(out), rc
}

// CheckpointFree frees a decoded checkpoint (nwep_checkpoint_free).
func CheckpointFree(cp unsafe.Pointer) {
	C.nwep_checkpoint_free((*C.nwep_checkpoint)(cp))
}

// CheckpointStaleness returns how stale a checkpoint is at now_secs (nwep_checkpoint_staleness).
func CheckpointStaleness(cp unsafe.Pointer, nowSecs int64) int {
	return int(C.nwep_checkpoint_staleness((*C.nwep_checkpoint)(cp), C.int64_t(nowSecs)))
}

// GenesisCheckpointCreate builds the founding checkpoint for a fresh network (nwep_genesis_checkpoint_create).
//
// blsSecrets is nFounders concatenated 32-byte secrets, blsPubkeys nFounders
// concatenated 48-byte keys, indices one byte per founder.
func GenesisCheckpointCreate(blsSecrets, blsPubkeys, indices []byte, nFounders, threshold int) ([]byte, int) {
	var outlen C.size_t
	probe := func(out *C.uint8_t) C.int {
		return C.nwep_genesis_checkpoint_create(bytePtr(blsSecrets), bytePtr(blsPubkeys), bytePtr(indices), C.size_t(nFounders), C.size_t(threshold), out, &outlen)
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
