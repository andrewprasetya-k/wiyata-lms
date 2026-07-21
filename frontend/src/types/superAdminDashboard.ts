export interface SuperAdminSchoolNeedsAttention {
  schoolId: string
  schoolName: string
  schoolCode: string
  createdAt: string
}

export interface SuperAdminTrendPoint {
  period: string
  count: number
}

export interface SuperAdminDashboardSummary {
  schoolsWithoutAdmin: SuperAdminSchoolNeedsAttention[]
  schoolsWithoutAdminTotal: number
  schoolsWithoutSetup: SuperAdminSchoolNeedsAttention[]
  schoolsWithoutSetupTotal: number
  schoolGrowthTrend: SuperAdminTrendPoint[]
  userGrowthTrend: SuperAdminTrendPoint[]
}
