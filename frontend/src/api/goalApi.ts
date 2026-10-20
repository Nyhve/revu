import { api } from "./client";
import type { CreateGoalRequest, Goal, UpdateGoalRequest } from "../types/goal";

export const goalApi = {
  getAll: () => api.get<Goal[]>("/goals"),
  getById: (id: number) => api.get<Goal>(`/goals/${id}`),
  create: (data: CreateGoalRequest) => api.post<Goal>("/goals", data),
  update: (id: number, data: UpdateGoalRequest) => api.put<Goal>(`/goals/${id}`, data),
  delete: (id: number) => api.delete<void>(`/goals/${id}`),
};