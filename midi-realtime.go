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
	"runtime"
	"time"

	"github.com/m13253/midi2ffxiv/winmm"
	cgc "github.com/m13253/cgc-go"
	"golang.org/x/sys/windows"
)

type midiQueueEvent struct {
	Time              time.Time
	Expiry            time.Time
	Message           []byte
	Realtime          bool
	FastForward       bool
	AlreadyTransposed bool
}

func (app *application) processMidiRealtime() {
	for {
		select {
		case r, ok := <-app.MidiRealtimeGoro:
			if !ok {
				return
			}
			cgc.RunOneRequest(app.ctx, r)
		case nextAction := <-app.midiOutQueue.NextAction():
			nextEvent := nextAction.Value.(*midiQueueEvent)
			err := app.sendMidiOutMessage(nextEvent)
			if err != nil {
				log.Println("Error: ", err)
			}
		case <-app.ctx.Done():
			return
		}
	}
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

	app.setMidiOutBank(app.MidiOutBank)
	app.setMidiOutPatch(app.MidiOutPatch)
	return nil
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
	_ = app.sendMidiOutMessage(&midiQueueEvent{
		Message: []byte{0xb0, 0x7b, 0x00},
	})
	app.MidiOutDevice = -1
	if app.hMidiOut == 0 {
		return
	}
	hMidiOut := app.hMidiOut
	app.hMidiOut = 0
	time.AfterFunc(1*time.Second, func() {
		_ = winmm.MidiOutClose(hMidiOut)
	})
}

func (app *application) setMidiOutBank(midiOutBank uint16) {
	app.keystrokeQueue.AddAction(&midiQueueEvent{
		Message:  []byte{0xb0, 0x00, uint8(midiOutBank>>15) & 0x7f},
		Realtime: true,
	}, time.Time{})
	app.keystrokeQueue.AddAction(&midiQueueEvent{
		Message:  []byte{0xb0, 0x20, uint8(midiOutBank) & 0x7f},
		Realtime: true,
	}, time.Time{})
	app.MidiOutBank = midiOutBank
}

func (app *application) setMidiOutPatch(midiOutPatch uint8) {
	app.keystrokeQueue.AddAction(&midiQueueEvent{
		Message:  []byte{0xc0, midiOutPatch & 0x7f},
		Realtime: true,
	}, time.Time{})
	app.MidiOutPatch = midiOutPatch
}

func (app *application) setMidiOutTranspose(midiOutTranspose int) {
	app.sendAllNoteOff(true)
	app.MidiOutTranspose = midiOutTranspose
}

func (app *application) onMidiInEvent(event []byte) {
	if len(event) == 0 {
		return
	}
	app.addMidiEvent(&midiQueueEvent{
		Time:     time.Now(),
		Message:  event,
		Realtime: true,
	})
}

func (app *application) addMidiEvent(event *midiQueueEvent) {
	channel := event.Message[0] & 0xf
	// Ignore percussion channel
	if channel == 9 {
		return
	}
	// Force channel 1
	filteredMessage := make([]byte, len(event.Message))
	copy(filteredMessage, event.Message)
	filteredMessage[0] &= 0xf0

	expiry := event.Expiry
	switch filteredMessage[0] {
	// Note off
	case 0x80:
		note := int(filteredMessage[1])
		if !event.AlreadyTransposed {
			note += app.MidiOutTranspose
			if note < 0x00 || note > 0x7f {
				return
			}
			filteredMessage[1] = uint8(note)
		}
	// Note on
	case 0x90:
		if event.FastForward {
			return
		}
		note := int(filteredMessage[1])
		if !event.AlreadyTransposed {
			note += app.MidiOutTranspose
			if note < 0x00 || note > 0x7f {
				return
			}
			filteredMessage[1] = uint8(note)
		}
		if filteredMessage[2] == 0 || filteredMessage[2] < app.MinTriggerVelocity {
			filteredMessage[0] = 0x80
		} else {
			if expiry.IsZero() && !event.Time.IsZero() {
				if event.Realtime {
					expiry = event.Time.Add(app.RealtimeMaxLatency)
				} else {
					expiry = event.Time.Add(app.PlaybackMaxLatency)
				}
			}
		}
	// After touch
	case 0xa0:
		if event.FastForward {
			return
		}
		note := int(filteredMessage[1])
		if !event.AlreadyTransposed {
			note += app.MidiOutTranspose
			if note < 0x00 || note > 0x7f {
				return
			}
			filteredMessage[1] = uint8(note)
		}
		if filteredMessage[2] == 0 {
			filteredMessage[0] = 0x80
		} else {
			if expiry.IsZero() && !event.Time.IsZero() {
				if event.Realtime {
					expiry = event.Time.Add(app.RealtimeMaxLatency)
				} else {
					expiry = event.Time.Add(app.PlaybackMaxLatency)
				}
			}
		}
	// Control change
	case 0xb0:
		// Block bank select
		if filteredMessage[1] == 0x00 || filteredMessage[1] == 0x20 {
			return
		}
	// Program change
	case 0xc0:
		return
	// Channel pressure
	case 0xd0:
		filteredMessage = filteredMessage[:2]
	// Pitch bend
	case 0xe0:
		return
	// System Messages
	case 0xf0:
	}
	app.keystrokeQueue.AddActionWithExpiry(&midiQueueEvent{
		Time:              event.Time,
		Expiry:            expiry,
		Message:           filteredMessage,
		Realtime:          event.Realtime,
		FastForward:       event.FastForward,
		AlreadyTransposed: true,
	}, event.Time, expiry)
}

func (app *application) sendMidiOutMessage(event *midiQueueEvent) error {
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
		err = winmm.MidiOutShortMsg(app.hMidiOut, uint32(event.Message[0])|(uint32(event.Message[1])<<8)|(uint32(event.Message[2])<<16))
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
				if err == nil {
					break
				}
				if midiOutError, ok := err.(winmm.MidiOutError); !ok || uint32(midiOutError) != winmm.MIDIERR_STILLPLAYING {
					break
				}
				runtime.Gosched()
			}
		}()
		err = winmm.MidiOutLongMsg(app.hMidiOut, midiHeader)
	}
	return err
}

func (app *application) sendAllNoteOff(realtime bool) {
	app.addMidiEvent(&midiQueueEvent{
		Message:  []byte{0xb0, 0x7b, 0x00},
		Realtime: realtime,
	})
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
