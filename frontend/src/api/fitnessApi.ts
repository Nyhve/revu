import { api } from "./client";
import type {
  Exercise,
  Workout,
  CreateWorkoutRequest,
  AddSetRequest,
} from "../types/fitness";

export const fitnessApi = {
  getExercises: () => api.get<Exercise[]>("/exercises"),
  createExercise: (name: string, muscleGroup: string) =>
    api.post<Exercise>("/exercises", { name, muscleGroup }),

  getWorkouts: () => api.get<Workout[]>("/workouts"),
  getWorkout: (id: number) => api.get<Workout>(`/workouts/${id}`),
  createWorkout: (data: CreateWorkoutRequest) =>
    api.post<Workout>("/workouts", data),
  completeWorkout: (id: number) =>
    api.post<Workout>(`/workouts/${id}/complete`),
  deleteWorkout: (id: number) => api.delete<void>(`/workouts/${id}`),

  addExercise: (workoutId: number, exerciseId: number) =>
    api.post<Workout>(`/workouts/${workoutId}/exercises`, { exerciseId }),
  removeExercise: (workoutId: number, workoutExerciseId: number) =>
    api.delete<Workout>(`/workouts/${workoutId}/exercises/${workoutExerciseId}`),

  addSet: (workoutId: number, workoutExerciseId: number, data: AddSetRequest) =>
    api.post<Workout>(
      `/workouts/${workoutId}/exercises/${workoutExerciseId}/sets`,
      data,
    ),
  removeSet: (workoutId: number, workoutExerciseId: number, setId: number) =>
    api.delete<Workout>(
      `/workouts/${workoutId}/exercises/${workoutExerciseId}/sets/${setId}`,
    ),
};