package log

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestLogging(t *testing.T) {
	// Save original output to restore later
	originalOutput := os.Stdout
	defer SetOutput(originalOutput)

	var buf bytes.Buffer
	SetOutput(&buf)

	t.Run("TraceLevel", func(t *testing.T) {
		buf.Reset()
		SetLevel(TRACE)

		Trace("trace message")
		Debug("debug message")
		Info("info message")
		Warning("warning message")
		Error("error message")

		output := buf.String()

		// All messages should appear
		if !strings.Contains(output, "[TRACE]") {
			t.Error("TRACE level should show trace messages")
		}
		if !strings.Contains(output, "[DEBUG]") {
			t.Error("TRACE level should show debug messages")
		}
		if !strings.Contains(output, "[INFO]") {
			t.Error("TRACE level should show info messages")
		}
		if !strings.Contains(output, "[WARN]") {
			t.Error("TRACE level should show warning messages")
		}
		if !strings.Contains(output, "[ERROR]") {
			t.Error("TRACE level should show error messages")
		}
	})

	t.Run("DebugLevel", func(t *testing.T) {
		buf.Reset()
		SetLevel(DEBUG)

		Trace("trace message")
		Debug("debug message")
		Info("info message")
		Warning("warning message")
		Error("error message")

		output := buf.String()

		// Trace should not appear, others should
		if strings.Contains(output, "[TRACE]") {
			t.Error("DEBUG level should not show trace messages")
		}
		if !strings.Contains(output, "[DEBUG]") {
			t.Error("DEBUG level should show debug messages")
		}
		if !strings.Contains(output, "[INFO]") {
			t.Error("DEBUG level should show info messages")
		}
		if !strings.Contains(output, "[WARN]") {
			t.Error("DEBUG level should show warning messages")
		}
		if !strings.Contains(output, "[ERROR]") {
			t.Error("DEBUG level should show error messages")
		}
	})

	t.Run("InfoLevel", func(t *testing.T) {
		buf.Reset()
		SetLevel(INFO)

		Trace("trace message")
		Debug("debug message")
		Info("info message")
		Warning("warning message")
		Error("error message")

		output := buf.String()

		// Trace and Debug should not appear
		if strings.Contains(output, "[TRACE]") {
			t.Error("INFO level should not show trace messages")
		}
		if strings.Contains(output, "[DEBUG]") {
			t.Error("INFO level should not show debug messages")
		}
		if !strings.Contains(output, "[INFO]") {
			t.Error("INFO level should show info messages")
		}
		if !strings.Contains(output, "[WARN]") {
			t.Error("INFO level should show warning messages")
		}
		if !strings.Contains(output, "[ERROR]") {
			t.Error("INFO level should show error messages")
		}
	})

	t.Run("WarnLevel", func(t *testing.T) {
		buf.Reset()
		SetLevel(WARN)

		Trace("trace message")
		Debug("debug message")
		Info("info message")
		Warning("warning message")
		Error("error message")

		output := buf.String()

		// Only WARN and ERROR should appear
		if strings.Contains(output, "[TRACE]") {
			t.Error("WARN level should not show trace messages")
		}
		if strings.Contains(output, "[DEBUG]") {
			t.Error("WARN level should not show debug messages")
		}
		if strings.Contains(output, "[INFO]") {
			t.Error("WARN level should not show info messages")
		}
		if !strings.Contains(output, "[WARN]") {
			t.Error("WARN level should show warning messages")
		}
		if !strings.Contains(output, "[ERROR]") {
			t.Error("WARN level should show error messages")
		}
	})

	t.Run("ErrorLevel", func(t *testing.T) {
		buf.Reset()
		SetLevel(ERROR)

		Trace("trace message")
		Debug("debug message")
		Info("info message")
		Warning("warning message")
		Error("error message")

		output := buf.String()

		// Only ERROR should appear
		if strings.Contains(output, "[TRACE]") {
			t.Error("ERROR level should not show trace messages")
		}
		if strings.Contains(output, "[DEBUG]") {
			t.Error("ERROR level should not show debug messages")
		}
		if strings.Contains(output, "[INFO]") {
			t.Error("ERROR level should not show info messages")
		}
		if strings.Contains(output, "[WARN]") {
			t.Error("ERROR level should not show warning messages")
		}
		if !strings.Contains(output, "[ERROR]") {
			t.Error("ERROR level should show error messages")
		}
	})

	t.Run("FormattingTest", func(t *testing.T) {
		buf.Reset()
		SetLevel(DEBUG)

		Debug("Debug with value: %d", 42)
		Info("Info with string: %s", "test")
		Warning("Warning with multiple: %s = %d", "count", 5)
		Error("Error with float: %.2f", 3.14159)

		output := buf.String()

		// Check that formatting worked correctly
		if !strings.Contains(output, "Debug with value: 42") {
			t.Error("Debug formatting failed")
		}
		if !strings.Contains(output, "Info with string: test") {
			t.Error("Info formatting failed")
		}
		if !strings.Contains(output, "Warning with multiple: count = 5") {
			t.Error("Warning formatting failed")
		}
		if !strings.Contains(output, "Error with float: 3.14") {
			t.Error("Error formatting failed")
		}
	})

	t.Run("OutputPreservesLevel", func(t *testing.T) {
		// Test that changing output preserves the current log level
		SetLevel(DEBUG)

		var buf1, buf2 bytes.Buffer

		// Set first output and test
		SetOutput(&buf1)
		Debug("debug to buf1")
		Info("info to buf1")

		// Change output and test again
		SetOutput(&buf2)
		Debug("debug to buf2")
		Info("info to buf2")

		// Both buffers should have debug messages since level should be preserved
		if !strings.Contains(buf1.String(), "[DEBUG]") {
			t.Error("First buffer should contain debug messages")
		}
		if !strings.Contains(buf2.String(), "[DEBUG]") {
			t.Error("Second buffer should contain debug messages - level was not preserved")
		}
	})
}
