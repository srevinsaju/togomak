package global

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

var hclContext *hcl.EvalContext
var hclDiagWriter hcl.DiagnosticWriter
var hclParser *hclparse.Parser

func SetHclEvalContext(ctx *hcl.EvalContext) {
	hclContext = ctx
}

func HclEvalContext() *hcl.EvalContext {
	return hclContext
}

func SetHclDiagWriter(w hcl.DiagnosticWriter) {
	hclDiagWriter = w
}

func HclDiagWriter() hcl.DiagnosticWriter {
	return hclDiagWriter
}

func SetHclParser(parser *hclparse.Parser) {
	hclParser = parser
}

func HclParser() *hclparse.Parser {
	return hclParser
}
