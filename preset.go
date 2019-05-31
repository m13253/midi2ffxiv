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
	RealtimeExtraDelay time.Duration
	PlaybackExtraDelay time.Duration
	RealtimeMaxLatency time.Duration
	PlaybackMaxLatency time.Duration
	SkillCooldown      time.Duration
	ModifierCooldown   time.Duration
	NtpSyncTimeout     time.Duration
	NtpCooldown        time.Duration
	MinTriggerVelocity uint8
	Keybinding         [128]keybindingPreset
	EmergencyStop      *keybindingPreset

	WebListenAddr string
	WebUsername   string
	WebPassword   string
}

var defaultPreset = preset{
	ConfigFile:         "midi2ffxiv.conf",
	IdleDuration:       1000 * time.Millisecond,
	RealtimeExtraDelay: 0 * time.Millisecond,
	PlaybackExtraDelay: 2000 * time.Millisecond,
	RealtimeMaxLatency: 300 * time.Millisecond,
	PlaybackMaxLatency: 300 * time.Millisecond,
	SkillCooldown:      125 * time.Millisecond,
	ModifierCooldown:   50 * time.Millisecond,
	NtpSyncTimeout:     5 * time.Second,
	NtpCooldown:        10 * time.Second,
	MinTriggerVelocity: 16,
	Keybinding: [128]keybindingPreset{
	    0x24: {false, false, false, 'Y'},
		0x25: {false, false, false, 'V'},
		0x26: {false, false, false, 'U'},
		0x27: {false, false, false, 'B'},
		0x28: {false, false, false, 'I'},
		0x29: {false, false, false, 'O'},
		0x2A: {false, false, false, 'N'},
		0x2B: {false, false, false, 'P'},
		0x2C: {false, false, false, 'M'},
		0x2D: {false, false, false, 'A'},
		0x2E: {false, false, false, ','},
		0x2F: {false, false, false, 'S'},

		0x30: {false, false, false, '9'},
		0x31: {false, false, false, 'K'},
		0x32: {false, false, false, 'O'},
		0x33: {false, false, false, 'L'},
		0x34: {false, false, false, 'Q'},
		0x35: {false, false, false, 'W'},
		0x36: {false, false, false, 'Z'},
		0x37: {false, false, false, 'E'},
		0x38: {false, false, false, 'X'},
		0x39: {false, false, false, 'R'},
		0x3a: {false, false, false, 'C'},
		0x3b: {false, false, false, 'T'},

		0x3c: {false, false, false, '1'},
		0x3d: {false, false, false, 'D'},
		0x3e: {false, false, false, '2'},
		0x3f: {false, false, false, 'F'},
		0x40: {false, false, false, '3'},
		0x41: {false, false, false, '4'},
		0x42: {false, false, false, 'G'},
		0x43: {false, false, false, '5'},
		0x44: {false, false, false, 'H'},
		0x45: {false, false, false, '6'},
		0x46: {false, false, false, 'J'},
		0x47: {false, false, false, '7'},

		0x48: {false, false, false, '8'},
	},
	EmergencyStop: &keybindingPreset{false, false, false, 0x1b},
	WebListenAddr: ":65300",
	WebUsername:   "",
	WebPassword:   "",
}
