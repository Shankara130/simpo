package services

import (
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// StockMetricsService tracks metrics for real-time stock system
// Story 4.2, Task 16: Add Monitoring and Metrics (AC: 6)
type StockMetricsService struct {
	logger *slog.Logger

	// Connection metrics (Task 16.1)
	activeConnections     atomic.Int64
	totalConnections      atomic.Int64
	connectionsByBranch   map[uint]int64
	connectionsByBranchMu sync.RWMutex

	// Event publishing metrics (Task 16.2)
	eventsPublished      atomic.Int64
	eventsPublishFailed  atomic.Int64
	publishRatePerSecond float64
	publishRateMu        sync.RWMutex
	publishHistory       []time.Time
	publishHistoryMu     sync.Mutex

	// Event delivery metrics (Task 16.3)
	eventsDelivered    atomic.Int64
	eventsDeliveredFailed atomic.Int64
	deliveryLatencies   []time.Duration
	deliveryLatenciesMu sync.Mutex

	// Stock reconciliation accuracy (Task 16.5)
	reconciliationChecks atomic.Int64
	reconciliationMatches atomic.Int64
	reconciliationMismatches atomic.Int64

	// Failure rate tracking (Task 16.4)
	failureWindowStart  time.Time
	failureWindowEvents atomic.Int64
	failureWindowFailed atomic.Int64

	// Alerting
	alertThreshold    float64 // 5% failure rate threshold (Task 16.4)
	alertCallback     func(string)
	alertCallbackMu   sync.RWMutex
	lastAlertTime     time.Time
	alertCooldown     time.Duration
}

// NewStockMetricsService creates a new metrics service
func NewStockMetricsService(logger *slog.Logger) *StockMetricsService {
	return &StockMetricsService{
		logger:              logger,
		connectionsByBranch: make(map[uint]int64),
		publishHistory:      make([]time.Time, 0, 60), // Keep last 60 publishes
		deliveryLatencies:   make([]time.Duration, 0, 1000), // Keep last 1000 latencies
		failureWindowStart:  time.Now(),
		alertThreshold:      0.05, // 5%
		alertCooldown:       5 * time.Minute, // Alert at most every 5 minutes
	}
}

// SetAlertCallback sets a callback function for alerts
// Story 4.2, Task 16.4: Alert on delivery failures (>5% failure rate)
func (s *StockMetricsService) SetAlertCallback(callback func(string)) {
	s.alertCallbackMu.Lock()
	defer s.alertCallbackMu.Unlock()
	s.alertCallback = callback
}

// Connection tracking methods
// Story 4.2, Task 16.1: Track WebSocket connection count

func (s *StockMetricsService) RecordConnection(branchID uint) {
	s.activeConnections.Add(1)
	s.totalConnections.Add(1)

	s.connectionsByBranchMu.Lock()
	s.connectionsByBranch[branchID]++
	s.connectionsByBranchMu.Unlock()

	s.logger.Info("WebSocket connection established",
		"branch_id", branchID,
		"active_connections", s.activeConnections.Load(),
	)
}

func (s *StockMetricsService) RecordDisconnection(branchID uint) {
	s.activeConnections.Add(-1)

	s.connectionsByBranchMu.Lock()
	s.connectionsByBranch[branchID]--
	s.connectionsByBranchMu.Unlock()

	s.logger.Info("WebSocket connection closed",
		"branch_id", branchID,
		"active_connections", s.activeConnections.Load(),
	)
}

func (s *StockMetricsService) GetActiveConnections() int64 {
	return s.activeConnections.Load()
}

func (s *StockMetricsService) GetTotalConnections() int64 {
	return s.totalConnections.Load()
}

func (s *StockMetricsService) GetConnectionsByBranch() map[uint]int64 {
	s.connectionsByBranchMu.RLock()
	defer s.connectionsByBranchMu.RUnlock()

	result := make(map[uint]int64, len(s.connectionsByBranch))
	for k, v := range s.connectionsByBranch {
		result[k] = v
	}
	return result
}

// Event publishing metrics
// Story 4.2, Task 16.2: Track event publishing rate

func (s *StockMetricsService) RecordEventPublished(success bool) {
	s.publishHistoryMu.Lock()
	now := time.Now()
	s.publishHistory = append(s.publishHistory, now)
	// Keep only last 60 seconds of history
	cutoff := now.Add(-60 * time.Second)
	for len(s.publishHistory) > 0 && s.publishHistory[0].Before(cutoff) {
		s.publishHistory = s.publishHistory[1:]
	}
	s.publishHistoryMu.Unlock()

	if success {
		s.eventsPublished.Add(1)
	} else {
		s.eventsPublishFailed.Add(1)
	}

	// Update rate calculation
	s.calculatePublishRate()

	// Check failure rate
	s.checkFailureRate()
}

func (s *StockMetricsService) calculatePublishRate() {
	s.publishHistoryMu.Lock()
	defer s.publishHistoryMu.Unlock()

	if len(s.publishHistory) < 2 {
		s.publishRatePerSecond = 0
		return
	}

	duration := s.publishHistory[len(s.publishHistory)-1].Sub(s.publishHistory[0]).Seconds()
	if duration > 0 {
		s.publishRatePerSecond = float64(len(s.publishHistory)) / duration
	}
}

func (s *StockMetricsService) GetPublishRate() float64 {
	s.publishRateMu.RLock()
	defer s.publishRateMu.RUnlock()
	return s.publishRatePerSecond
}

func (s *StockMetricsService) GetEventsPublished() int64 {
	return s.eventsPublished.Load()
}

func (s *StockMetricsService) GetEventsPublishFailed() int64 {
	return s.eventsPublishFailed.Load()
}

// Event delivery metrics
// Story 4.2, Task 16.3: Track event delivery latency

func (s *StockMetricsService) RecordEventDelivery(success bool, latency time.Duration) {
	if success {
		s.eventsDelivered.Add(1)
	} else {
		s.eventsDeliveredFailed.Add(1)
	}

	// Track latency (only for successful deliveries)
	if success && latency > 0 {
		s.deliveryLatenciesMu.Lock()
		s.deliveryLatencies = append(s.deliveryLatencies, latency)
		// Keep only last 1000 latencies
		if len(s.deliveryLatencies) > 1000 {
			s.deliveryLatencies = s.deliveryLatencies[len(s.deliveryLatencies)-1000:]
		}
		s.deliveryLatenciesMu.Unlock()
	}

	// Check failure rate
	s.checkFailureRate()
}

func (s *StockMetricsService) GetAverageLatency() time.Duration {
	s.deliveryLatenciesMu.Lock()
	defer s.deliveryLatenciesMu.Unlock()

	if len(s.deliveryLatencies) == 0 {
		return 0
	}

	var sum time.Duration
	for _, latency := range s.deliveryLatencies {
		sum += latency
	}
	return sum / time.Duration(len(s.deliveryLatencies))
}

func (s *StockMetricsService) GetPercentileLatency(percentile float64) time.Duration {
	s.deliveryLatenciesMu.Lock()
	defer s.deliveryLatenciesMu.Unlock()

	if len(s.deliveryLatencies) == 0 {
		return 0
	}

	// Simple percentile calculation
	// For production, use a more efficient algorithm
	sorted := make([]time.Duration, len(s.deliveryLatencies))
	copy(sorted, s.deliveryLatencies)

	// Simple bubble sort (good enough for small datasets)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	index := int(float64(len(sorted)) * percentile)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return sorted[index]
}

func (s *StockMetricsService) GetEventsDelivered() int64 {
	return s.eventsDelivered.Load()
}

func (s *StockMetricsService) GetEventsDeliveredFailed() int64 {
	return s.eventsDeliveredFailed.Load()
}

// Stock reconciliation accuracy metrics
// Story 4.2, Task 16.5: Log stock reconciliation accuracy metrics

func (s *StockMetricsService) RecordReconciliation(matched bool) {
	s.reconciliationChecks.Add(1)
	if matched {
		s.reconciliationMatches.Add(1)
	} else {
		s.reconciliationMismatches.Add(1)
	}
}

func (s *StockMetricsService) GetReconciliationAccuracy() float64 {
	checks := s.reconciliationChecks.Load()
	if checks == 0 {
		return 100.0 // No mismatches if no checks
	}

	matches := s.reconciliationMatches.Load()
	return (float64(matches) / float64(checks)) * 100.0
}

func (s *StockMetricsService) GetReconciliationChecks() int64 {
	return s.reconciliationChecks.Load()
}

func (s *StockMetricsService) GetReconciliationMatches() int64 {
	return s.reconciliationMatches.Load()
}

func (s *StockMetricsService) GetReconciliationMismatches() int64 {
	return s.reconciliationMismatches.Load()
}

// Failure rate monitoring and alerting
// Story 4.2, Task 16.4: Alert on delivery failures (>5% failure rate)

func (s *StockMetricsService) checkFailureRate() {
	s.failureWindowEvents.Add(1)

	// Calculate failure rate over current window
	totalEvents := s.failureWindowEvents.Load()
	failedEvents := s.eventsDeliveredFailed.Load() + s.eventsPublishFailed.Load()

	if totalEvents > 100 { // Minimum sample size
		failureRate := float64(failedEvents) / float64(totalEvents)

		if failureRate > s.alertThreshold {
			s.sendAlert(failureRate)
		}
	}

	// Reset window every minute
	if time.Since(s.failureWindowStart) > time.Minute {
		s.failureWindowStart = time.Now()
		s.failureWindowEvents.Store(0)
		s.failureWindowFailed.Store(0)
	}
}

func (s *StockMetricsService) sendAlert(failureRate float64) {
	// Check cooldown
	if time.Since(s.lastAlertTime) < s.alertCooldown {
		return
	}

	s.lastAlertTime = time.Now()

	s.logger.Warn("Real-time stock system failure rate exceeded threshold",
		"failure_rate", failureRate*100,
		"threshold", s.alertThreshold*100,
		"events_delivered", s.eventsDelivered.Load(),
		"events_failed", s.eventsDeliveredFailed.Load()+s.eventsPublishFailed.Load(),
	)

	s.alertCallbackMu.RLock()
	callback := s.alertCallback
	s.alertCallbackMu.RUnlock()

	if callback != nil {
		message := fmt.Sprintf("Real-time stock system failure rate: %.2f%% (threshold: %.2f%%)",
			failureRate*100, s.alertThreshold*100)
		go callback(message)
	}
}

// Get overall metrics snapshot
type MetricsSnapshot struct {
	Timestamp                time.Time
	ActiveConnections        int64
	TotalConnections         int64
	EventsPublished          int64
	EventsPublishFailed      int64
	PublishRatePerSecond     float64
	EventsDelivered          int64
	EventsDeliveredFailed    int64
	AverageLatency           time.Duration
	P95Latency               time.Duration
	P99Latency               time.Duration
	ReconciliationAccuracy   float64
	ReconciliationChecks     int64
	ReconciliationMatches    int64
	ReconciliationMismatches int64
}

func (s *StockMetricsService) GetSnapshot() *MetricsSnapshot {
	return &MetricsSnapshot{
		Timestamp:                time.Now(),
		ActiveConnections:        s.activeConnections.Load(),
		TotalConnections:         s.totalConnections.Load(),
		EventsPublished:          s.eventsPublished.Load(),
		EventsPublishFailed:      s.eventsPublishFailed.Load(),
		PublishRatePerSecond:     s.GetPublishRate(),
		EventsDelivered:          s.eventsDelivered.Load(),
		EventsDeliveredFailed:    s.eventsDeliveredFailed.Load(),
		AverageLatency:           s.GetAverageLatency(),
		P95Latency:               s.GetPercentileLatency(0.95),
		P99Latency:               s.GetPercentileLatency(0.99),
		ReconciliationAccuracy:   s.GetReconciliationAccuracy(),
		ReconciliationChecks:     s.reconciliationChecks.Load(),
		ReconciliationMatches:    s.reconciliationMatches.Load(),
		ReconciliationMismatches: s.reconciliationMismatches.Load(),
	}
}

// LogMetrics logs a summary of current metrics
func (s *StockMetricsService) LogMetrics() {
	snapshot := s.GetSnapshot()

	s.logger.Info("Real-time stock system metrics",
		"active_connections", snapshot.ActiveConnections,
		"total_connections", snapshot.TotalConnections,
		"events_published", snapshot.EventsPublished,
		"events_failed", snapshot.EventsPublishFailed+snapshot.EventsDeliveredFailed,
		"publish_rate_per_sec", snapshot.PublishRatePerSecond,
		"avg_latency_ms", snapshot.AverageLatency.Milliseconds(),
		"p95_latency_ms", snapshot.P95Latency.Milliseconds(),
		"p99_latency_ms", snapshot.P99Latency.Milliseconds(),
		"reconciliation_accuracy", snapshot.ReconciliationAccuracy,
	)
}
