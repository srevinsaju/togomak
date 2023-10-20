package ui

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/internal/x"
	"io"
)

// GetterProgressBar is a progress bar implementation for terraform downloads
type GetterProgressBar struct {
	Logger *logrus.Entry
	src    string
	pb     *ProgressWriter
}

func NewGetterProgressBar(logger *logrus.Entry, src string) *GetterProgressBar {
	return &GetterProgressBar{
		Logger: logger,
		src:    src,
	}
}

// Init initializes the progress bar
func (e *GetterProgressBar) Init() *GetterProgressBar {
	e.pb = NewProgressWriter(e.Logger, fmt.Sprintf("downloading %s", e.src))
	return e
}

// TrackProgress tracks the progress of the download using the default ui.ProgressWriter implementation
func (e *GetterProgressBar) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) (body io.ReadCloser) {
	for {
		_, err := io.CopyN(e.pb, stream, 1)
		if err != nil {
			x.Must(e.pb.Close())
			return stream
		}
	}
}
