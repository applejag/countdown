package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

func init() {
	pflag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: countdown <duration>")
	}
}

func main() {
	pflag.Parse()

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

	for {
		select {
		case <-ticker.C:
			passed := time.Since(start)
			fmt.Printf("\rRemaining: %s", (dur - passed).Truncate(time.Second))
		case <-timer.C:
			ticker.Stop()
			fmt.Printf("\rDone waiting for %s\n", dur)
			return
		}
	}
}

func printErr(err error) {
	fmt.Printf("countdown: err: %s\n", err)
}
