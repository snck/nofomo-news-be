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

func (c *AnthropicClient) Summarize(articles []SummaryInput) (*SummaryResult, error) {
	var sb strings.Builder
	for i, a := range articles {
		sb.WriteString(fmt.Sprintf("%d. Headline: %s\nSummary: %s\n\n", i+1, a.Headline, a.Detail))
	}

	resp, err := c.client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     c.model,
		MaxTokens: 2048,
		System: []anthropic.TextBlockParam{
			{Text: summarySystemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(sb.String())),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("anthropic API error: %w", err)
	}
	if len(resp.Content) == 0 {
		return nil, fmt.Errorf("no response from anthropic")
	}

	content := cleanJSONResponse(resp.Content[0].Text)

	var parsed struct {
		Paragraph string   `json:"paragraph"`
		Bullets   []string `json:"bullets"`
	}
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w, content: %s", err, content)
	}

	return &SummaryResult{
		Paragraph: parsed.Paragraph,
		Bullets:   parsed.Bullets,
		ModelUsed: c.modelName,
	}, nil
}

func (c *AnthropicClient) ClusterAndSummarize(articles []SummaryInput) (*ClusterSummaryResult, error) {
	clusterModel := anthropic.ModelClaudeSonnet4_6
	clusterModelName := "claude-sonnet-4-6"

	// Pass 1: Cluster & Rank
	userPrompt := formatArticlesForClustering(articles)

	resp, err := c.client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     clusterModel,
		MaxTokens: 4096,
		System: []anthropic.TextBlockParam{
			{Text: clusterRankPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("anthropic cluster pass error: %w", err)
	}
	if len(resp.Content) == 0 {
		return nil, fmt.Errorf("no response from anthropic (cluster pass)")
	}

	content := cleanJSONResponse(resp.Content[0].Text)

	var clusterResult struct {
		Clusters []struct {
			Topic            string `json:"topic"`
			ArticleIndices   []int  `json:"article_indices"`
			ImportanceReason string `json:"importance_reason"`
		} `json:"clusters"`
	}
	if err := json.Unmarshal([]byte(content), &clusterResult); err != nil {
		return nil, fmt.Errorf("failed to parse cluster response: %w, content: %s", err, content)
	}

	// Pass 2: Synthesize each cluster
	var stories []StorySummary
	for _, cluster := range clusterResult.Clusters {
		clusterArticles := gatherClusterArticles(articles, cluster.ArticleIndices)
		story, err := c.synthesizeCluster(clusterArticles, clusterModel)
		if err != nil {
			return nil, fmt.Errorf("anthropic synthesis error for cluster %q: %w", cluster.Topic, err)
		}
		stories = append(stories, *story)
	}

	return &ClusterSummaryResult{
		Stories:   stories,
		ModelUsed: clusterModelName,
	}, nil
}

func (c *AnthropicClient) synthesizeCluster(articles []SummaryInput, model anthropic.Model) (*StorySummary, error) {
	userPrompt := formatArticlesForSynthesis(articles)

	resp, err := c.client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     model,
		MaxTokens: 2048,
		System: []anthropic.TextBlockParam{
			{Text: synthesizePrompt},
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

	content := cleanJSONResponse(resp.Content[0].Text)

	var parsed struct {
		Stories []StorySummary `json:"stories"`
	}
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse synthesis response: %w, content: %s", err, content)
	}
	if len(parsed.Stories) == 0 {
		return nil, fmt.Errorf("no stories in synthesis response")
	}

	return &parsed.Stories[0], nil
}
