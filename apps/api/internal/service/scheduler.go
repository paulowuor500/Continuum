package service

import (
	"context"
	"log"
	"time"

	"continuum/api/internal/repository"
)

type Scheduler struct {
	repo *repository.Database
}

func NewScheduler(repo *repository.Database) *Scheduler {
	return &Scheduler{repo: repo}
}

// StartCheckLoop initializes a non-blocking background loop that executes scans periodically
func (s *Scheduler) StartCheckLoop(ctx context.Context, checkFrequency time.Duration) {
	ticker := time.NewTicker(checkFrequency)
	
	go func() {
		for {
			select {
			case <-ticker.C:
				s.evaluateVaultLifespans(ctx)
			case <-ctx.Done():
				ticker.Stop()
				log.Println("🛑 Background dead-man switch tracking ticker stopped cleanly.")
				return
			}
		}
	}()
}

// evaluateVaultLifespans mutates stale ACTIVE vaults to DORMANT safely under a timeout boundary
func (s *Scheduler) evaluateVaultLifespans(rootCtx context.Context) {
	// SECURITY BOUNDARY: Enforce a strict 5-second processing timeout boundary 
	// per execution scan loop so database connections never hang forever.
	ctx, cancel := context.WithTimeout(rootCtx, 5*time.Second)
	defer cancel()

	// PERFORMANCE OPTIMIZATION: Swapped out slow string concatenation for native 
	// integer-to-interval mathematical scaling factors.
	query := `
		UPDATE vaults 
		SET status = 'DORMANT', updated_at = NOW()
		WHERE status = 'ACTIVE' 
		  AND NOW() > (last_check_in_at + (check_in_interval_seconds * INTERVAL '1 second'));
	`
	
	// FIXED: Enforced context constraint boundary tracking via ExecContext
	result, err := s.repo.Db.ExecContext(ctx, query)
	if err != nil {
		log.Printf("❌ ERROR during dead-man loop evaluation: %v", err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("🛡️ Continuum Safety Alert: %d stale vault(s) shifted to DORMANT status.", rowsAffected)
	}
}