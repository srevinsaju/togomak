package ui

import "github.com/AlecAivazis/survey/v2"

func PromptYesNo(text string) bool {
	// show a prompt to the user using survey
	// return true if the user says yes

	var result bool
	err := survey.AskOne(&survey.Confirm{
		Message: text,
	}, &result)
	if err != nil {
		panic(err)
	}
	return result
}
