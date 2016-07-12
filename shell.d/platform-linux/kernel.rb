#!/usr/bin/env ruby

kernel_version = %x{ uname -r }.strip.chomp
kernel_arch    = %x{ uname -i }.strip.chomp
kernel_host    = %x{ uname -n }.strip.chomp
uptime_seconds = nil
uptime_idlesec = nil
booted_at      = nil

if File.readable?('/proc/uptime')
  begin
    uptime = File.read('/proc/uptime').lines.first.strip.chomp.split(/\s+/, 2).map(&:to_f)
    now    = Time.now

    uptime_seconds = uptime[0].round
    uptime_idlesec = uptime[1].round
    booted_at      = Time.at(now.to_i - uptime_seconds)
  rescue
  end
end

puts "kernel.version:str:#{ kernel_version }"
puts "kernel.hostname:str:#{ kernel_host }"
puts "arch:str:#{ kernel_arch }"
puts "uptime:int:#{ uptime_seconds }"
puts "booted_at:date:#{ booted_at.utc.strftime('%FT%TZ') }"

