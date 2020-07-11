package providers

import (
	"fmt"

	"github.com/satta/speeve/flow"
)

type EVEProvider interface {
	GetByte(*flow.Flow) []byte
}

func CreateProvider(ptype string, configFile string, name string) (EVEProvider, error) {
	switch ptype {
	case "template":
		return MakeTemplateProvider(configFile, name)
	case "tengo":
		return MakeTengoProvider(configFile, name)
	case "static":
		return MakeStaticProvider(configFile, name)
	default:
		return nil, fmt.Errorf("unknown provider type: %s", ptype)
	}
}
