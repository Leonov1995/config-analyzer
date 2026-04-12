package cli

import (
	"config-analyzer/internal/model"
	"fmt"

	"github.com/fatih/color"
)

func pluralizeIssue(n int) string {
	return pluralSuffix(n, "а", "ы", "")
}

func pluralizeFile(n int) string {
	return pluralSuffix(n, "", "а", "ов")
}

func pluralSuffix(n int, one, few, many string) string {
	abs := n
	if abs < 0 {
		abs = -abs
	}
	rem := abs % 100
	if rem >= 11 && rem <= 19 {
		return many
	}
	rem = abs % 10
	switch rem {
	case 1:
		return one
	case 2, 3, 4:
		return few
	default:
		return many
	}
}

var (
	green   = color.New(color.FgGreen, color.Bold)
	red     = color.New(color.FgRed, color.Bold)
	yellow  = color.New(color.FgYellow, color.Bold)
	magenta = color.New(color.FgMagenta, color.Bold)
	cyan    = color.New(color.FgCyan, color.Bold)
	bold    = color.New(color.Bold)
)

func printResult(reports []model.Report, summary *model.Summary) {
	for i, r := range reports {
		if i > 0 {
			fmt.Println()
		}

		printFile(r)
	}

	fmt.Println()
	printSummary(summary)
}

func printFile(r model.Report) {
	switch {
	case r.Err != nil:
		yellow.Printf("⚠ %s ", r.File)
		fmt.Printf("(ошибка парсинга: %v)\n", r.Err)

	case len(r.Issues) > 0:
		red.Printf("✖ %s ", r.File)
		fmt.Printf("(%d проблем%s)\n", len(r.Issues), pluralizeIssue(len(r.Issues)))
		printGroupedIssues(r.Issues)

	default:
		green.Printf("✔ %s (чисто)\n", r.File)
	}
}

func printGroupedIssues(issues []model.Issue) {
	grouped := map[model.Severity][]model.Issue{}

	for _, i := range issues {
		grouped[i.Severity] = append(grouped[i.Severity], i)
	}

	order := []model.Severity{
		model.HIGH,
		model.MEDIUM,
		model.LOW,
	}

	colors := map[model.Severity]*color.Color{
		model.HIGH:   red,
		model.MEDIUM: yellow,
		model.LOW:    magenta,
	}

	for _, sev := range order {
		items := grouped[sev]
		if len(items) == 0 {
			continue
		}

		c := colors[sev]
		c.Printf("  %s ", sev)
		fmt.Printf("(%d)\n", len(items))

		for _, i := range items {
			cyan.Printf("    - %s", i.Path)
			fmt.Printf(": %s\n", i.Message)

			if i.Recommendation != "" {
				fmt.Printf("      → %s\n", i.Recommendation)
			}
		}
	}
}

func printSummary(s *model.Summary) {
	if s.TotalIssues == 0 {
		green.Println("✔ все конфиги чистые")
		return
	}

	bold.Print("Итого: ")
	fmt.Printf(
		"%d файл%s, %d проблем%s\n",
		s.FilesWithIssues, pluralizeFile(s.FilesWithIssues),
		s.TotalIssues, pluralizeIssue(s.TotalIssues),
	)
}
