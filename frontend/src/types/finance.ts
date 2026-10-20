export type TransactionType = "INCOME" | "EXPENSE";

export type TransactionCategory =
  | "SALARY"
  | "FREELANCE"
  | "OTHER_INCOME"
  | "FOOD"
  | "TRANSPORT"
  | "HOUSING"
  | "HEALTH"
  | "ENTERTAINMENT"
  | "SHOPPING"
  | "EDUCATION"
  | "OTHER_EXPENSE";

export interface Transaction {
  id: number;
  amount: number;
  description: string;
  category: TransactionCategory;
  type: TransactionType;
  date: string;
  createdAt: string;
}

export interface MonthlyStats {
  year: number;
  month: number;
  totalIncome: number;
  totalExpenses: number;
  balance: number;
}

export interface CreateTransactionRequest {
  amount: number;
  description: string;
  category: TransactionCategory;
  type: TransactionType;
  date: string;
}

export interface MonthlyFinanceSummary {
  year: number;
  month: number;
  monthLabel: string;
  income: number;
  expenses: number;
}