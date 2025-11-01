package ai

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Dermato struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

func NewDermato(apiKey string) (*Dermato, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create GenAI client: %w", err)
	}
	return &Dermato{client: client}, nil
}

func (d *Dermato) Configure(instruction string, temp, topP float32, topK, maxTokens int32) {
	m := d.client.GenerativeModel("gemini-2.5-flash-lite")
	m.SetTemperature(temp)
	m.SetTopP(topP)
	m.SetTopK(topK)
	m.SetMaxOutputTokens(maxTokens)
	m.SystemInstruction = genai.NewUserContent(genai.Text(instruction))
	d.model = m
}

func (d *Dermato) GenerateResponse(ctx context.Context, req string) (string, error) {
	if d.model == nil {
		return "", fmt.Errorf("model not configured. Call Configure() first")
	}

	resp, err := d.model.GenerateContent(ctx, genai.Text(req))
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	text, err := extractFirstCandidateText(resp)
	if err != nil {
		return "", err
	}

	return text, nil
}

func (d *Dermato) GenerateImageResponse(ctx context.Context, imageData []byte, prompt string) (string, error) {
	if d.model == nil {
		return "", fmt.Errorf("model not configured. Call Configure() first")
	}
	if len(imageData) == 0 {
		return "", fmt.Errorf("no image data provided")
	}

	ct := http.DetectContentType(imageData) // "image/jpeg"
parts := strings.Split(ct, "/")
format := parts[len(parts)-1]           // "jpeg"

imagePart := genai.ImageData(format, imageData)

	resp, err := d.model.GenerateContent(ctx, imagePart, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate image response: %w", err)
	}

	text, err := extractFirstCandidateText(resp)
	if err != nil {
		return "", err
	}

	return text, nil
}



// helper to keep both methods consistent
func extractFirstCandidateText(resp *genai.GenerateContentResponse) (string, error) {
	if resp == nil {
		return "", fmt.Errorf("empty response from model")
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no response candidates generated")
	}

	cand := resp.Candidates[0]
	if cand == nil || cand.Content == nil {
		return "", fmt.Errorf("response candidate has no content")
	}

	var b strings.Builder
	for _, part := range cand.Content.Parts {
		switch v := part.(type) {
		case genai.Text:
			b.WriteString(string(v))
		// you can add other cases here if you want to support more types
		}
	}

	out := strings.TrimSpace(b.String())
	if out == "" {
		return "", fmt.Errorf("response candidate has no text parts")
	}

	return out, nil
}
