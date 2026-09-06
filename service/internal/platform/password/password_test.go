package password_test

import (
	"strings"
	"testing"

	"github.com/hermes-platform/go-service/internal/platform/password"
)

func newTestHasher(t *testing.T) *password.Hasher {
	t.Helper()
	return password.NewWithParams(password.TestParams)
}

func TestHashAndVerifyRoundTrip(t *testing.T) {
	h := newTestHasher(t)

	encoded, err := h.Hash("StrongPass123!")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if !strings.HasPrefix(encoded, "$argon2id$v=19$") {
		t.Fatalf("unexpected PHC prefix: %s", encoded)
	}

	ok, err := h.Verify("StrongPass123!", encoded)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !ok {
		t.Fatal("correct password must verify")
	}
}

func TestVerifyRejectsWrongPassword(t *testing.T) {
	h := newTestHasher(t)

	encoded, err := h.Hash("StrongPass123!")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}

	ok, err := h.Verify("WrongPass123!", encoded)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if ok {
		t.Fatal("wrong password must not verify")
	}
}

func TestHashUsesUniqueSalt(t *testing.T) {
	h := newTestHasher(t)

	a, err := h.Hash("same-password")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	b, err := h.Hash("same-password")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if a == b {
		t.Fatal("hashing the same password twice must produce different outputs (unique salt)")
	}

	// 两个哈希都必须能验证通过
	for _, encoded := range []string{a, b} {
		ok, err := h.Verify("same-password", encoded)
		if err != nil || !ok {
			t.Fatalf("each hash must verify its own password: ok=%v err=%v", ok, err)
		}
	}
}

func TestVerifyMalformedHash(t *testing.T) {
	h := newTestHasher(t)

	cases := map[string]string{
		"empty":            "",
		"not a hash":       "plaintext",
		"wrong algorithm":  "$bcrypt$v=19$m=65536,t=1,p=4$c2FsdA$aGFzaA",
		"bad version":      "$argon2id$v=99$m=65536,t=1,p=4$c2FsdA$aGFzaA",
		"missing segments": "$argon2id$v=19$m=65536,t=1,p=4$c2FsdA",
		"bad params":       "$argon2id$v=19$x=1$c2FsdA$aGFzaA",
		"bad base64 salt":  "$argon2id$v=19$m=64,t=1,p=1$!!!!$aGFzaA",
	}

	for name, encoded := range cases {
		t.Run(name, func(t *testing.T) {
			ok, _ := h.Verify("whatever", encoded)
			if ok {
				t.Fatal("malformed hash must never verify")
			}
		})
	}
}

func TestVerifyEmptyPasswordIsRejected(t *testing.T) {
	h := newTestHasher(t)

	encoded, err := h.Hash("StrongPass123!")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}

	ok, err := h.Verify("", encoded)
	if err != nil {
		t.Fatalf("empty password should not be an error, just a mismatch: %v", err)
	}
	if ok {
		t.Fatal("empty password must not verify")
	}
}

func TestHashRejectsEmptyPassword(t *testing.T) {
	h := newTestHasher(t)
	if _, err := h.Hash(""); err == nil {
		t.Fatal("hashing an empty password must fail")
	}
}

func TestDefaultParamsAreStrongEnough(t *testing.T) {
	// OWASP 建议 Argon2id 至少 m=19456 (19MiB) 或 m=65536 (64MiB) 配 t=1
	p := password.DefaultParams
	if p.Memory < 64*1024 {
		t.Errorf("memory too low: %d KiB", p.Memory)
	}
	if p.Time < 1 {
		t.Errorf("time cost too low: %d", p.Time)
	}
	if p.SaltLen < 16 {
		t.Errorf("salt too short: %d bytes", p.SaltLen)
	}
	if p.KeyLen < 32 {
		t.Errorf("key too short: %d bytes", p.KeyLen)
	}
}
