export interface LearningProject {
  id: number;
  title: string;
  description: string | null;
  topic: string;
  totalMinutes: number;
  createdAt: string;
}

export interface LearningSession {
  id: number;
  projectId: number;
  durationMinutes: number;
  notes: string | null;
  date: string;
  createdAt: string;
}

export interface CreateLearningProjectRequest {
  title: string;
  description?: string;
  topic: string;
}

export interface CreateLearningSessionRequest {
  durationMinutes: number;
  notes?: string;
  date: string;
}