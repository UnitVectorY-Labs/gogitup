package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// ANSI color and style constants.
const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	Gray   = "\033[90m"
)

// Writer wraps an io.Writer and provides styled terminal output helpers.
type Writer struct {
	Out io.Writer
}

// DefaultWriter is the default Writer targeting os.Stdout.
var DefaultWriter = &Writer{Out: os.Stdout}

// ErrorWriter is the default Writer targeting os.Stderr.
var ErrorWriter = &Writer{Out: os.Stderr}

// Success prints a green checkmark followed by the message.
func (w *Writer) Success(msg string) {
	fmt.Fprintf(w.Out, "%s✓%s %s\n", Green, Reset, msg)
}

// Error prints a red X followed by the message.
func (w *Writer) Error(msg string) {
	fmt.Fprintf(w.Out, "%s✗%s %s\n", Red, Reset, msg)
}

// Warn prints a yellow warning symbol followed by the message.
func (w *Writer) Warn(msg string) {
	fmt.Fprintf(w.Out, "%s⚠%s %s\n", Yellow, Reset, msg)
}

// Info prints a blue info symbol followed by the message.
func (w *Writer) Info(msg string) {
	fmt.Fprintf(w.Out, "%sℹ%s %s\n", Blue, Reset, msg)
}

// Header prints a bold cyan header line.
func (w *Writer) Header(msg string) {
	fmt.Fprintf(w.Out, "%s%s%s%s\n", Bold, Cyan, msg, Reset)
}

// PrintJSON marshals v as indented JSON and writes it to the writer.
func (w *Writer) PrintJSON(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	_, err = fmt.Fprintf(w.Out, "%s\n", data)
	return err
}

// StartProgress prints a working indicator message.
func (w *Writer) StartProgress(msg string) {
	fmt.Fprintf(w.Out, "%s⟳%s %s...\n", Cyan, Reset, msg)
}

// EndProgress prints a completion message.
func (w *Writer) EndProgress(msg string) {
	fmt.Fprintf(w.Out, "%s✓%s %s\n", Green, Reset, msg)
}

// Success prints a green checkmark followed by the message to stdout.
func Success(msg string) {
	DefaultWriter.Success(msg)
}

// Error prints a red X followed by the message to stderr.
func Error(msg string) {
	ErrorWriter.Error(msg)
}

// Warn prints a yellow warning symbol followed by the message to stdout.
func Warn(msg string) {
	DefaultWriter.Warn(msg)
}

// Info prints a blue info symbol followed by the message to stdout.
func Info(msg string) {
	DefaultWriter.Info(msg)
}

// Header prints a bold cyan header line to stdout.
func Header(msg string) {
	DefaultWriter.Header(msg)
}

// PrintJSON marshals v as indented JSON and writes it to stdout.
func PrintJSON(v any) error {
	return DefaultWriter.PrintJSON(v)
}

// StartProgress prints a working indicator message to stdout.
func StartProgress(msg string) {
	DefaultWriter.StartProgress(msg)
}

// EndProgress prints a completion message to stdout.
func EndProgress(msg string) {
	DefaultWriter.EndProgress(msg)
}
