package generator

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/satta/speeve/flow"
	"github.com/satta/speeve/providers"
	"github.com/satta/speeve/util"

	"github.com/satta/gommunityid"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

type ProviderConfig struct {
	Providers []struct {
		Name      string
		Type      string
		EventType string `yaml:"event_type"`
		Ports     struct {
			Src []uint16
			Dst []uint16
		}
		IPRanges struct {
			Src string
			Dst string
		}
		Proto          uint8
		ProtoString    string
		Weight         uint
		TopLevelFields map[string]string `yaml:"top_level_fields"`
	}
}

type ConfiguredProvider struct {
	SrcIPs        *util.IPSampler
	DstIPs        *util.IPSampler
	SrcPorts      []uint16
	DstPorts      []uint16
	Proto         uint8
	Sampler       *util.Sampler
	Provider      providers.EVEProvider
	EventType     string
	ExtraTopLevel string
}

type FlowGenerator struct {
	Providers   []ConfiguredProvider
	Sampler     *util.Sampler
	CommunityID gommunityid.CommunityID
	Buffer      *bytes.Buffer
}

func MakeFlowGenerator(configFile string) (*FlowGenerator, error) {
	pc := ProviderConfig{}
	fg := FlowGenerator{
		Providers: make([]ConfiguredProvider, 0),
		Sampler: &util.Sampler{
			Options: make([]uint, 0),
		},
		Buffer: new(bytes.Buffer),
	}
	cid, err := gommunityid.GetCommunityIDByVersion(1, 0)
	if err != nil {
		return nil, err
	}
	fg.CommunityID = cid
	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal([]byte(configData), &pc)
	if err != nil {
		return nil, err
	}
	for idx, pc := range pc.Providers {
		p, err := providers.CreateProvider(pc.Type, configFile, pc.Name)
		if err != nil {
			return nil, err
		}
		ss, err := util.MakeIPSampler(pc.IPRanges.Src)
		if err != nil {
			return nil, err
		}
		ds, err := util.MakeIPSampler(pc.IPRanges.Dst)
		if err != nil {
			return nil, err
		}
		if len(pc.Ports.Src) == 0 {
			log.Warnf("%s: src ports undefined, will use random high ports", pc.Name)
		}
		if len(pc.Ports.Dst) == 0 {
			log.Warnf("%s: dst ports undefined, will use random high ports", pc.Name)
		}
		if pc.Weight == 0 {
			return nil, fmt.Errorf("%s: weight cannot be undefined or 0", pc.Name)
		}
		if pc.Proto == 0 {
			return nil, fmt.Errorf("%s: protocol number cannot be undefined or 0", pc.Name)
		}
		if len(pc.EventType) == 0 {
			return nil, fmt.Errorf("%s: event type be cannot be undefined or empty", pc.Name)
		}
		// preprocess extra top level fields
		topLevelOut := ""
		if len(pc.TopLevelFields) > 0 {
			for k, v := range pc.TopLevelFields {
				topLevelOut += fmt.Sprintf(`, "%s": "%s"`, k, v)
			}
		}
		cfgp := ConfiguredProvider{
			Provider:      p,
			SrcIPs:        ss,
			DstIPs:        ds,
			SrcPorts:      pc.Ports.Src,
			DstPorts:      pc.Ports.Dst,
			Proto:         pc.Proto,
			EventType:     pc.EventType,
			ExtraTopLevel: topLevelOut,
		}
		fg.Providers = append(fg.Providers, cfgp)
		fg.Sampler.Add(uint(idx), pc.Weight)
	}
	return &fg, nil
}

const suricataTimestampFormat = "2006-01-02T15:04:05.999999-0700"

