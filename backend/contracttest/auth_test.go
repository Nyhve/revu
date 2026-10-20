package contracttest

import "testing"

func TestRegisterAndLogin(t *testing.T) {
	requireServer(t)
	a := anon(t)
	email := randomEmail()

	reg := expect(t, a.post("/api/auth/register", map[string]string{"email": email, "password": "password123"}), 201).json(t)
	hasKeys(t, reg, "id", "email", "createdAt")
	if reg["email"] != email {
		t.Fatalf("email mismatch: %v", reg)
	}
	isTimestamp(t, reg["createdAt"])

	expectError(t, a.post("/api/auth/register", map[string]string{"email": email, "password": "password123"}),
		409, "Email already registered: "+email)

	login := expect(t, a.post("/api/auth/login", map[string]string{"email": email, "password": "password123"}), 200).json(t)
	hasKeys(t, login, "token")

	expectError(t, a.post("/api/auth/login", map[string]string{"email": email, "password": "wrong-password"}),
		401, "Invalid email or password")
	expectError(t, a.post("/api/auth/login", map[string]string{"email": randomEmail(), "password": "whatever1"}),
		401, "Invalid email or password")
}

func TestRegisterValidation(t *testing.T) {
	requireServer(t)
	a := anon(t)
	expectError(t, a.post("/api/auth/register", map[string]string{"email": "not-an-email", "password": "password123"}),
		400, "email: Email must be valid")
	expectError(t, a.post("/api/auth/register", map[string]string{"email": randomEmail(), "password": "short"}),
		400, "password: Password must be at least 8 characters long")
	expectError(t, a.post("/api/auth/register", map[string]string{"email": "", "password": ""}),
		400, "Email must not be blank")
	expectError(t, a.post("/api/auth/login", map[string]string{"email": "", "password": "x"}),
		400, "email: Email must not be blank")
}

func TestProtectedEndpointsRequireToken(t *testing.T) {
	requireServer(t)
	for _, path := range []string{"/api/goals", "/api/tasks", "/api/habits", "/api/workouts",
		"/api/learning", "/api/finance", "/api/activities", "/api/user/preferences", "/api/exercises"} {
		r := anon(t).get(path)
		if legacy {
			if r.Status != 401 && r.Status != 403 {
				t.Fatalf("%s: expected 401/403, got %d", path, r.Status)
			}
			continue
		}
		expectError(t, r, 401, "Authentication required")
	}
	bad := &client{t: t, token: "not.a.jwt"}
	r := bad.get("/api/goals")
	if legacy {
		if r.Status != 401 && r.Status != 403 {
			t.Fatalf("expected 401/403 for invalid token, got %d", r.Status)
		}
		return
	}
	expectError(t, r, 401, "Invalid or expired token")
}

func TestPreferences(t *testing.T) {
	u := newUser(t)
	prefs := u.mustGet("/api/user/preferences").json(t)
	hasKeys(t, prefs, "goalsEnabled", "tasksEnabled", "habitsEnabled", "fitnessEnabled", "learningEnabled", "financeEnabled")
	for k, v := range prefs {
		if v != true {
			t.Fatalf("default %s should be true, got %v", k, v)
		}
	}

	update := map[string]bool{
		"goalsEnabled": true, "tasksEnabled": false, "habitsEnabled": true,
		"fitnessEnabled": false, "learningEnabled": true, "financeEnabled": false,
	}
	got := expect(t, u.put("/api/user/preferences", update), 200).json(t)
	for k, v := range update {
		if got[k] != v {
			t.Fatalf("%s: want %v got %v", k, v, got[k])
		}
	}
	again := u.mustGet("/api/user/preferences").json(t)
	if again["tasksEnabled"] != false || again["goalsEnabled"] != true {
		t.Fatalf("preferences not persisted: %v", again)
	}
}
