package providers

import (
	"bytes"
	"io/ioutil"
	"strconv"

	"github.com/d5/tengo"
	"github.com/d5/tengo/stdlib"
	"github.com/satta/speeve/flow"
	yaml "gopkg.in/yaml.v2"
)

type TengoProviderConfig struct {
	Providers []struct {
		Name  string
		Tengo string
	} `yaml:"providers"`
}

type TengoProvider struct {
	script   *tengo.Script
	compiled *tengo.Compiled
}

func MakeTengoProvider(configFile string, name string) (*TengoProvider, error) {
	tp := &TengoProvider{}
	t := TengoProviderConfig{}

	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal([]byte(configData), &t)
	if err != nil {
		return nil, err
	}
	for _, p := range t.Providers {
		if p.Name == name {
			s := tengo.NewScript([]byte(p.Tengo))
			s.SetImports(stdlib.GetModuleMap(stdlib.AllModuleNames()...))
			tp.script = s
			c, err := s.Compile()
			if err != nil {
				return nil, err
			}
			tp.compiled = c
			if err != nil {
				return nil, err
			}
			return tp, nil
		}

	}
	return tp, nil
}

type TengoProviderFields struct {
	Timestamp   string
	Srcip       string
	Dstip       string
	Srcport     string
	Dstport     string
	Communityid string
}

func (t *TengoProvider) GetByte(f *flow.Flow) []byte {
	var buf bytes.Buffer
	tf := TengoProviderFields{
		Timestamp:   "foo", // TODO
		Srcip:       f.SrcIP.String(),
		Dstip:       f.DstIP.String(),
		Srcport:     strconv.Itoa(int(f.SrcPort)),
		Dstport:     strconv.Itoa(int(f.DstPort)),
		Communityid: string(f.CommunityID),
	}
	_ = tf
	if err := t.compiled.Run(); err == nil {
		buf.WriteString(t.compiled.Get("encoded").String())
	}
	return buf.Bytes()
}
