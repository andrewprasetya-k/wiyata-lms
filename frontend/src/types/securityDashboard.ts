export interface BruteForceIncident {
  targetType: "email" | "ip";
  target: string;
  failureCount: number;
  lastAttemptAt: string;
}

export interface SuspiciousActivityItem {
  logId: string;
  action: string;
  severity?: string;
  userId?: string;
  userName?: string;
  userEmail?: string;
  createdAt: string;
}

export interface SecurityDashboardSummary {
  windowHours: number;
  generatedAt: string;
  failedLoginCount: number;
  bruteForceIncidents: BruteForceIncident[];
  passwordResetRequestedCount: number;
  passwordResetCompletedCount: number;
  suspiciousActivities: SuspiciousActivityItem[];
}
