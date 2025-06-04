package logging

import (
	"errors"
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
				line:   27,
			},
		},
		{
			title:    "outer frame",
			callback: outer,
			expected: Expected{
				frameN: 1,
				file:   "/stacktrace_test.go",
				name:   "func2",
				line:   32,
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

			if stack.String() == "" {
				t.Error("StackTrace string must generate string")
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
	type Expected = []stackInfo

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
			expected:   []stackInfo{},
		},
		{
			title:      "valid stacktrace error",
			stacktrace: NewStackTraceError("valid error"),
			expected: []stackInfo{
				[]frameInfo{
					{
						"file":     "/stacktrace_test.go",
						"function": "TestMarshalStack",
						"line":     "122",
					},
					{
						"function": "tRunner",
					},
					{
						"function": "goexit",
					},
				},
			},
		},
		{
			title: "joined stacktrace errors",
			stacktrace: errors.Join(
				NewStackTraceError("valid error 1"),
				fmt.Errorf("wrong error"),
				NewStackTraceError("valid error 2"),
			),
			expected: []stackInfo{
				[]frameInfo{
					{
						"file":     "/stacktrace_test.go",
						"function": "TestMarshalStack",
						"line":     "142",
					},
					{
						"function": "tRunner",
					},
					{
						"function": "goexit",
					},
				},
				[]frameInfo{
					{
						"file":     "/stacktrace_test.go",
						"function": "TestMarshalStack",
						"line":     "144",
					},
					{
						"function": "tRunner",
					},
					{
						"function": "goexit",
					},
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
				stackInfos, ok := data.([]stackInfo)
				if !ok {
					t.Fatalf("marshalled result is of the wrong type\ngot: %v\nwant: [][]map[string]string", reflect.TypeOf(data))
				}
				if len(tmp.expected) != len(stackInfos) {
					t.Fatalf("stackInfo array has incorrect length\ngot: %d\nwant: %d", len(stackInfos), len(tmp.expected))
				}

				for i, info := range stackInfos {
					expectInfo := tmp.expected[i]
					if len(expectInfo) != len(info) {
						t.Fatalf("stackInfo %d has incorrect length\ngot: %d\nwant: %d", i, len(info), len(expectInfo))
					}

					for j, frameInfo := range info {
						expect := expectInfo[j]["file"]
						if expect != "" && !strings.HasSuffix(frameInfo["file"], expect) {
							t.Errorf("frameInfo has incorrect file name suffix\ngot: %s\nwant: %s", frameInfo["file"], expect)
						}

						expect = expectInfo[j]["line"]
						if expect != "" && expect != frameInfo["line"] {
							t.Errorf("frameInfo has incorrect call line number\ngot: %s\nwant: %s", frameInfo["line"], expect)
						}

						expect = expectInfo[j]["function"]
						if expect != "" && expect != frameInfo["function"] {
							t.Errorf("frameInfo has incorrect function name\ngot: %s\nwant: %s", frameInfo["function"], expect)
						}
					}
				}
			}
		})
	}
}
