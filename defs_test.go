package quickbooks

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertDateEqual asserts that actual parses to the same time as expected.
// expected is in the QB wire format (RFC3339 timestamp or bare YYYY-MM-DD,
// without surrounding JSON quotes). Bare dates are anchored at midnight UTC
// in both expected and actual, so they compare equal regardless of host TZ.
func assertDateEqual(t *testing.T, expected string, actual Date) {
	t.Helper()
	expectedDate := Date{RawMessage: []byte(fmt.Sprintf("%q", expected))}
	e, err := expectedDate.In(time.UTC)
	require.NoErrorf(t, err, "parsing expected %q", expected)
	a, err := actual.In(time.UTC)
	require.NoErrorf(t, err, "parsing actual %s", actual.RawMessage)
	assert.Truef(t, a.Equal(e), "expected %s, got %s",
		e.Format(time.RFC3339), a.Format(time.RFC3339))
}
