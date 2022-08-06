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

	"github.com/m13253/midi2ffxiv/user32"
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
	lastNote            uint8
	lastNoteTime        time.Time
	lastModifierTime    time.Time
	clearModifiersTimer *time.Timer
}

func (app *application) processKeystrokes() {
	app.keyStatus = &keystrokeStatus{
		clearModifiersTimer: time.NewTimer(app.IdleDuration),
		lastNote:            0xff,
	}
	for {
		select {
		case r, ok := <-app.KeystrokeGoro:
			if !ok {
				return
			}
			cgc.RunOneRequest(app.ctx, r)
		case nextAction := <-app.keystrokeQueue.NextAction():
			nextEvent := nextAction.Value.(*midiQueueEvent)
			app.produceKeystroke(nextEvent)
		case now := <-app.keyStatus.clearModifiersTimer.C:
			app.clearModifiers(now)
		case <-app.ctx.Done():
			return
		}
	}
}

func (app *application) produceKeystroke(event *midiQueueEvent) {
	pInputs := []user32.INPUT_KEYBDINPUT{}
	now := time.Now()
	if event.Message[0] == 0x80 {
		if event.Realtime {
			app.midiOutQueue.AddAction(event, now)
		} else {
			app.midiOutQueue.AddAction(event, now.Add(app.PlaybackExtraDelay))
		}
		note := int(event.Message[1])
		if event.AlreadyTransposed {
			note -= app.MidiOutTranspose
			if note < 0x00 || note > 0x7f {
				return
			}
		}
		keybind := &app.Keybinding[note]
		if keybind.VirtualKeyCode == 0 {
			return
		}
		if app.keyStatus.pressedKeys[keybind.VirtualKeyCode].Pressed && app.keyStatus.pressedKeys[keybind.VirtualKeyCode].MidiNote == uint8(note) {
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
	} else if event.Message[0] == 0x90 {
		app.keyStatus.clearModifiersTimer.Stop()
		note := int(event.Message[1])
		if event.AlreadyTransposed {
			note -= app.MidiOutTranspose
			if note < 0x00 || note > 0x7f {
				return
			}
		}
		keybind := &app.Keybinding[note]
		if keybind.VirtualKeyCode == 0 {
			noteName, _ := noteIndexToName(uint8(note))
			log.Printf("Note %s out of range.\n", noteName)
			return
		}
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
			app.keyStatus.lastModifierTime = now
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
			app.keyStatus.lastModifierTime = now
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
			app.keyStatus.lastModifierTime = now
		}
		if !event.Realtime && app.ModifierCooldown != 0 {
			if len(pInputs) != 0 {
				_, err := user32.SendInput(pInputs)
				if err != nil {
					log.Println("Error: ", err)
				}
				app.printPressedKeys()
				pInputs = []user32.INPUT_KEYBDINPUT{}
			}
			waitTime := app.ModifierCooldown
			if waitTime != 0 {
				log.Printf("Modifier cooldown (playback) %s.\n", waitTime)
				if waitTime < time.Millisecond {
					waitTime = 0
				}
				time.Sleep(waitTime)
				now = now.Add(waitTime)
			}
		}
		if !app.keyStatus.lastNoteTime.IsZero() && ((event.Message[0] == 0x80 && app.keyStatus.lastNote == uint8(note)) || event.Message[0] == 0x90) && now.Sub(app.keyStatus.lastNoteTime) < app.SkillCooldown {
			waitTime := app.keyStatus.lastNoteTime.Add(app.SkillCooldown).Sub(now)
			log.Printf("Skill cooldown sleep %s.\n", waitTime)
			if waitTime < time.Millisecond {
				waitTime = 0
			}
			time.Sleep(waitTime)
			now = now.Add(waitTime)
		}
		if !event.Expiry.IsZero() && now.After(event.Expiry) {
			return
		}
		if event.Realtime {
			app.midiOutQueue.AddAction(event, now.Add(app.RealtimeExtraDelay))
		} else {
			app.midiOutQueue.AddAction(event, now.Add(app.PlaybackExtraDelay))
		}
		if event.Realtime && !app.keyStatus.lastModifierTime.IsZero() && now.Sub(app.keyStatus.lastModifierTime) < app.ModifierCooldown {
			if len(pInputs) != 0 {
				_, err := user32.SendInput(pInputs)
				if err != nil {
					log.Println("Error: ", err)
				}
				app.printPressedKeys()
				pInputs = []user32.INPUT_KEYBDINPUT{}
			}
			waitTime := app.keyStatus.lastModifierTime.Add(app.ModifierCooldown).Sub(now)
			log.Printf("Modifier cooldown (realtime) %s.\n", waitTime)
			if waitTime < time.Millisecond {
				waitTime = 0
			}
			time.Sleep(waitTime)
			now = now.Add(waitTime)
		}
		app.keyStatus.lastNote = uint8(note)
		app.keyStatus.lastNoteTime = now
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
		app.keyStatus.pressedKeys[keybind.VirtualKeyCode].MidiNote = uint8(note)
		app.keyStatus.pressedKeys[keybind.VirtualKeyCode].LastChange = now
		app.keyStatus.pressedKeys[keybind.VirtualKeyCode].LastPress = now
		app.keyStatus.pressedKeysCount++
	} else if event.Message[0] == 0xb0 {
		if event.Realtime {
			app.midiOutQueue.AddAction(event, now)
		} else {
			app.midiOutQueue.AddAction(event, now.Add(app.PlaybackExtraDelay))
		}
		if len(event.Message) > 1 && event.Message[1] == 0x7b {
			for i := 0; i < 256; i++ {
				if app.keyStatus.pressedKeys[i].Pressed {
					pInputs = append(pInputs, user32.INPUT_KEYBDINPUT{
						Type: user32.INPUT_KEYBOARD,
						Ki: user32.KEYBDINPUT{
							WVk:         0,
							WScan:       uint16(user32.MapVirtualKey(uint32(i), user32.MAPVK_VK_TO_VSC)),
							DwFlags:     user32.KEYEVENTF_SCANCODE | user32.KEYEVENTF_KEYUP,
							Time:        0,
							DwExtraInfo: 0,
						},
					})
					app.keyStatus.pressedKeys[i].Pressed = false
					app.keyStatus.pressedKeys[i].LastChange = now
					app.keyStatus.pressedKeys[i].LastRelease = now
					app.keyStatus.pressedKeysCount--
				}
			}
			app.keyStatus.clearModifiersTimer.Reset(0)
		}
	} else {
		if event.Realtime {
			app.midiOutQueue.AddAction(event, now)
		} else {
			app.midiOutQueue.AddAction(event, now.Add(app.PlaybackExtraDelay))
		}
	}
	if len(pInputs) != 0 {
		_, err := user32.SendInput(pInputs)
		if err != nil {
			log.Println("Error: ", err)
		}
		app.printPressedKeys()
	}
}

func (app *application) clearModifiers(now time.Time) {
	pInputs := []user32.INPUT_KEYBDINPUT{}
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
		app.keyStatus.lastModifierTime = now
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
		app.keyStatus.lastModifierTime = now
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
		app.keyStatus.lastModifierTime = now
	}
	if len(pInputs) != 0 {
		_, err := user32.SendInput(pInputs)
		if err != nil {
			log.Println("Error: ", err)
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
