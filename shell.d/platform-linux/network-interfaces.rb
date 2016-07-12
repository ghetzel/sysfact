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
%x{ ip link }.lines.each do |line|
  if line =~ /^\d+:\s+([^:]+):/
    interface_names << $~[1]
  end
end
interface_names.uniq!

global_addresses = []

# for each interface...
interface_names.each_index do |i|
  interface = interface_names[i]
  addresses = []
  info = {}

# get interface details
  %x{ ip address show dev #{ interface } }.lines.each do |line|
    case line
  # link state details
    when /^\d+: ([a-zA-Z]+[0-9]+): <([^\>]+)> (.*)/
      pairs = $~[3].split(' ')
      flags = $~[2].split(',').map(&:strip).map(&:downcase)
      flags.sort.each_index do |flag_i|
        info["flags.#{ flag_i }"] = flags[flag_i]
      end

      if flags.include?('pointopoint')
        info['pointopoint'] = true
      end

      pairs.each_index do |fi|
        case pairs[fi]
        when 'mtu'
          info['mtu'] = pairs[fi+1].to_i
        when 'qdisc'
          info['queue_discipline'] = pairs[fi+1]
        when 'qlen'
          info['max_queue_length'] = pairs[fi+1].to_i
        when 'state'
          info['state'] = pairs[fi+1].downcase unless pairs[fi+1] == 'UNKNOWN'
        when 'group'
          info['group'] = pairs[fi+1]
        end
      end

  # link-layer details
    when /^\s+link\/([^\s]+) ([^\s]+) brd ([^\s]+)/
      type = $~[1]
      brd  = $~[3]

      case type
      when 'ether'
        info['link_type'] = 'ethernet'
      when 'none'
        info['link_type'] = nil
      else
        info['link_type'] = type
      end

      info['mac']  = $~[2]

      if not [['0'], ['f']].include?(brd.downcase.delete(':').split('').uniq)
        info['mac_broadcast'] = brd
      end

  # IP details
    when /^\s+(inet|inet6|bridge|ipx|dnet|link|any) ([^\s]+) (brd|peer) ([^\s]+)/
      info['ip_family'] = $~[1]
      addr_i            = addresses.length
      addr              = $~[2]
      addresses        << addr
      ip, cidr          = addr.split('/', 2)

      cotype            = $~[3]

      if cotype == 'brd'
        brd               = $~[4]
      else
        peer              = $~[4]
      end

      global_addresses << ip

    # parse IP address
      info["addresses.#{ addr_i }.ip"]        = ip
      info["addresses.#{ addr_i }.cidr"]      = cidr.to_i          unless cidr.nil?
      info["addresses.#{ addr_i }.netmask"]   = cidr_to_mask(cidr) unless cidr.nil?
      info["addresses.#{ addr_i }.broadcast"] = brd                unless brd.nil?

    # add peer address (e.g.: pointopoint interfaces)
      unless peer.nil?
        peer_ip, peer_cidr = peer.split('/', 2)

        info["addresses.#{ addr_i }.peer.ip"]      = peer_ip
        info["addresses.#{ addr_i }.peer.cidr"]    = peer_cidr.to_i          unless peer_cidr.nil?
        info["addresses.#{ addr_i }.peer.netmask"] = cidr_to_mask(peer_cidr) unless peer_cidr.nil?
      end
    end
  end


# output interface details
  puts "network.interfaces.#{ i }.name:str:#{ interface }"

  info.each do |k,v|
    t = 'str'
    t = 'int' if v.is_a?(Integer)
    t = 'bool' if v.is_a?(TrueClass) or v.is_a?(FalseClass)

    puts "network.interfaces.#{ i }.#{ k }:#{ t }:#{ v }" unless v.nil?
  end


# include LLDP details for this interface
# %x{ lldpctl -f keyvalue em1 2> /dev/null }
#  if $? == 0
#    network.interfaces.#{ i }.switch.name
#    network.interfaces.#{ i }.switch.description
#    network.interfaces.#{ i }.switch.bridge:bool:true/false
#    network.interfaces.#{ i }.switch.router:bool:true/false
#    network.interfaces.#{ i }.switch.ip
#    network.interfaces.#{ i }.switch.port
#    network.interfaces.#{ i }.switch.tagged_vlans.0.id
#    network.interfaces.#{ i }.switch.tagged_vlans.0.name
#    network.interfaces.#{ i }.switch.tagged_vlans.1.id
#    network.interfaces.#{ i }.switch.tagged_vlans.1.name
#    network.interfaces.#{ i }.switch.tagged_vlans...
#    network.interfaces.#{ i }.switch.default_vlan
#    network.interfaces.#{ i }.switch.name


#  end
end

global_addresses.each_index do |i|
  puts "network.ip.#{ i }:str:#{ global_addresses[i] }"
end
