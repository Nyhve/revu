// Package fitness implements exercises (/api/exercises) and workouts with
// exercises and sets (/api/workouts).
package fitness

import (
	"github.com/gin-gonic/gin"

	"github.com/Nyhve/revu/backend/internal/activity"
	"github.com/Nyhve/revu/backend/internal/app"
	"github.com/Nyhve/revu/backend/internal/auth"
	"github.com/Nyhve/revu/backend/internal/database"
	"github.com/Nyhve/revu/backend/internal/decimal"
	"github.com/Nyhve/revu/backend/internal/httpx"
	"github.com/Nyhve/revu/backend/internal/store"
)

var muscleGroups = []string{"CHEST", "BACK", "LEGS", "SHOULDERS", "ARMS", "CORE", "FULL_BODY"}

// ---- requests ----

type CreateExerciseRequest struct {
	Name        *string `json:"name"`
	MuscleGroup *string `json:"muscleGroup"`
}

func (r CreateExerciseRequest) Validate() error {
	var v httpx.Validator
	v.NotBlank("name", r.Name, "Name must not be blank")
	v.MaxLen("name", r.Name, 255, "Name must not exceed 255 characters")
	if v.Required("muscleGroup", r.MuscleGroup != nil, "Muscle group must not be null") {
		v.OneOf("muscleGroup", r.MuscleGroup, muscleGroups)
	}
	return v.Err()
}

type CreateWorkoutRequest struct {
	Name  *string     `json:"name"`
	Date  *httpx.Date `json:"date"`
	Notes *string     `json:"notes"`
}

func (r CreateWorkoutRequest) Validate() error {
	var v httpx.Validator
	v.NotBlank("name", r.Name, "Name must not be blank")
	v.MaxLen("name", r.Name, 255, "Name must not exceed 255 characters")
	v.Required("date", r.Date != nil, "Date must not be null")
	return v.Err()
}

type AddExerciseRequest struct {
	ExerciseID *int64 `json:"exerciseId"`
}

func (r AddExerciseRequest) Validate() error {
	var v httpx.Validator
	v.Required("exerciseId", r.ExerciseID != nil, "Exercise ID must not be null")
	return v.Err()
}

type AddSetRequest struct {
	WeightKg *decimal.Decimal `json:"weightKg"`
	Reps     int              `json:"reps"`
}

func (r AddSetRequest) Validate() error {
	var v httpx.Validator
	if v.Required("weightKg", r.WeightKg != nil, "Weight must not be null") {
		v.DecimalMin("weightKg", r.WeightKg, decimal.Zero, "Weight must be positive")
		v.DecimalFits("weightKg", r.WeightKg, 6, 2, "Weight must not exceed 9999.99")
	}
	v.Min("reps", r.Reps, 1, "Reps must be at least 1")
	v.Max("reps", r.Reps, 100000, "Reps must not exceed 100000")
	return v.Err()
}

// ---- responses ----

type ExerciseResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	MuscleGroup string `json:"muscleGroup"`
}

type SetResponse struct {
	ID          int64               `json:"id"`
	SetNumber   int                 `json:"setNumber"`
	WeightKg    decimal.Decimal     `json:"weightKg"`
	Reps        int                 `json:"reps"`
	CompletedAt httpx.LocalDateTime `json:"completedAt"`
}

type WorkoutExerciseResponse struct {
	ID         int64            `json:"id"`
	Exercise   ExerciseResponse `json:"exercise"`
	OrderIndex int              `json:"orderIndex"`
	Sets       []SetResponse    `json:"sets"`
}

type WorkoutResponse struct {
	ID        int64                     `json:"id"`
	Name      string                    `json:"name"`
	Date      httpx.Date                `json:"date"`
	Notes     *string                   `json:"notes"`
	Completed bool                      `json:"completed"`
	Exercises []WorkoutExerciseResponse `json:"exercises"`
	CreatedAt httpx.LocalDateTime       `json:"createdAt"`
}

// ---- routes ----

type Handler struct{ *app.Deps }

