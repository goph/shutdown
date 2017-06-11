// Package shutdown provides tools for handling application shutdowns.
package shutdown

import "github.com/goph/stdlib/errors"

// Manager manages an application shutdown by calling the registered handlers.
type Manager struct {
	handlers     []Handler
	ErrorHandler errorHandler
}

// Handler is any function that has no parameters and can return an error.
//
// Shutdown handlers are the last resort for the application when there is no more
// flow control in the user's hand.
//
// Returned errors are logged.
type Handler func() error

// Func wraps a function withot an error return type.
//
// To make sure there are no silenced errors, panics are also recovered.
func Func(fn func()) Handler {
	return func() (err error) {
		defer func() {
			err = r(recover())
		}()

		fn()

		return err
	}
}

// NewManager creates a new Shutdown manager.
func NewManager() *Manager {
	return &Manager{
		ErrorHandler: &noopErrorHandler{},
	}
}

// Register appends new shutdown handlers to the list of existing ones.
func (m *Manager) Register(handlers ...Handler) {
	m.handlers = append(m.handlers, handlers...)
}

// RegisterAsFirst prepends new shutdown handlers to the list of existing ones.
func (m *Manager) RegisterAsFirst(handlers ...Handler) {
	m.handlers = append(handlers, m.handlers...)
}

// Shutdown is the panic recovery and shutdown handler.
//
// It should be called as the last method in `main` (eg. using defer).
func (m *Manager) Shutdown() {
	// Try recovering from panic first
	err := errors.Recover(recover())
	if err != nil {
		m.ErrorHandler.Handle(err)
	}

	// Loop through all the handlers and call them
	// Log any errors that may occur
	for _, handler := range m.handlers {
		err := handler()
		if err != nil {
			m.ErrorHandler.Handle(err)
		}
	}
}
