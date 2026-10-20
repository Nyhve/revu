export type GoalStatus = "ACTIVE" | "COMPLETED" | "ARCHIVED";

export interface Goal {
  id: number;
  title: string;
  description: string | null;
  targetDate: string | null;
  status: GoalStatus;
  createdAt: string;
  updatedAt: string;
}

export interface CreateGoalRequest {
  title: string;
  description?: string;
  targetDate?: string;
}

export interface UpdateGoalRequest {
  title: string;
  description?: string;
  targetDate?: string;
  status: GoalStatus;
}