// the layer 0 anchor node, an epoch checkpoint co-signer NW120515.

package sys

/*
#include <stdlib.h>
#include <nwep.h>
#include <nwep_trust.h>
*/
import "C"

import "unsafe"

// AnchorNodeCreate allocates an anchor node from its keys and bls share (nwep_anchor_node_create).
func AnchorNodeCreate(pubkey, privkey, blsSecret [32]byte, blsPubkey [48]byte, shareIndex, collectionWindowMs uint64) unsafe.Pointer {
	return unsafe.Pointer(C.nwep_anchor_node_create(arrPtr(&pubkey), arrPtr(&privkey), arrPtr(&blsSecret), arrPtr(&blsPubkey), C.uint64_t(shareIndex), C.uint64_t(collectionWindowMs)))
}

// AnchorNodeFree frees an anchor node (nwep_anchor_node_free).
func AnchorNodeFree(node unsafe.Pointer) {
	C.nwep_anchor_node_free((*C.nwep_anchor_node)(node))
}

// AnchorNodeCollectLogRoot records a server's reported log root for an epoch (nwep_anchor_node_collect_log_root).
func AnchorNodeCollectLogRoot(node unsafe.Pointer, epoch uint64, serverRoot [32]byte, serverLogSize uint64, localRoot [32]byte) int {
	return int(C.nwep_anchor_node_collect_log_root((*C.nwep_anchor_node)(node), C.uint64_t(epoch), arrPtr(&serverRoot), C.uint64_t(serverLogSize), arrPtr(&localRoot)))
}

// AnchorNodeDispatch answers a partial-signature request into buf, from a handler (nwep_anchor_node_dispatch).
func AnchorNodeDispatch(node unsafe.Pointer, requesterNodeID [32]byte, anchorIDs []byte, nAnchors int, req, buf unsafe.Pointer, nowSecs int64) int {
	return int(C.nwep_anchor_node_dispatch((*C.nwep_anchor_node)(node), arrPtr(&requesterNodeID), bytePtr(anchorIDs), C.size_t(nAnchors), (*C.nwep_message)(req), (*C.nwep_buf)(buf), C.int64_t(nowSecs)))
}

// AnchorNodeProducePartialSig produces this anchor's partial signature for a root (nwep_anchor_node_produce_partial_sig).
func AnchorNodeProducePartialSig(node unsafe.Pointer, epoch uint64, merkleRoot [32]byte, logSize uint64) (index uint8, sig [96]byte, rc int) {
	var idx C.uint8_t
	rc = int(C.nwep_anchor_node_produce_partial_sig((*C.nwep_anchor_node)(node), C.uint64_t(epoch), arrPtr(&merkleRoot), C.uint64_t(logSize), &idx, arrPtr(&sig)))
	return uint8(idx), sig, rc
}

// AnchorRequestPartialSig asks a peer anchor for its partial signature over a client (nwep_anchor_request_partial_sig).
func AnchorRequestPartialSig(client unsafe.Pointer, epoch uint64, merkleRoot [32]byte, logSize uint64, peerBLSPubkey [48]byte) (index uint8, sig [96]byte, rc int) {
	var idx C.uint8_t
	rc = int(C.nwep_anchor_request_partial_sig((*C.nwep_client)(client), C.uint64_t(epoch), arrPtr(&merkleRoot), C.uint64_t(logSize), arrPtr(&peerBLSPubkey), &idx, arrPtr(&sig)))
	return uint8(idx), sig, rc
}

// AnchorFinishCheckpoint aggregates partial signatures into a final checkpoint (nwep_anchor_finish_checkpoint).
//
// indices is one byte per partial, sigs nPartials concatenated 96-byte signatures,
// anchorBLSPks nAnchors concatenated 48-byte keys.
func AnchorFinishCheckpoint(epoch uint64, merkleRoot [32]byte, logSize uint64, indices, sigs []byte, nPartials int, anchorBLSPks []byte, nAnchors int) ([]byte, int) {
	var outlen C.size_t
	probe := func(out *C.uint8_t) C.int {
		return C.nwep_anchor_finish_checkpoint(C.uint64_t(epoch), arrPtr(&merkleRoot), C.uint64_t(logSize), bytePtr(indices), bytePtr(sigs), C.size_t(nPartials), bytePtr(anchorBLSPks), C.size_t(nAnchors), out, &outlen)
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
