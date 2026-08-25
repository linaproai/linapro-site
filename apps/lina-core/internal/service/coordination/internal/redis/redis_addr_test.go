// This file verifies GoFrame-compatible Redis address splitting.

package redis

import "testing"

// TestSplitRedisAddrsParsesStandaloneAndClusterEndpoints verifies comma
// separated addresses follow GoFrame redis config splitting.
func TestSplitRedisAddrsParsesStandaloneAndClusterEndpoints(t *testing.T) {
	standalone := splitRedisAddrs("127.0.0.1:6379")
	if len(standalone) != 1 || standalone[0] != "127.0.0.1:6379" {
		t.Fatalf("expected one standalone address, got %#v", standalone)
	}

	clusterAddrs := splitRedisAddrs("127.0.0.1:6379, 127.0.0.1:6370")
	if len(clusterAddrs) != 2 || clusterAddrs[0] != "127.0.0.1:6379" || clusterAddrs[1] != "127.0.0.1:6370" {
		t.Fatalf("expected two cluster addresses, got %#v", clusterAddrs)
	}

	if got := splitRedisAddrs(" , , "); len(got) != 0 {
		t.Fatalf("expected empty result for blank addresses, got %#v", got)
	}
}
