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

// Two well-formed 32-byte model ids. Built rather than written out so the
// length is obviously right: the loader rejects anything that is not 0x plus 64
// hex characters.
var (
	feedModelA = "0x" + strings.Repeat("a", 64)
	feedModelB = "0x" + strings.Repeat("b", 64)
)

// feed serves the published catalogue files, so the loader can be exercised
// without reaching active.mor.org.
type feed struct {
	// files maps a published name ("gateway_models.json") to its body. A name
	// absent from this map answers 404.
	files map[string]string

	// status, when nonzero, is answered for every request instead of a body.
	status int

	hits int32
}

func (f *feed) start(t *testing.T) *actions {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&f.hits, 1)
		if f.status != 0 {
			w.WriteHeader(f.status)
			_, _ = w.Write([]byte(`{"error":"nope"}`))
			return
		}
		body, ok := f.files[strings.TrimPrefix(r.URL.Path, "/")]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	// The router side is deliberately pointed at a dead address: these tests
	// assert on the feed path, and a fallback must surface as an error rather
	// than quietly reaching a real router.
	a := NewActions(client.NewApiGatewayClient("http://127.0.0.1:1", http.DefaultClient))
	a.feedBase = srv.URL
	return a
}

func (f *feed) requests() int { return int(atomic.LoadInt32(&f.hits)) }

// modelsJSON builds a published models file. bidsEach is how many bidDetail
// entries each model carries; 0 means the file carries none, as all_models.json
// does.
func modelsJSON(age time.Duration, bidsEach int, ids ...string) string {
	models := make([]map[string]interface{}, 0, len(ids))
	for i, id := range ids {
		m := map[string]interface{}{
			"Id":        id,
			"Name":      fmt.Sprintf("model-%d", i),
			"IsDeleted": false,
			"ModelType": "LLM",
		}
		if bidsEach > 0 {
			bd := make([]map[string]interface{}, bidsEach)
			for j := range bd {
				bd[j] = map[string]interface{}{"bidId": fmt.Sprintf("0xbid%d%d", i, j), "status": "healthy"}
			}
			m["bidDetail"] = bd
		}
		models = append(models, m)
	}
	body, _ := json.Marshal(map[string]interface{}{
		"last_updated":     time.Now().Add(-age).Unix(),
		"last_updated_utc": time.Now().Add(-age).UTC().Format(time.RFC3339),
		"models":           models,
	})
	return string(body)
}

// bidsJSON builds a published bids file granting each named model one live bid.
func bidsJSON(age time.Duration, modelIDs ...string) string {
	bids := make([]map[string]interface{}, 0, len(modelIDs))
	for i, id := range modelIDs {
		bids = append(bids, map[string]interface{}{
			"Id":             fmt.Sprintf("0xbid%d", i),
			"ModelAgentId":   id,
			"DeletedAt":      "0",
			"PricePerSecond": "1000",
		})
	}
	body, _ := json.Marshal(map[string]interface{}{
		"last_updated": time.Now().Add(-age).Unix(),
		"bids":         bids,
	})
	return string(body)
}

func idsOf(t *testing.T, models []interface{}) []string {
	t.Helper()
	out := make([]string, 0, len(models))
	for _, item := range models {
		m, ok := item.(map[string]interface{})
		if !ok {
			t.Fatalf("result item is %T, want a map", item)
		}
		out = append(out, m["Id"].(string))
	}
	return out
}

// TestFeedViewsFetchTheRightFiles: the three views are different catalogues, and
// only "all" needs a second file, because active and gateway already carry
// bidDetail on every entry.
func TestFeedViewsFetchTheRightFiles(t *testing.T) {
	for _, tc := range []struct {
		view      string
		wantFiles int
	}{
		{modelSourceGateway, 1},
		{modelSourceActive, 1},
		{modelSourceAll, 2},
	} {
		t.Run(tc.view, func(t *testing.T) {
			f := &feed{files: map[string]string{
				"gateway_models.json": modelsJSON(time.Minute, 1, feedModelA),
				"active_models.json":  modelsJSON(time.Minute, 1, feedModelA, feedModelB),
				"all_models.json":     modelsJSON(time.Minute, 0, feedModelA, feedModelB),
				"all_bids.json":       bidsJSON(time.Minute, feedModelA, feedModelB),
			}}
			a := f.start(t)

			models, err := a.loadModelsFromFeed(context.Background(), tc.view)
			if err != nil {
				t.Fatalf("%s view should load: %v", tc.view, err)
			}
			if len(models) == 0 {
				t.Fatal("no models returned")
			}
			if f.requests() != tc.wantFiles {
				t.Errorf("%s view made %d request(s), want %d", tc.view, f.requests(), tc.wantFiles)
			}
		})
	}
}

