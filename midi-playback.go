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
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/algoGuy/EasyMIDI/smf"
	"github.com/algoGuy/EasyMIDI/smfio"
	"github.com/algoGuy/EasyMIDI/vlq"
	cgc "github.com/m13253/cgc-go"
)

type midiFileBuffer struct {
	MidiTracks     []midiFileTrack
	TempoTable     []tempoEntry
	TicksPerBeat   uint16
	nextEventIndex int
	nextEventTimer *time.Timer
	fastForward    bool
}

type midiFileTrack []*midiFileEvent

type midiFileEvent struct {
	TicksElapsed int64
	Microseconds midiFileAbsoluteTime
	Message      []byte
}

type midiFileAbsoluteTime struct {
	Numerator   int64
	Denominator uint16 // = TicksPerBeat
}

type tempoEntry struct {
	TicksElapsed        int64
	MicrosecondsPerBeat uint32
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
			_ = cgc.RunOneRequest(app.ctx, r)
		case now := <-app.midiFileBuffer.nextEventTimer.C:
			app.playNextMidiEvent(now)
		case <-app.ctx.Done():
			return
		}
	}
}

func (app *application) setMidiPlaybackFile(midiFile io.Reader) error {
	var err error
	parsedFile, err := smfio.Read(midiFile)
	if err != nil {
		return err
	}
	midiTracks := make([]midiFileTrack, parsedFile.GetTracksNum())
	tempoTable := []tempoEntry{}
	division := parsedFile.GetDivision()
	if division.IsSMTPE() {
		return errors.New("MIDI with SMTPE timestamps is unsupported")
	}
	for trackID := range midiTracks {
		if parsedFile.GetFormat() == smf.Format2 {
			tempoTable = []tempoEntry{}
		}
		parsedTrack := parsedFile.GetTrack(uint16(trackID))
		track := make([]*midiFileEvent, 0, parsedTrack.Len())

		ticks := int64(0)
		msNumerator := int64(0)
		msDemonimator := division.GetTicks()
		msPerBeat := uint32(500000)
		nextTempoEntry := 0

		for it := parsedTrack.GetIterator(); it.MoveNext(); {
			event := it.GetValue()
			delta := int64(event.GetDTime())

			for nextTempoEntry < len(tempoTable) && ticks+delta > tempoTable[nextTempoEntry].TicksElapsed {
				delta -= tempoTable[nextTempoEntry].TicksElapsed - ticks
				msNumerator += (tempoTable[nextTempoEntry].TicksElapsed - ticks) * int64(msPerBeat)
				msPerBeat = tempoTable[nextTempoEntry].MicrosecondsPerBeat
				ticks = tempoTable[nextTempoEntry].TicksElapsed
				nextTempoEntry++
			}

			msNumerator += delta * int64(msPerBeat)
			ticks += delta

			var message []byte

			switch event.(type) {
			case *smf.MIDIEvent:
				message = []byte{event.GetStatus()}
				message = append(message, event.GetData()...)

			case *smf.MetaEvent:
				message = []byte{smf.MetaStatus, event.(*smf.MetaEvent).GetMetaType()}
				message = append(message, vlq.GetBytes(uint32(len(event.GetData())))...)
				message = append(message, event.GetData()...)

				if len(message) > 2 && message[1] == smf.MetaSetTempo {
					if len(message) != 6 || message[2] != 3 {
						return errors.New("Unrecognized MIDI tempo settings")
					}
					msPerBeat = (uint32(message[3]) << 16) | (uint32(message[4]) << 8) | uint32(message[5])
					tempoTable = append(tempoTable, tempoEntry{
						TicksElapsed:        ticks,
						MicrosecondsPerBeat: msPerBeat,
					})
				}

			case *smf.SysexEvent:
				message = []byte{event.GetStatus()}
				message = append(message, vlq.GetBytes(uint32(len(event.GetData())))...)
				message = append(message, event.GetData()...)

			default:
				return errors.New("Unrecognized MIDI event")
			}

			track = append(track, &midiFileEvent{
				TicksElapsed: ticks,
				Microseconds: midiFileAbsoluteTime{
					msNumerator,
					msDemonimator,
				},
				Message: message,
			})
		}

		midiTracks[trackID] = track
	}
	app.midiFileBuffer.MidiTracks = midiTracks
	app.midiFileBuffer.TempoTable = tempoTable
	app.midiFileBuffer.TicksPerBeat = division.GetTicks()
	return nil
}

func (app *application) playNextMidiEvent(now time.Time) {
	if !app.MidiPlaybackScheduleEnabled {
		return
	}
	if int(app.MidiPlaybackTrack) >= len(app.midiFileBuffer.MidiTracks) {
		log.Printf("Invalid track number (%d), max %d.\n", app.MidiPlaybackTrack, len(app.midiFileBuffer.MidiTracks)-1)
		return
	}
	playbackProgress := now.Add(app.NtpClockOffset).Add(app.MidiPlaybackOffset).Sub(app.MidiPlaybackSchedule)
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
	thisTrack := app.midiFileBuffer.MidiTracks[app.MidiPlaybackTrack]
	if index >= len(thisTrack) {
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
		lastNoteProgress := thisTrack[index-1].Microseconds.Duration()
		if lastNoteProgress > playbackProgress {
			app.resetMidiPlayback()
			return
		}
	}
	nextNoteProgress := thisTrack[index].Microseconds.Duration()
	if nextNoteProgress > playbackProgress {
		app.midiFileBuffer.nextEventTimer.Reset(nextNoteProgress - playbackProgress)
		if app.midiFileBuffer.fastForward {
			log.Println("Fast-forward off.")
			app.midiFileBuffer.fastForward = false
		}
		return
	}
	app.addMidiEvent(&midiQueueEvent{
		Time:              now.Add(-playbackProgress).Add(nextNoteProgress),
		Message:           thisTrack[index].Message,
		Realtime:          false,
		FastForward:       app.midiFileBuffer.fastForward,
		AlreadyTransposed: true,
	})
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
	if int(app.MidiPlaybackTrack) >= len(app.midiFileBuffer.MidiTracks) {
		return
	}
	if app.midiFileBuffer.nextEventIndex >= len(app.midiFileBuffer.MidiTracks[app.MidiPlaybackTrack]) {
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

func (m midiFileAbsoluteTime) Duration() time.Duration {
	return time.Duration(m.Numerator) * time.Microsecond / time.Duration(m.Denominator)
}
