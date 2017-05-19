package raygun

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/kaeuferportal/stack2struct"
	pkgerr "github.com/pkg/errors"
)

// Cause returns the underlying cause of the error, if possible.
// An error value has a cause if it implements the following
// interface:
//
//     type causer interface {
//            Cause() error
//     }
//
// If the error does not implement Cause, the original error will
// be returned. If the error is nil, nil will be returned without further
// investigation.
func cause(err error) error {
	type causer interface {
		Cause() error
	}

	for {
		cause, ok := err.(causer)
		if !ok {
			break
		}
		c := cause.Cause()
		if c == nil {
			break
		}
		err = c
	}
	return err
}

// Class returns the class that caused the error, if possible.
// An error value has a class if it implements the following
// interface:
//
//     type classer interface {
//            Class() string
//     }
//
// If the error does not implement Class, it returns an empty string
func class(err error) string {
	type classer interface {
		Class() string
	}

	if e, ok := err.(classer); ok {
		return e.Class()
	}

	return ""
}

// data returns additional data about the error, if possible.
// An error value has additional data if it implements the following
// interface:
//
//     type dataer interface {
//            Data() string
//     }
//
// If the error does not implement data, it returns nil
func data(err error) interface{} {
	type dataer interface {
		data() interface{}
	}

	if e, ok := err.(dataer); ok {
		return e.data()
	}

	return nil
}

// data returns the stacktrace of the error.
// It can rely on the errors internal interfaces, or create a new one from the current stacktrace
// It accepts different interfaces:
//
//	 type stackTracer1 interface {
//	 		StackTrace() pkgerr.StackTrace
//	 }
//
//	 type stackTracer2 interface {
//	 		StackTrace() []string
//	 }
//
// If the error does not implement data, it returns nil
func stacktrace(err error) StackTrace {
	type stackTracer1 interface {
		StackTrace() pkgerr.StackTrace
	}

	type stackTracer2 interface {
		StackTrace() []string
	}

	stack := StackTrace{}

	// Read pkg/errors stacktrace
	if e, ok := err.(stackTracer1); ok {
		s := e.StackTrace()

		for _, line := range s {
			n, err := strconv.Atoi(fmt.Sprintf("%d", line))
			if err != nil {
				n = -1
			}

			filename := fmt.Sprintf("%+s", line)
			parts := strings.Split(filename, "\n\t")
			packparts := strings.Split(parts[0], ".")
			pack := strings.Join(packparts[:len(packparts)-1], ".")

			stack.AddEntry(n, pack, parts[1], fmt.Sprintf("%n", line))
		}

		return stack
	}

	// Read juju/errors stacktrace
	if e, ok := err.(stackTracer2); ok {
		s := e.StackTrace()

		for _, line := range s {
			parts := strings.Split(line, ":")

			n, err := strconv.Atoi(parts[1])
			if err != nil {
				n = -1
			}

			stack.AddEntry(n, "", parts[0], "")
		}

		return stack
	}

	rawStackTrace := make([]byte, 1<<16)
	rawStackTrace = rawStackTrace[:runtime.Stack(rawStackTrace, false)]
	stack2struct.Parse(rawStackTrace, &stack)

	return stack[2:]
}

// arrayMapToStringMap converts a map[string][]string to a map[string]string
// by joining all values of the containing array and wrapping them in brackets
func arrayMapToStringMap(arrayMap map[string][]string) map[string]string {
	entries := make(map[string]string)
	for k, v := range arrayMap {
		if len(v) > 1 {
			entries[k] = fmt.Sprintf("[%s]", strings.Join(v, "; "))
		} else {
			entries[k] = v[0]
		}
	}
	return entries
}
