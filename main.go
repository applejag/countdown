package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

var flags = struct {
	noNotify bool
	showHelp bool
}{}

func init() {
	pflag.BoolVar(&flags.noNotify, "no-notify", false, "Disables notification via notify-send")
	pflag.BoolVarP(&flags.showHelp, "help", "h", false, "Show this help text")
	pflag.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: countdown <duration>

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

	args := pflag.Args()
	if len(args) == 0 {
		printErr(errors.New("missing argument"))
		pflag.Usage()
		os.Exit(1)
	}
	if len(args) > 1 {
		printErr(errors.New("too many arguments"))
		pflag.Usage()
		os.Exit(1)
	}

	arg := strings.TrimSpace(args[0])
	dur, err := time.ParseDuration(arg)
	if err != nil {
		printErr(fmt.Errorf("parse argument: %w", err))
		pflag.Usage()
		os.Exit(1)
	}

	ticker := time.NewTicker(time.Second)
	timer := time.NewTimer(dur)
	start := time.Now()

	printRemaining(dur, start)
	for {
		select {
		case <-ticker.C:
			printRemaining(dur, start)
		case <-timer.C:
			ticker.Stop()
			fmt.Printf("\rDone waiting for %s\n", dur)
			if !flags.noNotify {
				sendNotification(dur)
			}
			return
		}
	}
}

func printRemaining(dur time.Duration, start time.Time) {
	passed := time.Since(start)
	fmt.Printf("\r\x1B[0KRemaining: %s", (dur - passed).Round(time.Second))
}

func printErr(err error) {
	fmt.Printf("countdown: err: %s\n", err)
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
