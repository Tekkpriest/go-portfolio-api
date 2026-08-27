package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/mail"
	"strings"
)

const (
	maxNameLength    = 100
	maxEmailLength   = 254
	maxSubjectLength = 100
	maxCompanyLength = 100
	maxMessageLength = 5000
)

type ContactForm struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Company string `json:"company"`
	Message string `json:"message"`
}

type Mailer interface {
	SendMail(ctx context.Context, name, email, subject, company, message string) error
}

type ContactHandler struct {
	mailService Mailer
}

func NewContactHandler(m Mailer) *ContactHandler {
	return &ContactHandler{mailService: m}
}

func (h *ContactHandler) PostContact(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 65536)
	defer r.Body.Close()

	var form ContactForm

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&form); err != nil {
		WriteErrorResponse(w, r, "invalid json", err, http.StatusBadRequest)
		return
	}

	normalizeContactForm(&form)

	if form.Name == "" || form.Email == "" || form.Subject == "" || form.Message == "" {
		WriteErrorResponse(w, r, "required field(s) empty", nil, http.StatusBadRequest)
		return
	}

	if len(form.Name) > maxNameLength || len(form.Email) > maxEmailLength || len(form.Subject) > maxSubjectLength ||
		len(form.Company) > maxCompanyLength || len(form.Message) > maxMessageLength {
		WriteErrorResponse(w, r, "one or more fields have too many characters", nil, http.StatusBadRequest)
		return
	}

	if _, err := mail.ParseAddress(form.Email); err != nil {
		WriteErrorResponse(w, r, "invalid email address", err, http.StatusBadRequest)
		return
	}

	if err := h.mailService.SendMail(r.Context(), form.Name, form.Email, form.Subject, form.Company, form.Message); err != nil {
		WriteErrorResponse(w, r, "there was an error sending the email", err, http.StatusInternalServerError)
		return
	}

	WriteJSONResponse(w, map[string]bool{"ok": true}, http.StatusOK)
}

func normalizeContactForm(form *ContactForm) {
	form.Name = strings.TrimSpace(form.Name)
	form.Email = strings.TrimSpace(form.Email)
	form.Subject = strings.TrimSpace(form.Subject)
	form.Company = strings.TrimSpace(form.Company)
	form.Message = strings.TrimSpace(form.Message)

	form.Email = strings.ToLower(form.Email)
	form.Subject = strings.NewReplacer("\r", "", "\n", "").Replace(form.Subject)
}
