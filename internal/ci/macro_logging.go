package ci

import (
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/internal/blocks"
	"github.com/srevinsaju/togomak/v1/internal/global"
)

func (m *Macro) Logger() *logrus.Entry {
	return global.Logger().WithField(blocks.MacroBlock, m.Id)
}
