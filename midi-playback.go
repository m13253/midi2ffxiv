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
	"errors"
	"io"
	"time"

	"github.com/algoGuy/EasyMIDI/smf"
	"github.com/algoGuy/EasyMIDI/smfio"
	cgc "github.com/m13253/cgc-go"
)

type midiFileBuffer struct {
	MidiTracks     []midiFileTrack
	TempoTable     []tempoEntry
	TicksPerBeat   uint16
	nextEventIndex int
	nextEventTimer *time.Timer
	loopTimer      *time.Timer
}

type midiFileTrack []*midiFileEvent

type midiFileEvent struct {
	TicksElapsed uint64
	Microseconds midiFileAbsoluteTime
	Message      []byte
}

type midiFileAbsoluteTime struct {
	Numerator   uint64
	Denominator uint16 // = TicksPerBeat
}

type tempoEntry struct {
	TicksElapsed        uint64
	MicrosecondsPerBeat uint32
}

func (app *application) processMidiPlayback() {
	app.midiFileBuffer = &midiFileBuffer{
		nextEventTimer: time.NewTimer(0),
		loopTimer:      time.NewTimer(0),
	}
	for {
		now := time.Now()
		select {
		case r, ok := <-app.MidiPlaybackGoro:
			if !ok {
				return
			}
			_ = cgc.RunOneRequest(app.ctx, r)
		case <-app.midiFileBuffer.nextEventTimer.C:
			if app.MidiPlaybackScheduleEnabled && now.After(app.MidiPlaybackSchedule) {
			}
		case <-app.midiFileBuffer.loopTimer.C:
			if app.MidiPlaybackScheduleEnabled && now.After(app.MidiPlaybackSchedule) && app.MidiPlaybackLoopEnabled {
				app.resetMidiPlayback()
			}
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

		ticks := uint64(0)
		msNumerator := uint64(0)
		msDemonimator := division.GetTicks()
		msPerBeat := uint32(500000)
		nextTempoEntry := 0

		for it := parsedTrack.GetIterator(); it.MoveNext(); {
			event := it.GetValue()
			delta := uint64(event.GetDTime())

			for nextTempoEntry < len(tempoTable) && ticks+delta > tempoTable[nextTempoEntry].TicksElapsed {
				delta -= tempoTable[nextTempoEntry].TicksElapsed - ticks
				msNumerator += (tempoTable[nextTempoEntry].TicksElapsed - ticks) * uint64(msPerBeat)
				msPerBeat = tempoTable[nextTempoEntry].MicrosecondsPerBeat
				ticks = tempoTable[nextTempoEntry].TicksElapsed
				nextTempoEntry++
			}

			msNumerator += delta * uint64(msPerBeat)
			ticks += delta

			status := event.GetStatus()
			data := event.GetData()
			if status == smf.NoteOnStatus && data[0] == 0 {
				status = smf.NoteOffStatus
			} else if status == smf.MetaStatus && len(data) > 1 && data[0] == smf.MetaSetTempo {
				if len(data) != 5 || data[1] != 3 {
					return errors.New("Unrecognized MIDI tempo settings")
				}
				msPerBeat = (uint32(data[2]) << 16) | (uint32(data[3]) << 8) | uint32(data[4])
				tempoTable = append(tempoTable, tempoEntry{
					TicksElapsed:        ticks,
					MicrosecondsPerBeat: msPerBeat,
				})
			}

			message := make([]byte, len(data)+1)
			message[0] = status
			copy(message[1:], data)
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

func (app *application) setMidiPlaybackTrack(trackNumber uint16) {
	app.MidiPlaybackTrack = trackNumber
}

func (app *application) setMidiPlaybackOffset(offset time.Duration) {
	app.MidiPlaybackOffset = offset
}

func (app *application) resetMidiPlayback() {
	app.midiFileBuffer.nextEventIndex = 0
	app.midiFileBuffer.nextEventTimer.Reset(0)
}

func (m midiFileAbsoluteTime) Duration() time.Duration {
	return time.Duration(m.Numerator) * time.Microsecond / time.Duration(m.Denominator)
}
