# Bird2 SNMP Agent

Bird monitoring agent for SNMP AgentX protocol. Compatible with net-snmp.

Limited BGP4-MIB and ipv4 only peers:

## Installation on Keenetic router via opkg

This guide implies you already have bird2 configured with bgp peers and running.

Disable builtin snmp (if was configured before) via ndm shell:

    (config)> no service snmp
    Snmp::Manager: Disabled.
    (config)> system configuration save
    Core::System::StartupConfig: Saving (cli).

Install net-snmpd

    opkg install snmpd
    opkg install snmp-mibs

Configure net-snmpd

    > /opt/etc/snmp/snmpd.conf
    echo rocommunity public >> /opt/etc/snmp/snmpd.conf
    echo master agentx >> /opt/etc/snmp/snmpd.conf

Start snmpd

     /opt/etc/init.d/S47snmpd start

Install bird2snmp (select proper version and binary from releases)

    wget -O /opt/usr/bin/bird2snmp https://github.com/subuk/bird2snmp/releases/download/v0.1.0/bird2snmp.linux.mipsle
    chmod +x /opt/usr/bin/bird2snmp
    cat > /opt/etc/init.d/S81bird2snmp <<EOF
    #!/bin/sh
    ENABLED=yes
    PROCS=bird2snmp
    ARGS="--bird-sock=/opt/var/run/bird.ctl"
    PREARGS=""
    DESC=$PROCS
    PATH=/opt/sbin:/opt/bin:/opt/usr/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

    . /opt/etc/init.d/rc.func
    EOF
    chmod +x /opt/etc/init.d/S81bird2snmp
    /opt/etc/init.d/S81bird2snmp start

Check with snmpwalk:

    $ snmpwalk -m ALL -c public -v2c myhost.example.com bgp
    BGP4-MIB::bgpVersion = STRING: "4"
    BGP4-MIB::bgpLocalAs.0 = INTEGER: 64842
    BGP4-MIB::bgpPeerState.169.254.153.78 = INTEGER: established(6)
    BGP4-MIB::bgpPeerState.169.254.153.86 = INTEGER: established(6)
    BGP4-MIB::bgpPeerState.169.254.153.97 = INTEGER: idle(1)
    ...

## How to build

This command will generate set of binaries for all architectures and operating systems

    make release-binaries

To build only for current os/arch:

    make
