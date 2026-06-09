export interface SubjectClassItem {
  subjectClassId: string;
  subjectId: string;
  subjectName?: string;
  subjectCode?: string;
  teacherId: string;
  teacherName?: string;
}

export interface SubjectClassHeader {
  classId: string;
  classTitle: string;
  classCode: string;
}

export interface SubjectClassesByClassResponse {
  class: SubjectClassHeader;
  subjects: SubjectClassItem[];
}

export interface AssignSubjectClassPayload {
  classId: string;
  subjectId: string;
  teacherId: string;
}
