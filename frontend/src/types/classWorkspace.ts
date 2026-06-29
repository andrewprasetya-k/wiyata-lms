export interface ClassHeader {
  classId: string
  classTitle: string
  classCode: string
}

export interface SubjectClassItem {
  subjectClassId: string
  subjectId: string
  subjectName?: string
  subjectCode?: string
  subjectColor?: string
  teacherId: string
  teacherName?: string
}

export interface SubjectClassesResponse {
  class: ClassHeader
  subjects: SubjectClassItem[]
}

export interface MediaAttachment {
  mediaId: string
  mediaName?: string
  fileSize?: number
  mimeType?: string
  fileUrl?: string
  thumbnailUrl?: string
  ownerType?: string
  createdAt?: string
}

export interface MaterialItem {
  materialId: string
  subjectClassId: string
  subjectName?: string
  subjectColor?: string
  materialTitle: string
  materialDesc: string
  materialType: string
  creatorName?: string
  createdAt: string
  attachments?: MediaAttachment[]
}

export interface PaginatedResponse<T> {
  data: T[]
  totalItems: number
  page: number
  limit: number
  totalPages: number
}

export interface MaterialListWithSubjectResponse {
  subjectClass: {
    subjectClassId: string
    subjectCode: string
    subjectName?: string
    subjectColor?: string
    teacherId: string
    teacherName?: string
  }
  data: PaginatedResponse<MaterialItem>
}
