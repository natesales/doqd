# doqd

[![Go Report](https://goreportcard.com/badge/github.com/natesales/doqd?style=for-the-badge)](https://goreportcard.com/report/github.com/natesales/doqd)
[![License](https://img.shields.io/github/license/natesales/doqd?style=for-the-badge)](https://raw.githubusercontent.com/natesales/doqd/main/LICENSE)

DNS over QUIC implementation in Go
([draft-ietf-dprive-dnsoquic-02](https://datatracker.ietf.org/doc/draft-ietf-dprive-dnsoquic/?include_text=1))

This repo contains a client and server in conformance with [draft-ietf-dprive-dnsoquic-02](https://datatracker.ietf.org/doc/draft-ietf-dprive-dnsoquic/?include_text=1). Each acts as a proxy for queries from DoQ to plain DNS and vice versa.

### Quickstart

Start the server proxy

```bash
doqd server --cert cert.pem --key key.pem --upstream localhost:53 --listen localhost:8853
INFO[0000] starting QUIC listener on localhost:8853
```

Start the client proxy

```bash
doqd client --listen localhost:53 --upstream localhost:8853
INFO[0000] starting UDP listener on localhost:53
```

Query with dig through the client proxy

```bash
dig +short natesales.net @localhost
23.141.112.33
```

Query directly with the [q DNS client](https://github.com/natesales/q):

```bash
q A natesales.net @quic://localhost:8853
natesales.net. 9h40m58s A 23.141.112.33
```

### Interoperability

This DoQ implementation is designed to be in conformance with `draft-ietf-dprive-dnsoquic-02`, and therefore only offers the `doq-i02` TLS ALPN token. For experimental interop testing, `doq.Server` and `doq.Client` can be created with the `compat` parameter set to true to enable compatibility of other ALPN tokens.

### Tuning

As per the [quic-go wiki](https://github.com/lucas-clemente/quic-go/wiki/UDP-Receive-Buffer-Size), quic-go recommends increasing the maximum UDP receive buffer size and will show a warning if this value is too small. For DNS queries where the packet sizes are small to begin with, increasing the value won't yield a performance improvement so this is up to the operator.

### Local TLS

QUIC requires a TLS certificate. OpenSSL can be used to generate a self-signed local development cert:

```bash
openssl req -x509 -newkey rsa:4096 -sha256 -days 356 -nodes -keyout /tmp/key.pem -out /tmp/cert.pem -subj "/CN=localhost"
```
