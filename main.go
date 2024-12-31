package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"a21hc3NpZ25tZW50/service"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

// Initialize the services
var fileService = &service.FileService{}
var aiService = &service.AIService{Client: &http.Client{}}
var store = sessions.NewCookieStore([]byte("my-key"))

func getSession(r *http.Request) *sessions.Session {
	session, _ := store.Get(r, "chat-session")
	return session
}

func main() {
	// Load the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Retrieve the Hugging Face token from the environment variables
	token := os.Getenv("HUGGINGFACE_TOKEN")
	if token == "" {
		log.Fatal("HUGGINGFACE_TOKEN is not set in the .env file")
	}

	// Set up the router
	router := mux.NewRouter()

	// File upload endpoint
	router.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		// err := r.ParseMultipartForm(10 << 20) // max 10 mb
		// if err != nil {
		// 	http.Error(w, "Unable to parse form", http.StatusBadRequest)
		// 	return
		// }

		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Unable to retrieve file from form", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Read the file content
		buf, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Unable to read file", http.StatusInternalServerError)
			return
		}

		// Process the file
		data, err := fileService.ProcessFile(string(buf))
		if err != nil {
			http.Error(w, "Unable to process file", http.StatusInternalServerError)
			return
		}

		// get the question from form data
		query := r.FormValue("in_query")
		if query == "" {
			http.Error(w, "Question is required!", http.StatusBadRequest)
			return
		}

		// Analyze the data
		result, err := aiService.AnalyzeData(data, query, token)
		if err != nil {
			http.Error(w, "Unable to analyze data", http.StatusInternalServerError)
			return
		}

		// Respond with the result
		response := map[string]string{
			"status": "success",
			"answer": result,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}).Methods("POST")

	// Chat endpoint
	router.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		inputs := r.FormValue("query")
		if inputs == "" {
			http.Error(w, "Ask some question...", http.StatusBadRequest)
		}

		// err := json.NewDecoder(r.Body).Decode(&inputs)
		// if err != nil {
		// 	http.Error(w, "Unable to parse request body", http.StatusBadRequest)
		// 	return
		// }

		// Get the session
		session := getSession(r)
		context, ok := session.Values["context"].(string)
		if !ok {
			context = ""
		}

		// Chat with the AI service
		result, err := aiService.ChatWithAI(context, inputs, token)
		if err != nil {
			http.Error(w, "Unable to chat with AI", http.StatusInternalServerError)
			return
		}

		// Update the session context
		session.Values["context"] = context + "\n" + result.GeneratedText
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, "Unable to save session", http.StatusInternalServerError)
			return
		}

		// Respond with the result
		response := map[string]string{
			"status": "success",
			"answer": result.GeneratedText,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}).Methods("POST")

	// Enable CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"}, // Allow your React app's origin
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	}).Handler(router)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, corsHandler))
}
