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
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/olebedev/when/rules"
)

var _ rules.Rule = onlyHourRule{}

type onlyHourRule struct{}

func (onlyHourRule) Find(s string) *rules.Match {
	var h int
	var am string
	read, err := fmt.Sscanf(s, "%d%s", &h, &am)
	if errors.Is(err, io.EOF) && read == 1 {
		if h >= 0 && h <= 24 {
			return &rules.Match{
				Text:     s,
				Captures: []string{strconv.Itoa(h)},
				Applier:  applyHour,
			}
		}
	}
	if err != nil {
		return nil
	}
	if h < 0 {
		return nil
	}
	am = strings.ToLower(am)
	switch am {
	case "am", "pm":
		return &rules.Match{
			Text:     s,
			Captures: []string{strconv.Itoa(h), am},
			Applier:  applyHourAndAM,
		}
	default:
		return nil
	}
}

func applyHour(m *rules.Match, ctx *rules.Context, o *rules.Options, t time.Time) (bool, error) {
	h, _ := strconv.Atoi(m.Captures[0])
	zero := 0
	ctx.Hour = &h
	ctx.Minute = &zero
	ctx.Second = &zero
	return true, nil
}

func applyHourAndAM(m *rules.Match, ctx *rules.Context, o *rules.Options, t time.Time) (bool, error) {
	h, _ := strconv.Atoi(m.Captures[0])
	am := m.Captures[1]
	if h > 12 {
		return false, fmt.Errorf("an AM/PM time must have hour between 0-12, but got: %d", h)
	}
	if am == "pm" {
		h += 12
		if h == 24 {
			h = 0
		}
	}
	zero := 0
	ctx.Hour = &h
	ctx.Minute = &zero
	ctx.Second = &zero
	return true, nil
}
