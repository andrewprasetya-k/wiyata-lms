export interface StudentClassEnrollment {
  enrollmentId: string
  schoolId: string
  schoolUserId: string
  classId: string
  classTitle?: string
  role: 'student' | 'teacher' | string
  joinedAt: string
}
