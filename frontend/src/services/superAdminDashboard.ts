import { api } from './api'
import type { SuperAdminDashboardSummary } from '../types/superAdminDashboard'

export async function getSuperAdminDashboard() {
  const { data } = await api.get<SuperAdminDashboardSummary>('/dashboard/super-admin')
  return data
}
