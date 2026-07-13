<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from "vue";
import { useRouter } from "vue-router";
import { PhBuildings, PhCaretDown, PhCheck } from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import { useToastStore } from "../../stores/toast";
import type { ActiveContext, RoleName, SchoolInfo } from "../../types/auth";

interface SchoolContextOption {
  context: ActiveContext & { type: "school" };
  roleLabel: string;
  isActive: boolean;
}

interface SchoolContextGroup {
  school: SchoolInfo;
  initials: string;
  options: SchoolContextOption[];
}

const roleLabels: Record<RoleName, string> = {
  admin: "Admin",
  teacher: "Guru",
  student: "Siswa",
  super_admin: "Super Admin",
};

const auth = useAuthStore();
const router = useRouter();
const toast = useToastStore();

const isOpen = ref(false);
const isSwitching = ref(false);
const rootEl = ref<HTMLElement | null>(null);
const triggerEl = ref<HTMLButtonElement | null>(null);
const menuId = `context-switcher-${Math.random().toString(36).slice(2, 9)}`;

const activeDescription = computed(() => describeContext(auth.activeContext));
const hasMultipleContexts = computed(() => auth.availableContexts.length > 1);

const schoolGroups = computed<SchoolContextGroup[]>(() => {
  const groups = new Map<string, SchoolContextGroup>();

  for (const context of auth.availableContexts) {
    if (context.type !== "school") continue;
    const membership = auth.memberships.find(
      (item) =>
        item.school.id === context.schoolId &&
        item.schoolUserId === context.schoolUserId,
    );
    if (!membership) continue;

    const existing = groups.get(membership.school.id);
    const group = existing ?? {
      school: membership.school,
      initials: initials(membership.school.name || membership.school.code),
      options: [],
    };

    group.options.push({
      context,
      roleLabel: roleLabels[context.role],
      isActive: isSameContext(context, auth.activeContext),
    });
    groups.set(membership.school.id, group);
  }

  return [...groups.values()];
});

const platformContext = computed(() =>
  auth.availableContexts.find(
    (context): context is ActiveContext & { type: "platform" } =>
      context.type === "platform",
  ),
);

const isPlatformActive = computed(() =>
  platformContext.value
    ? isSameContext(platformContext.value, auth.activeContext)
    : false,
);

function toggleMenu() {
  if (!hasMultipleContexts.value || isSwitching.value) return;
  isOpen.value = !isOpen.value;
}

function closeMenu() {
  isOpen.value = false;
}

async function selectContext(target: ActiveContext) {
  if (isSwitching.value || isSameContext(target, auth.activeContext)) return;

  const previousContext = auth.activeContext;
  isSwitching.value = true;
  try {
    const landingRoute = auth.switchContext(target);
    if (!landingRoute) {
      toast.error("Konteks tersebut tidak tersedia untuk akun ini.");
      return;
    }
    closeMenu();
    await router.push(landingRoute);
  } catch {
    if (previousContext) {
      auth.switchContext(previousContext);
    }
    toast.error("Konteks belum bisa diganti. Coba lagi sebentar lagi.");
  } finally {
    isSwitching.value = false;
    await nextTick();
    triggerEl.value?.focus();
  }
}

function handleDocumentPointerDown(event: PointerEvent) {
  const root = rootEl.value;
  if (!root || root.contains(event.target as Node)) return;
  closeMenu();
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === "Escape") {
    closeMenu();
    triggerEl.value?.focus();
  }
}

onMounted(() => {
  document.addEventListener("pointerdown", handleDocumentPointerDown);
  document.addEventListener("keydown", handleKeydown);
});

onUnmounted(() => {
  document.removeEventListener("pointerdown", handleDocumentPointerDown);
  document.removeEventListener("keydown", handleKeydown);
});

function describeContext(context: ActiveContext | null) {
  if (context?.type === "platform") {
    return {
      title: "Platform Wiyata",
      subtitle: roleLabels.super_admin,
      initials: "PW",
      ariaLabel: "Konteks aktif Platform Wiyata, Super Admin",
    };
  }

  if (context?.type === "school") {
    const membership = auth.memberships.find(
      (item) =>
        item.school.id === context.schoolId &&
        item.schoolUserId === context.schoolUserId,
    );
    const schoolName = membership?.school.name || "Sekolah aktif";
    const roleLabel = roleLabels[context.role];
    return {
      title: schoolName,
      subtitle: roleLabel,
      initials: initials(schoolName),
      ariaLabel: `Konteks aktif ${schoolName}, ${roleLabel}`,
    };
  }

  return {
    title: "Konteks belum tersedia",
    subtitle: "Pilih konteks",
    initials: "WY",
    ariaLabel: "Konteks aktif belum tersedia",
  };
}

function initials(value: string) {
  const words = value
    .split(/\s+/)
    .map((item) => item.trim())
    .filter(Boolean);
  const source =
    words.length > 1 ? `${words[0][0]}${words[1][0]}` : value.slice(0, 2);
  return source.toUpperCase();
}

