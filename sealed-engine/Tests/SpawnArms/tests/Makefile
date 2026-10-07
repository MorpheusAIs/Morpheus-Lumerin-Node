# Gating-test harness for impl/src/sealed_spawn.c (TASK-tests.md, harness part).
# Builds the module from ../impl/src/*.c, builds and ad-hoc-signs the test
# children, generates the A-13 launch-requirement bytes the way
# context/probes/run-a13.sh does (a13.swift dumpfile route), compiles every
# tests/arms/*.c as an arm linked with the module and lib/harness.c, and runs
# each arm under a 60 s timeout. Arms print their own PASS/FAIL/SKIP line with
# a reason; the wrapper adds a line only for timeouts and bad exits.
# Children are never executed on the authoring machine; run ./run-on-test-mac.sh tests.
CC = clang
CFLAGS = -std=c17 -Wall -Wextra -I../include -Ilib
B = build
# a13.swift sits at ../probes on the test Mac (run-on-test-mac.sh rsyncs
# context/probes to <slot>/probes) and at ../context/probes in this tree.
A13 = $(or $(wildcard ../probes/a13.swift),$(wildcard ../context/probes/a13.swift))

# Arm programs: one file per arm, linked with the module and the harness.
# Wildcard so parallel arm authors add files without editing this Makefile.
ARMS = $(notdir $(basename $(wildcard arms/*.c)))

CHILDREN = child-ok child-alt child-gta child-crash child-noconn child-nosend \
	$(GEN_CHILDREN)
# Extra child sources tests/children/r2_*.c / r3_*.c: built and ad-hoc signed
# by the generic rules below; each gets CHILD_<NAME> (path) and HASH_<NAME>
# (its CD hash) exported to the arms by the `test` target. Existing explicit
# child rules are kept; the wildcard names never collide with them.
GEN_CHILDREN = $(notdir $(basename $(wildcard children/r2_*.c children/r3_*.c)))
MOD_SRC = $(wildcard ../impl/src/*.c)
MOD_OBJ = $(MOD_SRC:../impl/src/%.c=$(B)/mod-%.o)

cdhash = codesign -dvvv $(1) 2>&1 | awk -F= '/^CDHash=/{print $$2}'

.PHONY: test all clean

test: all
	@rm -f $(B)/m.*
	@export CHILD_OK=$(CURDIR)/$(B)/child-ok \
		CHILD_ALT=$(CURDIR)/$(B)/child-alt \
		CHILD_GTA=$(CURDIR)/$(B)/child-gta \
		CHILD_CRASH=$(CURDIR)/$(B)/child-crash \
		CHILD_NOCONN=$(CURDIR)/$(B)/child-noconn \
		CHILD_NOSEND=$(CURDIR)/$(B)/child-nosend \
		HASH_OK=$$($(call cdhash,$(B)/child-ok)) \
		HASH_ALT=$$($(call cdhash,$(B)/child-alt)) \
		HASH_GTA=$$($(call cdhash,$(B)/child-gta)) \
		HASH_CRASH=$$($(call cdhash,$(B)/child-crash)) \
		HASH_NOCONN=$$($(call cdhash,$(B)/child-noconn)) \
		HASH_NOSEND=$$($(call cdhash,$(B)/child-nosend)) \
		REQ_ALT=$(CURDIR)/$(B)/req-child-alt.bin \
		REQ_GTA=$(CURDIR)/$(B)/req-child-gta.bin \
		REQ_CRASH=$(CURDIR)/$(B)/req-child-crash.bin \
		REQ_NOCONN=$(CURDIR)/$(B)/req-child-noconn.bin \
		REQ_NOSEND=$(CURDIR)/$(B)/req-child-nosend.bin \
		REQ_OK=$(CURDIR)/$(B)/req-ok.bin \
		WORKDIR=$(CURDIR)/$(B); \
	for c in $(CHILDREN); do \
		u=`printf %s "$$c" | tr 'a-z-' 'A-Z_'`; \
		eval "CHILD_$$u='$(CURDIR)/$(B)/'$$c; HASH_$$u=`$(call cdhash,$(B)/$$c)`; REQ_$$u='$(CURDIR)/$(B)/req-'$$c'.bin'; export CHILD_$$u HASH_$$u REQ_$$u"; \
	done; \
	status=0; \
	for a in $(ARMS); do \
		rm -f $(B)/m.*; \
		out=`perl -e 'alarm shift; exec @ARGV' 60 ./$(B)/arm-$$a 2>&1`; rc=$$?; \
		echo "$$out"; \
		if [ $$rc -gt 128 ]; then echo "FAIL $$a: killed after 60 s (timeout)"; status=1; \
		elif [ $$rc -gt 1 ]; then echo "FAIL $$a: exit $$rc (no result line)"; status=1; \
		elif [ $$rc -eq 1 ]; then status=1; \
		elif ! printf '%s\n' "$$out" | grep -Eq '^(PASS|SKIP) '; then \
			echo "FAIL $$a: no PASS/SKIP result line"; status=1; fi; \
		if printf '%s\n' "$$out" | grep -q '^FAIL'; then status=1; fi; \
	done; exit $$status

# Generic per-child requirement bytes: for every test child, bytes naming that
# child's own CD hash (so an arm can spawn it pinned, requirement attached).
# req-ok.bin below keeps its explicit rule (other arms reference REQ_OK).
REQS = $(CHILDREN:%=$(B)/req-%.bin)

all: $(ARMS:%=$(B)/arm-%) $(CHILDREN:%=$(B)/%) $(REQS) $(B)/req-ok.bin

$(B):
	mkdir -p $(B)

# The module under test (never edited here; a missing sealed_spawn.c fails the
# build, which is expected while the other agent works).
$(B)/mod-%.o: ../impl/src/%.c $(wildcard ../include/*.h) | $(B)
	$(CC) $(CFLAGS) -c $< -o $@

$(B)/harness.o: lib/harness.c lib/harness.h | $(B)
	$(CC) $(CFLAGS) -c lib/harness.c -o $@

$(B)/arm-%: arms/%.c $(B)/harness.o $(MOD_OBJ) | $(B)
	$(CC) $(CFLAGS) $^ -o $@
	@codesign -f -s - --options runtime $@ 2>&1 | grep -v replacing || true

# Test children. child-ok/child-alt/child-gta are the same source with a
# different VARIANT, so each has its own CD hash (probes/marker.c pattern);
# every child is ad-hoc signed with hardened runtime like run-a13.sh.
$(B)/child-ok: children/good.c | $(B)
	$(CC) $(CFLAGS) -arch arm64 -DVARIANT=1 children/good.c -o $@
	@codesign -f -s - --options runtime -i test.child-ok $@ 2>&1 | grep -v replacing || true

$(B)/child-alt: children/good.c | $(B)
	$(CC) $(CFLAGS) -arch arm64 -DVARIANT=2 children/good.c -o $@
	@codesign -f -s - --options runtime -i test.child-alt $@ 2>&1 | grep -v replacing || true

$(B)/child-gta: children/good.c $(B)/gta.ent.plist | $(B)
	$(CC) $(CFLAGS) -arch arm64 -DVARIANT=3 children/good.c -o $@
	@codesign -f -s - --options runtime --entitlements $(B)/gta.ent.plist -i test.child-gta $@ 2>&1 | grep -v replacing || true

$(B)/child-crash: children/crash.c | $(B)
	$(CC) $(CFLAGS) -arch arm64 children/crash.c -o $@
	@codesign -f -s - --options runtime -i test.child-crash $@ 2>&1 | grep -v replacing || true

$(B)/child-noconn: children/noconn.c | $(B)
	$(CC) $(CFLAGS) -arch arm64 children/noconn.c -o $@
	@codesign -f -s - --options runtime -i test.child-noconn $@ 2>&1 | grep -v replacing || true

$(B)/child-nosend: children/nosend.c | $(B)
	$(CC) $(CFLAGS) -arch arm64 children/nosend.c -o $@
	@codesign -f -s - --options runtime -i test.child-nosend $@ 2>&1 | grep -v replacing || true

# Generic rule for the wildcard children (tests/children/r2_*.c, r3_*.c):
# same build/sign shape as the explicit rules (hardened runtime, identifier
# test.<name>). Only the generic wildcard names hit this rule.
$(B)/r2_%: children/r2_%.c | $(B)
	$(CC) $(CFLAGS) -arch arm64 $< -o $@
	@codesign -f -s - --options runtime -i test.$(@F) $@ 2>&1 | grep -v replacing || true

$(B)/r3_%: children/r3_%.c | $(B)
	$(CC) $(CFLAGS) -arch arm64 $< -o $@
	@codesign -f -s - --options runtime -i test.$(@F) $@ 2>&1 | grep -v replacing || true

$(B)/gta.ent.plist:
	printf '<?xml version="1.0" encoding="UTF-8"?>\n<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">\n<plist version="1.0"><dict><key>com.apple.security.get-task-allow</key><true/></dict></plist>\n' > $@

# Launch-requirement bytes naming child-ok's CD hash: encoded by a13.swift's
# Process.launchRequirement, dumped with the dumpfile argument (which exits
# before launching anything), exactly the route run-a13.sh measures (E1).
$(B)/reqgen: $(A13) | $(B)
	swiftc -O -o $@ $(A13)

$(B)/req-ok.bin: $(B)/reqgen $(B)/child-ok
	@H=$$($(call cdhash,$(B)/child-ok)); ./$(B)/reqgen $(B)/child-ok $$H dummy $@

# Generic per-child requirement bytes: for every test child, bytes naming that
# child's own CD hash (so an arm can spawn it pinned, requirement attached).
# req-ok.bin above keeps its explicit rule (other arms reference REQ_OK).

$(B)/req-%.bin: $(B)/reqgen $(B)/%
	@H=`$(call cdhash,$(B)/$*)`; ./$(B)/reqgen $(B)/$* $$H dummy $@

clean:
	rm -rf $(B)
