import { api } from "./api";
import { getSubjectClassesByClass as getSubjectClassesByClassWorkspace } from "./classWorkspace";
import type {
  AssignSubjectClassPayload,
  SubjectClassesByClassResponse,
} from "../types/adminSubjectClass";


export async function getSubjectClassesByClass(
  classId: string,
): Promise<SubjectClassesByClassResponse> {
  return getSubjectClassesByClassWorkspace(classId);
}

export async function assignSubjectClass(payload: AssignSubjectClassPayload) {
  const { data } = await api.post("/subject-classes/assign", payload);
  return data;
}

export async function deleteSubjectClass(subjectClassId: string) {
  const { data } = await api.delete(`/subject-classes/${subjectClassId}`);
  return data;
}
