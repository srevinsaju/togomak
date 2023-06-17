package ui

import (
	"fmt"
	"github.com/bcicen/jstream"
	"github.com/sirupsen/logrus"
	"io"
	"time"
)

type DockerProgressWriter struct {
	reader      io.Reader
	bytesRead   int64
	verb        string
	created     time.Time
	lastWritten time.Time
	status      string
	lastStatus  string
	writer      io.Writer
}

type ProgressWriter struct {
	bytesRead   int64
	verb        string
	created     time.Time
	lastWritten time.Time
	logger      *logrus.Entry
}

func NewDockerProgressWriter(reader io.Reader, writer io.Writer, verb string) *DockerProgressWriter {
	return &DockerProgressWriter{
		verb:        verb,
		reader:      reader,
		created:     time.Now(),
		lastWritten: time.Now().Add(-time.Second * 5),
		writer:      writer,
		status:      "",
		lastStatus:  "",
	}
}

func NewProgressWriter(logger *logrus.Entry, verb string) *ProgressWriter {
	return &ProgressWriter{
		bytesRead:   0,
		verb:        verb,
		created:     time.Now(),
		lastWritten: time.Now().Add(-time.Second * 5),
		logger:      logger,
	}
}

type Message struct {
	Status string `json:"status,omitempty"`
}

func (pr *DockerProgressWriter) Write(p []byte) (int, error) {
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

func (pr *ProgressWriter) Write(p []byte) (int, error) {
	err := pr.printProgress()
	pr.bytesRead += int64(len(p))
	return len(p), err
}

func (pr *DockerProgressWriter) printProgress() error {
	elapsed := time.Since(pr.lastWritten)
	var err error
	if elapsed.Seconds() >= 5 || pr.status != pr.lastStatus {
		pr.lastWritten = time.Now()
		pr.lastStatus = pr.status
		_, err = fmt.Fprintf(pr.writer, "Still %s... (%s) %s\n", pr.verb, pr.status, Bold(fmt.Sprintf("[%s elapsed]", time.Since(pr.created).Round(time.Second).String())))
	}
	return err
}

func (pr *ProgressWriter) printProgress() error {
	elapsed := time.Since(pr.lastWritten)
	var err error
	if elapsed.Seconds() >= 5 {
		pr.lastWritten = time.Now()
		pr.logger.Infof("Still %s... %s\n", pr.verb, Bold(fmt.Sprintf("[%s elapsed]", time.Since(pr.created).Round(time.Second).String())))
	}
	return err
}

func (pr *DockerProgressWriter) Close() error {
	_, err := fmt.Fprintf(pr.writer, "Completed %s in %s\n", pr.verb, Bold(fmt.Sprintf("[%s elapsed]", time.Since(pr.created).Round(time.Second).String())))
	return err
}

func (pr *ProgressWriter) Close() error {
	pr.logger.Infof("Completed %s in %s\n", pr.verb, Bold(fmt.Sprintf("[%s elapsed]", time.Since(pr.created).Round(time.Second).String())))
	return nil
}
