package contracttest

import (
	"testing"
	"time"
)

func TestFinance(t *testing.T) {
	u := newUser(t)
	now := time.Now()
	thisMonth := now.Format("2006-01") + "-01"

	expectError(t, u.post("/api/finance", map[string]any{
		"amount": 0, "description": "x", "category": "FOOD", "type": "EXPENSE", "date": thisMonth,
	}), 400, "amount: Amount must be greater than 0")

	salary := u.mustPost("/api/finance", map[string]any{
		"amount": 2500.5, "description": "Gehalt", "category": "SALARY", "type": "INCOME", "date": thisMonth,
	}).json(t)
	hasKeys(t, salary, "id", "amount", "description", "category", "type", "date", "createdAt")
	if num(t, salary, "amount") != 2500.5 {
		t.Fatalf("amount must be a JSON number: %s", salary["amount"])
	}
	food := u.mustPost("/api/finance", map[string]any{
		"amount": 120.25, "description": "Einkauf", "category": "FOOD", "type": "EXPENSE", "date": thisMonth,
	}).json(t)

	list := u.mustGet("/api/finance").list(t)
	if len(list) != 2 {
		t.Fatalf("expected 2 transactions: %v", list)
	}

	stats := u.mustGet(p("/api/finance/stats/%d/%d", now.Year(), int(now.Month()))).json(t)
	hasKeys(t, stats, "year", "month", "totalIncome", "totalExpenses", "balance")
	if num(t, stats, "totalIncome") != 2500.5 || num(t, stats, "totalExpenses") != 120.25 || num(t, stats, "balance") != 2380.25 {
		t.Fatalf("unexpected stats: %v", stats)
	}
	empty := u.mustGet("/api/finance/stats/2001/2").json(t)
	if num(t, empty, "totalIncome") != 0 || num(t, empty, "balance") != 0 {
		t.Fatalf("unexpected empty stats: %v", empty)
	}

	summary := u.mustGet("/api/finance/summary?months=6").list(t)
	if len(summary) != 6 {
		t.Fatalf("expected 6 months: %v", summary)
	}
	last := summary[5]
	hasKeys(t, last, "year", "month", "monthLabel", "income", "expenses")
	if int(num(t, last, "month")) != int(now.Month()) || num(t, last, "income") != 2500.5 || num(t, last, "expenses") != 120.25 {
		t.Fatalf("unexpected current month summary: %v", last)
	}
	labels := [12]string{"Jan", "Feb", "Mär", "Apr", "Mai", "Jun", "Jul", "Aug", "Sep", "Okt", "Nov", "Dez"}
	if want := p("%s %d", labels[now.Month()-1], now.Year()); last["monthLabel"] != want {
		t.Fatalf("label: want %q got %v", want, last["monthLabel"])
	}
	if num(t, summary[0], "income") != 0 {
		t.Fatalf("oldest month should be empty: %v", summary[0])
	}

	expect(t, u.del(p("/api/finance/%d", id(t, food))), 204)
	expectError(t, u.del(p("/api/finance/%d", id(t, food))), 404, "Transaction not found with id")
}

func TestFinanceInvalidMonth(t *testing.T) {
	fixed(t) // was a 500 before
	u := newUser(t)
	expect(t, u.get("/api/finance/stats/2026/13"), 400)
}

func TestFitness(t *testing.T) {
	u := newUser(t)

	name := randomName("Kniebeuge")
	ex := u.mustPost("/api/exercises", map[string]any{"name": name, "muscleGroup": "LEGS"}).json(t)
	hasKeys(t, ex, "id", "name", "muscleGroup")
	expectError(t, u.post("/api/exercises", map[string]any{"name": name, "muscleGroup": "LEGS"}), 409, "Exercise already exists with name: "+name)
	expectError(t, u.post("/api/exercises", map[string]any{"name": "x"}), 400, "muscleGroup: Muscle group must not be null")
	found := false
	for _, e := range u.mustGet("/api/exercises").list(t) {
		if e["name"] == name {
			found = true
		}
	}
	if !found {
		t.Fatal("created exercise missing from list")
	}

	w := u.mustPost("/api/workouts", map[string]any{"name": "Beintag", "date": today()}).json(t)
	hasKeys(t, w, "id", "name", "date", "notes", "completed", "exercises", "createdAt")
	if len(w["exercises"].([]any)) != 0 || w["completed"] != false {
		t.Fatalf("unexpected new workout: %v", w)
	}
	wid := id(t, w)

	w = expect(t, u.post(p("/api/workouts/%d/exercises", wid), map[string]any{"exerciseId": id(t, ex)}), 200).json(t)
	wes := w["exercises"].([]any)
	if len(wes) != 1 {
		t.Fatalf("expected one exercise: %v", w)
	}
	we := wes[0].(map[string]any)
	hasKeys(t, we, "id", "exercise", "orderIndex", "sets")
	weID := id(t, we)
	expectError(t, u.post(p("/api/workouts/%d/exercises", wid), map[string]any{"exerciseId": 999999999}), 404, "Exercise not found with id: 999999999")

	w = expect(t, u.post(p("/api/workouts/%d/exercises/%d/sets", wid, weID), map[string]any{"weightKg": 80, "reps": 8}), 201).json(t)
	set := w["exercises"].([]any)[0].(map[string]any)["sets"].([]any)[0].(map[string]any)
	hasKeys(t, set, "id", "setNumber", "weightKg", "reps", "completedAt")
	if num(t, set, "setNumber") != 1 || num(t, set, "weightKg") != 80 || num(t, set, "reps") != 8 {
		t.Fatalf("unexpected set: %v", set)
	}
	expectError(t, u.post(p("/api/workouts/%d/exercises/%d/sets", wid, weID), map[string]any{"weightKg": 80, "reps": 0}), 400, "reps: Reps must be at least 1")
	expectError(t, u.post(p("/api/workouts/%d/exercises/%d/sets", wid, weID), map[string]any{"reps": 5}), 400, "weightKg: Weight must not be null")

	u.mustPost(p("/api/workouts/%d/exercises/%d/sets", wid, weID), map[string]any{"weightKg": 85.5, "reps": 6})
	w = u.mustGet(p("/api/workouts/%d", wid)).json(t)
	sets := w["exercises"].([]any)[0].(map[string]any)["sets"].([]any)
	if len(sets) != 2 || num(t, sets[1].(map[string]any), "setNumber") != 2 || num(t, sets[1].(map[string]any), "weightKg") != 85.5 {
		t.Fatalf("unexpected sets: %v", sets)
	}

	w = expect(t, u.post(p("/api/workouts/%d/complete", wid), nil), 200).json(t)
	if w["completed"] != true {
		t.Fatalf("workout not completed: %v", w)
	}
	expectError(t, u.post(p("/api/workouts/%d/complete", wid), nil), 409, "Workout is already completed and cannot be modified")
	expectError(t, u.post(p("/api/workouts/%d/exercises/%d/sets", wid, weID), map[string]any{"weightKg": 1, "reps": 1}), 409, "already completed")

	acts := u.mustGet("/api/activities").list(t)
	if len(acts) != 1 || acts[0]["type"] != "WORKOUT_COMPLETED" {
		t.Fatalf("unexpected activities: %v", acts)
	}
	if l := u.mustGet("/api/workouts").list(t); len(l) != 1 {
		t.Fatalf("expected one workout: %v", l)
	}

	other := newUser(t)
	expectError(t, other.get(p("/api/workouts/%d", wid)), 404, p("Workout not found with id: %d", wid))
	expect(t, u.del(p("/api/workouts/%d", wid)), 204)
}

