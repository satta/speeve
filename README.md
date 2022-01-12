# ⛲ speeve

[![Status](https://github.com/satta/speeve/actions/workflows/go.yml/badge.svg)](https://github.com/satta/speeve/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/satta/speeve.svg)](https://pkg.go.dev/github.com/satta/speeve)

speeve is a fast, probabilistic [EVE-JSON](https://suricata.readthedocs.io/en/latest/output/eve/eve-json-format.html) generator for testing and benchmarking of EVE-consuming applications. It can generate up to hundreds of thousands of events per second, with varying IPs, ports, timestamps, accurate [Community ID](https://github.com/corelight/community-id-spec) values, etc. under a predefined traffic profile model. Additional metadata can be added based on static strings, [Go templates](https://golang.org/pkg/text/template/) and [Tengo](https://github.com/d5/tengo) scripts. 

🚧 Still alpha - work in progress! 🚧

## Building

Build like other Go tools:

```
$ go get -v -t github.com/satta/speeve/...
...
$ speeve spew -h
```

## Usage

```
$ speeve spew -h
The 'spew' command starts EVE-JSON generation.

Usage:
  speeve spew [flags]

Flags:
  -d, --duration duration   duration of run
  -h, --help                help for spew
  -j, --parallel uint32     number of generator tasks to run in parallel (default 2)
  -s, --persec uint32       number of events/s to emit (default 1000)
      --pproffile string    filename to write pprof profiling info into
  -p, --profile string      filename of traffic profile definition file (default "profile.yaml")
      --seed int            random seed for sampling
  -n, --total uint          total number of events to emit
  -v, --verbose             verbose mode

Global Flags:
      --config string   config file (default is $HOME/.speeve.yaml)
```

The only required parameter is the name of a profile YAML file to use (see
section below for details). Other parameters allow for configuration of
parallelism (`-j`, might be needed for higher event rates) and event count (`-n`)
as well as rate (`-s`, events per second).

Example:
```
$ speeve spew -p profile.yaml | head -n 2
INFO[0000] seed 1615822102344012700
WARN[0000] random_dns: src ports undefined, will use random high ports 
WARN[0000] random_dns: src ports undefined, will use random high ports 
WARN[0000] dns: src ports undefined, will use random high ports 
WARN[0000] dns: src ports undefined, will use random high ports 
{"timestamp":"2021-03-10T16:02:51.321482+0100", "event_type":"flow", "src_ip": "10.0.0.228", "src_port": 21220, "dst_ip": "10.0.0.157", "dst_port": 53, "proto": "UDP", "flow_id": 8687071865980767663, "community_id": "1:FtaON10zfcu6oNKiXPDVk7bZNAQ=", "flow": {"pkts_toserver": 197, "pkts_toclient": 781, "bytes_toserver": 7448, "bytes_toclient": 9615, "start": "2021-03-10T16:02:50.32156+0100", "end": "2021-03-10T16:02:51.321482+0100", "age": 99, "state": "new", "reason": "shutdown", "alerted": false}}
{"timestamp":"2021-03-10T16:02:51.321482+0100", "event_type":"dns", "src_ip": "10.0.0.228", "src_port": 21220, "dst_ip": "10.0.0.157", "dst_port": 53, "proto": "UDP", "flow_id": 8687071865980767663, "community_id": "1:FtaON10zfcu6oNKiXPDVk7bZNAQ=", "dns": {"version":2,"type":"answer","id":574,"flags":"8180","qr":true,"rd":true,"ra":true,"rrname":"github.com","rrtype":"AAAA","rcode":"NOERROR","authorities":[{"rrname":"github.com","rrtype":"SOA","ttl":5}]}}
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
    - name: random_dns
      type: tengo
      event_type: dns
      ipranges:
        src: 10.0.0.0/8
        dst: 10.0.0.0/8
      ports:
        src: []
        dst: [53]
      proto: 17
      tengo: >
        rnd := import("rand");
        fmt := import("fmt");
        text := import("text");
        chars := "abcdefghijklmnopqrstuvw";

        a := [];
        for i:=0; i<10; i++ {
           a = append(a, chars[rnd.intn(23)]);
        }

        dom := text.join(append(a, ".com"), "");
        ip := fmt.sprintf("%d.%d.%d.%d", rnd.intn(250), rnd.intn(250), rnd.intn(250), rnd.intn(250));
        tmpl := `{"version":2,"type":"answer","id":%d,"flags":"8180","qr":true,"rd":true,"ra":true,"rrname":"%s","rrtype":"A","rcode":"NOERROR","answers":[{"rrname":"%s","rrtype":"A","ttl":299,"rdata":"%s"}],"grouped":{"A":["%s"]}}`;
  
        encoded := fmt.sprintf(tmpl, rnd.intn(40000), dom, dom, ip, ip);
      weight: 50
```

## Author/Contact

Sascha Steinbiss

## License

MIT
