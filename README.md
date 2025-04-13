# 🦅 Bird2 SNMP Agent

[![Go Report Card](https://goreportcard.com/badge/github.com/subuk/bird2snmp)](https://goreportcard.com/report/github.com/subuk/bird2snmp)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8.svg)](https://golang.org/dl/)

A lightweight SNMP AgentX implementation for monitoring BIRD routing daemon through SNMP. This agent provides real-time monitoring of BGP peers and their states using the standard BGP4-MIB.

## ✨ Features

- 🚀 Real-time BGP peer monitoring
- 📊 SNMP AgentX protocol support
- 🔄 Automatic data refresh
- 🛠️ IPv4 BGP peer support
- 📈 Standard BGP4-MIB compliance

### Supported OIDs

| OID | Description |
|-----|-------------|
| bgpLocalAs | Local Autonomous System number |
| bgpPeerState | Current state of BGP peer |
| bgpPeerRemoteAddr | Remote peer IP address |
| bgpPeerFsmEstablishedTime | Time since BGP session establishment |
| bgpIdentifier | BGP router identifier |

## 🚀 Installation

### Prerequisites

- BIRD2 configured with BGP peers
- net-snmpd installed and configured

### Quick Start

1. Install net-snmpd:
```bash
opkg install snmpd snmp-mibs
```

2. Configure net-snmpd:
```bash
cat > /opt/etc/snmp/snmpd.conf <<EOF
rocommunity public
master agentx
EOF
```

3. Start snmpd:
```bash
/opt/etc/init.d/S47snmpd start
```

4. Install bird2snmp:
```bash
wget -O /opt/usr/bin/bird2snmp https://github.com/subuk/bird2snmp/releases/download/v0.1.0/bird2snmp.linux.mipsle
chmod +x /opt/usr/bin/bird2snmp
```

5. Create init script:
```bash
cat > /opt/etc/init.d/S81bird2snmp <<EOF
#!/bin/sh
ENABLED=yes
PROCS=bird2snmp
ARGS="--bird-sock=/opt/var/run/bird.ctl"
PREARGS=""
DESC=$PROCS
PATH=/opt/sbin:/opt/bin:/opt/usr/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

export TZ=/opt/etc/localtime

. /opt/etc/init.d/rc.func
EOF
chmod +x /opt/etc/init.d/S81bird2snmp
/opt/etc/init.d/S81bird2snmp start
```

## 🔍 Verification

Test the installation with snmpwalk:
```bash
snmpwalk -m ALL -c public -v2c myhost.example.com bgp
```

Expected output:
```
BGP4-MIB::bgpVersion = STRING: "4"
BGP4-MIB::bgpLocalAs.0 = INTEGER: 64842
BGP4-MIB::bgpPeerState.169.254.153.78 = INTEGER: established(6)
...
```

## 🛠️ Building

### Build for all platforms
```bash
make release-binaries
```

### Build for current platform
```bash
make
```

## ⚙️ Configuration

### Command Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `-s, --bird-sock` | BIRD socket path | `/run/bird/bird.ctl` |
| `-r, --bird-refresh-interval` | Data refresh interval | `3s` |
| `-x, --snmp-master-sock` | SNMP master socket path | `/var/agentx/master` |
| `-p, --snmp-priority` | SNMP registration priority | `127` |

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
