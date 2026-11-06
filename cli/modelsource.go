package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Where the model picker gets its catalogue from.
//
// The router path walks the chain: one bid lookup per model, 500+ of them, which
// is ~99s even with those lookups running concurrently and grows with the
// network. active.mor.org does that walk on a five-minute cron and publishes
// three curated views of the same sweep, so an app does not have to repeat it.
//
// The three views nest -- gateway subset of active subset of all -- so the flag
// is a widening dial. Their meanings are the project's, from
// docs/ecosystem/active-status.mdx, and they are about *bid health*, not model
// type:
//
//	all      non-deleted models and open bids, with a health stamp on each.
//	         Includes failures, which is the point: it is the ops view.
//	active   the polled working set: skipped, degraded and healthy bids.
//	gateway  the tested/openable set: healthy and degraded bids only.
//
// Note "gateway" keeps degraded as well as healthy. There happen to be no
// degraded bids most of the time, which makes it look like a healthy-only view;
// do not rely on that.
const (
	modelSourceRouter  = "router"
	modelSourceAll     = "all"
	modelSourceActive  = "active"
	modelSourceGateway = "gateway"

	// defaultModelSource keeps the chain as the source of truth unless the
	// operator asks otherwise. Nothing reaches active.mor.org by default, so
	// offline, air-gapped and private deployments are unaffected.
	defaultModelSource = modelSourceRouter
)

const (
	// defaultModelsFeedBase is overridden by MODELS_FEED_URL, which exists for
	// mirrors and for pointing tests at a local server.
	defaultModelsFeedBase = "https://active.mor.org"

	// modelsFeedTimeout bounds the whole fetch. The published files are ~60-95
	// KB on the wire and answer in about a second, so this is generous; it is
	// here because the API client is built with a zero-value http.Client that
	// has no timeout at all, and a feed that never answers must not hang the
	// picker.
	modelsFeedTimeout = 20 * time.Second

	// modelsFeedMaxAge is how stale a published sweep may be before the router
	// is used instead.
	//
	// The files are regenerated on a five-minute cron: four samples taken over
	// eleven days all carried a last_updated on a minute ≡ 4 (mod 5) at :10
	// seconds, with a ~75s gap to the file's Last-Modified. So a healthy feed
	// is at most ~6.5 minutes old, and one missed cycle ~11.5. Fifteen minutes
	// therefore tolerates a missed cycle with margin while still catching a
	// generator that has stopped -- which is the real risk, because a stalled
	// feed still parses perfectly and simply reports old prices.
	modelsFeedMaxAge = 15 * time.Minute

	// modelsFeedMaxBytes caps what is read from the feed. The largest file is
	// under 1 MB uncompressed today; this is a sanity bound so a misbehaving
	// host cannot make the CLI read forever.
	modelsFeedMaxBytes = 32 << 20 // 32 MiB
)

// errFeedStale reports a published sweep too old to use.
var errFeedStale = errors.New("models feed is stale")

// modelSourceValid reports whether s names a catalogue source.
func modelSourceValid(s string) bool {
	switch s {
	case modelSourceRouter, modelSourceAll, modelSourceActive, modelSourceGateway:
		return true
	}
	return false
}

// feedFiles returns the published files a view needs.
//
// Only "all" needs two. active_models.json and gateway_models.json carry a
// bidDetail array on every entry, so the bids are already there; all_models.json
// carries none, and 40 of its ~586 models have no live bid at all. Listing those
// would break the picker's standing contract that every model offered has a bid
// to open against, so the all view joins against all_bids.json.
func feedFiles(view string) (modelsFile, bidsFile string) {
	switch view {
	case modelSourceAll:
		return "all_models.json", "all_bids.json"
	case modelSourceActive:
		return "active_models.json", ""
	case modelSourceGateway:
		return "gateway_models.json", ""
	}
	return "", ""
}

// feedEnvelope is the shape every published file shares. Only the fields used
// here are declared; the feed carries more (health, enrichment) and unknown
// fields decode away harmlessly.
type feedEnvelope struct {
	LastUpdated int64             `json:"last_updated"`
	Models      []feedRecord      `json:"models"`
	Bids        []json.RawMessage `json:"bids"`
}

// feedRecord is one model as published. Name and Id are the only fields the
// picker reads, and both must survive validation, because the picker asserts
// their types without checking and panics on anything else.
type feedRecord struct {
	ID        string            `json:"Id"`
	Name      string            `json:"Name"`
	IsDeleted bool              `json:"IsDeleted"`
	ModelType string            `json:"ModelType"`
	BidDetail []json.RawMessage `json:"bidDetail"`
}

// feedBid is the subset of a published bid needed to decide whether a model has
// something open to session against.
type feedBid struct {
	ModelAgentID string `json:"ModelAgentId"`
	DeletedAt    string `json:"DeletedAt"`
}

// feedFresh reports whether a published sweep is recent enough to trust.
//
// The window is two-sided on purpose: a timestamp from the future means the
// generator's clock is wrong, and trusting it would hide staleness for as long
// as the skew lasts.
func feedFresh(lastUpdated int64, now time.Time) bool {
	if lastUpdated <= 0 {
		return false
	}
	published := time.Unix(lastUpdated, 0)
	return published.After(now.Add(-modelsFeedMaxAge)) && published.Before(now.Add(modelsFeedMaxAge))
}

