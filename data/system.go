package data

type System struct {
}

func (self System) Collect() map[string]interface{} {
	out := make(map[string]interface{})

	out[`uuid`] = shell(`dmidecode -s system-uuid`).String()
	out[`system.serial`] = shell(`dmidecode -s system-serial-number`).String()
	out[`system.vendor`] = shell(`dmidecode -s system-manufacturer`).String()
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

	return out
}
