package contracttest

import "testing"

func TestGoalsCRUD(t *testing.T) {
	u := newUser(t)
	other := newUser(t)

	expect(t, u.get("/api/goals"), 200).list(t)

	g := u.mustPost("/api/goals", map[string]any{"title": "Marathon laufen", "targetDate": "2027-05-01"}).json(t)
	hasKeys(t, g, "id", "title", "description", "targetDate", "status", "createdAt", "updatedAt")
	if g["status"] != "ACTIVE" || g["description"] != nil || g["targetDate"] != "2027-05-01" {
		t.Fatalf("unexpected goal: %v", g)
	}
	isTimestamp(t, g["createdAt"])
	gid := id(t, g)

	if l := u.mustGet("/api/goals").list(t); len(l) != 1 {
		t.Fatalf("expected 1 goal, got %d", len(l))
	}
	u.mustGet(p("/api/goals/%d", gid))
	expectError(t, other.get(p("/api/goals/%d", gid)), 404, p("Goal not found with id: %d", gid))

	expectError(t, u.post("/api/goals", map[string]any{"title": "  "}), 400, "title: Title must not be blank")
	long := make([]byte, 256)
	for i := range long {
		long[i] = 'a'
	}
	expectError(t, u.post("/api/goals", map[string]any{"title": string(long)}), 400, "title: Title must not exceed 255 characters")
	expectError(t, u.put(p("/api/goals/%d", gid), map[string]any{"title": "x"}), 400, "status: Status must not be null")

	upd := expect(t, u.put(p("/api/goals/%d", gid), map[string]any{
		"title": "Halbmarathon", "description": "erst mal", "status": "COMPLETED",
	}), 200).json(t)
	if upd["title"] != "Halbmarathon" || upd["status"] != "COMPLETED" || upd["targetDate"] != nil || upd["description"] != "erst mal" {
		t.Fatalf("unexpected update result: %v", upd)
	}

	feed := u.mustGet("/api/activities").list(t)
	if len(feed) != 2 || feed[0]["type"] != "GOAL_COMPLETED" || feed[1]["type"] != "GOAL_CREATED" {
		t.Fatalf("unexpected activity feed: %v", feed)
	}
	hasKeys(t, feed[0], "id", "type", "referenceId", "createdAt")
	if int64(num(t, feed[0], "referenceId")) != gid {
		t.Fatalf("activity references wrong goal: %v", feed[0])
	}

	expectError(t, other.del(p("/api/goals/%d", gid)), 404, "Goal not found")
	expect(t, u.del(p("/api/goals/%d", gid)), 204)
	expectError(t, u.del(p("/api/goals/%d", gid)), 404, "Goal not found")
}

func TestTasksCRUD(t *testing.T) {
	u := newUser(t)

	task := u.mustPost("/api/tasks", map[string]any{"title": "Steuern", "description": "bis Juli", "dueDate": "2026-07-31"}).json(t)
	hasKeys(t, task, "id", "title", "description", "dueDate", "completed", "createdAt", "updatedAt")
	if task["completed"] != false {
		t.Fatalf("new task must be open: %v", task)
	}
	tid := id(t, task)

	expectError(t, u.put(p("/api/tasks/%d", tid), map[string]any{"title": "Steuern"}), 400, "completed: Completed must not be null")
	done := expect(t, u.put(p("/api/tasks/%d", tid), map[string]any{"title": "Steuern", "completed": true}), 200).json(t)
	if done["completed"] != true || done["dueDate"] != nil {
		t.Fatalf("unexpected task: %v", done)
	}
	// completing again must not create a second TASK_COMPLETED activity
	expect(t, u.put(p("/api/tasks/%d", tid), map[string]any{"title": "Steuern", "completed": true}), 200)

	counts := map[string]int{}
	for _, a := range u.mustGet("/api/activities").list(t) {
		counts[a["type"].(string)]++
	}
	if counts["TASK_CREATED"] != 1 || counts["TASK_COMPLETED"] != 1 {
		t.Fatalf("unexpected activities: %v", counts)
	}

	daily := u.mustGet("/api/activities/daily?days=30").list(t)
	if len(daily) != 1 || num(t, daily[0], "count") != 2 {
		t.Fatalf("unexpected daily counts: %v", daily)
	}
	hasKeys(t, daily[0], "date", "count")
	isDate(t, daily[0]["date"])

	expectError(t, u.get("/api/tasks/999999999"), 404, "Task not found with id: 999999999")
	expect(t, u.del(p("/api/tasks/%d", tid)), 204)
}

func TestMalformedRequests(t *testing.T) {
	u := newUser(t)
	expect(t, u.post("/api/goals", "{not json"), 400)
	expect(t, u.post("/api/goals", ""), 400)
	expect(t, u.get("/api/goals/abc"), 400)
	expect(t, u.post("/api/habits", map[string]any{"name": "x", "frequency": "HOURLY"}), 400)
}
