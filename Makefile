BINARIES = ooze ooze-zfs ooze-nixos

.PHONY: build test install clean

build: $(BINARIES)

ooze: cmd/ooze/main.go
	go build -o $@ cmd/ooze

ooze-zfs: cmd/ooze-zfs/main.go
	go build -o $@ cmd/ooze-zfs

ooze-nixos: cmd/ooze-nixos/main.go
	go build -o $@ cmd/ooze-nixos

test:
	go test ./...

install: build
	install -d $(DESTDIR)/usr/local/bin
	install -m 755 $(BINARIES) $(DESTDIR)/usr/local/bin/

install-user: build
	mkdir -p ~/.local/bin
	install -m 755 $(BINARIES) ~/.local/bin/

clean:
	rm -f $(BINARIES)
