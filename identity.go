// the cryptographic identity layer, NodeID and Identity NW040200 NW090500.
//
// NodeID is the 32-byte sha-256(pubkey + "WEB/1") that names a node on the
// network and in the dht, an immutable value cheap to copy. Identity is an
// ed25519 keypair and the NodeID it derives to, what a server or client proves
// ownership of in the handshake NW090000. Identity holds a private key, so it
// keeps it in c memory and wipes it on Close NWG0700.

package nwep

import (
	"fmt"
	"unsafe"

	"nwep/sys"
)

// NodeID is the 32-byte sha-256 identity that names a node NW040200.
//
// it is the public half of an Identity and the key the dht resolves to an
// address. it is an immutable value, String is its base58 form and two are equal
// when their bytes match.
type NodeID [32]byte

// NodeIDFromBytes wraps 32 raw bytes as a NodeID without checking they name a key.
//
// use NodeIDFromPubkey or NodeIDFromBase58 for a checked one. returns an error
// when b is not exactly 32 bytes.
func NodeIDFromBytes(b []byte) (NodeID, error) {
	if len(b) != 32 {
		return NodeID{}, fmt.Errorf("nwep: node_id must be 32 bytes, got %d", len(b))
	}
	var n NodeID
	copy(n[:], b)
	return n, nil
}

// NodeIDFromBase58 parses a base58 node_id string, the inverse of Base58.
//
// returns the decoded NodeID.
// errors with a protocol error when text is not valid base58 of a 32-byte id.
func NodeIDFromBase58(text string) (NodeID, error) {
	id, rc := sys.NodeidFromBase58(text)
	if err := check(rc); err != nil {
		return NodeID{}, err
	}
	return NodeID(id), nil
}

// NodeIDFromPubkey derives the node_id of an ed25519 public key, sha-256(pubkey + "WEB/1").
//
// recovers the name of a key whose raw bytes do not carry it, for example one
// loaded from a pem NW040200.
//
// returns the derived NodeID.
// errors when pubkey is not a valid ed25519 point.
func NodeIDFromPubkey(pubkey [32]byte) (NodeID, error) {
	id, rc := sys.NodeidFromPubkey(pubkey)
	if err := check(rc); err != nil {
		return NodeID{}, err
	}
	return NodeID(id), nil
}

// Base58 encodes this node_id as a base58 string, also what String returns.
func (n NodeID) Base58() string {
	s, _ := sys.NodeidToBase58(n) // a 32-byte id is always encodable
	return s
}

// String returns the base58 form of the node_id.
func (n NodeID) String() string { return n.Base58() }

// Bytes returns a copy of the raw 32 identity bytes.
func (n NodeID) Bytes() []byte {
	b := make([]byte, 32)
	copy(b, n[:])
	return b
}

// Verify reports whether pubkey is the key this node_id was derived from NW040200.
func (n NodeID) Verify(pubkey [32]byte) bool {
	return sys.NodeidVerify(n, pubkey) == 0
}

// Identity is an ed25519 keypair and its derived NodeID, a node's proven identity.
//
// it owns the keypair in c memory so the private key never enters the go heap,
// call Close to wipe and free it NWG0700. pass it to a server or client builder
// to prove ownership in the handshake NW090000.
type Identity struct {
	kp     unsafe.Pointer
	nodeID NodeID
}

// GenerateIdentity creates a fresh ed25519 identity (nwep_identity_generate).
//
// returns the identity, whose Close the caller must defer to wipe the private key.
// errors when the system csprng fails.
func GenerateIdentity() (*Identity, error) {
	id, kp, rc := sys.IdentityGenerate()
	if err := check(rc); err != nil {
		return nil, err
	}
	return &Identity{kp: kp, nodeID: NodeID(id)}, nil
}

// LoadIdentityPEM restores an identity from pkcs8 pem (nwep_keypair_load_pem).
//
// returns the identity, Close it to wipe the private key.
// errors when the pem is malformed or not an ed25519 key.
func LoadIdentityPEM(pem []byte) (*Identity, error) {
	kp, rc := sys.KeypairLoadPem(pem)
	if err := check(rc); err != nil {
		return nil, err
	}
	id, rc := sys.NodeidFromPubkey(sys.KeypairPubKey(kp))
	if err := check(rc); err != nil {
		sys.KeypairFree(kp)
		return nil, err
	}
	return &Identity{kp: kp, nodeID: NodeID(id)}, nil
}

// NodeID returns the node_id this identity derives to.
func (id *Identity) NodeID() NodeID { return id.nodeID }

// PublicKey returns the ed25519 public half of the keypair.
func (id *Identity) PublicKey() [32]byte { return sys.KeypairPubKey(id.kp) }

// SavePEM encodes the keypair as pkcs8 pem (nwep_keypair_save_pem).
//
// returns secret pem bytes, wipe them after writing them somewhere safe.
func (id *Identity) SavePEM() ([]byte, error) {
	b, rc := sys.KeypairSavePem(id.kp)
	if err := check(rc); err != nil {
		return nil, err
	}
	return b, nil
}

// Sign signs msg with this identity's private key (nwep_ed25519_sign).
//
// the private key is copied out of c memory only for the call and the copy is
// wiped immediately after. returns the 64-byte signature.
func (id *Identity) Sign(msg []byte) ([]byte, error) {
	priv := sys.KeypairPrivKey(id.kp)
	sig, rc := sys.Ed25519Sign(msg, priv)
	sys.Zeroize(unsafe.Pointer(&priv[0]), len(priv))
	if err := check(rc); err != nil {
		return nil, err
	}
	out := make([]byte, 64)
	copy(out, sig[:])
	return out, nil
}

// Close wipes the private key and frees the keypair (nwep_zeroize then free).
//
// safe to call more than once. the Identity must not be used afterward.
func (id *Identity) Close() {
	if id.kp != nil {
		sys.KeypairFree(id.kp)
		id.kp = nil
	}
}

// Raw returns the underlying sys keypair pointer, the no-cliffs escape to L0 NWG0200.
//
// valid until Close. use it to call a sys function the safe api does not wrap.
func (id *Identity) Raw() unsafe.Pointer { return id.kp }

// Verify reports whether sig is a valid ed25519 signature over msg by pubkey (nwep_ed25519_verify).
func Verify(sig, msg []byte, pubkey [32]byte) bool {
	if len(sig) != 64 {
		return false
	}
	var s [64]byte
	copy(s[:], sig)
	return sys.Ed25519Verify(s, msg, pubkey) == 0
}
