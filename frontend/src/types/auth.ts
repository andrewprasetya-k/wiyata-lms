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
}

export interface AuthContextResponse {
  memberships: MembershipInfo[]
  globalRoles: RoleName[]
  defaultContext?: DefaultContext
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
