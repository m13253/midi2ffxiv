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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func (app *application) parseConfigFile() error {
	f, err := os.Open(app.ConfigFile)
	if err != nil {
		log.Printf("Unable to load %s, default settings applied.\n", app.ConfigFile)
		return nil
	}
	defer f.Close()
	buf := bufio.NewReader(f)
	for {
		line, lineerr := buf.ReadString('\n')
		if strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			if lineerr == io.EOF {
				break
			}
			continue
		}
		// I do not want to use reflect.Value, they are too ugly
		switch fields[0] {
		case "IdleDuration":
			err = app.parseConfigDuration(fields, &app.IdleDuration)
		case "MaxNoteDelay":
			err = app.parseConfigDuration(fields, &app.MaxNoteDelay)
		case "SkillCooldown":
			err = app.parseConfigDuration(fields, &app.SkillCooldown)
		case "ModifierCooldown":
			err = app.parseConfigDuration(fields, &app.ModifierCooldown)
		case "NtpSyncTimeout":
			err = app.parseConfigDuration(fields, &app.NtpSyncTimeout)
		case "NtpCooldown":
			err = app.parseConfigDuration(fields, &app.NtpCooldown)
		case "MinTriggerVelocity":
			err = app.parseConfigUint8(fields, &app.MinTriggerVelocity)
		case "Keybinding":
			err = app.parseConfigKeybinding(fields, &app.Keybinding)
		case "WebListenAddr":
			err = app.parseConfigString(fields, &app.WebListenAddr)
		case "WebUsername":
			err = app.parseConfigString(fields, &app.WebUsername)
		case "WebPassword":
			err = app.parseConfigString(fields, &app.WebPassword)
		default:
			err = fmt.Errorf("unrecognized option %q", fields[0])
		}
		if err != nil {
			return err
		}
		if lineerr == io.EOF {
			break
		}
	}
	return nil
}

func (app *application) parseConfigDuration(fields []string, dest *time.Duration) error {
	if len(fields) != 2 {
		return fmt.Errorf("syntax error in option %q", fields[0])
	}
	duration, err := time.ParseDuration(fields[1])
	if err != nil {
		return err
	}
	*dest = duration
	return nil
}

func (app *application) parseConfigUint8(fields []string, dest *uint8) error {
	if len(fields) != 2 {
		return fmt.Errorf("syntax error in option %q", fields[0])
	}
	value, err := strconv.ParseUint(fields[1], 0, 8)
	if err != nil {
		return err
	}
	*dest = uint8(value)
	return nil
}

func (app *application) parseConfigString(fields []string, dest *string) error {
	if len(fields) > 2 {
		return fmt.Errorf("space is not allowed in option %q", fields[0])
	}
	if len(fields) == 1 {
		*dest = ""
	} else {
		*dest = fields[1]
	}
	return nil
}

func (app *application) parseConfigKeybinding(fields []string, dest *[128]keybindingPreset) error {
	if len(fields) < 3 {
		return fmt.Errorf("syntax error in option %q", fields[0])
	}
	noteIndex, err := noteNameToIndex(fields[1])
	if err != nil {
		return err
	}
	keybind := keybindingPreset{}
	virtualKeyCode := fields[len(fields)-1]
	if len(virtualKeyCode) == 3 && virtualKeyCode[0] == '\'' && virtualKeyCode[2] == '\'' {
		keybind.VirtualKeyCode = bytes.ToUpper([]byte{virtualKeyCode[1]})[0]
	} else {
		value, err := strconv.ParseUint(virtualKeyCode, 0, 8)
		if err != nil {
			return err
		}
		keybind.VirtualKeyCode = uint8(value)
	}
	for _, i := range fields[2 : len(fields)-1] {
		if strings.EqualFold(i, "Ctrl") {
			keybind.Ctrl = true
		} else if strings.EqualFold(i, "Alt") {
			keybind.Alt = true
		} else if strings.EqualFold(i, "Shift") {
			keybind.Shift = true
		} else {
			return fmt.Errorf("unrecognized modifier %q", i)
		}
	}
	dest[noteIndex] = keybind
	return nil
}

