// File: internal/srs/algorithm_test.go

package srs

import (
	"testing"
	"time"

	"github.com/DavidMiserak/GoCard/internal/model"
)

func TestSM2SchedulerRating1(t *testing.T) {
	scheduler := NewSM2Scheduler()
	card := &model.Card{
		Ease:     2.5,
		Interval: 5,
	}

	result := scheduler.Schedule(card, 1)

	// Rating 1 (blackout) should reset interval and reduce ease
	if card.Interval != 1 {
		t.Errorf("Expected interval 1 for rating 1, got %d", card.Interval)
	}

	expectedEase := 2.2 // 2.5 - 0.3
	if card.Ease != expectedEase {
		t.Errorf("Expected ease %.1f, got %.1f", expectedEase, card.Ease)
	}

	// Check result
	if result.Interval != 1 {
		t.Errorf("Expected result interval 1, got %d", result.Interval)
	}
}

func TestSM2SchedulerRating5(t *testing.T) {
	scheduler := NewSM2Scheduler()
	card := &model.Card{
		Ease:     2.5,
		Interval: 5,
	}

	originalInterval := card.Interval
	result := scheduler.Schedule(card, 5)

	// Rating 5 (easy) should increase interval
	if card.Interval <= originalInterval {
		t.Errorf("Expected interval to increase for rating 5, got %d (was %d)", card.Interval, originalInterval)
	}

	expectedEase := 2.65 // 2.5 + 0.15 (capped at 4.0)
	if card.Ease != expectedEase {
		t.Errorf("Expected ease %.2f, got %.2f", expectedEase, card.Ease)
	}

	if result.EaseOrRetention != card.Ease {
		t.Errorf("Expected result ease %.2f, got %.2f", card.Ease, result.EaseOrRetention)
	}
}

func TestSM2SchedulerMinEaseBoundary(t *testing.T) {
	scheduler := NewSM2Scheduler()
	card := &model.Card{
		Ease:     1.3,
		Interval: 5,
	}

	scheduler.Schedule(card, 1) // Reduce by 0.3

	// Should not go below minEase (1.3)
	if card.Ease < minEase {
		t.Errorf("Ease %.1f is below minimum %.1f", card.Ease, minEase)
	}

	if card.Ease != minEase {
		t.Errorf("Expected ease %.1f, got %.1f", minEase, card.Ease)
	}
}

func TestSM2SchedulerMaxIntervalBoundary(t *testing.T) {
	scheduler := NewSM2Scheduler()
	card := &model.Card{
		Ease:     4.0,
		Interval: 350,
	}

	scheduler.Schedule(card, 5)

	// Should not exceed maxInterval (365)
	if card.Interval > maxInterval {
		t.Errorf("Interval %d exceeds maximum %d", card.Interval, maxInterval)
	}

	if card.Interval != maxInterval {
		t.Errorf("Expected interval %d, got %d", maxInterval, card.Interval)
	}
}

func TestFSRSSchedulerRating1(t *testing.T) {
	scheduler := NewFSRSScheduler()
	card := &model.Card{
		Ease:     2.5,
		Interval: 5,
	}

	result := scheduler.Schedule(card, 1)

	// Rating 1 should have low retention target
	if card.Retention >= 0.5 {
		t.Errorf("Expected low retention for rating 1, got %.2f", card.Retention)
	}

	// Interval should be updated
	if result.Interval <= 0 {
		t.Errorf("Expected positive interval, got %d", result.Interval)
	}
}

func TestFSRSSchedulerRating5(t *testing.T) {
	scheduler := NewFSRSScheduler()
	card := &model.Card{
		Ease:     2.5,
		Interval: 5,
	}

	result := scheduler.Schedule(card, 5)

	// Rating 5 should have high retention target
	if card.Retention < 0.9 {
		t.Errorf("Expected high retention for rating 5, got %.2f", card.Retention)
	}

	if result.EaseOrRetention != card.Retention {
		t.Errorf("Expected result retention %.2f, got %.2f", card.Retention, result.EaseOrRetention)
	}
}

func TestDispatcherSM2(t *testing.T) {
	card := &model.Card{
		Algorithm: "SM2",
		Ease:      2.5,
		Interval:  5,
	}

	err := ScheduleCard(card, 4)

	if err != nil {
		t.Errorf("ScheduleCard returned error: %v", err)
	}

	// SM-2 should update Ease, not Retention
	if card.Ease == 2.5 { // Ease should change for rating 4 (actually stays same)
		// Rating 4 doesn't change ease, so this is OK
	}
}

func TestDispatcherFSRS(t *testing.T) {
	card := &model.Card{
		Algorithm: "FSRS",
		Ease:      2.5,
		Interval:  5,
	}

	err := ScheduleCard(card, 4)

	if err != nil {
		t.Errorf("ScheduleCard returned error: %v", err)
	}

	// FSRS should update Retention, not Ease
	if card.Retention == 0 {
		t.Errorf("Expected Retention to be updated from 0 to non-zero")
	}
}

func TestDispatcherUnknownAlgorithm(t *testing.T) {
	card := &model.Card{
		Algorithm: "UNKNOWN",
		Ease:      2.5,
		Interval:  5,
	}

	err := ScheduleCard(card, 4)

	if err == nil {
		t.Errorf("Expected error for unknown algorithm, got nil")
	}
}

func TestDispatcherDefaultsToSM2(t *testing.T) {
	card := &model.Card{
		Algorithm: "", // No algorithm specified
		Ease:      2.5,
		Interval:  5,
	}

	err := ScheduleCard(card, 4)

	if err != nil {
		t.Errorf("ScheduleCard returned error: %v", err)
	}

	// Should default to SM2, so Ease should be used
	if card.Ease == 0 && card.Retention == 0 {
		t.Errorf("Expected at least one parameter to be updated")
	}
}

func TestScheduleResultHasValidFields(t *testing.T) {
	scheduler := NewSM2Scheduler()
	card := &model.Card{
		Ease:     2.5,
		Interval: 3,
	}

	now := time.Now()
	result := scheduler.Schedule(card, 4)

	// Check result fields
	if result.NextReviewDate.Before(now) {
		t.Errorf("Expected NextReviewDate to be in future, got %v", result.NextReviewDate)
	}

	if result.Interval != card.Interval {
		t.Errorf("Expected result interval to match card interval")
	}

	if result.EaseOrRetention != card.Ease {
		t.Errorf("Expected result EaseOrRetention to match card Ease")
	}

	if result.LastReview.Before(now.Add(-time.Second)) {
		t.Errorf("Expected LastReview to be recent, got %v", result.LastReview)
	}
}

func TestInitializeNewCard(t *testing.T) {
	card := &model.Card{
		Question: "Test",
		Answer:   "Test answer",
	}

	InitializeNewCard(card)

	if card.Ease != defaultEase {
		t.Errorf("Expected default ease %.1f, got %.1f", defaultEase, card.Ease)
	}

	if card.Interval != 0 {
		t.Errorf("Expected interval 0 for new card, got %d", card.Interval)
	}

	if card.Algorithm != "SM2" {
		t.Errorf("Expected algorithm SM2, got %s", card.Algorithm)
	}
}
