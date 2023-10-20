package global

import (
	"github.com/hashicorp/hcl/v2"
)

var hclContext *hcl.EvalContext

func SetHclEvalContext(ctx *hcl.EvalContext) {
	hclContext = ctx
}

func HclEvalContext() *hcl.EvalContext {
	return hclContext
}
