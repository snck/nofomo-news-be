package llm

type TransformInput struct {
	Headline string
	Detail   string
}

type TransformResult struct {
	Headline       string
	Detail         string
	Category       string
	SentimentScore int
	PromptVersion  string
	ModelUsed      string
}

type LLMClient interface {
	Transform(input TransformInput) (*TransformResult, error)
}
