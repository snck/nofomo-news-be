package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type AnthropicClient struct {
	client    *anthropic.Client
	model     anthropic.Model
	modelName string
}

func NewAnthropicClient(apiKey string) *AnthropicClient {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &AnthropicClient{
		client:    &client,
		model:     anthropic.ModelClaudeHaiku4_5,
		modelName: "claude-4.5-haiku",
	}
}

func (c *AnthropicClient) Transform(input TransformInput) (*TransformResult, error) {
	userPrompt := fmt.Sprintf("Headline: %s\nSummary: %s", input.Headline, input.Detail)

	resp, err := c.client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     c.model,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
		},
	})

	if err != nil {
		return nil, fmt.Errorf("anthropic API error: %w", err)
	}

	if len(resp.Content) == 0 {
		return nil, fmt.Errorf("no response from anthropic")
	}

	content := resp.Content[0].Text
	content = cleanJSONResponse(content)

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

func cleanJSONResponse(content string) string {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	// Some model responses include extra prose around JSON.
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end > start {
		content = content[start : end+1]
	}
	return content
}
