package handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"

	"github.com/resend/resend-go/v3"
)

type ContactForm struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Company string `json:"company"`
	Message string `json:"msg"`
}

func HandlePostContact(w http.ResponseWriter, r *http.Request) {
	emailFrom := os.Getenv("EMAIL_FROM")
	emailTo := os.Getenv("EMAIL_TO")
	if emailFrom == "" || emailTo == "" {
		log.Println("EMAIL_FROM or EMAIL_TO not found in .env")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var form ContactForm
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		http.Error(w, "JSON Encoding Error (E-Mail)", http.StatusBadRequest)
		return
	}

	if form.Name == "" || form.Email == "" || form.Subject == "" || form.Message == "" {
		http.Error(w, "Required field(s) empty", http.StatusBadRequest)
		return
	}

	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		log.Println("Resend API Key not found.")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	client := resend.NewClient(apiKey)

	htmlBody := fmt.Sprintf(`
		<p><strong>Name:</strong> %s</p>
		<p><strong>E-Mail:</strong> %s</p>
		<p><strong>Firma:</strong> %s</p>
		<p><strong>Nachricht:</strong>%s<br></p>
	`,
		html.EscapeString(form.Name),
		html.EscapeString(form.Email),
		html.EscapeString(form.Company),
		html.EscapeString(form.Message),
	)

	params := &resend.SendEmailRequest{
		From:    emailFrom,
		To:      []string{emailTo},
		Subject: html.EscapeString(form.Subject),
		Html:    htmlBody,
	}

	sent, err := client.Emails.Send(params)
	if err != nil {
		log.Printf("Resend Error: %v", err)
		http.Error(w, "There was an Error sending your E-Mail", http.StatusInternalServerError)
		return
	}

	log.Printf("E-mail sent: %s", sent.Id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	if err != nil {
		http.Error(w, "JSON Encoding Error (E-Mail):", http.StatusInternalServerError)
	}
}
