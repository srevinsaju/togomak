package togomak

import (
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/sirupsen/logrus"
)

type Togomak struct {
	Parser *hclparse.Parser

	Logger *logrus.Logger
}

func NewTogomak(cfg Config) *Togomak {
	return &Togomak{
		Parser: hclparse.NewParser(),
		Logger: NewLogger(cfg),
	}

}
