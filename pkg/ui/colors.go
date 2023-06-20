package ui

import (
	"fmt"
	"github.com/fatih/color"
	"os"
)

var Green = color.New(color.FgGreen).SprintFunc()
var HiCyan = color.New(color.FgHiCyan).SprintFunc()
var Red = color.New(color.FgRed).SprintFunc()
var Bold = color.New(color.Bold).SprintFunc()
var Blue = color.New(color.FgBlue).SprintFunc()
var Grey = color.New(color.FgHiBlack).SprintFunc()
var Yellow = color.New(color.FgYellow).SprintFunc()
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