// Each workout exercise must appear exactly once, regardless of its set count
// (the Spring backend repeated it once per set - the "updates all at once" bug).
func TestWorkoutExercisesAreNotDuplicated(t *testing.T) {
	fixed(t)
	u := newUser(t)
	ex := u.mustPost("/api/exercises", map[string]any{"name": randomName("Bankdrücken"), "muscleGroup": "CHEST"}).json(t)
	w := u.mustPost("/api/workouts", map[string]any{"name": "Push", "date": today()}).json(t)
	wid := id(t, w)
	w = expect(t, u.post(p("/api/workouts/%d/exercises", wid), map[string]any{"exerciseId": id(t, ex)}), 200).json(t)
	weID := id(t, w["exercises"].([]any)[0].(map[string]any))
	for i := 0; i < 3; i++ {
		u.mustPost(p("/api/workouts/%d/exercises/%d/sets", wid, weID), map[string]any{"weightKg": 60, "reps": 10})
	}
	for _, path := range []string{p("/api/workouts/%d", wid), "/api/workouts"} {
		body := u.mustGet(path)
		var exercises []any
		if path == "/api/workouts" {
			exercises = body.list(t)[0]["exercises"].([]any)
		} else {
			exercises = body.json(t)["exercises"].([]any)
		}
		if len(exercises) != 1 {
			t.Fatalf("%s: exercise listed %d times", path, len(exercises))
		}
	}
}

// Deleting a set renumbers the remaining sets 1..n.
func TestWorkoutSetRenumbering(t *testing.T) {
	fixed(t)
	u := newUser(t)
	ex := u.mustPost("/api/exercises", map[string]any{"name": randomName("Rudern"), "muscleGroup": "BACK"}).json(t)
	wid := id(t, u.mustPost("/api/workouts", map[string]any{"name": "Pull", "date": today()}).json(t))
	w := expect(t, u.post(p("/api/workouts/%d/exercises", wid), map[string]any{"exerciseId": id(t, ex)}), 200).json(t)
	weID := id(t, w["exercises"].([]any)[0].(map[string]any))
	var setIDs []int64
	for i := 0; i < 3; i++ {
		w = u.mustPost(p("/api/workouts/%d/exercises/%d/sets", wid, weID), map[string]any{"weightKg": 50 + i, "reps": 10}).json(t)
		sets := w["exercises"].([]any)[0].(map[string]any)["sets"].([]any)
		setIDs = append(setIDs, id(t, sets[len(sets)-1].(map[string]any)))
	}
	w = expect(t, u.del(p("/api/workouts/%d/exercises/%d/sets/%d", wid, weID, setIDs[1])), 200).json(t)
	sets := w["exercises"].([]any)[0].(map[string]any)["sets"].([]any)
	if len(sets) != 2 || num(t, sets[1].(map[string]any), "setNumber") != 2 {
		t.Fatalf("sets not renumbered: %v", sets)
	}
	w = u.mustPost(p("/api/workouts/%d/exercises/%d/sets", wid, weID), map[string]any{"weightKg": 70, "reps": 5}).json(t)
	sets = w["exercises"].([]any)[0].(map[string]any)["sets"].([]any)
	if num(t, sets[2].(map[string]any), "setNumber") != 3 {
		t.Fatalf("new set should be number 3: %v", sets)
	}
	w = expect(t, u.del(p("/api/workouts/%d/exercises/%d", wid, weID)), 200).json(t)
	if len(w["exercises"].([]any)) != 0 {
		t.Fatalf("exercise not removed: %v", w)
	}
}
