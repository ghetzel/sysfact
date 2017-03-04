#!/usr/bin/env ruby
SI = ['K', 'M', 'G', 'T', 'P', 'E', 'Z', 'Y']

# Quit unless we find zpool
if IO.popen('which zpool').read.empty?
  exit(0)
else
  pools = {}

  IO.popen("zpool list -H").read.split("\n").each do |pool_line|
      pool_name = pool_line.strip.chomp.split(/\s+/).first

      pools[pool_name.to_sym] ||= {
        'pool': pool_name,
      }

      IO.popen("zpool get all '#{ pool_name }'").read.split("\n").each do |line|
        next if line =~ /^NAME\s+/

        pool_name, property, value, source = line.strip.chomp.split(/\s+/, 4)
        next if value == '-'

        property.gsub!(/@/, '.')

        # handle sizes
        if value =~ /^(\d+(?:\.\d+)?)(K|M|G|T|P|E|Z|Y)$/
          factor = $~[1].to_f
          suffix = $~[2]
          exponent = SI.index(suffix)
          value = (factor * (1024**(exponent+1)))
        end

        # handle property-specific values
        case property
        when 'health'
          pools[pool_name.to_sym][property] = value.downcase
          pools[pool_name.to_sym]['health_value'] = case value.downcase.to_sym
          when :online then '0'
          when :degraded then '1'
          else '2'
          end

        when 'fragmentation', 'capacity'
          pools[pool_name.to_sym]["#{ property }_percent"] = value.to_s.gsub('%', '').to_f.to_s

        when 'dedupratio'
          pools[pool_name.to_sym][property] = value.gsub('x', '')

        else
          pools[pool_name.to_sym][property] = value
        end
      end
  end

  i = 0

  pools.each do |pool, properties|
    properties.each do |key, value|
      type = 'str'

      case value.to_s
      when 'on'
        type = 'bool'
        value = true
      when 'off'
        type = 'bool'
        value = false
      when /^-?\d+\.\d+$/
        type = 'float'
      when /^-?\d+$/
        type = 'int'
      end

      puts "zfs.pools.#{ i }.#{ key }:#{ type }:#{ value }"
    end

    i += 1
  end
end