// TestAllViewDropsModelsWithNoLiveBid is the reason the all view needs the bids
// file: all_models.json carries no bidDetail, and 40 of its ~586 models have no
// live bid. Offering those would break the picker's contract that every model
// listed has something to open against.
func TestAllViewDropsModelsWithNoLiveBid(t *testing.T) {
	f := &feed{files: map[string]string{
		"all_models.json": modelsJSON(time.Minute, 0, feedModelA, feedModelB),
		// Only model A has a bid.
		"all_bids.json": bidsJSON(time.Minute, feedModelA),
	}}
	a := f.start(t)

	models, err := a.loadModelsFromFeed(context.Background(), modelSourceAll)
	if err != nil {
		t.Fatalf("the all view should load: %v", err)
	}
	got := idsOf(t, models)
	if len(got) != 1 || got[0] != feedModelA {
		t.Errorf("only the model with a live bid should be offered, got %v", got)
	}
}

// TestDeletedBidsDoNotCountAsLive guards the DeletedAt comparison, which has to
// treat a nonzero stamp and an absent field as not-live.
func TestDeletedBidsDoNotCountAsLive(t *testing.T) {
	bids, _ := json.Marshal(map[string]interface{}{
		"last_updated": time.Now().Unix(),
		"bids": []map[string]interface{}{
			{"ModelAgentId": feedModelA, "DeletedAt": "1792180763"}, // deleted
			{"ModelAgentId": feedModelB, "DeletedAt": "0"},          // live
			{"DeletedAt": "0"}, // no model id
		},
	})
	f := &feed{files: map[string]string{
		"all_models.json": modelsJSON(time.Minute, 0, feedModelA, feedModelB),
		"all_bids.json":   string(bids),
	}}
	a := f.start(t)

	models, err := a.loadModelsFromFeed(context.Background(), modelSourceAll)
	if err != nil {
		t.Fatalf("should load: %v", err)
	}
	got := idsOf(t, models)
	if len(got) != 1 || got[0] != feedModelB {
		t.Errorf("only the model with a live bid should survive, got %v", got)
	}
}

// TestStaleFeedIsRefused is the core safety property. A stalled generator does
// not fail loudly: the JSON still parses and the prices in it simply go old.
func TestStaleFeedIsRefused(t *testing.T) {
	f := &feed{files: map[string]string{
		"gateway_models.json": modelsJSON(modelsFeedMaxAge+time.Minute, 1, feedModelA),
	}}
	a := f.start(t)

	_, err := a.loadModelsFromFeed(context.Background(), modelSourceGateway)
	if err == nil {
		t.Fatal("a feed older than the limit must be refused")
	}
	if !strings.Contains(err.Error(), "stale") {
		t.Errorf("the error should say it is stale, got %v", err)
	}
}

// TestFeedFreshnessWindow covers the bounds directly, including the future side:
// a timestamp ahead of now means the generator's clock is wrong, and trusting it
// would mask staleness for as long as the skew lasted.
func TestFeedFreshnessWindow(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	for _, tc := range []struct {
		name  string
		at    time.Time
		fresh bool
	}{
		{"just published", now, true},
		{"one cycle old", now.Add(-5 * time.Minute), true},
		{"inside the limit", now.Add(-modelsFeedMaxAge + time.Second), true},
		{"past the limit", now.Add(-modelsFeedMaxAge - time.Second), false},
		{"hours old", now.Add(-6 * time.Hour), false},
		{"far in the future", now.Add(modelsFeedMaxAge + time.Second), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := feedFresh(tc.at.Unix(), now); got != tc.fresh {
				t.Errorf("feedFresh(%s) = %v, want %v", tc.name, got, tc.fresh)
			}
		})
	}

	if feedFresh(0, now) {
		t.Error("a missing last_updated must not read as fresh")
	}
	if feedFresh(-1, now) {
		t.Error("a negative last_updated must not read as fresh")
	}
}

