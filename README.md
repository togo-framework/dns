<!-- togo-header -->
<div align="center">
  <img src=".github/assets/togo-mark.svg" alt="togo" height="64" />
  <h1>togo-framework/dns</h1>
  <p>One Provider contract over DNS, reverse proxies and API gateways.</p>
  <p>
    <a href="https://to-go.dev/marketplace"><img src="https://img.shields.io/badge/marketplace-to--go.dev-1FC7DC" alt="marketplace" /></a>
    <a href="https://pkg.go.dev/github.com/togo-framework/dns"><img src="https://pkg.go.dev/badge/github.com/togo-framework/dns.svg" alt="pkg.go.dev" /></a>
    <img src="https://img.shields.io/badge/license-MIT-blue" alt="MIT" />
  </p>
  <p><strong>Part of the <a href="https://to-go.dev">togo</a> framework.</strong></p>
</div>

## Install

```bash
togo install togo-framework/dns
```
<!-- /togo-header -->

`dns` is togo's **DNS / reverse-proxy / API-gateway** subsystem. One `Provider`
contract spans three control planes that normally have nothing in common:

| Capability | Drivers |
|------------|---------|
| Authoritative DNS records | [`dns-cloudflare`](https://github.com/togo-framework/dns-cloudflare) |
| Reverse-proxy hosts | [`dns-npm`](https://github.com/togo-framework/dns-npm) (Nginx Proxy Manager), [`dns-caddy`](https://github.com/togo-framework/dns-caddy) |
| API-gateway routes | [`dns-kong`](https://github.com/togo-framework/dns-kong) (+ Supabase) |

The default `log` driver records intent without touching anything — safe for dev
and tests. Pick a real driver with `DNS_DRIVER` and install its plugin.

## Usage

```go
import (
    "github.com/togo-framework/dns"
    _ "github.com/togo-framework/dns-cloudflare" // registers the driver
)

svc, _ := dns.FromKernel(k)
svc.UpsertRecord(ctx, "example.com", dns.Record{Type: "A", Name: "app", Content: "203.0.113.10", Proxied: true})
svc.UpsertProxyHost(ctx, dns.ProxyHost{Domain: "app.example.com", Upstream: "http://127.0.0.1:8080", SSL: true})
svc.UpsertRoute(ctx, dns.Route{Domain: "api.example.com", Path: "/v1", Upstream: "http://gateway:8000"})
```

Operations a driver doesn't support return `dns.ErrUnsupported`.

## Config

| Env | Meaning |
|-----|---------|
| `DNS_DRIVER` | `log` (default), `cloudflare`, `npm`, `caddy`, `kong` |

Each driver reads its own credentials — see that plugin's README.

<!-- togo-sponsors -->
---
<div align="center">
  <h3>Premium sponsors</h3>
  <p><a href="https://id8media.com"><strong>ID8 Media</strong></a> &nbsp;·&nbsp; <a href="https://one-studio.co"><strong>One Studio</strong></a></p>
  <p><sub>Support togo — <a href="https://github.com/sponsors/fadymondy">become a sponsor</a>.</sub></p>
</div>
<!-- /togo-sponsors -->
