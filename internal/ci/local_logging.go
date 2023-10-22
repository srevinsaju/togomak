package ci

import (
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/internal/global"
)

func (l *Local) Logger() *logrus.Entry {
	return global.Logger().WithField(LocalBlock, l.Key)
}
