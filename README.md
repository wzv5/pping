# pping

[![GoDoc](https://godoc.org/github.com/wzv5/pping?status.svg)](https://godoc.org/github.com/wzv5/pping)

tcp ping, tls ping, http ping, icmp ping, dns ping.

## Install

<https://github.com/wzv5/pping/releases/latest>

Or use [Scoop](https://scoop.sh):

``` text
scoop bucket add wzv5 https://github.com/wzv5/ScoopBucket
scoop install wzv5/pping
```

## Usage

``` text
$ pping
Usage:
  pping [command]

Available Commands:
  dns         dns ping
  help        Help about any command
  http        http ping
  icmp        icmp ping
  tcp         tcp ping
  tls         tls ping

Flags:
  -c, --count int           number of requests to send (default 4)
  -h, --help                help for pping
  -t, --infinite            ping the specified target until stopped
  -i, --interval duration   delay between each request (default 1s)
  -4, --ipv4                use IPv4
  -6, --ipv6                use IPv6
  -v, --version             version for pping

Use "pping [command] --help" for more information about a command.
```

tls ping (dns over tls):

``` text
$ pping tls 223.5.5.5 -p 853
Ping 223.5.5.5 (223.5.5.5):
16:58:28 [1] proto = TLS 1.3, connection = 20 ms, handshake = 22 ms, time = 42 ms
16:58:29 [2] proto = TLS 1.3, connection = 18 ms, handshake = 24 ms, time = 42 ms
16:58:31 [3] proto = TLS 1.3, connection = 19 ms, handshake = 25 ms, time = 44 ms
16:58:32 [4] proto = TLS 1.3, connection = 21 ms, handshake = 26 ms, time = 47 ms

        sent = 4, ok = 4, failed = 0 (0%)
        min = 42 ms, max = 47 ms, avg = 43 ms
```

http ping (sni proxy):

``` text
$ pping http https://www.google.com 127.0.0.2
Ping https://www.google.com:
16:59:34 [1] proto = HTTP/2.0, status = 200, length = 211727, time = 1105 ms
16:59:36 [2] proto = HTTP/2.0, status = 200, length = 211791, time = 1246 ms
16:59:38 [3] proto = HTTP/2.0, status = 200, length = 211721, time = 1159 ms
16:59:40 [4] proto = HTTP/2.0, status = 200, length = 211717, time = 1142 ms

        sent = 4, ok = 4, failed = 0 (0%)
        min = 1105 ms, max = 1246 ms, avg = 1163 ms
```
