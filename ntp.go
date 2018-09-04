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
	"time"

	"github.com/beevik/ntp"
)

func (app *application) processNTP() {
	_ = app.NtpGoro.RunLoop(app.ctx)
}

func (app *application) syncTime(ntpServer string) error {
	app.ntpMutex.RLock()
	if time.Now().Sub(app.NtpLastSync) < app.NtpCooldown {
		app.ntpMutex.RUnlock()
		return errors.New("Time already synchronized")
	}
	app.ntpMutex.RUnlock()
	ntpOptions := ntp.QueryOptions{
		Timeout: app.NtpSyncTimeout,
	}
	clockOffset := time.Duration(0)
	rootDistance := time.Duration(0)
	for i := 0; i < 4; i++ {
		ntpResponse, err := ntp.QueryWithOptions(ntpServer, ntpOptions)
		if err != nil {
			return err
		}
		err = ntpResponse.Validate()
		if err != nil {
			return err
		}
		clockOffset += ntpResponse.ClockOffset
		if ntpResponse.RootDistance > rootDistance {
			rootDistance = ntpResponse.RootDistance
		}
	}
	app.ntpMutex.Lock()
	app.NtpClockOffset = clockOffset / 4
	app.NtpMaxDeviation = rootDistance
	app.NtpLastSync = time.Now()
	app.ntpMutex.Unlock()
	return nil
}

func (app *application) getNtpOffset() (bool, time.Duration, time.Duration) {
	app.ntpMutex.RLock()
	synced := !app.NtpLastSync.IsZero()
	offset := app.NtpClockOffset
	deviation := app.NtpMaxDeviation
	app.ntpMutex.RUnlock()
	return synced, offset, deviation
}
