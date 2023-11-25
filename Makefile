.PHONY: all clean

all: midi2ffxiv.exe midi2ffxiv-$(shell date +%Y%m%d).zip

clean:
	rm -f -v midi2ffxiv.exe midi2ffxiv-????????.zip

midi2ffxiv.exe: kernel32/kernel32.go keystroke.go main.go midi-playback.go midi-realtime.go ntp.go parse-config.go preset.go user32/user32.go web.go winmm/winmm.go
	env GOOS=windows GOARCH=amd64 go get -d -v
	env GOOS=windows GOARCH=amd64 go build -ldflags "-X main.versionInfo=$(shell git describe --tags --long)" .

midi2ffxiv-%.zip: Readme.md midi2ffxiv.exe midi2ffxiv.exe.manifest midi2ffxiv.conf midi2ffxiv_no_modifier.conf screenshot.png demo/README.txt demo/*.mid web/index.html web/scripts.js web/styles.css
	rm -f -v "$@"
	zip -9 "$@" Readme.md midi2ffxiv.exe midi2ffxiv.exe.manifest midi2ffxiv.conf midi2ffxiv_no_modifier.conf screenshot.png demo/README.txt demo/*.mid web/index.html web/scripts.js web/styles.css