function isSameContext(
  a: ActiveContext | null | undefined,
  b: ActiveContext | null | undefined,
) {
  if (!a || !b || a.type !== b.type) return false;
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
</script>

<template>
  <div ref="rootEl" class="relative">
    <!-- Trigger: context name + role row -->
    <button
      ref="triggerEl"
      type="button"
      class="flex items-center gap-2.5 rounded-xl border border-border bg-surface-subtle px-2.5 py-2 text-left transition hover:border-[#d8d2c6] hover:bg-surface-strong focus:outline-none focus:ring-2 focus:ring-brand-line focus:ring-offset-2 focus:ring-offset-white disabled:cursor-default disabled:opacity-80"
      :class="isOpen ? 'border-brand-line bg-brand-soft' : ''"
      :disabled="!hasMultipleContexts || isSwitching"
      :aria-haspopup="hasMultipleContexts ? 'menu' : undefined"
      :aria-expanded="hasMultipleContexts ? isOpen : undefined"
      :aria-controls="hasMultipleContexts ? menuId : undefined"
      :aria-label="`${activeDescription.ariaLabel}${hasMultipleContexts ? '. Buka pilihan konteks.' : ''}`"
      @click="toggleMenu"
    >
      <span class="min-w-0 flex-1">
        <span class="block truncate text-[13px] font-semibold text-foreground">
          {{ activeDescription.title }}
        </span>
        <span class="text-[11px] text-[#7c7789]">{{
          activeDescription.subtitle
        }}</span>
      </span>
      <PhCaretDown
        v-if="hasMultipleContexts"
        :size="13"
        class="shrink-0 text-[#9b9589] transition-transform duration-150"
        :class="isOpen ? 'rotate-180' : ''"
      />
    </button>

    <!-- ── Dropdown (collapsed: opens right; expanded: opens below) -->
    <Transition
      enter-active-class="transition duration-150 ease-out"
      enter-from-class="scale-95 opacity-0"
      enter-to-class="scale-100 opacity-100"
      leave-active-class="transition duration-100 ease-in"
      leave-from-class="scale-100 opacity-100"
      leave-to-class="scale-95 opacity-0"
    >
      <div
        v-if="isOpen"
        :id="menuId"
        role="menu"
        class="absolute left-0 top-full z-50 mt-1.5 w-72 rounded-xl border border-border bg-surface p-2 text-left shadow-xl shadow-[#2f2b3a]/10"
        aria-label="Pilih konteks sekolah dan peran"
      >
        <div class="border-b border-border px-3 py-2">
          <p
            class="truncate text-sm font-semibold text-foreground"
            :title="activeDescription.title"
          >
            {{ activeDescription.title }}
          </p>
          <p class="text-xs text-[#7c7789]">{{ activeDescription.subtitle }}</p>
        </div>

        <div class="max-h-[70vh] overflow-y-auto py-2">
          <section v-if="schoolGroups.length" class="space-y-2">
            <div
              v-for="group in schoolGroups"
              :key="group.school.id"
              class="space-y-1"
            >
              <p
                class="truncate px-3 text-[11px] font-semibold uppercase tracking-[0.08em] text-[#9b9589]"
                :title="group.school.name"
              >
                {{ group.school.name }}
              </p>
              <button
                v-for="option in group.options"
                :key="`${option.context.schoolUserId}-${option.context.role}`"
                type="button"
                role="menuitem"
                class="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-left transition hover:bg-background focus:bg-background focus:outline-none disabled:cursor-default disabled:bg-brand-soft"
                :disabled="option.isActive || isSwitching"
                :aria-current="option.isActive ? 'true' : undefined"
                @click="selectContext(option.context)"
              >
                <span
                  class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-surface-strong text-[11px] font-semibold text-[#575269]"
                >
                  {{ group.initials }}
                </span>
                <span class="min-w-0 flex-1">
                  <span
                    class="block truncate text-sm font-medium text-foreground"
                    :title="group.school.name"
                  >
                    {{ group.school.name }}
                  </span>
                  <span class="text-xs text-[#7c7789]">{{
                    option.roleLabel
                  }}</span>
                </span>
                <PhCheck
                  v-if="option.isActive"
                  :size="16"
                  class="shrink-0 text-brand"
                />
              </button>
            </div>
          </section>

          <section
            v-if="platformContext"
            class="mt-2 border-t border-border pt-2"
          >
            <p
              class="px-3 text-[11px] font-semibold uppercase tracking-[0.08em] text-[#9b9589]"
            >
              Platform
            </p>
            <button
              type="button"
              role="menuitem"
              class="mt-1 flex w-full items-center gap-3 rounded-lg px-3 py-2 text-left transition hover:bg-background focus:bg-background focus:outline-none disabled:cursor-default disabled:bg-brand-soft"
              :disabled="isPlatformActive || isSwitching"
              :aria-current="isPlatformActive ? 'true' : undefined"
              @click="selectContext(platformContext)"
            >
              <span
                class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-brand-soft text-brand"
              >
                <PhBuildings :size="17" />
              </span>
              <span class="min-w-0 flex-1">
                <span class="block truncate text-sm font-medium text-foreground">
                  Platform Wiyata
                </span>
                <span class="text-xs text-[#7c7789]">Super Admin</span>
              </span>
              <PhCheck
                v-if="isPlatformActive"
                :size="16"
                class="shrink-0 text-brand"
              />
            </button>
          </section>
        </div>
      </div>
    </Transition>
  </div>
</template>
