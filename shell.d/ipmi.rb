#!/usr/bin/env ruby

# Quit unless we find ipmitool
if IO.popen('which ipmitool').read.empty?
  exit(0)
else
  if IO.popen("pgrep ipmitool").read.split("\n").length > 10
    STDERR.log("Too many ipmitool processes, skipping collection")
    puts "ipmi.collection_fault:bool:true"
    exit(0)
  end

  # get our IPMI data
  ipmi_data = %x{ ipmitool lan print 1 }.lines.collect{|l| l.strip.chomp.squeeze(' ') }
  bmc_data  = %x{ ipmitool bmc info }.lines.collect{|l| l.strip.chomp.squeeze(' ') }
  guid_data = %x{ ipmitool bmc guid }.lines.collect{|l| l.strip.chomp.squeeze(' ') }

  # Build our hashes out
  ipmi_hash = {
    :datasource  => "ipmitool",
    :ip          => ( ipmi_data.select{ |l| l =~ /^IP\ Address\ \:/i }.first.split(":")[1].strip          rescue nil ),
    :netmask     => ( ipmi_data.select{ |l| l =~ /^Subnet\ Mask\ \:/i }.first.split(":")[1].strip         rescue nil ),
    :gateway     => ( ipmi_data.select{ |l| l =~ /^Default\ Gateway\ IP\ \:/i }.first.split(":")[1].strip rescue nil ),
    :mac_address => ( ipmi_data.select{ |l| l =~ /^MAC\ Address\ \:/i }.first.split(":", 2)[1].strip      rescue nil ),
    :vlan        => ( ipmi_data.select{ |l| l =~ /^802.1q\ VLAN\ ID\ \:/i }.first.split(":")[1].strip     rescue nil ),
    :version     => ( bmc_data.select{ |l| l  =~ /^IPMI\ Version\ \:/i }.first.split(":")[1].strip        rescue nil ),
  }

  bmc_hash = {
    :firmware_version => ( bmc_data.select{ |l| l  =~ /^Firmware\ Revision\ \:/i }.first.split(":")[1].strip rescue nil ),
    :guid             => ( guid_data.select{ |l| l =~ /^System\ GUID\ \:/i }.first.split(":")[1].strip       rescue nil ),
  }

  # Print out our data
  ipmi_hash.each do |key, value|
    next if value.nil? or value.empty? or value == "Disabled"
    puts "ipmi.#{key}:str:#{value}"
  end

  bmc_hash.each do |key, value|
    next if value.nil? or value.empty? or value == "Disabled"
    puts "ipmi.bmc.#{key}:str:#{value}"
  end
end
