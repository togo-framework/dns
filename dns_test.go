package dns

import (
	"context"
	"testing"

	"github.com/togo-framework/togo"
)

func TestLogProviderRoundTrip(t *testing.T) {
	p := &logProvider{}
	if id, err := p.UpsertRecord(context.Background(), "example.com", Record{Type: "A", Name: "@", Content: "1.2.3.4"}); err != nil || id == "" {
		t.Fatalf("upsert record: id=%q err=%v", id, err)
	}
	if id, err := p.UpsertProxyHost(context.Background(), ProxyHost{Domain: "app.example.com", Upstream: "http://127.0.0.1:8080"}); err != nil || id == "" {
		t.Fatalf("upsert proxy: id=%q err=%v", id, err)
	}
	if id, err := p.UpsertRoute(context.Background(), Route{Domain: "api.example.com", Path: "/", Upstream: "http://svc:9000"}); err != nil || id == "" {
		t.Fatalf("upsert route: id=%q err=%v", id, err)
	}
}

func TestRegisterDriver(t *testing.T) {
	RegisterDriver("unit", func(k *togo.Kernel) (Provider, error) { return &logProvider{}, nil })
	found := false
	for _, n := range Drivers() {
		if n == "unit" {
			found = true
		}
	}
	if !found {
		t.Fatal("driver not registered")
	}
}
