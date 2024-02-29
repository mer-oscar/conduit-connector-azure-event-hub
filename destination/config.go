package destination

import "github.com/mer-oscar/conduit-connector-azure-event-hub/common"

//go:generate paramgen -output=config_paramgen.go Config

type Config struct {
	common.Config

	// BatchSamePartition is a boolean that controls whether to write batches of records to the same partition
	// or to divide them among available partitions according to the azure event hubs producer client
	BatchSamePartition bool `json:"batchSamePartition" default:"false"`
}
