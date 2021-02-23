# doq

[![Go Report](https://goreportcard.com/badge/github.com/natesales/doq?style=for-the-badge)](https://goreportcard.com/report/github.com/natesales/doq)
[![License](https://img.shields.io/github/license/natesales/doq?style=for-the-badge)](https://raw.githubusercontent.com/natesales/doq/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/natesales/doq?style=for-the-badge)](https://github.com/natesales/doq/releases)

DNS over QUIC implementation in Go ([draft-huitema-quic-dnsoquic-07](https://tools.ietf.org/html/draft-huitema-quic-dnsoquic-07))

### Setup

##### Create a self-signed TLS certificate
```bash
openssl req -x509 -newkey rsa:4096 -sha256 -days 356 -nodes -keyout key.pem -out cert.pem -subj "/CN=localhost"
```

##### Sysctl tuning
As per the [quic-go wiki](https://github.com/lucas-clemente/quic-go/wiki/UDP-Receive-Buffer-Size), quic-go recommends increasing the maximum UDP receive buffer size and will show a warning if this value is too small. For DNS queries where the packet sizes are small to begin with, increasing the value won't yield a performance improvement.
