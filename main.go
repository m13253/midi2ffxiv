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
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"./kernel32"
	"./user32"
	"./winmm"
)

type application struct {
	preset

	MidiInDevice      int
	MidiOutDevice     int
	MidiOutInstrument uint32
	MidiOutTranspose  int

	bQuitting   bool
	hWnd        uintptr
	hMidiIn     uintptr
	hMidiOut    uintptr
	sysexBuffer [2]*winmm.MIDIHDR

	pendingNotes        chan *midiMessage
	lastNoteOn          *midiMessage
	pressedKeys         [256]bool
	pressedKeysCount    int
	isCtrlDown          bool
	isAltDown           bool
	isShiftDown         bool
	clearModifiersTimer *time.Timer
	keysMutex           *sync.Mutex
}

type midiMessage struct {
	Time time.Time
	Msg  []byte
}

func main() {
	app := new(application)
	os.Exit(app.run(os.Args))
}

func (app *application) run(args []string) int {
	fmt.Println("MIDI2FFXIV")
	fmt.Println("Copyright (c) 2018 Star Brilliant")
	fmt.Println("=================================")
	fmt.Println()
	err := app.parseArgs(args)
	if err != nil {
		log.Println("Error: ", err)
		return 1
	}

	runtime.LockOSThread()
	_ = kernel32.SetPriorityClass(kernel32.GetCurrentProcess(), kernel32.HIGH_PRIORITY_CLASS)

	hWndClass, err := user32.RegisterClassEx(0, app.onMidiInMessage, 0, 0, 0, 0, 0, 0, 0, "midi2ffxiv", 0)
	if err != nil {
		fmt.Println("Error: ", err)
		return int(err.(syscall.Errno))
	}
	app.hWnd, err = user32.CreateWindowEx(0, uintptr(hWndClass), "midi2ffxiv", 0, 0, 0, 0, 0, user32.HWND_MESSAGE, 0, 0, nil)
	if err != nil {
		fmt.Println("Error: ", err)
		return int(err.(syscall.Errno))
	}

	app.pendingNotes = make(chan *midiMessage, 256)
	app.clearModifiersTimer = time.NewTimer(app.IdleDuration)
	app.keysMutex = new(sync.Mutex)

	go app.consumeStdin()
	go app.consumeMidiMessage()
	go app.clearModifiers()

	for !app.bQuitting {
		bResult, lpMsg, err := user32.GetMessage(app.hWnd, 0, 0)
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(int(err.(syscall.Errno)))
		}
		if bResult == 0 {
			break
		}
		_ = user32.TranslateMessage(lpMsg)
		_ = user32.DispatchMessage(lpMsg)
	}

	return 0
}

func (app *application) addNote(note []byte) {
	app.pendingNotes <- &midiMessage{
		Time: time.Now(),
		Msg:  note,
	}
}

func (app *application) clearModifiers() {
	for {
		<-app.clearModifiersTimer.C
		app.keysMutex.Lock()
		pInputs := []user32.INPUT_KEYBDINPUT{}
		if app.isCtrlDown {
			pInputs = append(pInputs, user32.INPUT_KEYBDINPUT{
				Type: user32.INPUT_KEYBOARD,
				Ki: user32.KEYBDINPUT{
					WVk:         0,
					WScan:       uint16(user32.MapVirtualKey(uint32(user32.VK_CONTROL), user32.MAPVK_VK_TO_VSC)),
					DwFlags:     user32.KEYEVENTF_SCANCODE | user32.KEYEVENTF_KEYUP,
					Time:        0,
					DwExtraInfo: 0,
				},
			})
			app.isCtrlDown = false
		}
		if app.isAltDown {
			pInputs = append(pInputs, user32.INPUT_KEYBDINPUT{
				Type: user32.INPUT_KEYBOARD,
				Ki: user32.KEYBDINPUT{
					WVk:         0,
					WScan:       uint16(user32.MapVirtualKey(uint32(user32.VK_MENU), user32.MAPVK_VK_TO_VSC)),
					DwFlags:     user32.KEYEVENTF_SCANCODE | user32.KEYEVENTF_KEYUP,
					Time:        0,
					DwExtraInfo: 0,
				},
			})
			app.isAltDown = false
		}
		if app.isShiftDown {
			pInputs = append(pInputs, user32.INPUT_KEYBDINPUT{
				Type: user32.INPUT_KEYBOARD,
				Ki: user32.KEYBDINPUT{
					WVk:         0,
					WScan:       uint16(user32.MapVirtualKey(uint32(user32.VK_SHIFT), user32.MAPVK_VK_TO_VSC)),
					DwFlags:     user32.KEYEVENTF_SCANCODE | user32.KEYEVENTF_KEYUP,
					Time:        0,
					DwExtraInfo: 0,
				},
			})
			app.isShiftDown = false
		}
		if len(pInputs) != 0 {
			for i := range pInputs {
				_, err := user32.SendInput(pInputs[i : i+1])
				if err != nil {
					fmt.Println("Error: ", err)
				}
			}
			app.printPressedKeys()
		}
		app.keysMutex.Unlock()
	}
}

