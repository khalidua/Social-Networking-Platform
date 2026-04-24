package security

import (
	"testing"
	"time"
)

func TestStateManagerGenerateAndValidate(t *testing.T) {
	manager := NewStateManager("secret", defaultTestDuration())

	state, err := manager.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if err := manager.Validate(state); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestStateManagerRejectsTamperedState(t *testing.T) {
	manager := NewStateManager("secret", defaultTestDuration())

	state, err := manager.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if err := manager.Validate(state + "tampered"); err == nil {
		t.Fatal("Validate() expected tampered state error")
	}
}

func defaultTestDuration() time.Duration {
	return 10 * time.Minute
}
