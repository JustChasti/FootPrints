package core

import (
	"slices"
	"testing"
	"time"

	"footprints/models/db"
)

const step = 10 * time.Minute

// 14 октября 2026 года это среда (ISO-день 3)
func at(h, m, s int) time.Time {
	return time.Date(2026, 10, 14, h, m, s, 0, time.UTC)
}

func task(start, end db.Clock, days ...int16) db.Task {
	return db.Task{TaskParams: db.TaskParams{
		Days:          days,
		StartTime:     start,
		EndTime:       end,
		CommitsPerDay: 1,
		Active:        true,
	}}
}

func TestSlots(t *testing.T) {
	workdays := []int16{1, 2, 3, 4, 5}

	cases := []struct {
		name      string
		task      db.Task
		now       time.Time
		wantCount int
		wantFirst time.Time
		wantLast  time.Time
	}{
		{"before the window gives the whole window", task("09:00", "18:00", workdays...), at(7, 55, 0), 54, at(9, 0, 0), at(17, 50, 0)},
		{"inside the window starts from the current tick", task("09:00", "18:00", workdays...), at(12, 34, 56), 33, at(12, 30, 0), at(17, 50, 0)},
		{"tick exactly on the grid point", task("09:00", "18:00", workdays...), at(12, 30, 0), 33, at(12, 30, 0), at(17, 50, 0)},
		{"last slot of the window", task("09:00", "18:00", workdays...), at(17, 50, 3), 1, at(17, 50, 0), at(17, 50, 0)},
		{"start is off the grid, rounded up", task("09:05", "18:00", workdays...), at(7, 0, 0), 53, at(9, 10, 0), at(17, 50, 0)},
		{"end is off the grid, last slot before it", task("09:00", "17:55", workdays...), at(7, 0, 0), 54, at(9, 0, 0), at(17, 50, 0)},
		{"end is off the grid, slot after it is dropped", task("09:00", "17:45", workdays...), at(7, 0, 0), 53, at(9, 0, 0), at(17, 40, 0)},
		{"window shorter than the step has one slot", task("09:00", "09:05", workdays...), at(7, 0, 0), 1, at(9, 0, 0), at(9, 0, 0)},
		{"window between two grid points has no slots", task("09:03", "09:08", workdays...), at(7, 0, 0), 0, time.Time{}, time.Time{}},
		{"window is closed", task("09:00", "18:00", workdays...), at(18, 0, 1), 0, time.Time{}, time.Time{}},
		{"long after the window", task("09:00", "18:00", workdays...), at(23, 50, 0), 0, time.Time{}, time.Time{}},
		{"not the day of the task", task("09:00", "18:00", 1, 2), at(7, 55, 0), 0, time.Time{}, time.Time{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := slots(c.task, c.now, step)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != c.wantCount {
				t.Fatalf("got %d slots, want %d: %v", len(got), c.wantCount, got)
			}
			if c.wantCount == 0 {
				return
			}
			if !got[0].Equal(c.wantFirst) || !got[len(got)-1].Equal(c.wantLast) {
				t.Errorf("first %v last %v, want %v and %v", got[0], got[len(got)-1], c.wantFirst, c.wantLast)
			}
			for i, s := range got {
				if s.Truncate(step) != s {
					t.Errorf("slot %v is not on the grid", s)
				}
				if i > 0 && s.Sub(got[i-1]) != step {
					t.Errorf("slots %v and %v are not one step apart", got[i-1], s)
				}
			}
		})
	}
}

func TestSlotsSunday(t *testing.T) {
	// воскресенье в Go это 0, в ISO это 7
	sunday := time.Date(2026, 10, 18, 7, 0, 0, 0, time.UTC)
	got, err := slots(task("09:00", "10:00", 7), sunday, step)
	if err != nil || len(got) != 6 {
		t.Fatalf("got %d slots and error %v, want 6 and nil", len(got), err)
	}
	got, err = slots(task("09:00", "10:00", 1, 2, 3, 4, 5, 6), sunday, step)
	if err != nil || len(got) != 0 {
		t.Fatalf("got %d slots and error %v, want 0 and nil", len(got), err)
	}
}

func TestSlotsBrokenWindow(t *testing.T) {
	if _, err := slots(task("abc", "18:00", 3), at(7, 0, 0), step); err == nil {
		t.Error("expected an error for a broken start_time")
	}
	if _, err := slots(task("09:00", "", 3), at(7, 0, 0), step); err == nil {
		t.Error("expected an error for a broken end_time")
	}
}

func TestPick(t *testing.T) {
	all := make([]time.Time, 20)
	for i := range all {
		all[i] = at(9, 0, 0).Add(time.Duration(i) * step)
	}

	for _, n := range []int{0, 1, 3, 19} {
		for range 1000 {
			got := pick(all, n)
			if len(got) != n {
				t.Fatalf("n=%d: got %d slots", n, len(got))
			}
			if !slices.IsSortedFunc(got, func(a, b time.Time) int { return a.Compare(b) }) {
				t.Fatalf("n=%d: not sorted: %v", n, got)
			}
			for i, s := range got {
				if !slices.ContainsFunc(all, s.Equal) {
					t.Fatalf("n=%d: %v is not among the slots", n, s)
				}
				if i > 0 && s.Equal(got[i-1]) {
					t.Fatalf("n=%d: slot %v is repeated", n, s)
				}
			}
		}
	}

	for _, n := range []int{20, 21, 100} {
		got := pick(all, n)
		if !slices.EqualFunc(got, all, time.Time.Equal) {
			t.Errorf("n=%d: want all %d slots, got %d", n, len(all), len(got))
		}
	}
	if got := pick(nil, 3); len(got) != 0 {
		t.Errorf("no slots: got %v", got)
	}
}

func TestPickIsRandom(t *testing.T) {
	all := make([]time.Time, 54)
	for i := range all {
		all[i] = at(9, 0, 0).Add(time.Duration(i) * step)
	}
	seen := map[time.Time]bool{}
	for range 1000 {
		seen[pick(all, 1)[0]] = true
	}
	if len(seen) < 40 {
		t.Errorf("only %d of %d slots were ever picked", len(seen), len(all))
	}
}

func TestOverdueReason(t *testing.T) {
	win := task("09:00", "18:00", 1, 2, 3, 4, 5)

	cases := []struct {
		name      string
		task      db.Task
		plannedAt time.Time
		now       time.Time
		overdue   bool
	}{
		{"fresh record inside the window", win, at(10, 0, 0), at(10, 0, 5), false},
		{"late but the window is still open", win, at(10, 30, 0), at(12, 0, 0), false},
		{"last minute of the window", win, at(17, 50, 0), at(17, 59, 59), false},
		{"window closed exactly now", win, at(17, 50, 0), at(18, 0, 0), true},
		{"window closed long ago", win, at(17, 50, 0), at(23, 0, 0), true},
		{"record from yesterday", win, at(12, 0, 0).AddDate(0, 0, -1), at(9, 10, 0), true},
		{"record from yesterday at the end of the day", win, at(23, 50, 0).AddDate(0, 0, -1), at(0, 0, 5), true},
		{"broken end_time", task("09:00", "abc", 3), at(10, 0, 0), at(10, 0, 5), true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reason := overdueReason(c.task, c.plannedAt, c.now)
			if (reason != "") != c.overdue {
				t.Errorf("reason %q, overdue want %v", reason, c.overdue)
			}
		})
	}
}
