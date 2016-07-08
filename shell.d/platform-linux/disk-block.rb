#!/usr/bin/env ruby

if File.directory?('/sys/block')
# get all block devices
  devices = Dir["/sys/block/*"].collect{|block|
    next unless File.exists?(File.join(block, 'device'))
    block
  }.compact

# output data on each device
  devices.each_index do |i|
    device    = devices[i]

    name      =   File.basename(device)
    p_blksz   = ( File.read(File.join(device, 'queue', 'physical_block_size')).lines.first.to_i      rescue nil )
    l_blksz   = ( File.read(File.join(device, 'queue', 'logical_block_size')).lines.first.to_i       rescue nil )
    size      = ( File.read(File.join(device, 'size')).lines.first.to_i * (l_blksz || 512)           rescue nil )
    removable = ( File.read(File.join(device, 'removable')).lines.first.strip.chomp == '1'           rescue nil )
    ssd       = ( File.read(File.join(device, 'queue', 'rotational')).lines.first.strip.chomp == '0' rescue nil )
    readonly  = ( File.read(File.join(device, 'ro')).lines.first.strip.chomp == '1'                  rescue nil )
    vendor    = ( File.read(File.join(device, 'device', 'vendor')).lines.first.strip.chomp           rescue nil )
    model     = ( File.read(File.join(device, 'device', 'model')).lines.first.strip.chomp            rescue nil )
    revision  = ( File.read(File.join(device, 'device', 'rev')).lines.first.strip.chomp              rescue nil )


  # output
    puts "disk.block.#{ i }.name:str:#{ name }"
    puts "disk.block.#{ i }.device:str:/dev/#{ name }"           if File.exists?("/dev/#{ name }")
    puts "disk.block.#{ i }.size:int:#{ size }"                  unless size.nil?
    puts "disk.block.#{ i }.vendor:str:#{ vendor }"              unless vendor.nil?
    puts "disk.block.#{ i }.model:str:#{ model }"                unless model.nil?
    puts "disk.block.#{ i }.revision:str:#{ revision }"          unless revision.nil?
    puts "disk.block.#{ i }.removable:bool:#{ removable }"       unless removable.nil?
    puts "disk.block.#{ i }.solidstate:bool:#{ ssd }"            unless ssd.nil?
    puts "disk.block.#{ i }.readonly:bool:#{ readonly }"         unless readonly.nil?
    puts "disk.block.#{ i }.blocksize.physical:int:#{ p_blksz }" unless p_blksz.nil?
    puts "disk.block.#{ i }.blocksize.logical:int:#{ l_blksz }"  unless l_blksz.nil?
  end
end
