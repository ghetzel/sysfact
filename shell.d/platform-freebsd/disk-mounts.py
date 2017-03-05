#!/usr/bin/env python2.7
import os
import re
from subprocess import check_output, check_call, CalledProcessError

# Mounted block devices (from '/proc/mounts' and 'df')

try:
    with open(os.devnull, 'w') as devnull:
        check_call(['which', 'mount', 'df'], stdout=devnull, stderr=devnull)

    try:
        i = 0

        for line in check_output(['mount', '-p']).split("\n"):
            device, mount, filesystem, flags, fs_dump, fs_pass = re.split('\s+', line.strip(), 5)

            with open(os.devnull, 'w') as devnull:
                dfout = check_output([
                    'df', '-P', '-k', mount,
                ], stderr=devnull).split("\n")[1]

            _fs, total, used, available, percent_used = re.split('\s+', dfout.strip(), 4)

            total = int(total) * 1024
            used = int(total) * 1024
            available = int(total) * 1024

            print("disk.mounts.{0}.mount:str:{1}".format(i, mount))
            print("disk.mounts.{0}.device:str:{1}".format(i, device))
            print("disk.mounts.{0}.filesystem:str:{1}".format(i, filesystem))
            print("disk.mounts.{0}.total:int:{1}".format(i, total))
            print("disk.mounts.{0}.available:int:{1}".format(i, available))
            print("disk.mounts.{0}.used:int:{1}".format(i, used))

            if re.match('^\d+%$', percent_used):
                print("disk.mounts.{0}.percent_used:float:{1}".format(
                    i, percent_used.sub('%', '')
                ))

            for flag in flags.split(','):
                if flag == 'rw':
                    print("disk.mounts.{0}.readonly:bool:false".format(i))
                elif flag == 'ro':
                    print("disk.mounts.{0}.readonly:bool:true".format(i))

            i += 1

    except:
        pass

except CalledProcessError:
    os.exit(0)
