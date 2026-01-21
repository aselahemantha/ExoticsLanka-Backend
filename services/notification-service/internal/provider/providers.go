package provider

import (
	"fmt"
	"log"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
)

type EmailProvider interface {
	SendEmail(to, subject, htmlContent string) (string, error)
}

type SMSProvider interface {
	SendSMS(to, message string) (string, error)
}

// SendGrid Implementation
type SendGridProvider struct {
	client    *sendgrid.Client
	fromEmail string
	fromName  string
}

func NewSendGridProvider(apiKey, fromEmail, fromName string) *SendGridProvider {
	return &SendGridProvider{
		client:    sendgrid.NewSendClient(apiKey),
		fromEmail: fromEmail,
		fromName:  fromName,
	}
}

func (p *SendGridProvider) SendEmail(to, subject, htmlContent string) (string, error) {
	from := mail.NewEmail(p.fromName, p.fromEmail)
	if p.client == nil || p.fromEmail == "test@example.com" { // Mock mode if key invalid or test
		log.Printf("[MOCK EMAIL] To: %s, Subject: %s\n", to, subject)
		return "mock-message-id", nil
	}

	toEmail := mail.NewEmail("", to)
	message := mail.NewSingleEmail(from, subject, toEmail, " ", htmlContent) // Content passed as HTML

	response, err := p.client.Send(message)
	if err != nil {
		return "", err
	}

	if response.StatusCode >= 400 {
		return "", fmt.Errorf("sendgrid error: status %d body %s", response.StatusCode, response.Body)
	}

	return response.Headers["X-Message-Id"][0], nil
}

// Twilio Implementation
type TwilioProvider struct {
	client     *twilio.RestClient
	fromNumber string
	accountSid string
}

func NewTwilioProvider(accountSid, authToken, fromNumber string) *TwilioProvider {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSid,
		Password: authToken,
	})
	return &TwilioProvider{
		client:     client,
		fromNumber: fromNumber,
		accountSid: accountSid,
	}
}

func (p *TwilioProvider) SendSMS(to, message string) (string, error) {
	if p.accountSid == "" || p.accountSid == "test-sid" {
		log.Printf("[MOCK SMS] To: %s, Message: %s\n", to, message)
		return "mock-sms-sid", nil
	}

	params := &api.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(p.fromNumber)
	params.SetBody(message)

	resp, err := p.client.Api.CreateMessage(params)
	if err != nil {
		return "", err
	}

	if resp.Sid != nil {
		return *resp.Sid, nil
	}
	return "", fmt.Errorf("no sid returned")
}