func Register(r *gin.RouterGroup, d *app.Deps) {
	h := Handler{d}

	ex := r.Group("/exercises")
	ex.GET("", httpx.H(h.listExercises))
	ex.POST("", httpx.H(h.createExercise))

	w := r.Group("/workouts")
	w.GET("", httpx.H(h.listWorkouts))
	w.GET("/:id", httpx.H(h.getWorkout))
	w.POST("", httpx.H(h.createWorkout))
	w.POST("/:id/complete", httpx.H(h.completeWorkout))
	w.DELETE("/:id", httpx.H(h.deleteWorkout))
	w.POST("/:id/exercises", httpx.H(h.addExercise))
	w.DELETE("/:id/exercises/:workoutExerciseId", httpx.H(h.removeExercise))
	w.POST("/:id/exercises/:workoutExerciseId/sets", httpx.H(h.addSet))
	w.DELETE("/:id/exercises/:workoutExerciseId/sets/:setId", httpx.H(h.removeSet))
}

// ---- exercises ----

func (h Handler) listExercises(c *gin.Context) error {
	rows, err := h.Q.ListExercises(c)
	if err != nil {
		return err
	}
	out := make([]ExerciseResponse, 0, len(rows))
	for _, e := range rows {
		out = append(out, ExerciseResponse{ID: e.ID, Name: e.Name, MuscleGroup: e.MuscleGroup})
	}
	return httpx.OK(c, out)
}

func (h Handler) createExercise(c *gin.Context) error {
	var req CreateExerciseRequest
	if err := httpx.Bind(c, &req); err != nil {
		return err
	}
	conflict := httpx.Conflict("Exercise already exists with name: %s", *req.Name)
	exists, err := h.Q.ExerciseExistsByName(c, *req.Name)
	if err != nil {
		return err
	}
	if exists {
		return conflict
	}
	e, err := h.Q.CreateExercise(c, store.CreateExerciseParams{Name: *req.Name, MuscleGroup: *req.MuscleGroup})
	if database.IsUniqueViolation(err) {
		return conflict
	}
	if err != nil {
		return err
	}
	return httpx.Created(c, ExerciseResponse{ID: e.ID, Name: e.Name, MuscleGroup: e.MuscleGroup})
}

// ---- workouts ----

func (h Handler) listWorkouts(c *gin.Context) error {
	workouts, err := h.Q.ListWorkouts(c, auth.UserID(c))
	if err != nil {
		return err
	}
	out, err := assemble(c, h.Q, workouts)
	if err != nil {
		return err
	}
	return httpx.OK(c, out)
}

func (h Handler) getWorkout(c *gin.Context) error {
	id, err := httpx.PathID(c, "id")
	if err != nil {
		return err
	}
	w, err := h.findWorkout(c, h.Q, id, false)
	if err != nil {
		return err
	}
	return h.respondWorkout(c, h.Q, w, 200)
}

func (h Handler) createWorkout(c *gin.Context) error {
	var req CreateWorkoutRequest
	if err := httpx.Bind(c, &req); err != nil {
		return err
	}
	w, err := h.Q.CreateWorkout(c, store.CreateWorkoutParams{
		UserID: auth.UserID(c),
		Name:   *req.Name,
		Date:   req.Date.Time(),
		Notes:  req.Notes,
		Now:    h.Clock.Now(),
	})
	if err != nil {
		return err
	}
	return h.respondWorkout(c, h.Q, w, 201)
}

func (h Handler) completeWorkout(c *gin.Context) error {
	id, err := httpx.PathID(c, "id")
	if err != nil {
		return err
	}
	uid := auth.UserID(c)
	now := h.Clock.Now()
	return h.mutate(c, id, 200, func(q *store.Queries, w *store.Workout) error {
		if err := q.CompleteWorkout(c, store.CompleteWorkoutParams{ID: id, UpdatedAt: now}); err != nil {
			return err
		}
		w.Completed = true
		return activity.Record(c, q, uid, activity.WorkoutCompleted, id, now)
	})
}

