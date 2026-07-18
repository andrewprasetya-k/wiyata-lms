import { api } from "./api";
import { getAssignmentCategoriesBySchool } from "./adminAcademic";
import type {
  AssignmentWithSubmissionsResponse,
  CreateAssignmentPayload,
  SchoolCategoriesResponse,
  TeacherAssignmentInboxResponse,
  TeacherSubmissionInboxResponse,
  TeacherSubjectClassSubmissionsResponse,
} from "../types/teacherAssignment";

export async function createAssignment(payload: CreateAssignmentPayload) {
  const { data } = await api.post("/assignments", payload);
  return data;
}

export async function updateAssignment(
  id: string,
  payload: Partial<CreateAssignmentPayload>,
) {
  const { data } = await api.patch(`/assignments/${id}`, payload);
  return data;
}

export async function deleteAssignment(id: string) {
  const { data } = await api.delete(`/assignments/${id}`);
  return data;
}

export async function getAssignmentCategories(
  schoolCode: string,
): Promise<SchoolCategoriesResponse> {
  return getAssignmentCategoriesBySchool(schoolCode);
}

export async function getAssignmentDetailWithSubmissions(assignmentId: string) {
  const { data } = await api.get<AssignmentWithSubmissionsResponse>(
    `/assignments/${assignmentId}`,
  );
  return data;
}

export async function getSubjectClassSubmissions(subjectClassId: string) {
  const { data } = await api.get<TeacherSubjectClassSubmissionsResponse>(
    `/assignments/subject-class/submissions/${subjectClassId}`,
  );
  return data;
}

export async function getTeacherSubmissionInbox() {
  const { data } = await api.get<TeacherSubmissionInboxResponse>(
    "/assignments/teacher-submissions",
  );
  return data;
}

export async function getTeacherAssignmentInbox() {
  const { data } = await api.get<TeacherAssignmentInboxResponse>(
    "/assignments/teacher-assignments",
  );
  return data;
}

export async function assessSubmission(
  submissionId: string,
  payload: { score: number; feedback: string },
) {
  const { data } = await api.post(
    `/assignments/assess/${submissionId}`,
    payload,
  );
  return data;
}
