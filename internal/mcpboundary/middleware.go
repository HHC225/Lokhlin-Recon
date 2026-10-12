// Package mcpboundary limits anonymous MCP to explicitly configured socket peers.
package mcpboundary

import (
	"errors"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// RouteMiddleware leaves the existing JWT/API-key policy in place by default.
// An enabled internal mode accepts only exact private socket peers and hosts.
func RouteMiddleware(authenticated gin.HandlerFunc) gin.HandlerFunc {
	if os.Getenv("MCP_INTERNAL_NOAUTH") != "1" {
		return authenticated
	}
	policy, err := loadPolicy(os.Getenv("MCP_INTERNAL_ALLOWED_IPS"), os.Getenv("MCP_INTERNAL_ALLOWED_HOSTS"))
	return func(c *gin.Context) {
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"message": "Internal MCP access policy is not configured"})
			return
		}
		if !policy.allows(c.Request) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Internal MCP access denied"})
			return
		}
		c.Next()
	}
}

type internalPolicy struct {
	peers map[netip.Addr]bool
	hosts map[string]bool
}

func loadPolicy(rawPeers, rawHosts string) (internalPolicy, error) {
	policy := internalPolicy{peers: map[netip.Addr]bool{}, hosts: map[string]bool{}}
	if strings.TrimSpace(rawPeers) == "" || strings.TrimSpace(rawHosts) == "" {
		return policy, errors.New("explicit peers and hosts are required")
	}
	for _, item := range strings.Split(rawPeers, ",") {
		item = strings.TrimSpace(item)
		var address netip.Addr
		var err error
		if strings.Contains(item, "/") {
			var prefix netip.Prefix
			prefix, err = netip.ParsePrefix(item)
			if err == nil && prefix.Bits() != prefix.Addr().BitLen() {
				err = errors.New("MCP peers must be individual addresses")
			}
			address = prefix.Addr()
		} else {
			address, err = netip.ParseAddr(item)
		}
		if err != nil || !address.IsValid() || !address.Unmap().IsPrivate() || address.Zone() != "" {
			return policy, errors.New("invalid internal MCP peer")
		}
		policy.peers[address.Unmap()] = true
	}
	for _, item := range strings.Split(rawHosts, ",") {
		host, valid := exactHost(strings.TrimSpace(item))
		if !valid {
			return policy, errors.New("invalid internal MCP host")
		}
		policy.hosts[host] = true
	}
	return policy, nil
}

func exactHost(raw string) (string, bool) {
	host, port, err := net.SplitHostPort(raw)
	number, portErr := strconv.Atoi(port)
	if err != nil || portErr != nil || number < 1 || number > 65535 || port != strconv.Itoa(number) || host == "" {
		return "", false
	}
	if address, err := netip.ParseAddr(host); err == nil {
		if address.Zone() != "" {
			return "", false
		}
	} else {
		for _, char := range host {
			if !(char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || char == '-' || char == '.') {
				return "", false
			}
		}
		if strings.HasPrefix(host, ".") || strings.HasSuffix(host, ".") || strings.Contains(host, "..") {
			return "", false
		}
	}
	return strings.ToLower(net.JoinHostPort(host, port)), true
}

func (policy internalPolicy) allows(r *http.Request) bool {
	// Forwarded, X-Forwarded-For, X-Real-IP and Gin ClientIP are deliberately
	// excluded: they do not identify the socket's actual container peer.
	peer, port, err := net.SplitHostPort(r.RemoteAddr)
	address, addressErr := netip.ParseAddr(peer)
	portNumber, portErr := strconv.Atoi(port)
	if err != nil || addressErr != nil || portErr != nil || portNumber < 1 || portNumber > 65535 || !policy.peers[address.Unmap()] {
		return false
	}
	host, valid := exactHost(r.Host)
	if !valid || !policy.hosts[host] {
		return false
	}
	origins := r.Header.Values("Origin")
	if len(origins) == 0 {
		return true
	}
	if len(origins) != 1 {
		return false
	}
	origin, err := url.Parse(origins[0])
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if err != nil || origin.Scheme != scheme || origin.User != nil || origin.Opaque != "" || origin.Path != "" || origin.RawPath != "" || origin.RawQuery != "" || origin.ForceQuery || origin.Fragment != "" || origin.RawFragment != "" || origins[0] != scheme+"://"+origin.Host {
		return false
	}
	originHost, valid := exactHost(origin.Host)
	return valid && originHost == host && policy.hosts[originHost]
}
