package core

import (
	"math/rand/v2"
	"slices"
	"time"

	"footprints/models/db"
)

// slots возвращает слоты таски на сегодняшний день now: точки сетки с шагом
// step внутри окна [start_time, end_time), не раньше текущего опроса (now,
// усечённого до шага). Если сегодня не день запуска таски, слотов нет
func slots(t db.Task, now time.Time, step time.Duration) ([]time.Time, error) {
	day := int16(now.Weekday()) // ISO: понедельник 1, воскресенье 7
	if day == 0 {
		day = 7
	}
	if !slices.Contains(t.Days, day) {
		return nil, nil
	}

	start, err := t.StartTime.At(now)
	if err != nil {
		return nil, err
	}
	end, err := t.EndTime.At(now)
	if err != nil {
		return nil, err
	}

	first := ceilTo(start, step)
	if cur := now.Truncate(step); cur.After(first) {
		first = cur
	}
	var out []time.Time
	for s := first; s.Before(end); s = s.Add(step) {
		out = append(out, s)
	}
	return out, nil
}

// ceilTo округляет время вверх до точки сетки
func ceilTo(t time.Time, step time.Duration) time.Time {
	r := t.Truncate(step)
	if r.Before(t) {
		r = r.Add(step)
	}
	return r
}

// pick выбирает n разных слотов случайно и отдаёт их по возрастанию. Если
// слотов не больше n, отдаёт все
func pick(slots []time.Time, n int) []time.Time {
	if n >= len(slots) {
		return slices.Clone(slots)
	}
	idx := rand.Perm(len(slots))[:n]
	slices.Sort(idx)
	out := make([]time.Time, n)
	for i, j := range idx {
		out[i] = slots[j]
	}
	return out
}

// overdueReason возвращает причину, по которой запись уже нельзя выполнять
// (она из прошлых суток или окно таски сегодня закрыто), или пустую строку,
// если запись ещё актуальна
func overdueReason(t db.Task, plannedAt, now time.Time) string {
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if plannedAt.Before(dayStart) {
		return "просрочена: запись из прошлых суток"
	}
	end, err := t.EndTime.At(now)
	if err != nil {
		return "просрочена: некорректное время окна таски"
	}
	if !now.Before(end) {
		return "просрочена: окно закрыто"
	}
	return ""
}
