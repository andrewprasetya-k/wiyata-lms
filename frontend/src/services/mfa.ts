import { api } from "./api";
import type { LoginResponse } from "../types/auth";

export interface MfaVerifyPayload {
  preAuthToken: string;
  code?: string;
  recoveryCode?: string;
}

// POST /login/mfa-verify — public. Completes a login paused for MFA.
export async function verifyMfaLogin(
  payload: MfaVerifyPayload,
): Promise<LoginResponse> {
  const { data } = await api.post<LoginResponse>("/login/mfa-verify", payload);
  return data;
}

export interface MfaEnrollResponse {
  secret: string;
  otpauthUrl: string;
}

// POST /login/mfa-setup/enroll — public, forced-setup flow (grace period
// expired). Driven by the preAuthToken from a mfaSetupRequired challenge.
export async function enrollMfaSetup(
  preAuthToken: string,
): Promise<MfaEnrollResponse> {
  const { data } = await api.post<MfaEnrollResponse>(
    "/login/mfa-setup/enroll",
    { preAuthToken },
  );
  return data;
}

export interface MfaSetupCompleteResponse extends LoginResponse {
  recoveryCodes: string[];
}

// POST /login/mfa-setup/confirm — public. Enables MFA and completes the
// login in one step; recoveryCodes are only ever returned here.
export async function confirmMfaSetup(
  preAuthToken: string,
  code: string,
): Promise<MfaSetupCompleteResponse> {
  const { data } = await api.post<MfaSetupCompleteResponse>(
    "/login/mfa-setup/confirm",
    { preAuthToken, code },
  );
  return data;
}

// GET /me/mfa/status — AuthRequired.
export async function getMfaStatus(): Promise<{ enabled: boolean }> {
  const { data } = await api.get<{ enabled: boolean }>("/me/mfa/status");
  return data;
}

// POST /me/mfa/enroll — AuthRequired. Self-service enrollment, initiated by
// an already logged-in user (not a forced flow) — the JWT is attached
// automatically by api.ts's request interceptor.
export async function enrollMfa(): Promise<MfaEnrollResponse> {
  const { data } = await api.post<MfaEnrollResponse>("/me/mfa/enroll");
  return data;
}

export interface MfaConfirmResponse {
  recoveryCodes: string[];
}

// POST /me/mfa/confirm — AuthRequired.
export async function confirmMfa(code: string): Promise<MfaConfirmResponse> {
  const { data } = await api.post<MfaConfirmResponse>("/me/mfa/confirm", {
    code,
  });
  return data;
}
