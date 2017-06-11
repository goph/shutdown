package shutdown

import (
	"errors"
	"fmt"
)

// errorHandler handles an error.
type errorHandler interface {
	Handle(err error)
}

// noopErrorHandler is a default fallback in case there is no error handler configured.
type noopErrorHandler struct{}

func (e *noopErrorHandler) Handle(err error) {}

// r accepts a recovered panic (if any) and converts it to an error (if necessary).
func r(r interface{}) (err error) {
	if r != nil {
		switch x := r.(type) {
		case string:
			err = errors.New(x)
		case error:
			err = x
		default:
			err = fmt.Errorf("Unknown panic, received: %v", r)
		}
	}

	return err
}
