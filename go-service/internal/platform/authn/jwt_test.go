package authn_test

import (
	"testing"
	"time"

	"github.com/hermes-platform/go-service/internal/platform/authn"
	"github.com/hermes-platform/go-service/internal/platform/config"
	"github.com/hermes-platform/go-service/internal/platform/errors"
)

const testSecret = "unit-test-secret-key-not-used-in-production"

func newIssuer(ttl time.Duration) *authn.Issuer {
	return authn.NewIssuer(config.AuthConfig{
		SecretKey:       testSecret,
		Issuer:          "hermes-platform-test",
		AccessTokenTTL:  ttl,
		RefreshTokenTTL: 168 * time.Hour,
	})
}

func TestIssueAndParseAccessToken(t *testing.T) {
	issuer := newIssuer(15 * time.Minute)

	token, expiresIn, err := issuer.IssueAccessToken(42, 3)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if token == "" {
		t.Fatal("token must not be empty")
	}
	if expiresIn != 900 {
		t.Fatalf("expires_in should be 900 seconds, got %d", expiresIn)
	}

	claims, err := issuer.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	userID, err := claims.UserID()
	if err != nil {
		t.Fatalf("user id: %v", err)
	}
	if userID != 42 {
		t.Errorf("user id = %d, want 42", userID)
	}
	if claims.TokenVersion != 3 {
		t.Errorf("token version = %d, want 3", claims.TokenVersion)
	}
}

func TestParseRejectsTokenSignedWithAnotherSecret(t *testing.T) {
	real := newIssuer(15 * time.Minute)
	forged := authn.NewIssuer(config.AuthConfig{
		SecretKey:      "a-completely-different-secret",
		Issuer:         "hermes-platform-test",
		AccessTokenTTL: 15 * time.Minute,
	})

	token, _, err := forged.IssueAccessToken(42, 0)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	if _, err := real.ParseAccessToken(token); !errors.Is(err, errors.KindUnauthenticated) {
		t.Fatalf("expected unauthenticated error, got %v", err)
	}
}

func TestExpiredTokenIsReportedDistinctly(t *testing.T) {
	issuer := newIssuer(-1 * time.Second)

	token, _, err := issuer.IssueAccessToken(42, 0)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	_, err = issuer.ParseAccessToken(token)
	if err == nil {
		t.Fatal("expired token must be rejected")
	}
	if err != authn.ErrTokenExpired {
		t.Fatalf("expected ErrTokenExpired, got %v", err)
	}
}

func TestParseRejectsGarbage(t *testing.T) {
	issuer := newIssuer(15 * time.Minute)

	for name, raw := range map[string]string{
		"empty":     "",
		"garbage":   "not.a.jwt",
		"truncated": "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiI0MiJ9",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := issuer.ParseAccessToken(raw); !errors.Is(err, errors.KindUnauthenticated) {
				t.Fatalf("expected unauthenticated error, got %v", err)
			}
		})
	}
}

func TestTokenTypeIsValidated(t *testing.T) {
	// 刷新令牌与访问令牌使用同一套签名，
	// 必须靠 typ 声明区分，否则刷新令牌可被当作访问令牌使用。
	issuer := newIssuer(15 * time.Minute)

	token, _, err := issuer.IssueAccessToken(42, 0)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	claims, err := issuer.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.TokenType != authn.TokenTypeAccess {
		t.Fatalf("token type = %q, want %q", claims.TokenType, authn.TokenTypeAccess)
	}
}

func TestTokenErrorsMapToExpectedHTTPStatus(t *testing.T) {
	cases := []struct {
		err        error
		wantStatus int
	}{
		{authn.ErrTokenInvalid, 401},
		{authn.ErrTokenExpired, 401},
		{authn.ErrTokenRevoked, 401},
		{authn.ErrRefreshTokenInvalid, 401},
		{authn.ErrLoginLocked, 400},
		{authn.ErrServiceTokenInvalid, 401},
	}

	for _, c := range cases {
		got := errors.KindOf(c.err).HTTPStatus()
		if got != c.wantStatus {
			t.Errorf("%v: status = %d, want %d", c.err, got, c.wantStatus)
		}
	}
}
