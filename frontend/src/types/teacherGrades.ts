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
