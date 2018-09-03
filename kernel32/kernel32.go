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

package kernel32

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	INVALID_HANDLE_VALUE uintptr = ^uintptr(0)

	ABOVE_NORMAL_PRIORITY_CLASS uint32 = 0x00008000
	BELOW_NORMAL_PRIORITY_CLASS uint32 = 0x00004000
	HIGH_PRIORITY_CLASS         uint32 = 0x00000080
	IDLE_PRIORITY_CLASS         uint32 = 0x00000040
	NORMAL_PRIORITY_CLASS       uint32 = 0x00000020
	REALTIME_PRIORITY_CLASS     uint32 = 0x00000100

	STD_INPUT_HANDLE  uint32 = ^uint32(10 - 1)
	STD_OUTPUT_HANDLE uint32 = ^uint32(11 - 1)
	STD_ERROR_HANDLE  uint32 = ^uint32(12 - 1)

	ENABLE_PROCESSED_INPUT uint32 = 0x0001

	KEY_EVENT          uint16 = 0x0001
	LEFT_CTRL_PRESSED  uint32 = 0x0008
	RIGHT_CTRL_PRESSED uint32 = 0x0004
)

type INPUT_RECORD_KEY_EVENT struct {
	EventType uint16
	KeyEvent  KEY_EVENT_RECORD
}

type KEY_EVENT_RECORD struct {
	BKeyDown          uint32
	WRepeatCount      uint16
	WVirtualKeyCode   uint16
	WVirtualScanCode  uint16
	UnicodeChar       uint16
	DwControlKeyState uint32
}

var (
	kernel32          *windows.LazyDLL
	getConsoleMode    *windows.LazyProc
	getCurrentProcess *windows.LazyProc
	getStdHandle      *windows.LazyProc
	readConsoleInput  *windows.LazyProc
	setConsoleMode    *windows.LazyProc
	setPriorityClass  *windows.LazyProc
)

func init() {
	kernel32 = windows.NewLazySystemDLL("Kernel32.dll")
	getConsoleMode = kernel32.NewProc("GetConsoleMode")
	getCurrentProcess = kernel32.NewProc("GetCurrentProcess")
	getStdHandle = kernel32.NewProc("GetStdHandle")
	readConsoleInput = kernel32.NewProc("ReadConsoleInputW")
	setConsoleMode = kernel32.NewProc("SetConsoleMode")
	setPriorityClass = kernel32.NewProc("SetPriorityClass")
}

func GetConsoleMode(hConsoleHandle uintptr) (bResult bool, lpMode uint32, err error) {
	r1, _, err := getConsoleMode.Call(hConsoleHandle, uintptr(unsafe.Pointer(&lpMode)))
	if int32(r1) == 0 {
		return
	}
	return true, lpMode, nil
}

func GetStdHandle(nStdHandle uint32) (hStdHandle uintptr) {
	hStdHandle, _, _ = getStdHandle.Call(uintptr(nStdHandle))
	return
}

func GetCurrentProcess() (hProcess uintptr) {
	hProcess, _, _ = getCurrentProcess.Call()
	return
}

func ReadConsoleInput(hConsoleInput uintptr, lpBuffer []INPUT_RECORD_KEY_EVENT, nLength uint32) (bResult bool, lpNumberOfEventsRead uint32, err error) {
	r1, _, err := readConsoleInput.Call(hConsoleInput, uintptr(unsafe.Pointer(&lpBuffer[0])), uintptr(nLength), uintptr(unsafe.Pointer(&lpNumberOfEventsRead)))
	if int32(r1) == 0 {
		return
	}
	return true, lpNumberOfEventsRead, nil
}

func SetConsoleMode(hConsoleHandle uintptr, dwMode uint32) (bResult bool, err error) {
	r1, _, err := setConsoleMode.Call(hConsoleHandle, uintptr(dwMode))
	if int32(r1) == 0 {
		return
	}
	return true, nil
}

func SetPriorityClass(hProcess uintptr, dwPriorityClass uint32) (err error) {
	r1, _, err := setPriorityClass.Call(hProcess, uintptr(dwPriorityClass))
	if int32(r1) == 0 {
		return
	}
	return nil
}
