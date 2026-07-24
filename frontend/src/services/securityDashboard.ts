import { api } from "./api";
import type { SecurityDashboardSummary } from "../types/securityDashboard";

export async function getAdminSecurityDashboard(schoolId: string) {
  const { data } = await api.get<SecurityDashboardSummary>(
    `/dashboard/admin/${schoolId}/security`,
  );
  return data;
}

export async function getSuperAdminSecurityDashboard() {
  const { data } = await api.get<SecurityDashboardSummary>(
    "/dashboard/super-admin/security",
  );
  return data;
}