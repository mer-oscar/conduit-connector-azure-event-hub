package destination

import "github.com/mer-oscar/conduit-connector-azure-event-hub/common"

//go:generate paramgen -output=config_paramgen.go Config

type Config struct {
	common.Config
}
