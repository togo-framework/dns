// Package dns is togo's DNS / reverse-proxy / API-gateway subsystem: a single
// Provider contract over wildly different control planes — authoritative DNS
// (Cloudflare), reverse proxies (Nginx Proxy Manager, Caddy) and API gateways
// (Kong). The safe dev default "log" driver records intent without touching a
// real control plane. Real drivers ship as plugins that call dns.RegisterDriver
// and depend on this package. Select one with DNS_DRIVER.
//
// Install: `togo install togo-framework/dns` (blank-import registers it), then a
// driver, e.g. `togo install togo-framework/dns-cloudflare`.
package dns

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/togo-framework/togo"
)

// ErrUnsupported is returned by a driver for operations it does not implement
// (a DNS provider has no proxy hosts; a reverse proxy has no DNS records).
var ErrUnsupported = errors.New("dns: operation not supported by this driver")

// Record is an authoritative DNS record within a zone.
type Record struct {
	ID      string // provider record id (set on read/upsert)
	Type    string // A, AAAA, CNAME, TXT, MX, NS, SRV…
	Name    string // host within the zone ("@" for apex, "www", …)
	Content string // value (IP, target host, text…)
	TTL     int     // seconds; 1 = automatic on most providers
	Proxied bool    // provider-side proxy/CDN (Cloudflare orange-cloud)
	Prio    int     // priority for MX/SRV
}

// ProxyHost maps an inbound domain to an upstream origin on a reverse proxy.
type ProxyHost struct {
	ID       string
	Domain   string // public hostname, e.g. app.example.com
	Upstream string // origin URL, e.g. http://127.0.0.1:8080
	SSL      bool    // request/force TLS (Let's Encrypt where supported)
}

// Route maps an inbound host/path to an upstream service on an API gateway.
type Route struct {
	ID       string
	Domain   string         // host to match
	Path     string         // path prefix to match ("/" for all)
	Upstream string         // upstream service URL
	Plugins  map[string]any // gateway-specific plugin config (rate-limit, auth…)
}

// Provider is implemented by driver plugins. Implement the operations your
// control plane supports; return ErrUnsupported for the rest.
type Provider interface {
	// DNS records
	UpsertRecord(ctx context.Context, zone string, r Record) (id string, err error)
	DeleteRecord(ctx context.Context, zone, id string) error
	ListRecords(ctx context.Context, zone string) ([]Record, error)
	// Reverse-proxy hosts
	UpsertProxyHost(ctx context.Context, h ProxyHost) (id string, err error)
	DeleteProxyHost(ctx context.Context, id string) error
	// API-gateway routes
	UpsertRoute(ctx context.Context, rt Route) (id string, err error)
	DeleteRoute(ctx context.Context, id string) error
}

// DriverFactory builds a Provider from the kernel (env-configured).
type DriverFactory func(k *togo.Kernel) (Provider, error)

var (
	regMu   sync.RWMutex
	drivers = map[string]DriverFactory{}
)

// RegisterDriver registers a DNS driver by name (call from a plugin's init()).
func RegisterDriver(name string, f DriverFactory) {
	regMu.Lock()
	drivers[name] = f
	regMu.Unlock()
}

// Drivers lists the registered driver names (unordered).
func Drivers() []string {
	regMu.RLock()
	defer regMu.RUnlock()
	out := make([]string, 0, len(drivers))
	for n := range drivers {
		out = append(out, n)
	}
	return out
}

func init() {
	RegisterDriver("log", func(k *togo.Kernel) (Provider, error) { return &logProvider{k: k}, nil })

	togo.RegisterProviderFunc("dns", togo.PriorityService, func(k *togo.Kernel) error {
		name := os.Getenv("DNS_DRIVER")
		if name == "" {
			name = "log" // safe dev default: record intent, change nothing
		}
		regMu.RLock()
		f, ok := drivers[name]
		regMu.RUnlock()
		if !ok {
			return fmt.Errorf("dns: unknown driver %q (install its plugin, e.g. togo install togo-framework/dns-%s)", name, name)
		}
		p, err := f(k)
		if err != nil {
			return err
		}
		k.Set("dns", &Service{provider: p, driver: name})
		return nil
	})
}

