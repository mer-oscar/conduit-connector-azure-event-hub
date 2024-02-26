package azure_event_hub

import (
	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/mer-oscar/conduit-connector-azure-event-hub/destination"
	"github.com/mer-oscar/conduit-connector-azure-event-hub/source"
)

// Connector combines all constructors for each plugin in one struct.
var Connector = sdk.Connector{
	NewSpecification: Specification,
	NewSource:        source.NewSource,
	NewDestination:   destination.NewDestination,
}
