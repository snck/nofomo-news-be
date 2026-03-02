package llm

import (
	"fmt"
	"strings"
)

const maxDetailChars = 200

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func formatArticlesForClustering(articles []SummaryInput) string {
	var sb strings.Builder
	for i, a := range articles {
		sb.WriteString(fmt.Sprintf("[%d] Headline: %s\n", i, a.Headline))
		sb.WriteString(fmt.Sprintf("    Summary: %s\n", truncate(a.Detail, maxDetailChars)))
		sb.WriteString(fmt.Sprintf("    Publisher: %s\n", a.Publisher))
		sb.WriteString(fmt.Sprintf("    Published: %s\n", a.PublishedAt.Format("2006-01-02 15:04")))
		if len(a.Symbols) > 0 {
			sb.WriteString(fmt.Sprintf("    Symbols: %s\n", strings.Join(a.Symbols, ", ")))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func formatArticlesForSynthesis(articles []SummaryInput) string {
	var sb strings.Builder
	for i, a := range articles {
		sb.WriteString(fmt.Sprintf("[%d] Headline: %s\n", i, a.Headline))
		sb.WriteString(fmt.Sprintf("    Summary: %s\n", a.Detail))
		sb.WriteString(fmt.Sprintf("    Publisher: %s\n", a.Publisher))
		sb.WriteString(fmt.Sprintf("    Published: %s\n", a.PublishedAt.Format("2006-01-02 15:04")))
		if len(a.Symbols) > 0 {
			sb.WriteString(fmt.Sprintf("    Symbols: %s\n", strings.Join(a.Symbols, ", ")))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func gatherClusterArticles(allArticles []SummaryInput, indices []int) []SummaryInput {
	var result []SummaryInput
	for _, idx := range indices {
		if idx >= 0 && idx < len(allArticles) {
			result = append(result, allArticles[idx])
		}
	}
	return result
}
