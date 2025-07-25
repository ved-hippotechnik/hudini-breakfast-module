package services

import (
	"context"
	"encoding/json"
	"hudini-breakfast-module/internal/logging"
)

// MockPushProvider is a mock implementation for testing
type MockPushProvider struct {
	sentNotifications []MockNotification
}

type MockNotification struct {
	Tokens       []string
	Notification *PushNotification
}

// NewMockPushProvider creates a new mock push provider
func NewMockPushProvider() *MockPushProvider {
	return &MockPushProvider{
		sentNotifications: make([]MockNotification, 0),
	}
}

// Send sends a push notification to a single device
func (m *MockPushProvider) Send(ctx context.Context, token string, notification *PushNotification) error {
	m.sentNotifications = append(m.sentNotifications, MockNotification{
		Tokens:       []string{token},
		Notification: notification,
	})
	
	data, _ := json.Marshal(notification)
	logging.WithField("token", token).WithField("notification", string(data)).Info("Mock push notification sent")
	
	return nil
}

// SendBatch sends push notifications to multiple devices
func (m *MockPushProvider) SendBatch(ctx context.Context, tokens []string, notification *PushNotification) error {
	m.sentNotifications = append(m.sentNotifications, MockNotification{
		Tokens:       tokens,
		Notification: notification,
	})
	
	data, _ := json.Marshal(notification)
	logging.WithField("tokens", tokens).WithField("notification", string(data)).Info("Mock batch push notifications sent")
	
	return nil
}

// GetSentNotifications returns all sent notifications for testing
func (m *MockPushProvider) GetSentNotifications() []MockNotification {
	return m.sentNotifications
}

// FCMPushProvider implements Firebase Cloud Messaging
type FCMPushProvider struct {
	serverKey string
	// In production, this would use the FCM SDK
}

// NewFCMPushProvider creates a new FCM push provider
func NewFCMPushProvider(serverKey string) *FCMPushProvider {
	return &FCMPushProvider{
		serverKey: serverKey,
	}
}

// Send sends a push notification via FCM
func (f *FCMPushProvider) Send(ctx context.Context, token string, notification *PushNotification) error {
	// In production, this would use the FCM SDK
	// For now, just log
	logging.WithField("token", token).Info("FCM push notification would be sent")
	return nil
}

// SendBatch sends push notifications to multiple devices via FCM
func (f *FCMPushProvider) SendBatch(ctx context.Context, tokens []string, notification *PushNotification) error {
	// FCM supports multicast for up to 500 tokens at once
	const batchSize = 500
	
	for i := 0; i < len(tokens); i += batchSize {
		end := i + batchSize
		if end > len(tokens) {
			end = len(tokens)
		}
		
		batch := tokens[i:end]
		// In production, send multicast message
		logging.WithField("batch_size", len(batch)).Info("FCM batch push notifications would be sent")
	}
	
	return nil
}

// APNSPushProvider implements Apple Push Notification Service
type APNSPushProvider struct {
	teamID      string
	keyID       string
	privateKey  string
	production  bool
}

// NewAPNSPushProvider creates a new APNS push provider
func NewAPNSPushProvider(teamID, keyID, privateKey string, production bool) *APNSPushProvider {
	return &APNSPushProvider{
		teamID:     teamID,
		keyID:      keyID,
		privateKey: privateKey,
		production: production,
	}
}

// Send sends a push notification via APNS
func (a *APNSPushProvider) Send(ctx context.Context, token string, notification *PushNotification) error {
	// In production, this would use the APNS HTTP/2 API
	logging.WithField("token", token).Info("APNS push notification would be sent")
	return nil
}

// SendBatch sends push notifications to multiple devices via APNS
func (a *APNSPushProvider) SendBatch(ctx context.Context, tokens []string, notification *PushNotification) error {
	// APNS doesn't support batch, so send individually
	for _, token := range tokens {
		if err := a.Send(ctx, token, notification); err != nil {
			logging.WithError(err).WithField("token", token).Warn("Failed to send APNS notification")
		}
	}
	return nil
}

// WebPushProvider implements Web Push Protocol
type WebPushProvider struct {
	vapidPublicKey  string
	vapidPrivateKey string
	subject         string
}

// NewWebPushProvider creates a new Web Push provider
func NewWebPushProvider(publicKey, privateKey, subject string) *WebPushProvider {
	return &WebPushProvider{
		vapidPublicKey:  publicKey,
		vapidPrivateKey: privateKey,
		subject:         subject,
	}
}

// Send sends a web push notification
func (w *WebPushProvider) Send(ctx context.Context, token string, notification *PushNotification) error {
	// In production, this would use the Web Push Protocol
	logging.WithField("token", token).Info("Web push notification would be sent")
	return nil
}

// SendBatch sends web push notifications to multiple devices
func (w *WebPushProvider) SendBatch(ctx context.Context, tokens []string, notification *PushNotification) error {
	// Send individually
	for _, token := range tokens {
		if err := w.Send(ctx, token, notification); err != nil {
			logging.WithError(err).WithField("token", token).Warn("Failed to send web push notification")
		}
	}
	return nil
}

// MockEmailProvider is a mock email provider for testing
type MockEmailProvider struct {
	sentEmails []MockEmail
}

type MockEmail struct {
	To      string
	Subject string
	Body    string
}

// NewMockEmailProvider creates a new mock email provider
func NewMockEmailProvider() *MockEmailProvider {
	return &MockEmailProvider{
		sentEmails: make([]MockEmail, 0),
	}
}

// Send sends a mock email
func (m *MockEmailProvider) Send(ctx context.Context, to string, subject string, body string) error {
	m.sentEmails = append(m.sentEmails, MockEmail{
		To:      to,
		Subject: subject,
		Body:    body,
	})
	
	logging.WithField("to", to).WithField("subject", subject).Info("Mock email sent")
	return nil
}

// GetSentEmails returns all sent emails for testing
func (m *MockEmailProvider) GetSentEmails() []MockEmail {
	return m.sentEmails
}

// MockSMSProvider is a mock SMS provider for testing
type MockSMSProvider struct {
	sentMessages []MockSMS
}

type MockSMS struct {
	To      string
	Message string
}

// NewMockSMSProvider creates a new mock SMS provider
func NewMockSMSProvider() *MockSMSProvider {
	return &MockSMSProvider{
		sentMessages: make([]MockSMS, 0),
	}
}

// Send sends a mock SMS
func (m *MockSMSProvider) Send(ctx context.Context, to string, message string) error {
	m.sentMessages = append(m.sentMessages, MockSMS{
		To:      to,
		Message: message,
	})
	
	logging.WithField("to", to).WithField("message", message).Info("Mock SMS sent")
	return nil
}

// GetSentMessages returns all sent messages for testing
func (m *MockSMSProvider) GetSentMessages() []MockSMS {
	return m.sentMessages
}