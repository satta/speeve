package providers

import (
	"io/ioutil"

	"github.com/satta/speeve/flow"
	yaml "gopkg.in/yaml.v2"
)

type StaticProviderConfig struct {
	Providers []struct {
		Name   string
		Static string
	} `yaml:"providers"`
}

type StaticProvider struct {
	Static string
}

func MakeStaticProvider(configFile string, name string) (*StaticProvider, error) {
	tp := &StaticProvider{}
	t := StaticProviderConfig{}

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
			tp.Static = p.Static
			return tp, nil
		}

	}
	return tp, nil
}

type StaticProviderFields struct {
	Timestamp   string
	Srcip       string
	Dstip       string
	Srcport     string
	Dstport     string
	Communityid string
}

func (t *StaticProvider) GetByte(f *flow.Flow) []byte {
	return []byte(t.Static)
}
