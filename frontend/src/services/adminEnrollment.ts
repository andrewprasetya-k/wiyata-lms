import { api } from "./api";
import type {
  ClassEnrollmentsResponse,
  CreateEnrollmentPayload,
} from "../types/adminEnrollment";

export async function getClassEnrollments(
  classId: string,
  params: { page?: number; limit?: number; search?: string } = {},
) {
  const { data } = await api.get<ClassEnrollmentsResponse>(`/enrollments/class/${classId}`, {
    params: {
      page: params.page ?? 1,
      limit: params.limit ?? 50,
      search: params.search || undefined,
    },
  });
  return data;
}

export async function createClassEnrollments(payload: CreateEnrollmentPayload) {
  const { data } = await api.post("/enrollments", payload);
  return data;
}

export async function deleteEnrollment(enrollmentId: string) {
  const { data } = await api.delete(`/enrollments/${enrollmentId}`);
  return data;
}
