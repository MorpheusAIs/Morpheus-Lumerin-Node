prefix = /usr/local
PROG = mor-cli

all: $(PROG)

# Build the whole package. `go build main.go` compiles only that file and
# drops modelsource.go, which is how the v7.20.0 release executable job failed.
$(PROG):
	go build -o $(PROG) .

build:
	go build -o mor-cli .

clean:
	rm -f $(PROG)

install: $(PROG)
	mkdir -p $(DESTDIR)$(prefix)/bin
	cp $< $(DESTDIR)$(prefix)/bin/

uninstall:
	rm $(DESTDIR)$(prefix)/bin/$(PROG)

run:
	go run .

.PHONY: all build clean install uninstall run

