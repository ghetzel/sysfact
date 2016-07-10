#!/usr/bin/env ruby

def print_output_if(field, cmd)
  value = %x{ #{ cmd } 2> /dev/null }.lines.reject{|l| l =~ /^\s*#/ }.first.strip.chomp
  return if value.empty?
  return if value.downcase == 'not specified'
  return if value.downcase == 'to be filled by o.e.m.'

  puts "#{ field }:str:#{ value }"
rescue
  nil
end

fqdn = %x{ uname -n }.strip.chomp
puts "fqdn:str:#{ fqdn }"

print_output_if 'uuid',                     'dmidecode -s system-uuid'
print_output_if 'system.serial',            'dmidecode -s system-serial-number'
print_output_if 'system.vendor',            'dmidecode -s system-manufacturer'
print_output_if 'system.model',             'dmidecode -s system-product-name'
print_output_if 'system.revision',          'dmidecode -s system-version'
print_output_if 'system.bios.vendor',       'dmidecode -s bios-vendor'
print_output_if 'system.bios.version',      'dmidecode -s bios-version'
print_output_if 'system.bios.release',      'dmidecode -s bios-release-date'
print_output_if 'system.board.vendor',      'dmidecode -s baseboard-manufacturer'
print_output_if 'system.board.model',       'dmidecode -s baseboard-product-name'
print_output_if 'system.board.version',     'dmidecode -s baseboard-version'
print_output_if 'system.board.asset_tag',   'dmidecode -s baseboard-asset-tag'
print_output_if 'system.board.serial',      'dmidecode -s baseboard-serial-number 2> /dev/null | tr -d "."'
print_output_if 'system.chassis.vendor',    'dmidecode -s chassis-manufacturer'
print_output_if 'system.chassis.version',   'dmidecode -s chassis-version'
print_output_if 'system.chassis.serial',    'dmidecode -s chassis-serial-number'
print_output_if 'system.chassis.asset_tag', 'dmidecode -s chassis-asset-tag'
print_output_if 'system.chassis.type',      'dmidecode -s chassis-type'
