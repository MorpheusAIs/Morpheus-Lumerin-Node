# Measurement probes (FIXES-2 4a etc.): built and run on the test Mac only,
# never on the authoring machine (SLOT=<name> ./run-on-test-mac.sh tests/probes).
# Probes are informational: they print MEASURED lines and exit 0; their output
# is recorded in impl/NOTES.md.
CC = clang
CFLAGS = -std=c17 -Wall -Wextra -Werror

.PHONY: test clean

test: nosend_probe exc_probe exc_abrt exc_trap
	./nosend_probe
	./exc_probe

nosend_probe: nosend_probe.c
	$(CC) $(CFLAGS) -o $@ $<

exc_probe: exc_probe.c
	$(CC) $(CFLAGS) -o $@ $<

exc_abrt: exc_abrt.c
	$(CC) $(CFLAGS) -arch arm64 -o $@ $<
	@codesign -f -s - --options runtime -i test.exc-abrt $@ 2>&1 | grep -v replacing || true

exc_trap: exc_trap.c
	$(CC) $(CFLAGS) -arch arm64 -o $@ $<
	@codesign -f -s - --options runtime -i test.exc-trap $@ 2>&1 | grep -v replacing || true

clean:
	rm -f nosend_probe exc_probe exc_abrt exc_trap
