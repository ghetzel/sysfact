#!/usr/bin/env python2.7
import os
import re
from subprocess import check_output, check_call, CalledProcessError

INFO_FIELD_MAP = {
    'model_family':     'family',
    'device_model':     'model',
    'lu_wwn_device_id': 'wwn',
    'user_capacity':    'capacity',
}

try:
    with open(os.devnull, 'w') as devnull:
        check_call(['which', 'smartctl'], stdout=devnull, stderr=devnull)

    try:
        i = 0
        attr_i = 0

        for devline in check_output(['smartctl', '--scan']).split("\n"):
            parts = re.split(r'\s+', devline.strip())

            if len(parts):
                device = parts[0]
                devtype = None

                if len(parts) >= 3:
                    devtype = parts[2]

            with open(os.devnull, 'w') as devnull:
                smartout = check_output([
                    'smartctl', '-i', '-A', device,
                ], stderr=devnull).split("\n")

            print("disk.smart.{}.device:str:{}".format(i, device))

            if devtype:
                print("disk.smart.{}.device_type:str:{}".format(i, devtype))

            in_section_info = False
            in_section_data = False

            for line in smartout:
                try:
                    if 'INFORMATION SECTION' in line.upper():
                        in_section_info = True
                        in_section_data = False

                    elif 'DATA SECTION' in line.upper():
                        in_section_info = False
                        in_section_data = True

                    elif in_section_info:
                        field, value = re.split(r':\s+', line.strip(), 1)
                        field = re.sub(r'\s+', '_', field.lower())
                        value = value.strip()
                        dtype = 'str'

                        if field == 'user_capacity':
                            value = re.sub(r'\s+bytes.*$', '', value)
                            value = int(value.replace(',', ''))
                            dtype = 'int'

                        elif field == 'lu_wwn_device_id':
                            value = re.sub(r'\s+', '', value)

                        if field in INFO_FIELD_MAP:
                            key = INFO_FIELD_MAP[field]

                            print("disk.smart.{}.{}:{}:{}".format(i, key, dtype, value))

                    elif in_section_data and re.match(r'^\s*\d+\s', line):
                        attr_id, attr_name, flag, value, worst, threshold, attr_type, updated, \
                            when_failed, raw = re.split(r'\s+', line.strip(), 9)

                        attr_name = attr_name.lower()
                        attr_type = attr_type.lower()
                        updated = updated.lower()

                        if when_failed == '-':
                            when_failed = None

                        print("disk.smart.{}.attributes.{}.id:str:{}".format(i, attr_i, attr_id))
                        print("disk.smart.{}.attributes.{}.name:str:{}".format(i, attr_i, attr_name))
                        print("disk.smart.{}.attributes.{}.flag:int:{}".format(i, attr_i, int(flag, 16)))
                        print("disk.smart.{}.attributes.{}.value:int:{}".format(i, attr_i, int(value)))
                        print("disk.smart.{}.attributes.{}.raw_value:int:{}".format(i, attr_i, int(raw)))
                        print("disk.smart.{}.attributes.{}.worst:int:{}".format(i, attr_i, int(worst)))
                        print("disk.smart.{}.attributes.{}.threshold:int:{}".format(i, attr_i, int(threshold)))
                        print("disk.smart.{}.attributes.{}.type:str:{}".format(i, attr_i, attr_type))
                        print("disk.smart.{}.attributes.{}.update_freq:str:{}".format(i, attr_i, updated))

                        if when_failed:
                            print("disk.smart.{}.attributes.{}.when_failed:str:{}".format(i, attr_i, when_failed))

                        attr_i += 1

                except:
                    continue

            i += 1

    except:
        pass

except CalledProcessError:
    os.exit(0)
