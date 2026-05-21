import type { DefaultContext, MembershipInfo, RoleName, UserInfo } from '../types/auth'

const TOKEN_KEY = 'edv_token'
const USER_KEY = 'edv_user'
const MEMBERSHIPS_KEY = 'edv_memberships'
const GLOBAL_ROLES_KEY = 'edv_global_roles'
const DEFAULT_CONTEXT_KEY = 'edv_default_context'
const ACTIVE_SCHOOL_KEY = 'edv_active_school_id'
const ACTIVE_ROLES_KEY = 'edv_active_roles'
const ACTIVE_CLASS_KEY = 'edv_active_class_id'

export function getStoredToken() {
  return localStorage.getItem(TOKEN_KEY)
}

export function getActiveSchoolId() {
  return localStorage.getItem(ACTIVE_SCHOOL_KEY)
}

export function persistSession(payload: {
  token: string
  user: UserInfo | null
  memberships: MembershipInfo[]
  globalRoles: RoleName[]
  defaultContext?: DefaultContext
  activeSchoolId: string | null
  activeRoles: RoleName[]
}) {
  localStorage.setItem(TOKEN_KEY, payload.token)
  localStorage.setItem(USER_KEY, JSON.stringify(payload.user))
  localStorage.setItem(MEMBERSHIPS_KEY, JSON.stringify(payload.memberships))
  localStorage.setItem(GLOBAL_ROLES_KEY, JSON.stringify(payload.globalRoles))
  localStorage.setItem(DEFAULT_CONTEXT_KEY, JSON.stringify(payload.defaultContext ?? null))
  if (payload.activeSchoolId) {
    localStorage.setItem(ACTIVE_SCHOOL_KEY, payload.activeSchoolId)
  } else {
    localStorage.removeItem(ACTIVE_SCHOOL_KEY)
  }
  localStorage.setItem(ACTIVE_ROLES_KEY, JSON.stringify(payload.activeRoles))
}

export function readStoredSession() {
  return {
    token: localStorage.getItem(TOKEN_KEY),
    user: parseJSON<UserInfo | null>(localStorage.getItem(USER_KEY), null),
    memberships: parseJSON<MembershipInfo[]>(localStorage.getItem(MEMBERSHIPS_KEY), []),
    globalRoles: parseJSON<RoleName[]>(localStorage.getItem(GLOBAL_ROLES_KEY), []),
    defaultContext: parseJSON<DefaultContext | undefined>(
      localStorage.getItem(DEFAULT_CONTEXT_KEY),
      undefined,
    ),
    activeSchoolId: localStorage.getItem(ACTIVE_SCHOOL_KEY),
    activeRoles: parseJSON<RoleName[]>(localStorage.getItem(ACTIVE_ROLES_KEY), []),
  }
}

export function clearStoredSession() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
  localStorage.removeItem(MEMBERSHIPS_KEY)
  localStorage.removeItem(GLOBAL_ROLES_KEY)
  localStorage.removeItem(DEFAULT_CONTEXT_KEY)
  localStorage.removeItem(ACTIVE_SCHOOL_KEY)
  localStorage.removeItem(ACTIVE_ROLES_KEY)
  localStorage.removeItem(ACTIVE_CLASS_KEY)
}

function parseJSON<T>(raw: string | null, fallback: T): T {
  if (!raw) return fallback
  try {
    return JSON.parse(raw) as T
  } catch {
    return fallback
  }
}
