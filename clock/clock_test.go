package clock

import (
	"testing"
	"time"

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

func TestAlarmEvent(t *testing.T) {
	clock, err := NewClock()
	require.NoError(t, err)
	require.NotNil(t, clock, "Expected Clock to be not nil")

	defer clock.Destroy()

	// Use a channel for signaling
	readyChan := make(chan string, 1)

	callbacks := EventCallbacks{
		OnAlarm: func(time string, msg string) {
			// Non-blocking send to channel
			select {
			case readyChan <- msg:
			default:
				// Avoid blocking if channel is full or test already timed out
			}
		},
	}

	alarmMsg := "this is my alarm"
	err = clock.SetAlarm(1000, alarmMsg)
	require.NoError(t, err)

	// Register callback only on the receiver
	clock.RegisterCallbacks(callbacks)

	// Verification - Wait on channel with timeout
	select {
	case receivedMsg := <-readyChan:
		// Mark as called implicitly since we received on channel
		if receivedMsg != alarmMsg {
			t.Errorf("OnAlarm called with wrong alarm message: got %q, want %q", receivedMsg, alarmMsg)
		}
	case <-time.After(2 * time.Second):
		// If timeout occurs, the channel receive failed.
		t.Errorf("Timed out waiting for OnAlarm callback on readyChan")
	}

}
