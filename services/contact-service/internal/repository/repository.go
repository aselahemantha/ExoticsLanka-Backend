package repository

import (
	"context"
	"fmt"

	"github.com/aselahemantha/exoticsLanka/services/contact-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateInquiry(ctx context.Context, inq *domain.Inquiry) (*domain.Inquiry, error)
	GetInquiries(ctx context.Context, status, subject, priority, search string, page, limit int) ([]domain.Inquiry, int64, error)
	GetInquiryByID(ctx context.Context, id uuid.UUID) (*domain.Inquiry, error)
	UpdateInquiry(ctx context.Context, inq *domain.Inquiry) error
	GetInquiryStats(ctx context.Context) (*domain.InquiryStats, error)

	// Validation
	CheckRateLimit(ctx context.Context, email, ip string) (int, error)
	GetTodayCount(ctx context.Context) (int, error)
}

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) CreateInquiry(ctx context.Context, inq *domain.Inquiry) (*domain.Inquiry, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO contact_inquiries (
			name, email, phone, subject, message, priority, user_id, ip_address, user_agent, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, status, created_at, updated_at
	`, inq.Name, inq.Email, inq.Phone, inq.Subject, inq.Message, inq.Priority, inq.UserID, inq.IPAddress, inq.UserAgent).Scan(
		&inq.ID, &inq.Status, &inq.CreatedAt, &inq.UpdatedAt,
	)
	return inq, err
}

func (r *postgresRepository) GetInquiries(ctx context.Context, status, subject, priority, search string, page, limit int) ([]domain.Inquiry, int64, error) {
	// Base Query
	query := `
		SELECT ci.id, ci.name, ci.email, ci.phone, ci.subject, ci.message, ci.status, ci.priority,
			   ci.admin_response, ci.responded_by, ci.responded_at, ci.user_id, ci.created_at, ci.updated_at
		FROM contact_inquiries ci
		WHERE 1=1
	`
	args := []interface{}{}
	argIdx := 1

	if status != "" {
		query += fmt.Sprintf(" AND ci.status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	if subject != "" {
		query += fmt.Sprintf(" AND ci.subject = $%d", argIdx)
		args = append(args, subject)
		argIdx++
	}
	if priority != "" {
		query += fmt.Sprintf(" AND ci.priority = $%d", argIdx)
		args = append(args, priority)
		argIdx++
	}
	if search != "" {
		// Basic search across email, name, message
		query += fmt.Sprintf(" AND (ci.email ILIKE $%d OR ci.name ILIKE $%d OR ci.message ILIKE $%d)", argIdx, argIdx, argIdx)
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern)
		argIdx++
	}

	// Count Query
	cQuery := "SELECT COUNT(*) FROM contact_inquiries ci WHERE 1=1"
	// Reconstruct args for count
	cArgs := []interface{}{}
	cArgIdx := 1

	if status != "" {
		cQuery += fmt.Sprintf(" AND ci.status = $%d", cArgIdx)
		cArgs = append(cArgs, status)
		cArgIdx++
	}
	if subject != "" {
		cQuery += fmt.Sprintf(" AND ci.subject = $%d", cArgIdx)
		cArgs = append(cArgs, subject)
		cArgIdx++
	}
	if priority != "" {
		cQuery += fmt.Sprintf(" AND ci.priority = $%d", cArgIdx)
		cArgs = append(cArgs, priority)
		cArgIdx++
	}
	if search != "" {
		cQuery += fmt.Sprintf(" AND (ci.email ILIKE $%d OR ci.name ILIKE $%d OR ci.message ILIKE $%d)", cArgIdx, cArgIdx, cArgIdx)
		cArgs = append(cArgs, "%"+search+"%")
		cArgIdx++
	}

	var total int64
	if err := r.db.QueryRow(ctx, cQuery, cArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Pagination
	query += fmt.Sprintf(" ORDER BY ci.created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, (page-1)*limit)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var inquiries []domain.Inquiry
	for rows.Next() {
		var inq domain.Inquiry
		err := rows.Scan(
			&inq.ID, &inq.Name, &inq.Email, &inq.Phone, &inq.Subject, &inq.Message, &inq.Status, &inq.Priority,
			&inq.AdminResponse, &inq.RespondedBy, &inq.RespondedAt, &inq.UserID, &inq.CreatedAt, &inq.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		inquiries = append(inquiries, inq)
	}

	return inquiries, total, nil
}

func (r *postgresRepository) GetInquiryByID(ctx context.Context, id uuid.UUID) (*domain.Inquiry, error) {
	query := `
		SELECT id, name, email, phone, subject, message, status, priority,
			   admin_response, responded_by, responded_at, user_id, ip_address, user_agent, created_at, updated_at
		FROM contact_inquiries
		WHERE id = $1
	`
	var inq domain.Inquiry
	err := r.db.QueryRow(ctx, query, id).Scan(
		&inq.ID, &inq.Name, &inq.Email, &inq.Phone, &inq.Subject, &inq.Message, &inq.Status, &inq.Priority,
		&inq.AdminResponse, &inq.RespondedBy, &inq.RespondedAt, &inq.UserID, &inq.IPAddress, &inq.UserAgent, &inq.CreatedAt, &inq.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &inq, nil
}

func (r *postgresRepository) UpdateInquiry(ctx context.Context, inq *domain.Inquiry) error {
	_, err := r.db.Exec(ctx, `
		UPDATE contact_inquiries 
		SET status = $1, priority = $2, admin_response = $3, responded_by = $4, responded_at = $5, updated_at = NOW()
		WHERE id = $6
	`, inq.Status, inq.Priority, inq.AdminResponse, inq.RespondedBy, inq.RespondedAt, inq.ID)
	return err
}

func (r *postgresRepository) GetInquiryStats(ctx context.Context) (*domain.InquiryStats, error) {
	stats := &domain.InquiryStats{
		ByStatus:   make(map[string]int),
		BySubject:  make(map[string]int),
		ByPriority: make(map[string]int),
	}

	// Total
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM contact_inquiries").Scan(&stats.Total); err != nil {
		return nil, err
	}

	// By Status
	rows, err := r.db.Query(ctx, "SELECT status, COUNT(*) FROM contact_inquiries GROUP BY status")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var s string
		var c int
		if err := rows.Scan(&s, &c); err == nil {
			stats.ByStatus[s] = c
		}
	}

	// By Subject
	rows2, err := r.db.Query(ctx, "SELECT subject, COUNT(*) FROM contact_inquiries GROUP BY subject")
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var s string
		var c int
		if err := rows2.Scan(&s, &c); err == nil {
			stats.BySubject[s] = c
		}
	}

	// By Priority
	rows3, err := r.db.Query(ctx, "SELECT priority, COUNT(*) FROM contact_inquiries GROUP BY priority")
	if err != nil {
		return nil, err
	}
	defer rows3.Close()
	for rows3.Next() {
		var s string
		var c int
		if err := rows3.Scan(&s, &c); err == nil {
			stats.ByPriority[s] = c
		}
	}

	return stats, nil
}

func (r *postgresRepository) CheckRateLimit(ctx context.Context, email, ip string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM contact_inquiries 
		WHERE (email = $1 OR ip_address = $2)
		AND created_at > NOW() - INTERVAL '1 hour'
	`, email, ip).Scan(&count)
	return count, err
}

func (r *postgresRepository) GetTodayCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM contact_inquiries 
		WHERE DATE(created_at) = CURRENT_DATE
	`).Scan(&count)
	return count, err
}
