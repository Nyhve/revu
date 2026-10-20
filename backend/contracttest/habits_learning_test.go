package contracttest

import "testing"

func TestHabits(t *testing.T) {
	u := newUser(t)

	expectError(t, u.post("/api/habits", map[string]any{"name": "Lesen"}), 400, "frequency: Frequency must not be null")

	h := u.mustPost("/api/habits", map[string]any{"name": "Lesen", "frequency": "DAILY"}).json(t)
	hasKeys(t, h, "id", "name", "description", "frequency", "createdAt")
	hid := id(t, h)

	entry := u.mustPost(p("/api/habits/%d/entries", hid), nil).json(t)
	hasKeys(t, entry, "id", "habitId", "date", "createdAt")
	isDate(t, entry["date"])
	if int64(num(t, entry, "habitId")) != hid {
		t.Fatalf("wrong habitId: %v", entry)
	}
	expectError(t, u.post(p("/api/habits/%d/entries", hid), nil), 409, "Habit already completed today")

	todays := u.mustGet("/api/habits/entries/today").list(t)
	if len(todays) != 1 || id(t, todays[0]) != id(t, entry) {
		t.Fatalf("unexpected today entries: %v", todays)
	}
	if l := u.mustGet(p("/api/habits/%d/entries", hid)).list(t); len(l) != 1 {
		t.Fatalf("expected 1 entry, got %v", l)
	}

	cons := u.mustGet("/api/habits/consistency?days=28").list(t)
	if len(cons) != 1 {
		t.Fatalf("expected 1 consistency row: %v", cons)
	}
	hasKeys(t, cons[0], "habitId", "name", "frequency", "totalDays", "completedDays", "consistencyPercent")
	if num(t, cons[0], "totalDays") != 28 || num(t, cons[0], "completedDays") != 1 || num(t, cons[0], "consistencyPercent") != 3.6 {
		t.Fatalf("unexpected consistency: %v", cons[0])
	}

	weekly := u.mustPost("/api/habits", map[string]any{"name": "Sauna", "frequency": "WEEKLY", "description": "Fr"}).json(t)
	cons = u.mustGet("/api/habits/consistency").list(t)
	if len(cons) != 2 || num(t, cons[1], "totalDays") != 4 || num(t, cons[1], "consistencyPercent") != 0 {
		t.Fatalf("unexpected weekly consistency: %v", cons)
	}
	u.mustPost(p("/api/habits/%d/entries", id(t, weekly)), nil)

	activities := u.mustGet("/api/activities").list(t)
	if len(activities) != 2 || activities[0]["type"] != "HABIT_COMPLETED" {
		t.Fatalf("unexpected activities: %v", activities)
	}

	expect(t, u.del(p("/api/habits/%d/entries/%d", hid, id(t, entry))), 204)
	if l := u.mustGet(p("/api/habits/%d/entries", hid)).list(t); len(l) != 0 {
		t.Fatalf("entry not deleted: %v", l)
	}
	expect(t, u.del(p("/api/habits/%d", hid)), 204)
	expectError(t, u.get(p("/api/habits/%d/entries", hid)), 404, p("Habit not found with id: %d", hid))
}

// The entry id must belong to the habit in the URL (cross-user deletion was possible before).
func TestHabitEntryOwnership(t *testing.T) {
	fixed(t)
	victim := newUser(t)
	attacker := newUser(t)

	vh := victim.mustPost("/api/habits", map[string]any{"name": "Meditation", "frequency": "DAILY"}).json(t)
	ve := victim.mustPost(p("/api/habits/%d/entries", id(t, vh)), nil).json(t)
	ah := attacker.mustPost("/api/habits", map[string]any{"name": "Egal", "frequency": "DAILY"}).json(t)

	expect(t, attacker.del(p("/api/habits/%d/entries/%d", id(t, ah), id(t, ve))), 404)
	if l := victim.mustGet(p("/api/habits/%d/entries", id(t, vh))).list(t); len(l) != 1 {
		t.Fatalf("victim's entry was deleted")
	}
}

func TestLearning(t *testing.T) {
	u := newUser(t)

	expectError(t, u.post("/api/learning", map[string]any{"title": "Go"}), 400, "topic: Topic must not be blank")
	proj := u.mustPost("/api/learning", map[string]any{"title": "Go lernen", "topic": "Programmierung"}).json(t)
	hasKeys(t, proj, "id", "title", "description", "topic", "totalMinutes", "createdAt")
	if num(t, proj, "totalMinutes") != 0 {
		t.Fatalf("new project must have 0 minutes: %v", proj)
	}
	pid := id(t, proj)

	expectError(t, u.post(p("/api/learning/%d/sessions", pid), map[string]any{"durationMinutes": 0, "date": today()}),
		400, "durationMinutes: Duration must be at least 1 minute")
	expectError(t, u.post(p("/api/learning/%d/sessions", pid), map[string]any{"durationMinutes": 5}),
		400, "date: Date must not be null")

	s1 := u.mustPost(p("/api/learning/%d/sessions", pid), map[string]any{"durationMinutes": 45, "date": "2026-10-01", "notes": "Kapitel 1"}).json(t)
	hasKeys(t, s1, "id", "projectId", "durationMinutes", "notes", "date", "createdAt")
	u.mustPost(p("/api/learning/%d/sessions", pid), map[string]any{"durationMinutes": 30, "date": "2026-10-02"})

	projects := u.mustGet("/api/learning").list(t)
	if len(projects) != 1 || num(t, projects[0], "totalMinutes") != 75 {
		t.Fatalf("unexpected projects: %v", projects)
	}
	if l := u.mustGet(p("/api/learning/%d/sessions", pid)).list(t); len(l) != 2 {
		t.Fatalf("expected 2 sessions: %v", l)
	}

	acts := u.mustGet("/api/activities").list(t)
	if len(acts) != 2 || acts[0]["type"] != "LEARNING_SESSION_COMPLETED" || int64(num(t, acts[0], "referenceId")) != pid {
		t.Fatalf("unexpected activities: %v", acts)
	}

	expect(t, u.del(p("/api/learning/%d/sessions/%d", pid, id(t, s1))), 204)
	if projects = u.mustGet("/api/learning").list(t); num(t, projects[0], "totalMinutes") != 30 {
		t.Fatalf("total not updated: %v", projects)
	}
	expect(t, u.del(p("/api/learning/%d", pid)), 204)
	expectError(t, u.get(p("/api/learning/%d/sessions", pid)), 404, p("Learning project not found with id: %d", pid))
}

func TestLearningSessionOwnership(t *testing.T) {
	fixed(t)
	victim := newUser(t)
	attacker := newUser(t)
	vp := victim.mustPost("/api/learning", map[string]any{"title": "A", "topic": "B"}).json(t)
	vs := victim.mustPost(p("/api/learning/%d/sessions", id(t, vp)), map[string]any{"durationMinutes": 10, "date": today()}).json(t)
	ap := attacker.mustPost("/api/learning", map[string]any{"title": "C", "topic": "D"}).json(t)

	expect(t, attacker.del(p("/api/learning/%d/sessions/%d", id(t, ap), id(t, vs))), 404)
	if l := victim.mustGet(p("/api/learning/%d/sessions", id(t, vp))).list(t); len(l) != 1 {
		t.Fatalf("victim's session was deleted")
	}
}
