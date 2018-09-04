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
        document.getElementById("float-container").appendChild(el);
        setTimeout(function onHide() {
            el.style.marginTop = -el.offsetHeight + "px";
        }, 3196);
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
        document.getElementById("float-container").appendChild(el);
        setTimeout(function onHide() {
            el.style.marginTop = -el.offsetHeight + "px";
        }, 3196);
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
        var numFailed = 0;
        var countLoad = function countLoad(event, response) {
            numLoaded++;
            if (numLoaded == 3) {
                doSynthInstrumentUpdate();
            }
        }
        var countError = function countError(event, error) {
            numFailed++;
            if (numFailed == 1) {
                reportError(error);
            }
        }
        requestHTTP("GET", "/midi-output-bank", null, function onLoad(event, response) {
            var value = response["bank"];
            suppressEvents = true;
            synthBank.value = value;
            suppressEvents = false;
            countLoad();
        }, countError);
        requestHTTP("GET", "/midi-output-patch", null, function onLoad(event, response) {
            var value = response["patch"];
            suppressEvents = true;
            synthPatch.value = +value + 1;
            suppressEvents = false;
            countLoad();
        }, countError);
        requestHTTP("GET", "/midi-output-transpose", null, function onLoad(event, response) {
            var value = response["transpose"];
            suppressEvents = true;
            synthTranspose.value = value;
            suppressEvents = false;
            countLoad();
        }, countError);
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
            if (numLoaded == 3) {
                reportMessage("MIDI instrument changed to " + text + ".");
            }
        }
        var countError = function countError(event, error) {
            numFailed++;
            if (numFailed == 1) {
                reportError(error);
            }
        }
        requestHTTP("PUT", "/midi-output-bank", synthBank.value, countLoad, countError);
        requestHTTP("PUT", "/midi-output-patch", +synthPatch.value - 1, countLoad, countError);
        requestHTTP("PUT", "/midi-output-transpose", synthTranspose.value, countLoad, countError);
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

    function updateServerTime() {
        doUpdateServerTime();
        setTimeout(updateServerTime, 5000);
    }

    function onNTPSyncClicked() {
        var button = this;
        var server = document.getElementById("ntp-server").value || "0.beevik-ntp.pool.ntp.org";
        button.disabled = true;
        setTimeout(function onTimeout() {
            button.disabled = false;
        }, 5000);
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

    function onSchedSetClicked() {
        reportError("Feature not implemented yet");
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
    document.getElementById("sched-set").addEventListener("click", onSchedSetClicked);

    doMidiInputRefresh(true);
    doMidiOutputRefresh(true);
    doSynthInstrumentRefresh();
    doMIDITrackNumberRefresh();
    doMIDIOffsetMsRefresh();
    updateServerTime();
    requestAnimationFrame(displayServerTime);

})();
