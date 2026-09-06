package bootstrap

import "testing"

func TestSuperuserIdentity(t *testing.T) {
	t.Parallel()

	username, email := superuserIdentity("admin@example.com")
	if username != "admin" || email != "admin@example.com" {
		t.Fatalf("email form: username=%q email=%q", username, email)
	}

	username, email = superuserIdentity("superadmin")
	if username != "superadmin" || email != "superadmin@hermes.local" {
		t.Fatalf("username form: username=%q email=%q", username, email)
	}

	username, email = superuserIdentity("ab@example.com")
	if username != "admin" || email != "ab@example.com" {
		t.Fatalf("short local-part: username=%q email=%q", username, email)
	}
}
