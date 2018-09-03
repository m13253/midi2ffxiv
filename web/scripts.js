(function () {
    function requestHTTP(method, url, onLoad, onError) {
        var xhr = new XMLHttpRequest();
        xhr.onload = onLoad;
        xhr.onerror = onError;
        xhr.open(method, url);
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
    function onMidiInputRefreshClicked() { }
    function onMidiInputDeviceChanged() { }
    function onMidiOutputRefreshClicked() { }
    function onMidiOutputDeviceChanged() { }
    function onSynthBankChanged() { }
    function onSynthPatchChanged() { }
    function onSynthTransposeChanged() { }
    function onSynthInstrumentChanged() {
        var synthBank = document.getElementById("synth-bank");
        var synthPatch = document.getElementById("synth-patch");
        var synthTranspose = document.getElementById("synth-transpose");
        switch (this.selectedIndex) {
            case 0:
                synthBank.value = 0;
                synthPatch.value = 47;
                synthTranspose.value = 0;
                break;
            case 1:
                synthBank.value = 0;
                synthPatch.value = 1;
                synthTranspose.value = 12;
                break;
            case 2:
                synthBank.value = 0;
                synthPatch.value = 26;
                synthTranspose.value = -12;
                break;
            case 3:
                synthBank.value = 0;
                synthPatch.value = 46;
                synthTranspose.value = 0;
                break;
            case 4:
                synthBank.value = 0;
                synthPatch.value = 74;
                synthTranspose.value = 0;
                break;
            case 5:
                synthBank.value = 0;
                synthPatch.value = 69;
                synthTranspose.value = 0;
                break;
            case 6:
                synthBank.value = 0;
                synthPatch.value = 72;
                synthTranspose.value = 0;
                break;
            case 7:
                synthBank.value = 0;
                synthPatch.value = 73;
                synthTranspose.value = 0;
                break;
            case 8:
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
})();
