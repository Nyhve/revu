import { api } from "./client";
import type {
  Transaction,
  MonthlyStats,
  CreateTransactionRequest,
} from "../types/finance";
import type { MonthlyFinanceSummary } from "../types/finance";


export const financeApi = {
  getAll: () => api.get<Transaction[]>("/finance"),
  create: (data: CreateTransactionRequest) =>
    api.post<Transaction>("/finance", data),
  delete: (id: number) => api.delete<void>(`/finance/${id}`),
  getMonthlyStats: (year: number, month: number) =>
    api.get<MonthlyStats>(`/finance/stats/${year}/${month}`),
  getLastMonthsSummary: (months: number) =>
    api.get<MonthlyFinanceSummary[]>(`/finance/summary?months=${months}`),
};