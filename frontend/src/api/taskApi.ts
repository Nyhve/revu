import { api } from "./client";
import type { CreateTaskRequest, Task, UpdateTaskRequest } from "../types/task";

export const taskApi = {
  getAll: () => api.get<Task[]>("/tasks"),
  getById: (id: number) => api.get<Task>(`/tasks/${id}`),
  create: (data: CreateTaskRequest) => api.post<Task>("/tasks", data),
  update: (id: number, data: UpdateTaskRequest) => api.put<Task>(`/tasks/${id}`, data),
  delete: (id: number) => api.delete<void>(`/tasks/${id}`),
};