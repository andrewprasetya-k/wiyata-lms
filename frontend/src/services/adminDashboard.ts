import { api } from './api'
import type { AdminDashboardSummary } from '../types/adminDashboard'

export async function getAdminDashboard(schoolId: string) {
  const { data } = await api.get<AdminDashboardSummary>(`/dashboard/admin/${schoolId}`)
  return data
}
