// This file parses Redis addresses using the GoFrame redis config convention:
// a single host:port is standalone, and comma-separated hosts use Cluster.

package redis

import "strings"

// splitRedisAddrs splits one GoFrame-style address value into Redis nodes.
func splitRedisAddrs(address string) []string {
	parts := strings.Split(address, ",")
	addrs := make([]string, 0, len(parts))
	for _, part := range parts {
		addr := strings.TrimSpace(part)
		if addr == "" {
			continue
		}
		addrs = append(addrs, addr)
	}
	return addrs
}
