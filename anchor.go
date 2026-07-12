// the anchor node, an epoch checkpoint co-signer in the threshold set NW120515.
//
// an anchor collects the log root each epoch, produces its bls partial signature,
// and the partials are aggregated into a checkpoint. dispatch answers a peer's
// partial-signature request from inside a server handler.

package nwep

import (
	"unsafe"

	"github.com/levresearch/nwep-go/sys"
)

// AnchorNode is one member of the threshold anchor set co-signing checkpoints NW120500.
type AnchorNode struct {
	ptr unsafe.Pointer
}

// NewAnchorNode allocates an anchor node from its ed25519 keys and bls share (nwep_anchor_node_create).
func NewAnchorNode(pubkey, privkey, blsSecret [32]byte, blsPubkey [48]byte, shareIndex, collectionWindowMs uint64) *AnchorNode {
	return &AnchorNode{ptr: sys.AnchorNodeCreate(pubkey, privkey, blsSecret, blsPubkey, shareIndex, collectionWindowMs)}
}

// CollectLogRoot records a server's reported log root for an epoch (nwep_anchor_node_collect_log_root).
func (a *AnchorNode) CollectLogRoot(epoch uint64, serverRoot [32]byte, serverLogSize uint64, localRoot [32]byte) error {
	return check(sys.AnchorNodeCollectLogRoot(a.ptr, epoch, serverRoot, serverLogSize, localRoot))
}

// Dispatch answers a peer's partial-signature request into res, from a handler (nwep_anchor_node_dispatch).
//
// anchorIDs is nAnchors concatenated 32-byte requester anchor ids.
func (a *AnchorNode) Dispatch(requesterNodeID [32]byte, anchorIDs []byte, nAnchors int, req *Message, res *Responder, nowSecs int64) error {
	return check(sys.AnchorNodeDispatch(a.ptr, requesterNodeID, anchorIDs, nAnchors, req.ptr, res.buf, nowSecs))
}

// ProducePartialSig produces this anchor's partial signature for a root (nwep_anchor_node_produce_partial_sig).
//
// returns the anchor's share index and its 96-byte partial signature.
func (a *AnchorNode) ProducePartialSig(epoch uint64, merkleRoot [32]byte, logSize uint64) (index uint8, sig [96]byte, err error) {
	index, sig, rc := sys.AnchorNodeProducePartialSig(a.ptr, epoch, merkleRoot, logSize)
	return index, sig, check(rc)
}

// Close frees the anchor node (nwep_anchor_node_free).
func (a *AnchorNode) Close() {
	if a.ptr != nil {
		sys.AnchorNodeFree(a.ptr)
		a.ptr = nil
	}
}

// Raw returns the underlying sys anchor node pointer, the no-cliffs escape NWG0200.
func (a *AnchorNode) Raw() unsafe.Pointer { return a.ptr }

// RequestPartialSig asks a peer anchor for its partial signature over a client (nwep_anchor_request_partial_sig).
//
// returns the peer's share index and its 96-byte partial signature.
func RequestPartialSig(client *Client, epoch uint64, merkleRoot [32]byte, logSize uint64, peerBLSPubkey [48]byte) (index uint8, sig [96]byte, err error) {
	index, sig, rc := sys.AnchorRequestPartialSig(client.ptr, epoch, merkleRoot, logSize, peerBLSPubkey)
	return index, sig, check(rc)
}

// FinishCheckpoint aggregates partial signatures into a final checkpoint (nwep_anchor_finish_checkpoint).
//
// indices is one byte per partial, sigs nPartials concatenated 96-byte signatures,
// anchorBLSPks nAnchors concatenated 48-byte keys. returns the encoded checkpoint.
func FinishCheckpoint(epoch uint64, merkleRoot [32]byte, logSize uint64, indices, sigs []byte, nPartials int, anchorBLSPks []byte, nAnchors int) ([]byte, error) {
	b, rc := sys.AnchorFinishCheckpoint(epoch, merkleRoot, logSize, indices, sigs, nPartials, anchorBLSPks, nAnchors)
	if err := check(rc); err != nil {
		return nil, err
	}
	return b, nil
}
