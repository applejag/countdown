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

// Package fuzzytime contains helper functions to parse times in a more fuzzy
// and natural manner, compared to the built-in time.Parse which only attempts
// a single format.
package fuzzytime

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/common"
	"github.com/olebedev/when/rules/en"
)

var (
	// ErrUnknownFormat is returned when no time layout format was matched.
	ErrUnknownFormat = errors.New("unknown time format")
	// ErrTimeAlreadyPassed is returned when the resulting time is in the past.
	ErrTimeAlreadyPassed = errors.New("time already passed")
	// ErrDurationNegative is returned by Parse when duration was negative.
	ErrDurationNegative = fmt.Errorf("%w: duration cannot be negative", ErrTimeAlreadyPassed)
)

var w *when.Parser

func init() {
	w = when.New(nil)
	w.Add(en.All...)
	w.Add(common.All...)
	w.Add(optionalHourRule{})
	w.Add(onlyHourRule{})
}

// ParseFuture a duration, a list of known formats, and lastly via the `when`
// fuzzy parsing package, and returns the time on the first match it finds.
//
// The result will be adjusted to try fit into the future, otherwise an error
// is returned.
func ParseFuture(s string, base time.Time) (time.Time, error) {
	if strings.EqualFold(s, "now") {
		return time.Now(), nil
	}
	if t, ok := ParseDelta(s, base); ok {
		if t.Before(base) {
			return time.Time{}, ErrDurationNegative
		}
		return t, nil
	}
	if t, err := ParseKnownLayouts(s); err == nil {
		if t.Before(base) {
			return time.Time{}, fmt.Errorf("%w: %s", ErrTimeAlreadyPassed, t.Format(time.Stamp))
		}
		return t, nil
	}
	return parseWhenFuture(s, base)
}

func parseWhenFuture(s string, base time.Time) (time.Time, error) {
	d, err := parseWhenDiff(s, base)
	if err != nil {
		return time.Time{}, err
	}
	if d.t.After(base) {
		return d.t, nil
	}
	// At this point, the parsed time is thought to be in the past.
	// Can we assume the future?
	switch {
	case d.hasY && d.hasM && d.hasD:
		return time.Time{}, fmt.Errorf("%w: must be today or future day: %d-%d-%d", ErrTimeAlreadyPassed, d.y, d.m, d.d)
	case !d.hasM && !d.hasD && d.hasH:
		fmt.Println("Woa", d.t, "     ", d.t.Add(24*time.Hour))
		return d.t.Add(24 * time.Hour), nil
	default:
		return time.Time{}, fmt.Errorf("%w: %s", ErrTimeAlreadyPassed, d.t.Format(time.Stamp))
	}
}

type whenDiff struct {
	t                                           time.Time
	y, d, h, min, s, ns                         int
	m                                           time.Month
	hasY, hasM, hasD, hasH, hasMin, hasS, hasNs bool
}

func parseWhenDiff(s string, base time.Time) (whenDiff, error) {
	t1, err := ParseWhen(s, base)
	if err != nil {
		return whenDiff{}, err
	}
	t1y, t1m, t1d := t1.Date()
	t1h, t1min, t1s := t1.Clock()
	t1ns := t1.Nanosecond()
	t2, err := ParseWhen(s, time.Date(
		t1y^1, t1m^1, t1d^1,
		t1h^1, t1min^1, t1s^1, t1ns^1,
		t1.Location()))
	if err != nil {
		return whenDiff{}, err
	}
	t2y, t2m, t2d := t2.Date()
	t2h, t2min, t2s := t2.Clock()
	t2ns := t2.Nanosecond()
	return whenDiff{
		t: t1,
		y: t1y, hasY: t1y == t2y,
		m: t1m, hasM: t1m == t2m,
		d: t1d, hasD: t1d == t2d,
		h: t1h, hasH: t1h == t2h,
		min: t1min, hasMin: t1min == t2min,
		s: t1s, hasS: t1s == t2s,
		ns: t1ns, hasNs: t1ns == t2ns,
	}, nil
}

var knownLayouts = []string{
	time.RFC3339,
	time.RFC3339Nano,
	time.RFC822,
	time.RFC822Z,
	time.RFC850,
	time.RFC1123,
	time.RFC1123Z,
	time.Stamp,
}

// ParseKnownLayouts attempts to parse the string according to the date formats
// defined in the IETF RFC822, RFC580, RFC1123, or RFC3339.
func ParseKnownLayouts(s string) (time.Time, error) {
	for _, layout := range knownLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, ErrUnknownFormat
}

// ParseWhen performs a fuzzy time parsing via the `when` package.
func ParseWhen(s string, base time.Time) (time.Time, error) {
	r, err := w.Parse(s, base.Truncate(time.Second))
	if err != nil {
		return time.Time{}, err
	}
	if r == nil {
		return time.Time{}, ErrUnknownFormat
	}
	return r.Time, nil
}

// ParseDelta attempts to parse the string as a time.Duration and adds that to
// the current time.
func ParseDelta(s string, base time.Time) (time.Time, bool) {
	if len(s) < 2 {
		return time.Time{}, false
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return time.Time{}, false
	}
	return base.Add(d), true
}
