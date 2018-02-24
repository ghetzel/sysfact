#!/usr/bin/env python2.7
import os
import sys
import re
from subprocess import check_output, CalledProcessError

RX_IFCONFIG_OPTIONS = re.compile('^options=\d+<([^>]+)>')
RX_IFCONFIG_ETHER = re.compile('^ether ((?:[0-9a-fA-F]+:){5}[0-9a-fA-F]+)')
RX_IFCONFIG_INET4 = re.compile('^inet ([^\s]+) netmask ([^\s]+) broadcast (.*)')
RX_IFCONFIG_INET6 = re.compile('^inet6 ([^\s]+) prefixlen (\d+) scopeid 0x(\d+)')
RX_IFCONFIG_STATUS = re.compile('^status: (.*)')
RX_NETSTAT_PHY = re.compile('<([^>]+)>')


def cidr_to_mask(cidr):
    bits = (
        ("1"*int(cidr)) + ("0"*(32-int(cidr)))
    )

    return '.'.join([
        str(int(bits[0:7], 2)),
        str(int(bits[8:15], 2)),
        str(int(bits[16:23], 2)),
        str(int(bits[24:31], 2)),
    ])

try:
    interface_names = []

    # get all interface names
    interface_names = list(set(re.split('\s+', check_output([
        'ifconfig', '-l',
    ]).strip())))

    global_addresses = []

    # for each interface...
    for i, interface in enumerate(interface_names):
        addresses = []
        info = {}

        # get interface details
        for line in check_output(['ifconfig', interface]).split("\n"):
            line = line.strip()

            # -------------------------------------------------------------------------------------
            m = re.match(
                '^{0}:\s+flags=\d+<([^>]+)> metric (\d+) mtu (\d+)$'.format(interface),
                line
            )
            if m is not None:
                flags = [v.lower() for v in str(m.group(1)).split(',')]
                metric = int(m.group(2))
                mtu = int(m.group(3))

                print("network.interfaces.{0}.name:str:{1}".format(i, interface))
                print("network.interfaces.{0}.metric:int:{1}".format(i, metric))
                print("network.interfaces.{0}.mtu:int:{1}".format(i, mtu))

                for fi, flag in enumerate(flags):
                    print("network.interfaces.{0}.flags.{1}:str:{2}".format(i, fi, flag))

                continue

            # -------------------------------------------------------------------------------------
            m = RX_IFCONFIG_OPTIONS.match(line)
            if m is not None:
                options = [v.lower() for v in str(m.group(1)).split(',')]

                for oi, option in enumerate(options):
                    print("network.interfaces.{0}.options.{1}:str:{2}".format(i, oi, option))

                continue

            # -------------------------------------------------------------------------------------
            m = RX_IFCONFIG_ETHER.match(line)
            if m is not None:
                print("network.interfaces.{0}.mac:int:{1}".format(
                    i,
                    str(m.group(1)).lower()
                ))
                continue

            # -------------------------------------------------------------------------------------
            m = RX_IFCONFIG_INET4.match(line)
            if m is not None:
                ip = str(m.group(1))
                netmask = str(m.group(2))
                broadcast = str(m.group(3))
                cidr = (netmask.lower().count('f') * 4)
                netmask = cidr_to_mask(cidr)

                ai = len(addresses)

                print("network.interfaces.{0}.addresses.{1}.ip:str:{2}".format(i, ai, ip))
                print("network.interfaces.{0}.addresses.{1}.netmask:str:{2}".format(
                    i, ai, netmask
                ))
                print("network.interfaces.{0}.addresses.{1}.cidr:int:{2}".format(i, ai, cidr))
                print("network.interfaces.{0}.addresses.{1}.broadcast:str:{2}".format(
                    i, ai, broadcast
                ))
                print("network.interfaces.{0}.addresses.{1}.family:str:inet4".format(i, ai))

                addresses.append(ip)
                global_addresses.append(ip)
                continue

            # -------------------------------------------------------------------------------------
            m = RX_IFCONFIG_INET6.match(line)
            if m is not None:
                ip = str(m.group(1))
                cidr = str(m.group(2))
                scopeid = str(m.group(3))

                ai = len(addresses)

                print("network.interfaces.{0}.addresses.{1}.ip:str:{2}".format(i, ai, ip))
                print("network.interfaces.{0}.addresses.{1}.cidr:int:{2}".format(i, ai, cidr))
                print("network.interfaces.{0}.addresses.{1}.scopeid:str:{2}".format(
                    i, ai, scopeid
                ))
                print("network.interfaces.{0}.addresses.{1}.family:str:inet6".format(i, ai))

                addresses.append(ip)
                global_addresses.append(ip)
                continue

            # -------------------------------------------------------------------------------------
            m = RX_IFCONFIG_INET6.match(line)
            if m is not None:
                print("network.interfaces.{0}.status:str:{1}".format(i, m.group(1)))
                continue

            # get interface statistics
            phy = 0
            log = 0

            for line in check_output([
                'netstat', '-i', '-W', '-b', '-n', '-I', interface,
            ]).split("\n"):
                line = line.strip()

                if line.startswith('Name'):
                    continue

                # aggregate details for the physical interface
                m = RX_NETSTAT_PHY.match(line)
                if m is not None:
                    name = str(m.group(1))

                    parts = line.split()

                    if len(parts) < 12:
                        continue

                    ns_interface = parts[0]
                    mtu = parts[1]
                    network = parts[2]
                    mac = parts[3]
                    ipkt = parts[4]
                    ierr = parts[5]
                    idrop = parts[6]
                    ibytes = parts[7]
                    opkt = parts[8]
                    oerr = parts[9]
                    obytes = parts[10]
                    coll = parts[11]

                    print(
                        "network.interfaces.{0}.statistics.physical.{1}.id:str:{2}".format(
                            i, phy, name
                        )
                    )

                    print(
                        "network.interfaces.{0}.statistics.physical.{1}.mtu:int:{2}".format(
                            i, phy, int(mtu)
                        )
                    )

                    print(
                        "network.interfaces.{0}.statistics.physical.{1}.mac:str:{2}".format(
                            i, phy, mac
                        )
                    )

                    print(
                        "network.interfaces.{0}.statistics.physical.{1}.rx.packets:int:{2}".format(
                            i, phy, int(ipkt)
                        )
                    )

                    print(
                        "network.interfaces.{0}.statistics.physical.{1}.rx.errors:int:{2}".format(
                            i, phy, int(ierr)
                        )
                    )

                    print(
                        "network.interfaces.{0}.statistics.physical.{1}.rx.dropped:int:{2}".format(
                            i, phy, int(idrop)
                        )
                    )

                    print(
                        "network.interfaces.{0}.statistics.physical.{1}.rx.bytes:int:{2}".format(
                            i, phy, int(ibytes)
                        )
                    )

                    print(
                        "network.interfaces.{0}.statistics.physical.{1}.tx.packets:int:{2}".format(
                            i, phy, int(opkt)
                        )
                    )

                    print(
                        "network.interfaces.{0}.statistics.physical.{1}.tx.errors:int:{2}".format(
                            i, phy, int(oerr)
                        )
                    )

                    print(
                        "network.interfaces.{0}.statistics.physical.{1}.tx.bytes:int:{2}".format(
                            i, phy, int(obytes)
                        )
                    )

                    print(
                        (
                            "network.interfaces.{0}.statistics.physical."
                            "{1}.tx.collisions:int:{2}"
                        ).format(
                            i, phy, int(coll)
                        )
                    )

                    phy += 1
                else:
                    parts = line.split()

                    if len(parts) < 12:
                        continue

                    ns_interface = parts[0]
                    network = parts[2]
                    ip = parts[3]
                    ipkt = parts[4]
                    ibytes = parts[7]
                    opkt = parts[8]
                    obytes = parts[10]

                    pair = network.split('/', 1)

                    if len(pair) > 1:
                        network = pair[0]
                        cidr = pair[1]
                    else:
                        network = pair[0]
                        cidr = 32

                    print(
                        "network.interfaces.{0}.statistics.logical.{1}.network:str:{2}".format(
                            i, log, network
                        )
                    )

                    print(
                        "network.interfaces.{0}.statistics.logical.{1}.cidr:int:{2}".format(
                            i, log, int(cidr)
                        )
                    )

                    print(
                        "network.interfaces.{0}.statistics.logical.{1}.address:str:{2}".format(
                            i, log, ip
                        )
                    )

                    print(
                        "network.interfaces.{0}.statistics.logical.{1}.rx.packets:int:{2}".format(
                            i, log, int(ipkt)
                        )
                    )

                    print(
                        "network.interfaces.{0}.statistics.logical.{1}.rx.bytes:int:{2}".format(
                            i, log, int(ibytes)
                        )
                    )

                    print(
                        "network.interfaces.{0}.statistics.logical.{1}.tx.packets:int:{2}".format(
                            i, log, int(opkt)
                        )
                    )

                    print(
                        "network.interfaces.{0}.statistics.logical.{1}.tx.bytes:int:{2}".format(
                            i, log, int(obytes)
                        )
                    )

                    log += 1

    for gi, address in enumerate(global_addresses):
        print("network.ip.{0}:str:{1}".format(gi, address))

except CalledProcessError:
    sys.exit(0)
