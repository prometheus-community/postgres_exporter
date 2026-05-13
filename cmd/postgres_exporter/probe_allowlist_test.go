package main

import (
	"context"
	"net"
	"net/url"
	"testing"
)

func TestTargetAllowlist_DefaultAllowsAll(t *testing.T) {
	t.Parallel()

	allowlist, err := newTargetAllowlist("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowlist.allows(context.Background(), "anything", "5432") {
		t.Fatal("expected empty allowlist to allow all targets")
	}
}

func TestTargetAllowlist_HostPortAndValidation(t *testing.T) {
	t.Parallel()

	allowlist, err := newTargetAllowlist("db.internal:5432")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowlist.allows(context.Background(), "db.internal", "5432") {
		t.Fatal("expected host:port rule to match")
	}
	if allowlist.allows(context.Background(), "db.internal", "5433") {
		t.Fatal("expected host:port rule to reject other ports")
	}
	if _, err := newTargetAllowlist("db.internal"); err == nil {
		t.Fatal("expected host without port to be rejected")
	}
}

func TestTargetAllowlist_CIDRMatching(t *testing.T) {
	t.Parallel()

	allowlist := &targetAllowlist{
		cidrs:    []*net.IPNet{mustCIDR(t, "10.0.0.0/8")},
		endpoints: map[string]struct{}{},
		resolve: func(_ context.Context, host string) ([]net.IP, error) {
			if host == "db.internal" {
				return []net.IP{net.ParseIP("10.1.2.3")}, nil
			}
			return []net.IP{net.ParseIP("192.168.1.1")}, nil
		},
	}

	if !allowlist.allows(context.Background(), "db.internal", "5432") {
		t.Fatal("expected CIDR rule to allow matching IP")
	}
	connectHost, ok := allowlist.resolveConnectionHost(context.Background(), "db.internal", "5432")
	if !ok {
		t.Fatal("expected CIDR rule to return allowed connection host")
	}
	if connectHost != "10.1.2.3" {
		t.Fatalf("expected resolved IP to be pinned for connection, got %q", connectHost)
	}
	if allowlist.allows(context.Background(), "other.internal", "5432") {
		t.Fatal("expected CIDR rule to reject non-matching IP")
	}
}

func TestRewriteDSNHost_HostnameToIPUsesHostaddr(t *testing.T) {
	t.Parallel()

	out, err := rewriteDSNHost("postgresql://user:pass@db.internal:5432/postgres?sslmode=verify-full", "10.1.2.3", "5432")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	u, err := url.Parse(out)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if got := u.Hostname(); got != "db.internal" {
		t.Fatalf("unexpected host: got %q, want %q", got, "db.internal")
	}
	if got := u.Query().Get("sslmode"); got != "verify-full" {
		t.Fatalf("unexpected sslmode: got %q, want %q", got, "verify-full")
	}
	if got := u.Query().Get("hostaddr"); got != "10.1.2.3" {
		t.Fatalf("unexpected hostaddr: got %q, want %q", got, "10.1.2.3")
	}
}

func TestRewriteDSNHost_IPTargetDoesNotSetHostaddr(t *testing.T) {
	t.Parallel()

	out, err := rewriteDSNHost("postgresql://user:pass@10.1.1.10:5432/postgres?sslmode=disable", "10.1.2.3", "5432")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	u, err := url.Parse(out)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if got := u.Hostname(); got != "10.1.2.3" {
		t.Fatalf("unexpected host: got %q, want %q", got, "10.1.2.3")
	}
	if got := u.Query().Get("hostaddr"); got != "" {
		t.Fatalf("unexpected hostaddr: got %q, want empty", got)
	}
}

func mustCIDR(t *testing.T, cidr string) *net.IPNet {
	t.Helper()
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		t.Fatalf("invalid test CIDR %q: %v", cidr, err)
	}
	return ipnet
}
