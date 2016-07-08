#!/usr/bin/env ruby

os = {}

if File.readable?('/etc/redhat-release')
  os[:family] = 'redhat'

  if File.read('/etc/redhat-release').lines.first.chomp.strip =~ /^(.*) release ([0-9\.]+)/
    os[:distribution] = $~[1]
    os[:version]      = $~[2]
    os[:installed_at] = ( File.stat('/root/install.log').mtime.strftime('%Y-%m-%dT%H:%M:%S%z') rescue nil )
  end

elsif File.executable?('/usr/bin/lsb_release')
  os[:family]       = 'debian'
  os[:distribution] = %x{ lsb_release --short --id }.lines.first.chomp
  os[:version]      = %x{ lsb_release --short --release }.lines.first.chomp
  os[:codename]     = %x{ lsb_release --short --codename }.lines.first.chomp
  os[:description]  = %x{ lsb_release --short --description }.lines.first.chomp
end


os.each do |field, value|
  next if value.nil? or value.empty?
  puts "os.#{ field }:str:#{ value }"
end
