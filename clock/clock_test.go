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
	alarmChan := make(chan alarmEvent, 1)

	callbacks := EventCallbacks{
		OnAlarm: func(event alarmEvent) {
			// Non-blocking send to channel
			select {
			case alarmChan <- event:
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
	case receivedAlarm := <-alarmChan:
		epochSeconds := time.Now().Unix()

		// Mark as called implicitly since we received on channel
		if receivedAlarm.Msg != alarmMsg {
			t.Errorf("OnAlarm called with wrong alarm message: got %q, want %q", receivedAlarm.Msg, alarmMsg)
		}

		if epochSeconds-receivedAlarm.Time > 1 {
			t.Errorf("Alarm was set at %d but current time is %d", receivedAlarm.Time, epochSeconds)
		}

	case <-time.After(2 * time.Second):
		// If timeout occurs, the channel receive failed.
		t.Errorf("Timed out waiting for OnAlarm callback on alarmChan")
	}

}
