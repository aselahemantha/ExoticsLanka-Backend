package repository

import (
	"context"
	"fmt"

	"github.com/aselahemantha/exoticsLanka/services/reports-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateReport(ctx context.Context, report *domain.Report) (*domain.Report, error)
	GetReports(ctx context.Context, status, reason string, page, limit int) ([]domain.Report, int64, error)
	GetReportByID(ctx context.Context, id uuid.UUID) (*domain.Report, error)
	UpdateReport(ctx context.Context, report *domain.Report) error
	GetReportStats(ctx context.Context) (*domain.ReportStats, error)

	// Validations
	HasRecentReport(ctx context.Context, reporterID, listingID uuid.UUID) (bool, error)
	GetPendingReportsCount(ctx context.Context, listingID uuid.UUID) (int, error)

	// Cross-Domain (Shared DB)
	UpdateListingStatus(ctx context.Context, listingID uuid.UUID, status string) error
	UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error
}

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) CreateReport(ctx context.Context, report *domain.Report) (*domain.Report, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO listing_reports (listing_id, reporter_id, reason, details)
		VALUES ($1, $2, $3, $4)
		RETURNING id, status, created_at, updated_at
	`, report.ListingID, report.ReporterID, report.Reason, report.Details).Scan(
		&report.ID, &report.Status, &report.CreatedAt, &report.UpdatedAt,
	)
	return report, err
}

func (r *postgresRepository) GetReports(ctx context.Context, status, reason string, page, limit int) ([]domain.Report, int64, error) {
	// Base Query
	query := `
		SELECT r.id, r.listing_id, r.reporter_id, r.reason, r.details, r.status, 
               r.admin_notes, r.action_taken, r.resolved_at, r.resolved_by, r.created_at,
               cl.title, u.email as reporter_email
		FROM listing_reports r
		JOIN car_listings cl ON r.listing_id = cl.id
		LEFT JOIN users u ON r.reporter_id = u.id
		WHERE 1=1
	`
	args := []interface{}{}
	argIdx := 1

	if status != "" {
		query += fmt.Sprintf(" AND r.status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	if reason != "" {
		query += fmt.Sprintf(" AND r.reason = $%d", argIdx)
		args = append(args, reason)
		argIdx++
	}

	// Count Query
	countQuery := "SELECT COUNT(*) FROM listing_reports r WHERE 1=1"
	if status != "" {
		countQuery += fmt.Sprintf(" AND r.status = $%d", 1)
	}
	// Note: Re-using args logic for count is tricky if we just append.
	// Simplified: Run separate count query logic or window function.
	// For MVP, window function is cleaner but standard count is safer for large datasets.
	// Let's rely on standard count with re-built args.
	cArgs := []interface{}{}
	cArgIdx := 1
	cQuery := "SELECT COUNT(*) FROM listing_reports r WHERE 1=1"
	if status != "" {
		cQuery += fmt.Sprintf(" AND r.status = $%d", cArgIdx)
		cArgs = append(cArgs, status)
		cArgIdx++
	}
	if reason != "" {
		cQuery += fmt.Sprintf(" AND r.reason = $%d", cArgIdx)
		cArgs = append(cArgs, reason)
		cArgIdx++
	}

	var total int64
	if err := r.db.QueryRow(ctx, cQuery, cArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Pagination
	query += fmt.Sprintf(" ORDER BY r.created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, (page-1)*limit)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reports []domain.Report
	for rows.Next() {
		var rpt domain.Report
		var reporterEmail *string
		var listingTitle string

		err := rows.Scan(
			&rpt.ID, &rpt.ListingID, &rpt.ReporterID, &rpt.Reason, &rpt.Details, &rpt.Status,
			&rpt.AdminNotes, &rpt.ActionTaken, &rpt.ResolvedAt, &rpt.ResolvedBy, &rpt.CreatedAt,
			&listingTitle, &reporterEmail,
		)
		if err != nil {
			return nil, 0, err
		}

		// Map extra fields to summary structs
		rpt.Listing = &domain.ListingSummary{ID: rpt.ListingID, Title: listingTitle}
		if rpt.ReporterID != nil && reporterEmail != nil {
			rpt.Reporter = &domain.UserSummary{ID: *rpt.ReporterID, Email: *reporterEmail}
		}

		reports = append(reports, rpt)
	}

	return reports, total, nil
}

func (r *postgresRepository) GetReportByID(ctx context.Context, id uuid.UUID) (*domain.Report, error) {
	query := `
		SELECT r.id, r.listing_id, r.reporter_id, r.reason, r.details, r.status, 
               r.admin_notes, r.action_taken, r.resolved_at, r.resolved_by, r.created_at, r.updated_at,
               cl.title, cl.price, cl.status,
               u.id as reporter_id, u.email as reporter_email,
               s.id as listing_owner_id, s.name as listing_owner_name, s.email as listing_owner_email
		FROM listing_reports r
		JOIN car_listings cl ON r.listing_id = cl.id
		JOIN users s ON cl.user_id = s.id
		LEFT JOIN users u ON r.reporter_id = u.id
		WHERE r.id = $1
	`
	var rpt domain.Report
	var lListing domain.ListingSummary
	var lOwner domain.UserSummary
	var rReporter domain.UserSummary
	var rReporterID *uuid.UUID
	var rReporterEmail *string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&rpt.ID, &rpt.ListingID, &rpt.ReporterID, &rpt.Reason, &rpt.Details, &rpt.Status,
		&rpt.AdminNotes, &rpt.ActionTaken, &rpt.ResolvedAt, &rpt.ResolvedBy, &rpt.CreatedAt, &rpt.UpdatedAt,
		&lListing.Title, &lListing.Price, &lListing.Status,
		&rReporterID, &rReporterEmail,
		&lOwner.ID, &lOwner.Name, &lOwner.Email,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Reconstruct complex objects
	lListing.ID = rpt.ListingID
	lListing.User = lOwner
	rpt.Listing = &lListing

	if rReporterID != nil && rReporterEmail != nil {
		rReporter.ID = *rReporterID
		rReporter.Email = *rReporterEmail
		rpt.Reporter = &rReporter
	}

	return &rpt, nil
}

func (r *postgresRepository) UpdateReport(ctx context.Context, report *domain.Report) error {
	_, err := r.db.Exec(ctx, `
		UPDATE listing_reports 
		SET status = $1, admin_notes = $2, action_taken = $3, resolved_at = $4, resolved_by = $5, updated_at = NOW()
		WHERE id = $6
	`, report.Status, report.AdminNotes, report.ActionTaken, report.ResolvedAt, report.ResolvedBy, report.ID)
	return err
}

func (r *postgresRepository) GetReportStats(ctx context.Context) (*domain.ReportStats, error) {
	stats := &domain.ReportStats{
		ByStatus: make(map[string]int),
		ByReason: make(map[string]int),
	}

	// Total
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM listing_reports").Scan(&stats.TotalReports); err != nil {
		return nil, err
	}

	// By Status
	rows, err := r.db.Query(ctx, "SELECT status, COUNT(*) FROM listing_reports GROUP BY status")
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

	// By Reason
	rows2, err := r.db.Query(ctx, "SELECT reason, COUNT(*) FROM listing_reports GROUP BY reason")
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var re string
		var c int
		if err := rows2.Scan(&re, &c); err == nil {
			stats.ByReason[re] = c
		}
	}

	return stats, nil
}

func (r *postgresRepository) HasRecentReport(ctx context.Context, reporterID, listingID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM listing_reports 
			WHERE listing_id = $1 AND reporter_id = $2 
			AND created_at > NOW() - INTERVAL '24 hours'
		)
	`, listingID, reporterID).Scan(&exists)
	return exists, err
}

func (r *postgresRepository) GetPendingReportsCount(ctx context.Context, listingID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM listing_reports 
		WHERE listing_id = $1 AND status = 'pending'
	`, listingID).Scan(&count)
	return count, err
}

func (r *postgresRepository) UpdateListingStatus(ctx context.Context, listingID uuid.UUID, status string) error {
	_, err := r.db.Exec(ctx, "UPDATE car_listings SET status = $1 WHERE id = $2", status, listingID)
	return err
}

func (r *postgresRepository) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error {
	// Assuming 'users' table has a 'status' or similar field (often 'is_active' or 'role').
	// If the schema doesn't have status, we might need to check other services.
	// For now, let's assume standard 'status' field for moderation.
	// If it fails, we catch it.
	_, err := r.db.Exec(ctx, "UPDATE users SET status = $1 WHERE id = $2", status, userID)
	return err
}
