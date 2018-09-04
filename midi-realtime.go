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
	"runtime"
	"time"

	"./winmm"
	"golang.org/x/sys/windows"
)

type midiRealtimeEvent struct {
	Time    time.Time
	Message []byte
}

func (app *application) listMidiInDevices() []string {
	midiInDeviceCount := winmm.MidiInGetNumDevs()
	results := make([]string, midiInDeviceCount)
	for i := uint32(0); i < midiInDeviceCount; i++ {
		deviceName, _ := getMidiInDevName(uintptr(i))
		results[i] = deviceName
	}
	return results
}

func (app *application) listMidiOutDevices() []string {
	midiOutDeviceCount := winmm.MidiOutGetNumDevs()
	results := make([]string, midiOutDeviceCount)
	for i := uint32(0); i < midiOutDeviceCount; i++ {
		deviceName, _ := getMidiOutDevName(uintptr(i))
		results[i] = deviceName
	}
	return results
}

func (app *application) openMidiInDevice(midiInDevice int) error {
	app.closeMidiInDevice()
	if midiInDevice < 0 {
		return nil
	}
	midiInDeviceCount := winmm.MidiInGetNumDevs()
	if midiInDevice >= int(midiInDeviceCount) {
		return winmm.MidiInError(winmm.MMSYSERR_BADDEVICEID)
	}

	hMidiIn, err := winmm.MidiInOpen(uint32(midiInDevice), app.hWnd, 0, winmm.CALLBACK_WINDOW|winmm.MIDI_IO_STATUS)
	if err != nil {
		return err
	}

	for i := range app.sysexBuffer {
		app.sysexBuffer[i] = &winmm.MIDIHDR{
			LpData:         &new([65536]byte)[0],
			DwBufferLength: 65536,
		}
		err = winmm.MidiInPrepareHeader(hMidiIn, app.sysexBuffer[i])
		if err != nil {
			_ = winmm.MidiInClose(hMidiIn)
			return err
		}
		err = winmm.MidiInAddBuffer(hMidiIn, app.sysexBuffer[i])
		if err != nil {
			_ = winmm.MidiInClose(hMidiIn)
			return err
		}
	}

	app.MidiInDevice = midiInDevice
	app.hMidiIn = hMidiIn
	err = winmm.MidiInStart(app.hMidiIn)
	if err != nil {
		app.closeMidiInDevice()
		return err
	}

	return nil
}

func (app *application) openMidiOutDevice(midiOutDevice int) error {
	app.closeMidiOutDevice()
	if midiOutDevice < 0 {
		return nil
	}
	MidiOutDeviceCount := winmm.MidiOutGetNumDevs()
	if midiOutDevice >= int(MidiOutDeviceCount) {
		return winmm.MidiOutError(winmm.MMSYSERR_BADDEVICEID)
	}

	hMidiOut, err := winmm.MidiOutOpen(uint32(midiOutDevice), app.hWnd, 0, winmm.CALLBACK_NULL)
	if err != nil {
		return err
	}

	app.MidiOutDevice = midiOutDevice
	app.hMidiOut = hMidiOut

	err = app.setMidiOutBank(app.MidiOutBank)
	if err != nil {
		return err
	}
	return app.setMidiOutPatch(app.MidiOutPatch)
}

func (app *application) closeMidiInDevice() {
	app.MidiInDevice = -1
	if app.hMidiIn == 0 {
		return
	}
	for i := range app.sysexBuffer {
		_ = winmm.MidiInUnprepareHeader(app.hMidiIn, app.sysexBuffer[i])
	}
	_ = winmm.MidiInClose(app.hMidiIn)
	app.hMidiIn = 0
}

func (app *application) closeMidiOutDevice() {
	app.MidiOutDevice = -1
	if app.hMidiOut == 0 {
		return
	}
	_ = app.sendAllNoteOff()
	_ = winmm.MidiOutClose(app.hMidiOut)
	app.hMidiOut = 0
}

func (app *application) setMidiOutBank(midiOutBank uint16) error {
	err := winmm.MidiOutShortMsg(app.hMidiOut, 0x0000b0|((uint32(midiOutBank)<<23)&0x7f0000))
	if err != nil {
		return err
	}
	err = winmm.MidiOutShortMsg(app.hMidiOut, 0x0020b0|((uint32(midiOutBank)<<16)&0x7f0000))
	if err != nil {
		return err
	}
	app.MidiOutBank = midiOutBank
	return nil
}

