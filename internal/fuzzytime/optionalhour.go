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

package fuzzytime

import (
	"regexp"
	"strconv"
	"time"

	"github.com/olebedev/when/rules"
)

var _ rules.Rule = optionalHourRule{}

type optionalHourRule struct{}

var optionalHourRegex = regexp.MustCompile(`^(\d*):(\d*):?(\d*)$`)

func (optionalHourRule) Find(str string) *rules.Match {
	h, m, s := matchOptionalTimes(str)
	if h == -1 && m == -1 && s == -1 {
		return nil
	}
	return &rules.Match{
		Text:     str,
		Captures: []string{strconv.Itoa(h), strconv.Itoa(m), strconv.Itoa(s)},
		Applier:  applyHourMinSec,
	}
}

func applyHourMinSec(m *rules.Match, ctx *rules.Context, o *rules.Options, t time.Time) (bool, error) {
	hour, _ := strconv.Atoi(m.Captures[0])
	min, _ := strconv.Atoi(m.Captures[1])
	sec, _ := strconv.Atoi(m.Captures[2])

	ctx.Hour = minusOneAsNil(hour)
	ctx.Minute = minusOneAsZero(min)
	ctx.Second = minusOneAsZero(sec)

	if newTime, err := ctx.Time(t); err == nil && newTime.Before(t) {
		newHour := t.Hour() + 1
		ctx.Hour = &newHour
	}
	return true, nil
}

func minusOneAsNil(n int) *int {
	if n == -1 {
		return nil
	}
	return &n
}

func minusOneAsZero(n int) *int {
	if n == -1 {
		zero := 0
		return &zero
	}
	return &n
}

func matchOptionalTimes(s string) (hour, min, sec int) {
	matches := optionalHourRegex.FindAllStringSubmatch(s, 1)
	if len(matches) != 1 {
		return -1, -1, -1
	}
	groups := matches[0]
	hour, hourOK := tryParseUintWithLimit(groups[1], 24)
	if !hourOK {
		return -1, -1, -1
	}
	min, minOK := tryParseUintWithLimit(groups[2], 60)
	if !minOK {
		return -1, -1, -1
	}
	sec, secOK := tryParseUintWithLimit(groups[3], 60)
	if !secOK {
		return -1, -1, -1
	}
	return
}

func tryParseUintWithLimit(s string, limit int) (int, bool) {
	if s == "" {
		return -1, true
	}
	num, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return -1, false
	}
	if num > uint64(limit) {
		return -1, false
	}
	return int(num), true
}

//func applyMinute(m *rules.Match, ctx *rules.Context, o *rules.Options, t time.Time) (bool, error) {
//	min, _ := strconv.Atoi(m.Captures[0])
//	zero := 0
//	ctx.Hour = &zero
//	ctx.Minute = &min
//	ctx.Second = &zero
//	return true, nil
//}
