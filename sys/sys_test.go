package sys

import (
	"strings"
	"testing"
)

// TestVersion proves cgo links libnwep and reads a value back across the boundary.
func TestVersion(t *testing.T) {
	v := Version()
	if !strings.HasPrefix(v, "0.") {
		t.Fatalf("unexpected version %q", v)
	}
}

// TestIdentityRoundTrip exercises generate, the derived name, base58, and verify.
func TestIdentityRoundTrip(t *testing.T) {
	id, kp, rc := IdentityGenerate()
	if rc != 0 {
		t.Fatalf("generate rc=%d %s", rc, Strerror(rc))
	}
	defer KeypairFree(kp)

	// the node_id must be the derived name of the keypair's public half.
	pub := KeypairPubKey(kp)
	if NodeidVerify(id, pub) != 0 {
		t.Fatal("node_id does not verify against its own pubkey")
	}
	derived, rc := NodeidFromPubkey(pub)
	if rc != 0 || derived != id {
		t.Fatalf("from_pubkey did not reproduce the node_id rc=%d", rc)
	}

	// base58 round-trips back to the same bytes.
	b58, rc := NodeidToBase58(id)
	if rc != 0 {
		t.Fatalf("to_base58 rc=%d", rc)
	}
	back, rc := NodeidFromBase58(b58)
	if rc != 0 || back != id {
		t.Fatalf("base58 round trip failed rc=%d got %x want %x", rc, back, id)
	}
}

// TestPemRoundTrip saves a keypair to pem and loads it back to the same identity.
func TestPemRoundTrip(t *testing.T) {
	id, kp, rc := IdentityGenerate()
	if rc != 0 {
		t.Fatalf("generate rc=%d", rc)
	}
	defer KeypairFree(kp)

	pem, rc := KeypairSavePem(kp)
	if rc != 0 {
		t.Fatalf("save_pem rc=%d %s", rc, Strerror(rc))
	}
	kp2, rc := KeypairLoadPem(pem)
	if rc != 0 {
		t.Fatalf("load_pem rc=%d %s", rc, Strerror(rc))
	}
	defer KeypairFree(kp2)

	loaded, rc := NodeidFromPubkey(KeypairPubKey(kp2))
	if rc != 0 || loaded != id {
		t.Fatalf("loaded keypair has a different identity rc=%d", rc)
	}
}

// TestSignVerify signs a message and verifies it under the matching pubkey.
func TestSignVerify(t *testing.T) {
	_, kp, rc := IdentityGenerate()
	if rc != 0 {
		t.Fatalf("generate rc=%d", rc)
	}
	defer KeypairFree(kp)
	pub := KeypairPubKey(kp)
	priv := KeypairPrivKey(kp)

	msg := []byte("web/1 known answer")
	sig, rc := Ed25519Sign(msg, priv)
	if rc != 0 {
		t.Fatalf("sign rc=%d %s", rc, Strerror(rc))
	}
	if Ed25519Verify(sig, msg, pub) != 0 {
		t.Fatal("signature did not verify under the matching pubkey")
	}
	// a tampered message must fail to verify.
	if Ed25519Verify(sig, []byte("tampered"), pub) == 0 {
		t.Fatal("verify accepted a tampered message")
	}
}