func (app *application) consumeMidiMessage() {
	for {
		nextNote := <-app.pendingNotes
		now := time.Now()

		if (nextNote.Msg[0] == 0x90 || nextNote.Msg[0] == 0xa0) && now.Sub(nextNote.Time) > app.MaxNoteDelay {
			continue
		}

		if app.lastNoteOn != nil && ((nextNote.Msg[0] == 0x80 && nextNote.Msg[1] == app.lastNoteOn.Msg[1]) || nextNote.Msg[0] == 0x90) && now.Sub(app.lastNoteOn.Time) < app.SkillCooldown {
			time.Sleep(app.lastNoteOn.Time.Add(app.SkillCooldown).Sub(now))
			now = time.Now()
		}

		err := app.sendMidiOutMessage(nextNote)
		if err != nil {
			fmt.Println("Error: ", err)
		}

		if nextNote.Msg[0] == 0x80 || nextNote.Msg[0] == 0x90 {
			app.produceKeystroke(nextNote)
		}

		if nextNote.Msg[0] == 0x90 {
			nextNote.Time = now
			app.lastNoteOn = nextNote
		}
	}
}

func (app *application) consumeStdin() {
	var readBuffer [512]byte
	for {
		n, err := os.Stdin.Read(readBuffer[:])
		if n == 0 || err != nil {
			break
		}
	}
}

func (app *application) delayReturn(code int) int {
	fmt.Fprint(os.Stderr, "Press Ctrl-C to exit...")
	time.Sleep(1 * time.Minute)
	fmt.Fprintln(os.Stderr)
	return code
}

func (app *application) parseArgs(args []string) error {
	app.preset = defaultPreset
	app.MidiInDevice = -1
	app.MidiOutDevice = -1
	app.MidiOutInstrument = 46
	app.MidiOutTranspose = 0
	if len(args) >= 2 {
		value, err := strconv.ParseInt(args[1], 0, 32)
		if err != nil {
			return err
		}
		app.MidiInDevice = int(value)
	}
	if len(args) >= 3 {
		value, err := strconv.ParseInt(args[2], 0, 32)
		if err != nil {
			return err
		}
		app.MidiOutDevice = int(value)
	}
	if len(args) == 4 {
		switch strings.ToLower(args[3]) {
		case "harp":
			app.MidiOutInstrument = 46
			app.MidiOutTranspose = 0
		case "grandpiano", "piano":
			app.MidiOutInstrument = 0
			app.MidiOutTranspose = 12
		case "steelguitar", "lute":
			app.MidiOutInstrument = 25
			app.MidiOutTranspose = -12
		case "pizzicato", "fiddle":
			app.MidiOutInstrument = 45
			app.MidiOutTranspose = 0
		case "flute":
			app.MidiOutInstrument = 73
			app.MidiOutTranspose = 0
		case "oboe":
			app.MidiOutInstrument = 68
			app.MidiOutTranspose = 0
		case "clarinet":
			app.MidiOutInstrument = 71
			app.MidiOutTranspose = 0
		case "piccolo", "fife":
			app.MidiOutInstrument = 72
			app.MidiOutTranspose = 0
		case "panpipes", "panflute":
			app.MidiOutInstrument = 75
			app.MidiOutTranspose = 0
		default:
			value, err := strconv.ParseUint(args[3], 0, 32)
			if err != nil {
				return err
			}
			app.MidiOutInstrument = uint32(value - 1)
			app.MidiOutTranspose = 0
		}
	}
	if len(args) >= 5 {
		return errors.New("wrong number of arguments")
	}
	return nil
}

