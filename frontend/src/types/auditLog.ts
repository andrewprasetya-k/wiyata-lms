export type AuditLogSeverity = "LOW" | "MEDIUM" | "HIGH";
export type AuditLogScope = "platform" | "school";

export interface AuditLogListItem {
  logId: string;
  action: string;
  entityType?: string;
  entityId?: string;
  scope?: AuditLogScope | string;
  severity?: AuditLogSeverity | string;
  schoolId?: string;
  schoolName?: string;
  schoolCode?: string;
  actorUserId: string;
  actorName?: string;
  actorEmail?: string;
  correlationId?: string;
  createdAt: string;
}

export interface AuditLogDetail extends AuditLogListItem {
  metadata: string;
  ipAddress?: string;
  userAgent?: string;
}

export interface AuditLogListResponse {
  data: AuditLogListItem[];
  totalItems: number;
  page: number;
  limit: number;
  totalPages: number;
}

export interface AuditLogFilters {
  schoolId?: string;
  scope?: string;
  action?: string;
  entityType?: string;
  severity?: string;
  actorUserId?: string;
  dateFrom?: string;
  dateTo?: string;
  correlationId?: string;
  search?: string;
  page?: number;
  limit?: number;
}
