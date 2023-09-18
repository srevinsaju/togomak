package ci

import (
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/global"
)

func (m *Module) Logger() *logrus.Entry {
	return global.Logger().WithField("module", m.Id)
}
