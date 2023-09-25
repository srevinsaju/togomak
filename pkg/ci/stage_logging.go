package ci

import (
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/global"
)

func (s *Stage) Logger() *logrus.Entry {
	return global.Logger().WithField("stage", s.Id)
}
