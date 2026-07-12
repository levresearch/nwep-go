// the layer 0 identity calls, keypairs, node_ids, and ed25519 signing NW040200 NW090500.

package sys

/*
#include <stdlib.h>
#include <nwep.h>
*/
import "C"

import "unsafe"

// keypairSize is the byte size of an nwep_keypair, for the c allocations below.
var keypairSize = C.size_t(unsafe.Sizeof(C.nwep_keypair{}))

// nodeIDToC reinterprets a go node_id array as the c struct (identical layout).
func nodeIDToC(nodeID [NodeIDSize]byte) C.nwep_node_id {
	return *(*C.nwep_node_id)(unsafe.Pointer(&nodeID))
}

// IdentityGenerate generates a fresh ed25519 identity (nwep_identity_generate).
//
// the keypair is allocated in c memory and returned as an opaque pointer so the
// private half never enters the go heap, free it with KeypairFree. returns the
// derived node_id, the keypair pointer, and the raw c return code.
func IdentityGenerate() (nodeID [NodeIDSize]byte, keypair unsafe.Pointer, rc int) {
	kp := C.calloc(1, keypairSize)
	var id C.nwep_node_id
	rc = int(C.nwep_identity_generate(&id, (*C.nwep_keypair)(kp)))
	if rc != 0 {
		C.free(kp)
		return nodeID, nil, rc
	}
	nodeID = *(*[NodeIDSize]byte)(unsafe.Pointer(&id.bytes[0]))
	return nodeID, kp, rc
}

// KeypairFree zeroizes the private half then frees a keypair (from IdentityGenerate or KeypairLoadPem).
func KeypairFree(keypair unsafe.Pointer) {
	if keypair == nil {
		return
	}
	C.nwep_zeroize(keypair, keypairSize)
	C.free(keypair)
}

// KeypairPubKey copies out the 32-byte public half of a keypair (no secret material).
func KeypairPubKey(keypair unsafe.Pointer) [PubKeySize]byte {
	kp := (*C.nwep_keypair)(keypair)
	return *(*[PubKeySize]byte)(unsafe.Pointer(&kp.pub_[0]))
}

// KeypairPrivKey copies out the 32-byte private half (secret material, wipe after use).
//
// this is the raw escape hatch, the safe layer keeps the private key in c memory
// and never calls this. zero the returned array when done.
func KeypairPrivKey(keypair unsafe.Pointer) [PrivKeySize]byte {
	kp := (*C.nwep_keypair)(keypair)
	return *(*[PrivKeySize]byte)(unsafe.Pointer(&kp.priv_[0]))
}

// KeypairSavePem encodes a keypair as unencrypted pkcs8 pem (nwep_keypair_save_pem).
//
// returns the pem bytes (secret material, wipe after use) and the c return code.
func KeypairSavePem(keypair unsafe.Pointer) ([]byte, int) {
	// an ed25519 pkcs8 pem is well under 512 bytes, so one ample buffer suffices.
	buf := make([]byte, 512)
	outlen := C.size_t(len(buf))
	rc := int(C.nwep_keypair_save_pem((*C.uint8_t)(unsafe.Pointer(&buf[0])), &outlen, (*C.nwep_keypair)(keypair)))
	if rc != 0 {
		return nil, rc
	}
	return buf[:outlen], rc
}

// KeypairLoadPem decodes pem bytes into a keypair (nwep_keypair_load_pem).
//
// the keypair is allocated in c memory like IdentityGenerate, free it with
// KeypairFree. returns the keypair pointer and the c return code.
func KeypairLoadPem(pem []byte) (keypair unsafe.Pointer, rc int) {
	kp := C.calloc(1, keypairSize)
	var p *C.uint8_t
	if len(pem) > 0 {
		p = (*C.uint8_t)(unsafe.Pointer(&pem[0]))
	}
	rc = int(C.nwep_keypair_load_pem((*C.nwep_keypair)(kp), p, C.size_t(len(pem))))
	if rc != 0 {
		C.free(kp)
		return nil, rc
	}
	return kp, rc
}

// NodeidVerify checks node_id is sha-256(pubkey + "WEB/1"), constant time (nwep_nodeid_verify).
func NodeidVerify(nodeID [NodeIDSize]byte, pubkey [PubKeySize]byte) int {
	id := nodeIDToC(nodeID)
	return int(C.nwep_nodeid_verify(&id, (*C.uint8_t)(unsafe.Pointer(&pubkey[0]))))
}

// NodeidFromPubkey derives the node_id of an ed25519 public key (nwep_nodeid_from_pubkey).
func NodeidFromPubkey(pubkey [PubKeySize]byte) (nodeID [NodeIDSize]byte, rc int) {
	var id C.nwep_node_id
	rc = int(C.nwep_nodeid_from_pubkey(&id, (*C.uint8_t)(unsafe.Pointer(&pubkey[0]))))
	if rc == 0 {
		nodeID = *(*[NodeIDSize]byte)(unsafe.Pointer(&id.bytes[0]))
	}
	return
}

// NodeidToBase58 encodes a node_id as a base58 string (nwep_nodeid_to_base58).
func NodeidToBase58(nodeID [NodeIDSize]byte) (string, int) {
	id := nodeIDToC(nodeID)
	// a 32-byte id is at most 44 base58 chars, so 64 bytes is always enough.
	buf := make([]byte, 64)
	outlen := C.size_t(len(buf))
	rc := int(C.nwep_nodeid_to_base58((*C.char)(unsafe.Pointer(&buf[0])), &outlen, &id))
	if rc != 0 {
		return "", rc
	}
	return string(buf[:outlen]), rc
}

// NodeidFromBase58 decodes a base58 string into a node_id (nwep_nodeid_from_base58).
func NodeidFromBase58(text string) (nodeID [NodeIDSize]byte, rc int) {
	cs := C.CString(text)
	defer C.free(unsafe.Pointer(cs))
	var id C.nwep_node_id
	rc = int(C.nwep_nodeid_from_base58(&id, cs, C.size_t(len(text))))
	if rc == 0 {
		nodeID = *(*[NodeIDSize]byte)(unsafe.Pointer(&id.bytes[0]))
	}
	return
}

// Ed25519Sign signs msg with a private key, returning the 64-byte signature (nwep_ed25519_sign).
func Ed25519Sign(msg []byte, privkey [PrivKeySize]byte) (sig [64]byte, rc int) {
	var mp *C.uint8_t
	if len(msg) > 0 {
		mp = (*C.uint8_t)(unsafe.Pointer(&msg[0]))
	}
	rc = int(C.nwep_ed25519_sign(
		(*C.uint8_t)(unsafe.Pointer(&sig[0])),
		mp, C.size_t(len(msg)),
		(*C.uint8_t)(unsafe.Pointer(&privkey[0])),
	))
	return
}

// Ed25519Verify verifies a 64-byte signature over msg under pubkey (nwep_ed25519_verify).
func Ed25519Verify(sig [64]byte, msg []byte, pubkey [PubKeySize]byte) int {
	var mp *C.uint8_t
	if len(msg) > 0 {
		mp = (*C.uint8_t)(unsafe.Pointer(&msg[0]))
	}
	return int(C.nwep_ed25519_verify(
		(*C.uint8_t)(unsafe.Pointer(&sig[0])),
		mp, C.size_t(len(msg)),
		(*C.uint8_t)(unsafe.Pointer(&pubkey[0])),
	))
}
