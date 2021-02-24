# doq

[![Go Report](https://goreportcard.com/badge/github.com/natesales/doq?style=for-the-badge)](https://goreportcard.com/report/github.com/natesales/doq)
[![License](https://img.shields.io/github/license/natesales/doq?style=for-the-badge)](https://raw.githubusercontent.com/natesales/doq/main/LICENSE)

DNS over QUIC implementation in Go
([draft-ietf-dprive-dnsoquic-02](https://datatracker.ietf.org/doc/draft-ietf-dprive-dnsoquic/?include_text=1))

The `pkg` directory contains a DoQ client and server implementation in conformance
with [draft-ietf-dprive-dnsoquic-02](https://datatracker.ietf.org/doc/draft-ietf-dprive-dnsoquic/?include_text=1), as
well as a UDP-DoQ proxy (`cmd/server`), a CLI client (`cmd/client`), and a DoQ-UDP proxy (`cmd/clientproxy`).

### Quickstart

Start the DoQ server
```shell
➜ sudo ./server -backend 9.9.9.9:53 -tlsCompat -listen localhost:784 -tlsCompat
INFO[0000] starting quic listener on quic://localhost:784
```

Query with the DoQ client
```
➜ ./client -server localhost:784 -insecureSkipVerify -queryName natesales.net -queryType A
;; opcode: QUERY, status: NOERROR, id: 50003
;; flags: qr rd ra; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 1

;; QUESTION SECTION:
;natesales.net.	IN	 A

;; ANSWER SECTION:
natesales.net.	42341	IN	A	23.141.112.33

;; ADDITIONAL SECTION:

;; OPT PSEUDOSECTION:
; EDNS: version 0; flags: do; udp: 1232
```

Start the client proxy
```bash
➜ ./clientproxy -listen localhost:6000 -server localhost:784 -insecureSkipVerify
INFO[0000] opening DoQ connection to localhost:784
INFO[0000] starting UDP listener on localhost:6000
```

Query with dig through the client proxy
```bash
➜ dig natesales.net @localhost -p 6000
; <<>> DiG 9.11.5-P4-5.1+deb10u3-Debian <<>> natesales.net @localhost -p 6000
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 34905
;; flags: qr rd ra; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 1

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 1232
;; QUESTION SECTION:
;natesales.net.			IN	A

;; ANSWER SECTION:
natesales.net.		41624	IN	A	23.141.112.33

;; Query time: 24 msec
;; SERVER: 127.0.0.1#6000(127.0.0.1)
;; WHEN: Wed Feb 24 00:07:45 PST 2021
;; MSG SIZE  rcvd: 71
```

### Interoperability

This DoQ implementation is designed to be in conformance with `draft-ietf-dprive-dnsoquic-02`, and therefore only
announces the `doq-i02` TLS protocol. For experimental interop testing, `doq.Server` and `doq.Client` can be created
with the `compat` parameter set to true to enable backwards compatibility of TLS protocols.

### Sysctl tuning

As per the [quic-go wiki](https://github.com/lucas-clemente/quic-go/wiki/UDP-Receive-Buffer-Size), quic-go recommends
increasing the maximum UDP receive buffer size and will show a warning if this value is too small. For DNS queries where
the packet sizes are small to begin with, increasing the value won't yield a performance improvement.

### TLS

QUIC requires a TLS certificate. OpenSSL can be used to generate a self-signed local development cert:

```bash
openssl req -x509 -newkey rsa:4096 -sha256 -days 356 -nodes -keyout key.pem -out cert.pem -subj "/CN=localhost"
```