func (h Handler) deleteWorkout(c *gin.Context) error {
	id, err := httpx.PathID(c, "id")
	if err != nil {
		return err
	}
	n, err := h.Q.DeleteWorkout(c, store.DeleteWorkoutParams{ID: id, UserID: auth.UserID(c)})
	if err != nil {
		return err
	}
	if n == 0 {
		return httpx.NotFound("Workout", id)
	}
	return httpx.NoContent(c)
}

func (h Handler) addExercise(c *gin.Context) error {
	id, err := httpx.PathID(c, "id")
	if err != nil {
		return err
	}
	var req AddExerciseRequest
	if err := httpx.Bind(c, &req); err != nil {
		return err
	}
	return h.mutate(c, id, 200, func(q *store.Queries, _ *store.Workout) error {
		if _, err := q.GetExercise(c, *req.ExerciseID); database.IsNoRows(err) {
			return httpx.NotFound("Exercise", *req.ExerciseID)
		} else if err != nil {
			return err
		}
		idx, err := q.NextWorkoutExerciseIndex(c, id)
		if err != nil {
			return err
		}
		return q.CreateWorkoutExercise(c, store.CreateWorkoutExerciseParams{
			WorkoutID: id, ExerciseID: *req.ExerciseID, OrderIndex: idx,
		})
	})
}

func (h Handler) removeExercise(c *gin.Context) error {
	id, err := httpx.PathID(c, "id")
	if err != nil {
		return err
	}
	weID, err := httpx.PathID(c, "workoutExerciseId")
	if err != nil {
		return err
	}
	return h.mutate(c, id, 200, func(q *store.Queries, _ *store.Workout) error {
		n, err := q.DeleteWorkoutExercise(c, store.DeleteWorkoutExerciseParams{ID: weID, WorkoutID: id})
		if err != nil {
			return err
		}
		if n == 0 {
			return httpx.NotFound("Workout exercise", weID)
		}
		return q.RenumberWorkoutExercises(c, id)
	})
}

func (h Handler) addSet(c *gin.Context) error {
	id, err := httpx.PathID(c, "id")
	if err != nil {
		return err
	}
	weID, err := httpx.PathID(c, "workoutExerciseId")
	if err != nil {
		return err
	}
	var req AddSetRequest
	if err := httpx.Bind(c, &req); err != nil {
		return err
	}
	now := h.Clock.Now()
	return h.mutate(c, id, 201, func(q *store.Queries, _ *store.Workout) error {
		if err := requireWorkoutExercise(c, q, weID, id); err != nil {
			return err
		}
		n, err := q.NextSetNumber(c, weID)
		if err != nil {
			return err
		}
		return q.CreateWorkoutSet(c, store.CreateWorkoutSetParams{
			WorkoutExerciseID: weID,
			SetNumber:         n,
			WeightKg:          req.WeightKg.Numeric(),
			Reps:              int32(req.Reps),
			CompletedAt:       now,
		})
	})
}

func (h Handler) removeSet(c *gin.Context) error {
	id, err := httpx.PathID(c, "id")
	if err != nil {
		return err
	}
	weID, err := httpx.PathID(c, "workoutExerciseId")
	if err != nil {
		return err
	}
	setID, err := httpx.PathID(c, "setId")
	if err != nil {
		return err
	}
	return h.mutate(c, id, 200, func(q *store.Queries, _ *store.Workout) error {
		if err := requireWorkoutExercise(c, q, weID, id); err != nil {
			return err
		}
		n, err := q.DeleteWorkoutSet(c, store.DeleteWorkoutSetParams{ID: setID, WorkoutExerciseID: weID})
		if err != nil {
			return err
		}
		if n == 0 {
			return httpx.NotFound("Workout set", setID)
		}
		return q.RenumberWorkoutSets(c, weID)
	})
}

// ---- helpers ----

