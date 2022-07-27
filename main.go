package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fatih/color"
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

	timer := time.NewTimer(dur)
	ticker := time.NewTicker(time.Second)
	start := time.Now()

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

func printRemaining(dur time.Duration, start time.Time) {
	passed := time.Since(start)
	if color.NoColor {
		fmt.Printf("Remaining: %s\n", (dur - passed).Round(time.Second))
	} else {
		fmt.Printf("\r\x1B[0K%s Remaining: %s",
			colorCmdName.Sprint("countdown:"),
			colorDuration.Sprint((dur - passed).Round(time.Second)),
		)
	}
}

func printDone(dur time.Duration) {
	if color.NoColor {
		fmt.Printf("Done waiting for %s\n", dur)
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
