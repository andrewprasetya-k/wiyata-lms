import { api } from "./api";
import type {
  AssignSubjectClassPayload,
  SubjectClassesByClassResponse,
} from "../types/adminSubjectClass";

export async function getSubjectClassesByClass(classId: string) {
  const { data } = await api.get<SubjectClassesByClassResponse>(
    `/subject-classes/class/${classId}`,
  );
  return data;
}

export async function assignSubjectClass(payload: AssignSubjectClassPayload) {
  const { data } = await api.post("/subject-classes/assign", payload);
  return data;
}
