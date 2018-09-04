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
	"fmt"
	"log"
	"time"

	"./user32"
	cgc "github.com/m13253/cgc-go"
)

type keystroke struct {
	Pressed     bool
	MidiNote    uint8
	LastChange  time.Time
	LastPress   time.Time
	LastRelease time.Time
}

type keystrokeStatus struct {
	pressedKeys         [256]keystroke
	pressedKeysCount    int
	ctrl                keystroke
	alt                 keystroke
	shift               keystroke
	lastNoteOn          *midiRealtimeEvent
	clearModifiersTimer *time.Timer
}

func (app *application) processKeystrokes() {
	app.keyStatus = &keystrokeStatus{
		clearModifiersTimer: time.NewTimer(app.IdleDuration),
	}
	for {
		select {
		case r, ok := <-app.KeystrokeGoro:
			if !ok {
				return
			}
			_ = cgc.RunOneRequest(app.ctx, r)
		case <-app.keyStatus.clearModifiersTimer.C:
			app.clearModifiers()
		case <-app.ctx.Done():
			return
		}
	}
}

func (app *application) produceKeystroke(event *midiRealtimeEvent) {
	pInputs := []user32.INPUT_KEYBDINPUT{}
	now := time.Now()
	if event.Message[0] == 0x90 {
		app.keyStatus.clearModifiersTimer.Stop()
		keybind := app.Keybinding[int(event.Message[1])]
		if app.keyStatus.pressedKeys[keybind.VirtualKeyCode].Pressed {
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
			app.keyStatus.pressedKeys[keybind.VirtualKeyCode].Pressed = false
			app.keyStatus.pressedKeys[keybind.VirtualKeyCode].LastChange = now
			app.keyStatus.pressedKeys[keybind.VirtualKeyCode].LastRelease = now
			app.keyStatus.pressedKeysCount--
		}
		if app.keyStatus.ctrl.Pressed != keybind.Ctrl {
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
			if keybind.Ctrl {
				app.keyStatus.ctrl.Pressed = true
				app.keyStatus.ctrl.LastChange = now
				app.keyStatus.ctrl.LastPress = now
			} else {
				app.keyStatus.ctrl.Pressed = false
				app.keyStatus.ctrl.LastChange = now
				app.keyStatus.ctrl.LastRelease = now
			}
		}
		if app.keyStatus.alt.Pressed != keybind.Alt {
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
			if keybind.Alt {
				app.keyStatus.alt.Pressed = true
				app.keyStatus.alt.LastChange = now
				app.keyStatus.alt.LastPress = now
			} else {
				app.keyStatus.alt.Pressed = false
				app.keyStatus.alt.LastChange = now
				app.keyStatus.alt.LastRelease = now
			}
		}
		if app.keyStatus.shift.Pressed != keybind.Shift {
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
			if keybind.Shift {
				app.keyStatus.shift.Pressed = true
				app.keyStatus.shift.LastChange = now
				app.keyStatus.shift.LastPress = now
			} else {
				app.keyStatus.shift.Pressed = false
				app.keyStatus.shift.LastChange = now
				app.keyStatus.shift.LastRelease = now
			}
		}
		if now.Sub(app.keyStatus.ctrl.LastChange) < app.ModifierCooldown || now.Sub(app.keyStatus.alt.LastChange) < app.ModifierCooldown || now.Sub(app.keyStatus.shift.LastChange) < app.ModifierCooldown || now.Sub(app.keyStatus.pressedKeys[keybind.VirtualKeyCode].LastRelease) < app.ModifierCooldown {
			for i := range pInputs {
				_, err := user32.SendInput(pInputs[i : i+1])
				if err != nil {
					fmt.Println("Error: ", err)
				}
			}
			app.printPressedKeys()
			pInputs = pInputs[:0]
			time.Sleep(app.ModifierCooldown)
			now = time.Now()
		}
		if app.keyStatus.lastNoteOn != nil && ((event.Message[0] == 0x80 && event.Message[1] == app.keyStatus.lastNoteOn.Message[1]) || event.Message[0] == 0x90) && now.Sub(app.keyStatus.lastNoteOn.Time) < app.SkillCooldown {
			time.Sleep(app.keyStatus.lastNoteOn.Time.Add(app.SkillCooldown).Sub(now))
			now = time.Now()
		}
		event.Time = now
		app.keyStatus.lastNoteOn = event
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
		app.keyStatus.pressedKeys[keybind.VirtualKeyCode].Pressed = true
		app.keyStatus.pressedKeys[keybind.VirtualKeyCode].MidiNote = event.Message[1]
		app.keyStatus.pressedKeys[keybind.VirtualKeyCode].LastPress = now
		app.keyStatus.pressedKeysCount++
	} else if event.Message[0] == 0x80 {
		keybind := app.Keybinding[int(event.Message[1])]
		if app.keyStatus.pressedKeys[keybind.VirtualKeyCode].Pressed && app.keyStatus.pressedKeys[keybind.VirtualKeyCode].MidiNote == event.Message[1] {
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
			app.keyStatus.pressedKeys[keybind.VirtualKeyCode].Pressed = false
			app.keyStatus.pressedKeys[keybind.VirtualKeyCode].LastRelease = now
			app.keyStatus.pressedKeysCount--
		}
		if app.keyStatus.pressedKeysCount == 0 {
			app.keyStatus.clearModifiersTimer.Reset(app.IdleDuration)
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

func (app *application) clearModifiers() {
	pInputs := []user32.INPUT_KEYBDINPUT{}
	now := time.Now()
	if app.keyStatus.ctrl.Pressed {
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
		app.keyStatus.ctrl.Pressed = false
		app.keyStatus.ctrl.LastChange = now
		app.keyStatus.ctrl.LastRelease = now
	}
	if app.keyStatus.alt.Pressed {
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
		app.keyStatus.alt.Pressed = false
		app.keyStatus.alt.LastChange = now
		app.keyStatus.alt.LastRelease = now
	}
	if app.keyStatus.shift.Pressed {
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
		app.keyStatus.shift.Pressed = false
		app.keyStatus.shift.LastChange = now
		app.keyStatus.shift.LastRelease = now
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

func (app *application) printPressedKeys() {
	pressedKeysCount := 0
	line := "["
	if app.keyStatus.ctrl.Pressed {
		line += " Ctrl"
	}
	if app.keyStatus.alt.Pressed {
		line += " Alt"
	}
	if app.keyStatus.shift.Pressed {
		line += " Shift"
	}
	for i, v := range app.keyStatus.pressedKeys {
		if v.Pressed {
			line += fmt.Sprintf(" %q", rune(i))
			pressedKeysCount++
		}
	}
	line += " ]"
	log.Println(line)
	if pressedKeysCount != app.keyStatus.pressedKeysCount {
		panic(fmt.Sprintf("pressedKeysCount (%d) != keyStatus.pressedKeysCount (%d)", pressedKeysCount, app.keyStatus.pressedKeysCount))
	}
}
