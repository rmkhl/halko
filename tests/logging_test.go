package tests

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/rmkhl/halko/types/log"
)

func TestLogging(t *testing.T) {
	// Save original output to restore later
	originalOutput := os.Stdout
	defer log.SetOutput(originalOutput)

	var buf bytes.Buffer
	log.SetOutput(&buf)

	t.Run("TraceLevel", func(t *testing.T) {
		buf.Reset()
		log.SetLevel(log.TRACE)

		log.Trace("trace message")
		log.Debug("debug message")
		log.Info("info message")
		log.Warning("warning message")
		log.Error("error message")

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
		log.SetLevel(log.DEBUG)

		log.Trace("trace message")
		log.Debug("debug message")
		log.Info("info message")
		log.Warning("warning message")
		log.Error("error message")

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
		log.SetLevel(log.INFO)

		log.Trace("trace message")
		log.Debug("debug message")
		log.Info("info message")
		log.Warning("warning message")
		log.Error("error message")

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
		log.SetLevel(log.WARN)

		log.Trace("trace message")
		log.Debug("debug message")
		log.Info("info message")
		log.Warning("warning message")
		log.Error("error message")

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
		log.SetLevel(log.ERROR)

		log.Trace("trace message")
		log.Debug("debug message")
		log.Info("info message")
		log.Warning("warning message")
		log.Error("error message")

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
		log.SetLevel(log.DEBUG)

		log.Debug("Debug with value: %d", 42)
		log.Info("Info with string: %s", "test")
		log.Warning("Warning with multiple: %s = %d", "count", 5)
		log.Error("Error with float: %.2f", 3.14159)

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
		log.SetLevel(log.DEBUG)

		var buf1, buf2 bytes.Buffer

		// Set first output and test
		log.SetOutput(&buf1)
		log.Debug("debug to buf1")
		log.Info("info to buf1")

		// Change output and test again
		log.SetOutput(&buf2)
		log.Debug("debug to buf2")
		log.Info("info to buf2")

		// Both buffers should have debug messages since level should be preserved
		if !strings.Contains(buf1.String(), "[DEBUG]") {
			t.Error("First buffer should contain debug messages")
		}
		if !strings.Contains(buf2.String(), "[DEBUG]") {
			t.Error("Second buffer should contain debug messages - level was not preserved")
		}
	})
}
