package email

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/resend/resend-go/v3"
)

type mockResendClient struct {
	lastParams  *resend.SendEmailRequest
	errToReturn error
}

func (m *mockResendClient) SendWithContext(ctx context.Context, params *resend.SendEmailRequest) (*resend.SendEmailResponse, error) {
	m.lastParams = params
	if m.errToReturn != nil {
		return nil, m.errToReturn
	}
	return &resend.SendEmailResponse{Id: "msg_123456"}, nil
}

func TestMailService_SendMail(t *testing.T) {
	tests := []struct {
		TestName        string
		TestSenderName  string
		TestEmail       string
		TestSubject     string
		TestCompany     string
		TestMessage     string
		MockErr         error
		WantErr         bool
		WantErrContains string
		WantHTMLContain []string
	}{
		{
			TestName:       "success and HTML escaping",
			TestSenderName: "<script>alert(1)</script>",
			TestEmail:      "test@example.com",
			TestSubject:    "TestSubject",
			TestCompany:    "TestCompany",
			TestMessage:    "This is a test message with &. Have fun.",
			WantHTMLContain: []string{
				"&lt;script&gt;alert(1)&lt;/script&gt;",
				"This is a test message with &amp;. Have fun.",
			},
		},
		{
			TestName:        "resend api error is wrapped",
			TestSenderName:  "TestName",
			TestEmail:       "test@example.com",
			TestSubject:     "TestSubject",
			TestCompany:     "TestCompany",
			TestMessage:     "TestMessage",
			MockErr:         errors.New("unauthorized api key"),
			WantErr:         true,
			WantErrContains: "resend api error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			mockClient := &mockResendClient{errToReturn: tc.MockErr}
			service := &MailService{
				client:    mockClient,
				emailFrom: "from@example.com",
				emailTo:   "to@example.com",
			}

			err := service.SendMail(t.Context(), tc.TestSenderName, tc.TestEmail, tc.TestSubject, tc.TestCompany, tc.TestMessage)

			if tc.WantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tc.WantErrContains) {
					t.Errorf("expected error to contain %q, got %q", tc.WantErrContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("expected success, got error: %v", err)
			}
			if mockClient.lastParams == nil {
				t.Fatal("expected SendWithContext to be called")
			}
			if mockClient.lastParams.From != "from@example.com" {
				t.Errorf("expected From %q, got %q", "from@example.com", mockClient.lastParams.From)
			}
			if mockClient.lastParams.To[0] != "to@example.com" {
				t.Errorf("expected To %q, got %q", "to@example.com", mockClient.lastParams.To[0])
			}
			for _, want := range tc.WantHTMLContain {
				if !strings.Contains(mockClient.lastParams.Html, want) {
					t.Errorf("expected HTML body to contain %q, got: %s", want, mockClient.lastParams.Html)
				}
			}
		})
	}
}
