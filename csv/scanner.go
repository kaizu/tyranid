// https://github.com/smartystreets/scanners/blob/master/csv/scanner.go

package csv

import (
	"encoding/csv"
	"io"
)

// Scanner wraps a csv.Reader via an API similar to that of bufio.Scanner.
type Scanner struct {
	reader *csv.Reader
	record []string
	err    error

	continueOnError bool
}

// NewScanner returns a scanner configured with the provided options.
func NewScanner(reader io.Reader, options ...Option) *Scanner {
	return new(Scanner).initialize(reader).configure(options)
}
func (this *Scanner) initialize(reader io.Reader) *Scanner {
	this.reader = csv.NewReader(reader)
	return this
}
func (this *Scanner) configure(options []Option) *Scanner {
	for _, configure := range options {
		configure(this)
	}
	return this
}

// Scan advances the Scanner to the next record, which will then be available
// through the Record method. It returns false when the scan stops, either by
// reaching the end of the input or an error. After Scan returns false, the
// Error method will return any error that occurred during scanning, except
// that if it was io.EOF, Error will return nil.
func (this *Scanner) Scan() bool {
	if this.eof() {
		return false
	}
	this.record, this.err = this.reader.Read()
	return !this.eof()
}

func (this *Scanner) eof() bool {
	if this.err == io.EOF {
		return true
	}
	if this.err == nil {
		return false
	}
	return !this.continueOnError
}

// Record returns the most recent record generated by a call to Scan as a
// []string. See *csv.Reader.ReuseRecord for details on the strategy for
// reusing the underlying array: https://golang.org/pkg/encoding/csv/#Reader
func (this *Scanner) Record() []string {
	return this.record
}

// Error returns the last non-nil error produced by Scan (if there was one).
// It will not ever return io.EOF. This method may be called at any point
// during or after scanning but the underlying err will be reset by each call
// to Scan.
func (this *Scanner) Error() error {
	if this.err == io.EOF {
		return nil
	}
	return this.err
}
