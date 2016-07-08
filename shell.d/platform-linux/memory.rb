#!/usr/bin/env ruby

# System memory utilization, configuration

def toBytes(derp)
  if derp.downcase.include?('tb')
    return derp.gsub(/[^0-9,.]/, "").to_i * 1099511627776
  end
  if derp.downcase.include?('gb')
    return derp.gsub(/[^0-9,.]/, "").to_i * 1073741824
  end
  if derp.downcase.include?('mb')
    return derp.gsub(/[^0-9,.]/, "").to_i * 1048576
  end
  if derp.downcase.include?('kb')
    return derp.gsub(/[^0-9,.]/, "").to_i * 1024
  end
  return derp.gsub(/[^0-9,.]/, "").to_i
end

# Total memory information.
if File.readable?('/proc/meminfo')
  begin
    File.read('/proc/meminfo').lines.each do |line|
      if line.split(' ')[0] == 'MemTotal:'
        puts "memory.total:int:#{line.split(' ')[1].strip.chomp.to_i * 1024}"
      end
      if line.split(' ')[0] == 'MemFree:'
        puts "memory.free:int:#{line.split(' ')[1].strip.chomp.to_i * 1024}"
      end
      if line.split(' ')[0] == 'SwapTotal:'
        puts "memory.swap:int:#{line.split(' ')[1].strip.chomp.to_i * 1024}"
      end
    end
  rescue
    puts "FATAL: Could not read /proc/meminfo"
    exit 2
  end
else
  puts "FATAL: Could not read /proc/meminfo"
  exit 2
end

# Bank information.
begin
  dmi16 = %x{ dmidecode -t 16 2> /dev/null }
  raise "dmidecode: #{ $? }" unless $? == 0

  bindex = 0
  dmi16.split('Physical Memory Array').each do |bank|
    if bindex != 0
      bank.lines.each do |line|
        if line.strip.chomp.split(':')[0] == 'Maximum Capacity'
          cap = line.strip.chomp.split(':')[1].strip
          puts "memory.bank.#{bindex-1}.capacity:int:#{toBytes(cap)}"
        end
        if line.strip.chomp.split(':')[0] == 'Error Correction Type'
          if line.strip.chomp.split(':')[1].downcase.include?('none')
            puts "memory.bank.#{bindex-1}.ecc:bool:false"
            puts "memory.bank.#{bindex-1}.ecc_type:str:none"

          else
            puts "memory.bank.#{bindex-1}.ecc:bool:true"
            puts "memory.bank.#{bindex-1}.ecc_type:str:#{line.strip.strip.chomp.split(':')[1]}"
          end
        end
        if line.strip.chomp.split(':')[0] == 'Number Of Devices'
          puts "memory.bank.#{bindex-1}.slot_count:int:#{line.strip.chomp.split(':')[1].chomp.strip}"
        end
      end
    end
    bindex += 1
  end
rescue Exception => e
  STDERR.puts("Unable to read physical memory layout details: #{ e.message }")
end

# Dimm information.
begin
  dmi17 = %x{ dmidecode -t 17 2> /dev/null }
  raise "dmidecode: #{ $? }" unless $? == 0

  dindex = 0
  populated = 0
  empty = 0
  unpopulated = false
  dmi17.split('Memory Device').each do |dimm|
    if dindex != 0
      dimm.lines.each do |line|
        if line.strip.chomp.split(':')[0] == "Size"
          if toBytes(line.strip.chomp.split(':')[1].strip.chomp) == 0
            puts "memory.slots.#{dindex-1}.empty:bool:true"
            unpopulated = true
          else
            unpopulated = false
          end
        end
      end
      if !unpopulated
        dimm.lines.each do |line|
          if line.strip.chomp.split(':')[0] == "Size"
            puts "memory.slots.#{dindex-1}.size:int:#{toBytes(line.strip.chomp.split(':')[1].strip.chomp)}"
          end
          if line.strip.chomp.split(':')[0] == "Locator"
            puts "memory.slots.#{dindex-1}.name:str:#{line.strip.chomp.split(':')[1].strip.chomp}"
          end
          if line.strip.chomp.split(':')[0] == "Form Factor"
            puts "memory.slots.#{dindex-1}.form:str:#{line.strip.chomp.split(':')[1].strip.chomp}"
          end
          if line.strip.chomp.split(':')[0] == "Data Width"
            puts "memory.slots.#{dindex-1}.bits:int:#{line.strip.chomp.split(':')[1].strip.chomp.gsub(/[^0-9,.]/, '')}"
          end
          if line.strip.chomp.split(':')[0] == "Speed"
            puts "memory.slots.#{dindex-1}.speed:int:#{line.strip.chomp.split(':')[1].strip.chomp.gsub(/[^0-9,.]/, '')}"
          end
          if line.strip.chomp.split(':')[0] == "Type"
            puts "memory.slots.#{dindex-1}.type:str:#{line.strip.chomp.split(':')[1].strip.chomp}"
          end
          if line.strip.chomp.split(':')[0] == "Manufacturer"
            puts "memory.slots.#{dindex-1}.make:str:#{line.strip.chomp.split(':')[1].strip.chomp}"
          end
          if line.strip.chomp.split(':')[0] == "Part Number"
            puts "memory.slots.#{dindex-1}.model:str:#{line.strip.chomp.split(':')[1].strip.chomp}"
          end
          if line.strip.chomp.split(':')[0] == "Serial Number"
            puts "memory.slots.#{dindex-1}.serial:str:#{line.strip.chomp.split(':')[1].strip.chomp}"
          end
          if line.strip.chomp.split(':')[0] == "Asset Tag"
            puts "memory.slots.#{dindex-1}.asset_tag:str:#{line.strip.chomp.split(':')[1].strip.chomp}"
          end
        end
        populated +=1
      else
        empty +=1
      end
    end
  dindex+= 1
  end
  puts "memory.slot_count:int:#{populated+empty}"
  puts "memory.slot_empty:int:#{empty}"
  puts "memory.slot_populated:int:#{populated}"

rescue Exception => e
  STDERR.puts("Unable to read physical memory layout details: #{ e.message }")
end
