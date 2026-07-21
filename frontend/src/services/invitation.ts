import { api } from './api'

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
  existingUser: boolean
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

// For an already-authenticated existing user: identity comes from the
// Authorization header the `api` client already attaches, not from a
// submitted name/password. The backend verifies the logged-in account's
// email matches the invitation before accepting.
export async function acceptInvitationAuthenticated(token: string) {
  const { data } = await api.post<AcceptInvitationResponse>(
    `/invitations/${encodeURIComponent(token)}/accept-authenticated`,
  )
  return data
}
