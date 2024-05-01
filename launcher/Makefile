prefix = /usr/local
PROG = mor-launch

all: $(PROG)

$(PROG): cmd/main.go
	go build -o $@ $^

clean:
	rm -f $(PROG)

install: $(PROG)
	mkdir -p $(DESTDIR)$(prefix)/bin
	cp $< $(DESTDIR)$(prefix)/bin/

uninstall:
	rm $(DESTDIR)$(prefix)/bin/$(PROG)

run:
	go run cmd/main.go

.PHONY: all clean install uninstall run

