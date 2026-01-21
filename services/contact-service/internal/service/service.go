package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aselahemantha/exoticsLanka/services/contact-service/internal/domain"
	"github.com/aselahemantha/exoticsLanka/services/contact-service/internal/repository"
	"github.com/google/uuid"
)

type Service interface {
	SubmitInquiry(ctx context.Context, req domain.CreateInquiryRequest, metadata map[string]interface{}) (*domain.Inquiry, error)
	GetInquiries(ctx context.Context, status, subject, priority, search string, page, limit int) ([]domain.Inquiry, domain.Pagination, error)
	GetInquiry(ctx context.Context, id uuid.UUID) (*domain.Inquiry, error)
	RespondInquiry(ctx context.Context, id uuid.UUID, adminID uuid.UUID, req domain.RespondInquiryRequest) (*domain.Inquiry, error)
	GetStats(ctx context.Context) (*domain.InquiryStats, error)
}

type service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &service{repo: repo}
}

func (s *service) SubmitInquiry(ctx context.Context, req domain.CreateInquiryRequest, metadata map[string]interface{}) (*domain.Inquiry, error) {
	// Extract metadata
	ip, _ := metadata["ipAddress"].(string)
	userAgent, _ := metadata["userAgent"].(string)
	var userID *uuid.UUID
	if uid, ok := metadata["userId"].(uuid.UUID); ok {
		userID = &uid
	}

	// 1. Rate Limit Check
	count, err := s.repo.CheckRateLimit(ctx, req.Email, ip)
	if err != nil {
		return nil, err
	}
	if count >= 5 {
		return nil, fmt.Errorf("too many inquiries. please try again later")
	}

	// 2. Generate Reference Number Logic
	// Note: We aren't storing it in DB schema provided, but we return it.
	// We can calculate it based on today's count.
	// However, without a DB column, this reference number is ephemeral unless we add it to the schema or calculate it from ID/Date.
	// Given the plan strictly follows provided schema and schema doesn't have it, we will generate it for the *response object*
	// but purely as a display value derived from date + daily count.
	// To make it persistent/consistent, ideally schema should have it.
	// For now, let's treat it as a return value decoration or just skip persistence if schema is immutable.
	// The plan says "Generate Reference Number ... Create".
	// Let's retrieve today's count to compute the suffix.
	todayCount, err := s.repo.GetTodayCount(ctx)
	if err != nil {
		return nil, err
	}
	refNum := fmt.Sprintf("INQ-%s-%03d", time.Now().Format("20060102"), todayCount+1)

	// 3. Determine Priority
	priority := "normal"
	if req.Subject == "complaint" {
		priority = "high"
	}
	if req.Subject == "support" && strings.Contains(strings.ToLower(req.Message), "urgent") {
		priority = "high"
	}

	// 4. Create Inquiry
	inq := &domain.Inquiry{
		Name:      req.Name,
		Email:     req.Email,
		Phone:     &req.Phone,
		Subject:   req.Subject,
		Message:   req.Message,
		Priority:  priority,
		UserID:    userID,
		IPAddress: ip,
		UserAgent: userAgent,
	}

	createdInq, err := s.repo.CreateInquiry(ctx, inq)
	if err != nil {
		return nil, err
	}

	// Decorate with generated ref number
	createdInq.ReferenceNumber = refNum

	// 5. Send Email Notification (Stub)
	// emailService.SendConfirmation(...)

	return createdInq, nil
}

func (s *service) GetInquiries(ctx context.Context, status, subject, priority, search string, page, limit int) ([]domain.Inquiry, domain.Pagination, error) {
	inquiries, total, err := s.repo.GetInquiries(ctx, status, subject, priority, search, page, limit)
	if err != nil {
		return nil, domain.Pagination{}, err
	}

	pagination := domain.Pagination{
		Page:  page,
		Limit: limit,
		Total: total,
	}
	if limit > 0 {
		pagination.TotalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	// Decorate Reference Number if needed (mocked for list view as we don't store it)
	// Ideally we store it. Since we didn't add column, we can't show consistent refs in list.
	// We will skip ref number in list for now or generate a hash-based one from ID.

	return inquiries, pagination, nil
}

func (s *service) GetInquiry(ctx context.Context, id uuid.UUID) (*domain.Inquiry, error) {
	return s.repo.GetInquiryByID(ctx, id)
}

func (s *service) RespondInquiry(ctx context.Context, id uuid.UUID, adminID uuid.UUID, req domain.RespondInquiryRequest) (*domain.Inquiry, error) {
	inq, err := s.repo.GetInquiryByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if inq == nil {
		return nil, fmt.Errorf("inquiry not found")
	}

	now := time.Now()
	if req.Status != "" {
		inq.Status = req.Status
	} else {
		inq.Status = "responded"
	}
	if req.Priority != "" {
		inq.Priority = req.Priority
	}

	inq.AdminResponse = &req.AdminResponse
	inq.RespondedBy = &adminID
	inq.RespondedAt = &now

	if err := s.repo.UpdateInquiry(ctx, inq); err != nil {
		return nil, err
	}

	// Send Email Response (Stub)

	return inq, nil
}

func (s *service) GetStats(ctx context.Context) (*domain.InquiryStats, error) {
	return s.repo.GetInquiryStats(ctx)
}
