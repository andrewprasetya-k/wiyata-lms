import type { SchoolHeader } from "./adminAcademic";

export interface AdminUserItem {
  userId: string;
  fullName: string;
  email: string;
  isActive: boolean;
  createdAt: string;
}

export interface AdminUserPaginatedResponse {
  data: AdminUserItem[];
  totalItems: number;
  page: number;
  limit: number;
  totalPages: number;
}

export interface SchoolMemberItem {
  schoolUserId: string;
  userId: string;
  fullName?: string;
  email?: string;
  schoolId: string;
  schoolName?: string;
  schoolCode?: string;
  roles?: string[];
  createdAt: string;
}

export interface SchoolMembersResponse {
  school: SchoolHeader;
  members: {
    data: SchoolMemberItem[];
    totalItems: number;
    page: number;
    limit: number;
    totalPages: number;
  };
}

export interface RoleItem {
  roleId: string;
  roleName: string;
  createdAt: string;
}

export interface SyncUserRolesPayload {
  roleIds: string[];
}