func (app *application) printPressedKeys() {
	pressedKeysCount := 0
	line := "\t ["
	if app.isCtrlDown {
		line += " Ctrl"
	}
	if app.isAltDown {
		line += " Alt"
	}
	if app.isShiftDown {
		line += " Shift"
	}
	for i, v := range app.pressedKeys {
		if v {
			line += fmt.Sprintf(" %q", rune(i))
			pressedKeysCount++
		}
	}
	line += " ]"
	fmt.Println(line)
	if pressedKeysCount != app.pressedKeysCount {
		panic(fmt.Sprintf("pressedKeysCount (%d) != app.pressedKeysCount (%d)", pressedKeysCount, app.pressedKeysCount))
	}
}

func (app *application) printUsage() {
	fmt.Printf("Usage: %s MidiInDevice [MidiOutDevice Instrument]\n", filepath.Base(os.Args[0]))
	fmt.Println()
	fmt.Println("List of MIDI IN devices:")
	midiInDeviceCount := winmm.MidiInGetNumDevs()
	for i := uint32(0); i < midiInDeviceCount; i++ {
		deviceName, _ := getMidiInDevName(uintptr(i))
		fmt.Printf("  %d: %s\n", i, deviceName)
	}
	fmt.Println()
	fmt.Println("List of MIDI OUT devices:")
	midiOutDeviceCount := winmm.MidiOutGetNumDevs()
	for i := uint32(0); i < midiOutDeviceCount; i++ {
		deviceName, _ := getMidiOutDevName(uintptr(i))
		fmt.Printf("  %d: %s\n", i, deviceName)
	}
	fmt.Println()
	fmt.Println("List of instrument sounds:")
	fmt.Println("  Harp:        General MIDI 0:47")
	fmt.Println("  GrandPiano:  General MIDI 0:1")
	fmt.Println("  SteelGuitar: General MIDI 0:26")
	fmt.Println("  Pizzicato:   General MIDI 0:46")
	fmt.Println("  Flute:       General MIDI 0:74")
	fmt.Println("  Oboe:        General MIDI 0:69")
	fmt.Println("  Clarinet:    General MIDI 0:72")
	fmt.Println("  Piccolo:     General MIDI 0:73")
	fmt.Println("  Panpipes:    General MIDI 0:76")
	fmt.Println()
}

func (app *application) processMidiMessage(midiMsg []byte) {
	channel := midiMsg[0] & 0xf
	// Ignore percussion channel
	if channel == 9 {
		return
	}
	midiMsg[0] = midiMsg[0] & 0xf0
	switch midiMsg[0] {
	// Note off
	case 0x80:
		if app.Keybinding[int(midiMsg[1])].VirtualKeyCode == 0 {
			break
		}
		app.addNote(midiMsg)
	// Note on
	case 0x90:
		if app.Keybinding[int(midiMsg[1])].VirtualKeyCode == 0 {
			break
		}
		if midiMsg[2] < app.MinTriggerVelocity {
			midiMsg[0] = 0x80
		}
		app.addNote(midiMsg)
	// After touch
	case 0xa0:
		if app.Keybinding[int(midiMsg[1])].VirtualKeyCode == 0 {
			break
		}
		if midiMsg[2] == 0 {
			midiMsg[0] = 0x80
		}
		app.addNote(midiMsg)
	// Control change
	case 0xb0:
		// Block bank select
		if midiMsg[1] == 0x00 || midiMsg[1] == 0x20 {
			break
		}
		app.addNote(midiMsg)
	// Channel pressure
	case 0xd0:
		app.addNote(midiMsg)
	}
}

