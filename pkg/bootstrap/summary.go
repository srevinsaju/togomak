package bootstrap

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/ui"
	"strings"
)

const (
	bullet     = "•"
	space      = " "
	bottomLeft = "└"
	horizontal = "─"
)

var (
	subtle = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	bold   = lipgloss.NewStyle().Bold(true)
	list   = lipgloss.NewStyle().
		MarginRight(2)

	listHeader = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(subtle).
			MarginRight(2).
			Render

	success   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	fail      = lipgloss.AdaptiveColor{Light: "#FF5F5F", Dark: "#FF5F5F"}
	checkMark = lipgloss.NewStyle().SetString("✓").
			Foreground(success).
			PaddingRight(1).
			String()
	crossMark = lipgloss.NewStyle().SetString("✗").
			Foreground(fail).
			PaddingRight(1).
			String()

	ciPassed = func(s string) string {
		return checkMark + lipgloss.NewStyle().
			Render(s)
	}
	ciFailed = func(s string) string {
		return crossMark + lipgloss.NewStyle().
			Render(s)
	}
	docStyle = lipgloss.NewStyle().Padding(1, 2, 1, 2)
)

func Summary(ctx *context.Context) {

	stageStatuses := ctx.Data["stage"].(map[string]interface{})

	doc := strings.Builder{}
	var mainArr []string
	for i, layers := range ctx.Graph.TopoSortedLayers() {

		if i == 0 {
			continue
		}
		var arr []string
		arr = append(arr, bold.Render(fmt.Sprintf("Pass %d", i)))
		for _, layer := range layers {
			v, ok := stageStatuses[layer]
			if !ok {
				ctx.Logger.Tracef("Trying to find key for summary: %s", layer)
				panic("stage not found for stats check")
			}
			stage := v.(map[string]interface{})
			status := stage["status"].(map[string]interface{})
			var renderedText string
			if status["success"].(bool) {
				renderedText = strings.Repeat("  ", i) + bold.Render(bottomLeft+horizontal) + bullet + space + ciPassed(layer)

			} else {
				renderedText = strings.Repeat("  ", i) + bold.Render(bottomLeft+horizontal) + bullet + space + ciFailed(layer)
			}

			deps := ctx.Graph.Dependencies(layer)
			if len(deps) > 0 {

				for k, _ := range deps {
					if k == "root" {
						continue
					}
					renderedText += ui.Grey(" -> " + k)
				}
			}

			arr = append(arr, renderedText)

		}
		mainArr = append(mainArr, list.Render(lipgloss.JoinVertical(lipgloss.Top, arr...)))
	}

	//fmt.Println(arr)

	lists := lipgloss.JoinVertical(lipgloss.Top, append([]string{listHeader("Summary")}, mainArr...)...)
	doc.WriteString(lists)
	fmt.Println(docStyle.Render(doc.String()))

}
