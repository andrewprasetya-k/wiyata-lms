import type { SchoolHeader } from "./adminAcademic";

export interface AdminClassItem {
  classId: string;
  schoolId: string;
  schoolName?: string;
  termId: string;
  termName?: string;
  academicYearName?: string;
  classCode: string;
  classTitle: string;
  classDesc: string;
  createdBy: string;
  creatorName?: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface AdminClassPaginatedData {
  data: AdminClassItem[];
  totalItems: number;
  page: number;
  limit: number;
  totalPages: number;
}

export interface AdminClassListResponse {
  school?: SchoolHeader;
  data: AdminClassPaginatedData;
}

export interface GetAdminClassesParams {
  schoolCode: string;
  termId: string;
  page?: number;
  limit?: number;
  search?: string;
}

export interface CreateAdminClassPayload {
  schoolId: string;
  termId: string;
  classCode: string;
  classTitle: string;
  classDesc: string;
}

export interface UpdateAdminClassPayload {
  classTitle?: string;
  classDesc?: string;
  isActive?: boolean;
}
