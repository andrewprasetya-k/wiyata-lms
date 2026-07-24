import { api } from "./api";
import type { SecurityDashboardSummary } from "../types/securityDashboard";

// super_admin only (Phase 11.5.2) — the school-admin variant of this
// endpoint was removed on the backend, so there's no equivalent function
// here anymore.
export async function getSuperAdminSecurityDashboard() {
  const { data } = await api.get<SecurityDashboardSummary>(
    "/dashboard/super-admin/security",
  );
  return data;
}