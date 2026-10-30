package main

import (
	"strings"
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/client"
)

const (
	testModelID  = "0xc85309216df7ef7902fe96810914fac7ba479f43f16ef383ba95bea39bb5ba76"
	testProvider = "0x010208ec11f3a844dd2f5003a2807dde982ebb03"
)

func testItem() client.SessionListItem {
	return client.SessionListItem{
		Session:        "0x2b5b7c295de206820e042221b83a25324d57fd2c757a92594c465c4b1976b7e4",
		ModelORAgent:   testModelID,
		Provider:       testProvider,
		Bid:            "0x09254036bb7f5a2c24c8c29f32109ce5583863f01cc8eec2fcd21fa16f79ce65",
		PricePerSecond: 27777777777777,
		EndsAt:         2000000000,
	}
}

// lineFor returns the rendered line whose label matches, e.g. "provider".
func lineFor(t *testing.T, out, label string) string {
	t.Helper()
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "- "+label+":") {
			return line
		}
	}
	t.Fatalf("no %q line in output:\n%s", label, out)
	return ""
}

// TestSessionListLabelsMatchTheirValues is the regression test: the model id
// was printed under the label "provider", and the provider address -- already
// decoded from the response -- was never shown.
func TestSessionListLabelsMatchTheirValues(t *testing.T) {
	out := formatSessionListItem(testItem(), time.Unix(1000000000, 0))

	providerLine := lineFor(t, out, "provider")
	if !strings.Contains(providerLine, testProvider) {
		t.Errorf("provider line should carry the provider address, got: %s", providerLine)
	}
	if strings.Contains(providerLine, testModelID) {
		t.Errorf("provider line is still showing the model id: %s", providerLine)
	}

	modelLine := lineFor(t, out, "model")
	if !strings.Contains(modelLine, testModelID) {
		t.Errorf("model line should carry the model id, got: %s", modelLine)
	}
	if strings.Contains(modelLine, testProvider) {
		t.Errorf("model line is showing the provider address: %s", modelLine)
	}
}

func TestSessionListShowsTheSessionAndPrice(t *testing.T) {
	item := testItem()
	out := formatSessionListItem(item, time.Unix(1000000000, 0))

	if !strings.Contains(out, item.Session) {
		t.Errorf("output should name the session id:\n%s", out)
	}
	// Printed as a plain integer, not Go's %v on a uint64 wrapper.
	if !strings.Contains(out, "27777777777777") {
		t.Errorf("output should carry the price per second:\n%s", out)
	}
}

func TestSessionListExpiryIsRelativeToNow(t *testing.T) {
	item := testItem()
	item.EndsAt = 1_500_000_000

	before := formatSessionListItem(item, time.Unix(1_400_000_000, 0))
	if !strings.Contains(lineFor(t, before, "expired"), "false") {
		t.Errorf("a session ending in the future must not read as expired:\n%s", before)
	}

	after := formatSessionListItem(item, time.Unix(1_600_000_000, 0))
	if !strings.Contains(lineFor(t, after, "expired"), "true") {
		t.Errorf("a session past its end must read as expired:\n%s", after)
	}
}

func TestSessionListClosedReflectsTheCloseoutReceipt(t *testing.T) {
	item := testItem()

	open := formatSessionListItem(item, time.Unix(1000000000, 0))
	if !strings.Contains(lineFor(t, open, "closed"), "false") {
		t.Errorf("no closeout receipt should read as not closed:\n%s", open)
	}

	item.CloseoutReceipt = "2b5b7c295de20682..."
	closed := formatSessionListItem(item, time.Unix(1000000000, 0))
	if !strings.Contains(lineFor(t, closed, "closed"), "true") {
		t.Errorf("a closeout receipt should read as closed:\n%s", closed)
	}
}
