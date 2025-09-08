package scheduler

import (
	"sync"
	"time"

	"github.com/gi4nks/ambros/v3/internal/models"
)

// Scheduler manages scheduled command executions
type Scheduler struct {
	schedules map[string]*models.Schedule
	mutex     sync.RWMutex
	stop      chan struct{}
}

// NewScheduler creates a new scheduler instance
func NewScheduler() *Scheduler {
	return &Scheduler{
		schedules: make(map[string]*models.Schedule),
		stop:      make(chan struct{}),
	}
}

// AddSchedule adds a new schedule for a command
func (s *Scheduler) AddSchedule(commandID string, schedule *models.Schedule) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.schedules[commandID] = schedule
}

// Start begins the scheduling service
func (s *Scheduler) Start() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.checkSchedules()
			case <-s.stop:
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *Scheduler) checkSchedules() {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	now := time.Now()
	for _, schedule := range s.schedules {
		if schedule.Enabled && now.After(schedule.NextRun) {
			// Execute command (implementation to be added)
			schedule.LastRun = now
			// Update next run based on cron expression
		}
	}
}

// Stop halts the scheduler
func (s *Scheduler) Stop() {
	close(s.stop)
}
