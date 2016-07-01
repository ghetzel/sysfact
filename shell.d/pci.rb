#!/usr/bin/env ruby

begin
# TODO: remove this once we figure out a better way to index documents
#       presently this causes many nodes to exceed 32KB docsize, causing indexing to fail
  exit 0

  lspci = %x{lspci -vmmknnD}

# exit early with no output if this is missing
  exit 0 unless $? == 0

  i = 0
  curdev = {}

  lspci.split("\n").each do |line|
    k, v = line.strip.chomp.split(/:\s+/, 2)

    if not v.nil? and not v.empty?
      k = k.downcase.to_sym
      v = v.strip


      case k
      when :slot
        domain, bus, slot_func = v.split(':', 3)
        slot, function = slot_func.split('.', 2)

        curdev = {
          :domain   => ("%04x" % domain.to_i(16) rescue nil),
          :bus      => ("%02x" % bus.to_i(16) rescue nil),
          :slot     => ("%02x" % slot.to_i(16) rescue nil),
          :function => ("%02x" % function.to_i(16) rescue nil),
        }

      when :class
        name, code, rest = v.split(/[\[\]]/, 3)

        curdev[:class_name] = name.strip
        curdev[:class]      = ("%04x" % code.to_i(16) rescue nil)
      when :vendor
        name, code, rest = v.split(/[\[\]]/, 3)

        curdev[:vendor_name] = name.strip
        curdev[:vendor]      = ("%04x" % code.to_i(16) rescue nil)
      when :device
        name, code, rest = v.split(/[\[\]]/, 3)

        curdev[:device_name] = name.strip
        curdev[:device]      = ("%04x" % code.to_i(16) rescue nil)
      when :svendor
        name, code, rest = v.split(/[\[\]]/, 3)

        curdev[:subsystem_vendor_name] = name.strip
        curdev[:subsystem_vendor]      = ("%04x" % code.to_i(16) rescue nil)
      when :sdevice
        name, code, rest = v.split(/[\[\]]/, 3)

        curdev[:subsystem_device_name] = name.strip
        curdev[:subsystem_device]      = ("%04x" % code.to_i(16) rescue nil)
      when :rev
        curdev[:revision] = ("%02x" % v.to_i(16) rescue nil)
      when :driver
        curdev[:kernel_driver] = v
      end

    else
      if not curdev.empty?
        curdev.each do |kk, vv|
          typ = 'str'

          if vv.is_a?(Float)
            typ = 'float'
          elsif vv.is_a?(Integer)
            typ = 'int'
          end


          puts "pci.devices.#{ i }.#{ kk }:#{ typ }:#{ vv }"
        end

        i += 1
      end
    end
  end
end
