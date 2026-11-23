package logging

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"
	"testing/slogtest"
)

func TestSlogHandlerConformance(t *testing.T) {
	buf := &bytes.Buffer{}

	newHandler := func(*testing.T) slog.Handler {
		buf.Reset()

		return NewSlogHandler(WithWriter(buf), WithLevel(DebugLevel))
	}

	result := func(t *testing.T) map[string]any {
		t.Helper()

		line := bytes.TrimSpace(buf.Bytes())
		if len(line) == 0 {
			t.Fatal("handler wrote no output")
		}

		var m map[string]any
		if err := json.Unmarshal(line, &m); err != nil {
			t.Fatalf("output is not a JSON object: %v\n%s", err, line)
		}

		// Map the GCP field names back to the keys slog expects.
		for gcp, key := range map[string]string{
			"timestamp": slog.TimeKey,
			"severity":  slog.LevelKey,
			"message":   slog.MessageKey,
		} {
			if v, ok := m[gcp]; ok {
				delete(m, gcp)
				m[key] = v
			}
		}

		return m
	}

	slogtest.Run(t, newHandler, result)
}
