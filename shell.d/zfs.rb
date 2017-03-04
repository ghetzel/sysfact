#!/usr/bin/env ruby
SI = ['K', 'M', 'G', 'T', 'P', 'E', 'Z', 'Y']

def autotype(value)
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

    return [value, type]
end

def state_to_status(health)
  case health.downcase.to_sym
  when :online then '0'
  when :degraded then '1'
  else '2'
  end
end

# Quit unless we find zpool
if IO.popen('which zpool').read.empty?
  exit(0)
else
  pools = {}

  IO.popen("zpool list -H").read.split("\n").each do |pool_line|
      pool_name = pool_line.strip.chomp.split(/\s+/).first

      pools[pool_name.to_sym] ||= {
        'name': pool_name,
      }

      # get pool properties
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
          pools[pool_name.to_sym]['health_value'] = state_to_status(value)

        when 'fragmentation', 'capacity'
          pools[pool_name.to_sym]["#{ property }_percent"] = value.to_s.gsub('%', '').to_f.to_s

        when 'dedupratio'
          pools[pool_name.to_sym][property] = value.gsub('x', '')

        else
          pools[pool_name.to_sym][property] = value
        end
      end

      # get status of zvols and disks
      in_config = false
      current_zvol = nil
      zvols = {}
      devices = {}
      errors = false

      IO.popen("zpool status '#{ pool_name }'").read.split("\n").each do |line|
        case line
        when /^\s*scan:/

        when /^\s*config:/
          in_config = true
          next

        when /^\s*errors:/
          in_config = false

          _, message = line.strip.chomp.split(': ', 2)

          if message != 'No known data errors'
            errors = true
          end
        end

        if in_config
          line = line.strip.chomp
          next if line.empty?

          name, state, read_err, write_err, cksum_err = line.split(/\s+/)
          next if name == pool_name

          case name
          when /^NAME/
            next
          when /^(mirror|raidz)/
            current_zvol = name.to_sym

            zvols[current_zvol] = {
              'name' => name,
              'type' => name.gsub(/-\d+$/, ''),
              'state' => state.downcase.to_sym,
              'state_value' => state_to_status(state),
              'errors.read' => read_err.to_i,
              'errors.write' => write_err.to_i,
              'errors.cksum' => cksum_err.to_i,
              'devices' => [],
            }
          else
            if current_zvol.nil?
              devices[name.to_sym] = {
                'name' => name,
                'state' => state.downcase.to_sym,
                'state_value' => state_to_status(state),
                'errors.read' => read_err.to_i,
                'errors.write' => write_err.to_i,
                'errors.cksum' => cksum_err.to_i,
              }
            else
              zvols[current_zvol]['devices'] << {
                'name' => name,
                'state' => state.downcase.to_sym,
                'state_value' => state_to_status(state),
                'errors.read' => read_err.to_i,
                'errors.write' => write_err.to_i,
                'errors.cksum' => cksum_err.to_i,
              }
            end
          end
        end
      end

      pools[pool_name.to_sym]['zvol'] = zvols
      pools[pool_name.to_sym]['devices'] = devices
      pools[pool_name.to_sym]['errors'] = errors
  end

  i = 0

  pools.each do |pool, properties|
    properties.each do |key, value|
      value, type = autotype(value)

      case key
      when 'zvol'
        zi = 0
        value.each do |name, zvol|
          zvol.each do |k, v|
            case k
            when 'devices'
              zdi = 0

              v.each do |device|
                device.each do |dk, dv|
                  dv, t = autotype(dv)
                  puts "zfs.pools.#{ i }.zvol.#{ zi }.devices.#{ zdi }.#{ dk }:#{ t }:#{ dv }"
                end

                zdi += 1
              end

            else
              v, t = autotype(v)
              puts "zfs.pools.#{ i }.zvol.#{ zi }.#{ k }:#{ t }:#{ v }"
            end
          end

          zi += 1
        end

      when 'devices'
        di = 0
        value.each do |name, device|
          device.each do |k, v|
            v, t = autotype(v)
            puts "zfs.pools.#{ i }.devices.#{ di }.#{ k }:#{ t }:#{ v }"
          end

          di += 1
        end

      else
        puts "zfs.pools.#{ i }.#{ key }:#{ type }:#{ value }"
      end
    end

    i += 1
  end
end
