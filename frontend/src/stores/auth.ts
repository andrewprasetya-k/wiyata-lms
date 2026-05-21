import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { api } from '../services/api'
import { clearStoredSession, persistSession, readStoredSession } from '../services/session'
import { useActiveClassStore } from './activeClass'
import type {
  DefaultContext,
  LoginPayload,
  LoginResponse,
  MembershipInfo,
  RoleName,
  UserInfo,
} from '../types/auth'

const rolePriority: RoleName[] = ['super_admin', 'admin', 'teacher', 'student']

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(null)
  const user = ref<UserInfo | null>(null)
  const memberships = ref<MembershipInfo[]>([])
  const globalRoles = ref<RoleName[]>([])
  const defaultContext = ref<DefaultContext | undefined>()
  const activeSchoolId = ref<string | null>(null)
  const activeRoles = ref<RoleName[]>([])
  const isRestored = ref(false)

  const isAuthenticated = computed(() => Boolean(token.value))
  const allRoles = computed<RoleName[]>(() => {
    const roles = new Set<RoleName>([...globalRoles.value, ...activeRoles.value])
    return [...roles]
  })

  function applySession(response: LoginResponse) {
    token.value = response.token
    user.value = response.user
    memberships.value = response.memberships ?? []
    globalRoles.value = response.globalRoles ?? []
    defaultContext.value = response.defaultContext
    activeSchoolId.value = response.defaultContext?.schoolId ?? null
    activeRoles.value = response.defaultContext?.roles ?? response.memberships?.[0]?.roles ?? []

    persistSession({
      token: token.value,
      user: user.value,
      memberships: memberships.value,
      globalRoles: globalRoles.value,
      defaultContext: defaultContext.value,
      activeSchoolId: activeSchoolId.value,
      activeRoles: activeRoles.value,
    })
  }

  async function login(payload: LoginPayload) {
    const { data } = await api.post<LoginResponse>('/login', payload)
    applySession(data)
    return data
  }

  function logout() {
    const activeClass = useActiveClassStore()
    token.value = null
    user.value = null
    memberships.value = []
    globalRoles.value = []
    defaultContext.value = undefined
    activeSchoolId.value = null
    activeRoles.value = []
    activeClass.reset()
    clearStoredSession()
  }

  function restoreSession() {
    if (isRestored.value) return
    const stored = readStoredSession()
    token.value = stored.token
    user.value = stored.user
    memberships.value = stored.memberships
    globalRoles.value = stored.globalRoles
    defaultContext.value = stored.defaultContext
    activeSchoolId.value = stored.activeSchoolId
    activeRoles.value = stored.activeRoles
    isRestored.value = true
  }

  function hasAnyRole(roles: RoleName[]) {
    if (roles.length === 0) return true
    return roles.some((role) => allRoles.value.includes(role))
  }

  function primaryRole() {
    return rolePriority.find((role) => allRoles.value.includes(role)) ?? null
  }

  return {
    token,
    user,
    memberships,
    globalRoles,
    defaultContext,
    activeSchoolId,
    activeRoles,
    isAuthenticated,
    allRoles,
    login,
    logout,
    restoreSession,
    hasAnyRole,
    primaryRole,
  }
})
