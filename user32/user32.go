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

package user32

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	HWND_MESSAGE          uintptr = ^uintptr(2)
	INPUT_MOUSE           uint32  = 0
	INPUT_KEYBOARD        uint32  = 1
	INPUT_HARDWARE        uint32  = 2
	MAPVK_VK_TO_VSC       uint32  = 0
	MAPVK_VSC_TO_VK       uint32  = 1
	MAPVK_VK_TO_CHAR      uint32  = 2
	MAPVK_VSC_TO_VK_EX    uint32  = 3
	KEYEVENTF_EXTENDEDKEY uint32  = 0x1
	KEYEVENTF_KEYUP       uint32  = 0x2
	KEYEVENTF_UNICODE     uint32  = 0x4
	KEYEVENTF_SCANCODE    uint32  = 0x8
	VK_SHIFT              uint16  = 0x10
	VK_CONTROL            uint16  = 0x11
	VK_MENU               uint16  = 0x12
)

type INPUT_KEYBDINPUT struct {
	Type    uint32
	Ki      KEYBDINPUT
	Padding [2]uint32
}

type KEYBDINPUT struct {
	WVk         uint16
	WScan       uint16
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

type MSG struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct {
		X int32
		Y int32
	}
}

type WNDCLASSEX struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  uintptr
	LpszClassName uintptr
	HIconSm       uintptr
}

type WindowProc func(hWnd uintptr, uMsg uint32, wParam, lParam uintptr) uintptr

var (
	user32           *windows.LazyDLL
	createWindowEx   *windows.LazyProc
	defWindowProc    *windows.LazyProc
	dispatchMessage  *windows.LazyProc
	getMessage       *windows.LazyProc
	mapVirtualKey    *windows.LazyProc
	translateMessage *windows.LazyProc
	registerClassEx  *windows.LazyProc
	sendInput        *windows.LazyProc
	//DefWindowProcAddr uintptr
)

func init() {
	user32 = windows.NewLazySystemDLL("User32.dll")
	createWindowEx = user32.NewProc("CreateWindowExW")
	defWindowProc = user32.NewProc("DefWindowProcW")
	dispatchMessage = user32.NewProc("DispatchMessageW")
	getMessage = user32.NewProc("GetMessageW")
	mapVirtualKey = user32.NewProc("MapVirtualKeyW")
	translateMessage = user32.NewProc("TranslateMessage")
	registerClassEx = user32.NewProc("RegisterClassExW")
	sendInput = user32.NewProc("SendInput")
	//DefWindowProcAddr = defWindowProc.Addr()
}

func CreateWindowEx(dwExStyle uint32, lpClassName uintptr, lpWindowName string, dwStyle uint32, x, y, nWidth, nHeight int32, hWndParent, hMenu, hInstance uintptr, lpParam unsafe.Pointer) (hWnd uintptr, err error) {
	szWindowName, err := windows.UTF16PtrFromString(lpWindowName)
	if err != nil {
		return
	}
	hWnd, _, err = createWindowEx.Call(uintptr(dwExStyle), lpClassName, uintptr(unsafe.Pointer(szWindowName)), uintptr(dwStyle), uintptr(x), uintptr(y), uintptr(nWidth), uintptr(nHeight), hWndParent, hMenu, hInstance, uintptr(lpParam))
	if hWnd == 0 {
		return
	}
	return hWnd, nil
}

func DefWindowProc(hWnd uintptr, uMsg uint32, wParam, lParam uintptr) (lResult uintptr) {
	lResult, _, _ = defWindowProc.Call(hWnd, uintptr(uMsg), wParam, lParam)
	return
}

func DispatchMessage(lpMsg *MSG) (lResult uintptr) {
	lResult, _, _ = dispatchMessage.Call(uintptr(unsafe.Pointer(lpMsg)))
	return
}

func GetMessage(hWnd uintptr, wMsgFilterMin, wMsgFilterMax uint32) (bResult int32, lpMsg *MSG, err error) {
	lpMsg = new(MSG)
	r1, _, err := getMessage.Call(uintptr(unsafe.Pointer(lpMsg)), hWnd, uintptr(wMsgFilterMin), uintptr(wMsgFilterMax))
	bResult = int32(r1)
	if bResult == -1 {
		return
	}
	return bResult, lpMsg, nil
}

func MapVirtualKey(uCode, uMapType uint32) (uResult uint32) {
	r1, _, _ := mapVirtualKey.Call(uintptr(uCode), uintptr(uMapType))
	uResult = uint32(r1)
	return
}

func TranslateMessage(lpMsg *MSG) (bResult bool) {
	r1, _, _ := translateMessage.Call(uintptr(unsafe.Pointer(lpMsg)))
	return r1 != 0
}

func RegisterClassEx(style uint32, lpfnWndProc WindowProc, cbClsExtra, cbWndExtra int32, hInstance, hIcon, hCursor, hbrBackground, lpszMenuName uintptr, lpszClassName string, hIconSm uintptr) (hClass uint16, err error) {
	lpwcx := new(WNDCLASSEX)
	lpwcx.CbSize = uint32(unsafe.Sizeof(*lpwcx))
	lpwcx.Style = style
	if lpfnWndProc != nil {
		lpwcx.LpfnWndProc = windows.NewCallback(lpfnWndProc)
	}
	lpwcx.CbClsExtra = cbClsExtra
	lpwcx.CbWndExtra = cbWndExtra
	lpwcx.HInstance = hInstance
	lpwcx.HIcon = hIcon
	lpwcx.HCursor = hCursor
	lpwcx.HbrBackground = hbrBackground
	lpwcx.LpszMenuName = lpszMenuName
	szClassName, err := windows.UTF16PtrFromString(lpszClassName)
	if err != nil {
		return
	}
	lpwcx.LpszClassName = uintptr(unsafe.Pointer(szClassName))
	lpwcx.HIconSm = hIconSm
	r1, _, err := registerClassEx.Call(uintptr(unsafe.Pointer(lpwcx)))
	hClass = uint16(r1)
	if hClass == 0 {
		return
	}
	return hClass, nil
}

func SendInput(pInputs []INPUT_KEYBDINPUT) (nResult uint32, err error) {
	r1, _, err := sendInput.Call(uintptr(uint32(len(pInputs))), uintptr(unsafe.Pointer(&pInputs[0])), unsafe.Sizeof(pInputs[0]))
	nResult = uint32(r1)
	if nResult != uint32(len(pInputs)) {
		return
	}
	return nResult, nil
}
