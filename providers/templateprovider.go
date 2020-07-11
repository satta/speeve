package providers

import (
	"bytes"
	"io/ioutil"
	"strconv"
	"text/template"

	"github.com/satta/speeve/flow"
	yaml "gopkg.in/yaml.v2"
)

type TemplateProviderConfig struct {
	Providers []struct {
		Name     string
		Template string
	} `yaml:"providers"`
}

type TemplateProvider struct {
	tpl *template.Template
}

func MakeTemplateProvider(configFile string, name string) (*TemplateProvider, error) {
	tp := &TemplateProvider{}
	t := TemplateProviderConfig{}

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
			tmpl, err := template.New(p.Name).Parse(p.Template)
			if err != nil {
				return nil, err
			}
			tp.tpl = tmpl
			return tp, nil
		}

	}
	return tp, nil
}

type TemplateProviderFields struct {
	Timestamp   string
	Srcip       string
	Dstip       string
	Srcport     string
	Dstport     string
	Communityid string
}

func (t *TemplateProvider) GetByte(f *flow.Flow) []byte {
	var buf bytes.Buffer
	tf := TemplateProviderFields{
		Timestamp:   f.Timestamp,
		Srcip:       f.SrcIP.String(),
		Dstip:       f.DstIP.String(),
		Srcport:     strconv.Itoa(int(f.SrcPort)),
		Dstport:     strconv.Itoa(int(f.DstPort)),
		Communityid: string(f.CommunityID),
	}
	t.tpl.Execute(&buf, tf)
	return buf.Bytes()
}
