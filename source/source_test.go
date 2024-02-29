package source

import (
	"context"
	"testing"

	"github.com/matryer/is"
)

func TestTeardownSource_NoOpen(t *testing.T) {
	is := is.New(t)
	con := New()
	err := con.Teardown(context.Background())
	is.NoErr(err)
}
