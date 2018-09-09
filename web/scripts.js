(function () {

    "use strict";

    function requestHTTP(method, url, body, onLoad, onError) {
        var xhr = new XMLHttpRequest();
        xhr.onerror = function onerror(event) {
            return onError.bind(xhr)(event, "Connection error")
        }
        xhr.onabort = function onabort(event) {
            return onError.bind(xhr)(event, "Connection aborted");
        }
        xhr.onload = function onload(event) {
            if (xhr.status < 200 || xhr.status >= 300) {
                return onError.bind(xhr)(event, xhr.response);
            }
            try {
                return onLoad.bind(xhr)(event, JSON.parse(xhr.response));
            } catch (e) {
                return onError.bind(xhr)(event, e);
            }
        }
        xhr.open(method, url, true);
        xhr.send(body);
    }

    function reportError(message) {
        console.error(message);
        var content = document.createTextNode(message);
        var el = document.createElement("div");
        el.classList.add("float-message", "error");
        el.appendChild(content);
        var container = document.getElementById("float-container");
        container.insertBefore(el, container.firstElementChild);
        el.style.marginBottom = -el.offsetHeight + "px";
        el.classList.add("animated");
        setTimeout(function onHidden() {
            el.remove();
        }, 3400);
    }

    function reportMessage(message) {
        console.log(message);
        var content = document.createTextNode(message);
        var el = document.createElement("div");
        el.classList.add("float-message");
        el.appendChild(content);
        var container = document.getElementById("float-container");
        container.insertBefore(el, container.firstElementChild);
        el.style.marginBottom = -el.offsetHeight + "px";
        el.classList.add("animated");
        setTimeout(function onHidden() {
            el.remove();
        }, 3400);
    }

    function clearSelect(element) {
        while (element.firstElementChild) {
            element.firstElementChild.remove();
        }
    }

    function addSelectOption(element, text, value) {
        var option = document.createElement("option");
        option.text = text;
        option.value = value;
        element.appendChild(option);
    }

    var suppressEvents = false;

    function doVersionInfoUpdate() {
        requestHTTP("GET", "/version", null, function onLoad(event, response) {
            if (response["version_info"]) {
                document.getElementById("version-info").innerText = "Version " + response["version_info"] + ".";
            }
        }, function onError(event, error) {
        });
    }

    function doMidiInputRefresh(quiet) {
        requestHTTP("GET", "/midi-input-device", null, function onLoad(event, response) {
            var list = document.getElementById("midi-input-device");
            suppressEvents = true;
            try {
                clearSelect(list);
                addSelectOption(list, "(None)", "-1");
                var devices = response["devices"];
                for (var i = 0; i < devices.length; i++) {
                    addSelectOption(list, devices[i], i);
                }
                list.value = response["selected"];
            } finally {
                suppressEvents = false;
            }
            if (!quiet) {
                reportMessage("MIDI input device updated.");
            }
        }, function onError(event, error) {
            reportError("Failed to retrieve MIDI input devices.");
        });
    }

    function onMidiInputRefreshClicked() {
        return doMidiInputRefresh(false);
    }

    function onMidiInputDeviceChanged() {
        if (suppressEvents) { return; }
        var value = this.value;
        var text = this.options[this.selectedIndex].text;
        requestHTTP("PUT", "/midi-input-device", value, function onLoad(event, response) {
            reportMessage("MIDI input device changed to " + text + ".");
        }, function onError(event, error) {
            reportError(error);
        })
    }

    function doMidiOutputRefresh(quiet) {
        requestHTTP("GET", "/midi-output-device", null, function onLoad(event, response) {
            var list = document.getElementById("midi-output-device");
            suppressEvents = true;
            try {
                clearSelect(list);
                addSelectOption(list, "(None)", "-1");
                var devices = response["devices"];
                for (var i = 0; i < devices.length; i++) {
                    addSelectOption(list, devices[i], i);
                }
                list.value = response["selected"];
            } finally {
                suppressEvents = false;
            }
            if (!quiet) {
                reportMessage("MIDI output device updated.");
            }
        }, function onError(event, error) {
            reportError("Failed to retrieve MIDI output devices.");
        });
    }

    function onMidiOutputRefreshClicked() {
        return doMidiOutputRefresh(false);
    }

    function onMidiOutputDeviceChanged() {
        if (suppressEvents) { return; }
        var value = this.value;
        var text = this.options[this.selectedIndex].text;
        requestHTTP("PUT", "/midi-output-device", value, function onLoad(event, response) {
            reportMessage("MIDI output device changed to " + text + ".");
        }, function onError(event, error) {
            reportError(error);
        })
    }

    var instrumentTable = {
        "0:47": [0, 47, 0],
        "0:1": [0, 1, 12],
        "0:26": [0, 26, -12],
        "0:46": [0, 46, 0],
        "0:74": [0, 74, 12],
        "0:69": [0, 69, 12],
        "0:72": [0, 72, 0],
        "0:73": [0, 73, 24],
        "0:76": [0, 76, 12],
    }

    function doSynthInstrumentUpdate() {
        var synthBank = document.getElementById("synth-bank");
        var synthPatch = document.getElementById("synth-patch");
        var synthTranspose = document.getElementById("synth-transpose");
        var synthInstrument = document.getElementById("synth-instrument");
        var instrument = "";
        for (var i in instrumentTable) {
            if (instrumentTable[i][0] === +synthBank.value && instrumentTable[i][1] === +synthPatch.value && instrumentTable[i][2] === +synthTranspose.value) {
                instrument = i;
            }
        }
        suppressEvents = true;
        synthInstrument.value = instrument;
        suppressEvents = false;
    }

    function onSynthBankChanged() {
        if (suppressEvents) { return; }
        var value = this.value || "0";
        doSynthInstrumentUpdate();
        requestHTTP("PUT", "/midi-output-bank", value, function onLoad(event, response) {
            reportMessage("MIDI bank changed to " + value + ".");
        }, function onError(event, error) {
            reportError(error);
        })
    }

    function onSynthPatchChanged() {
        if (suppressEvents) { return; }
        var value = this.value || "47";
        doSynthInstrumentUpdate();
        requestHTTP("PUT", "/midi-output-patch", +value - 1, function onLoad(event, response) {
            reportMessage("MIDI patch changed to " + value + ".");
        }, function onError(event, error) {
            reportError(error);
        })
    }

    function onSynthTransposeChanged() {
        if (suppressEvents) { return; }
        var value = this.value || "0";
        doSynthInstrumentUpdate();
        requestHTTP("PUT", "/midi-output-transpose", value, function onLoad(event, response) {
            reportMessage("MIDI patch changed to " + value + ".");
        }, function onError(event, error) {
            reportError(error);
        })
    }

    function doSynthInstrumentRefresh() {
        if (suppressEvents) { return; }
        var synthBank = document.getElementById("synth-bank");
        var synthPatch = document.getElementById("synth-patch");
        var synthTranspose = document.getElementById("synth-transpose");
        var numLoaded = 0;
        var countLoad = function countLoad(event, response) {
            numLoaded++;
            if (numLoaded === 3) {
                doSynthInstrumentUpdate();
            }
        }
        requestHTTP("GET", "/midi-output-bank", null, function onLoad(event, response) {
            var value = response["bank"];
            suppressEvents = true;
            synthBank.value = value;
            suppressEvents = false;
            countLoad();
        }, function onError(event, error) {
        });
        requestHTTP("GET", "/midi-output-patch", null, function onLoad(event, response) {
            var value = response["patch"];
            suppressEvents = true;
            synthPatch.value = +value + 1;
            suppressEvents = false;
            countLoad();
        }, function onError(event, error) {
        });
        requestHTTP("GET", "/midi-output-transpose", null, function onLoad(event, response) {
            var value = response["transpose"];
            suppressEvents = true;
            synthTranspose.value = value;
            suppressEvents = false;
            countLoad();
        }, function onError(event, error) {
        });
    }

    function onSynthInstrumentChanged() {
        if (suppressEvents) { return; }
        var synthBank = document.getElementById("synth-bank");
        var synthPatch = document.getElementById("synth-patch");
        var synthTranspose = document.getElementById("synth-transpose");
        var value = this.value;
        var text = this.options[this.selectedIndex].text;
        var instrument = instrumentTable[value];
        if (instrument) {
            suppressEvents = true;
            synthBank.value = instrument[0];
            synthPatch.value = instrument[1];
            synthTranspose.value = instrument[2];
            suppressEvents = false;
        }

        var numLoaded = 0;
        var numFailed = 0;
        var countLoad = function countLoad(event, response) {
            numLoaded++;
            if (numLoaded === 3) {
                reportMessage("MIDI instrument changed to " + text + ".");
            }
        }
        var countError = function countError(event, error) {
            numFailed++;
            if (numFailed === 1) {
                reportError(error);
            }
        }
        requestHTTP("PUT", "/midi-output-bank", synthBank.value, countLoad, countError);
        requestHTTP("PUT", "/midi-output-patch", +synthPatch.value - 1, countLoad, countError);
        requestHTTP("PUT", "/midi-output-transpose", synthTranspose.value, countLoad, countError);
    }

    function doNTPServerUpdate() {
        requestHTTP("GET", "/ntp-sync-server", null, function onLoad(event, response) {
            var server = response["server"];
            if (server !== "0.beevik-ntp.pool.ntp.org") {
                document.getElementById("ntp-server").value = server;
            }
        }, function onError(event, error) {
        });
    }

    var serverTime = {
        "synced": false,
        "offset": 0,
        "max_deviation": 0,
    };

    function displayServerTime() {
        var el = document.getElementById("current-time");
        try {
            var now = new Date(Date.now() + serverTime["offset"] * 1000);
            var hours = now.getHours();
            var minutes = now.getMinutes();
            var seconds = now.getSeconds();
            var milliseconds = now.getMilliseconds();
            hours = hours < 10 ? "0" + hours : "" + hours;
            minutes = minutes < 10 ? "0" + minutes : "" + minutes;
            seconds = seconds < 10 ? "0" + seconds : "" + seconds;
            milliseconds = milliseconds < 100 ? milliseconds < 10 ? "00" + milliseconds : "0" + milliseconds : "" + milliseconds;
            var maxDeviation = serverTime["synced"] ? "\xb1" + Math.ceil(serverTime["max_deviation"] * 1000) + " ms" : "unsynced"
            el.value = hours + " : " + minutes + " : " + seconds + "." + milliseconds + " (" + maxDeviation + ")";
        } catch (e) {
            el.value = "-- : -- : --.--- (unsynced)";
        }
        requestAnimationFrame(displayServerTime);
    }

    function doUpdateServerTime() {
        var before = Date.now();
        requestHTTP("GET", "/current-time", null, function onLoad(event, response) {
            var after = Date.now();
            var rtt = (after - before) * 0.001;
            response["time"] += rtt / 2;
            response["offset"] = response["time"] - after * 0.001;
            response["max_deviation"] += rtt;
            serverTime = response;
        }, function onError(event, error) {
            serverTime["synced"] = false;
        });
    }

    function updateAllStates(state) {
        switch (state) {
            case 0:
                doVersionInfoUpdate();
                doMidiInputRefresh(true);
                doMidiOutputRefresh(true);
                doSynthInstrumentRefresh();
                doNTPServerUpdate();
                doUpdateServerTime();
                doMIDITrackNumberRefresh();
                doMIDIOffsetMsRefresh();
                doSchedulerRefresh();
                return setTimeout(updateAllStates, 1000, 1);
            case 1:
                doVersionInfoUpdate();
                return setTimeout(updateAllStates, 1000, 2);
            case 2:
                doMidiInputRefresh(true);
                doMidiOutputRefresh(true);
                return setTimeout(updateAllStates, 1000, 3);
            case 3:
                if (document.activeElement !== document.getElementById("synth-bank") && document.activeElement !== document.getElementById("synth-patch") && document.activeElement !== document.getElementById("synth-transpose")) {
                    doSynthInstrumentRefresh();
                }
                return setTimeout(updateAllStates, 1000, 4);
            case 4:
                doUpdateServerTime();
                return setTimeout(updateAllStates, 1000, 5);
            case 5:
                if (document.activeElement !== document.getElementById("midi-track-number")) {
                    doMIDITrackNumberRefresh();
                }
                if (document.activeElement !== document.getElementById("midi-offset-ms")) {
                    doMIDIOffsetMsRefresh();
                }
                return setTimeout(updateAllStates, 1000, 6);
            case 6:
                if (document.activeElement !== document.getElementById("sched-start-time") && document.activeElement !== document.getElementById("sched-loop-interval")) {
                    doSchedulerRefresh();
                }
                return setTimeout(updateAllStates, 1000, 1);
        }
    }

    function onNTPSyncClicked() {
        var button = this;
        var server = document.getElementById("ntp-server").value || "0.beevik-ntp.pool.ntp.org";
        button.disabled = true;
        setTimeout(function onTimeout() {
            button.disabled = false;
        }, 4000);
        requestHTTP("PUT", "/ntp-sync-server", server, function onLoad(event, response) {
            doUpdateServerTime();
            reportMessage("Time synced to " + server);
        }, function onError(event, error) {
            reportError(error);
        });
    }

    function onCurrentTimeCopyClicked() {
        var el = document.getElementById("sched-start-time");
        var now = new Date(Date.now() + serverTime["offset"] * 1000 + 5000);
        var hours = now.getHours();
        var minutes = now.getMinutes();
        var seconds = now.getSeconds();
        hours = hours < 10 ? "0" + hours : "" + hours;
        minutes = minutes < 10 ? "0" + minutes : "" + minutes;
        seconds = seconds < 10 ? "0" + seconds : "" + seconds;
        el.value = hours + " : " + minutes + " : " + seconds;
        reportMessage("Scheduled time is set to 5 seconds later.");
    }

    function onMIDIFileChanged() {
        if (this.files.length > 0) {
            var file = this.files[0];
            requestHTTP("PUT", "/midi-playback-file", file, function onLoad(event, response) {
                reportMessage("MIDI file loaded: " + file.name);
            }, function onError(event, error) {
                reportError(error);
            });
        }
    }

    function doMIDITrackNumberRefresh() {
        requestHTTP("GET", "/midi-playback-track", null, function onLoad(event, response) {
            document.getElementById("midi-track-number").value = response["track"];
        }, function onError(event, error) {
        });
    }

    function onMIDITrackNumberChanged() {
        if (suppressEvents) { return; }
        var value = this.value || "1";
        requestHTTP("PUT", "/midi-playback-track", value, function onLoad(event, response) {
            reportMessage("MIDI track changed to #" + value + ".");
        }, function onError(event, error) {
            reportError(error);
        })
    }

    function doMIDIOffsetMsRefresh() {
        requestHTTP("GET", "/midi-playback-offset", null, function onLoad(event, response) {
            document.getElementById("midi-offset-ms").value = Math.round(response["offset"] * 1000);
        }, function onError(event, error) {
        });
    }

    function onMIDIOffsetMsChanged() {
        if (suppressEvents) { return; }
        var value = this.value || "0";
        requestHTTP("PUT", "/midi-playback-offset", value * 0.001, function onLoad(event, response) {
        }, function onError(event, error) {
            reportError(error);
        })
    }

    var schedulerEnabled = false;

    function doSchedulerRefresh() {
        requestHTTP("GET", "/scheduler", null, function onLoad(event, response) {
            schedulerEnabled = response["enabled"];
            if (schedulerEnabled) {
                document.getElementById("sched-set").classList.add("pure-button-primary");
            } else {
                document.getElementById("sched-set").classList.remove("pure-button-primary");
            }
            var startTime = response["start_time"] !== null ? new Date(response["start_time"] * 1000) : null;
            if (startTime !== null) {
                var startTimeHours = startTime.getHours();
                var startTimeMinutes = startTime.getMinutes();
                var startTimeSeconds = startTime.getSeconds();
                startTimeHours = startTimeHours < 10 ? "0" + startTimeHours : "" + startTimeHours;
                startTimeMinutes = startTimeMinutes < 10 ? "0" + startTimeMinutes : "" + startTimeMinutes;
                startTimeSeconds = startTimeSeconds < 10 ? "0" + startTimeSeconds : "" + startTimeSeconds;
                document.getElementById("sched-start-time").value = startTimeHours + " : " + startTimeMinutes + " : " + startTimeSeconds;
            } else {
                document.getElementById("sched-start-time").value = "";
            }
            var loopEnabled = response["loop_enabled"];
            document.getElementById("sched-loop-enabled").checked = loopEnabled;
            var loopInterval = response["loop_interval"];
            if (loopInterval > 0) {
                var loopIntervalHours = Math.trunc(loopInterval / 3600);
                var loopIntervalMinutes = Math.trunc((loopInterval % 3600) / 60);
                var loopIntervalSeconds = Math.trunc(loopInterval % 60);
                loopIntervalMinutes = loopIntervalMinutes < 10 ? "0" + loopIntervalMinutes : "" + loopIntervalMinutes;
                loopIntervalSeconds = loopIntervalSeconds < 10 ? "0" + loopIntervalSeconds : "" + loopIntervalSeconds;
                document.getElementById("sched-loop-interval").value = loopIntervalHours + " : " + loopIntervalMinutes + " : " + loopIntervalSeconds;
            } else {
                document.getElementById("sched-loop-interval").value = "";
            }
        }, function onError(event, error) {
        });
    }

    function onSchedulerChanged() {
        if (suppressEvents) { return; }
        var timeRegEx = /(\d{1,2})\s*:\s*(\d{1,2})\s*:\s*(\d{1,2})/;
        var durationRegEx = /(?:(?:(\d+)\s*:\s*)?(\d+)\s*:\s*)?(\d+)/;
        var startTimeMatch = document.getElementById("sched-start-time").value.match(timeRegEx);
        if (!startTimeMatch && !schedulerEnabled) {
            reportError("Invalid start time.");
            return;
        }
        var loopEnabled = document.getElementById("sched-loop-enabled").checked;
        var loopIntervalMatch = document.getElementById("sched-loop-interval").value.match(durationRegEx);
        if (!loopIntervalMatch && !schedulerEnabled && loopEnabled) {
            reportError("Invalid loop interval.");
            return;
        }
        var now = new Date();
        var startTime = null;
        if (startTimeMatch) {
            startTime = new Date(now.getFullYear(), now.getMonth(), now.getDate(), +startTimeMatch[1], +startTimeMatch[2], +startTimeMatch[3], 0);
            if (startTime.getTime() < now.getTime()) {
                if (now.getTime() - startTime.getTime() > 43200000) {
                    startTime.setDate(startTime.getDate() + 1);
                }
            } else {
                if (startTime.getTime() - now.getTime() > 43200000) {
                    startTime.setDate(startTime.getDate() - 1);
                }
            }
        }
        var loopInterval = 0;
        if (loopIntervalMatch) {
            var loopIntervalHours = +(loopIntervalMatch[1] || 0)
            var loopIntervalMinutes = +(loopIntervalMatch[2] || 0)
            var loopintervalSeconds = +loopIntervalMatch[3];
            loopInterval = loopIntervalHours * 3600 + loopIntervalMinutes * 60 + loopintervalSeconds;
        }

        var button = document.getElementById("sched-set");
        if (this === button) {
            schedulerEnabled = !schedulerEnabled;
            if (schedulerEnabled) {
                button.disabled = true;
                setTimeout(function onTimeout() {
                    button.disabled = false;
                }, 1000);
            }
        } else if (this === document.getElementById("sched-start-time") || (loopEnabled && this === document.getElementById("sched-loop-interval"))) {
            schedulerEnabled = false;
        }

        var body = {
            "enabled": schedulerEnabled,
            "start_time": startTime !== null ? startTime.getTime() * 0.001 : null,
            "loop_enabled": loopEnabled,
            "loop_interval": loopInterval,
        };
        requestHTTP("PUT", "/scheduler", JSON.stringify(body), function onLoad(event, response) {
            schedulerEnabled = response["enabled"];
            if (schedulerEnabled) {
                button.classList.add("pure-button-primary");
            } else {
                button.classList.remove("pure-button-primary");
            }
        }, function onError(event, error) {
            reportError(error);
        });
    }

    document.getElementById("midi-input-refresh").addEventListener("click", onMidiInputRefreshClicked);
    document.getElementById("midi-input-device").addEventListener("change", onMidiInputDeviceChanged);
    document.getElementById("midi-output-refresh").addEventListener("click", onMidiOutputRefreshClicked);
    document.getElementById("midi-output-device").addEventListener("change", onMidiOutputDeviceChanged);
    document.getElementById("synth-bank").addEventListener("change", onSynthBankChanged);
    document.getElementById("synth-patch").addEventListener("change", onSynthPatchChanged);
    document.getElementById("synth-transpose").addEventListener("change", onSynthTransposeChanged);
    document.getElementById("synth-instrument").addEventListener("change", onSynthInstrumentChanged);
    document.getElementById("ntp-sync").addEventListener("click", onNTPSyncClicked);
    document.getElementById("current-time-copy").addEventListener("click", onCurrentTimeCopyClicked);
    document.getElementById("midi-file").addEventListener("change", onMIDIFileChanged);
    document.getElementById("midi-track-number").addEventListener("change", onMIDITrackNumberChanged);
    document.getElementById("midi-offset-ms").addEventListener("change", onMIDIOffsetMsChanged);
    document.getElementById("sched-start-time").addEventListener("change", onSchedulerChanged);
    document.getElementById("sched-set").addEventListener("click", onSchedulerChanged);
    document.getElementById("sched-loop-enabled").addEventListener("change", onSchedulerChanged);
    document.getElementById("sched-loop-interval").addEventListener("change", onSchedulerChanged);

    document.getElementById("midi-file").value = "";
    updateAllStates(0);
    requestAnimationFrame(displayServerTime);

})();
