package model_test

import (
	"testing"
	"time"

	"github.com/TTitcombe/questlog/internal/model"
)

func timePtr(t time.Time) *time.Time { return &t }

func TestHasCycle_NoCycle(t *testing.T) {
	tracks := []model.Track{
		{Name: "a", DependsOn: []string{}},
		{Name: "b", DependsOn: []string{"a"}},
		{Name: "c", DependsOn: []string{"a"}},
	}
	if model.HasCycle(tracks) {
		t.Fatal("expected no cycle")
	}
}

func TestHasCycle_WithCycle(t *testing.T) {
	tracks := []model.Track{
		{Name: "a", DependsOn: []string{"b"}},
		{Name: "b", DependsOn: []string{"a"}},
	}
	if !model.HasCycle(tracks) {
		t.Fatal("expected cycle to be detected")
	}
}

func TestHasCycle_SelfLoop(t *testing.T) {
	tracks := []model.Track{
		{Name: "a", DependsOn: []string{"a"}},
	}
	if !model.HasCycle(tracks) {
		t.Fatal("expected self-loop to be detected as cycle")
	}
}

func TestHasCycle_SingleNodeNoCycle(t *testing.T) {
	tracks := []model.Track{
		{Name: "a", DependsOn: []string{}},
	}
	if model.HasCycle(tracks) {
		t.Fatal("expected single node with no deps to have no cycle")
	}
}

func TestHasCycle_ThreeNodeCycle(t *testing.T) {
	tracks := []model.Track{
		{Name: "a", DependsOn: []string{"c"}},
		{Name: "b", DependsOn: []string{"a"}},
		{Name: "c", DependsOn: []string{"b"}},
	}
	if !model.HasCycle(tracks) {
		t.Fatal("expected three-node cycle to be detected")
	}
}

func TestComputeGoalProgress_ParallelTracks(t *testing.T) {
	goal := model.Goal{Slug: "g", Title: "G"}
	done := timePtr(time.Now())
	tracks := []model.Track{
		{
			Name: "maths", GoalSlug: "g", DependsOn: []string{},
			Milestones: []model.Milestone{{ID: "m1", Description: "done", CompletedAt: done}},
		},
		{Name: "coding", GoalSlug: "g", DependsOn: []string{}},
	}
	resources := map[string][]model.Resource{
		"maths":  {{IsCore: true, Status: model.StatusDone}},
		"coding": {{IsCore: true, Status: model.StatusUnread}},
	}

	gp := model.ComputeGoalProgress(goal, tracks, resources)

	if gp.TracksComplete != 1 {
		t.Errorf("expected 1 track complete, got %d", gp.TracksComplete)
	}
	if gp.TracksTotal != 2 {
		t.Errorf("expected 2 tracks total, got %d", gp.TracksTotal)
	}

	for _, tp := range gp.Tracks {
		if tp.Track.Name == "maths" && tp.Status != model.TrackStatusComplete {
			t.Errorf("maths should be complete")
		}
		if tp.Track.Name == "coding" && tp.Status != model.TrackStatusAvailable {
			t.Errorf("coding should be available, got %v", tp.Status)
		}
	}
}

func TestComputeGoalProgress_LockedByDep(t *testing.T) {
	goal := model.Goal{Slug: "g", Title: "G"}
	tracks := []model.Track{
		{Name: "a", GoalSlug: "g", DependsOn: []string{}},
		{Name: "b", GoalSlug: "g", DependsOn: []string{"a"}},
	}
	resources := map[string][]model.Resource{
		"a": {{IsCore: true, Status: model.StatusUnread}},
		"b": {{IsCore: true, Status: model.StatusUnread}},
	}

	gp := model.ComputeGoalProgress(goal, tracks, resources)

	for _, tp := range gp.Tracks {
		if tp.Track.Name == "b" && tp.Status != model.TrackStatusLocked {
			t.Errorf("b should be locked while a is incomplete, got %v", tp.Status)
		}
	}
}

func TestComputeGoalProgress_UnlockedAfterDepComplete(t *testing.T) {
	goal := model.Goal{Slug: "g", Title: "G"}
	done := timePtr(time.Now())
	tracks := []model.Track{
		{
			Name: "a", GoalSlug: "g", DependsOn: []string{},
			Milestones: []model.Milestone{{ID: "m1", Description: "x", CompletedAt: done}},
		},
		{Name: "b", GoalSlug: "g", DependsOn: []string{"a"}},
	}
	resources := map[string][]model.Resource{
		"a": {{IsCore: true, Status: model.StatusDone}},
		"b": {{IsCore: true, Status: model.StatusUnread}},
	}

	gp := model.ComputeGoalProgress(goal, tracks, resources)

	for _, tp := range gp.Tracks {
		if tp.Track.Name == "b" && tp.Status != model.TrackStatusAvailable {
			t.Errorf("b should be available once a is complete, got %v", tp.Status)
		}
	}
}

func TestComputeGoalProgress_NoSignalsNeverComplete(t *testing.T) {
	goal := model.Goal{Slug: "g", Title: "G"}
	tracks := []model.Track{
		{Name: "empty", GoalSlug: "g", DependsOn: []string{}},
	}
	resources := map[string][]model.Resource{"empty": {}}

	gp := model.ComputeGoalProgress(goal, tracks, resources)

	if gp.TracksComplete != 0 {
		t.Error("track with no signals should never be complete")
	}
}
