#!/usr/bin/env ruby

# Mounted block devices (from '/proc/mounts' and 'df')

if File.readable?('/proc/mounts')
  begin
    i = 0
    File.read('/proc/mounts').lines.each do |line|
        if line.strip.chomp.split(' ')[0].chars.first == '/'
            dfout = %x{ df -P --block-size 1 #{ line.strip.chomp.split(' ')[1] } 2> /dev/null }.lines.to_a[1]

            device, mount, filesystem = line.strip.chomp.split(' ')
            _fs, total, used, available, percent_used = dfout.strip.chomp.split(' ')

            puts "disk.mounts.#{ i }.mount:str:#{ mount }"
            puts "disk.mounts.#{ i }.device:str:#{ device }"
            puts "disk.mounts.#{ i }.filesystem:str:#{ filesystem }"

            puts "disk.mounts.#{ i }.total:int:#{ total }"
            puts "disk.mounts.#{ i }.available:int:#{ available }"
            puts "disk.mounts.#{ i }.used:int:#{ used }"

            if percent_used =~ /^\d+%$/
                puts "disk.mounts.#{ i }.percent_used:float:#{ percent_used.tr('%','') }"
            end

            findex = 0
            line.strip.chomp.split(' ')[3].split(',').each do |flag|
                puts "disk.mounts.#{ i }.flags.#{ findex }:str:#{ flag }"
                findex += 1
            end

            i += 1
        end
    end
  rescue
  end
end
