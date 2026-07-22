export interface AdminEnrollmentTrend {
  className: string
  totalEnrolled: number
  teachers: number
  students: number
}

export interface AdminRecentActivity {
  userName: string
  action: string
  timestamp: string
}

export interface AdminClassWithoutTeacher {
  classId: string
  className: string
}

export interface AdminContentLessSubjectClass {
  subjectClassId: string
  className: string
  subjectName: string
}

export interface AdminSubjectWithoutAssessmentWeight {
  subjectId: string
  subjectName: string
}

export interface AdminGradingBacklogClass {
  classId: string
  className: string
  backlogCount: number
}

export interface AdminDashboardSummary {
  totalStudents: number
  totalTeachers: number
  totalClasses: number
  activeClasses: number
  enrollmentTrends: AdminEnrollmentTrend[]
  recentActivities: AdminRecentActivity[]
  classesWithoutTeacher: AdminClassWithoutTeacher[]
  classesWithoutTeacherTotal: number
  contentLessSubjectClasses: AdminContentLessSubjectClass[]
  contentLessSubjectClassesTotal: number
  subjectsWithoutAssessmentWeight: AdminSubjectWithoutAssessmentWeight[]
  subjectsWithoutAssessmentWeightTotal: number
  backlogTotal: number
  backlogClasses: AdminGradingBacklogClass[]
}
