export interface ClassGradeReportClass {
  classId: string
  classTitle: string
  classCode: string
}

export interface ClassGradeReportSubject {
  subjectId: string
  subjectName: string
  subjectCode: string
}

export interface ClassGradeReportStudent {
  studentId: string
  studentName: string
  studentEmail: string
  finalGrade: number
  letterGrade: string
}

export interface ClassGradeReportResponse {
  class: ClassGradeReportClass
  subject: ClassGradeReportSubject
  students: ClassGradeReportStudent[]
}

export type StudentGradeAssignmentStatus = 'not_submitted' | 'submitted' | 'graded'

export interface StudentGradeCategoryBreakdown {
  categoryId: string
  categoryName: string
  averageScore: number
  weightedScore: number
  weight: number
  assignmentCount: number
}

export interface StudentGradeAssignment {
  assignmentId: string
  assignmentTitle: string
  categoryName: string
  deadline?: string | null
  status: StudentGradeAssignmentStatus
  submittedAt?: string | null
  score?: number | null
  feedback?: string | null
  assessedAt?: string | null
  assessorName?: string | null
}

export interface StudentGradeDetailResponse {
  studentId: string
  studentName: string
  studentEmail: string
  class: ClassGradeReportClass
  subject: ClassGradeReportSubject
  finalGrade: number
  letterGrade: string
  breakdown: StudentGradeCategoryBreakdown[]
  assignments: StudentGradeAssignment[]
}
