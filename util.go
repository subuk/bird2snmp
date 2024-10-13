package main

import (
	"net"

	"github.com/posteo/go-agentx/value"
)

func ipToOid(ip net.IP) []uint32 {
	ret := []uint32{}
	for _, x := range ip.To4() {
		ret = append(ret, uint32(x))
	}
	return ret
}

// Compare returns an integer comparing two SNMP OIDs lexicographically.
// The result will be :
//
//	 0 if x == y
//	-1 if x < y
//	+1 if x > y
func compareOids(x value.OID, y value.OID) int {
	var i int
	for i = 0; i < min(len(x), len(y)); i++ {
		if x[i] < y[i] {
			return -1
		}
		if x[i] > y[i] {
			return 1
		}
	}
	if len(x) < len(y) {
		return -1
	}
	if len(x) > len(y) {
		return 1
	}
	return 0
}
