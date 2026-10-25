package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/client"
)

// marketplace serves /blockchain/models plus a per-model bids endpoint, so the
// picker's loading path can be exercised without a router.
type marketplace struct {
	models []map[string]interface{}
	// bidsFor maps a model id to the bids response for it. A model absent from
	// this map answers 500, standing in for a throttled RPC.
	bidsFor map[string][]map[string]interface{}

	delay time.Duration // artificial latency per bid lookup

	inFlight int32
	maxSeen  int32
	calls    int32
}

func (m *marketplace) start(t *testing.T) *actions {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/blockchain/models" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"models": m.models})
			return
		}

		if strings.HasPrefix(r.URL.Path, "/blockchain/models/") && strings.HasSuffix(r.URL.Path, "/bids") {
			cur := atomic.AddInt32(&m.inFlight, 1)
			for {
				old := atomic.LoadInt32(&m.maxSeen)
				if cur <= old || atomic.CompareAndSwapInt32(&m.maxSeen, old, cur) {
					break
				}
			}
			atomic.AddInt32(&m.calls, 1)
			if m.delay > 0 {
				time.Sleep(m.delay)
			}
			atomic.AddInt32(&m.inFlight, -1)

			id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/blockchain/models/"), "/bids")
			bids, ok := m.bidsFor[id]
			if !ok {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"over rate limit"}`))
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"bids": bids})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	return NewActions(client.NewApiGatewayClient(srv.URL, http.DefaultClient))
}

func model(id, name string, deleted bool) map[string]interface{} {
	return map[string]interface{}{"Id": id, "Name": name, "IsDeleted": deleted, "ModelType": "LLM"}
}

func liveBid(id string) map[string]interface{} {
	return map[string]interface{}{"Id": id, "DeletedAt": "0", "PricePerSecond": "1000"}
}

func deletedBid(id string) map[string]interface{} {
	return map[string]interface{}{"Id": id, "DeletedAt": "1792180763", "PricePerSecond": "1000"}
}

func namesOf(t *testing.T, models []interface{}) []string {
	t.Helper()
	out := make([]string, 0, len(models))
	for _, item := range models {
		m, ok := item.(map[string]interface{})
		if !ok {
			t.Fatalf("result item is %T, want a map", item)
		}
		out = append(out, m["Name"].(string))
	}
	return out
}

// TestLoadModelsWithBidsFiltersDeletedModels is the regression test for the
// filter that stopped working: the router reports deletion as IsDeleted, while
// the code read DeletedAt, so deleted models were offered for sessions.
func TestLoadModelsWithBidsFiltersDeletedModels(t *testing.T) {
	mp := &marketplace{
		models: []map[string]interface{}{
			model("0xaaa", "live-model", false),
			model("0xbbb", "deleted-model", true),
		},
		bidsFor: map[string][]map[string]interface{}{
			"0xaaa": {liveBid("0xbid1")},
			"0xbbb": {liveBid("0xbid2")},
		},
	}
	a := mp.start(t)

	got, err := a.loadModelsWithBids(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if names := namesOf(t, got); len(names) != 1 || names[0] != "live-model" {
		t.Errorf("got %v, want only live-model", names)
	}
	// The deleted model must not even be asked about.
	if n := atomic.LoadInt32(&mp.calls); n != 1 {
		t.Errorf("made %d bid lookups, want 1 (the deleted model should be skipped)", n)
	}
}

func TestLoadModelsWithBidsSkipsModelsWithoutLiveBids(t *testing.T) {
	mp := &marketplace{
		models: []map[string]interface{}{
			model("0xaaa", "has-live-bid", false),
			model("0xbbb", "only-deleted-bids", false),
			model("0xccc", "no-bids-at-all", false),
		},
		bidsFor: map[string][]map[string]interface{}{
			"0xaaa": {liveBid("0xbid1"), deletedBid("0xbid2")},
			"0xbbb": {deletedBid("0xbid3")},
			"0xccc": {},
		},
	}
	a := mp.start(t)

	got, err := a.loadModelsWithBids(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	names := namesOf(t, got)
	if len(names) != 1 || names[0] != "has-live-bid" {
		t.Errorf("got %v, want only has-live-bid", names)
	}
	// The surviving model must carry only its live bids.
	bids := got[0].(map[string]interface{})["bids"].([]interface{})
	if len(bids) != 1 {
		t.Errorf("got %d bids attached, want 1 (the deleted one excluded)", len(bids))
	}
}

// TestLoadModelsWithBidsSurvivesFailures covers the behaviour change: one
// failing lookup used to abort the whole command, losing the catalogue.
func TestLoadModelsWithBidsSurvivesFailures(t *testing.T) {
	mp := &marketplace{
		models: []map[string]interface{}{
			model("0xaaa", "first", false),
			model("0xbbb", "throttled", false), // absent from bidsFor -> 500
			model("0xccc", "third", false),
		},
		bidsFor: map[string][]map[string]interface{}{
			"0xaaa": {liveBid("0xbid1")},
			"0xccc": {liveBid("0xbid3")},
		},
	}
	a := mp.start(t)

	got, err := a.loadModelsWithBids(context.Background())
	if err != nil {
		t.Fatalf("one failed lookup must not fail the whole load: %v", err)
	}
	if names := namesOf(t, got); len(names) != 2 || names[0] != "first" || names[1] != "third" {
		t.Errorf("got %v, want [first third] in catalogue order", names)
	}
}

func TestLoadModelsWithBidsErrorsWhenEverythingFails(t *testing.T) {
	mp := &marketplace{
		models:  []map[string]interface{}{model("0xaaa", "a", false), model("0xbbb", "b", false)},
		bidsFor: map[string][]map[string]interface{}{}, // every lookup 500s
	}
	a := mp.start(t)

	_, err := a.loadModelsWithBids(context.Background())
	if err == nil {
		t.Fatal("want an error when no model could be loaded")
	}
	if !strings.Contains(err.Error(), "failed") {
		t.Errorf("error should say the lookups failed, got: %v", err)
	}
}

// TestLoadModelsWithBidsPreservesOrder guards the ordering the picker relies
// on: results are indexed by position, not appended as goroutines finish.
func TestLoadModelsWithBidsPreservesOrder(t *testing.T) {
	const n = 40
	mp := &marketplace{
		models:  make([]map[string]interface{}, 0, n),
		bidsFor: map[string][]map[string]interface{}{},
	}
	var want []string
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("0x%03d", i)
		name := fmt.Sprintf("model-%03d", i)
		mp.models = append(mp.models, model(id, name, false))
		mp.bidsFor[id] = []map[string]interface{}{liveBid("bid" + id)}
		want = append(want, name)
	}
	// Stagger the responses so completion order differs from request order.
	mp.delay = 5 * time.Millisecond
	a := mp.start(t)

	got, err := a.loadModelsWithBids(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	names := namesOf(t, got)
	if len(names) != n {
		t.Fatalf("got %d models, want %d", len(names), n)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("position %d is %q, want %q -- catalogue order was not preserved", i, names[i], want[i])
		}
	}
}

// TestLoadModelsWithBidsIsConcurrentButBounded is the point of the change:
// lookups overlap, and no more than bidFetchWorkers run at once so the shared
// public RPCs are not hammered.
func TestLoadModelsWithBidsIsConcurrentButBounded(t *testing.T) {
	const n = 50
	mp := &marketplace{
		models:  make([]map[string]interface{}, 0, n),
		bidsFor: map[string][]map[string]interface{}{},
		delay:   20 * time.Millisecond,
	}
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("0x%03d", i)
		mp.models = append(mp.models, model(id, fmt.Sprintf("m%03d", i), false))
		mp.bidsFor[id] = []map[string]interface{}{liveBid("bid" + id)}
	}
	a := mp.start(t)

	start := time.Now()
	got, err := a.loadModelsWithBids(context.Background())
	elapsed := time.Since(start)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != n {
		t.Fatalf("got %d models, want %d", len(got), n)
	}

	peak := atomic.LoadInt32(&mp.maxSeen)
	if peak > int32(bidFetchWorkers) {
		t.Errorf("peak concurrency %d exceeded the %d-worker bound", peak, bidFetchWorkers)
	}
	if peak < 2 {
		t.Errorf("peak concurrency %d: lookups did not overlap at all", peak)
	}

	// Serial would be ~n*delay; allow generous slack for slow CI.
	serial := time.Duration(n) * mp.delay
	if elapsed > serial/2 {
		t.Errorf("took %v, which is not meaningfully faster than serial (%v)", elapsed, serial)
	}
}

func TestLoadModelsWithBidsRejectsAnUnexpectedShape(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"unexpected":"shape"}`))
	}))
	defer srv.Close()
	a := NewActions(client.NewApiGatewayClient(srv.URL, http.DefaultClient))

	// Previously this panicked on a type assertion.
	_, err := a.loadModelsWithBids(context.Background())
	if err == nil {
		t.Fatal("want an error for a response with no models array")
	}
}

