import { api } from "./api";
import type {
  AuditLogDetail,
  AuditLogFilters,
  AuditLogListResponse,
} from "../types/auditLog";

function buildParams(filters: AuditLogFilters) {
  return {
    schoolId: filters.schoolId || undefined,
    scope: filters.scope || undefined,
    action: filters.action || undefined,
    entityType: filters.entityType || undefined,
    severity: filters.severity || undefined,
    actorUserId: filters.actorUserId || undefined,
    dateFrom: filters.dateFrom || undefined,
    dateTo: filters.dateTo || undefined,
    correlationId: filters.correlationId || undefined,
    search: filters.search || undefined,
    page: filters.page ?? 1,
    limit: filters.limit ?? 20,
  };
}

// Platform-wide search — super admin only (GET /api/logs), optional schoolId
// filter narrows without needing the school-scoped route below.
export async function getPlatformAuditLogs(filters: AuditLogFilters) {
  const { data } = await api.get<AuditLogListResponse>("/logs", {
    params: buildParams(filters),
  });
  return data;
}

// School-pinned search — school admin (and super admin acting on one
// school). schoolId is enforced server-side regardless of filters.schoolId.
export async function getSchoolAuditLogs(
  schoolId: string,
  filters: AuditLogFilters,
) {
  const { data } = await api.get<AuditLogListResponse>(
    `/logs/school/${schoolId}/search`,
    { params: buildParams(filters) },
  );
  return data;
}

export async function getPlatformAuditLogDetail(id: string) {
  const { data } = await api.get<AuditLogDetail>(`/logs/${id}`);
  return data;
}

export async function getSchoolAuditLogDetail(schoolId: string, id: string) {
  const { data } = await api.get<AuditLogDetail>(
    `/logs/school/${schoolId}/entries/${id}`,
  );
  return data;
}
