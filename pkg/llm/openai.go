package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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

	content := cleanJSONResponse(resp.Choices[0].Message.Content)

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

func (c *OpenAIClient) Summarize(articles []SummaryInput) (*SummaryResult, error) {
	var sb strings.Builder
	for i, a := range articles {
		sb.WriteString(fmt.Sprintf("%d. Headline: %s\nSummary: %s\n\n", i+1, a.Headline, a.Detail))
	}

	resp, err := c.client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Model: c.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(summarySystemPrompt),
			openai.UserMessage(sb.String()),
		},
	})

	if err != nil {
		return nil, fmt.Errorf("openai API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from openai")
	}

	content := cleanJSONResponse(resp.Choices[0].Message.Content)

	var parsed struct {
		Paragraph string   `json:"paragraph"`
		Bullets   []string `json:"bullets"`
	}

	err = json.Unmarshal([]byte(content), &parsed)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w, content: %s", err, content)
	}

	return &SummaryResult{
		Paragraph: parsed.Paragraph,
		Bullets:   parsed.Bullets,
		ModelUsed: c.modelName,
	}, nil
}

func (c *OpenAIClient) ClusterAndSummarize(articles []SummaryInput) (*ClusterSummaryResult, error) {
	// Pass 1: Cluster & Rank
	userPrompt := formatArticlesForClustering(articles)

	resp, err := c.client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT4_1Mini,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(clusterRankPrompt),
			openai.UserMessage(userPrompt),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("openai cluster pass error: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from openai (cluster pass)")
	}

	content := cleanJSONResponse(resp.Choices[0].Message.Content)

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
		story, err := c.synthesizeCluster(clusterArticles)
		if err != nil {
			return nil, fmt.Errorf("openai synthesis error for cluster %q: %w", cluster.Topic, err)
		}
		stories = append(stories, *story)
	}

	return &ClusterSummaryResult{
		Stories:   stories,
		ModelUsed: "gpt-4.1-mini",
	}, nil
}

func (c *OpenAIClient) synthesizeCluster(articles []SummaryInput) (*StorySummary, error) {
	userPrompt := formatArticlesForSynthesis(articles)

	resp, err := c.client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT4_1Mini,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(synthesizePrompt),
			openai.UserMessage(userPrompt),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("openai API error: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from openai")
	}

	content := cleanJSONResponse(resp.Choices[0].Message.Content)

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
