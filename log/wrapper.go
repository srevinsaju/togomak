package log

import (
	"context"
	"fmt"
	"github.com/acarl005/stripansi"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/pkg/meta"
	"github.com/srevinsaju/togomak/pkg/x"
	"google.golang.org/genproto/googleapis/api/monitoredres"
	"os"
)
import "cloud.google.com/go/logging"

var hostname string
var googleCloudLoggerClient *logging.Client

func GoogleCloudLoggerClient() (*logging.Client, error) {
	if googleCloudLoggerClient != nil {
		return googleCloudLoggerClient, nil
	}
	loggerContext := context.Background()
	hostname = x.MustReturn(os.Hostname()).(string)
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		panic("GOOGLE_CLOUD_PROJECT environment variable is not set")
	}

	// initialize the client
	client, err := logging.NewClient(loggerContext, projectID)
	if err != nil {
		return nil, err
	}
	googleCloudLoggerClient = client
	return googleCloudLoggerClient, nil
}

type NormalHook struct {
}

func (h NormalHook) Fire(entry *logrus.Entry) error {
	fmt.Println(entry.Message)
	return nil
}
func (h NormalHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

type GoogleCloudLoggerHook struct {
}

func (h GoogleCloudLoggerHook) Fire(entry *logrus.Entry) error {
	// upload to google cloud logging
	// using google cloud API
	// https://cloud.google.com/logging/docs/reference/libraries#client-libraries-install-gow
	client, err := GoogleCloudLoggerClient()
	if err != nil {
		return err
	}
	logger := client.Logger(meta.AppName)
	if err != nil {
		return err
	}
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
			"version": meta.Version,
			"host":    hostname,
		},
		Resource: &monitoredres.MonitoredResource{Type: "global"},
		Trace:    "togomak",
		Severity: severityLevel,
		Labels: map[string]string{
			"app":          meta.AppName,
			"version":      meta.Version,
			"instanceName": meta.AppName,
			"instanceId":   meta.GetCorrelationId().String(),
		},
	})
	return nil
}

func (h GoogleCloudLoggerHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.InfoLevel, logrus.WarnLevel, logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel}
}
