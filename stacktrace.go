package logging

import (
	"errors"
	"fmt"
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

func marshalStack(err error) interface{} {
	var stErr StackTraceError
	if !errors.As(err, &stErr) {
		return nil
	}

	info := make([]map[string]string, 0)
	stack := stErr.Stack
	for _, frame := range stack.Frames {
		file := frame.File
		line := frame.Line
		name := frame.Func.Name()
		i := strings.LastIndex(name, ".")
		if i != -1 {
			name = name[i+1:]
		}

		info = append(info, map[string]string{
			"file":     file,
			"function": name,
			"line":     fmt.Sprintf("%d", line),
		})
	}

	return info
}
