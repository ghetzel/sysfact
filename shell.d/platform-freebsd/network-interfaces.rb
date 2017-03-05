#!/usr/bin/env ruby

def cidr_to_mask(cidr)
  bits = (("1"*cidr.to_i) + ("0"*(32-cidr.to_i)))

  return [
    bits[0..7].to_i(2),
    bits[8..15].to_i(2),
    bits[16..23].to_i(2),
    bits[24..31].to_i(2),
  ].join('.')
end

interface_names = []

# get all interface names
interface_names = %x{ ifconfig -l }.strip.chomp.split(/\s+/).uniq

global_addresses = []

# for each interface...
interface_names.each_index do |i|
  interface = interface_names[i]
  addresses = []
  info = {}

# get interface details
  IO.popen("ifconfig '#{ interface }'").read.split("\n").each do |line|
    line = line.strip.chomp

    case line
    when /^#{ interface }:\s+flags=\d+<([^>]+)> metric (\d+) mtu (\d+)$/
      flags = $~[1].to_s.split(',').collect{|i| i.downcase }
      metric = $~[2].to_i
      mtu = $~[3].to_i

      puts "network.interfaces.#{ i }.name:str:#{ interface }"
      puts "network.interfaces.#{ i }.metric:int:#{ metric }"
      puts "network.interfaces.#{ i }.mtu:int:#{ mtu }"

      flags.each_index do |fi|
        puts "network.interfaces.#{ i }.flags.#{ fi }:str:#{ flags[fi] }"
      end

    when /^options=\d+<([^>]+)>/
      options = $~[1].to_s.split(',').collect{|i| i.downcase }

      options.each_index do |oi|
        puts "network.interfaces.#{ i }.options.#{ oi }:str:#{ options[oi] }"
      end

    when /^ether ((?:[0-9a-fA-F]+:){5}[0-9a-fA-F]+)/
      puts "network.interfaces.#{ i }.mac:int:#{ $~[1].to_s.downcase }"

    when /^inet ([^\s]+) netmask ([^\s]+) broadcast (.*)/
      ip = $~[1].to_s
      netmask = $~[2].to_s
      broadcast = $~[3].to_s
      cidr = (netmask.downcase.count('f') * 4)
      netmask = cidr_to_mask(cidr)

      ai = addresses.length

      puts "network.interfaces.#{ i }.addresses.#{ ai }.ip:str:#{ ip }"
      puts "network.interfaces.#{ i }.addresses.#{ ai }.netmask:str:#{ netmask }"
      puts "network.interfaces.#{ i }.addresses.#{ ai }.cidr:int:#{ cidr }"
      puts "network.interfaces.#{ i }.addresses.#{ ai }.broadcast:str:#{ broadcast }"
      puts "network.interfaces.#{ i }.addresses.#{ ai }.family:str:inet4"

      addresses << ip
      global_addresses << ip

    when /^inet6 ([^\s]+) prefixlen (\d+) scopeid 0x(\d+)/
      ip = $~[1].to_s
      cidr = $~[2].to_s
      scopeid = $~[3].to_s

      ai = addresses.length

      puts "network.interfaces.#{ i }.addresses.#{ ai }.ip:str:#{ ip }"
      puts "network.interfaces.#{ i }.addresses.#{ ai }.cidr:int:#{ cidr }"
      puts "network.interfaces.#{ i }.addresses.#{ ai }.scopeid:str:#{ scopeid }"
      puts "network.interfaces.#{ i }.addresses.#{ ai }.family:str:inet6"

      addresses << ip
      global_addresses << ip

    when /^status: (.*)/
      puts "network.interfaces.#{ i }.status:str:#{ $~[1] }"
    end

    # get interface statistics
    phy = 0
    log = 0

    IO.popen("netstat -i -W -b -n -I '#{ interface }'").read.split("\n").each do |line|
      line = line.strip.chomp
      next if line =~ /^Name\s+/

      # aggregate details for the physical interface
      case line
      when /<([^>]+)>/
        name = $~[1].to_s

        ns_interface, mtu, network, mac, ipkt, ierr, idrop, ibytes, opkt, oerr, obytes, coll = line.split(/\s+/)

        puts "network.interfaces.#{ i }.statistics.physical.#{ phy }.id:str:#{ name }"
        puts "network.interfaces.#{ i }.statistics.physical.#{ phy }.mtu:int:#{ mtu.to_i }"
        puts "network.interfaces.#{ i }.statistics.physical.#{ phy }.mac:str:#{ mac }"
        puts "network.interfaces.#{ i }.statistics.physical.#{ phy }.rx.packets:int:#{ ipkt.to_i }"
        puts "network.interfaces.#{ i }.statistics.physical.#{ phy }.rx.errors:int:#{ ierr.to_i }"
        puts "network.interfaces.#{ i }.statistics.physical.#{ phy }.rx.dropped:int:#{ idrop.to_i }"
        puts "network.interfaces.#{ i }.statistics.physical.#{ phy }.rx.bytes:int:#{ ibytes.to_i }"
        puts "network.interfaces.#{ i }.statistics.physical.#{ phy }.tx.packets:int:#{ opkt.to_i }"
        puts "network.interfaces.#{ i }.statistics.physical.#{ phy }.tx.errors:int:#{ oerr.to_i }"
        puts "network.interfaces.#{ i }.statistics.physical.#{ phy }.tx.bytes:int:#{ obytes.to_i }"
        puts "network.interfaces.#{ i }.statistics.physical.#{ phy }.tx.collisions:int:#{ coll.to_i }"

        phy += 1
      else
        ns_interface, _, network, ip, ipkt, _, _, ibytes, opkt, _, obytes, _ = line.split(/\s+/)
        network, cidr = network.split('/', 2)

        puts "network.interfaces.#{ i }.statistics.logical.#{ log }.network:str:#{ network }"
        puts "network.interfaces.#{ i }.statistics.logical.#{ log }.cidr:int:#{ cidr.to_i }"
        puts "network.interfaces.#{ i }.statistics.logical.#{ log }.address:str:#{ ip }"
        puts "network.interfaces.#{ i }.statistics.logical.#{ log }.rx.packets:int:#{ ipkt.to_i }"
        puts "network.interfaces.#{ i }.statistics.logical.#{ log }.rx.bytes:int:#{ ibytes.to_i }"
        puts "network.interfaces.#{ i }.statistics.logical.#{ log }.tx.packets:int:#{ opkt.to_i }"
        puts "network.interfaces.#{ i }.statistics.logical.#{ log }.tx.bytes:int:#{ obytes.to_i }"

        log += 1
      end
    end
  end
end

global_addresses.each_index do |i|
  puts "network.ip.#{ i }:str:#{ global_addresses[i] }"
end
