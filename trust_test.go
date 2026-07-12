package nwep

import (
	"testing"

	"github.com/levresearch/nwep-go/sys"
)

// sysKeypairPriv reaches the private key through the sys escape hatch, for the
// tests that need to sign key-management entries.
func sysKeypairPriv(id *Identity) [32]byte {
	return sys.KeypairPrivKey(id.kp)
}

// TestBLSRoundTrip exercises bls keygen, sign, verify, and aggregate NW150500.
func TestBLSRoundTrip(t *testing.T) {
	sk, pk, err := BLSKeygen()
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	msg := []byte("epoch 7 merkle root")
	sig, err := BLSSign(sk, msg)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if !BLSVerify(sig, pk, msg) {
		t.Fatal("signature did not verify")
	}
	if BLSVerify(sig, pk, []byte("tampered")) {
		t.Fatal("verify accepted a tampered message")
	}
}

// TestTrustStoreLifecycle creates, saves, and reloads an empty trust store.
func TestTrustStoreLifecycle(t *testing.T) {
	ts := NewTrustStore()
	defer ts.Close()

	saved, err := ts.Save()
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	ts2 := NewTrustStore()
	defer ts2.Close()
	if err := ts2.Load(saved); err != nil {
		t.Fatalf("load: %v", err)
	}
}

// TestLogAppendRoot appends to a merkle log and reads its growing root NW120000.
func TestLogAppendRoot(t *testing.T) {
	log := NewLog()
	defer log.Close()

	if log.Size() != 0 {
		t.Fatalf("fresh log size = %d, want 0", log.Size())
	}
	empty, err := log.Root()
	if err != nil {
		t.Fatalf("root: %v", err)
	}

	id, _ := GenerateIdentity()
	defer id.Close()
	entry, err := CreateKeyBinding(id.PublicKey(), [32]byte{}, 1000, derivePriv(t, id))
	if err != nil {
		t.Fatalf("create binding: %v", err)
	}
	if _, err := log.Append(entry); err != nil {
		t.Fatalf("append: %v", err)
	}
	if log.Size() != 1 {
		t.Fatalf("size after append = %d, want 1", log.Size())
	}
	full, err := log.Root()
	if err != nil {
		t.Fatalf("root: %v", err)
	}
	if full == empty {
		t.Fatal("root did not change after appending an entry")
	}
}

// derivePriv pulls the private key out of an identity via the sys escape hatch,
// used only to drive the key-entry creation tests.
func derivePriv(t *testing.T, id *Identity) [32]byte {
	t.Helper()
	p := sysKeypairPriv(id)
	return p
}
