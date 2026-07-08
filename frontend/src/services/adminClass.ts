import { api } from "./api";
import type {
  AdminClassListResponse,
  CreateAdminClassPayload,
  GetAdminClassesParams,
  UpdateAdminClassPayload,
} from "../types/adminClass";

export async function getAdminClasses(params: GetAdminClassesParams) {
  const { data } = await api.get<AdminClassListResponse>("/classes", {
    params: {
      schoolCode: params.schoolCode,
      termId: params.termId,
      page: params.page ?? 1,
      limit: params.limit ?? 50,
      search: params.search || undefined,
    },
  });
  return data;
}

export async function createAdminClass(payload: CreateAdminClassPayload) {
  const { data } = await api.post("/classes", payload);
  return data;
}

export async function updateAdminClass(classId: string, payload: UpdateAdminClassPayload) {
  const { data } = await api.patch(`/classes/${classId}`, payload);
  return data;
}