// TestFeedFailuresAreRefusedNotFatal: every way the feed can go wrong has to
// come back as an error the caller can fall back on.
func TestFeedFailuresAreRefusedNotFatal(t *testing.T) {
	fresh := time.Now().Unix()

	for _, tc := range []struct {
		name  string
		feed  *feed
		dead  bool
		files map[string]string
	}{
		{name: "server error", feed: &feed{status: http.StatusInternalServerError}},
		{name: "not found", feed: &feed{files: map[string]string{}}},
		{
			name: "html instead of json",
			feed: &feed{files: map[string]string{"gateway_models.json": "<!doctype html><h1>nope</h1>"}},
		},
		{
			name: "truncated json",
			feed: &feed{files: map[string]string{"gateway_models.json": `{"models":[`}},
		},
		{
			name: "no models key",
			feed: &feed{files: map[string]string{
				"gateway_models.json": fmt.Sprintf(`{"last_updated":%d}`, fresh)}},
		},
		{
			name: "every record unusable",
			feed: &feed{files: map[string]string{
				"gateway_models.json": fmt.Sprintf(
					`{"last_updated":%d,"models":[{"Id":123},{"Name":"no id"},{"Id":"0xshort","Name":"x"}]}`, fresh)}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := tc.feed.start(t)
			if _, err := a.loadModelsFromFeed(context.Background(), modelSourceGateway); err == nil {
				t.Errorf("%s should be refused so the caller can fall back", tc.name)
			}
		})
	}

	t.Run("unreachable host", func(t *testing.T) {
		a := NewActions(nil)
		a.feedBase = "http://127.0.0.1:1"
		if _, err := a.loadModelsFromFeed(context.Background(), modelSourceGateway); err == nil {
			t.Error("an unreachable feed should be refused")
		}
	})
}

// TestJunkRecordsAreSkippedNotPanicked matters more than it looks: the picker
// asserts Name and Id to string without checking, so a malformed entry reaching
// it panics inside the full-screen program.
func TestJunkRecordsAreSkippedNotPanicked(t *testing.T) {
	body := fmt.Sprintf(`{"last_updated":%d,"models":[
		{"Id":%q,"Name":"good","bidDetail":[{"status":"healthy"}]},
		{"Id":%q,"Name":"","bidDetail":[{"status":"healthy"}]},
		{"Id":"0xnothex","Name":"bad id","bidDetail":[{"status":"healthy"}]},
		{"Id":%q,"Name":"deleted","IsDeleted":true,"bidDetail":[{"status":"healthy"}]},
		{"Id":%q,"Name":"no bids"}
	]}`, time.Now().Unix(), feedModelA, feedModelB, feedModelB, feedModelB)

	f := &feed{files: map[string]string{"gateway_models.json": body}}
	a := f.start(t)

	models, err := a.loadModelsFromFeed(context.Background(), modelSourceGateway)
	if err != nil {
		t.Fatalf("one good record should be enough to load: %v", err)
	}
	got := idsOf(t, models)
	if len(got) != 1 || got[0] != feedModelA {
		t.Errorf("only the valid record should survive, got %v", got)
	}
}

// TestFeedResultMatchesTheRouterShape: the picker reads Name and Id off these
// maps with unchecked assertions, and loadModelsWithBids attaches a bids key.
// The feed path has to produce the same shape.
func TestFeedResultMatchesTheRouterShape(t *testing.T) {
	f := &feed{files: map[string]string{
		"gateway_models.json": modelsJSON(time.Minute, 2, feedModelA),
	}}
	a := f.start(t)

	models, err := a.loadModelsFromFeed(context.Background(), modelSourceGateway)
	if err != nil {
		t.Fatalf("should load: %v", err)
	}

	m, ok := models[0].(map[string]interface{})
	if !ok {
		t.Fatalf("entry is %T, want map[string]interface{}", models[0])
	}
	if _, ok := m["Name"].(string); !ok {
		t.Errorf("Name must be a string, got %T -- the picker asserts it unchecked", m["Name"])
	}
	if _, ok := m["Id"].(string); !ok {
		t.Errorf("Id must be a string, got %T -- the picker asserts it unchecked", m["Id"])
	}
	bids, ok := m["bids"].([]interface{})
	if !ok {
		t.Fatalf("bids must be []interface{} for parity with the router path, got %T", m["bids"])
	}
	if len(bids) != 2 {
		t.Errorf("bids should carry one entry per published bid, got %d", len(bids))
	}
}

