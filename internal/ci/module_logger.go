package ci

import (
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/internal/global"
)

func (m *Module) Logger() *logrus.Entry {
	return global.Logger().WithField("module", m.Id)
}
