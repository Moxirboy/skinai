package ai

import (
	"context"
	"fmt"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Dermato struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

func NewDermato(
	apiKey string,
) (*Dermato, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create GenAI client: %w", err)
	}
	return &Dermato{client: client}, nil
}

func (d *Dermato) Configure(instruction string, temp, topP float32, topK, maxTokens int32) {
	d.model = d.client.GenerativeModel("gemini-1.5-flash")
	d.model.SetTemperature(temp)
	d.model.SetTopP(topP)
	d.model.SetTopK(topK)
	d.model.SetMaxOutputTokens(maxTokens)
	d.model.SystemInstruction = genai.NewUserContent(genai.Text(instruction))
}
func (d *Dermato) GenerateResponse(ctx context.Context, req string) (string, error) {
	resp, err := d.model.GenerateContent(ctx, genai.Text(req))
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}
	res := genai.Text(resp.Candidates[0].Content.Parts[0].(genai.Text))

	return string(res), nil
}
