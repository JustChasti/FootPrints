package db

import (
	"slices"
	"testing"
)

func validParams() TaskParams {
	return TaskParams{
		RepoURL:       "https://bitbucket.org/team/repo",
		Branch:        "main",
		CommitName:    "PROJ-1 work",
		Days:          Weekdays{1, 2, 3, 4, 5},
		StartTime:     "09:00",
		EndTime:       "18:00",
		CommitsPerDay: 1,
		Active:        true,
	}
}

func TestValidateRejects(t *testing.T) {
	cases := map[string]func(p *TaskParams){
		"repo_url empty":       func(p *TaskParams) { p.RepoURL = "" },
		"repo_url no scheme":   func(p *TaskParams) { p.RepoURL = "bitbucket.org/team/repo" },
		"repo_url ftp":         func(p *TaskParams) { p.RepoURL = "ftp://bitbucket.org/team/repo" },
		"branch empty":         func(p *TaskParams) { p.Branch = "  " },
		"commit_name empty":    func(p *TaskParams) { p.CommitName = "" },
		"days empty":           func(p *TaskParams) { p.Days = nil },
		"day 0":                func(p *TaskParams) { p.Days = Weekdays{0, 1} },
		"day 8":                func(p *TaskParams) { p.Days = Weekdays{1, 8} },
		"day repeated":         func(p *TaskParams) { p.Days = Weekdays{1, 3, 1} },
		"start not a time":     func(p *TaskParams) { p.StartTime = "abc" },
		"start 25:00":          func(p *TaskParams) { p.StartTime = "25:00" },
		"end 18:60":            func(p *TaskParams) { p.EndTime = "18:60" },
		"end empty":            func(p *TaskParams) { p.EndTime = "" },
		"start after end":      func(p *TaskParams) { p.StartTime, p.EndTime = "18:00", "09:00" },
		"start equals end":     func(p *TaskParams) { p.StartTime, p.EndTime = "09:00", "09:00" },
		"commits_per_day zero": func(p *TaskParams) { p.CommitsPerDay = 0 },
		"commits_per_day < 0":  func(p *TaskParams) { p.CommitsPerDay = -3 },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			p := validParams()
			mutate(&p)
			if err := p.Validate(); err == nil {
				t.Fatalf("expected an error, got nil for %+v", p)
			}
		})
	}
}

func TestValidateAccepts(t *testing.T) {
	p := validParams()
	if err := p.Validate(); err != nil {
		t.Fatalf("valid params rejected: %v", err)
	}
}

func TestValidateNormalizes(t *testing.T) {
	p := validParams()
	p.RepoURL = "  https://bitbucket.org/team/repo "
	p.Branch = " dev "
	p.Days = Weekdays{5, 1, 3}
	p.StartTime = "9:00"
	p.EndTime = "18:05"
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.RepoURL != "https://bitbucket.org/team/repo" || p.Branch != "dev" {
		t.Errorf("strings are not trimmed: %q %q", p.RepoURL, p.Branch)
	}
	if !slices.Equal(p.Days, Weekdays{1, 3, 5}) {
		t.Errorf("days are not sorted: %v", p.Days)
	}
	if p.StartTime != "09:00" || p.EndTime != "18:05" {
		t.Errorf("time is not normalized: %q %q", p.StartTime, p.EndTime)
	}
}