// feedHTTPClient is the client used for the feed, and never the API client.
//
// client.getRequest cannot be reused for this: it prepends the router's base URL
// and attaches the router's basic-auth header unconditionally, which would send
// router credentials to a third-party host.
func feedHTTPClient() *http.Client {
	return &http.Client{
		Timeout: modelsFeedTimeout,
		// A catalogue host has no business redirecting the fetch elsewhere.
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 0 && req.URL.Host != via[0].URL.Host {
				return fmt.Errorf("refusing cross-host redirect to %s", req.URL.Host)
			}
			return nil
		},
	}
}

// fetchFeedFile reads and decodes one published file.
func fetchFeedFile(ctx context.Context, base, name string) (*feedEnvelope, error) {
	url := base + "/" + name

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request for %s: %w", name, err)
	}
	// Accept-Encoding is left to the transport, which asks for gzip and
	// decompresses transparently: ~860 KB becomes ~92 KB on the wire.

	resp, err := feedHTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned status %d", name, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, modelsFeedMaxBytes))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", name, err)
	}

	var env feedEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("decode %s: %w", name, err)
	}
	return &env, nil
}

// modelsWithLiveBids returns the ids from a bids file that have at least one bid
// still open. A deleted bid carries a nonzero deletion timestamp; compared as
// text because the field has been seen as both a JSON string and a number, and
// an absent value must not count as live.
func modelsWithLiveBids(raw []json.RawMessage) map[string]int {
	live := make(map[string]int, len(raw))
	for _, item := range raw {
		var bid feedBid
		if err := json.Unmarshal(item, &bid); err != nil {
			continue
		}
		if bid.ModelAgentID == "" || fmt.Sprint(bid.DeletedAt) != "0" {
			continue
		}
		live[bid.ModelAgentID]++
	}
	return live
}

// usableFeedModels converts published records into the shape the picker already
// consumes, dropping anything it could not safely display.
//
// Validation is not defensiveness for its own sake: chat/chat/new.go asserts
// Name and Id to string without checking, so one malformed entry would panic
// inside the full-screen program. Junk entries are skipped the way
// loadModelsWithBids skips them rather than failing the whole catalogue.
func usableFeedModels(view string, models *feedEnvelope, bids *feedEnvelope) []interface{} {
	var liveBidCount map[string]int
	if bids != nil {
		liveBidCount = modelsWithLiveBids(bids.Bids)
	}

	out := make([]interface{}, 0, len(models.Models))
	for _, rec := range models.Models {
		if rec.IsDeleted || rec.Name == "" || !looksLikeModelID(rec.ID) {
			continue
		}

		// Every model offered must have something to open against. The active
		// and gateway views are already curated to that; all is not.
		bidCount := len(rec.BidDetail)
		if liveBidCount != nil {
			bidCount = liveBidCount[rec.ID]
		}
		if bidCount == 0 {
			continue
		}

		out = append(out, map[string]interface{}{
			"Id":        rec.ID,
			"Name":      rec.Name,
			"IsDeleted": false,
			"ModelType": rec.ModelType,
			// Kept for shape parity with the router path, which attaches the
			// bids it fetched. Nothing reads this today; it is the
			// has-a-live-bid signal.
			"bids": make([]interface{}, bidCount),
		})
	}
	return out
}

// looksLikeModelID reports whether s is a 32-byte hex model id.
func looksLikeModelID(s string) bool {
	if len(s) != 66 || s[:2] != "0x" {
		return false
	}
	for _, c := range s[2:] {
		isHex := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
		if !isHex {
			return false
		}
	}
	return true
}

// loadModelsFromFeed returns the models for a view, or an error explaining why
// the caller should use the router instead.
//
// Every failure is a fallback, not a fatal: a feed that is unreachable,
// malformed, stale, or empty after validation means the chain is used. The feed
// is an accelerator for discovery, never a dependency, and it never decides what
// gets staked -- the router rates bids on chain when the session is opened.
func (a *actions) loadModelsFromFeed(ctx context.Context, view string) ([]interface{}, error) {
	modelsFile, bidsFile := feedFiles(view)
	if modelsFile == "" {
		return nil, fmt.Errorf("unknown models source %q", view)
	}

	base := a.feedBase
	if base == "" {
		base = defaultModelsFeedBase
	}

	models, err := fetchFeedFile(ctx, base, modelsFile)
	if err != nil {
		return nil, err
	}

	if !feedFresh(models.LastUpdated, time.Now()) {
		published := "never"
		if models.LastUpdated > 0 {
			published = time.Unix(models.LastUpdated, 0).UTC().Format(time.RFC3339)
		}
		return nil, fmt.Errorf("%w: published %s, limit is %s", errFeedStale, published, modelsFeedMaxAge)
	}

	var bids *feedEnvelope
	if bidsFile != "" {
		bids, err = fetchFeedFile(ctx, base, bidsFile)
		if err != nil {
			return nil, err
		}
	}

	usable := usableFeedModels(view, models, bids)
	if len(usable) == 0 {
		return nil, fmt.Errorf("%s carried no usable models", modelsFile)
	}

	fmt.Printf("Loaded %d models from the %s view of %s, published %s ago\n",
		len(usable), view, base, time.Since(time.Unix(models.LastUpdated, 0)).Round(time.Second))

	return usable, nil
}
