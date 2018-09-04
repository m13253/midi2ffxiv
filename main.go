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
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	cgc "github.com/m13253/cgc-go"

	"./kernel32"
	"./user32"
	"./winmm"
)

type application struct {
	preset

	Quit          context.CancelFunc
	MidiGoro      cgc.Executor
	PlaybackGoro  cgc.Executor
	KeystrokeGoro cgc.Executor

	MidiInDevice         int
	MidiOutDevice        int
	MidiOutBank          uint16
	MidiOutPatch         uint8
	MidiOutTranspose     int
	MidiPlaybackFile     io.ReadSeeker
	MidiPlaybackTrack    uint16
	MidiPlaybackOffset   float64
	MidiPlaybackSchedule time.Time
	MidiPlaybackLoop     time.Duration
	MidiPlaybackEnabled  bool
	NtpOffset            float64

	hWnd        uintptr
	hMidiIn     uintptr
	hMidiOut    uintptr
	sysexBuffer [2]*winmm.MIDIHDR

	ctx          context.Context
	midiBuffer   *midiBuffer
	pendingNotes chan *midiMessage
	lastNoteOn   *midiMessage
	keyStatus    *keystrokeStatus
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
	runtime.LockOSThread()
	_ = kernel32.SetPriorityClass(kernel32.GetCurrentProcess(), kernel32.HIGH_PRIORITY_CLASS)

	fmt.Println("MIDI2FFXIV")
	fmt.Println("Copyright (c) 2018 Star Brilliant")
	fmt.Println("=================================")
	fmt.Println()

	app.preset = defaultPreset

	app.ctx, app.Quit = context.WithCancel(context.Background())
	app.MidiGoro = cgc.NewBuffered(1)
	app.KeystrokeGoro = cgc.NewBuffered(1)

	app.MidiInDevice = -1
	app.MidiOutDevice = -1
	app.MidiOutBank = 0
	app.MidiOutPatch = 46
	app.MidiOutTranspose = 0

	app.pendingNotes = make(chan *midiMessage, 256)

	err := app.startWebServer()
	if err != nil {
		fmt.Println("Error: ", err)
		return app.delayReturn(1)
	}

	hWndClass, err := user32.RegisterClassEx(0, app.windowProc, 0, 0, 0, 0, 0, 0, 0, "midi2ffxiv", 0)
	if err != nil {
		fmt.Println("Error: ", err)
		return app.delayReturn(int(err.(syscall.Errno)))
	}
	app.hWnd, err = user32.CreateWindowEx(0, uintptr(hWndClass), "midi2ffxiv", 0, 0, 0, 0, 0, user32.HWND_MESSAGE, 0, 0, nil)
	if err != nil {
		fmt.Println("Error: ", err)
		return app.delayReturn(int(err.(syscall.Errno)))
	}

	go app.consumeStdin()
	go app.MidiGoro.RunLoop(app.ctx)
	go app.processKeystrokes()
	go app.processMidi()
	go app.waitForQuit()

	for {
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

	app.Quit()

	return 0
}

func (app *application) processMidi() {
	for {
		select {
		case nextNote := <-app.pendingNotes:
			now := time.Now()

			if (nextNote.Msg[0] == 0x90 || nextNote.Msg[0] == 0xa0) && now.Sub(nextNote.Time) > app.MaxNoteDelay {
				continue
			}

			if app.lastNoteOn != nil && ((nextNote.Msg[0] == 0x80 && nextNote.Msg[1] == app.lastNoteOn.Msg[1]) || nextNote.Msg[0] == 0x90) && now.Sub(app.lastNoteOn.Time) < app.SkillCooldown {
				time.Sleep(app.lastNoteOn.Time.Add(app.SkillCooldown).Sub(now))
				now = time.Now()
			}

			_, err := app.MidiGoro.Submit(app.ctx, func(context.Context) (interface{}, error) {
				return nil, app.sendMidiOutMessage(nextNote)
			})
			if err != nil {
				fmt.Println("Error: ", err)
			}

			if nextNote.Msg[0] == 0x80 || nextNote.Msg[0] == 0x90 {
				_, _ = app.KeystrokeGoro.Submit(app.ctx, func(context.Context) (interface{}, error) {
					app.produceKeystroke(nextNote)
					return nil, nil
				})
			}

			if nextNote.Msg[0] == 0x90 {
				nextNote.Time = now
				app.lastNoteOn = nextNote
			}
		case <-app.ctx.Done():
			break
		}
	}
}

func (app *application) consumeStdin() {
	hStdin := kernel32.GetStdHandle(kernel32.STD_INPUT_HANDLE)
	if hStdin == 0 || hStdin == kernel32.INVALID_HANDLE_VALUE {
		return
	}
	_, dwMode, err := kernel32.GetConsoleMode(hStdin)
	if err == nil {
		dwMode &= ^kernel32.ENABLE_PROCESSED_INPUT
		_, _ = kernel32.SetConsoleMode(hStdin, dwMode)
	}
	var lpBuffer [16]kernel32.INPUT_RECORD_KEY_EVENT
	for {
		bResult, lpNumberOfEventsRead, _ := kernel32.ReadConsoleInput(hStdin, lpBuffer[:], uint32(len(lpBuffer)))
		if !bResult || lpNumberOfEventsRead == 0 {
			break
		}
		for _, event := range lpBuffer[:lpNumberOfEventsRead] {
			if event.EventType == kernel32.KEY_EVENT && event.KeyEvent.WVirtualKeyCode == 'C' && (event.KeyEvent.DwControlKeyState&(kernel32.LEFT_CTRL_PRESSED|kernel32.RIGHT_CTRL_PRESSED)) != 0 {
				app.Quit()
			}
		}
	}
}

func (app *application) delayReturn(code int) int {
	fmt.Fprint(os.Stderr, "Press Ctrl-C to exit...")
	time.Sleep(1 * time.Minute)
	fmt.Fprintln(os.Stderr)
	return code
}

func (app *application) waitForQuit() {
	<-app.ctx.Done()
	_, _ = user32.PostMessage(app.hWnd, user32.WM_QUIT, 0, 0)
}

func (app *application) windowProc(hWnd uintptr, uMsg uint32, wParam, lParam uintptr) uintptr {
	switch uMsg {
	case winmm.MM_MIM_OPEN:
	case winmm.MM_MIM_CLOSE:
	case winmm.MM_MIM_DATA, winmm.MM_MIM_MOREDATA:
		midiMsg := []byte{byte(lParam), byte(lParam >> 8), byte(lParam >> 16)}
		app.MidiGoro.Submit(app.ctx, func(context.Context) (interface{}, error) {
			app.onMidiInMessage(midiMsg)
			return nil, nil
		})
	case winmm.MM_MIM_LONGDATA:
		midiHeader := (*winmm.MIDIHDR)(unsafe.Pointer(lParam))
		midiMsg := make([]byte, midiHeader.DwBytesRecorded)
		copy(midiMsg, (*[65536]byte)(unsafe.Pointer(midiHeader.LpData))[:midiHeader.DwBytesRecorded])
		app.MidiGoro.Submit(app.ctx, func(context.Context) (interface{}, error) {
			app.onMidiInMessage(midiMsg)
			return nil, nil
		})
		err := winmm.MidiInAddBuffer(app.hMidiIn, midiHeader)
		if err != nil {
			fmt.Println("Error: ", err)
		}
	case winmm.MM_MIM_ERROR:
		midiMsg := []byte{byte(lParam), byte(lParam >> 8), byte(lParam >> 16)}
		fmt.Printf("Invalid MIDI message: %x\n", midiMsg)
	case winmm.MM_MIM_LONGERROR:
		midiHeader := (*winmm.MIDIHDR)(unsafe.Pointer(lParam))
		midiMsg := make([]byte, midiHeader.DwBytesRecorded)
		copy(midiMsg, (*[65536]byte)(unsafe.Pointer(midiHeader.LpData))[:midiHeader.DwBytesRecorded])
		fmt.Printf("Invalid MIDI message: %x\n", midiMsg)
		err := winmm.MidiInAddBuffer(app.hMidiIn, midiHeader)
		if err != nil {
			fmt.Println("Error: ", err)
		}
	default:
		return user32.DefWindowProc(hWnd, uMsg, wParam, lParam)
	}
	return 0
}
