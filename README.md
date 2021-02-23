# doq

[![Go Report](https://goreportcard.com/badge/github.com/natesales/doq?style=for-the-badge)](https://goreportcard.com/report/github.com/natesales/doq)
[![License](https://img.shields.io/github/license/natesales/doq?style=for-the-badge)](https://raw.githubusercontent.com/natesales/doq/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/natesales/doq?style=for-the-badge)](https://github.com/natesales/doq/releases)

DNS over QUIC implementation in Go
([draft-ietf-dprive-dnsoquic-02](https://datatracker.ietf.org/doc/draft-ietf-dprive-dnsoquic/?include_text=1))

### Overview

The `pkg` directory contains a DoQ client and server implementation in conformance
with ([draft-ietf-dprive-dnsoquic-02](https://datatracker.ietf.org/doc/draft-ietf-dprive-dnsoquic/?include_text=1)), as
well as a UDP-DoQ proxy (`cmd/server`), and a CLI client (`cmd/client`).

### Interoperability

This DoQ implementation is designed to be in conformance with `draft-ietf-dprive-dnsoquic-02`, and therefore only
announces the `doq-i02` TLS protocol. For experimental interop testing, `doq.Server` and `doq.Client` can be created
with the `compat` parameter set to true to enable `doq-i02`, `dq`, and `doq` TLS protocols.

### Sysctl tuning

As per the [quic-go wiki](https://github.com/lucas-clemente/quic-go/wiki/UDP-Receive-Buffer-Size), quic-go recommends
increasing the maximum UDP receive buffer size and will show a warning if this value is too small. For DNS queries where
the packet sizes are small to begin with, increasing the value won't yield a performance improvement.

### TLS

QUIC requires a TLS certificate. OpenSSL can be used to generate a self-signed local development cert:

```bash
openssl req -x509 -newkey rsa:4096 -sha256 -days 356 -nodes -keyout key.pem -out cert.pem -subj "/CN=localhost"
```