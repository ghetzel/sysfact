# sysfact

A utility for collecting and formatting system information.

## Overview

The `sysfact` utility provides a simple, structured mechanism for collecting and formatting pieces of information about Linux, BSD, macOS, and other operating systems. It achieves this using _collection scripts_, which run in parallel when `sysfact` is invoked to gather details. Collection scripts can be written in any language, as they are called via a shell process-- they only need to be executable to run. Scripts emit details, one per line, to standard output with a simple format: `metricname:type:value\n`

## Installation

### Binaries

Soon. _Soon..._

### From Source

```
go get -u github.com/ghetzel/sysfact/cmd/sysfact
```

## Formatting

For example, here is a some output that would be usable by `sysfact`:

```
demo.things.enabled:bool:true
demo.items.0.name:str:Item 1: Things
demo.items.0.count:int:42
demo.items.0.factor:float:3.14
demo.items.1.name:str:Item 2: Stuff
demo.items.1.count:int:99
demo.items.1.factor:float:2.178
demo.items.1.active:bool:true
```

This is what the above would look like using the "json" format option:

```json
{
	"demo": {
		"items": [
			{
				"count": 42,
				"factor": 3.14,
				"name": "Item 1: Things"
			},
			{
				"active": true,
				"count": 99,
				"factor": 2.178,
				"name": "Item 2: Stuff"
			}
		],
		"things": {
			"enabled": true
		}
	}
}
```

Output can be formatted into JSON, YAML, as a flat list of "key=value" items, or in several popular timeseries string formats; like InfluxDB's [Line Protocol](https://docs.influxdata.com/influxdb/latest/write_protocols/line_protocol_tutorial/), OpenTSDB/KairosDB, and Graphite.

## Output

Finally, the results of formatting the collected data can be output, either via standard output (the default), or using the built-in HTTP client. You can specify a URL, HTTP method, request headers, query string parameters (as command line flags), and timeouts. This provides a flexible and powerful mechanism for using `sysfact` as a component in monitoring, inventory collection, and other data collection systems.

# File Management & Templating

Sysfact also has a built-in mechanism for copying directory trees from one location to another, while also allowing for individual files and filenames to make use of data gathered by a Sysfact report. This is a simple but powerful mechanism for providing very bare-bones management of files on a system.

For example, lets take the following directory tree:

```
~/.config/sysfact/apply/
├── any
│   └── any
│       ├── .bashrc
│       ├── .config
│       │   └── gtk-3.0
│       │       └── settings.ini
│       └── .mpd
│           └── mpd.conf
└── ubuntu
    └── x86_64
        ├── @.bashrc
        ├── .config
        │   └── systemd
        │       └── user
        │           ├── default.target.wants
        │           │   └── sysfact-apply.service -> ../sysfact-apply.service
        │           ├── sysfact-apply.service
        │           ├── sysfact-apply.timer
        │           └── timer.target.wants
        │               └── sysfact-apply.timer -> ../sysfact-apply.timer
        └── @.sysfact-[[os.distribution]]-[[os.version]].json
```

By default, these files and directories represent files that you want to copy to their respective locations relative to your home directory (`~`). When you run `sysfact apply`, each file in this tree will be visited and copied into your home directory. Intermediate directories will be created, and the files will be placed in them.

## Using Templates to Generate Files

Note the `@` symbol in the filename `@.bashrc`. This tells Sysfact that the file should be treated as a template. The template engine will pass the complete Sysfact Report as input, along with any additional values you provide, and generate a text file as output with the values interopolated in. The resulting filename will have the leading `@` omitted. In this example, the rendered template will be placed at `~/.bashrc`.

Templates are standard [Golang text/template](https://golang.org/pkg/text/template/#pkg-overview) files, with additional functions provided by the [Diecast Standard Function Library](https://ghetzel.github.io/diecast/#funcref).

## Environment-Specific Overrides

Note that there are several roots in play here. The bulk of the files reside in `./any/any/`, but some of them are in the `./ubuntu/x86_64/` directory. Sysfact allows you to specify files and directories that should only be copied to the destination directory under specific OS, Distribution, Archicture, or OS Family combinations. Below is a list of the default order used to determine which files are copied over. You may also provide additional search patterns that will be appended to this default list. Patterns may use any value that appears in the `sysfact` report.

    - `<srcdir>/any/any/`
    - `<srcdir>/any/${arch}/`
    - `<srcdir>/${os.platform}/any/`
    - `<srcdir>/${os.platform}/${arch}/`
    - `<srcdir>/${os.family}/any/`
    - `<srcdir>/${os.family}/${arch}/`
    - `<srcdir>/${os.distribution}/any/`
    - `<srcdir>/${os.distribution}/${arch}/`
    - `<srcdir>/${os.distribution}-${os.version}/any/`
    - `<srcdir>/${os.distribution}-${os.version}/${arch}/`
    - `<srcdir>/${domain}/`
    - `<srcdir>/${hostname}/`
    - `<srcdir>/${hostname}.${domain}/`
    - `<srcdir>/${fqdn}/`
    - `<srcdir>/${uuid}/`

It is entirely possible to have the same filenames reside in multiple roots, with more-specific ones overwriting less specific versions. For example, the file `./any/any/.bashrc` would be copied to `~/.bashrc` first. If `sysfact` is being run on a 64-bit Ubuntu installation (any version), the `./ubuntu/x86_64/@.bashrc` file will be read, rendered as a template (because of the leading `@`), and overwrite the `~/.bashrc` file that was copied from before. This simple but powerful mechanism allows for very flexible file structures to be created that adapt to the needs of the system being configured.

## Templated Filenames

File and directory names themselves can also be templated by specifying values surrounded by double square brackets (`[[` and `]]`). The values inside the square brackets work the same way as the patterns mentioned above, including fallback values and `%`-formatting directives.

Take a look at the source file `./ubuntu/x86_64/@.sysfact-[[os.distribution]]-[[os.version]].json`. There's a fair bit going on here:

- This file will only be created on 64-bit Ubuntu hosts
- It will be treated as a template.
- The filename itself will _also_ be expanded. If run on a 64-bit Ubuntu 18.04 machine, the resulting file will be placed at `~/.sysfact-ubuntu-18.04.json`.

## Default Paths

When `sysfact` is run as a normal (i.e.: non-root) user, the default source and destination paths are:

- **srcdir**: `~/.config/sysfact/apply/`
- **destdir**: `~/`

When run as root, `sysfact` defaults to using:

- **srcdir**: `/etc/sysfact/apply/`
- **destdir**: `/`

This allows `sysfact` to be used by both regular users to manage their home directories, or by administrators to configure entire systems.
