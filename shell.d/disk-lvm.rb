#!/usr/bin/env ruby

# Volume group information
begin
  vgd = %x{ vgdisplay -c 2> /dev/null  }

# exit early with no output if this is missing  
  exit 0 unless $? == 0

  id = 0
  vgs = Hash.new
  vgd.lines.each do |vg|
  	vgs[vg.split(':')[0].strip.chomp] = id
    puts "disk.lvm.groups.#{ id }.name:str:#{ vg.split(':')[0].strip.chomp }"
    puts "disk.lvm.groups.#{ id }.size:int:#{ vg.split(':')[11].strip.chomp }"
    puts "disk.lvm.groups.#{ id }.uuid:str:#{ vg.split(':')[16].strip.chomp }"
    puts "disk.lvm.groups.#{ id }.extents.allocated:int:#{ vg.split(':')[14].strip.chomp }"
    puts "disk.lvm.groups.#{ id }.extents.free:int:#{ vg.split(':')[15].strip.chomp }"
    puts "disk.lvm.groups.#{ id }.extents.size:int:#{ vg.split(':')[12].strip.chomp }"
    puts "disk.lvm.groups.#{ id }.extents.total:int:#{ vg.split(':')[13].strip.chomp }"
    id += 1
  end
end

# Volume group information
begin
  vgd = %x{ pvdisplay -c 2> /dev/null  }

# exit early with no output if this is missing
  exit 0 unless $? == 0

  sectors = Hash.new
  diskid = 0
  vgd.lines.each do |vg|
  	id = vgs[vg.split(':')[1].strip.chomp]
    puts "disk.lvm.groups.#{ id }.disks.#{ diskid }.name:str:#{ vg.split(':')[0].strip.chomp }"
    puts "disk.lvm.groups.#{ id }.disks.#{ diskid }.vgname:str:#{ vg.split(':')[1].strip.chomp }"
    puts "disk.lvm.groups.#{ id }.disks.#{ diskid }.size:int:#{ vg.split(':')[2].strip.chomp }"
    puts "disk.lvm.groups.#{ id }.disks.#{ diskid }.uuid:str:#{ vg.split(':')[11].strip.chomp }"
    puts "disk.lvm.groups.#{ id }.disks.#{ diskid }.extents.allocated:int:#{ vg.split(':')[10].strip.chomp }"
    puts "disk.lvm.groups.#{ id }.disks.#{ diskid }.extents.free:int:#{ vg.split(':')[9].strip.chomp }"
    puts "disk.lvm.groups.#{ id }.disks.#{ diskid }.extents.size:int:#{ vg.split(':')[7].strip.chomp }"
    puts "disk.lvm.groups.#{ id }.disks.#{ diskid }.extents.total:int:#{ vg.split(':')[8].strip.chomp }"

    # if the device is a partition get sectorsize from its parent
    blockdev = %x{ lsblk -nrso NAME,TYPE #{ vg.split(':')[0].strip.chomp } | grep -v ' part' }.split(' ')[0].strip.chomp
    sectorsze = %x{ cat /sys/block/#{ blockdev }/queue/hw_sector_size }.strip.chomp.to_i

    puts "disk.lvm.groups.#{ id }.disks.#{ diskid }.sector_size:int:#{ sectorsze }"
    sectors[vg.split(':')[1].strip.chomp] = sectorsze
    diskid += 1
  end
end

# Logical Volume information
begin
  vgd = %x{ lvdisplay -c 2> /dev/null  }
  
# exit early with no output if this is missing  
  exit 0 unless $? == 0

  volumeid = 0
  vgd.lines.each do |vg|
  	id = vgs[vg.split(':')[1].strip.chomp]
    puts "disk.lvm.groups.#{ id }.volumes.#{ volumeid }.name:str:#{ vg.split(':')[0].strip.chomp }"
    puts "disk.lvm.groups.#{ id }.volumes.#{ volumeid }.vgname:str:#{ vg.split(':')[1].strip.chomp }"
    puts "disk.lvm.groups.#{ id }.volumes.#{ volumeid }.size:int:#{ vg.split(':')[6].strip.chomp.to_i * sectors[vg.split(':')[1].strip.chomp] }"
    puts "disk.lvm.groups.#{ id }.volumes.#{ volumeid }.sectors:int:#{ vg.split(':')[6].strip.chomp }"
    puts "disk.lvm.groups.#{ id }.volumes.#{ volumeid }.extents.associated:int:#{ vg.split(':')[7].strip.chomp }"
    puts "disk.lvm.groups.#{ id }.volumes.#{ volumeid }.extents.allocated:int:#{ vg.split(':')[8].strip.chomp }"
    volumeid += 1
  end
end

