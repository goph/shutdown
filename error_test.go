package shutdown

import (
	"testing"

	"fmt"
)

// testErrorHandler is a test implementation of errorHandler
type testErrorHandler struct {
	errors []error
}

// Handle takes care of an error by logging it.
func (h *testErrorHandler) Handle(err error) {
	h.errors = append(h.errors, err)
}

// Last returns the last error (if any).
func (h *testErrorHandler) Last() error {
	count := len(h.errors)

	// Return the last error (if any)
	if count > 0 {
		return h.errors[count-1]
	}

	return nil
}

func createRecoverFunc(p interface{}) func() error {
	return func() (err error) {
		defer func() {
			err = r(recover())
		}()

		panic(p)
	}
}

func TestRecover_ErrorPanic(t *testing.T) {
	err := fmt.Errorf("internal error")

	f := createRecoverFunc(err)

	if got, want := f(), err; got != want {
		t.Fatalf("expected to recover a specific error, received: %v", got)
	}
}

func TestRecover_StringPanic(t *testing.T) {
	f := createRecoverFunc("internal error")

	if got, want := f().Error(), "internal error"; got != want {
		t.Fatalf("expected to recover a specific error, received: %v", got)
	}
}

func TestRecover_AnyPanic(t *testing.T) {
	f := createRecoverFunc(123)

	if got, want := f().Error(), "Unknown panic, received: 123"; got != want {
		t.Fatalf("expected to recover a specific error, received: %v", got)
	}
}
