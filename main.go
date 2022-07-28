// SPDX-FileCopyrightText: 2022 Kalle Fagerberg
//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/jilleJr/countdown/internal/fuzzytime"
	"github.com/mattn/go-colorable"
	"github.com/spf13/pflag"
)

var flags = struct {
	noNotify bool
	color    string
	showHelp bool
}{
	color: "auto",
}

var (
	stdout = colorable.NewColorableStdout()
	stderr = colorable.NewColorableStderr()

	colorCmdName   = color.New(color.FgHiBlack)
	colorErrPrefix = color.New(color.FgHiRed, color.Bold)
	colorErr       = color.New(color.FgRed)
	colorDuration  = color.New(color.FgHiMagenta, color.Bold)
	colorDone      = color.New(color.FgGreen)
)

func init() {
	pflag.StringVar(&flags.color, "color", flags.color, `Colored output, either "always", "never", or "auto"`)
	pflag.BoolVar(&flags.noNotify, "no-notify", false, "Disables notification via notify-send")
	pflag.BoolVarP(&flags.showHelp, "help", "h", false, "Show this help text")
	pflag.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: countdown <duration>

The <duration> argument is a Go time.Duration formatted string.
Examples:

  countdown 10s        // 10 seconds
  countdown 10m        // 10 minutes
  countdown 1m30s      // 1 minute and 30 seconds
  countdown 1h20m30s   // 1 hour, 20 minutes, and 30 seconds

  countdown 12:00      // countdown to the next 12 o'clock

Flags:
`)
		pflag.PrintDefaults()
	}
}

func main() {
	pflag.Parse()

	if flags.showHelp {
		pflag.Usage()
		os.Exit(0)
	}

	flags.color = strings.ToLower(flags.color)
	switch flags.color {
	case "auto":
	// Do nothing
	case "never":
		color.NoColor = true
	case "always":
		color.NoColor = false
	default:
		printErr(errors.New(`invalid --color value, must be one of "always", "never", or "auto"`))
		os.Exit(1)
	}

	args := pflag.Args()
	if len(args) == 0 {
		pflag.Usage()
		os.Exit(0)
	}
	if len(args) > 1 {
		printErr(errors.New("too many arguments"))
		pflag.Usage()
		os.Exit(1)
	}

	end, err := fuzzytime.ParseFuture(strings.TrimSpace(args[0]), time.Now())
	if err != nil {
		printErr(fmt.Errorf("parse time or duration: %w", err))
		os.Exit(1)
	}

	start := time.Now()
	dur := end.Sub(start).Round(time.Millisecond)
	timer := time.NewTimer(dur)
	ticker := time.NewTicker(time.Second)

	printEnd(end)
	printRemaining(dur, start)
	for {
		select {
		case <-timer.C:
			ticker.Stop()
			printDone(dur)
			if !flags.noNotify {
				sendNotification(dur)
			}
			return
		case <-ticker.C:
			printRemaining(dur, start)
		}
	}
}

func printEnd(end time.Time) {
	if color.NoColor {
		fmt.Printf("countdown: Waiting for: %s\n", end.Format(time.Stamp))
	} else {
		fmt.Printf("%s Waiting for: %s\n",
			colorCmdName.Sprint("countdown:"),
			colorDuration.Sprint(end.Format(time.Stamp)),
		)
	}
}

func printRemaining(dur time.Duration, start time.Time) {
	passed := time.Since(start)
	if color.NoColor {
		fmt.Printf("countdown: Remaining:   %s\n", (dur - passed).Round(time.Second))
	} else {
		fmt.Printf("\r\x1B[0K%s Remaining:   %s",
			colorCmdName.Sprint("countdown:"),
			colorDuration.Sprint((dur - passed).Round(time.Second)),
		)
	}
}

func printDone(dur time.Duration) {
	if color.NoColor {
		fmt.Printf("countdown: Done waiting for %s\n", dur)
	} else {
		fmt.Printf("\r\x1B[0K%s %s %s\n",
			colorCmdName.Sprint("countdown:"),
			colorDone.Sprint("Done waiting for:"),
			colorDuration.Sprint(dur.Round(time.Second)),
		)
	}
}

func printErr(err error) {
	if color.NoColor {
		fmt.Fprintf(os.Stderr, "countdown: err: %s\n", err)
	} else {
		fmt.Fprintf(stderr, "%s %s %s\n",
			colorCmdName.Sprint("countdown:"),
			colorErrPrefix.Sprint("err:"),
			colorErr.Sprint(err),
		)
	}
}

func sendNotification(dur time.Duration) {
	exe, err := exec.LookPath("notify-send")
	if errors.Is(err, exec.ErrNotFound) {
		printErr(err)
		return
	} else if err != nil {
		printErr(fmt.Errorf("find notify-send: %w", err))
		return
	}

	if err := exec.Command(exe,
		"--urgency=critical",
		"--app-name=countdown",
		"Countdown expired!",
		fmt.Sprintf("Done counting down from %s", dur),
	).Run(); err != nil {
		printErr(fmt.Errorf("run notify-send: %w", err))
	}
}
