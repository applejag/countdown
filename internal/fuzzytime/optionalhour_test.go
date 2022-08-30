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
	"testing"
	"time"
)

func TestMatchOptionalHour(t *testing.T) {
	tests := []struct {
		name  string
		str   string
		wantH int
		wantM int
		wantS int
	}{
		{
			name:  "empty str",
			str:   "",
			wantH: -1,
			wantM: -1,
			wantS: -1,
		},
		{
			name:  "only minute short",
			str:   ":40",
			wantH: -1,
			wantM: 40,
			wantS: -1,
		},
		{
			name:  "only minute with delim",
			str:   ":40:",
			wantH: -1,
			wantM: 40,
			wantS: -1,
		},
		{
			name:  "only second",
			str:   "::34",
			wantH: -1,
			wantM: -1,
			wantS: 34,
		},
		{
			name:  "only hour short",
			str:   "10:",
			wantH: 10,
			wantM: -1,
			wantS: -1,
		},
		{
			name:  "only hour with sec delim",
			str:   "10::",
			wantH: 10,
			wantM: -1,
			wantS: -1,
		},
		{
			name:  "all 3",
			str:   "10:40:34",
			wantH: 10,
			wantM: 40,
			wantS: 34,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h, m, s := matchOptionalTimes(tc.str)
			if h != tc.wantH {
				t.Errorf("want h=%d, got %d", tc.wantH, h)
			}
			if m != tc.wantM {
				t.Errorf("want m=%d, got %d", tc.wantM, m)
			}
			if s != tc.wantS {
				t.Errorf("want s=%d, got %d", tc.wantS, s)
			}
		})
	}
}

func TestFoo(t *testing.T) {
	ParseFuture(":01", time.Now())
}
