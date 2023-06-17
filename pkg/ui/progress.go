package ui

import (
	"fmt"
	"github.com/bcicen/jstream"
	"io"
	"time"
)

type ProgressWriter struct {
	reader      io.Reader
	bytesRead   int64
	verb        string
	created     time.Time
	lastWritten time.Time
	status      string
	lastStatus  string
	writer      io.Writer
}

func NewProgressWriter(reader io.Reader, writer io.Writer, verb string) *ProgressWriter {
	return &ProgressWriter{
		verb:        verb,
		reader:      reader,
		created:     time.Now(),
		lastWritten: time.Now(),
		writer:      writer,
		status:      "",
		lastStatus:  "",
	}
}

type Message struct {
	Status string `json:"status,omitempty"`
}

func (pr *ProgressWriter) Write(p []byte) (int, error) {
	d := jstream.NewDecoder(pr.reader, 0)

	for mv := range d.Stream() {
		m, ok := mv.Value.(map[string]interface{})

		if ok && m["status"] != nil {
			pr.status = m["status"].(string)
			pr.printProgress()
		}
	}
	return d.Pos(), nil
}

func (pr *ProgressWriter) printProgress() error {
	elapsed := time.Since(pr.lastWritten)
	var err error
	if elapsed.Seconds() >= 5 || pr.status != pr.lastStatus {
		pr.lastWritten = time.Now()
		pr.lastStatus = pr.status
		_, err = fmt.Fprintf(pr.writer, "Still %s... (%s) %s\n", pr.verb, pr.status, Bold(fmt.Sprintf("[%s elapsed]", time.Since(pr.created).Round(time.Second).String())))
	}
	return err
}

func (pr *ProgressWriter) Close() error {
	_, err := fmt.Fprintf(pr.writer, "Completed %s in %s\n", pr.verb, Bold(fmt.Sprintf("[%s elapsed]", time.Since(pr.created).Round(time.Second).String())))
	return err
}
