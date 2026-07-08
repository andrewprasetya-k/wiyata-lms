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

export interface AdminDashboardSummary {
  totalStudents: number
  totalTeachers: number
  totalClasses: number
  activeClasses: number
  enrollmentTrends: AdminEnrollmentTrend[]
  recentActivities: AdminRecentActivity[]
}
