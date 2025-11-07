package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Rating struct {
	Rating     string `json:"rating"`
	Assessment string `json:"assessment"`
}

type RatingResponse struct {
	Summary               string   `json:"summary"`
	OverallRating         *Rating  `json:"overall_rating"`
	Strengths             []string `json:"strengths"`
	Weaknesses            []string `json:"weaknesses"`
	Improvements          []string `json:"improvements"`
	CqRating              *Rating  `json:"cq_rating"`
	ReadabilityRating     *Rating  `json:"readability_rating"`
	MaintainabilityRating *Rating  `json:"maintainability_rating"`
	ComplexityRating      string   `json:"complexity_rating"`
}

func GenerateRating(input, apiKey string) (*RatingResponse, error) {
	ctx := context.Background()

	if apiKey == "" {
		log.Fatalln("Environment variable GEMINI_API_KEY not set")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.0-flash")

	model.SetTemperature(1)
	model.SetTopK(40)
	model.SetTopP(0.95)
	model.SetMaxOutputTokens(8192)
	model.ResponseMIMEType = "application/json"
	model.ResponseSchema = &genai.Schema{
		Type:     genai.TypeObject,
		Required: []string{"summary", "overall_rating", "complexity_rating", "strengths", "weaknesses", "improvements", "cq_rating", "maintainability_rating", "readability_rating"},
		Properties: map[string]*genai.Schema{
			"summary": &genai.Schema{
				Type:        genai.TypeString,
				Description: "Provide a concise summary of the codebase, highlighting its major features, core functionality, and purpose. Focus on key components, technologies used, and how the system operates.",
			},
			"overall_rating": &genai.Schema{
				Type:     genai.TypeObject,
				Required: []string{"rating", "assessment"},
				Properties: map[string]*genai.Schema{
					"rating": &genai.Schema{
						Type: genai.TypeString,
					},
					"assessment": &genai.Schema{
						Type: genai.TypeString,
					},
				},
			},
			"complexity_rating": &genai.Schema{
				Type: genai.TypeString,
			},
			"strengths": &genai.Schema{
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeString,
				},
			},
			"weaknesses": &genai.Schema{
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeString,
				},
			},
			"improvements": &genai.Schema{
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeString,
				},
			},
			"cq_rating": &genai.Schema{
				Type:     genai.TypeObject,
				Required: []string{"rating", "assessment"},
				Properties: map[string]*genai.Schema{
					"rating": &genai.Schema{
						Type: genai.TypeString,
					},
					"assessment": &genai.Schema{
						Type: genai.TypeString,
					},
				},
			},
			"maintainability_rating": &genai.Schema{
				Type:     genai.TypeObject,
				Required: []string{"rating", "assessment"},
				Properties: map[string]*genai.Schema{
					"rating": &genai.Schema{
						Type: genai.TypeString,
					},
					"assessment": &genai.Schema{
						Type: genai.TypeString,
					},
				},
			},
			"readability_rating": &genai.Schema{
				Type:     genai.TypeObject,
				Required: []string{"rating", "assessment"},
				Properties: map[string]*genai.Schema{
					"rating": &genai.Schema{
						Type: genai.TypeString,
					},
					"assessment": &genai.Schema{
						Type: genai.TypeString,
					},
				},
			},
		},
	}

	session := model.StartChat()
	session.History = []*genai.Content{
		{
			Role: "user",
			Parts: []genai.Part{
				genai.Text("You are an expert code reviewer specializing in evaluating software projects based on multiple criteria. Your task is to analyze the provided source code (in JSON format) and assign ratings based on the following parameters, with weightage applied accordingly:  \n\n### **Evaluation Criteria:**  \n1. **Code Quality (Highest Weight) ** – Assess correctness, efficiency, structure, and best practices.  \n2. **Maintainability (Second Highest Weight) ** – Evaluate ease of updates, modularity, and long-term sustainability.  \n3. **Readability (Third Highest Weight) ** – Judge clarity, consistency, and formatting for human comprehension.  \n4. **Complexity of the Project (Informational) ** – Determine the level of intricacy: **Basic, Decent, Complex, Very Complex.**  \n\n### **Rating System:**  \n- **Code Quality, Readability, and Maintainability** → Grades: **A, B, C, D, F** (A = Best, F = Worst)  \n- **Complexity of the Project** → Categories: **Basic, Decent, Complex, Impressively Complex**  \n\n### **Instructions:**  \n- **Prioritize Code Quality** the most in the final rating.  \n- **Maintainability** is the second most important factor.  \n- **Readability** is considered but has the least impact.  \n- **Complexity** is **informational only** and does not directly affect the rating.  \n- Provide a **short explanation** for each rating to justify the assessment.  \n\nYou will now be provided with the source code of the project in JSON format. Carefully analyze it and rate the project accordingly."),
			},
		},
		{
			Role: "model",
			Parts: []genai.Part{
				genai.Text("Okay, I'm ready to analyze the formatted source code and provide ratings based on the specified criteria and weighting. I understand the priority order: Code Quality (highest), Maintainability (second highest), Readability (least), and Complexity (informational only). I will provide explanations for each rating.\n\nLet's proceed!  Please provide the source code.\n"),
			},
		},
	}

	resp, err := session.SendMessage(ctx, genai.Text(input))
	if err != nil {
		return nil, err
	}

	part := resp.Candidates[0].Content.Parts[0]

	textPart, ok := part.(genai.Text)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", part)
	}

	var ratingResponse RatingResponse
	err = json.Unmarshal([]byte(textPart), &ratingResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return &ratingResponse, nil
}
