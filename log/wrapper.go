package log

import (
	"context"
	"github.com/acarl005/stripansi"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/pkg/meta"
	"github.com/srevinsaju/togomak/pkg/x"
	"github.com/srevinsaju/togomak/pkg/client/logging"
	"os"
)

var togomakHostname string
var togomakTrackingClient *logging.Client

type TrackingServerClientConfig struct {
	URL string 

}

func TogomakTrackingClient(config TrackingServerClientConfig) (*logging.Client, error) {
	if togomakTrackingClient != nil {
		return togomakTrackingClient, nil
	}
	loggerContext := context.Background()
	togomakHostname = x.MustReturn(os.Hostname()).(string)
	
	// initialize the client
	client, err := logging.NewClient(loggerContext, config.URL)
	if err != nil {
		return nil, err
	}
	togomakTrackingClient = client
	return togomakTrackingClient, nil
}

type TogomakHook struct {
	TrackingServer string
}

func (h TogomakHook) Fire(entry *logrus.Entry) error {
	// upload to google cloud logging
	// using google cloud API
	// https://cloud.google.com/logging/docs/reference/libraries#client-libraries-install-gow
	client, err := TogomakTrackingClient(TrackingServerClientConfig{
		URL: h.TrackingServer,

	})
	if err != nil {
		return err
	}
	
	logger := client.Logger()
	if err != nil {
		return err
	}
	
	logger.Log(logging.Entry{
		Payload: map[string]interface{}{
			"message": stripansi.Strip(entry.Message),
			"labels":  entry.Data,
			"app":     meta.AppName,
			"version": meta.Version,
			"host":    hostname,
		},
		Severity: entry.Level.String(),
		Labels: map[string]string{
			"app":          meta.AppName,
			"version":      meta.Version,
			"instanceName": meta.AppName,
			"instanceId":   meta.GetCorrelationId().String(),
		},
	})
	return nil
}

func (h TogomakHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.InfoLevel, logrus.WarnLevel, logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel}
}
