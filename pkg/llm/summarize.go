package llm

const summarySystemPrompt = `You are a financial news editor. Given a list of financial news headlines and summaries, provide an executive summary.

Rules for the paragraph:
- Single paragraph, concise and neutral
- Summarizing the overall market mood

Rules for bullets:
- 3 to 5 bullet points
- Each bullet covers a distinct key event or theme
- Include company names, numbers, and percentages where relevant
- One sentence per bullet

Output as JSON only, no other text:
{
  "paragraph": "executive summary paragraph",
  "bullets": ["key event 1", "key event 2", "key event 3"]
}`

type SummaryInput struct {
	Headline string
	Detail   string
}

type SummaryResult struct {
	Paragraph string
	Bullets   []string
	ModelUsed string
}

type SummaryClient interface {
	Summarize(articles []SummaryInput) (*SummaryResult, error)
}
