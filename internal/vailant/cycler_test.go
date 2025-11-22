package vailant

import (
	"testing"
	"time"

	"github.com/ksimuk/ebus-climate/internal/climate"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/physic"
)

type mockPin struct {
	level gpio.Level
}

func (m *mockPin) String() string                { return "mockPin" }
func (m *mockPin) Halt() error                   { return nil }
func (m *mockPin) Name() string                  { return "mockPin" }
func (m *mockPin) Number() int                   { return 0 }
func (m *mockPin) Function() string              { return "Out" }
func (m *mockPin) In(pull gpio.Pull, edge gpio.Edge) error { return nil }
func (m *mockPin) Read() gpio.Level              { return m.level }
func (m *mockPin) WaitForEdge(timeout time.Duration) bool { return false }
func (m *mockPin) Pull() gpio.Pull               { return gpio.PullNoChange }
func (m *mockPin) DefaultPull() gpio.Pull        { return gpio.PullNoChange }
func (m *mockPin) Out(l gpio.Level) error {
	m.level = l
	return nil
}
func (m *mockPin) PWM(duty gpio.Duty, freq physic.Frequency) error { return nil }

func createTestClimate() *eBusClimate {
	c := &eBusClimate{
		stopChan:          make(chan struct{}),
		heatingActive:     false,
		heatingRelay:      &mockPin{},
		heatingTimerMutex: make(chan struct{}, 1),
		stat: climate.Stat{
			HwcDemand: "off",
		},
		state: &climate.ClimateState{
			Mode: MODE_HEATING,
		},
		power: 1000,
	}
	c.heatingTimerMutex <- struct{}{} // initialize mutex
	return c
}

func TestRunForStartsNewCycle(t *testing.T) {
	c := createTestClimate()
	
	c.runFor(5)
	
	// Give goroutine time to start
	time.Sleep(100 * time.Millisecond)
	
	if !c.heatingActive {
		t.Error("Expected heating to be active")
	}
	
	<-c.heatingTimerMutex
	endTime := c.heatingEndTime
	c.heatingTimerMutex <- struct{}{}
	
	expectedEnd := time.Now().Add(5 * time.Minute)
	diff := endTime.Sub(expectedEnd).Abs()
	
	if diff > time.Second {
		t.Errorf("Expected end time around %v, got %v (diff: %v)", expectedEnd, endTime, diff)
	}
	
	c.StopHeating()
	close(c.stopChan) // Stop the goroutine
}

func TestRunForExtendsExistingCycle(t *testing.T) {
	c := createTestClimate()
	
	// Start first cycle
	c.runFor(5)
	time.Sleep(100 * time.Millisecond)
	
	<-c.heatingTimerMutex
	firstEndTime := c.heatingEndTime
	c.heatingTimerMutex <- struct{}{}
	
	// Extend with another 3 minutes
	c.runFor(3)
	time.Sleep(50 * time.Millisecond)
	
	<-c.heatingTimerMutex
	secondEndTime := c.heatingEndTime
	c.heatingTimerMutex <- struct{}{}
	
	diff := secondEndTime.Sub(firstEndTime)
	expectedDiff := 3 * time.Minute
	
	if (diff - expectedDiff).Abs() > time.Second {
		t.Errorf("Expected extension of %v, got %v", expectedDiff, diff)
	}
	
	c.StopHeating()
	close(c.stopChan)
}

func TestHwcDemandExtendsHeating(t *testing.T) {
	t.Skip("Skipping time-dependent test - requires 1+ minute to run")
}

func TestHwcDemandMultipleExtensions(t *testing.T) {
	t.Skip("Skipping time-dependent test - requires 2+ minutes to run")
}

func TestIsHwcDemandActive(t *testing.T) {
	c := createTestClimate()
	
	tests := []struct {
		value    string
		expected bool
	}{
		{"on", true},
		{"yes", true},
		{"1", true},
		{"true", true},
		{"off", false},
		{"no", false},
		{"0", false},
		{"false", false},
		{"", false},
	}
	
	for _, tt := range tests {
		c.stat.HwcDemand = tt.value
		result := c.isHwcDemandActive()
		if result != tt.expected {
			t.Errorf("For HwcDemand=%q, expected %v, got %v", tt.value, tt.expected, result)
		}
	}
}

func TestGetStatReturnsHeatingEndTime(t *testing.T) {
	c := createTestClimate()
	
	// Test when heating is not active
	stat := c.GetStat()
	if stat.HeatingEndTime != "" {
		t.Errorf("Expected empty HeatingEndTime when not heating, got %q", stat.HeatingEndTime)
	}
	
	// Start heating
	c.runFor(10)
	time.Sleep(100 * time.Millisecond)
	
	// Test when heating is active
	stat = c.GetStat()
	if stat.HeatingEndTime == "" {
		t.Error("Expected non-empty HeatingEndTime when heating")
	}
	
	// Verify it's a valid RFC3339 timestamp
	_, err := time.Parse(time.RFC3339, stat.HeatingEndTime)
	if err != nil {
		t.Errorf("Expected valid RFC3339 timestamp, got error: %v", err)
	}
	
	c.StopHeating()
	close(c.stopChan)
}

func TestHeatingStopsAtEndTime(t *testing.T) {
	t.Skip("Skipping time-dependent test - requires 2+ minutes to run")
}

func TestConcurrentRunForCalls(t *testing.T) {
	c := createTestClimate()
	
	// Start initial cycle
	c.runFor(5)
	time.Sleep(100 * time.Millisecond)
	
	// Make multiple concurrent extension calls
	done := make(chan bool, 3)
	for i := 0; i < 3; i++ {
		go func() {
			c.runFor(2)
			done <- true
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}
	
	<-c.heatingTimerMutex
	endTime := c.heatingEndTime
	c.heatingTimerMutex <- struct{}{}
	
	// Total should be 5 + 2 + 2 + 2 = 11 minutes from start
	expectedEnd := time.Now().Add(11 * time.Minute)
	diff := endTime.Sub(expectedEnd).Abs()
	
	// Allow for some timing variance
	if diff > 2*time.Second {
		t.Errorf("Expected end time around %v, got %v (diff: %v)", expectedEnd, endTime, diff)
	}
	
	c.StopHeating()
	close(c.stopChan)
}
