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

package winmm

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

type MidiInError uint32
type MidiOutError uint32

const (
	MMSYSERR_NOERROR  uint32 = 0
	CALLBACK_NULL     uint32 = 0
	CALLBACK_WINDOW   uint32 = 0x10000
	CALLBACK_THREAD   uint32 = 0x20000
	CALLBACK_FUNCTION uint32 = 0x30000
	MIDI_IO_STATUS    uint32 = 32
	MM_MIM_OPEN       uint32 = 0x3c1
	MM_MIM_CLOSE      uint32 = 0x3c2
	MM_MIM_DATA       uint32 = 0x3c3
	MM_MIM_LONGDATA   uint32 = 0x3c4
	MM_MIM_ERROR      uint32 = 0x3c5
	MM_MIM_LONGERROR  uint32 = 0x3c6
	MM_MOM_OPEN       uint32 = 0x3c7
	MM_MOM_CLOSE      uint32 = 0x3c8
	MM_MOM_DONE       uint32 = 0x3c9
	MM_MIM_MOREDATA   uint32 = 0x3cc
)

type MIDIINCAPS struct {
	WMid           uint16
	WPid           uint16
	VDriverVersion uint32
	SzPname        [32]uint16
	DwSupport      uint32
}

type MIDIHDR struct {
	LpData          *byte
	DwBufferLength  uint32
	DwBytesRecorded uint32
	DwUser          uintptr
	DwFlags         uint32
	LpNext          uintptr
	Reserved        uintptr
	DwOffset        uint32
	DwReserved      [4]uintptr
}

type MIDIOUTCAPS struct {
	WMid           uint16
	WPid           uint16
	VDriverVersion uint32
	SzPname        [32]uint16
	WTechnology    uint16
	WVoices        uint16
	WNotes         uint16
	WChannelMask   uint16
	DwSupport      uint32
}

type MidiInProc func(hMidiIn uintptr, wMsg uint32, dwInstance, dwParam1, dwParam2 uintptr) uintptr

var (
	winmm                  *windows.LazyDLL
	midiInAddBuffer        *windows.LazyProc
	midiInClose            *windows.LazyProc
	midiInGetDevCaps       *windows.LazyProc
	midiInGetErrorText     *windows.LazyProc
	midiInGetNumDevs       *windows.LazyProc
	midiInOpen             *windows.LazyProc
	midiInPrepareHeader    *windows.LazyProc
	midiInStart            *windows.LazyProc
	midiInStop             *windows.LazyProc
	midiInUnprepareHeader  *windows.LazyProc
	midiOutClose           *windows.LazyProc
	midiOutGetDevCaps      *windows.LazyProc
	midiOutGetErrorText    *windows.LazyProc
	midiOutGetNumDevs      *windows.LazyProc
	midiOutLongMsg         *windows.LazyProc
	midiOutOpen            *windows.LazyProc
	midiOutPrepareHeader   *windows.LazyProc
	midiOutShortMsg        *windows.LazyProc
	midiOutUnprepareHeader *windows.LazyProc
)

func init() {
	winmm = windows.NewLazySystemDLL("winmm.dll")
	midiInAddBuffer = winmm.NewProc("midiInAddBuffer")
	midiInClose = winmm.NewProc("midiInClose")
	midiInGetDevCaps = winmm.NewProc("midiInGetDevCapsW")
	midiInGetErrorText = winmm.NewProc("midiInGetErrorTextW")
	midiInGetNumDevs = winmm.NewProc("midiInGetNumDevs")
	midiInOpen = winmm.NewProc("midiInOpen")
	midiInPrepareHeader = winmm.NewProc("midiInPrepareHeader")
	midiInStart = winmm.NewProc("midiInStart")
	midiInStop = winmm.NewProc("midiInStop")
	midiInUnprepareHeader = winmm.NewProc("midiInUnprepareHeader")
	midiOutClose = winmm.NewProc("midiOutClose")
	midiOutGetDevCaps = winmm.NewProc("midiOutGetDevCapsW")
	midiOutGetErrorText = winmm.NewProc("midiOutGetErrorTextW")
	midiOutGetNumDevs = winmm.NewProc("midiOutGetNumDevs")
	midiOutLongMsg = winmm.NewProc("midiOutLongMsg")
	midiOutOpen = winmm.NewProc("midiOutOpen")
	midiOutPrepareHeader = winmm.NewProc("midiOutPrepareHeader")
	midiOutShortMsg = winmm.NewProc("midiOutShortMsg")
	midiOutUnprepareHeader = winmm.NewProc("midiOutUnprepareHeader")
}

func (e MidiInError) Error() string {
	var buffer [256]uint16
	r1, _, _ := midiInGetErrorText.Call(uintptr(e), uintptr(unsafe.Pointer(&buffer)), 256)
	if uint32(r1) != MMSYSERR_NOERROR {
		return fmt.Sprintf("%08x", uint32(e))
	}
	return windows.UTF16ToString(buffer[:])
}

func (e MidiOutError) Error() string {
	var buffer [256]uint16
	r1, _, _ := midiOutGetErrorText.Call(uintptr(e), uintptr(unsafe.Pointer(&buffer)), 256)
	if uint32(r1) != MMSYSERR_NOERROR {
		return fmt.Sprintf("%08x", uint32(e))
	}
	return windows.UTF16ToString(buffer[:])
}

func MidiInAddBuffer(hMidiIn uintptr, lpMidiInHdr *MIDIHDR) (err error) {
	r1, _, _ := midiInAddBuffer.Call(hMidiIn, uintptr(unsafe.Pointer(lpMidiInHdr)), unsafe.Sizeof(*lpMidiInHdr))
	if uint32(r1) != MMSYSERR_NOERROR {
		return MidiInError(r1)
	}
	return nil
}

