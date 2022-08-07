package ui

import (
	"github.com/fatih/color"
)

var Green = color.New(color.FgGreen).SprintFunc()
var Red = color.New(color.FgRed).SprintFunc()
var Grey = color.New(color.FgHiBlack).SprintFunc()
var Yellow = color.New(color.FgYellow).SprintFunc()
var Plus = color.New(color.FgHiWhite).SprintFunc()("+")
var SubStage = Grey("==>")
var SubSubStage = Grey("-->")
var Matrix = Yellow("matrix")
var Stage = Yellow("stage")
var Options = Grey("options")
var FailLazy = Grey("fail-lazy")

var True = Green("true")
var False = Red("false")
