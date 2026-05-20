<script setup lang="ts">
import { PhSignOut } from "@phosphor-icons/vue";
import { useRoute, useRouter } from "vue-router";
import { useAuthStore } from "../../stores/auth";
import type { NavItem } from "../../types/navigation";

defineProps<{
  items: NavItem[];
  label: string;
}>();

const auth = useAuthStore();
const route = useRoute();
const router = useRouter();

function logout() {
  auth.logout();
  router.push("/login");
}

function isActive(to: string) {
  return route.path === to || route.path.startsWith(`${to}/`);
}
</script>

<template>
  <aside
    class="flex h-full w-16 flex-col items-center border-r border-[#ebe7df] bg-white/95 px-0 py-4"
  >
    <div
      class="mb-3 flex h-9 w-9 items-center justify-center rounded-xl bg-[#4f46e5] text-[13px] font-medium text-white"
    >
      Ev
    </div>

    <nav class="flex flex-1 flex-col items-center gap-1" :aria-label="label">
      <RouterLink
        v-for="item in items"
        :key="item.label"
        :title="item.label"
        class="relative flex h-10 w-10 items-center justify-center rounded-xl text-[#a3a1aa] transition hover:bg-[#f3f1ec] hover:text-[#3f3a4a]"
        :class="isActive(item.to) ? 'bg-[#eef2ff] text-[#4f46e5]' : ''"
        :to="item.to"
      >
        <component :is="item.icon" :size="20" weight="regular" />
        <span
          v-if="item.hasDot"
          class="absolute right-2 top-2 h-1.5 w-1.5 rounded-full border border-white bg-[#4f46e5]"
        />
      </RouterLink>
    </nav>

    <button
      title="Logout"
      class="mb-3 flex h-10 w-10 items-center justify-center rounded-xl text-[#a3a1aa] transition hover:bg-[#f3f1ec] hover:text-[#3f3a4a]"
      type="button"
      @click="logout"
    >
      <PhSignOut :size="19" />
    </button>

    <div
      class="flex h-8 w-8 items-center justify-center rounded-full bg-[#4f46e5] text-[11px] font-medium text-white"
    >
      {{ auth.user?.fullName?.slice(0, 2).toUpperCase() || "EV" }}
    </div>
  </aside>
</template>
