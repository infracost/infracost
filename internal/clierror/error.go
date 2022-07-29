package clierror

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/maruel/panicparse/v2/stack"
	log "github.com/sirupsen/logrus"
)

// SanitizedError allows errors to be wrapped with a sanitized message for sending upstream
type SanitizedError interface {
	SanitizedError() string
	SanitizedStack() string
}

type CLIError struct {
	sanitizedMsg string
	err          error
}

func NewCLIError(err error, sanitizedMsg string) *CLIError {
	return &CLIError{
		sanitizedMsg: sanitizedMsg,
		err:          err,
	}
}

func (e *CLIError) Error() string {
	return e.err.Error()
}

func (e *CLIError) SanitizedError() string {
	return e.sanitizedMsg
}

func (e *CLIError) SanitizedStack() string {
	return ""
}

// PanicError is used to collect goroutine panics into an error interface so
// that we can do type assertion on err checking.
type PanicError struct {
	err   error
	stack []byte
}

func NewPanicError(err error, stack []byte) *PanicError {
	return &PanicError{
		err:   err,
		stack: stack,
	}
}

func (p *PanicError) Error() string {
	return fmt.Sprintf("%s\n%s", p.err.Error(), p.stack)
}

func (p *PanicError) SanitizedError() string {
	return p.err.Error()
}

func (p *PanicError) SanitizedStack() string {
	sanitizedStack := p.stack
	sanitizedStack, err := processStack(sanitizedStack)
	if err != nil {
		log.Debugf("Could not sanitize stack: %s", err)
	}

	return string(sanitizedStack)
}

// processStack processes the raw stack trace into a format that can be reported upstream.
// This strips out the absolute path information of the local machine and removes any function args
// so that the reported value always equals the same string value as an equivalent stack trace.
// Adapted from: https://pkg.go.dev/github.com/maruel/panicparse/v2/stack#example-package-Text
func processStack(rawStack []byte) ([]byte, error) {
	stream := bytes.NewReader(rawStack)
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	s, suffix, err := stack.ScanSnapshot(stream, ioutil.Discard, stack.DefaultOpts())
	if err != nil && err != io.EOF {
		return []byte{}, err
	}

	// Find out similar goroutine traces and group them into buckets.
	buckets := s.Aggregate(stack.AnyValue).Buckets

	// Calculate padding for alignment.
	srcLen := 0
	for _, bucket := range buckets {
		for _, line := range bucket.Signature.Stack.Calls {
			if l := len(fmt.Sprintf("%s:%d", line.ImportPath, line.Line)); l > srcLen {
				srcLen = l
			}
		}
	}

	for _, bucket := range buckets {
		// Print the goroutine header.
		extra := ""
		if s := bucket.SleepString(); s != "" {
			extra += " [" + s + "]"
		}
		if bucket.Locked {
			extra += " [locked]"
		}

		if len(bucket.CreatedBy.Calls) != 0 {
			extra += fmt.Sprintf(" [Created by %s.%s @ %s:%d]", bucket.CreatedBy.Calls[0].Func.DirName, bucket.CreatedBy.Calls[0].Func.Name, bucket.CreatedBy.Calls[0].SrcName, bucket.CreatedBy.Calls[0].Line)
		}
		fmt.Fprintf(w, "%d: %s%s\n", len(bucket.IDs), bucket.State, extra)

		// Print the stack lines.
		for _, line := range bucket.Stack.Calls {
			_, err := fmt.Fprintf(w,
				"   %-*s %s()\n",
				srcLen,
				fmt.Sprintf("%s:%d", line.RelSrcPath, line.Line),
				line.Func.Name)
			if err != nil {
				return []byte{}, err
			}
		}
		if bucket.Stack.Elided {
			_, err := w.WriteString("    (...)\n")
			if err != nil {
				return []byte{}, err
			}
		}
	}

	// If there was any remaining data in the pipe, dump it now.
	if len(suffix) != 0 {
		_, err := w.Write(suffix)
		if err != nil {
			return []byte{}, err
		}
	}
	_, err = io.Copy(w, stream)
	if err != nil {
		return []byte{}, err
	}

	w.Flush()

	return buf.Bytes(), nil
}
