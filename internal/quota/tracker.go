package quota

import (
	"fmt"
	"sync"
	"time"

	"github.com/ivanjtm/YunoChallenge/internal/model"
)

type Tracker struct {
	mu         sync.Mutex
	processors map[string]model.Processor
	usage      map[string]int
	overrides  map[string]model.ProcessorOverride
	resetDate  time.Time
}

func NewTracker(processors []model.Processor) *Tracker {
	procMap := make(map[string]model.Processor)
	for _, p := range processors {
		procMap[p.ID] = p
	}
	return &Tracker{
		processors: procMap,
		usage:      make(map[string]int),
		overrides:  make(map[string]model.ProcessorOverride),
		resetDate:  time.Now().UTC().Truncate(24 * time.Hour),
	}
}

func (t *Tracker) resetIfNewDay(now time.Time) {
	today := now.UTC().Truncate(24 * time.Hour)
	if today.After(t.resetDate) {
		t.usage = make(map[string]int)
		t.resetDate = today
	}
}

func (t *Tracker) IsAvailable(processorID string, now time.Time) (bool, string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.resetIfNewDay(now)

	if override, ok := t.overrides[processorID]; ok {
		if override.Available != nil && !*override.Available {
			return false, "Processor marked as unavailable (simulated)"
		}
		if override.AtCapacity != nil && *override.AtCapacity {
			return false, "Processor marked as at capacity (simulated)"
		}
	}

	proc, ok := t.processors[processorID]
	if !ok {
		return false, fmt.Sprintf("Unknown processor: %s", processorID)
	}

	if proc.DailyQuota > 0 {
		used := t.usage[processorID]
		if override, ok := t.overrides[processorID]; ok && override.QuotaUsed != nil {
			used = *override.QuotaUsed
		}
		if used >= proc.DailyQuota {
			return false, fmt.Sprintf("Daily quota exhausted: %d/%d used", used, proc.DailyQuota)
		}
	}

	return true, ""
}

func (t *Tracker) Consume(processorID string, now time.Time) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.resetIfNewDay(now)
	t.usage[processorID]++
	return nil
}

func (t *Tracker) SetOverrides(overrides map[string]model.ProcessorOverride) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for k, v := range overrides {
		t.overrides[k] = v
	}
}

func (t *Tracker) ResetOverrides() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.overrides = make(map[string]model.ProcessorOverride)
}

func (t *Tracker) Status(now time.Time) []model.QuotaStatus {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.resetIfNewDay(now)

	var statuses []model.QuotaStatus
	for id, proc := range t.processors {
		used := t.usage[id]
		if override, ok := t.overrides[id]; ok && override.QuotaUsed != nil {
			used = *override.QuotaUsed
		}

		remaining := proc.DailyQuota - used
		if remaining < 0 {
			remaining = 0
		}

		available, reason := true, ""

		if override, ok := t.overrides[id]; ok {
			if override.Available != nil && !*override.Available {
				available = false
				reason = "Processor marked as unavailable (simulated)"
			} else if override.AtCapacity != nil && *override.AtCapacity {
				available = false
				reason = "Processor marked as at capacity (simulated)"
			}
		}

		if available && proc.DailyQuota > 0 && used >= proc.DailyQuota {
			available = false
			reason = fmt.Sprintf("Daily quota exhausted: %d/%d", used, proc.DailyQuota)
		}

		statuses = append(statuses, model.QuotaStatus{
			ProcessorID:       id,
			DailyQuota:        proc.DailyQuota,
			UsedToday:         used,
			Remaining:         remaining,
			IsAvailable:       available,
			UnavailableReason: reason,
		})
	}
	return statuses
}
