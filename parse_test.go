package main

import (
	"net"
	"reflect"
	"testing"
	"time"
)

var StatusInDefault = `
BIRD 2.15.1 ready.
BIRD 2.15.1
Router ID is 192.168.32.79
Hostname is infra2
Current server time is 2024-10-13 14:39:40.531
Last reboot on 2024-10-12 20:41:10.197
Last reconfiguration on 2024-10-13 09:25:06.844
Daemon is up and running
`

func TestParseShowStatus(t *testing.T) {
	type args struct {
		in string
	}
	tests := []struct {
		name string
		args args
		want ShowStatus
	}{
		{name: "show status", args: args{in: StatusInDefault}, want: ShowStatus{RouterId: net.IP{192, 168, 32, 79}.To16(), Hostname: "infra2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseShowStatus(tt.args.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseShowStatus() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

var showProtocolsAllDefault = `
BIRD 2.15.1 ready.
Name       Proto      Table      State  Since         Info
helpers    Static     master4    up     2024-10-12 20:41:10
  Channel ipv4
    State:          UP
    Table:          master4
    Preference:     200
    Input filter:   ACCEPT
    Output filter:  ACCEPT
    Routes:         0 imported, 0 exported, 0 preferred
    Route change stats:     received   rejected   filtered    ignored   accepted
      Import updates:              0          0          0          0          0
      Import withdraws:            0          0        ---          0          0
      Export updates:              0          0          0        ---          0
      Export withdraws:            0        ---        ---        ---          0

device1    Device     ---        up     2024-10-12 20:41:10

direct1    Direct     ---        up     2024-10-12 20:41:10
  Channel ipv4
    State:          UP
    Table:          master4
    Preference:     240
    Input filter:   ACCEPT
    Output filter:  REJECT
    Routes:         2 imported, 0 exported, 2 preferred
    Route change stats:     received   rejected   filtered    ignored   accepted
      Import updates:              2          0          0          0          2
      Import withdraws:            0          0        ---          0          0
      Export updates:              0          0          0        ---          0
      Export withdraws:            0        ---        ---        ---          0
  Channel ipv6
    State:          UP
    Table:          master6
    Preference:     240
    Input filter:   ACCEPT
    Output filter:  REJECT
    Routes:         2 imported, 0 exported, 2 preferred
    Route change stats:     received   rejected   filtered    ignored   accepted
      Import updates:              2          0          0          0          2
      Import withdraws:            0          0        ---          0          0
      Export updates:              0          0          0        ---          0
      Export withdraws:            0        ---        ---        ---          0

ber1_gw1   BGP        ---        up     2024-10-12 20:41:14  Established
  BGP state:          Established
    Neighbor address: 192.168.32.1
    Neighbor AS:      64846
    Local AS:         64846
    Neighbor ID:      192.168.32.1
    Local capabilities
      Multiprotocol
        AF announced: ipv4
      Route refresh
      Graceful restart
      4-octet AS numbers
      Enhanced refresh
      Long-lived graceful restart
    Neighbor capabilities
      Multiprotocol
        AF announced: ipv4
      Route refresh
      4-octet AS numbers
    Session:          internal multihop AS4
    Source address:   192.168.32.79
    Hold timer:       10.639/15
    Keepalive timer:  0.627/5
    Send hold timer:  24.635/30
  Channel ipv4
    State:          UP
    Table:          master4
    Preference:     100
    Input filter:   (unnamed)
    Output filter:  (unnamed)
    Routes:         21 imported, 0 exported, 21 preferred
    Route change stats:     received   rejected   filtered    ignored   accepted
      Import updates:           1674          0          0         18       1656
      Import withdraws:          459          0        ---          0        459
      Export updates:           1658       1656          2        ---          0
      Export withdraws:          459        ---        ---        ---          0
    BGP Next hop:   192.168.32.79
    IGP IPv4 table: master4

xxx_gw1    BGP        ---        start  2024-10-13 09:25:06  Active        Socket: No route to host
  BGP state:          Active
    Neighbor address: 192.168.32.253
    Neighbor AS:      64846
    Local AS:         64846
    Connect delay:    3.102/5
    Last error:       Socket: No route to host
  Channel ipv4
    State:          DOWN
    Table:          master4
    Preference:     100
    Input filter:   (unnamed)
    Output filter:  (unnamed)
    IGP IPv4 table: master4

`

func mustParseTime(t time.Time, err error) time.Time {
	if err != nil {
		panic(err)
	}
	return t
}

func TestParseShowProtocolsAll(t *testing.T) {
	type args struct {
		in string
	}
	tests := []struct {
		name string
		args args
		want []ProtocolBGPStatus
	}{
		{name: "show protocols all", args: args{in: showProtocolsAllDefault}, want: []ProtocolBGPStatus{
			{
				Name:            "ber1_gw1",
				Table:           "",
				Up:              true,
				State:           "Established",
				Since:           mustParseTime(time.Parse(time.DateTime, "2024-10-12 20:41:14")),
				NeighborAddress: net.IPv4(192, 168, 32, 1),
				LocalAs:         64846,
				Channels: map[string]ProtocolBGPChannel{
					"ipv4": {Name: "ipv4", Imported: 21, Exported: 0, Preferred: 21},
				},
			},
			{
				Name:            "xxx_gw1",
				Table:           "",
				Up:              false,
				State:           "Active",
				Since:           mustParseTime(time.Parse(time.DateTime, "2024-10-13 09:25:06")),
				NeighborAddress: net.IPv4(192, 168, 32, 253),
				LocalAs:         64846,
				Channels:        map[string]ProtocolBGPChannel{},
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseShowProtocolsAll(tt.args.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseShowProtocolsAll() = %v, want %v", got, tt.want)
			}
		})
	}
}