func MidiInClose(hMidiIn uintptr) (err error) {
	r1, _, _ := midiInClose.Call(hMidiIn)
	if uint32(r1) != MMSYSERR_NOERROR {
		return MidiInError(r1)
	}
	return nil
}

func MidiInGetDevCaps(uDeviceID uintptr) (lpMidiInCaps *MIDIINCAPS, err error) {
	lpMidiInCaps = new(MIDIINCAPS)
	r1, _, _ := midiInGetDevCaps.Call(uDeviceID, uintptr(unsafe.Pointer(lpMidiInCaps)), uintptr(uint32(unsafe.Sizeof(*lpMidiInCaps))))
	if uint32(r1) != MMSYSERR_NOERROR {
		return nil, MidiInError(r1)
	}
	return lpMidiInCaps, nil
}

func MidiInGetNumDevs() uint32 {
	r1, _, _ := midiInGetNumDevs.Call()
	return uint32(r1)
}

func MidiInOpen(uDeviceID uint32, dwCallback uintptr, dwCallbackInstance uintptr, dwFlags uint32) (hMidiIn uintptr, err error) {
	r1, _, _ := midiInOpen.Call(uintptr(unsafe.Pointer(&hMidiIn)), uintptr(uDeviceID), dwCallback, dwCallbackInstance, uintptr(dwFlags))
	if uint32(r1) != MMSYSERR_NOERROR {
		return 0, MidiInError(r1)
	}
	return hMidiIn, nil
}

func MidiInPrepareHeader(hMidiIn uintptr, lpMidiInHdr *MIDIHDR) (err error) {
	r1, _, _ := midiInPrepareHeader.Call(hMidiIn, uintptr(unsafe.Pointer(lpMidiInHdr)), unsafe.Sizeof(*lpMidiInHdr))
	if uint32(r1) != MMSYSERR_NOERROR {
		return MidiInError(r1)
	}
	return nil
}

func MidiInStart(hMidiIn uintptr) (err error) {
	r1, _, _ := midiInStart.Call(hMidiIn)
	if uint32(r1) != MMSYSERR_NOERROR {
		return MidiInError(r1)
	}
	return nil
}

func MidiInUnprepareHeader(hMidiIn uintptr, lpMidiInHdr *MIDIHDR) (err error) {
	r1, _, _ := midiInUnprepareHeader.Call(hMidiIn, uintptr(unsafe.Pointer(lpMidiInHdr)), unsafe.Sizeof(*lpMidiInHdr))
	if uint32(r1) != MMSYSERR_NOERROR {
		return MidiInError(r1)
	}
	return nil
}

func MidiOutClose(hMidiOut uintptr) (err error) {
	r1, _, _ := midiOutClose.Call(hMidiOut)
	if uint32(r1) != MMSYSERR_NOERROR {
		return MidiOutError(r1)
	}
	return nil
}

func MidiOutGetDevCaps(uDeviceID uintptr) (lpMidiOutCaps *MIDIOUTCAPS, err error) {
	lpMidiOutCaps = new(MIDIOUTCAPS)
	r1, _, _ := midiOutGetDevCaps.Call(uDeviceID, uintptr(unsafe.Pointer(lpMidiOutCaps)), uintptr(uint32(unsafe.Sizeof(*lpMidiOutCaps))))
	if uint32(r1) != MMSYSERR_NOERROR {
		return nil, MidiOutError(r1)
	}
	return lpMidiOutCaps, nil
}

func MidiOutGetNumDevs() uint32 {
	r1, _, _ := midiOutGetNumDevs.Call()
	return uint32(r1)
}

func MidiOutLongMsg(hmo uintptr, lpMidiOutHdr *MIDIHDR) (err error) {
	r1, _, _ := midiOutLongMsg.Call(hmo, uintptr(unsafe.Pointer(lpMidiOutHdr)), unsafe.Sizeof(*lpMidiOutHdr))
	if uint32(r1) != MMSYSERR_NOERROR {
		return MidiOutError(r1)
	}
	return nil
}

func MidiOutOpen(uDeviceID uint32, dwCallback uintptr, dwCallbackInstance uintptr, dwFlags uint32) (hMidiOut uintptr, err error) {
	r1, _, _ := midiOutOpen.Call(uintptr(unsafe.Pointer(&hMidiOut)), uintptr(uDeviceID), dwCallback, dwCallbackInstance, uintptr(dwFlags))
	if uint32(r1) != MMSYSERR_NOERROR {
		return 0, MidiOutError(r1)
	}
	return hMidiOut, nil
}

func MidiOutPrepareHeader(hMidiOut uintptr, lpMidiOutHdr *MIDIHDR) (err error) {
	r1, _, _ := midiOutPrepareHeader.Call(hMidiOut, uintptr(unsafe.Pointer(lpMidiOutHdr)), unsafe.Sizeof(*lpMidiOutHdr))
	if uint32(r1) != MMSYSERR_NOERROR {
		return MidiOutError(r1)
	}
	return nil
}

func MidiOutShortMsg(hmo uintptr, dwMsg uint32) (err error) {
	r1, _, _ := midiOutShortMsg.Call(hmo, uintptr(dwMsg))
	if uint32(r1) != MMSYSERR_NOERROR {
		return MidiOutError(r1)
	}
	return nil
}

func MidiOutUnprepareHeader(hMidiOut uintptr, lpMidiOutHdr *MIDIHDR) (err error) {
	r1, _, _ := midiOutUnprepareHeader.Call(hMidiOut, uintptr(unsafe.Pointer(lpMidiOutHdr)), unsafe.Sizeof(*lpMidiOutHdr))
	if uint32(r1) != MMSYSERR_NOERROR {
		return MidiOutError(r1)
	}
	return nil
}
