import { defineStore } from "pinia";
import { computed, ref } from "vue";
import { api } from "../services/api";
import {
  clearStoredSession,
  persistSession,
  readStoredSession,
} from "../services/session";
import { useActiveClassStore } from "./activeClass";
import type {
  ActiveContext,
  AuthContextResponse,
  DefaultContext,
  LoginPayload,
  LoginResponse,
  MembershipInfo,
  RoleName,
  SchoolRole,
  UserInfo,
} from "../types/auth";

const schoolRolePriority: SchoolRole[] = ["admin", "teacher", "student"];

const landingRouteByRole: Record<RoleName, string> = {
  super_admin: "/superadmin/dashboard",
  admin: "/admin/dashboard",
  teacher: "/teacher/dashboard",
  student: "/student/dashboard",
};

export const useAuthStore = defineStore("auth", () => {
  const token = ref<string | null>(null);
  const user = ref<UserInfo | null>(null);
  const memberships = ref<MembershipInfo[]>([]);
  const globalRoles = ref<RoleName[]>([]);
  const defaultContext = ref<DefaultContext | undefined>();
  const activeContext = ref<ActiveContext | null>(null);
  const isRestored = ref(false);
  const isContextReady = ref(false);
  const isContextInitialized = ref(false);
  const contextVersion = ref(0);
  let contextRefreshPromise: Promise<boolean> | null = null;
  let contextRefreshGeneration = 0;

  const isAuthenticated = computed(() => Boolean(token.value));

  const availableContexts = computed<ActiveContext[]>(() => {
    const seen = new Set<string>();
    const contexts: ActiveContext[] = [];

    for (const membership of memberships.value) {
      for (const role of schoolRolePriority) {
        if (!membership.roles.includes(role)) continue;
        const key = `school:${membership.school.id}:${membership.schoolUserId}:${role}`;
        if (seen.has(key)) continue;
        seen.add(key);
        contexts.push({
          type: "school",
          schoolId: membership.school.id,
          schoolUserId: membership.schoolUserId,
          role,
        });
      }
    }

    if (globalRoles.value.includes("super_admin")) {
      contexts.push({ type: "platform", role: "super_admin" });
    }

    return contexts;
  });

  const activeRole = computed<RoleName | null>(
    () => activeContext.value?.role ?? null,
  );
  const activeSchoolId = computed(() =>
    activeContext.value?.type === "school"
      ? activeContext.value.schoolId
      : null,
  );
  const activeMembership = computed(() => {
    const context = activeContext.value;
    if (context?.type !== "school") return null;
    return (
      memberships.value.find(
        (membership) =>
          membership.school.id === context.schoolId &&
          membership.schoolUserId === context.schoolUserId,
      ) ?? null
    );
  });
  const activeSchoolUserId = computed(
    () => activeMembership.value?.schoolUserId ?? "",
  );
  // Compatibility surface: these intentionally expose only the selected active role.
  const activeRoles = computed<RoleName[]>(() =>
    activeRole.value ? [activeRole.value] : [],
  );
  const allRoles = computed<RoleName[]>(() => activeRoles.value);

  function applySession(response: LoginResponse) {
    token.value = response.token;
    user.value = response.user;
    memberships.value = response.memberships ?? [];
    globalRoles.value = response.globalRoles ?? [];
    defaultContext.value = response.defaultContext;

    const stored = readStoredSession();
    activeContext.value = chooseActiveContext({
      preferredContext: stored.activeContext,
      legacySchoolId: stored.activeSchoolId,
      defaultContext: defaultContext.value,
    });

    isContextReady.value = true;
    isContextInitialized.value = true;
    contextRefreshGeneration += 1;
    contextRefreshPromise = null;
    persistCurrentSession();
  }

  async function login(payload: LoginPayload) {
    const { data } = await api.post<LoginResponse>("/login", payload);
    applySession(data);
    return data;
  }

  function logout() {
    const activeClass = useActiveClassStore();
    token.value = null;
    user.value = null;
    memberships.value = [];
    globalRoles.value = [];
    defaultContext.value = undefined;
    activeContext.value = null;
    isContextReady.value = false;
    isContextInitialized.value = false;
    contextRefreshGeneration += 1;
    contextRefreshPromise = null;
    contextVersion.value += 1;
    activeClass.reset();
    clearStoredSession();
  }

  function restoreSession() {
    if (isRestored.value) return;
    const stored = readStoredSession();
    token.value = stored.token;
    user.value = stored.user;
    memberships.value = stored.memberships;
    globalRoles.value = stored.globalRoles;
    defaultContext.value = stored.defaultContext;
    activeContext.value = chooseActiveContext({
      preferredContext: stored.activeContext,
      legacySchoolId: stored.activeSchoolId,
      defaultContext: stored.defaultContext,
    });
    isContextReady.value = !token.value || Boolean(activeContext.value);
    isRestored.value = true;
    persistCurrentSession();
  }

  async function ensureUserContext() {
    if (!token.value) {
      isContextReady.value = true;
      isContextInitialized.value = true;
      return false;
    }
    if (isContextInitialized.value) {
      isContextReady.value = true;
      return Boolean(activeContext.value);
    }
    return requestUserContext();
  }

  async function refreshUserContext() {
    return requestUserContext();
  }

  async function requestUserContext() {
    if (!token.value) {
      isContextReady.value = true;
      isContextInitialized.value = true;
      return false;
    }
    if (contextRefreshPromise) {
      return contextRefreshPromise;
    }

    const requestToken = token.value;
    const requestGeneration = contextRefreshGeneration;

    const request = api
      .get<AuthContextResponse>("/me/context")
      .then(({ data }) => {
        if (
          token.value !== requestToken ||
          requestGeneration !== contextRefreshGeneration
        ) {
          return false;
        }
        memberships.value = data.memberships ?? [];
        globalRoles.value = data.globalRoles ?? [];
        defaultContext.value = data.defaultContext;
        reconcileActiveContext();
        persistCurrentSession();
        return true;
      })
      .catch(() => {
        // Keep the local persisted context for transient refresh failures.
        return false;
      })
      .finally(() => {
        if (
          token.value === requestToken &&
          requestGeneration === contextRefreshGeneration
        ) {
          isContextReady.value = true;
          isContextInitialized.value = true;
        }
        if (contextRefreshPromise === request) {
          contextRefreshPromise = null;
        }
      });

    contextRefreshPromise = request;
    return contextRefreshPromise;
  }

  function reconcileActiveContext() {
    const previousContext = activeContext.value;
    activeContext.value = chooseActiveContext({
      preferredContext: activeContext.value,
      defaultContext: defaultContext.value,
    });
    if (!sameNullableContext(previousContext, activeContext.value)) {
      contextVersion.value += 1;
      useActiveClassStore().reset();
      window.dispatchEvent(new CustomEvent("wiyata:context-changed"));
    }
  }

  function switchContext(targetContext: ActiveContext) {
    const validContext = availableContexts.value.find((context) =>
      sameContext(context, targetContext),
    );
    if (!validContext) {
      return null;
    }

    activeContext.value = validContext;
    contextVersion.value += 1;
    useActiveClassStore().reset();
    persistCurrentSession();
    window.dispatchEvent(new CustomEvent("wiyata:context-changed"));
    return landingRouteForContext(validContext);
  }

  function hasAnyRole(roles: RoleName[]) {
    if (roles.length === 0) return true;
    return activeRole.value ? roles.includes(activeRole.value) : false;
  }

  function primaryRole() {
    return activeRole.value;
  }

  function landingRoute(context: ActiveContext | null = activeContext.value) {
    return landingRouteForContext(context);
  }

  function persistCurrentSession() {
    if (!token.value) return;
    // For platform context (super_admin), activeSchoolId.value is null because the
    // computed only returns a schoolId for type:'school' contexts. However, backend
    // RequireRole always needs a SchoolId to look up roles — and for super_admin that
    // school is always the system school (code "000000"), guaranteed by CreateSuperAdmin.
    // Fall back to the system school UUID from memberships so the API interceptor
    // can include it in requests, restoring the pre-ab26e05 behaviour.
    const superAdminSchoolId = import.meta.env.VITE_SUPERADMIN_SCHOOL_ID;
    console.log("SUPERADMIN_SCHOOL_ID", superAdminSchoolId);
    const effectiveSchoolId =
      activeSchoolId.value ??
      memberships.value.find((m) => m.school.code === superAdminSchoolId)
        ?.school.id ??
      null;
    persistSession({
      token: token.value,
      user: user.value,
      memberships: memberships.value,
      globalRoles: globalRoles.value,
      defaultContext: defaultContext.value,
      activeContext: activeContext.value,
      activeSchoolId: effectiveSchoolId,
      activeRoles: activeRoles.value,
    });
  }

  function chooseActiveContext(options: {
    preferredContext?: ActiveContext | null;
    legacySchoolId?: string | null;
    defaultContext?: DefaultContext;
  }) {
    if (options.preferredContext && isValidContext(options.preferredContext)) {
      return matchingContext(options.preferredContext);
    }

    if (options.legacySchoolId) {
      const legacyContext = contextForSchool(options.legacySchoolId);
      if (legacyContext) return legacyContext;
    }

    if (options.defaultContext) {
      const defaultSchoolContext = contextForDefault(options.defaultContext);
      if (defaultSchoolContext) return defaultSchoolContext;
    }

    return availableContexts.value[0] ?? null;
  }

  function contextForDefault(context: DefaultContext) {
    const membership = memberships.value.find(
      (item) =>
        item.school.id === context.schoolId &&
        item.schoolUserId === context.schoolUserId,
    );
    if (!membership) return null;
    return contextForMembership(membership, context.roles);
  }

  function contextForSchool(schoolId: string) {
    const membership = memberships.value.find(
      (item) => item.school.id === schoolId,
    );
    if (!membership) return null;
    return contextForMembership(membership, membership.roles);
  }

  function contextForMembership(
    membership: MembershipInfo,
    preferredRoles: RoleName[],
  ) {
    const role =
      schoolRolePriority.find(
        (candidate) =>
          membership.roles.includes(candidate) &&
          preferredRoles.includes(candidate),
      ) ??
      schoolRolePriority.find((candidate) =>
        membership.roles.includes(candidate),
      );
    if (!role) return null;
    return {
      type: "school",
      schoolId: membership.school.id,
      schoolUserId: membership.schoolUserId,
      role,
    } satisfies ActiveContext;
  }

  function isValidContext(context: ActiveContext) {
    return availableContexts.value.some((item) => sameContext(item, context));
  }

  function matchingContext(context: ActiveContext) {
    return (
      availableContexts.value.find((item) => sameContext(item, context)) ?? null
    );
  }

  return {
    token,
    user,
    memberships,
    globalRoles,
    defaultContext,
    activeContext,
    activeRole,
    activeSchoolId,
    activeMembership,
    activeSchoolUserId,
    activeRoles,
    availableContexts,
    contextVersion,
    isAuthenticated,
    isContextReady,
    allRoles,
    login,
    logout,
    restoreSession,
    ensureUserContext,
    refreshUserContext,
    reconcileActiveContext,
    switchContext,
    hasAnyRole,
    primaryRole,
    landingRoute,
  };
});

function sameContext(a: ActiveContext, b: ActiveContext) {
  if (a.type !== b.type) return false;
  if (a.type === "platform" && b.type === "platform") return a.role === b.role;
  if (a.type === "school" && b.type === "school") {
    return (
      a.schoolId === b.schoolId &&
      a.schoolUserId === b.schoolUserId &&
      a.role === b.role
    );
  }
  return false;
}

function sameNullableContext(a: ActiveContext | null, b: ActiveContext | null) {
  if (!a && !b) return true;
  if (!a || !b) return false;
  return sameContext(a, b);
}

function landingRouteForContext(context: ActiveContext | null) {
  if (!context) return "/unauthorized";
  return landingRouteByRole[context.role];
}
