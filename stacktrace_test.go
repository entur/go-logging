package logging

import (
	"strings"
	"testing"
)

func TestNewStack(t *testing.T) {
	type Expected struct {
		frameN int
		file   string
		name   string
		line   int
	}

	type Test struct {
		title    string
		callback func() StackTrace
		expected Expected
	}

	// Anonymous func1
	inner := func() StackTrace {
		return NewStackTrace(2)
	}

	// Anonymous func2
	outer := func() StackTrace {
		return inner()
	}

	var tests = []Test{
		{
			title:    "inner frame",
			callback: inner,
			expected: Expected{
				file: "/stacktrace_test.go",
				name: "func1",
				line: 24,
			},
		},
		{
			title:    "outer frame",
			callback: outer,
			expected: Expected{
				frameN: 1,
				file:   "/stacktrace_test.go",
				name:   "func2",
				line:   29,
			},
		},
	}

	for _, test := range tests {
		tmp := test
		t.Run(tmp.title, func(t *testing.T) {
			t.Parallel()

			stack := tmp.callback()
			if len(stack.Frames)+1 < tmp.expected.frameN {
				t.Errorf("specified frameN is not within vakud stack frame range\ngot: %d\nwant: <%d", tmp.expected.frameN, len(stack.Frames))
			} else {
				frame := stack.Frames[tmp.expected.frameN]

				file := frame.File
				line := frame.Line
				name := frame.Func.Name()
				i := strings.LastIndex(name, ".")
				if i != -1 {
					name = name[i+1:]
				}

				if tmp.expected.file != "" && !strings.HasSuffix(file, tmp.expected.file) {
					t.Errorf("frame has file name suffix\ngot: %s\nwant: %s", file, tmp.expected.file)
				}
				if tmp.expected.line != 0 && line != tmp.expected.line {
					t.Errorf("frame has incorrect call line number\ngot: %d\nwant: %d", line, tmp.expected.line)
				}
				if tmp.expected.name != "" && name != tmp.expected.name {
					t.Errorf("frame has incorrect function name\ngot: %s\nwant: %s", name, tmp.expected.name)
				}
			}
		})
	}
}
