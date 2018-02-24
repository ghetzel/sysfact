# sysfact

A utility for collecting and formatting system information.

## Overview

The `sysfact` utility provides a simple, structured mechanism for collecting and formatting pieces of information about Linux, BSD, macOS, and other operating systems.  It achieves this using _collection scripts_, which run in parallel when `sysfact` is invoked to gather details.  Collection scripts can be written in any language, as they are called via a shell process-- they only need to be executable to run.  Scripts emit details, one per line, to standard output with a simple format: `metricname:type:value\n`

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

Finally, the results of formatting the collected data can be output, either via standard output (the default), or using the built-in HTTP client.  You can specify a URL, HTTP method, request headers, query string parameters (as command line flags), and timeouts.  This provides a flexible and powerful mechanism for using `sysfact` as a component in monitoring, inventory collection, and other system data collection systems.
