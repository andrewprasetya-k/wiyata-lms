import { api } from './api'

export interface SchoolRegistrationRequestPayload {
  schoolName: string
  npsn?: string
  picName: string
  picEmail: string
  picPhone?: string
  picRole?: string
  message?: string
}

export interface SchoolRegistrationRequestSummary {
  requestId: string
  schoolName: string
  picName: string
  picEmail: string
  status: 'pending' | 'approved' | 'rejected'
  createdAt: string
}

export interface SchoolRegistrationRequestResponse {
  message: string
  request: SchoolRegistrationRequestSummary
}

export interface InvitationSchool {
  schoolId: string
  schoolCode: string
  schoolName: string
}

export interface InvitationMetadata {
  invitationId: string
  email: string
  role: string
  school: InvitationSchool
  expiresAt: string
  status: 'valid'
}

export interface AcceptInvitationPayload {
  name: string
  password: string
  confirmPassword: string
}

export interface AcceptedInvitationUser {
  userId: string
  fullName: string
  email: string
}

export interface AcceptInvitationResponse {
  message: string
  user: AcceptedInvitationUser
  school: InvitationSchool
  role: string
}

export type SchoolRegistrationStatus = 'pending' | 'approved' | 'rejected'

export interface SchoolRegistrationRequestDetail {
  requestId: string
  requesterUserId?: string
  schoolName: string
  npsn?: string
  picName: string
  picEmail: string
  picPhone?: string
  picRole?: string
  message?: string
  status: SchoolRegistrationStatus
  reviewedBy?: string
  reviewedAt?: string
  reviewNote?: string
  createdAt: string
  updatedAt: string
}

export interface SchoolRegistrationRequestListResponse {
  data: SchoolRegistrationRequestDetail[]
  totalItems: number
  page: number
  limit: number
  totalPages: number
}

export interface ApproveSchoolRegistrationPayload {
  schoolCode: string
  schoolName?: string
  note?: string
}

export interface ApproveSchoolRegistrationResponse {
  message: string
  request: SchoolRegistrationRequestDetail
  school: {
    schoolId: string
    schoolCode: string
    schoolName: string
  }
  admin: {
    userId: string
    fullName: string
    email: string
    schoolUserId: string
    role: string
  }
}

export interface RejectSchoolRegistrationPayload {
  reason?: string
}

export interface RejectSchoolRegistrationResponse {
  message: string
  request: SchoolRegistrationRequestDetail
}

export async function submitSchoolRegistrationRequest(
  payload: SchoolRegistrationRequestPayload,
) {
  const { data } = await api.post<SchoolRegistrationRequestResponse>(
    '/school-registration-requests',
    payload,
  )
  return data
}

export async function getInvitation(token: string) {
  const { data } = await api.get<InvitationMetadata>(
    `/invitations/${encodeURIComponent(token)}`,
  )
  return data
}

export async function acceptInvitation(
  token: string,
  payload: AcceptInvitationPayload,
) {
  const { data } = await api.post<AcceptInvitationResponse>(
    `/invitations/${encodeURIComponent(token)}/accept`,
    payload,
  )
  return data
}

export async function getSchoolRegistrationRequests(params: {
  status?: SchoolRegistrationStatus
  page?: number
  limit?: number
}) {
  const { data } = await api.get<SchoolRegistrationRequestListResponse>(
    '/super-admin/school-registration-requests',
    {
      params: {
        status: params.status,
        page: params.page ?? 1,
        limit: params.limit ?? 10,
      },
    },
  )
  return data
}

export async function getSchoolRegistrationRequestDetail(id: string) {
  const { data } = await api.get<SchoolRegistrationRequestDetail>(
    `/super-admin/school-registration-requests/${id}`,
  )
  return data
}

export async function approveSchoolRegistrationRequest(
  id: string,
  payload: ApproveSchoolRegistrationPayload,
) {
  const { data } = await api.patch<ApproveSchoolRegistrationResponse>(
    `/super-admin/school-registration-requests/${id}/approve`,
    payload,
  )
  return data
}

export async function rejectSchoolRegistrationRequest(
  id: string,
  payload: RejectSchoolRegistrationPayload,
) {
  const { data } = await api.patch<RejectSchoolRegistrationResponse>(
    `/super-admin/school-registration-requests/${id}/reject`,
    payload,
  )
  return data
}
