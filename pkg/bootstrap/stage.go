package bootstrap

import (
	"github.com/imdario/mergo"
	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/schema"
	"regexp"
	"strings"
)

var StageIdValidation = regexp.MustCompile(`^([a-zA-Z_]+)$`)

func StageValidate(ctx *context.Context, data schema.SchemaConfig) {

	stages := map[string]string{}

	validateLog := ctx.Logger.WithField("context", "validate")

	// check if duplicate ID is present
	for _, stage := range data.Stages {
		if !StageIdValidation.MatchString(stage.Id) {
			validateLog.Fatalf("Stage ID must contain only alphabets: %s", stage.Id)
		}
		if _, ok := stages[stage.Id]; ok {
			validateLog.Fatal("Duplicate stage ID: " + stage.Id)
		}
		stages[stage.Id] = stage.Id
	}

	// extend the current stage if .extends is present
	for i, stage := range data.Stages {
		if stage.Extends != "" && strings.HasPrefix(stage.Extends, ".") {

			extendsKey := stage.Extends[1:]
			validateLog.Debugf("Extending stage %s with %s", stage.Id, extendsKey)

			_, ok := stages[extendsKey]
			if !ok {
				validateLog.Fatal("Stage " + stage.Id + " extends non-existing stage " + extendsKey)
			}
			extendsStage := data.Stages.GetStageById(extendsKey)

			err := mergo.Merge(&data.Stages[i], extendsStage)
			if err != nil {
				validateLog.Fatal("merge of extends stage failed ", err)
			}

		}
		// TODO: implement extends from git
	}
}
