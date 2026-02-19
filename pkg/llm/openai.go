package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const promptVersion = "v1"
const systemPrompt = `You are a financial news editor. Your job is to rewrite news headlines and summaries in a neutral, calm tone.

Rules:
1. Remove urgency words (BREAKING, NOW, ALERT, JUST IN)
2. Remove ALL CAPS
3. Replace emotional verbs:
   - crash, plummet, tank → dropped, decreased
   - explode, soar, skyrocket → rose, increased
4. Remove judgmental words (smart, dumb, crazy, shocking, terrifying)
5. Add uncertainty to predictions (will → may, could, might)
6. Remove dramatic metaphors (bloodbath, shockwave, chaos)
7. Keep all facts: numbers, names, dates, percentages

Output as JSON only, no other text:
{
  "headline": "transformed headline",
  "summary": "transformed summary",
  "category": "one of: Earnings, Market Movement, Economy, Crypto, Mergers & Acquisitions, Policy & Regulation, Company News, Analysis",
  "sentiment_score": 1-10 how emotional was the original (10 = very emotional)
}`

type OpenAIClient struct {
	client    *openai.Client
	model     openai.ChatModel
	modelName string
}

func NewOpenAIClient(apiKey string) *OpenAIClient {
	client := openai.NewClient(option.WithAPIKey(apiKey))
	return &OpenAIClient{
		client:    &client,
		model:     openai.ChatModelGPT4oMini,
		modelName: "gpt-4o-mini",
	}
}

func (c *OpenAIClient) Transform(input TransformInput) (*TransformResult, error) {
	userPrompt := fmt.Sprintf("Headline: %s\nSummary: %s", input.Headline, input.Detail)

	resp, err := c.client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Model: c.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		},
	})

	if err != nil {
		return nil, fmt.Errorf("openai API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from openai")
	}

	content := resp.Choices[0].Message.Content

	var parsed struct {
		Headline       string `json:"headline"`
		Summary        string `json:"summary"`
		Category       string `json:"category"`
		SentimentScore int    `json:"sentiment_score"`
	}

	err = json.Unmarshal([]byte(content), &parsed)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w, content: %s", err, content)
	}

	return &TransformResult{
		Headline:       parsed.Headline,
		Detail:         parsed.Summary,
		Category:       parsed.Category,
		SentimentScore: parsed.SentimentScore,
		PromptVersion:  promptVersion,
		ModelUsed:      c.modelName,
	}, nil
}
