export interface SubjectClassItem {
  subjectClassId: string;
  subjectId: string;
  subjectName?: string;
  subjectCode?: string;
  subjectColor?: string;
  teacherId: string;
  teacherName?: string;
}

export interface ClassHeader {
  classId: string;
  classTitle: string;
  classCode: string;
}

export interface SubjectClassesByClassResponse {
  class: ClassHeader;
  subjects: SubjectClassItem[];
}

export interface AssignSubjectClassPayload {
  classId: string;
  subjectId: string;
  teacherId: string;
}
