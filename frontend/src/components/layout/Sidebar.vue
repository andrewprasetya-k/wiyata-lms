<script setup lang="ts">
import { ref } from "vue";
import { PhSignOut, PhSidebarSimple } from "@phosphor-icons/vue";
import { useRoute } from "vue-router";
import { useAuthStore } from "../../stores/auth";
import { useLogoutConfirm } from "../../composables/useLogoutConfirm";
import type { NavItem } from "../../types/navigation";

defineProps<{
  items: NavItem[];
  label: string;
  profileTo: string;
}>();

const auth = useAuthStore();
const route = useRoute();
const { confirmLogout } = useLogoutConfirm();

const isCollapsed = ref(
  localStorage.getItem("wiyata_sidebar_collapsed") === "true",
);

function toggle() {
  isCollapsed.value = !isCollapsed.value;
  localStorage.setItem("wiyata_sidebar_collapsed", String(isCollapsed.value));
}

function isActive(to: string) {
  return route.path === to || route.path.startsWith(`${to}/`);
}
</script>

<template>
  <aside
    class="flex h-full flex-col border-r border-border bg-surface/95 transition-[width] duration-200 ease-in-out"
    :class="isCollapsed ? 'w-18' : 'w-62'"
  >
    <!-- ── Header: logo + brand + toggle -->
    <div
      class="flex shrink-0 items-center"
      :class="isCollapsed ? 'flex-col gap-1.5 px-0 py-4' : 'gap-2.5 px-3 py-4'"
    >
      <img
        src="/logo_fix.svg"
        alt="Wiyata"
        class="h-7 w-7 shrink-0 rounded-lg object-contain"
      />

      <!-- Brand label — flex-1 only in expanded; h-0 flex-none in collapsed to take zero space in flex-col -->
      <span
        class="overflow-hidden whitespace-nowrap text-[15px] font-semibold tracking-tight text-foreground transition-[opacity,transform] duration-150"
        :class="
          isCollapsed
            ? 'pointer-events-none h-0 flex-none -translate-x-1 opacity-0'
            : 'flex-1 translate-x-0 opacity-100'
        "
        aria-hidden="true"
      >
        Wiyata Workspace
      </span>

      <!-- Collapse / expand toggle -->
      <button
        class="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg text-[#a3a1aa] transition hover:bg-surface-strong hover:text-[#3f3a4a]"
        :title="isCollapsed ? 'Buka sidebar' : 'Tutup sidebar'"
        type="button"
        @click="toggle"
      >
        <PhSidebarSimple
          :size="16"
          class="transition-transform duration-200"
          :class="isCollapsed ? 'rotate-180' : ''"
        />
      </button>
    </div>

    <!-- ── Navigation -->
    <nav
      class="flex flex-1 flex-col gap-0.5 overflow-y-auto px-2 py-2"
      :aria-label="label"
    >
      <RouterLink
        v-for="item in items"
        :key="item.label"
        :title="isCollapsed ? item.label : undefined"
        class="relative flex h-10 items-center rounded-xl text-[#a3a1aa] transition hover:bg-surface-strong hover:text-gray-700 cursor-pointer"
        :class="[
          isCollapsed ? 'mx-auto w-4 justify-center px-6' : 'w-full gap-3 px-3',
          isActive(item.to)
            ? 'bg-surface-strong text-foreground'
            : item.emphasized
              ? 'text-[#575269]'
              : '',
        ]"
        :to="item.to"
      >
        <component
          :is="item.icon"
          :size="20"
          weight="regular"
          class="shrink-0 text-muted"
        />

        <!-- Label — visible only when expanded -->
        <span
          class="flex-1 truncate text-sm font-medium transition-[opacity,transform] duration-150 text-muted"
          :class="
            isCollapsed
              ? 'pointer-events-none w-0 -translate-x-1 opacity-0'
              : 'translate-x-0 opacity-100'
          "
        >
          {{ item.label }}
        </span>

        <!-- hasDot indicator -->
        <span
          v-if="item.hasDot"
          class="absolute rounded-full border border-white bg-[#3f3a4a]"
          :class="
            isCollapsed
              ? 'right-1.5 top-1.5 h-1.5 w-1.5'
              : 'right-2 top-2 h-1.5 w-1.5'
          "
        />

        <!-- Badge count -->
        <Transition
          enter-active-class="transition duration-200 ease-out"
          enter-from-class="scale-95 opacity-0"
          enter-to-class="scale-100 opacity-100"
          leave-active-class="transition duration-150 ease-out"
          leave-from-class="scale-100 opacity-100"
          leave-to-class="scale-95 opacity-0"
          mode="out-in"
        >
          <span
            v-if="item.badgeCount"
            :key="item.badgeLabel || String(item.badgeCount)"
            class="absolute inline-flex min-w-[1.1rem] items-center justify-center rounded-full bg-brand px-1.5 py-0.5 text-[10px] font-semibold leading-none text-white shadow-sm"
            :class="isCollapsed ? '-right-1.5 -top-1 ' : 'right-2'"
            :aria-label="
              item.badgeAriaLabel || `${item.badgeCount} chat belum dibaca`
            "
          >
            {{ item.badgeLabel || item.badgeCount }}
          </span>
        </Transition>
      </RouterLink>
    </nav>

    <!-- ── Bottom: logout + profile -->
    <div class="shrink-0 space-y-2 px-2 py-4">
      <!-- Logout -->
      <button
        class="relative flex h-10 items-center rounded-xl text-danger transition hover:bg-danger hover:text-white/95 cursor-pointer"
        :class="
          isCollapsed ? 'mx-auto w-4 justify-center px-6' : 'w-full gap-3 px-3'
        "
        :title="isCollapsed ? 'Logout' : undefined"
        type="button"
        @click="confirmLogout"
      >
        <PhSignOut :size="19" class="shrink-0" />
        <span
          class="flex-1 truncate text-left text-sm font-medium transition-[opacity,transform] duration-150"
          :class="
            isCollapsed
              ? 'pointer-events-none w-4 -translate-x-1 opacity-0'
              : 'translate-x-0 opacity-100'
          "
        >
          Logout
        </span>
      </button>

      <!-- Profile -->
      <RouterLink
        :to="profileTo"
        :title="isCollapsed ? 'Buka profil' : undefined"
        :aria-label="isCollapsed ? 'Buka profil' : undefined"
        class="relative flex h-10 items-center rounded-xl transition hover:bg-surface-strong"
        :class="
          isCollapsed ? 'mx-auto w-5 justify-center px-6' : 'w-full gap-3 px-3'
        "
      >
        <span
          class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-brand text-[11px] font-medium text-white transition hover:bg-brand-hover"
          :class="
            isActive(profileTo)
              ? 'ring-2 ring-brand-line ring-offset-2 ring-offset-white'
              : ''
          "
        >
          {{ auth.user?.fullName?.slice(0, 2).toUpperCase() || "EV" }}
        </span>
        <span
          class="min-w-0 flex-1 transition-[opacity,transform] duration-150"
          :class="
            isCollapsed
              ? 'pointer-events-none w-0 -translate-x-1 opacity-0'
              : 'translate-x-0 opacity-100'
          "
        >
          <span class="block truncate text-sm font-medium text-foreground">
            {{ auth.user?.fullName || "Pengguna" }}
          </span>
        </span>
      </RouterLink>
    </div>
  </aside>
</template>
