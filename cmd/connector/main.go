package main

import (
	sdk "github.com/conduitio/conduit-connector-sdk"

	azure-event-hub "github.com/mer-oscar/conduit-connector-azure-event-hub"
)

func main() {
	sdk.Serve(azure-event-hub.Connector)
}
