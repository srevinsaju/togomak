package logging

import (
	"context"
	"github.com/acarl005/stripansi"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/internal/meta"
	"github.com/srevinsaju/togomak/v1/internal/x"
	"google.golang.org/genproto/googleapis/api/monitoredres"
	"os"
)
import "cloud.google.com/go/logging"

var hostname string

func googleCloudLoggingClient(project string) (*logging.Client, error) {
	loggerContext := context.Background()
	hostname = x.MustReturn(os.Hostname()).(string)

	// initialize the client
	client, err := logging.NewClient(loggerContext, project)
	return client, err
}

type GoogleCloudLoggerHook struct {
	client  *logging.Client
	cfg     Config
	project string
}

func NewGoogleCloudLoggerHook(cfg Config, project string) (*GoogleCloudLoggerHook, error) {
	client, err := googleCloudLoggingClient(project)
	return &GoogleCloudLoggerHook{cfg: cfg, client: client, project: project}, err
}

func (h *GoogleCloudLoggerHook) Fire(entry *logrus.Entry) error {
	// upload to google cloud logging
	// using google cloud API
	// https://cloud.google.com/logging/docs/reference/libraries#client-libraries-install-gow
	client := h.client
	logger := client.Logger(meta.AppName)
	severityLevel := logging.Default
	switch entry.Level {
	case logrus.DebugLevel:
		severityLevel = logging.Debug
	case logrus.InfoLevel:
		severityLevel = logging.Info
	case logrus.WarnLevel:
		severityLevel = logging.Warning
	case logrus.ErrorLevel:
		severityLevel = logging.Error
	case logrus.FatalLevel:
		severityLevel = logging.Critical
	case logrus.PanicLevel:
		severityLevel = logging.Alert
	}
	logger.Log(logging.Entry{
		Payload: map[string]interface{}{
			"message": stripansi.Strip(entry.Message),
			"labels":  entry.Data,
			"app":     meta.AppName,
			"version": meta.AppVersion,
			"host":    hostname,
		},
		Resource: &monitoredres.MonitoredResource{Type: "global"},
		Trace:    "togomak",
		Severity: severityLevel,
		Labels: map[string]string{
			"app":          meta.AppName,
			"version":      meta.AppVersion,
			"instanceName": meta.AppName,
			"instanceId":   h.cfg.CorrelationID,
		},
	})
	return nil
}

func (h *GoogleCloudLoggerHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.InfoLevel, logrus.WarnLevel, logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel}
}