// Build constructs a Service for the named driver without booting the kernel.
// The togo CLI's `proxy` runner uses it to resolve a provider standalone; the
// stock drivers (cloudflare/npm/caddy/kong) are env-configured, so a nil kernel
// is fine. Pass a real kernel only if a driver needs kernel services.
func Build(name string, k *togo.Kernel) (*Service, error) {
	if name == "" {
		name = "log"
	}
	regMu.RLock()
	f, ok := drivers[name]
	regMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("dns: unknown driver %q (install its plugin, e.g. togo install togo-framework/dns-%s)", name, name)
	}
	p, err := f(k)
	if err != nil {
		return nil, err
	}
	return &Service{provider: p, driver: name}, nil
}

// Service is the dns runtime stored on the kernel (k.Get("dns")).
type Service struct {
	provider Provider
	driver   string
}

// Provider returns the active driver implementation.
func (s *Service) Provider() Provider { return s.provider }

// Driver returns the active driver name.
func (s *Service) Driver() string { return s.driver }

func (s *Service) UpsertRecord(ctx context.Context, zone string, r Record) (string, error) {
	return s.provider.UpsertRecord(ctx, zone, r)
}
func (s *Service) DeleteRecord(ctx context.Context, zone, id string) error {
	return s.provider.DeleteRecord(ctx, zone, id)
}
func (s *Service) ListRecords(ctx context.Context, zone string) ([]Record, error) {
	return s.provider.ListRecords(ctx, zone)
}
func (s *Service) UpsertProxyHost(ctx context.Context, h ProxyHost) (string, error) {
	return s.provider.UpsertProxyHost(ctx, h)
}
func (s *Service) DeleteProxyHost(ctx context.Context, id string) error {
	return s.provider.DeleteProxyHost(ctx, id)
}
func (s *Service) UpsertRoute(ctx context.Context, rt Route) (string, error) {
	return s.provider.UpsertRoute(ctx, rt)
}
func (s *Service) DeleteRoute(ctx context.Context, id string) error {
	return s.provider.DeleteRoute(ctx, id)
}

// FromKernel fetches the dns service from the kernel container.
func FromKernel(k *togo.Kernel) (*Service, bool) {
	v, ok := k.Get("dns")
	if !ok {
		return nil, false
	}
	s, ok := v.(*Service)
	return s, ok
}

// logProvider records intent via the kernel logger — the safe default.
type logProvider struct{ k *togo.Kernel }

func (l *logProvider) log(op string, kv ...any) {
	if l.k != nil && l.k.Log != nil {
		l.k.Log.Info("dns (log driver) "+op, kv...)
	}
}
func (l *logProvider) UpsertRecord(_ context.Context, zone string, r Record) (string, error) {
	l.log("upsert record", "zone", zone, "type", r.Type, "name", r.Name, "content", r.Content)
	return "log_record", nil
}
func (l *logProvider) DeleteRecord(_ context.Context, zone, id string) error {
	l.log("delete record", "zone", zone, "id", id)
	return nil
}
func (l *logProvider) ListRecords(context.Context, string) ([]Record, error) { return nil, nil }
func (l *logProvider) UpsertProxyHost(_ context.Context, h ProxyHost) (string, error) {
	l.log("upsert proxy host", "domain", h.Domain, "upstream", h.Upstream)
	return "log_proxy", nil
}
func (l *logProvider) DeleteProxyHost(_ context.Context, id string) error {
	l.log("delete proxy host", "id", id)
	return nil
}
func (l *logProvider) UpsertRoute(_ context.Context, rt Route) (string, error) {
	l.log("upsert route", "domain", rt.Domain, "path", rt.Path, "upstream", rt.Upstream)
	return "log_route", nil
}
func (l *logProvider) DeleteRoute(_ context.Context, id string) error {
	l.log("delete route", "id", id)
	return nil
}
