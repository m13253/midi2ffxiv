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
	"log"
	"time"

	cgc "github.com/m13253/cgc-go"
	"github.com/m13253/midimark"
)

type midiFileBuffer struct {
	sequence       *midimark.Sequence
	nextEventIndex int
	nextEventTimer *time.Timer
	fastForward    bool
}

func (app *application) warningCallback(err error) {
	log.Println(err)
}

func (app *application) processMidiPlayback() {
	app.midiFileBuffer = &midiFileBuffer{
		nextEventTimer: time.NewTimer(0),
	}
	for {
		select {
		case r, ok := <-app.MidiPlaybackGoro:
			if !ok {
				return
			}
			cgc.RunOneRequest(app.ctx, r)
		case now := <-app.midiFileBuffer.nextEventTimer.C:
			app.playNextMidiEvent(now)
		case <-app.ctx.Done():
			return
		}
	}
}

func (app *application) setMidiPlaybackFile(midiFile io.ReadSeeker) error {
	var err error

	sequence, err := midimark.DecodeSequenceFromSMF(midiFile, app.warningCallback)
	if err != nil {
		return err
	}
	app.midiFileBuffer.sequence = sequence
	return nil
}

func (app *application) playNextMidiEvent(now time.Time) {
	if !app.MidiPlaybackScheduleEnabled {
		return
	}
	track := app.MidiPlaybackTrack
	if len(app.midiFileBuffer.sequence.Tracks) == 0 {
		log.Println("MIDI file contains no track.")
		return
	}
	if len(app.midiFileBuffer.sequence.Tracks) == 1 {
		track = 0
	}
	if int(track) >= len(app.midiFileBuffer.sequence.Tracks) {
		log.Printf("Invalid track number (%d), max %d.\n", app.MidiPlaybackTrack, len(app.midiFileBuffer.sequence.Tracks)-1)
		return
	}
	playbackProgress := now.Add(app.NtpClockOffset).Add(app.MidiPlaybackOffset).Add(app.ModifierCooldown).Sub(app.MidiPlaybackSchedule)
	if playbackProgress < 0 {
		app.midiFileBuffer.nextEventIndex = 0
		app.midiFileBuffer.nextEventTimer.Reset(-playbackProgress)
		if app.midiFileBuffer.fastForward {
			log.Println("Fast-forward off.")
			app.midiFileBuffer.fastForward = false
		}
		return
	}
	if app.MidiPlaybackLoopEnabled && app.MidiPlaybackLoop > 0 {
		playbackProgress %= app.MidiPlaybackLoop
	}
	index := app.midiFileBuffer.nextEventIndex
	thisTrack := app.midiFileBuffer.sequence.Tracks[track]
	if index >= len(thisTrack.Events) {
		if app.MidiPlaybackLoopEnabled {
			app.midiFileBuffer.nextEventIndex = 0
			waitTime := app.MidiPlaybackLoop - playbackProgress
			if waitTime < 0 {
				waitTime = 0
			}
			log.Printf("Will loop in %s\n", waitTime)
			app.midiFileBuffer.nextEventTimer.Reset(waitTime)
			if app.midiFileBuffer.fastForward {
				log.Println("Fast-forward off.")
				app.midiFileBuffer.fastForward = false
			}
		} else {
			log.Println("Track finished.")
			_ = app.MidiRealtimeGoro.SubmitNoWait(app.ctx, func(context.Context) (interface{}, error) {
				app.sendAllNoteOff(false)
				return nil, nil
			})
		}
		return
	}
	if index > 0 {
		lastNoteProgress := thisTrack.ConvertAbsTickToDuration(thisTrack.Events[index-1].Common().AbsTick)
		if lastNoteProgress > playbackProgress {
			app.resetMidiPlayback()
			return
		}
	}
	nextNoteProgress := thisTrack.ConvertAbsTickToDuration(thisTrack.Events[index].Common().AbsTick)
	if nextNoteProgress > playbackProgress {
		app.midiFileBuffer.nextEventTimer.Reset(nextNoteProgress - playbackProgress)
		if app.midiFileBuffer.fastForward {
			log.Println("Fast-forward off.")
			app.midiFileBuffer.fastForward = false
		}
		return
	}
	message, err := thisTrack.Events[index].EncodeRealtime()
	if err != nil {
		log.Println(err)
		// fall-through
	}
	if len(message) != 0 {
		app.addMidiEvent(&midiQueueEvent{
			Time:              now.Add(-playbackProgress).Add(nextNoteProgress),
			Message:           message,
			Realtime:          false,
			FastForward:       app.midiFileBuffer.fastForward,
			AlreadyTransposed: true,
		})
	}
	app.midiFileBuffer.nextEventIndex = index + 1
	app.midiFileBuffer.nextEventTimer.Reset(0)
}

func (app *application) setMidiPlaybackTrack(trackNumber uint16) {
	if app.MidiPlaybackTrack == trackNumber {
		return
	}
	app.MidiPlaybackTrack = trackNumber
	app.resetMidiPlayback()
}

func (app *application) setMidiPlaybackOffset(offset time.Duration) {
	fmt.Printf("Set playback offset to %s.\n", offset)
	app.MidiPlaybackOffset = offset
	track := app.MidiPlaybackTrack
	if len(app.midiFileBuffer.sequence.Tracks) == 1 {
		track = 0
	}
	if int(track) >= len(app.midiFileBuffer.sequence.Tracks) {
		return
	}
	if app.midiFileBuffer.nextEventIndex >= len(app.midiFileBuffer.sequence.Tracks[track].Events) {
		app.midiFileBuffer.nextEventIndex = 0
	}
	app.midiFileBuffer.nextEventTimer.Reset(0)
}

func (app *application) getMidiPlaybackScheduler() (enabled bool, startTime time.Time, loopEnabled bool, loopInterval time.Duration) {
	return app.MidiPlaybackScheduleEnabled, app.MidiPlaybackSchedule, app.MidiPlaybackLoopEnabled, app.MidiPlaybackLoop
}

func (app *application) setMidiPlaybackScheduler(enabled bool, startTime time.Time, loopEnabled bool, loopInterval time.Duration) {
	app.MidiPlaybackScheduleEnabled = enabled
	app.MidiPlaybackSchedule = startTime
	app.MidiPlaybackLoopEnabled = loopEnabled
	app.MidiPlaybackLoop = loopInterval
	app.resetMidiPlayback()
}

func (app *application) resetMidiPlayback() {
	log.Println("Reset playback.")
	_ = app.MidiRealtimeGoro.SubmitNoWait(app.ctx, func(context.Context) (interface{}, error) {
		app.sendAllNoteOff(false)
		return nil, nil
	})
	app.midiFileBuffer.nextEventIndex = 0
	app.midiFileBuffer.nextEventTimer.Reset(0)
	if !app.midiFileBuffer.fastForward {
		log.Println("Fast-forward on.")
		app.midiFileBuffer.fastForward = true
	}
}