func (fg *FlowGenerator) makeFlowForProvider(p ConfiguredProvider) *flow.Flow {
	srcPort := uint16(0)
	dstPort := uint16(0)
	if len(p.SrcPorts) == 0 {
		srcPort = uint16(rand.Intn(63000) + 1024)
	} else {
		srcPort = p.SrcPorts[rand.Intn(len(p.SrcPorts))]
	}
	if len(p.DstPorts) == 0 {
		dstPort = uint16(rand.Intn(63000) + 1024)
	} else {
		dstPort = p.DstPorts[rand.Intn(len(p.DstPorts))]
	}
	f := &flow.Flow{
		SrcIP:     p.SrcIPs.GetIP(),
		DstIP:     p.DstIPs.GetIP(),
		SrcPort:   srcPort,
		DstPort:   dstPort,
		Proto:     p.Proto,
		Timestamp: time.Now().Format(suricataTimestampFormat),
	}
	ft := gommunityid.MakeFlowTuple(f.SrcIP, f.DstIP, f.SrcPort, f.DstPort, f.Proto)
	f.CommunityID = fg.CommunityID.CalcBase64(ft)
	return f
}

func (fg *FlowGenerator) EmitFlow(out chan<- []byte) {
	selectedProvider := fg.Providers[fg.Sampler.Sample()]
	flow := fg.makeFlowForProvider(selectedProvider)

	protoStr := "UNKNOWN"
	switch flow.Proto {
	case 6:
		protoStr = "TCP"
	case 17:
		protoStr = "UDP"
	}

	flowID := rand.Uint64() / 2

	flowStart := fmt.Sprintf(`{"timestamp":"%s", "event_type":"flow", "src_ip": "%s", "src_port": %d, "dst_ip": "%s", "dst_port": %d, "proto": "%s", "flow_id": %d, "community_id": "%s"`,
		flow.Timestamp,
		flow.SrcIP.String(),
		flow.SrcPort,
		flow.DstIP.String(),
		flow.DstPort,
		protoStr,
		flowID,
		flow.CommunityID)
	alerted := "false"
	if selectedProvider.EventType == "alert" {
		alerted = "true"
	}
	state := []string{"new", "established", "closed"}[rand.Intn(3)]
	reason := []string{"timeout", "forced", "shutdown"}[rand.Intn(3)]
	flowJSON := fmt.Sprintf(`{"pkts_toserver": %d, "pkts_toclient": %d, "bytes_toserver": %d, "bytes_toclient": %d, "start": "%s", "end": "%s", "age": %d, "state": "%s", "reason": "%s", "alerted": %s}`,
		rand.Intn(1000),
		rand.Intn(1000),
		rand.Intn(10000),
		rand.Intn(10000),
		time.Now().Add(-1*time.Second).Format(suricataTimestampFormat),
		flow.Timestamp,
		rand.Intn(100),
		state,
		reason,
		alerted)
	fg.Buffer.WriteString(flowStart)
	fg.Buffer.WriteString(selectedProvider.ExtraTopLevel)
	fg.Buffer.WriteString(fmt.Sprintf(`, "flow": %s}`, flowJSON))
	fg.Buffer.WriteByte('\n')
	// We intentionally use Buffer.String() here to ensure we pass a copy
	// of the buffer content
	out <- []byte(fg.Buffer.String())
	fg.Buffer.Reset()

	genericStart := fmt.Sprintf(`{"timestamp":"%s", "event_type":"%s", "src_ip": "%s", "src_port": %d, "dst_ip": "%s", "dst_port": %d, "proto": "%s", "flow_id": %d, "community_id": "%s"`,
		flow.Timestamp,
		selectedProvider.EventType,
		flow.SrcIP.String(),
		flow.SrcPort,
		flow.DstIP.String(),
		flow.DstPort,
		protoStr,
		flowID,
		flow.CommunityID)
	providerJSON := string(selectedProvider.Provider.GetByte(flow))
	fg.Buffer.WriteString(genericStart)
	fg.Buffer.WriteString(selectedProvider.ExtraTopLevel)
	fg.Buffer.WriteString(fmt.Sprintf(`, "%s": %s}`, selectedProvider.EventType, providerJSON))
	fg.Buffer.WriteByte('\n')
	// We intentionally use Buffer.String() here to ensure we pass a copy
	// of the buffer content
	out <- []byte(fg.Buffer.String())
	fg.Buffer.Reset()
}
