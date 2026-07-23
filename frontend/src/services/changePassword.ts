import { api } from "./api";

export interface ChangePasswordPayload {
  currentPassword: string;
  newPassword: string;
}

export interface ChangePasswordResponse {
  message: string;
}

export async function changePassword(payload: ChangePasswordPayload) {
  const { data } = await api.patch<ChangePasswordResponse>(
    "/me/change-password",
    payload,
  );
  return data;
}
