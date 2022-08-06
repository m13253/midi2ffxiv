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
	"log"
	"os"
	"runtime"
	"sync"
	"syscall"
	"time"
	"unsafe"

	actionqueue "github.com/m13253/actionqueue-go"
	cgc "github.com/m13253/cgc-go"

	"github.com/m13253/midi2ffxiv/kernel32"
	"github.com/m13253/midi2ffxiv/user32"
	"github.com/m13253/midi2ffxiv/winmm"
)

var versionInfo string

type application struct {
	preset

	Quit             context.CancelFunc
	NtpGoro          cgc.Executor
	MidiRealtimeGoro cgc.Executor
	MidiPlaybackGoro cgc.Executor
	KeystrokeGoro    cgc.Executor

	MidiInDevice                int
	MidiOutDevice               int
	MidiOutBank                 uint16
	MidiOutPatch                uint8
	MidiOutTranspose            int
	MidiPlaybackTrack           uint16
	MidiPlaybackOffset          time.Duration
	MidiPlaybackSchedule        time.Time
	MidiPlaybackScheduleEnabled bool
	MidiPlaybackLoop            time.Duration
	MidiPlaybackLoopEnabled     bool
	NtpSyncServer               string
	NtpLastSync                 time.Time
	NtpClockOffset              time.Duration
	NtpMaxDeviation             time.Duration

	ctx context.Context

	hWnd        uintptr
	hMidiIn     uintptr
	hMidiOut    uintptr
	sysexBuffer [2]*winmm.MIDIHDR

	midiOutQueue   *actionqueue.Queue
	keystrokeQueue *actionqueue.Queue

	keyStatus *keystrokeStatus

	midiFileBuffer *midiFileBuffer

	ntpMutex *sync.RWMutex
}

func main() {
	app := new(application)
	os.Exit(app.run(os.Args))
}

func (app *application) run(args []string) int {
	runtime.LockOSThread()
	_ = kernel32.SetPriorityClass(kernel32.GetCurrentProcess(), kernel32.HIGH_PRIORITY_CLASS)

	fmt.Println("MIDI2FFXIV")
	if versionInfo != "" {
		fmt.Printf("Version: %s\n", versionInfo)
	}
	fmt.Println("Copyright (c) 2018 Star Brilliant")
	fmt.Println("=================================")
	fmt.Println()

	app.preset = defaultPreset
	err := app.parseConfigFile()
	if err != nil {
		log.Println("Error: ", err)
		return app.delayReturn(1)
	}

	app.ctx, app.Quit = context.WithCancel(context.Background())

	app.KeystrokeGoro = cgc.NewBuffered(1)
	app.MidiRealtimeGoro = cgc.NewBuffered(1)
	app.NtpGoro = cgc.NewBuffered(1)
	app.MidiPlaybackGoro = cgc.NewBuffered(1)

	app.MidiInDevice = -1
	app.MidiOutDevice = -1
	app.MidiOutBank = 0
	app.MidiOutPatch = 46
	app.MidiOutTranspose = 0
	app.MidiPlaybackTrack = 1

	app.midiOutQueue = actionqueue.New()
	app.midiOutQueue.Run(app.ctx)
	app.keystrokeQueue = actionqueue.New()
	app.keystrokeQueue.Run(app.ctx)

	app.ntpMutex = new(sync.RWMutex)

	err = app.startWebServer()
	if err != nil {
		log.Println("Error: ", err)
		return app.delayReturn(1)
	}

	hWndClass, err := user32.RegisterClassEx(0, app.windowProc, 0, 0, 0, 0, 0, 0, 0, "midi2ffxiv", 0)
	if err != nil {
		log.Println("Error: ", err)
		return app.delayReturn(int(err.(syscall.Errno)))
	}
	app.hWnd, err = user32.CreateWindowEx(0, uintptr(hWndClass), "midi2ffxiv", 0, 0, 0, 0, 0, user32.HWND_MESSAGE, 0, 0, nil)
	if err != nil {
		log.Println("Error: ", err)
		return app.delayReturn(int(err.(syscall.Errno)))
	}
	if app.EmergencyStop != nil && app.EmergencyStop.VirtualKeyCode != 0 {
		hotkeyModifiers := user32.MOD_NOREPEAT
		if app.EmergencyStop.Ctrl {
			hotkeyModifiers |= user32.MOD_CONTROL
		}
		if app.EmergencyStop.Alt {
			hotkeyModifiers |= user32.MOD_ALT
		}
		if app.EmergencyStop.Shift {
			hotkeyModifiers |= user32.MOD_SHIFT
		}
		_, err := user32.RegisterHotKey(app.hWnd, 1, uint32(hotkeyModifiers), uint32(app.EmergencyStop.VirtualKeyCode))
		if err != nil {
			log.Println("Failed to register emergency stop hotkey, is another instance running?")
			log.Println("Error: ", err)
		}
	}

	go app.consumeStdin()
	go app.processKeystrokes()
	go app.processMidiPlayback()
	go app.processMidiRealtime()
	go app.processNTP()
	go app.waitForQuit()

	for {
		bResult, lpMsg, err := user32.GetMessage(app.hWnd, 0, 0)
		if err != nil {
			log.Println("Error: ", err)
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
			} else if event.EventType == kernel32.KEY_EVENT && event.KeyEvent.WVirtualKeyCode == 'P' && (event.KeyEvent.DwControlKeyState&(kernel32.LEFT_CTRL_PRESSED|kernel32.RIGHT_CTRL_PRESSED)) != 0 && (event.KeyEvent.DwControlKeyState&(kernel32.LEFT_ALT_PRESSED|kernel32.RIGHT_ALT_PRESSED)) != 0 && (event.KeyEvent.DwControlKeyState&kernel32.SHIFT_PRESSED) != 0 {
				log.Println("Stack trace requested")
				buf := make([]byte, 1024)
				for {
					n := runtime.Stack(buf, true)
					if n < len(buf) {
						fmt.Println(string(buf[:n]))
						break
					}
					buf = make([]byte, 2*len(buf))
				}
			}
		}
	}
}

