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
