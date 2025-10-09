package models

type AIRequest struct {
	Model          string    `json:"model"`
	Messages       []Message `json:"messages"`
	Temperature    float64   `json:"temperature"`
	ResponseFormat *string   `json:"response_format"`
}

type Message struct {
	Role        string        `json:"role"`
	Content     string        `json:"content"`
	Refusal     *string       `json:"refusal"`
	Annotations []interface{} `json:"annotations"`
}

// AIResponse là cấu trúc dữ liệu trả về từ AI service
type AIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created uint64   `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
	Error   *Error   `json:"error,omitempty"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	Logprobs     *string `json:"logprobs"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens        int `json:"prompt_tokens"`
	CompletionTokens    int `json:"completion_tokens"`
	TotalTokens         int `json:"total_tokens"`
	PromptTokensDetails struct {
		CachedTokens int `json:"cached_tokens"`
		AudioTokens  int `json:"audio_tokens"`
	} `json:"prompt_tokens_details"`
	CompletionTokensDetails struct {
		ReasoningTokens          int `json:"reasoning_tokens"`
		AudioTokens              int `json:"audio_tokens"`
		AcceptedPredictionTokens int `json:"accepted_prediction_tokens"`
		RejectedPredictionTokens int `json:"rejected_prediction_tokens"`
	} `json:"completion_tokens_details"`
}

type Error struct {
	Message string `json:"message"`
}