func (app *application) delayReturn(code int) int {
	fmt.Fprint(os.Stderr, "\nPress Ctrl-C to exit...")
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
	case user32.WM_HOTKEY:
		log.Println("Emergency stop pressed!")
		app.MidiPlaybackGoro.SubmitNoWait(app.ctx, func(context.Context) (interface{}, error) {
			app.setMidiPlaybackScheduler(false, app.MidiPlaybackSchedule, app.MidiPlaybackLoopEnabled, app.MidiPlaybackLoop)
			return nil, nil
		})
	case winmm.MM_MIM_OPEN:
		// no-op
	case winmm.MM_MIM_CLOSE:
		// no-op
	case winmm.MM_MIM_DATA, winmm.MM_MIM_MOREDATA:
		midiEvent := []byte{byte(lParam), byte(lParam >> 8), byte(lParam >> 16)}
		app.MidiRealtimeGoro.SubmitNoWait(app.ctx, func(context.Context) (interface{}, error) {
			app.onMidiInEvent(midiEvent)
			return nil, nil
		})
	case winmm.MM_MIM_LONGDATA:
		midiHeader := (*winmm.MIDIHDR)(unsafe.Pointer(lParam))
		midiEvent := make([]byte, midiHeader.DwBytesRecorded)
		copy(midiEvent, (*[65536]byte)(unsafe.Pointer(midiHeader.LpData))[:midiHeader.DwBytesRecorded])
		app.MidiRealtimeGoro.Submit(app.ctx, func(context.Context) (interface{}, error) {
			app.onMidiInEvent(midiEvent)
			return nil, nil
		})
		err := winmm.MidiInAddBuffer(app.hMidiIn, midiHeader)
		if err != nil {
			log.Println("Error: ", err)
		}
	case winmm.MM_MIM_ERROR:
		midiEvent := []byte{byte(lParam), byte(lParam >> 8), byte(lParam >> 16)}
		log.Printf("Invalid MIDI message: %x\n", midiEvent)
	case winmm.MM_MIM_LONGERROR:
		midiHeader := (*winmm.MIDIHDR)(unsafe.Pointer(lParam))
		midiEvent := make([]byte, midiHeader.DwBytesRecorded)
		copy(midiEvent, (*[65536]byte)(unsafe.Pointer(midiHeader.LpData))[:midiHeader.DwBytesRecorded])
		log.Printf("Invalid MIDI message: %x\n", midiEvent)
		err := winmm.MidiInAddBuffer(app.hMidiIn, midiHeader)
		if err != nil {
			log.Println("Error: ", err)
		}
	default:
		return user32.DefWindowProc(hWnd, uMsg, wParam, lParam)
	}
	return 0
}
