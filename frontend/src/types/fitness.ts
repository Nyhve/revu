export type MuscleGroup =
  | "CHEST"
  | "BACK"
  | "LEGS"
  | "SHOULDERS"
  | "ARMS"
  | "CORE"
  | "FULL_BODY";

export interface Exercise {
  id: number;
  name: string;
  muscleGroup: MuscleGroup;
}

export interface WorkoutSet {
  id: number;
  setNumber: number;
  weightKg: number;
  reps: number;
  completedAt: string;
}

export interface WorkoutExercise {
  id: number;
  exercise: Exercise;
  orderIndex: number;
  sets: WorkoutSet[];
}

export interface Workout {
  id: number;
  name: string;
  date: string;
  notes: string | null;
  completed: boolean;
  exercises: WorkoutExercise[];
  createdAt: string;
}

export interface CreateWorkoutRequest {
  name: string;
  date: string;
  notes?: string;
}

export interface AddSetRequest {
  weightKg: number;
  reps: number;
}