// +build windows

/*
   MIDI2FFXIV-Realtime
   Copyright (C) 2017-2018 Star Brilliant <m13253@hotmail.com>

   Permission is hereby granted, free of charge, to any person obtaining a
   copy of this software and associated documentation files (the "Software"),
   to deal in the Software without restriction, including without limitation
   the rights to use, copy, modify, merge, publish, distribute, sublicense,
   and/or sell copies of the Software, and to permit persons to whom the
   Software is furnished to do so, subject to the following conditions:

   The above copyright notice and this permission notice shall be included in
   all copies or substantial portions of the Software.

   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
   FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
   DEALINGS IN THE SOFTWARE.
*/

package main

import (
	"time"
)

type keybindingPreset struct {
	VirtualKeyCode uint8
	Ctrl           bool
	Alt            bool
	Shift          bool
}

type preset struct {
	IdleDuration       time.Duration
	MaxNoteDelay       time.Duration
	SkillCooldown      time.Duration
	MinTriggerVelocity uint8
	Keybinding         [128]keybindingPreset
}

var defaultPreset = preset{
	IdleDuration:       1 * time.Second,
	MaxNoteDelay:       300 * time.Millisecond,
	SkillCooldown:      140 * time.Millisecond,
	MinTriggerVelocity: 16,
	Keybinding: [128]keybindingPreset{
		0x30: {'Q', true, false, false},
		0x31: {'2', true, false, false},
		0x32: {'W', true, false, false},
		0x33: {'3', true, false, false},
		0x34: {'E', true, false, false},
		0x35: {'R', true, false, false},
		0x36: {'5', true, false, false},
		0x37: {'T', true, false, false},
		0x38: {'6', true, false, false},
		0x39: {'Y', true, false, false},
		0x3a: {'7', true, false, false},
		0x3b: {'U', true, false, false},

		0x3c: {'Q', false, false, false},
		0x3d: {'2', false, false, false},
		0x3e: {'W', false, false, false},
		0x3f: {'3', false, false, false},
		0x40: {'E', false, false, false},
		0x41: {'R', false, false, false},
		0x42: {'5', false, false, false},
		0x43: {'T', false, false, false},
		0x44: {'6', false, false, false},
		0x45: {'Y', false, false, false},
		0x46: {'7', false, false, false},
		0x47: {'U', false, false, false},

		0x48: {'Q', false, false, true},
		0x49: {'2', false, false, true},
		0x4a: {'W', false, false, true},
		0x4b: {'3', false, false, true},
		0x4c: {'E', false, false, true},
		0x4d: {'R', false, false, true},
		0x4e: {'5', false, false, true},
		0x4f: {'T', false, false, true},
		0x50: {'6', false, false, true},
		0x51: {'Y', false, false, true},
		0x52: {'7', false, false, true},
		0x53: {'U', false, false, true},

		0x54: {'I', false, false, true},
	},
}
