package main

import (
	"math/big"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"
)

// $ sudo birdc show status
// BIRD 2.15.1 ready.
// BIRD 2.15.1
// Router ID is 192.168.32.79
// Hostname is infra2
// Current server time is 2024-10-13 14:39:40.531
// Last reboot on 2024-10-12 20:41:10.197
// Last reconfiguration on 2024-10-13 09:25:06.844
// Daemon is up and running
type ShowStatus struct {
	RouterId net.IP
	Hostname string
}

func ParseShowStatus(in string) ShowStatus {
	status := ShowStatus{}
	for _, line := range strings.Split(in, "\n") {
		routerIdPref := "Router ID is "
		if strings.HasPrefix(line, routerIdPref) {
			status.RouterId = net.ParseIP(strings.TrimSpace(line[len(routerIdPref):]))
		}
		hostnamePref := "Hostname is "
		if strings.HasPrefix(line, hostnamePref) {
			status.Hostname = strings.TrimSpace(line[len(hostnamePref):])
		}
	}
	return status
}

// pnz2_gw1   BGP        ---        up     2024-10-12 19:14:52  Established
//
//	BGP state:          Established
//	  Neighbor address: 169.254.153.78
//	  Neighbor AS:      64844
//	  Local AS:         64842
//	  Neighbor ID:      192.168.131.1
//	  Local capabilities
//	    Multiprotocol
//	      AF announced: ipv4
//	    Route refresh
//	    Graceful restart
//	    4-octet AS numbers
//	    Enhanced refresh
//	    Long-lived graceful restart
//	  Neighbor capabilities
//	    Multiprotocol
//	      AF announced: ipv4
//	    Route refresh
//	    Graceful restart
//	    4-octet AS numbers
//	    Enhanced refresh
//	    Long-lived graceful restart
//	  Session:          external AS4
//	  Source address:   169.254.153.77
//	  Hold timer:       186.508/240
//	  Keepalive timer:  67.492/80
//	Channel ipv4
//	  State:          UP
//	  Table:          master4
//	  Preference:     100
//	  Input filter:   (unnamed)
//	  Output filter:  (unnamed)
//	  Routes:         10 imported, 29 exported, 2 preferred
//	  Route change stats:     received   rejected   filtered    ignored   accepted
//	    Import updates:            339          0          0          0        339
//	    Import withdraws:          503          0        ---        174        329
//	    Export updates:           1038        329          0        ---        709
//	    Export withdraws:          258        ---        ---        ---        467
//	  BGP Next hop:   169.254.153.77

type ProtocolBGPChannel struct {
	Name      string
	Imported  int
	Exported  int
	Preferred int
}

type ProtocolBGPStatus struct {
	Name            string
	Table           string
	Up              bool
	Since           time.Time
	State           string
	NeighborAddress net.IP
	LocalAs         int
	Channels        map[string]ProtocolBGPChannel
}

func ParseShowProtocolsAll(in string) []ProtocolBGPStatus {
	lines := strings.Split(in, "\n")
	state := "new"

	protocols := []ProtocolBGPStatus{}
	var proto *ProtocolBGPStatus

	for _, line := range lines {
		if len(line) < 1 {
			continue
		}
		if line[0] != ' ' && line[0] != '\t' {
			state = "new"
		}
		switch state {
		case "new":
			state = "parse_any_proto"
			var items []string
			for _, item := range strings.Split(line, " ") {
				item = strings.TrimSpace(item)
				if item == "" {
					continue
				}
				items = append(items, strings.TrimSpace(item))
			}
			if len(items) < 6 {
				continue
			}
			if items[1] != "BGP" {
				continue
			}
			state = "parse_bgp_proto"
			if proto != nil {
				protocols = append(protocols, *proto)
			}
			proto = &ProtocolBGPStatus{Channels: map[string]ProtocolBGPChannel{}}

			proto.Name = items[0]
			if items[2] != "---" {
				proto.Table = items[2]
			}
			if items[3] == "up" {
				proto.Up = true
			}
			if t, err := time.ParseInLocation(time.DateTime, items[4]+" "+items[5], time.Local); err == nil {
				proto.Since = t
			}
			continue
		case "parse_bgp_proto":
			if strings.HasPrefix(line, "  BGP state:") {
				items := strings.Split(line, ":")
				if len(items) > 1 {
					proto.State = strings.TrimSpace(items[1])
				}
			}
			if strings.HasPrefix(line, "    Neighbor address:") {
				items := strings.Split(line, ":")
				if len(items) > 1 {
					proto.NeighborAddress = net.ParseIP(strings.TrimSpace(items[1]))
				}
			}
			if strings.HasPrefix(line, "    Local AS:") {
				items := strings.Split(line, ":")
				if len(items) > 1 {
					localAs, err := strconv.ParseInt(strings.TrimSpace(items[1]), 10, 32)
					if err == nil {
						proto.LocalAs = int(localAs)
					}
				}
			}

			if strings.HasPrefix(line, "  Channel ") {
				channelName := strings.SplitN(strings.TrimSpace(line), " ", 2)[1]
				state = "parse_bgp_proto_channel_" + channelName
				continue
			}
		case "parse_bgp_proto_channel_ipv4":
			if strings.HasPrefix(line, "    Routes:") {
				routestatsLine := strings.Split(strings.TrimSpace(line), ":")[1]
				routestatsLineParts := strings.Split(routestatsLine, ",")
				channelstat := ProtocolBGPChannel{Name: "ipv4"}
				for _, statpart := range routestatsLineParts {
					statpartItems := strings.SplitN(strings.TrimSpace(statpart), " ", 2)
					if len(statpartItems) != 2 {
						continue
					}
					statValueRaw := statpartItems[0]
					statName := statpartItems[1]
					statValue, err := strconv.Atoi(statValueRaw)
					if err != nil {
						continue
					}
					switch statName {
					default:
						continue
					case "imported":
						channelstat.Imported = statValue
					case "exported":
						channelstat.Exported = statValue
					case "preferred":
						channelstat.Preferred = statValue
					}
				}
				proto.Channels[channelstat.Name] = channelstat
			}
		}
	}
	if proto != nil {
		protocols = append(protocols, *proto)
	}
	sort.Slice(protocols, func(i int, j int) bool {
		ia := big.NewInt(0)
		ia.SetBytes(protocols[i].NeighborAddress.To16())
		ja := big.NewInt(0)
		ja.SetBytes(protocols[j].NeighborAddress.To16())
		return ia.Cmp(ja) == -1
	})
	return protocols
}
