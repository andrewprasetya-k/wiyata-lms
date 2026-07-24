export type SchoolRole = 'admin' | 'teacher' | 'student'
export type RoleName = SchoolRole | 'super_admin'

export type ActiveContext =
  | {
      type: 'school'
      schoolId: string
      schoolUserId: string
      role: SchoolRole
    }
  | {
      type: 'platform'
      role: 'super_admin'
    }

export interface UserInfo {
  id: string
  fullName: string
  email: string
}

export interface SchoolInfo {
  id: string
  code: string
  name: string
}

export interface MembershipInfo {
  schoolUserId: string
  school: SchoolInfo
  roles: RoleName[]
  isDefault: boolean
}

export interface DefaultContext {
  schoolId: string
  schoolUserId: string
  roles: RoleName[]
}

export interface LoginResponse {
  token: string
  user: UserInfo
  memberships: MembershipInfo[]
  globalRoles: RoleName[]
  defaultContext?: DefaultContext
  // Only present when the user hasn't enrolled in MFA yet but is still
  // within the grace period — a dismissible reminder, never a blocker.
  mfaGraceDaysRemaining?: number
}

// What POST /login and POST /register return instead of LoginResponse when
// the password was correct but a second step still stands between here and
// a completed login. Distinguished from LoginResponse by the absence of a
// `token` field (see isLoginChallenge below).
export interface LoginChallengeResponse {
  mfaRequired?: boolean
  mfaSetupRequired?: boolean
  preAuthToken: string
}

export type LoginOutcome =
  | { kind: 'success'; data: LoginResponse }
  | { kind: 'mfaRequired'; preAuthToken: string }
  | { kind: 'mfaSetupRequired'; preAuthToken: string }

export interface AuthContextResponse {
  memberships: MembershipInfo[]
  globalRoles: RoleName[]
  defaultContext?: DefaultContext
  emailVerified: boolean
  emailVerifiedAt?: string
}

export interface LoginPayload {
  email: string
  password: string
}

export interface RegisterPayload {
  fullName: string
  email: string
  password: string
}
