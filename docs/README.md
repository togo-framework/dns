# dns — docs

**DNS / Proxy / Gateway.** Manage DNS records, reverse-proxy hosts and API-gateway routes through one pluggable driver.

## Install

```bash
togo install togo-framework/dns
```

Select a driver via **dns.provider in togo.yaml (or DNS_DRIVER)**. Drive it from the CLI with **`togo proxy`**.

## Interface

`Provider` — `UpsertRecord`/`DeleteRecord`/`ListRecords`, `UpsertProxyHost`/`DeleteProxyHost`, `UpsertRoute`/`DeleteRoute`.

## Configuration

| Env var | Description |
|---|---|
| `DNS_DRIVER` | Selects the DNS/proxy driver (alternative to togo.yaml `dns.provider`). |

## Usage & notes

Powers `togo proxy`. Drivers self-register; the CLI resolves with `dns.Build(name, k)`. Configure in `togo.yaml`:
```yaml
dns:
  provider: cloudflare   # cloudflare|npm|caddy|kong
  zone: example.com
```
Providers implement the subset they support and return `ErrUnsupported` for the rest.

## Example

```bash
togo proxy:host:add app.example.com http://localhost:3000 --provider cloudflare --dry-run
```

## Links

- [Driver plugins](https://to-go.dev/marketplace)
- [Marketplace](https://to-go.dev/marketplace)
- [Source](https://github.com/togo-framework/dns)
