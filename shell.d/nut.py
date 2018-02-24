#!/usr/bin/env python
import os
import sys
from subprocess import check_output, check_call, CalledProcessError


def autotype(value):
    if value is None:
        return 'str', ''

    try:
        return 'int', int(value)
    except ValueError:
        pass

    try:
        return 'float', float(value)
    except ValueError:
        pass

    if value.lower() in ['true', 'on', 'yes']:
        return 'bool', 'true'

    if value.lower() in ['false', 'off', 'no']:
        return 'bool', 'false'

    return 'str', '{}'.format(value)

try:
    with open(os.devnull, 'w') as devnull:
        check_call(
            ['which', 'upsc'],
            stdout=devnull,
            stderr=devnull,
            universal_newlines=True
        )

        try:
            ups_list = check_output(
                ['upsc', '-l'],
                stderr=devnull,
                universal_newlines=True
            ).split("\n")

            i = 0

            for ups in ups_list:
                if not len(ups.strip()):
                    continue

                try:
                    details = check_output(
                        ['upsc', ups],
                        stderr=devnull,
                        universal_newlines=True
                    ).split("\n")

                    for detail in details:
                        detail = detail.strip()

                        if not len(detail):
                            continue

                        parts = detail.split(': ', 1)
                        key = None
                        value = None

                        if len(parts) == 2:
                            key = parts[0]
                            value = parts[1]
                        else:
                            key = parts[0]

                        if 'version' in key:
                            typ = 'str'
                            value = '{}'.format(value)
                        else:
                            typ, value = autotype(value)

                        key = key.rstrip(':')

                        if key and len(key) and len('{}'.format(value)):
                            if key == 'ups.status':
                                statuses = value.strip().split(' ')

                                if 'OL' in statuses:
                                    print('nut.ups.{}.ups.powersource:str:mains'.format(i))
                                elif 'OB' in statuses:
                                    print('nut.ups.{}.ups.powersource:str:battery'.format(i))

                                if 'CHRG' in statuses:
                                    print('nut.ups.{}.ups.charging:bool:true'.format(i))
                                elif 'DISCHRG' in statuses:
                                    print('nut.ups.{}.ups.charging:bool:false'.format(i))

                            else:
                                print('nut.ups.{}.{}:{}:{}'.format(i, key, typ, value))
                except:
                    pass
                finally:
                    print('nut.ups.{}.name:str:{}'.format(i, ups))
                    i += 1
        except:
            pass

except CalledProcessError:
    sys.exit(0)
