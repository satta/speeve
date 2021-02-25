# ⛲ speeve

speeve is a fast, probabilistic [EVE-JSON](https://suricata.readthedocs.io/en/latest/output/eve/eve-json-format.html) generator for testing and benchmarking of EVE-consuming applications. It can generate up to hundreds of thousands of events per second, with varying IPs, ports, timestamps, accurate [Community ID](https://github.com/corelight/community-id-spec) values, etc. under a predefined traffic profile model. Additional metadata can be added based on static strings, [Go templates](https://golang.org/pkg/text/template/) and [Tengo](https://github.com/d5/tengo) scripts. 

🚧 Very alpha - work in progress! 🚧

## Building

Build dependency:
* `libpcap-dev`

Like any good Go program:

```
$ go get -t ./...
$ go build ./...
$ go install -v ./...
...
$ speeve spew -h
```

## Traffic profile configuration

```yaml
providers:
    - name: alert
      type: static
      event_type: alert
      ipranges:
        src: 10.0.0.0/16
        dst: 10.0.0.0/29
      ports:
        src: [3452, 32333]
        dst: [80]
      proto: 6
      static: '{"action": "allowed", "gid": 1, "signature_id": 2001999, "rev": 9, "signature": "ET MALWARE BTGrab.com Spyware Downloading Ads", "category": "A Network Trojan was detected", "severity": 1}'
      weight: 100
    - name: tengotest
      type: tengo
      event_type: foo
      ipranges:
        src: 10.0.0.0/27
        dst: 10.0.0.0/27
      ports:
        src: [1125, 1232, 39488]
        dst: [80]
      proto: 6
      tengo: >
        encoded := "{\"foo\": 42}"
      weight: 2
    - name: dns
      type: template
      event_type: dns
      ipranges:
        src: 10.0.0.0/24
        dst: 10.0.0.0/24
      ports:
        src: [12222, 13333, 4441]
        dst: [53]
      proto: 17
      template: '{"yep": "{{.Srcip}}"}'
      weight: 100
```

## Author/Contact

Sascha Steinbiss

## License

MIT
