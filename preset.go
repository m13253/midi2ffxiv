// +build windows

/*
   MIDI2FFXIV
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
	Ctrl           bool
	Alt            bool
	Shift          bool
	VirtualKeyCode uint8
}

type preset struct {
	ConfigFile string

	IdleDuration       time.Duration
	MaxNoteDelay       time.Duration
	SkillCooldown      time.Duration
	ModifierCooldown   time.Duration
	NtpSyncTimeout     time.Duration
	NtpCooldown        time.Duration
	MinTriggerVelocity uint8
	Keybinding         [128]keybindingPreset

	WebListenAddr string
	WebUsername   string
	WebPassword   string
}

var defaultPreset = preset{
	ConfigFile:         "midi2ffxiv.conf",
	IdleDuration:       1 * time.Second,
	MaxNoteDelay:       300 * time.Millisecond,
	SkillCooldown:      140 * time.Millisecond,
	ModifierCooldown:   50 * time.Millisecond,
	NtpSyncTimeout:     5 * time.Second,
	NtpCooldown:        10 * time.Second,
	MinTriggerVelocity: 16,
	Keybinding: [128]keybindingPreset{
		0x30: {true, false, false, 'Q'},
		0x31: {true, false, false, '2'},
		0x32: {true, false, false, 'W'},
		0x33: {true, false, false, '3'},
		0x34: {true, false, false, 'E'},
		0x35: {true, false, false, 'R'},
		0x36: {true, false, false, '5'},
		0x37: {true, false, false, 'T'},
		0x38: {true, false, false, '6'},
		0x39: {true, false, false, 'Y'},
		0x3a: {true, false, false, '7'},
		0x3b: {true, false, false, 'U'},

		0x3c: {false, false, false, 'Q'},
		0x3d: {false, false, false, '2'},
		0x3e: {false, false, false, 'W'},
		0x3f: {false, false, false, '3'},
		0x40: {false, false, false, 'E'},
		0x41: {false, false, false, 'R'},
		0x42: {false, false, false, '5'},
		0x43: {false, false, false, 'T'},
		0x44: {false, false, false, '6'},
		0x45: {false, false, false, 'Y'},
		0x46: {false, false, false, '7'},
		0x47: {false, false, false, 'U'},

		0x48: {false, false, true, 'Q'},
		0x49: {false, false, true, '2'},
		0x4a: {false, false, true, 'W'},
		0x4b: {false, false, true, '3'},
		0x4c: {false, false, true, 'E'},
		0x4d: {false, false, true, 'R'},
		0x4e: {false, false, true, '5'},
		0x4f: {false, false, true, 'T'},
		0x50: {false, false, true, '6'},
		0x51: {false, false, true, 'Y'},
		0x52: {false, false, true, '7'},
		0x53: {false, false, true, 'U'},

		0x54: {false, false, true, 'I'},
	},
	WebListenAddr: ":65300",
	WebUsername:   "",
	WebPassword:   "",
}