var (
	noteIndexToNameTable = [128]string{
		"C-1", "C#-1", "D-1", "Eb-1", "E-1", "F-1", "F#-1", "G-1", "Ab-1", "A-1", "Bb-1", "B-1", "C0", "C#0", "D0", "Eb0", "E0", "F0", "F#0", "G0", "Ab0", "A0", "Bb0", "B0", "C1", "C#1", "D1", "Eb1", "E1", "F1", "F#1", "G1", "Ab1", "A1", "Bb1", "B1", "C2", "C#2", "D2", "Eb2", "E2", "F2", "F#2", "G2", "Ab2", "A2", "Bb2", "B2", "C3", "C#3", "D3", "Eb3", "E3", "F3", "F#3", "G3", "Ab3", "A3", "Bb3", "B3", "C4", "C#4", "D4", "Eb4", "E4", "F4", "F#4", "G4", "Ab4", "A4", "Bb4", "B4", "C5", "C#5", "D5", "Eb5", "E5", "F5", "F#5", "G5", "Ab5", "A5", "Bb5", "B5", "C6", "C#6", "D6", "Eb6", "E6", "F6", "F#6", "G6", "Ab6", "A6", "Bb6", "B6", "C7", "C#7", "D7", "Eb7", "E7", "F7", "F#7", "G7", "Ab7", "A7", "Bb7", "B7", "C8", "C#8", "D8", "Eb8", "E8", "F8", "F#8", "G8", "Ab8", "A8", "Bb8", "B8", "C9", "C#9", "D9", "Eb9", "E9", "F9", "F#9", "G9",
	}
	noteNameToIndexTable = map[string]uint8{
		"C-1": 0x00, "C#-1": 0x01, "Db-1": 0x01, "D-1": 0x02, "D#-1": 0x03, "Eb-1": 0x03, "E-1": 0x04, "F-1": 0x05, "F#-1": 0x06, "Gb-1": 0x06, "G-1": 0x07, "G#-1": 0x08, "Ab-1": 0x08, "A-1": 0x09, "A#-1": 0x0a, "Bb-1": 0x0a, "B-1": 0x0b, "C0": 0x0c, "C#0": 0x0d, "Db0": 0x0d, "D0": 0x0e, "D#0": 0x0f, "Eb0": 0x0f, "E0": 0x10, "F0": 0x11, "F#0": 0x12, "Gb0": 0x12, "G0": 0x13, "G#0": 0x14, "Ab0": 0x14, "A0": 0x15, "A#0": 0x16, "Bb0": 0x16, "B0": 0x17, "C1": 0x18, "C#1": 0x19, "Db1": 0x19, "D1": 0x1a, "D#1": 0x1b, "Eb1": 0x1b, "E1": 0x1c, "F1": 0x1d, "F#1": 0x1e, "Gb1": 0x1e, "G1": 0x1f, "G#1": 0x20, "Ab1": 0x20, "A1": 0x21, "A#1": 0x22, "Bb1": 0x22, "B1": 0x23, "C2": 0x24, "C#2": 0x25, "Db2": 0x25, "D2": 0x26, "D#2": 0x27, "Eb2": 0x27, "E2": 0x28, "F2": 0x29, "F#2": 0x2a, "Gb2": 0x2a, "G2": 0x2b, "G#2": 0x2c, "Ab2": 0x2c, "A2": 0x2d, "A#2": 0x2e, "Bb2": 0x2e, "B2": 0x2f, "C3": 0x30, "C#3": 0x31, "Db3": 0x31, "D3": 0x32, "D#3": 0x33, "Eb3": 0x33, "E3": 0x34, "F3": 0x35, "F#3": 0x36, "Gb3": 0x36, "G3": 0x37, "G#3": 0x38, "Ab3": 0x38, "A3": 0x39, "A#3": 0x3a, "Bb3": 0x3a, "B3": 0x3b, "C4": 0x3c, "C#4": 0x3d, "Db4": 0x3d, "D4": 0x3e, "D#4": 0x3f, "Eb4": 0x3f, "E4": 0x40, "F4": 0x41, "F#4": 0x42, "Gb4": 0x42, "G4": 0x43, "G#4": 0x44, "Ab4": 0x44, "A4": 0x45, "A#4": 0x46, "Bb4": 0x46, "B4": 0x47, "C5": 0x48, "C#5": 0x49, "Db5": 0x49, "D5": 0x4a, "D#5": 0x4b, "Eb5": 0x4b, "E5": 0x4c, "F5": 0x4d, "F#5": 0x4e, "Gb5": 0x4e, "G5": 0x4f, "G#5": 0x50, "Ab5": 0x50, "A5": 0x51, "A#5": 0x52, "Bb5": 0x52, "B5": 0x53, "C6": 0x54, "C#6": 0x55, "Db6": 0x55, "D6": 0x56, "D#6": 0x57, "Eb6": 0x57, "E6": 0x58, "F6": 0x59, "F#6": 0x5a, "Gb6": 0x5a, "G6": 0x5b, "G#6": 0x5c, "Ab6": 0x5c, "A6": 0x5d, "A#6": 0x5e, "Bb6": 0x5e, "B6": 0x5f, "C7": 0x60, "C#7": 0x61, "Db7": 0x61, "D7": 0x62, "D#7": 0x63, "Eb7": 0x63, "E7": 0x64, "F7": 0x65, "F#7": 0x66, "Gb7": 0x66, "G7": 0x67, "G#7": 0x68, "Ab7": 0x68, "A7": 0x69, "A#7": 0x6a, "Bb7": 0x6a, "B7": 0x6b, "C8": 0x6c, "C#8": 0x6d, "Db8": 0x6d, "D8": 0x6e, "D#8": 0x6f, "Eb8": 0x6f, "E8": 0x70, "F8": 0x71, "F#8": 0x72, "Gb8": 0x72, "G8": 0x73, "G#8": 0x74, "Ab8": 0x74, "A8": 0x75, "A#8": 0x76, "Bb8": 0x76, "B8": 0x77, "C9": 0x78, "C#9": 0x79, "Db9": 0x79, "D9": 0x7a, "D#9": 0x7b, "Eb9": 0x7b, "E9": 0x7c, "F9": 0x7d, "F#9": 0x7e, "Gb9": 0x7e, "G9": 0x7f,
	}
)

func noteIndexToName(index uint8) (string, error) {
	if index > 0x7f {
		return fmt.Sprintf("0x%02x", index), fmt.Errorf("note 0x%02x out of range", index)
	}
	return noteIndexToNameTable[index], nil
}

func noteNameToIndex(name string) (uint8, error) {
	if index, ok := noteNameToIndexTable[name]; ok {
		return index, nil
	}
	if index, err := strconv.ParseUint(name, 0, 7); err == nil {
		return uint8(index), nil
	}
	return 0xff, fmt.Errorf("unrecognized note name %q", name)
}
