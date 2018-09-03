(function () {

    function requestHTTP(method, url, body, onLoad, onError) {
        var xhr = new XMLHttpRequest();
        xhr.onerror = onError.bind(xhr);
        xhr.onabort = xhr.onerror;
        xhr.onload = function onload(event) {
            try {
                return onLoad.bind(xhr)(event);
            } catch (e) {
                console.error(e);
                xhr.onerror(event, e);
            }
        }
        xhr.open(method, url, true);
        xhr.responseType = "json";
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
        requestHTTP("GET", "/midi-input-device", null, function onLoad() {
            var list = document.getElementById("midi-input-device");
            suppressEvents = true;
            try {
                clearSelect(list);
                addSelectOption(list, "(None)", "-1");
                var devices = this.response["devices"];
                for (var i = 0; i < devices.length; i++) {
                    addSelectOption(list, devices[i], i.toString());
                }
                list.value = this.response["selected"];
            } finally {
                suppressEvents = false;
            }
            if (!quiet) {
                reportMessage("MIDI input device updated.");
            }
        }, function onError() {
            if (!quiet) {
                reportError("Failed to retrieve MIDI input devices.");
            }
        });
    }

    function onMidiInputRefreshClicked() {
        return doMidiInputRefresh(false);
    }

    function onMidiInputDeviceChanged() {
        if (suppressEvents) { return; }
        var deviceID = document.getElementById("midi-input-device").value;
        requestHTTP("PUT", "/midi-input-device", deviceID, function onLoad() {
            if (deviceID >= 0) {
                reportMessage("MIDI input device changed to #" + deviceID + ".");
            } else {
                reportMessage("MIDI input device changed to (None).");
            }
        }, function onError() {
            reportError("Failed to change MIDI input device.");
        })
    }

    function doMidiOutputRefresh(quiet) {
        requestHTTP("GET", "/midi-output-device", null, function onLoad() {
            var list = document.getElementById("midi-output-device");
            clearSelect(list);
            addSelectOption(list, "(None)", "-1");
            var devices = this.response["devices"];
            for (var i = 0; i < devices.length; i++) {
                addSelectOption(list, devices[i], i.toString());
            }
            list.value = this.response["selected"];
            if (!quiet) {
                reportMessage("MIDI output device updated.");
            }
        }, function onError() {
            if (!quiet) {
                reportError("Failed to retrieve MIDI output devices.");
            }
        });
    }

    function onMidiOutputRefreshClicked() {
        return doMidiOutputRefresh(false);
    }

    function onMidiOutputDeviceChanged() {
        if (suppressEvents) { return; }
        var deviceID = document.getElementById("midi-output-device").value;
        requestHTTP("PUT", "/midi-output-device", deviceID, function onLoad() {
            if (deviceID >= 0) {
                reportMessage("MIDI output device changed to #" + deviceID + ".");
            } else {
                reportMessage("MIDI output device changed to (None).");
            }
        }, function onError() {
            reportError("Failed to change MIDI output device.");
        })
    }

    function onSynthBankChanged() { }

    function onSynthPatchChanged() { }

    function onSynthTransposeChanged() { }

    function onSynthInstrumentChanged() {
        var synthBank = document.getElementById("synth-bank");
        var synthPatch = document.getElementById("synth-patch");
        var synthTranspose = document.getElementById("synth-transpose");
        switch (this.value) {
            case "47":
                synthBank.value = 0;
                synthPatch.value = 47;
                synthTranspose.value = 0;
                break;
            case "1":
                synthBank.value = 0;
                synthPatch.value = 1;
                synthTranspose.value = 12;
                break;
            case "26":
                synthBank.value = 0;
                synthPatch.value = 26;
                synthTranspose.value = -12;
                break;
            case "46":
                synthBank.value = 0;
                synthPatch.value = 46;
                synthTranspose.value = 0;
                break;
            case "74":
                synthBank.value = 0;
                synthPatch.value = 74;
                synthTranspose.value = 0;
                break;
            case "69":
                synthBank.value = 0;
                synthPatch.value = 69;
                synthTranspose.value = 0;
                break;
            case "72":
                synthBank.value = 0;
                synthPatch.value = 72;
                synthTranspose.value = 0;
                break;
            case "73":
                synthBank.value = 0;
                synthPatch.value = 73;
                synthTranspose.value = 0;
                break;
            case "76":
                synthBank.value = 0;
                synthPatch.value = 76;
                synthTranspose.value = 0;
                break;
        }
    }

    function onNTPSyncClicked() { }

    function onCurrentTimeCopyClicked() { }

    function onMIDIFileChanged() { }

    function onMIDITrackNumberChanged() { }

    function onMIDIOffsetMsChanged() { }

    function onSchedSetClicked() { }

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

})();