func TestLoadModelsWithBidsToleratesJunkEntries(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/blockchain/models" {
			// A string, a model with no Id, and a usable one.
			_, _ = w.Write([]byte(`{"models":["junk",{"Name":"no-id","IsDeleted":false},{"Id":"0xaaa","Name":"ok","IsDeleted":false}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"bids":[{"Id":"0xbid","DeletedAt":"0"}]}`))
	}))
	defer srv.Close()
	a := NewActions(client.NewApiGatewayClient(srv.URL, http.DefaultClient))

	got, err := a.loadModelsWithBids(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if names := namesOf(t, got); len(names) != 1 || names[0] != "ok" {
		t.Errorf("got %v, want only the well-formed model", names)
	}
}

func TestLiveBids(t *testing.T) {
	cases := []struct {
		name string
		in   map[string]interface{}
		want int
	}{
		{
			name: "keeps bids with a zero deletion stamp",
			in:   map[string]interface{}{"bids": []interface{}{liveBid("a"), liveBid("b")}},
			want: 2,
		},
		{
			name: "drops deleted bids",
			in:   map[string]interface{}{"bids": []interface{}{liveBid("a"), deletedBid("b")}},
			want: 1,
		},
		{
			name: "a numeric zero still counts as live",
			in:   map[string]interface{}{"bids": []interface{}{map[string]interface{}{"DeletedAt": 0}}},
			want: 1,
		},
		{
			name: "an absent stamp is not treated as live",
			in:   map[string]interface{}{"bids": []interface{}{map[string]interface{}{"Id": "a"}}},
			want: 0,
		},
		{
			name: "no bids key",
			in:   map[string]interface{}{},
			want: 0,
		},
		{
			name: "junk entries are skipped",
			in:   map[string]interface{}{"bids": []interface{}{"nope", liveBid("a")}},
			want: 1,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := len(liveBids(tc.in)); got != tc.want {
				t.Errorf("got %d live bids, want %d", got, tc.want)
			}
		})
	}
}

// guard against the worker bound being set to something that defeats either
// purpose of the change.
func TestBidFetchWorkersIsSane(t *testing.T) {
	if bidFetchWorkers < 2 {
		t.Errorf("bidFetchWorkers = %d, which is serial again", bidFetchWorkers)
	}
	if bidFetchWorkers > 32 {
		t.Errorf("bidFetchWorkers = %d, too many in flight against a shared RPC", bidFetchWorkers)
	}
}
