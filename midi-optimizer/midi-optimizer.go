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
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/m13253/midimark"
)

const (
	SkillCooldown time.Duration = 125 * time.Millisecond
)

type NoteOnRecord struct {
	Event   *midimark.EventNoteOn
	OldTick int64
	OldTime time.Duration
	NewTick int64
	NewTime time.Duration
}

func warningCallback(err error) {
	log.Println(err)
}

func newProgressTicker() (*time.Ticker, <-chan time.Time) {
	ticker := time.NewTicker(100 * time.Millisecond)
	c := make(chan time.Time, 1)
	go func() {
		for t := range ticker.C {
			select {
			case c <- t:
			default:
			}
		}
	}()
	return ticker, c
}

func main() {
	var input, output *os.File
	var err error
	switch len(os.Args) {
	case 0:
		fmt.Print("Usage: midi-optimizer INPUT.mid [OUTPUT.mid]\n\n")
		os.Exit(1)
	case 1:
		fmt.Printf("Usage: %s INPUT.mid [OUTPUT.mid]\n\n", os.Args[0])
		os.Exit(1)
	case 2:
		input, err = os.Open(os.Args[1])
		if err != nil {
			log.Fatalln(err)
		}
		defer input.Close()
		output, err = os.Create(strings.TrimSuffix(os.Args[1], filepath.Ext(os.Args[1])) + "-ffxiv.mid")
		if err != nil {
			log.Fatalln(err)
		}
		defer output.Close()
	case 3:
		input, err = os.Open(os.Args[1])
		if err != nil {
			log.Fatalln(err)
		}
		defer input.Close()
		output, err = os.Create(os.Args[2])
		if err != nil {
			log.Fatalln(err)
		}
		defer output.Close()
	}
	seq, err := midimark.DecodeSequenceFromSMF(input, warningCallback)
	if err != nil {
		log.Fatalln(err)
	}
	progressTicker, progressChan := newProgressTicker()
	for trackID, mtrk := range seq.Tracks {
		records := make([]*NoteOnRecord, 0)
		for _, event := range mtrk.Events {
			if ev, ok := event.(*midimark.EventNoteOn); ok {
				tick := ev.AbsTick
				realtime := mtrk.ConvertAbsTickToDuration(tick)
				records = append(records, &NoteOnRecord{
					Event:   ev,
					OldTick: tick,
					OldTime: realtime,
					NewTick: tick,
					NewTime: realtime,
				})
			}
		}
		sort.Slice(records, func(i, j int) bool {
			return records[i].OldTick < records[j].OldTick ||
				(records[i].OldTick == records[j].OldTick && records[i].Event.Key < records[j].Event.Key) ||
				(records[i].OldTick == records[j].OldTick && records[i].Event.Key == records[j].Event.Key && records[i].Event.FilePosition == records[j].Event.FilePosition)
		})
		totalUnresolved := 0
		var round int64
		for round = 0; true; round++ {
			numUnresolved := 0
			left, right := -1, -1
			for i := 0; i < len(records)-1; i++ {
				if records[i+1].NewTime-records[i].NewTime < SkillCooldown {
					numUnresolved++
					if left < 0 || right < 0 || records[i+1].NewTime-records[i].NewTime < records[right].NewTime-records[left].NewTime {
						left = i
						right = i + 1
					}
				}
			}
			select {
			case <-progressChan:
				if numUnresolved != 1 {
					fmt.Printf("\rTrack %d/%d, round #%d, %d conflicts unresolved.  ", trackID+1, len(seq.Tracks), round, numUnresolved)
				} else {
					fmt.Printf("\rTrack %d/%d, round #%d, %d conflict unresolved.  ", trackID+1, len(seq.Tracks), round, numUnresolved)
				}
			default:
			}
			if totalUnresolved < numUnresolved {
				totalUnresolved = numUnresolved
			}
			if left < 0 || right < 0 {
				break
			}

			for left != 0 && records[left].NewTime-records[left-1].NewTime < SkillCooldown {
				left--
			}
			for right < len(records)-1 && records[right+1].NewTime-records[right].NewTime < SkillCooldown {
				right++
			}
			if records[left].NewTick > 0 {
				records[left].NewTick--
				records[right].NewTick++
			} else {
				records[left].NewTick++
				records[right].NewTick += 2
			}
			records[left].NewTime = mtrk.ConvertAbsTickToDuration(records[left].NewTick)
			records[right].NewTime = mtrk.ConvertAbsTickToDuration(records[right].NewTick)
		}
		if totalUnresolved != 1 {
			fmt.Printf("\rTrack %d/%d, round #%d, %d conflicts resolved.  \n", trackID+1, len(seq.Tracks), round, totalUnresolved)
		} else {
			fmt.Printf("\rTrack %d/%d, round #%d, %d conflict resolved.  \n", trackID+1, len(seq.Tracks), round, totalUnresolved)
		}
		maxTick := int64(0)
		for _, record := range records {
			if record.NewTick > maxTick {
				maxTick = record.NewTick
			}
			offset := record.NewTick - record.OldTick
			record.Event.AbsTick += offset
			if record.Event.RelatedNoteOff != nil {
				record.Event.RelatedNoteOff.AbsTick += offset
				if record.Event.RelatedNoteOff.AbsTick > maxTick {
					maxTick = record.Event.RelatedNoteOff.AbsTick
				}
			}
		}
		for i := 0; i < len(records)-1; i++ {
			if records[i].Event.RelatedNoteOff != nil && records[i].Event.RelatedNoteOff.AbsTick > records[i+1].NewTick && records[i+1].NewTick >= records[i].NewTick {
				records[i].Event.RelatedNoteOff.AbsTick = records[i+1].NewTick
			}
		}
		if len(mtrk.Events) != 0 && mtrk.Events[len(mtrk.Events)-1].Common().AbsTick < maxTick {
			mtrk.Events[len(mtrk.Events)-1].Common().AbsTick = maxTick
		}
		sort.Slice(mtrk.Events, func(i, j int) bool {
			return mtrk.Events[i].Common().AbsTick < mtrk.Events[j].Common().AbsTick || (mtrk.Events[i].Common().AbsTick == mtrk.Events[j].Common().AbsTick && mtrk.Events[i].Common().FilePosition < mtrk.Events[j].Common().FilePosition)
		})
		mtrk.ConvertAbsToDeltaTick()
	}
	progressTicker.Stop()
	err = seq.EncodeSMF(output)
	if err != nil {
		log.Fatalln(err)
	}
}
