package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockMailer struct {
	wasCalled bool
	mockErr   error
	got       ContactForm
}

func (m *mockMailer) SendMail(ctx context.Context, name, email, subject, company, message string) error {
	m.wasCalled = true
	m.got = ContactForm{
		Name:    name,
		Email:   email,
		Subject: subject,
		Company: company,
		Message: message,
	}
	return m.mockErr
}

func TestContactHandler_PostContact(t *testing.T) {
	tests := []struct {
		TestName         string
		Body             string
		Form             *ContactForm
		MailerErr        error
		WantStatusCode   int
		WantMailerCalled bool
		WantErrMessage   string
	}{
		{
			TestName: "valid form is accepted and forwarded to mailer",
			Form: &ContactForm{
				Name:    "TestName",
				Email:   "test@example.com",
				Subject: "TestSubject",
				Company: "TestCompany",
				Message: "Hello World!",
			},
			WantStatusCode:   http.StatusOK,
			WantMailerCalled: true,
		},
		{
			TestName: "invalid email address is rejected",
			Form: &ContactForm{
				Name:    "TestName",
				Email:   "invalid-mail",
				Subject: "TestSubject",
				Company: "TestCompany",
				Message: "Hello Invalid World!",
			},
			WantStatusCode: http.StatusBadRequest,
			WantErrMessage: "invalid email address",
		},
		{
			TestName: "missing required field is rejected",
			Form: &ContactForm{
				Email:   "test@example.com",
				Subject: "TestSubject",
				Message: "Hello!",
			},
			WantStatusCode: http.StatusBadRequest,
			WantErrMessage: "required field(s) empty",
		},
		{
			TestName: "mailer error results in 500",
			Form: &ContactForm{
				Name:    "TestName",
				Email:   "test@example.com",
				Subject: "TestSubject",
				Company: "TestCompany",
				Message: "MailServiceError",
			},
			MailerErr:        errors.New("resend API is down"),
			WantStatusCode:   http.StatusInternalServerError,
			WantMailerCalled: true,
			WantErrMessage:   "there was an error sending the email",
		},
		{
			TestName:       "malformed JSON body is rejected",
			Body:           "{invalid-json",
			WantStatusCode: http.StatusBadRequest,
			WantErrMessage: "invalid json",
		},
	}

	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockMail := &mockMailer{mockErr: tc.MailerErr}
			handler := NewContactHandler(mockMail)

			body := []byte(tc.Body)
			if tc.Form != nil {
				body, _ = json.Marshal(tc.Form)
			}

			req := httptest.NewRequestWithContext(t.Context(), "POST", "/api/contact", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			handler.PostContact(rec, req)

			if rec.Code != tc.WantStatusCode {
				t.Errorf("expected status %d, got %d", tc.WantStatusCode, rec.Code)
			}
			if mockMail.wasCalled != tc.WantMailerCalled {
				t.Errorf("expected mailer called=%v, got %v", tc.WantMailerCalled, mockMail.wasCalled)
			}

			if tc.WantStatusCode == http.StatusOK {
				if mockMail.got != *tc.Form {
					t.Errorf("expected mailer to receive %+v, got %+v", *tc.Form, mockMail.got)
				}
				var resp map[string]bool
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode success body: %v", err)
				}
				if !resp["ok"] {
					t.Errorf("expected json response {'ok':true}, got %v", resp)
				}
				return
			}

			var errResp APIError
			if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
				t.Fatalf("failed to decode error response body: %v", err)
			}
			if errResp.Error != tc.WantErrMessage {
				t.Errorf("expected error message %q, got %q", tc.WantErrMessage, errResp.Error)
			}
		})
	}
}

func TestNormalizeContactForm(t *testing.T) {
	tests := []struct {
		TestName string
		Input    ContactForm
		Want     ContactForm
	}{
		{
			TestName: "trims whitespace in all fields",
			Input:    ContactForm{Name: "  TestName  ", Email: " test@example.com ", Subject: " TestSubject ", Company: " TestCompany ", Message: " Hello "},
			Want:     ContactForm{Name: "TestName", Email: "test@example.com", Subject: "TestSubject", Company: "TestCompany", Message: "Hello"},
		},
		{
			TestName: "lowercases email",
			Input:    ContactForm{Email: "Test@Example.COM"},
			Want:     ContactForm{Email: "test@example.com"},
		},
		{
			TestName: "strips CRLF from subject",
			Input:    ContactForm{Subject: "Hello\r\nBcc: evil@example.com"},
			Want:     ContactForm{Subject: "HelloBcc: evil@example.com"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			form := tc.Input
			normalizeContactForm(&form)
			if form != tc.Want {
				t.Errorf("got %+v, want %+v", form, tc.Want)
			}
		})
	}
}
