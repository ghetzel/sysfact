#!/usr/bin/env ruby

# Quit unless we find zpool
if IO.popen('which zpool').read.empty?
  exit(0)
else
  pools = {}

  IO.popen("zpool get -H -p all").read.split("\n").each do |line|
    pool_name, property, value, source = line.strip.chomp.split("\t", 4)
    next if value == '-'

    property.gsub!(/@/, '.')

    pools[pool_name.to_sym] ||= {
      'name': pool_name,
    }

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

  i = 0

  pools.each do |pool, properties|
    properties.each do |key, value|
      type = 'str'

      case value
      when 'on'
        type = 'bool'
        value = true
      when 'off'
        type = 'bool'
        value = false
      when /^\-?[0-9]+\.[0-9]+$/
        type = 'float'
      when /^\-?[0-9]+$/
        type = 'int'
      end

      puts "zfs.pools.#{ pool }.#{ key }:#{ type }:#{ value }"
    end

    i += 1
  end
end
