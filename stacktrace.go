package logging

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

type StackTrace struct {
	Frames []runtime.Frame
}

func (s *StackTrace) String() string {
	var builder strings.Builder

	for _, frame := range s.Frames {
		builder.WriteString(frame.Func.Name())
		builder.WriteString("\n\t")
		builder.WriteString(frame.File)
		builder.WriteString(":")
		builder.WriteString(fmt.Sprintf("%d", frame.Line))
		builder.WriteString("\n")
	}

	return builder.String()
}

func NewStackTrace(skip int) StackTrace {
	const limit = 32

	pc := make([]uintptr, limit)
	frames := make([]runtime.Frame, 0, limit)

	// Gather frame info
	pc = pc[:runtime.Callers(skip, pc)]
	it := runtime.CallersFrames(pc)
	for {
		frame, more := it.Next()
		frames = append(frames, frame)

		if !more {
			break
		}
	}

	return StackTrace{
		Frames: frames,
	}
}

type StackTraceError struct {
	err   error
	Stack StackTrace
}

func (e StackTraceError) Unwrap() error {
	return e.err
}

func (e StackTraceError) Error() string {
	return fmt.Sprintf("stacktraced: %s", e.err)
}

func NewStackTraceError(format string, a ...any) error {
	return StackTraceError{
		err:   fmt.Errorf(format, a...),
		Stack: NewStackTrace(3),
	}
}

// See https://cs.opensource.google/go/go/+/master:src/errors/join.go;l=53
type joinError interface {
	Unwrap() []error
}

// Extract the underlying error array from joined errors
func joinedErrs(err error) []error {
	var errs []error

	t := reflect.TypeOf(err)
	if t.Kind().String() == "ptr" && t.String() == "*errors.joinError" {
		joinErr, ok := err.(joinError)
		if ok {
			errs = joinErr.Unwrap()
		}
	}

	return errs
}

type frameInfo = map[string]string

type stackInfo = []frameInfo

func marshalStack(err error) interface{} {
	var stErr StackTraceError
	if !errors.As(err, &stErr) {
		return nil
	}

	errs := joinedErrs(err)
	num := len(errs)
	if num < 1 {
		errs = []error{stErr}
	}

	stackInfos := make([]stackInfo, 0, num)
	for i := range errs {
		err = errs[i]
		if !errors.As(err, &stErr) {
			continue
		}

		frames := stErr.Stack.Frames
		num = len(frames)
		if num < 1 {
			continue
		}

		info := make(stackInfo, num)
		for j, frame := range frames {
			file := frame.File
			line := frame.Line
			name := frame.Func.Name()
			i := strings.LastIndex(name, ".")
			if i != -1 {
				name = name[i+1:]
			}

			info[j] = frameInfo{
				"file":     file,
				"function": name,
				"line":     fmt.Sprintf("%d", line),
			}
		}

		stackInfos = append(stackInfos, info)
	}

	return stackInfos
}