// mutate locks the user's workout, rejects changes to completed workouts,
// applies fn in one transaction and responds with the updated workout.
func (h Handler) mutate(c *gin.Context, id int64, status int, fn func(q *store.Queries, w *store.Workout) error) error {
	var w store.Workout
	var out WorkoutResponse
	err := h.InTx(c, func(q *store.Queries) error {
		var err error
		w, err = h.findWorkout(c, q, id, true)
		if err != nil {
			return err
		}
		if w.Completed {
			return httpx.Conflict("Workout is already completed and cannot be modified")
		}
		if err := fn(q, &w); err != nil {
			return err
		}
		list, err := assemble(c, q, []store.Workout{w})
		if err != nil {
			return err
		}
		out = list[0]
		return nil
	})
	if err != nil {
		return err
	}
	c.JSON(status, out)
	return nil
}

func (h Handler) findWorkout(c *gin.Context, q *store.Queries, id int64, forUpdate bool) (store.Workout, error) {
	var (
		w   store.Workout
		err error
	)
	if forUpdate {
		w, err = q.GetWorkoutForUpdate(c, store.GetWorkoutForUpdateParams{ID: id, UserID: auth.UserID(c)})
	} else {
		w, err = q.GetWorkout(c, store.GetWorkoutParams{ID: id, UserID: auth.UserID(c)})
	}
	if database.IsNoRows(err) {
		return w, httpx.NotFound("Workout", id)
	}
	return w, err
}

func (h Handler) respondWorkout(c *gin.Context, q *store.Queries, w store.Workout, status int) error {
	list, err := assemble(c, q, []store.Workout{w})
	if err != nil {
		return err
	}
	c.JSON(status, list[0])
	return nil
}

func requireWorkoutExercise(c *gin.Context, q *store.Queries, weID, workoutID int64) error {
	ok, err := q.WorkoutExerciseExists(c, store.WorkoutExerciseExistsParams{ID: weID, WorkoutID: workoutID})
	if err != nil {
		return err
	}
	if !ok {
		return httpx.NotFound("Workout exercise", weID)
	}
	return nil
}

// assemble loads exercises and sets for the given workouts with two queries
// and builds the nested responses (each exercise appears exactly once).
func assemble(c *gin.Context, q *store.Queries, workouts []store.Workout) ([]WorkoutResponse, error) {
	out := make([]WorkoutResponse, 0, len(workouts))
	if len(workouts) == 0 {
		return out, nil
	}
	ids := make([]int64, len(workouts))
	for i, w := range workouts {
		ids[i] = w.ID
	}
	wes, err := q.ListWorkoutExercises(c, ids)
	if err != nil {
		return nil, err
	}
	weIDs := make([]int64, len(wes))
	for i, we := range wes {
		weIDs[i] = we.ID
	}
	sets, err := q.ListWorkoutSets(c, weIDs)
	if err != nil {
		return nil, err
	}

	setsByWE := make(map[int64][]SetResponse)
	for _, s := range sets {
		setsByWE[s.WorkoutExerciseID] = append(setsByWE[s.WorkoutExerciseID], SetResponse{
			ID:          s.ID,
			SetNumber:   int(s.SetNumber),
			WeightKg:    decimal.FromNumeric(s.WeightKg),
			Reps:        int(s.Reps),
			CompletedAt: httpx.LocalDateTime(s.CompletedAt),
		})
	}
	exByWorkout := make(map[int64][]WorkoutExerciseResponse)
	for _, we := range wes {
		s := setsByWE[we.ID]
		if s == nil {
			s = []SetResponse{}
		}
		exByWorkout[we.WorkoutID] = append(exByWorkout[we.WorkoutID], WorkoutExerciseResponse{
			ID: we.ID,
			Exercise: ExerciseResponse{
				ID:          we.ExerciseID,
				Name:        we.ExerciseName,
				MuscleGroup: we.ExerciseMuscleGroup,
			},
			OrderIndex: int(we.OrderIndex),
			Sets:       s,
		})
	}
	for _, w := range workouts {
		ex := exByWorkout[w.ID]
		if ex == nil {
			ex = []WorkoutExerciseResponse{}
		}
		out = append(out, WorkoutResponse{
			ID:        w.ID,
			Name:      w.Name,
			Date:      httpx.Date(w.Date),
			Notes:     w.Notes,
			Completed: w.Completed,
			Exercises: ex,
			CreatedAt: httpx.LocalDateTime(w.CreatedAt),
		})
	}
	return out, nil
}
