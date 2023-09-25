package ci

import (
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/global"
)

func (s *Data) Logger() *logrus.Entry {
	return global.Logger().WithField("data", s.Id)
}
