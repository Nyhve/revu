import { api } from "./client";
import type {
  LearningProject,
  LearningSession,
  CreateLearningProjectRequest,
  CreateLearningSessionRequest,
} from "../types/learning";

export const learningApi = {
  getAll: () => api.get<LearningProject[]>("/learning"),
  create: (data: CreateLearningProjectRequest) =>
    api.post<LearningProject>("/learning", data),
  delete: (id: number) => api.delete<void>(`/learning/${id}`),

  getSessions: (projectId: number) =>
    api.get<LearningSession[]>(`/learning/${projectId}/sessions`),
  addSession: (projectId: number, data: CreateLearningSessionRequest) =>
    api.post<LearningSession>(`/learning/${projectId}/sessions`, data),
  deleteSession: (projectId: number, sessionId: number) =>
    api.delete<void>(`/learning/${projectId}/sessions/${sessionId}`),
};