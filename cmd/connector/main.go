package main

import (
	sdk "github.com/conduitio/conduit-connector-sdk"

	azure_event_hub "github.com/mer-oscar/conduit-connector-azure-event-hub"
)

func main() {
	sdk.Serve(azure_event_hub.Connector)
}
