package data

import (
	"encoding/xml"
	"fmt"
	"os/exec"
	"strings"

	"github.com/ghetzel/go-stockutil/log"
)

type geomConfig struct {
	End          int64  `xml:"end"`
	Entries      int    `xml:"entries"`
	First        int64  `xml:"first"`
	FwHeads      int64  `xml:"fwheads"`
	FwSectors    int64  `xml:"fwsectors"`
	Index        int    `xml:"index"`
	Label        string `xml:"label"`
	Last         int64  `xml:"last"`
	Length       int64  `xml:"length"`
	Modified     bool   `xml:"modified"`
	Offset       int64  `xml:"offset"`
	SecOffset    int64  `xml:"secoffset"`
	Scheme       string `xml:"scheme"`
	Start        int64  `xml:"start"`
	RotationRate int    `xml:"rotationrate"`
	WWID         string `xml:"lunid"`
	State        string `xml:"state"`
	Description  string `xml:"descr"`
	Identifier   string `xml:"ident"`
	Type         string `xml:"type"`
	RawType      string `xml:"rawtype"`
	RawUUID      string `xml:"rawuuid"`
	EFIMedia     string `xml:"efimedia"`
}

type geomTarget struct {
	ID           string         `xml:"id,attr,omitempty"`
	GEOM         geomDescriptor `xml:"geom"`
	Mode         string         `xml:"mode"`
	Name         string         `xml:"name"`
	MediaSize    int64          `xml:"mediasize"`
	SectorSize   int64          `xml:"sectorsize"`
	StripeSize   int64          `xml:"stripesize"`
	StripeOffset int64          `xml:"stripeoffset"`
	Config       geomConfig     `xml:"config"`
}

type geomDescriptor struct {
	ID        string       `xml:"id,attr,omitempty"`
	Name      string       `xml:"name,omitempty"`
	Rank      int          `xml:"rank,omitempty"`
	Config    geomConfig   `xml:"config,omitempty"`
	Consumers []geomTarget `xml:"consumer,omitempty"`
	Providers []geomTarget `xml:"provider,omitempty"`
}

type geomClass struct {
	ID   string         `xml:"id,attr"`
	Name string         `xml:"name"`
	GEOM geomDescriptor `xml:"geom"`
}

type geomXml struct {
	XMLName xml.Name    `xml:"mesh"`
	Classes []geomClass `xml:"class"`
}

type BlockDevices struct {
	BlockDeviceRoot string
}

func (self BlockDevices) Collect() map[string]interface{} {
	out := make(map[string]interface{})
	devid := 0

	var geom geomXml

	if data, err := exec.Command(`sysctl`, `-n`, `kern.geom.confxml`).Output(); err == nil {
		if err := xml.Unmarshal(data, &geom); err == nil {
			log.Dump(geom)

			for _, cls := range geom.Classes {
				switch strings.ToUpper(cls.Name) {
				case `DISK`:
					for k, v := range self.collectDevice(&cls) {
						out[fmt.Sprintf("disk.block.%d.%s", devid, k)] = v
					}

					devid += 1
				}
			}
		} else {
			log.Warningf("Failed to parse XML: %v", err)
		}
	} else {
		log.Warningf("Failed to retrieve XML: %v", err)
	}

	return out
}

func (self BlockDevices) collectDevice(cls *geomClass) map[string]interface{} {
	var provider geomTarget

	if len(cls.GEOM.Providers) > 0 {
		provider = cls.GEOM.Providers[0]
	} else {
		return nil
	}

	return map[string]interface{}{
		`name`:               cls.Name,
		`device`:             fmt.Sprintf("/dev/%s", cls.Name),
		`size`:               provider.MediaSize,
		`removable`:          false,
		`ssd`:                (provider.Config.RotationRate == 0),
		`rotation_rate`:      provider.Config.RotationRate,
		`wwid`:               provider.Config.WWID,
		`model`:              strings.TrimSpace(provider.Config.Description),
		`serial`:             strings.TrimSpace(provider.Config.Identifier),
		`blocksize.physical`: provider.StripeSize,
		`blocksize.logical`:  provider.SectorSize,
	}
}
