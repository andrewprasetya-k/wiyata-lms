import { api } from './api'

export interface VerifyEmailResponse {
  message: string
  emailVerifiedAt: string
}

export interface ResendVerificationResponse {
  message: string
}

export async function verifyEmail(token: string) {
  const { data } = await api.post<VerifyEmailResponse>('/verify-email', { token })
  return data
}

export async function resendVerificationEmail() {
  const { data } = await api.post<ResendVerificationResponse>('/me/resend-verification')
  return data
}
