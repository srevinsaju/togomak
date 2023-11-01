package ui

import (
	"fmt"
	"github.com/acarl005/stripansi"
	"github.com/fatih/color"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"os"
)

var Green = color.New(color.FgGreen).SprintFunc()
var HiCyan = color.New(color.FgHiCyan).SprintFunc()
var Red = color.New(color.FgRed).SprintFunc()
var Bold = color.New(color.Bold).SprintFunc()
var Blue = color.New(color.FgBlue).SprintFunc()
var Grey = color.New(color.FgHiBlack).SprintFunc()
var Yellow = color.New(color.FgYellow).SprintFunc()
var HiYellow = color.New(color.FgHiYellow).SprintFunc()
var HiRed = color.New(color.FgHiRed).SprintFunc()
var Italic = color.New(color.Italic).SprintFunc()
var Plus = color.New(color.FgHiWhite).SprintFunc()("+")
var SubStage = Grey("==>")
var SubSubStage = Grey("-->")
var Matrix = Blue("matrix")
var Stage = Blue("stage")
var Options = Grey("options")
var FailLazy = Grey("fail-lazy")

var True = Green("true")
var False = Red("false")

func Error(message string) {
	fmt.Println(Red(message))
	os.Exit(1)
}

func Success(message string, args ...interface{}) {
	fmt.Println(Green(fmt.Sprintf(message, args...)))
}

func DeprecationWarning(message string, args ...string) {
	fmt.Println(HiYellow("[deprecated] "), message, args)
}

var AnsiFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "color",
			Type: cty.String,
		},
		{
			Name: "message",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		color := args[0].AsString()
		message := args[1].AsString()
		return cty.StringVal(Color(color, message)), nil
	},
})

var StripAnsiFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "message",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		message := args[0].AsString()
		return cty.StringVal(stripansi.Strip(message)), nil
	},
})

func Color(color string, message string) string {
	switch color {
	case "green":
		return Green(message)
	case "red":
		return Red(message)
	case "blue":
		return Blue(message)
	case "yellow":
		return Yellow(message)
	case "bold":
		return Bold(message)
	case "italic":
		return Italic(message)
	case "cyan":
		return HiCyan(message)
	case "grey":
		return Grey(message)
	case "hi-yellow":
		return HiYellow(message)
	default:
		return message
	}
}
