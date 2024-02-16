package azure-event-hub_test

import (
	"context"
	"testing"

	azure-event-hub "github.com/mer-oscar/conduit-connector-azure-event-hub"
	"github.com/matryer/is"
)

func TestTeardownSource_NoOpen(t *testing.T) {
	is := is.New(t)
	con := azure-event-hub.NewSource()
	err := con.Teardown(context.Background())
	is.NoErr(err)
}