func (app *application) produceKeystroke(note *midiMessage) {
	app.keysMutex.Lock()
	defer app.keysMutex.Unlock()

	pInputs := []user32.INPUT_KEYBDINPUT{}
	if note.Msg[0] == 0x90 {
		keybind := app.Keybinding[int(note.Msg[1])]
		if app.pressedKeys[keybind.VirtualKeyCode] {
			pInputs = append(pInputs, user32.INPUT_KEYBDINPUT{
				Type: user32.INPUT_KEYBOARD,
				Ki: user32.KEYBDINPUT{
					WVk:         0,
					WScan:       uint16(user32.MapVirtualKey(uint32(keybind.VirtualKeyCode), user32.MAPVK_VK_TO_VSC)),
					DwFlags:     user32.KEYEVENTF_SCANCODE | user32.KEYEVENTF_KEYUP,
					Time:        0,
					DwExtraInfo: 0,
				},
			})
			app.pressedKeysCount--
		}
		if app.isCtrlDown != keybind.Ctrl {
			dwFlags := user32.KEYEVENTF_SCANCODE | user32.KEYEVENTF_KEYUP
			if keybind.Ctrl {
				dwFlags = user32.KEYEVENTF_SCANCODE
			}
			pInputs = append(pInputs, user32.INPUT_KEYBDINPUT{
				Type: user32.INPUT_KEYBOARD,
				Ki: user32.KEYBDINPUT{
					WVk:         0,
					WScan:       uint16(user32.MapVirtualKey(uint32(user32.VK_CONTROL), user32.MAPVK_VK_TO_VSC)),
					DwFlags:     dwFlags,
					Time:        0,
					DwExtraInfo: 0,
				},
			})
			app.isCtrlDown = keybind.Ctrl
		}
		if app.isAltDown != keybind.Alt {
			dwFlags := user32.KEYEVENTF_SCANCODE | user32.KEYEVENTF_KEYUP
			if keybind.Alt {
				dwFlags = user32.KEYEVENTF_SCANCODE
			}
			pInputs = append(pInputs, user32.INPUT_KEYBDINPUT{
				Type: user32.INPUT_KEYBOARD,
				Ki: user32.KEYBDINPUT{
					WVk:         0,
					WScan:       uint16(user32.MapVirtualKey(uint32(user32.VK_MENU), user32.MAPVK_VK_TO_VSC)),
					DwFlags:     dwFlags,
					Time:        0,
					DwExtraInfo: 0,
				},
			})
			app.isAltDown = keybind.Alt
		}
		if app.isShiftDown != keybind.Shift {
			dwFlags := user32.KEYEVENTF_SCANCODE | user32.KEYEVENTF_KEYUP
			if keybind.Shift {
				dwFlags = user32.KEYEVENTF_SCANCODE
			}
			pInputs = append(pInputs, user32.INPUT_KEYBDINPUT{
				Type: user32.INPUT_KEYBOARD,
				Ki: user32.KEYBDINPUT{
					WVk:         0,
					WScan:       uint16(user32.MapVirtualKey(uint32(user32.VK_SHIFT), user32.MAPVK_VK_TO_VSC)),
					DwFlags:     dwFlags,
					Time:        0,
					DwExtraInfo: 0,
				},
			})
			app.isShiftDown = keybind.Shift
		}
		pInputs = append(pInputs, user32.INPUT_KEYBDINPUT{
			Type: user32.INPUT_KEYBOARD,
			Ki: user32.KEYBDINPUT{
				WVk:         0,
				WScan:       uint16(user32.MapVirtualKey(uint32(keybind.VirtualKeyCode), user32.MAPVK_VK_TO_VSC)),
				DwFlags:     user32.KEYEVENTF_SCANCODE,
				Time:        0,
				DwExtraInfo: 0,
			},
		})
		app.pressedKeys[keybind.VirtualKeyCode] = true
		app.pressedKeysCount++
		app.clearModifiersTimer.Stop()
	} else if note.Msg[0] == 0x80 {
		keybind := app.Keybinding[int(note.Msg[1])]
		if app.pressedKeys[keybind.VirtualKeyCode] {
			pInputs = append(pInputs, user32.INPUT_KEYBDINPUT{
				Type: user32.INPUT_KEYBOARD,
				Ki: user32.KEYBDINPUT{
					WVk:         0,
					WScan:       uint16(user32.MapVirtualKey(uint32(keybind.VirtualKeyCode), user32.MAPVK_VK_TO_VSC)),
					DwFlags:     user32.KEYEVENTF_SCANCODE | user32.KEYEVENTF_KEYUP,
					Time:        0,
					DwExtraInfo: 0,
				},
			})
			app.pressedKeys[keybind.VirtualKeyCode] = false
			app.pressedKeysCount--
		}
		if app.pressedKeysCount == 0 {
			app.clearModifiersTimer.Reset(app.IdleDuration)
		}
	}
	if len(pInputs) != 0 {
		for i := range pInputs {
			_, err := user32.SendInput(pInputs[i : i+1])
			if err != nil {
				fmt.Println("Error: ", err)
			}
		}
		app.printPressedKeys()
	}
}
