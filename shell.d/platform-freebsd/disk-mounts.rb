#!/usr/bin/env ruby

# Mounted block devices (from '/proc/mounts' and 'df')

if IO.popen('which mount df').read.empty?
  exit(0)
else
  begin
    i = 0
    IO.popen('mount -p').lines.each do |line|
        device, mount, filesystem, flags, dump, pass  = line.strip.chomp.split(/\s+/, 6)

        dfout = %x{ df -P -k '#{ mount }' 2> /dev/null }.lines.to_a[1]
        _fs, total, used, available, percent_used = dfout.strip.chomp.split(/\s+/)

        total     = total.to_i * 1024
        used      = total.to_i * 1024
        available = total.to_i * 1024

        puts "disk.mounts.#{ i }.mount:str:#{ mount }"
        puts "disk.mounts.#{ i }.device:str:#{ device }"
        puts "disk.mounts.#{ i }.filesystem:str:#{ filesystem }"

        puts "disk.mounts.#{ i }.total:int:#{ total }"
        puts "disk.mounts.#{ i }.available:int:#{ available }"
        puts "disk.mounts.#{ i }.used:int:#{ used }"

        if percent_used =~ /^\d+%$/
            puts "disk.mounts.#{ i }.percent_used:float:#{ percent_used.tr('%','') }"
        end

        flags.split(',').each do |flag|
            case flag
            when 'rw'
                puts "disk.mounts.#{ i }.readonly:bool:false"
            when 'ro'
                puts "disk.mounts.#{ i }.readonly:bool:true"
            end
        end

        i += 1
    end
  rescue
  end
end
