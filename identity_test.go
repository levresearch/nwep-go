package nwep

import (
	"bytes"
	"errors"
	"testing"
)

// TestIdentityVertical exercises the L1 identity path end to end.
func TestIdentityVertical(t *testing.T) {
	id, err := GenerateIdentity()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	defer id.Close()

	// the node_id verifies against the identity's own public key.
	if !id.NodeID().Verify(id.PublicKey()) {
		t.Fatal("node_id does not verify against its own pubkey")
	}

	// base58 round-trips back to the same node_id.
	back, err := NodeIDFromBase58(id.NodeID().Base58())
	if err != nil || back != id.NodeID() {
		t.Fatalf("base58 round trip: %v", err)
	}

	// from_pubkey reproduces the node_id.
	derived, err := NodeIDFromPubkey(id.PublicKey())
	if err != nil || derived != id.NodeID() {
		t.Fatalf("from_pubkey: %v", err)
	}
}

// TestIdentityPEM round-trips an identity through pem.
func TestIdentityPEM(t *testing.T) {
	id, err := GenerateIdentity()
	if err != nil {
		t.Fatal(err)
	}
	defer id.Close()

	pem, err := id.SavePEM()
	if err != nil {
		t.Fatalf("save pem: %v", err)
	}
	loaded, err := LoadIdentityPEM(pem)
	if err != nil {
		t.Fatalf("load pem: %v", err)
	}
	defer loaded.Close()
	if loaded.NodeID() != id.NodeID() {
		t.Fatal("loaded identity has a different node_id")
	}
}

// TestSignVerify signs with an identity and verifies with the free function.
func TestSignVerify(t *testing.T) {
	id, err := GenerateIdentity()
	if err != nil {
		t.Fatal(err)
	}
	defer id.Close()

	msg := []byte("web/1 known answer")
	sig, err := id.Sign(msg)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if !Verify(sig, msg, id.PublicKey()) {
		t.Fatal("signature did not verify")
	}
	if Verify(sig, []byte("tampered"), id.PublicKey()) {
		t.Fatal("verify accepted a tampered message")
	}
}

// TestErrorTaxonomy checks the shared error type classifies and matches.
func TestErrorTaxonomy(t *testing.T) {
	_, err := NodeIDFromBase58("not valid base58 !!!")
	if err == nil {
		t.Fatal("expected an error from bad base58")
	}
	var e *Error
	if !errors.As(err, &e) {
		t.Fatalf("error is not *Error: %T", err)
	}
	if e.Family() != FamilyProtocol {
		t.Errorf("expected protocol family, got %d", e.Family())
	}
}

// TestShamir splits and recombines a secret.
func TestShamir(t *testing.T) {
	secret := []byte("a 32 byte recovery key, exactly!")
	shares, err := SplitSecret(secret, 2, 3)
	if err != nil {
		t.Fatalf("split: %v", err)
	}
	shareLen := len(shares) / 3
	// any two of the three shares recombine the secret.
	got, err := CombineShares(shares[:2*shareLen], 2, shareLen)
	if err != nil {
		t.Fatalf("combine: %v", err)
	}
	if !bytes.Equal(got, secret) {
		t.Fatalf("recovered %q want %q", got, secret)
	}
}
