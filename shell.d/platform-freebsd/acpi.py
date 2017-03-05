#!/usr/bin/env python2.7
import os
import re
from subprocess import check_output

VALUE_RX = re.compile('C$')
FLOAT_RX = re.compile('^\-?[0-9]+\.[0-9]+$')
INT_RX = re.compile('^\-?[0-9]+$')
ACPI_RX = re.compile('^hw\.acpi\.')
ACPI_THERMAL_RX = re.compile('^hw\.acpi\.thermal\.tz\d+\.')


def process_value(key, value):
    if key in [
        'field.PSV',
        'field.HOT',
        'field.CRT',
        'field.ACx',
        'temperature'
    ]:
        if isinstance(value, list):
            value = [VALUE_RX.sub('', i) for i in value]
        else:
            value = VALUE_RX.sub('', str(value))

    elif key in [
        'passive_cooling',
        'reset_video',
        'handle_reboot',
        'disable_on_reboot',
        'verbose',
        's4bios'
    ]:
        if value == '1':
            value = 'true'
        else:
            value = 'false'

    if value in ['-1', 'NONE']:
        return None
    else:
        return value


def get_type(value):
    if value == 'true' or value == 'false':
        return 'bool'
    elif FLOAT_RX.match(value):
        return 'float'
    elif INT_RX.match(value):
        return 'int'
    else:
        return 'str'


with open(os.devnull, 'w') as devnull:
    acpi_output = [
        i.strip() for i in check_output([
            'sysctl', '-a'
        ], stderr=devnull).split("\n") if ACPI_RX.match(i)
    ]

if len(acpi_output):
    zone_count = [i for i in acpi_output if ACPI_THERMAL_RX.match(i)]
    zone_count = [i.split('.')[3] for i in zone_count]
    zone_count = [i for i in zone_count if i is not None]
    zone_count = [i for i in zone_count if i.startswith('tz')]
    zone_count = sorted(zone_count)
    zone_count = list(set(zone_count))
    zone_count = [int(i.replace('tz', '')) for i in zone_count]
    zone_count = zone_count[-1]

    for tz in range(zone_count + 1):
        ACPI_TZ_RX = re.compile('^hw\.acpi\.thermal\.tz{0}\.'.format(tz))
        tz_output = [i for i in acpi_output if ACPI_TZ_RX.match(i)]

        if not len(tz_output):
            break

        out = []

        for i in tz_output:
            key, value = i.split(': ', 1)
            key = key.split('.')[-1]
            key = re.sub('^_', 'field.', key)

            if key in ['field.ACx', 'supported_sleep_state']:
                value = value.split(' ')

            value = process_value(key, value)

            if value is None:
                continue
            elif isinstance(value, list):
                i = 0

                for v in value:
                    v = process_value(key, v)

                    if v is not None:
                        out.append("acpi.thermal.zones.{0}.{1}.{2}:{3}:{4}".format(
                            tz, key, i, get_type(v), v
                        ))

                        i += 1
            else:
                out.append("acpi.thermal.zones.{0}.{1}:{2}:{3}".format(
                    tz, key, get_type(value), value
                ))

        for line in sorted(list(set(out))):
            if len(line):
                print(line)
