#!/usr/bin/env ruby

def process_value(key, value)
    case key
    when 'field.PSV', 'field.HOT', 'field.CRT', 'field.ACx', 'temperature'
        if value.is_a?(Array)
            value = value.collect{|i| i.sub(/C$/, '') }
        else
            value = value.sub(/C$/, '')
        end
    when 'passive_cooling', 'reset_video', 'handle_reboot', 'disable_on_reboot', 'verbose', 's4bios'
        value = (value == '1' ? 'true' : 'false')
    end

    case value
    when '-1', 'NONE'
        return nil
    else
        return value
    end

end

def get_type(value)
    if value == 'true' || value == 'false'
        return 'bool'
    elsif value =~ /^\-?[0-9]+\.[0-9]+$/
        return 'float'
    elsif value =~ /^\-?[0-9]+$/
        return 'int'
    else
        return 'str'
    end
end

acpi_output = IO.popen("sysctl -a").read.split("\n").select{|i| i =~ /^hw\.acpi\./ }.map(&:strip).compact

if not acpi_output.empty?
    zone_count = acpi_output.select{|i| /^hw\.acpi\.thermal\.tz\d+\./ }.collect{|i|
        i.split('.')[3]
    }.compact.select{|i|
        i.start_with?('tz')
    }.sort.uniq.collect{|i|
        i.sub('tz', '').to_i
    }.last

    (zone_count + 1).times do |tz|
        tz_output = acpi_output.select{|i| /^hw\.acpi\.thermal\.tz#{ tz }\./ }

        if tz_output.empty?
            break
        end

        tz_output.collect{|i|
            key, value = i.split(': ', 2)
            key = key.split('.').last
            key = key.gsub(/^_/, 'field.')

            case key
            when 'field.ACx', 'supported_sleep_state'
                value = value.split(' ')
            end

            value = process_value(key, value)

            if value.nil?
                next
            elsif value.is_a?(Array)
                i = 0

                value.collect{|v|
                    v = process_value(key, v)

                    if v.nil?
                        nil
                    else
                        line = "acpi.thermal.zones.#{ tz }.#{ key }.#{ i }:#{ get_type(v) }:#{ v }"
                        i+=1
                        line
                    end
                }.compact
            else
                [ "acpi.thermal.zones.#{ tz }.#{ key }:#{ get_type(value) }:#{ value }" ]
            end
        }.compact.flatten.uniq.each do |line|
            puts line
        end
    end
end

# hw.acpi.thermal.tz1._TSP: 10
# hw.acpi.thermal.tz1._TC2: 5
# hw.acpi.thermal.tz1._TC1: 1
# hw.acpi.thermal.tz1._ACx: -1 -1 -1 -1 -1 -1 -1 -1 -1 -1
# hw.acpi.thermal.tz1._CRT: 106.0C
# hw.acpi.thermal.tz1._HOT: -1
# hw.acpi.thermal.tz1._PSV: 106.0C
# hw.acpi.thermal.tz1.thermal_flags: 0
# hw.acpi.thermal.tz1.passive_cooling: 1
# hw.acpi.thermal.tz1.active: -1
# hw.acpi.thermal.tz1.temperature: 29.8C
# hw.acpi.thermal.tz0._TSP: -1
# hw.acpi.thermal.tz0._TC2: -1
# hw.acpi.thermal.tz0._TC1: -1
# hw.acpi.thermal.tz0._ACx: 85.0C 55.0C 0.0C 0.0C 0.0C -1 -1 -1 -1 -1
# hw.acpi.thermal.tz0._CRT: 106.0C
# hw.acpi.thermal.tz0._HOT: -1
# hw.acpi.thermal.tz0._PSV: -1
# hw.acpi.thermal.tz0.thermal_flags: 0
# hw.acpi.thermal.tz0.passive_cooling: 0
# hw.acpi.thermal.tz0.active: 2
# hw.acpi.thermal.tz0.temperature: 27.8C
# hw.acpi.thermal.user_override: 0
# hw.acpi.thermal.polling_rate: 10
# hw.acpi.thermal.min_runtime: 0
# hw.acpi.cpu.cx_lowest: C1
# hw.acpi.reset_video: 0
# hw.acpi.handle_reboot: 1
# hw.acpi.disable_on_reboot: 0
# hw.acpi.verbose: 0
# hw.acpi.s4bios: 0
# hw.acpi.sleep_delay: 1
# hw.acpi.suspend_state: S3
# hw.acpi.standby_state: NONE
# hw.acpi.lid_switch_state: NONE
# hw.acpi.sleep_button_state: S3
# hw.acpi.power_button_state: S5
# hw.acpi.supported_sleep_state: S3 S4 S5


# echo "acpi.thermal."

# echo "os.family:str:freebsd"
# echo "os.distribution:str:$(uname -s)"
# echo "os.version:str:$(uname -r)"
# echo "os.description:str:$(uname -s) $(uname -r) $(uname -i)"