// TestRouterSourceNeverTouchesTheFeed guards the default: nothing reaches
// active.mor.org unless the operator asks for it.
func TestRouterSourceNeverTouchesTheFeed(t *testing.T) {
	if defaultModelSource != modelSourceRouter {
		t.Fatalf("the default source is %q, want %q", defaultModelSource, modelSourceRouter)
	}

	f := &feed{files: map[string]string{
		"gateway_models.json": modelsJSON(time.Minute, 1, feedModelA),
	}}
	a := f.start(t)

	// The router here is a dead address, so this must fail -- the point is that
	// it fails having never contacted the feed.
	_, _ = a.loadModels(context.Background(), modelSourceRouter)

	if f.requests() != 0 {
		t.Errorf("the router source made %d feed request(s), want 0", f.requests())
	}
}

// TestFeedFallsBackToRouter: a refused feed must hand over to the chain rather
// than failing the command.
func TestFeedFallsBackToRouter(t *testing.T) {
	// A marketplace standing in for the router, and a stale feed beside it.
	mkt := &marketplace{
		models:  []map[string]interface{}{model(feedModelB, "from-router", false)},
		bidsFor: map[string][]map[string]interface{}{feedModelB: {liveBid("0xb1")}},
	}
	a := mkt.start(t)

	stale := &feed{files: map[string]string{
		"gateway_models.json": modelsJSON(modelsFeedMaxAge+time.Hour, 1, feedModelA),
	}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&stale.hits, 1)
		body, ok := stale.files[strings.TrimPrefix(r.URL.Path, "/")]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	a.feedBase = srv.URL

	models, err := a.loadModels(context.Background(), modelSourceGateway)
	if err != nil {
		t.Fatalf("a stale feed should fall back to the router, not fail: %v", err)
	}
	if stale.requests() == 0 {
		t.Error("the feed should have been tried first")
	}
	if names := namesOf(t, models); len(names) != 1 || names[0] != "from-router" {
		t.Errorf("the catalogue should have come from the router, got %v", names)
	}
}

func TestModelSourceValid(t *testing.T) {
	for _, s := range []string{modelSourceRouter, modelSourceAll, modelSourceActive, modelSourceGateway} {
		if !modelSourceValid(s) {
			t.Errorf("%q should be a valid source", s)
		}
	}
	for _, s := range []string{"", "Router", "chain", "feed", "gateway_models.json"} {
		if modelSourceValid(s) {
			t.Errorf("%q should not be a valid source", s)
		}
	}
}

// TestFeedFilesPerView pins which published files each view reads. Mixing them
// up is the mistake the project's own docs warn about: active_models.json is the
// ACTIVE layer, not GATEWAY.
func TestFeedFilesPerView(t *testing.T) {
	for _, tc := range []struct{ view, models, bids string }{
		{modelSourceGateway, "gateway_models.json", ""},
		{modelSourceActive, "active_models.json", ""},
		{modelSourceAll, "all_models.json", "all_bids.json"},
		{modelSourceRouter, "", ""},
		{"nonsense", "", ""},
	} {
		gotModels, gotBids := feedFiles(tc.view)
		if gotModels != tc.models || gotBids != tc.bids {
			t.Errorf("feedFiles(%q) = (%q, %q), want (%q, %q)",
				tc.view, gotModels, gotBids, tc.models, tc.bids)
		}
	}
}

func TestLooksLikeModelID(t *testing.T) {
	valid := "0x" + strings.Repeat("ab", 32)
	for _, tc := range []struct {
		in   string
		want bool
	}{
		{valid, true},
		{"0x" + strings.Repeat("AB", 32), true},  // uppercase hex digits are fine
		{"0X" + strings.Repeat("ab", 32), false}, // an uppercase prefix is not what the feed publishes
		{"", false},
		{"0x", false},
		{strings.Repeat("ab", 32), false},        // no prefix
		{"0x" + strings.Repeat("ab", 31), false}, // too short
		{"0x" + strings.Repeat("ab", 33), false}, // too long
		{"0x" + strings.Repeat("zz", 32), false}, // not hex
	} {
		if got := looksLikeModelID(tc.in); got != tc.want {
			t.Errorf("looksLikeModelID(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

// TestFeedMaxAgeIsSane keeps the threshold tied to the publishing cadence: the
// files are regenerated every five minutes with a ~75s publish lag, so anything
// below ~7 minutes would fall back on a healthy feed, and anything beyond an
// hour would stop catching a stalled generator.
func TestFeedMaxAgeIsSane(t *testing.T) {
	if modelsFeedMaxAge < 7*time.Minute || modelsFeedMaxAge > time.Hour {
		t.Errorf("modelsFeedMaxAge = %s, want between 7m and 1h", modelsFeedMaxAge)
	}
}