func (app *application) setMidiOutPatch(midiOutPatch uint8) error {
	err := winmm.MidiOutShortMsg(app.hMidiOut, 0x00c0|((uint32(midiOutPatch)<<8)&0x7f00))
	if err != nil {
		return err
	}
	app.MidiOutPatch = midiOutPatch
	return nil
}

func (app *application) setMidiOutTranspose(midiOutTranspose int) {
	app.MidiOutTranspose = midiOutTranspose
}

func (app *application) onMidiInEvent(event []byte) {
	channel := event[0] & 0xf
	// Ignore percussion channel
	if channel == 9 {
		return
	}
	// Force channel 1
	event[0] = event[0] & 0xf0
	switch event[0] {
	// Note off
	case 0x80:
		if app.Keybinding[int(event[1])].VirtualKeyCode == 0 {
			return
		}
	// Note on
	case 0x90:
		if app.Keybinding[int(event[1])].VirtualKeyCode == 0 {
			return
		}
		if event[2] < app.MinTriggerVelocity {
			event[0] = 0x80
		}
	// After touch
	case 0xa0:
		if app.Keybinding[int(event[1])].VirtualKeyCode == 0 {
			return
		}
		if event[2] == 0 {
			event[0] = 0x80
		}
	// Control change
	case 0xb0:
		// Block bank select
		if event[1] == 0x00 || event[1] == 0x20 {
			return
		}
	// Program change
	case 0xc0:
		event = event[:2]
	// Channel pressure
	case 0xd0:
		event = event[:2]
	// Pitch bend
	case 0xe0:
		return
	// System Messages
	case 0xf0:
	}
	app.pendingNotes <- &midiRealtimeEvent{
		Time:    time.Now(),
		Message: event,
	}
}

func (app *application) sendMidiOutMessage(event *midiRealtimeEvent) error {
	if app.MidiOutDevice == -1 {
		return nil
	}
	var err error
	switch len(event.Message) {
	case 1:
		err = winmm.MidiOutShortMsg(app.hMidiOut, uint32(event.Message[0]))
	case 2:
		err = winmm.MidiOutShortMsg(app.hMidiOut, uint32(event.Message[0])|(uint32(event.Message[1])<<8))
	case 3:
		if event.Message[0] == 0x80 || event.Message[0] == 0x90 || event.Message[0] == 0xa0 {
			noteName := int(event.Message[1]) + app.MidiOutTranspose
			if noteName >= 0x00 || noteName <= 0x7f {
				err = winmm.MidiOutShortMsg(app.hMidiOut, uint32(event.Message[0])|(uint32(noteName)<<8)|(uint32(event.Message[2])<<16))
			}
		} else {
			err = winmm.MidiOutShortMsg(app.hMidiOut, uint32(event.Message[0])|(uint32(event.Message[1])<<8)|(uint32(event.Message[2])<<16))
		}
	default:
		buffer := make([]byte, len(event.Message))
		midiHeader := &winmm.MIDIHDR{
			LpData:          &buffer[0],
			DwBufferLength:  uint32(len(event.Message)),
			DwBytesRecorded: uint32(len(event.Message)),
		}
		copy(buffer, event.Message)
		err = winmm.MidiOutPrepareHeader(app.hMidiOut, midiHeader)
		if err != nil {
			return err
		}
		defer func() {
			for {
				err := winmm.MidiOutUnprepareHeader(app.hMidiOut, midiHeader)
				if uint32(err.(winmm.MidiOutError)) != winmm.MIDIERR_STILLPLAYING {
					break
				}
				runtime.Gosched()
			}
		}()
		err = winmm.MidiOutLongMsg(app.hMidiOut, midiHeader)
	}
	return err
}

func (app *application) sendAllNoteOff() error {
	for i := uint32(0x007bb0); i <= 0x007bbf; i++ {
		err := winmm.MidiOutShortMsg(app.hMidiOut, i)
		if err != nil {
			return err
		}
	}
	return nil
}

func getMidiInDevName(uDeviceID uintptr) (string, error) {
	lpMidiInCaps, err := winmm.MidiInGetDevCaps(uDeviceID)
	if err != nil {
		return fmt.Sprintf("(Error: %s)", err.Error()), err
	}
	return windows.UTF16ToString(lpMidiInCaps.SzPname[:]), nil
}

func getMidiOutDevName(uDeviceID uintptr) (string, error) {
	lpMidiOutCaps, err := winmm.MidiOutGetDevCaps(uDeviceID)
	if err != nil {
		return fmt.Sprintf("(Error: %s)", err.Error()), err
	}
	return windows.UTF16ToString(lpMidiOutCaps.SzPname[:]), nil
}
