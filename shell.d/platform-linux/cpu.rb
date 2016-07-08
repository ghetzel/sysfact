#!/usr/bin/env ruby

if File.readable?('/proc/cpuinfo')
  cpuinfo = File.read('/proc/cpuinfo').lines.collect{|l| l.strip.chomp.squeeze(' ') }

# pull CPU data
#   At the moment, this will use details from the first processor to represent all of them.
#
#   Systems that support multiple physical processors with different capabilities is
#   currently an unsupported edge case.
#
#   This simplifies the output and reduces the complexity at the expense of handling a rare condition
#
  logical_count  = ( cpuinfo.select{|l| l =~ /^processor\s+:/i }.length                                                 rescue nil )
  physical_count = ( cpuinfo.select{|l| l =~ /^physical id\s+:/i }.collect{|l| l.split(/\s*:\s*/, 2).last }.uniq.length rescue nil )
  speed          = ( cpuinfo.select{|l| l =~ /^cpu mhz\s+:/i }.first.split(/\s*:\s*/, 2).last.to_f.round(0)             rescue nil )
  model          = ( cpuinfo.select{|l| l =~ /^model name\s+:/i }.first.split(/\s*:\s*/, 2).last                        rescue nil )
  flags          = ( cpuinfo.select{|l| l =~ /^flags\s+:/i }.first.split(/\s*:\s*/, 2).last.split(/\s+/)                rescue nil )

# dump CPU data
  puts "cpu.count:int:#{ logical_count }"     unless logical_count.nil?
  puts "cpu.physical:int:#{ physical_count }" unless physical_count.nil?
  puts "cpu.speed:int:#{ speed }"             unless speed.nil?
  puts "cpu.model:str:#{ model }"             unless model.nil?

# dump all CPU flags
  unless flags.nil?
    flags.sort.each_index do |i|
      flag = flags[i]

      puts "cpu.flags.#{ i }:str:#{ flag }"
    end
  end
end
