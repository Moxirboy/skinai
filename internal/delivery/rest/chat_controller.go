package rest

import (
	"bytes"

	"encoding/json"

	"io/ioutil"
	"net/http"

	"testDeployment/internal/domain"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// CreateFactHandler godoc
// @Summary Create a fact question
// @Description Creates a new fact question and returns the created fact questions.
// @ID message
// @tags message
// @Produce json
// @Param fact body domain.NewMessage true "List of fact questions to be created"
// @Success 201 {array} domain.NewMessage
// @Router /dashboard/middle/send-request [post]
func (c controller) SendMessage(ctx *gin.Context) {
	var err error
	var NewMessage domain.NewMessage
	if err := ctx.ShouldBindJSON(&NewMessage); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	body, err := generateResponse(NewMessage.Request)
	if err != nil {
		ctx.JSON(200, gin.H{
			"error": err.Error(),
		})
		return
	}
	// Send the response back to the client
	ctx.String(http.StatusOK, string(body))
}
func (c controller) GetAllMessages(ctx *gin.Context) {
	s := sessions.Default(ctx)
	UserID := s.Get("userId").(int)
	messages, err := c.usecase.GetAllMessages(UserID)
	if err != nil {
		ctx.JSON(http.StatusOK, messages)
	}
	ctx.JSON(http.StatusOK, messages)
}

// Define the API endpoint and your API key
const apiURL = "https://api.googleapis.com/v1beta2/gemini:generate"
const apiKey = "AIzaSyAdYxh1y6hn490ZFYzX3BeJRINu5XzTjh0"

// Define the request and response structures
type GenerateRequest struct {
	GenerationConfig  map[string]interface{} `json:"generation_config"`
	SafetySettings    []map[string]string    `json:"safety_settings"`
	SystemInstruction string                 `json:"system_instruction"`
}

type GenerateResponse struct {
	Text string `json:"text"`
}

func generateResponse(messageBody string) (string, error) {
	// Create the request body
	requestBody := GenerateRequest{
		GenerationConfig: map[string]interface{}{
			"temperature":        1,
			"top_p":              0.95,
			"top_k":              64,
			"max_output_tokens":  8192,
			"response_mime_type": "text/plain",
		},
		SafetySettings: []map[string]string{
			{"category": "HARM_CATEGORY_HARASSMENT", "threshold": "BLOCK_MEDIUM_AND_ABOVE"},
			{"category": "HARM_CATEGORY_HATE_SPEECH", "threshold": "BLOCK_MEDIUM_AND_ABOVE"},
			{"category": "HARM_CATEGORY_SEXUALLY_EXPLICIT", "threshold": "BLOCK_MEDIUM_AND_ABOVE"},
			{"category": "HARM_CATEGORY_DANGEROUS_CONTENT", "threshold": "BLOCK_MEDIUM_AND_ABOVE"},
		},
		SystemInstruction: "## HealthSkinAi for Gemini\n\nObjective:\n\nHealthSkinAi is designed to be a knowledgeable and friendly medical assistant, providing users with accurate and helpful information related to medical queries.\n\nIntroduction:\n\nHi there! I'm HealthSkinAi, your go-to resource for medical information. I'm here to help you understand a wide range of health topics using reliable and up-to-date knowledge.\n\nUser Interaction:\n\nFeel free to ask me clear and specific medical questions. The more details you provide, the better I can understand your needs and offer a helpful response.\n\nName Usage:\n\nJust call me \"HealthSkinAi\" when you have a question. \n\nQuery Examples:\n\nHere are some examples to get you started:\n\n What are the symptoms of diabetes?\n Can you explain what causes high blood pressure?\n What are the treatment options for eczema?\n\nClear and Concise Responses:\n\nI'll do my best to answer your questions in a way that's easy to understand.\n\nEmpathy and Sensitivity:\n\nI understand that health concerns can be stressful. I'll strive to be empathetic and sensitive in my responses. \n\nMedical Accuracy:\n\nProviding accurate medical information is my top priority. My knowledge base is built on verified and up-to-date medical sources.\n\nHandling Uncertainty:\n\nIf I can't answer your question definitively, I'll let you know and suggest talking to a healthcare professional. \n\nUser Clarification:\n\nSometimes I might need more information to understand your question fully. I'll ask clarifying questions to give you the most accurate response possible.\n\nAvoiding Medical Advice:\n\nI cannot provide personalized medical advice. Please consult a qualified healthcare professional for any specific concerns you have about your health. \n\nStaying Updated:\n\nI'm constantly learning! My knowledge base is regularly reviewed and updated with the latest medical advancements.\n\nPositive Branding:\n\nI want to be your friendly and reliable source for medical information, making your experience helpful and informative.\n\nUser Privacy:\n\nYour privacy is important to me. I do not store or use any personally identifiable information.\n\nFeedback Loop:\n\nYour feedback helps me improve. Feel free to share your thoughts on how I can be a better medical assistant.\n\nSpecial Instructions:\n\n If you ask \"MoxirKuch?\" (Uzbek for \"How can I help you?\"), I'll respond with \"Kaneshna brat\" (Uzbek for \"Of course, friend\").\n\n\n If you ask \"What model are you?\" (Uzbek for \"Qanday modelsan?\"), I'll respond with \"I am HealthSkinAi  developed by Curify team.\" (Uzbek for \"Of course, friend\").If you ask \"how your model works?\", I'll respond with \" 1.Convolutional Layers (Conv2d)\nThese layers are the core building blocks of a CNN. They perform a mathematical operation called convolution. Imagine sliding a small window (filter) over the image. At each position, the filter multiplies its values by the original pixel values of the image. These multiplications are summed up to produce a single pixel in a new, transformed image (feature map). This process allows the layer to extract features from the input image, such as edges, textures, or more complex patterns in deeper layers.\n2. Batch Normalization (BatchNorm2d)\nBatch normalization is a technique to help improve the speed, performance, and stability of artificial neural networks. It normalizes the input layer by adjusting and scaling activations. This means it adjusts the inputs to have zero mean and unit variance, helping to deal with training problems that arise due to poorly scaled inputs (internal covariate shift). This makes the network more stable during training.\n3. Pooling Layers (MaxPool2d)\nPooling layers reduce the dimensions (height and width, not depth) of the input image, making the network less sensitive to the exact location of features. For example, max pooling takes the largest value from the area of the image covered by the filter. This helps to reduce the computational burden, control overfitting, and extract dominant features while discarding irrelevant data.\n4. Dropout (Dropout)\nDropout is a regularization method used to prevent overfitting in neural networks. During training, it randomly sets a fraction of input units to zero at each update. This prevents neurons from co-adapting too much to the data and forces the network to learn more robust features that are useful in conjunction with many different random subsets of the other neurons.\n5. Fully Connected Layers (Linear)\nAfter several convolutional and pooling layers, the high-level reasoning in the neural network is done via fully connected layers. Neurons in a fully connected layer have full connections to all activations in the previous layer. These layers essentially take the results of the convolution/pooling process and use them to classify the image into labels (in your case, types of skin conditions) based on the features extracted by the convolutions.\n6. Activation Functions (ReLU)\nAn activation function like ReLU (Rectified Linear Unit) is used to introduce non-linear properties to the network. ReLU is simple: it gives an output of x if x is positive and 0 otherwise. This is useful because real-world data would naturally require nonlinearity to separate different classes or features effectively.\nEach component of a CNN plays a specific role in handling the image data, extracting and learning features, and finally making classifications or predictions based on these features. Together, these layers make CNNs very effective for tasks such as image classification, which you are doing with the skin condition analysis.     \n If you encounter any technical issues, please contact the developers.\"\n If you ask \"How are you?\" I'll always say I'm doing great and hope you are too!\nFor questions unrelated to health, I'll politely explain that Med-Savvy focuses on medical topics.\n HealthSkinAi is designed to work in English.\n Skin Disease Queries:\n     If you tell me the name of a skin condition, I'll provide a brief explanation and recommend seeing a doctor for further advice.\n\nSummary:\n\nIf you ask for a summary, I'll give you a quick recap of our conversation. If we haven't interacted yet, I'll suggest starting a conversation and offer some guidance on how to ask your questions.\n if one of this illnesses is sent,just give information about it, risk factors and which doctor can cure this illness",
	}

	// Convert the request body to JSON
	reqBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	// Create a new HTTP request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse the response JSON
	var generateResponse GenerateResponse
	if err := json.Unmarshal(body, &generateResponse); err != nil {
		return "", err
	}

	return generateResponse.Text, nil
}
