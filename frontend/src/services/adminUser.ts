import { api } from "./api";
import type {
  AdminUserPaginatedResponse,
  RoleItem,
  SchoolMembersResponse,
  SyncUserRolesPayload,
} from "../types/adminUser";

export async function getAdminUsers(params: { page?: number; limit?: number; search?: string }) {
  const { data } = await api.get<AdminUserPaginatedResponse>("/users", {
    params: {
      page: params.page ?? 1,
      limit: params.limit ?? 20,
      search: params.search || undefined,
    },
  });
  return data;
}

export async function getSchoolMembers(
  schoolCode: string,
  params: { page?: number; limit?: number; search?: string },
) {
  const { data } = await api.get<SchoolMembersResponse>(`/school-users/school/${schoolCode}`, {
    params: {
      page: params.page ?? 1,
      limit: params.limit ?? 50,
      search: params.search || undefined,
    },
  });
  return data;
}

export async function getRoles() {
  const { data } = await api.get<RoleItem[]>("/rbac/roles");
  return data;
}

export async function syncUserRoles(schoolUserId: string, payload: SyncUserRolesPayload) {
  const { data } = await api.patch(`/rbac/user-roles/${schoolUserId}`, payload);
  return data;
}
