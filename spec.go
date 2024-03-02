package azure_event_hub

import (
	sdk "github.com/conduitio/conduit-connector-sdk"
)

// version is set during the build process with ldflags (see Makefile).
// Default version matches default from runtime/debug.
var version = "(devel)"

// Specification returns the connector's specification.
func Specification() sdk.Specification {
	return sdk.Specification{
		Name:        "azure-event-hub",
		Summary:     "Azure Event Hub Summary",
		Description: "Azure Event Hub Description",
		Version:     version,
		Author:      "Oscar Villavicencio",
	}
}
