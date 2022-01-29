package data

import (
	"github.com/ghetzel/go-stockutil/stringutil"
)

type System struct {
}

func (self System) Collect() map[string]interface{} {
	out := make(map[string]interface{})

	out[`uuid`] = shellfne(
		`dmidecode -s system-uuid`,
		`sysctl -n kern.uuid`,
	).String()

	if sn := shell(`dmidecode -s system-serial-number`); !sn.IsZero() {
		out[`system.serial`] = sn.String()
	}

	out[`system.vendor`] = shellfne(
		`dmidecode -s system-manufacturer`,
		`sysctl -n machdep.cpu.brand_string | cut -d' ' -f1`,
	).String()

	out[`system.model`] = shell(`dmidecode -s system-product-name`).String()
	out[`system.revision`] = shell(`dmidecode -s system-version`).String()
	out[`system.bios.vendor`] = shell(`dmidecode -s bios-vendor`).String()
	out[`system.bios.version`] = shell(`dmidecode -s bios-version`).String()
	out[`system.bios.release`] = shell(`dmidecode -s bios-release-date`).String()
	out[`system.board.vendor`] = shell(`dmidecode -s baseboard-manufacturer`).String()
	out[`system.board.model`] = shell(`dmidecode -s baseboard-product-name`).String()
	out[`system.board.version`] = shell(`dmidecode -s baseboard-version`).String()
	out[`system.board.asset_tag`] = shell(`dmidecode -s baseboard-asset-tag`).String()
	out[`system.board.serial`] = shell(`dmidecode -s baseboard-serial-number`).String()
	out[`system.chassis.vendor`] = shell(`dmidecode -s chassis-manufacturer`).String()
	out[`system.chassis.version`] = shell(`dmidecode -s chassis-version`).String()
	out[`system.chassis.serial`] = shell(`dmidecode -s chassis-serial-number`).String()
	out[`system.chassis.asset_tag`] = shell(`dmidecode -s chassis-asset-tag`).String()
	out[`system.chassis.type`] = shell(`dmidecode -s chassis-type`).String()

	if sn := shell(`system_profiler SPHardwareDataType`); !sn.IsZero() {
		for _, line := range sn.Split("\n") {
			var k, v = stringutil.SplitPairTrimSpace(line, `:`)

			switch k {
			case `Serial Number (system)`:
				out[`system.serial`] = v
			case `Hardware UUID`:
				out[`uuid`] = v
			case `Model Name`:
				out[`system.model`] = v
			case `System Firmware Version`:
				out[`system.darwin.system_firmware_version`] = v
			case `OS Loader Version`:
				out[`system.darwin.os_loader_version`] = v
			case `Provisioning UDID`:
				out[`system.darwin.provisioning_udid`] = v
			}
		}
	}

	return out
}
