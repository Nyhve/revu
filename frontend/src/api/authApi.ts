import { api } from "./client";
import type { LoginRequest, LoginResponse, RegisterRequest, UserResponse } from "../types/auth";

export const authApi = {
  register: (data: RegisterRequest) => api.post<UserResponse>("/auth/register", data),
  login: (data: LoginRequest) => api.post<LoginResponse>("/auth/login", data),
};