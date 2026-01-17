package jobs

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type JobScheduler struct {
	db *pgxpool.Pool
}

func NewJobScheduler(db *pgxpool.Pool) *JobScheduler {
	return &JobScheduler{db: db}
}

func (s *JobScheduler) Start() {
	go s.runDailyJobs()
	go s.runHourlyJobs()
}

func (s *JobScheduler) runDailyJobs() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Run once on startup or wait? usually wait, but for now let's just wait for ticker
	// Or maybe run immediately if we want to ensure consistency on restart

	for range ticker.C {
		ctx := context.Background()
		log.Println("Running daily jobs...")

		if err := s.updateDaysListed(ctx); err != nil {
			log.Printf("Error updating days listed: %v", err)
		}

		if err := s.expireListings(ctx); err != nil {
			log.Printf("Error expiring listings: %v", err)
		}

		if err := s.recalculateHealthScores(ctx); err != nil {
			log.Printf("Error recalculating health scores: %v", err)
		}
	}
}

func (s *JobScheduler) runHourlyJobs() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		log.Println("Running hourly jobs...")

		if err := s.updateBrandCounts(ctx); err != nil {
			log.Printf("Error updating brand counts: %v", err)
		}
	}
}

func (s *JobScheduler) updateDaysListed(ctx context.Context) error {
	_, err := s.db.Exec(ctx, `
		UPDATE car_listings 
		SET days_listed = days_listed + 1,
			is_new = CASE WHEN days_listed < 7 THEN TRUE ELSE FALSE END
		WHERE status = 'active'
	`)
	return err
}

func (s *JobScheduler) expireListings(ctx context.Context) error {
	_, err := s.db.Exec(ctx, `
		UPDATE car_listings 
		SET status = 'expired' 
		WHERE status = 'active' 
		  AND expires_at IS NOT NULL 
		  AND expires_at < NOW()
	`)
	return err
}

func (s *JobScheduler) updateBrandCounts(ctx context.Context) error {
	_, err := s.db.Exec(ctx, `
		UPDATE car_brands b
		SET listing_count = (
			SELECT COUNT(*) 
			FROM car_listings cl 
			WHERE cl.make = b.name AND cl.status = 'active'
		)
	`)
	return err
}

func (s *JobScheduler) recalculateHealthScores(ctx context.Context) error {
	// Simple implementation using DB logic or fetching and updating
	// For performance, pure SQL update if possible, but logic is complex.
	// Implementing simplified version here or full logic requires fetching.
	// Let's defer full implementation to avoid fetching all listings
	log.Println("Recalculating health scores (placeholder)")
	return nil
}
