import { api } from "./api";

export interface ForgotPasswordResponse {
  message: string;
}

export async function requestPasswordReset(email: string) {
  const { data } = await api.post<ForgotPasswordResponse>(
    "/forgot-password",
    { email },
  );
  return data;
}

export interface PasswordResetMetadata {
  status: string;
  expiresAt: string;
}

export async function getPasswordResetMetadata(token: string) {
  const { data } = await api.get<PasswordResetMetadata>(
    `/reset-password/${encodeURIComponent(token)}`,
  );
  return data;
}

export interface ResetPasswordResponse {
  message: string;
}

export async function resetPassword(token: string, newPassword: string) {
  const { data } = await api.post<ResetPasswordResponse>(
    `/reset-password/${encodeURIComponent(token)}`,
    { newPassword },
  );
  return data;
}
