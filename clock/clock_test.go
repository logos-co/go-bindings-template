package clock

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Test basic creation, cleanup, and reset
func TestLifecycle(t *testing.T) {
	clock, err := NewClock()
	require.NoError(t, err)
	require.NotNil(t, clock, "Expected Clock to be not nil")

	err = clock.Destroy()
	require.NoError(t, err)
}
