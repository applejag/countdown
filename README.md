<!--
SPDX-FileCopyrightText: 2021 Kalle Fagerberg

SPDX-License-Identifier: CC0-1.0
-->

# countdown

[![REUSE status](https://api.reuse.software/badge/github.com/jilleJr/countdown)](https://api.reuse.software/info/github.com/jilleJr/countdown)

I needed a small countdown util. So I made one

Sends a notification when the time has expired.

## Install

Requires Go 1.18 (or later)

```sh
go install github.com/jilleJr/countdown@latest
```

## Usage

```console
$ countdown --help
Usage: countdown <duration>

The <duration> argument is a Go time.Duration formatted string.
Examples:

  countdown 10s        // 10 seconds
  countdown 10m        // 10 minutes
  countdown 1m30s      // 1 minute and 30 seconds
  countdown 1h20m30s   // 1 hour, 20 minutes, and 30 seconds

Flags:
      --color string   Colored output, either "always", "never", or "auto" (default "auto")
  -h, --help           Show this help text
      --no-notify      Disables notification via notify-send
```

## License

Written and maintained by [@jilleJr](https://github.com/jilleJr).
Code is licensed under the GNU GPL 3.0 or later,
with misc. documents licensed under CC0 1.0.

This repository is [REUSE](https://reuse.software/) compliant.
