export interface SuperAdminSchoolNeedsAttention {
  schoolId: string
  schoolName: string
  schoolCode: string
  createdAt: string
}

export interface SuperAdminDashboardSummary {
  schoolsWithoutAdmin: SuperAdminSchoolNeedsAttention[]
  schoolsWithoutAdminTotal: number
  schoolsWithoutSetup: SuperAdminSchoolNeedsAttention[]
  schoolsWithoutSetupTotal: number
}
