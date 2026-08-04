package attestation

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestDeriveAttestationURL_Phase1Port(t *testing.T) {
	got, err := DeriveAttestationURL("https://provider.example.com:3333/v1")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://provider.example.com:29343"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestDeriveBackendAttestationURL_Phase2Port(t *testing.T) {
	got, err := DeriveBackendAttestationURL("https://secretai-rytn.scrtlabs.com:21434/v1")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://secretai-rytn.scrtlabs.com:21434"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// overridePorts swaps the probed attestation ports for the duration of a test.
func overridePorts(t *testing.T, ports []string) {
	t.Helper()
	old := backendAttestationPorts
	backendAttestationPorts = ports
	t.Cleanup(func() { backendAttestationPorts = old })
}

func TestResolveAttestationURL_FallsBackToSecondPort(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cpu", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "quote")
	})
	server := httptest.NewTLSServer(mux)
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	// Port 1 refuses connections; the resolver must fall back to the live port.
	overridePorts(t, []string{"1", serverURL.Port()})

	bv := NewBackendVerifier("http://unused", nil, nil, nil, testLog())
	got, err := bv.ResolveAttestationURL(context.Background(), "https://127.0.0.1:9999/v1")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://127.0.0.1:" + serverURL.Port()
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestResolveAttestationURL_PrimaryPortWins(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cpu", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "quote")
	})
	server := httptest.NewTLSServer(mux)
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	overridePorts(t, []string{serverURL.Port(), "1"})

	bv := NewBackendVerifier("http://unused", nil, nil, nil, testLog())
	got, err := bv.ResolveAttestationURL(context.Background(), "https://127.0.0.1:9999/v1")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://127.0.0.1:" + serverURL.Port()
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestResolveAttestationURL_NoPortAnswers_ReturnsPrimary(t *testing.T) {
	// Both ports refuse connections; the resolver must return the primary URL
	// so the subsequent attestation surfaces a clear error against it.
	overridePorts(t, []string{"1", "2"})

	bv := NewBackendVerifier("http://unused", nil, nil, nil, testLog())
	got, err := bv.ResolveAttestationURL(context.Background(), "https://127.0.0.1:9999/v1")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://127.0.0.1:1"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
