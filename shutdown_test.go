package shutdown

import (
	"testing"

	"fmt"
)

func CreateShutdownFunc() (Handler, *bool) {
	var called bool

	f := Func(func() {
		called = true
	})

	return f, &called
}

func CreateOrderedShutdownFuncs(num int) ([]Handler, *[]int) {
	funcs := make([]Handler, num)
	called := []int{}

	for index := 0; index < num; index++ {
		funcs[index] = func(index int) Handler {
			return Func(func() {
				called = append(called, index+1)
			})
		}(index)
	}

	return funcs, &called
}

func TestFunc_CallsUnderlyingFunc(t *testing.T) {
	f, called := CreateShutdownFunc()

	var err error

	if got, want := f(), err; got != want {
		t.Fatalf("wrapped functions are expected to return nil, error received: %v", got)
	}

	if *called != true {
		t.Fatal("the wrapped function is expected to be called")
	}
}

func TestFunc_RecoversErrorPanic(t *testing.T) {
	err := fmt.Errorf("internal error")

	f := Func(func() {
		panic(err)
	})

	if got, want := f(), err; got != want {
		t.Fatalf("expected to recover a specific error, received: %v", got)
	}
}

func TestFunc_RecoversStringPanic(t *testing.T) {
	f := Func(func() {
		panic("internal error")
	})

	if got, want := f().Error(), "internal error"; got != want {
		t.Fatalf("expected to recover a specific error, received: %v", got)
	}
}

func TestFunc_RecoversAnyPanic(t *testing.T) {
	f := Func(func() {
		panic(123)
	})

	if got, want := f().Error(), "Unknown panic, received: 123"; got != want {
		t.Fatalf("expected to recover a specific error, received: %v", got)
	}
}

func TestNewManager(t *testing.T) {
	manager := NewManager()

	// Test falling back to NullHandler
	manager.Register(func() error {
		return fmt.Errorf("error")
	})

	manager.Shutdown()
}

func TestManager_Register(t *testing.T) {
	f, called := CreateShutdownFunc()

	manager := NewManager()

	manager.Register(f)
	manager.Shutdown()

	if *called != true {
		t.Fatal("the shutdown handler is expected to be called")
	}
}

func TestManager_Register_ExecutedInOrder(t *testing.T) {
	funcs, called := CreateOrderedShutdownFuncs(2)

	manager := NewManager()

	manager.Register(funcs[0], funcs[1])
	manager.Shutdown()

	if got, want := (*called)[0], 1; got != want {
		t.Fatal("the first shutdown handler is expected to be called first")
	}

	if got, want := (*called)[1], 2; got != want {
		t.Fatal("the second shutdown handler is expected to be called second")
	}
}

func TestManager_RegisterAsFirst(t *testing.T) {
	funcs, called := CreateOrderedShutdownFuncs(2)

	manager := NewManager()

	manager.Register(funcs[1])
	manager.RegisterAsFirst(funcs[0])
	manager.Shutdown()

	if got, want := (*called)[0], 1; got != want {
		t.Fatal("the first shutdown handler is expected to be called first")
	}

	if got, want := (*called)[1], 2; got != want {
		t.Fatal("the second shutdown handler is expected to be called second")
	}
}

func TestManager_Shutdown(t *testing.T) {
	errorHandler := &testErrorHandler{}
	manager := NewManager()
	manager.ErrorHandler = errorHandler

	manager.Shutdown()

	if errorHandler.Last() != nil {
		t.Fatal("shutting down not emit an error")
	}
}

func TestManager_Shutdown_HandleErrors(t *testing.T) {
	errorHandler := &testErrorHandler{}
	manager := NewManager()
	manager.ErrorHandler = errorHandler

	err := fmt.Errorf("error")

	manager.Register(func() error {
		return err
	})

	manager.Shutdown()

	if errorHandler.Last() != err {
		t.Fatal("errors ocurred during shutdown should be handled")
	}
}

func TestManager_Shutdown_RecoverFromPanic(t *testing.T) {
	errorHandler := &testErrorHandler{}
	manager := NewManager()
	manager.ErrorHandler = errorHandler

	err := fmt.Errorf("error")

	func() {
		defer manager.Shutdown()

		func() {
			panic(err)
		}()
	}()

	if errorHandler.Last() != err {
		t.Fatal("errors ocurred during shutdown should be handled")
	}
}
