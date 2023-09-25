package ci

import (
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/global"
)

func (m *Macro) Logger() *logrus.Entry {
	return global.Logger().WithField(MacroBlock, m.Id)
}
