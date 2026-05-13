// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

const defaultPostgresPort = "5432"

type ipResolver func(context.Context, string) ([]net.IP, error)

type targetAllowlist struct {
	cidrs    []*net.IPNet
	endpoints map[string]struct{}
	resolve  ipResolver
}

func newTargetAllowlist(raw string) (*targetAllowlist, error) {
	allowlist := &targetAllowlist{
		endpoints: map[string]struct{}{},
		resolve: func(ctx context.Context, host string) ([]net.IP, error) {
			return net.DefaultResolver.LookupIP(ctx, "ip", host)
		},
	}

	for _, entry := range strings.Split(raw, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if _, cidr, err := net.ParseCIDR(entry); err == nil {
			allowlist.cidrs = append(allowlist.cidrs, cidr)
			continue
		}

		host, port, err := net.SplitHostPort(entry)
		if err != nil {
			return nil, fmt.Errorf("invalid allowlist entry %q: use CIDR or host:port (for example 10.0.0.0/8 or db.internal:5432)", entry)
		}
		if host == "" {
			return nil, fmt.Errorf("invalid allowlist entry %q: empty host", entry)
		}
		if !isValidPort(port) {
			return nil, fmt.Errorf("invalid allowlist entry %q: invalid port", entry)
		}
		allowlist.endpoints[normalizeEndpoint(host, port)] = struct{}{}
	}

	return allowlist, nil
}

func (a *targetAllowlist) allows(ctx context.Context, host, port string) bool {
	_, ok := a.resolveConnectionHost(ctx, host, port)
	return ok
}

func (a *targetAllowlist) resolveConnectionHost(ctx context.Context, host, port string) (string, bool) {
	if a == nil || (len(a.cidrs) == 0 && len(a.endpoints) == 0) {
		return host, true
	}
	if port == "" {
		port = defaultPostgresPort
	}
	if _, ok := a.endpoints[normalizeEndpoint(host, port)]; ok {
		return host, true
	}
	if len(a.cidrs) == 0 {
		return "", false
	}

	ips := []net.IP{}
	if ip := net.ParseIP(host); ip != nil {
		ips = append(ips, ip)
	} else {
		resolved, err := a.resolve(ctx, host)
		if err != nil {
			return "", false
		}
		ips = resolved
	}

	for _, ip := range ips {
		for _, cidr := range a.cidrs {
			if cidr.Contains(ip) {
				return ip.String(), true
			}
		}
	}
	return "", false
}

func targetEndpointFromDSN(dsn string) (string, string, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return "", "", fmt.Errorf("invalid DSN: %w", err)
	}
	host := strings.TrimSpace(u.Hostname())
	if host == "" {
		return "", "", fmt.Errorf("DSN has no host")
	}
	port := strings.TrimSpace(u.Port())
	if port == "" {
		port = defaultPostgresPort
	}
	return host, port, nil
}

func isValidPort(port string) bool {
	p, err := strconv.Atoi(port)
	if err != nil {
		return false
	}
	return p > 0 && p <= 65535
}

func normalizeEndpoint(host, port string) string {
	return strings.ToLower(strings.TrimSpace(host)) + ":" + strings.TrimSpace(port)
}

func rewriteDSNHost(connectionString, host, port string) (string, error) {
	u, err := url.Parse(connectionString)
	if err != nil {
		return "", fmt.Errorf("invalid DSN: %w", err)
	}
	originalHost := strings.TrimSpace(u.Hostname())
	if originalHost == "" {
		return "", fmt.Errorf("DSN has no host")
	}
	if strings.TrimSpace(port) == "" {
		port = strings.TrimSpace(u.Port())
	}

	query := u.Query()
	resolvedIsIP := net.ParseIP(host) != nil
	originalIsIP := net.ParseIP(originalHost) != nil

	if resolvedIsIP && !originalIsIP {
		// Keep hostname for TLS hostname verification (sslmode=verify-full)
		// while pinning the actual connection address to the allowlisted IP.
		query.Set("hostaddr", host)
		u.Host = net.JoinHostPort(originalHost, port)
	} else {
		query.Del("hostaddr")
		u.Host = net.JoinHostPort(host, port)
	}
	u.RawQuery = query.Encode()
	return u.String(), nil
}
