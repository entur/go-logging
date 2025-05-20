package logging

import (
	"fmt"
	"reflect"
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
				frameN: 0,
				file:   "/stacktrace_test.go",
				name:   "func1",
				line:   26,
			},
		},
		{
			title:    "outer frame",
			callback: outer,
			expected: Expected{
				frameN: 1,
				file:   "/stacktrace_test.go",
				name:   "func2",
				line:   31,
			},
		},
	}

	for _, test := range tests {
		tmp := test
		t.Run(tmp.title, func(t *testing.T) {
			t.Parallel()

			stack := tmp.callback()
			if len(stack.Frames)+1 < tmp.expected.frameN {
				t.Fatalf("specified frameN is not within valid stack frame range\ngot: %d\nwant: <%d", len(stack.Frames), tmp.expected.frameN)
			}

			frame := stack.Frames[tmp.expected.frameN]

			file := frame.File
			line := frame.Line
			name := frame.Func.Name()
			i := strings.LastIndex(name, ".")
			if i != -1 {
				name = name[i+1:]
			}

			expect := tmp.expected.file
			if expect != "" && !strings.HasSuffix(file, expect) {
				t.Errorf("frame has file name suffix\ngot: %s\nwant: %s", file, expect)
			}

			expectI := tmp.expected.line
			if expectI != 0 && line != expectI {
				t.Errorf("frame has incorrect call line number\ngot: %d\nwant: %d", line, expectI)
			}

			expect = tmp.expected.name
			if expect != "" && name != expect {
				t.Errorf("frame has incorrect function name\ngot: %s\nwant: %s", name, expect)
			}
		})
	}
}

func TestMarshalStack(t *testing.T) {
	type Expected = []map[string]string

	type Test struct {
		title      string
		stacktrace error
		expected   Expected
	}

	var tests = []Test{
		{
			title:      "wrong error",
			stacktrace: fmt.Errorf("wrong error"),
			expected:   nil,
		},
		{
			title:      "empty stacktrace error",
			stacktrace: StackTraceError{},
			expected:   []map[string]string{},
		},
		{
			title:      "valid stacktrace error",
			stacktrace: NewStackTraceError("valid error"),
			expected: []map[string]string{
				{
					"file":     "/stacktrace_test.go",
					"function": "TestMarshalStack",
					"line":     "117",
				},
				{
					"function": "tRunner",
				},
				{
					"function": "goexit",
				},
			},
		},
	}

	for _, test := range tests {
		tmp := test
		t.Run(tmp.title, func(t *testing.T) {
			t.Parallel()

			data := marshalStack(tmp.stacktrace)
			if tmp.expected == nil {
				if data != nil {
					t.Fatalf("marshalled result value is incorrect\ngot: %+v\nwant: nil", data)
				}
			} else {
				info, ok := data.([]map[string]string)
				if !ok {
					t.Fatalf("marshalled result is of the wrong type\ngot: %v\nwant: []map[string]string", reflect.TypeOf(data))
				}
				if len(tmp.expected) != len(info) {
					t.Fatalf("stack info has incorrect length\ngot: %d\nwant: %d", len(info), len(tmp.expected))
				}

				for i, frameInfo := range info {
					expect := tmp.expected[i]["file"]
					if expect != "" && !strings.HasSuffix(frameInfo["file"], expect) {
						t.Errorf("frame info has incorrect file name suffix\ngot: %s\nwant: %s", frameInfo["file"], expect)
					}

					expect = tmp.expected[i]["line"]
					if expect != "" && expect != frameInfo["line"] {
						t.Errorf("frame info has incorrect call line number\ngot: %s\nwant: %s", frameInfo["line"], expect)
					}

					expect = tmp.expected[i]["function"]
					if expect != "" && expect != frameInfo["function"] {
						t.Errorf("frame info has incorrect function name\ngot: %s\nwant: %s", frameInfo["function"], expect)
					}
				}
			}
		})
	}
}
