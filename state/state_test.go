package state_test

import (
	"testing"

	"github.com/vickychhetri/nova/state"
)

func TestStateValue(t *testing.T) {
	count := state.Int(10)
	if count.Get() != 10 {
		t.Fatalf("expected 10, got %d", count.Get())
	}

	var observed int
	unsub := count.Subscribe(func(v int) {
		observed = v
	})

	count.Set(25)
	if observed != 25 {
		t.Fatalf("expected 25, got %d", observed)
	}

	count.Update(func(c int) int { return c + 5 })
	if observed != 30 || count.Get() != 30 {
		t.Fatalf("expected 30, got %d (observed %d)", count.Get(), observed)
	}

	unsub()
	count.Set(100)
	if observed != 30 {
		t.Fatalf("expected observed to stay 30 after unsubscribe, got %d", observed)
	}
}

func TestStateComputed(t *testing.T) {
	firstName := state.String("John")
	lastName := state.String("Doe")

	fullName := state.Compute(func() string {
		return firstName.Get() + " " + lastName.Get()
	})

	if fullName.Get() != "John Doe" {
		t.Fatalf("expected 'John Doe', got '%s'", fullName.Get())
	}

	var updates []string
	unsub := fullName.Subscribe(func(val string) {
		updates = append(updates, val)
	})
	defer unsub()

	firstName.Set("Jane")
	if fullName.Get() != "Jane Doe" {
		t.Fatalf("expected 'Jane Doe', got '%s'", fullName.Get())
	}

	lastName.Set("Smith")
	if fullName.Get() != "Jane Smith" {
		t.Fatalf("expected 'Jane Smith', got '%s'", fullName.Get())
	}

	if len(updates) != 2 || updates[0] != "Jane Doe" || updates[1] != "Jane Smith" {
		t.Fatalf("unexpected update history: %+v", updates)
	}
}

func TestStateEffectAndBatch(t *testing.T) {
	a := state.Int(1)
	b := state.Int(2)
	var sum int
	var runs int

	dispose := state.Effect(func() func() {
		runs++
		sum = a.Get() + b.Get()
		return nil
	})
	defer dispose()

	if sum != 3 || runs != 1 {
		t.Fatalf("initial effect failed: sum=%d, runs=%d", sum, runs)
	}

	state.Batch(func() {
		a.Set(10)
		b.Set(20)
	})

	if sum != 30 {
		t.Fatalf("expected sum 30 after batch, got %d", sum)
	}
}
