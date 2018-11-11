package data

import (
	"runtime"

	"github.com/ghetzel/go-stockutil/fileutil"
	"github.com/ghetzel/go-stockutil/pathutil"
	"github.com/ghetzel/go-stockutil/rxutil"
)

type OS struct {
}

func (self OS) Collect() map[string]interface{} {
	out := make(map[string]interface{})

	out[`os.platform`] = runtime.GOOS

	if pathutil.IsNonemptyFile(`/etc/redhat-release`) {
		out[`os.family`] = `redhat`

		if content, err := fileutil.ReadAllString(`/etc/redhat-release`); err == nil {
			if match := rxutil.Match(`^(?P<distribution>.*) release (?P<version>[0-9\.]+)`, content); match != nil {
				out[`os.distribution`] = match.Group(`distribution`)
				out[`os.version`] = match.Group(`version`)
			}
		}
	} else if pathutil.IsNonemptyExecutableFile(`/usr/bin/lsb_release`) {
		out[`os.family`] = `debian`
		out[`os.distribution`] = shell(`lsb_release --short --id`).String()
		out[`os.version`] = shell(`lsb_release --short --release`).String()
		out[`os.codename`] = shell(`lsb_release --short --codename`).String()
		out[`os.description`] = shell(`lsb_release --short --description`).String()
	} else {
		out[`os.family`] = `unknown`
	}

	return out
}
