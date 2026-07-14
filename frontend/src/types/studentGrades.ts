export type GradebookAssignmentStatus =
  | "not_submitted"
  | "submitted"
  | "graded";

export interface GradebookClass {
  classId: string;
  className: string;
  classCode: string;
}

export interface GradebookAssignment {
  assignmentId: string;
  assignmentTitle: string;
  categoryName: string;
  deadline?: string | null;
  status: GradebookAssignmentStatus;
  submittedAt?: string | null;
  score?: number | null;
  feedback?: string | null;
  assessedAt?: string | null;
  assessorName?: string | null;
}

export interface GradebookSubject {
  subjectClassId: string;
  subjectId: string;
  subjectName: string;
  subjectCode: string;
  finalGrade?: number | null;
  gradedCount: number;
  submittedCount: number;
  pendingCount: number;
  assignments: GradebookAssignment[];
}

export interface GradebookSummary {
  subjectCount: number;
  gradedAssignmentCount: number;
  submittedAssignmentCount: number;
  pendingAssessmentCount: number;
}

export interface MyGradebookResponse {
  class: GradebookClass;
  subjects: GradebookSubject[];
  summary: GradebookSummary;
}
