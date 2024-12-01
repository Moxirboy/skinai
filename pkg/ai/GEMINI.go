package ai

import (
	"context"
	"fmt"
	"net/http"
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

func (d *Dermato) GenerateImageResponse(ctx context.Context, imageData []byte, prompt string) (string, error) {
	// Validate inputs
	if d.model == nil {
		return "", fmt.Errorf("model not configured. Call Configure() first")
	}

	if len(imageData) == 0 {
		return "", fmt.Errorf("no image data provided")
	}

	// Create image part
	imagePart := genai.ImageData(http.DetectContentType(imageData), imageData)

	// Generate content with image and optional text prompt
	resp, err := d.model.GenerateContent(ctx, imagePart, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate image response: %w", err)
	}

	// Extract response text
	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no response candidates generated for image")
	}

	var responseText string
	for _, part := range resp.Candidates[0].Content.Parts {
		switch v := part.(type) {
		case genai.Text:
			responseText += string(v)
		}
	}

	if responseText == "" {
		return "", fmt.Errorf("generated image response is empty")
	}

	return responseText, nil
}
