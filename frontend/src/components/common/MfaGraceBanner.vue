<script setup lang="ts">
import { computed, ref } from "vue";
import { RouterLink } from "vue-router";
import { useAuthStore } from "../../stores/auth";
import type { RoleName } from "../../types/auth";

const auth = useAuthStore();
// Lightweight/session-scoped only — resets on a full reload, which is
// an acceptable trade-off for a low-stakes reminder banner (no need for
// persisted dismissal state for this).
const dismissed = ref(false);

const profileRouteNameByRole: Record<RoleName, string> = {
  super_admin: "superadmin-profile",
  admin: "admin-profile",
  teacher: "teacher-profile",
  student: "student-profile",
};

const profileRouteName = computed(() =>
  auth.activeRole ? profileRouteNameByRole[auth.activeRole] : null,
);
</script>

<template>
  <div
    v-if="!dismissed && auth.user && auth.mfaGraceDaysRemaining !== undefined"
    class="flex flex-col items-center justify-center gap-2 bg-brand-soft px-4 py-2.5 text-center text-sm text-[#3730a3] sm:flex-row"
  >
    <span>
      Aktifkan verifikasi dua langkah — tersisa
      {{ auth.mfaGraceDaysRemaining }} hari sebelum wajib diaktifkan.
    </span>
    <RouterLink
      v-if="profileRouteName"
      :to="{ name: profileRouteName }"
      class="font-medium underline underline-offset-2 transition hover:text-[#26216b]"
    >
      Aktifkan sekarang
    </RouterLink>
    <button
      type="button"
      class="text-xs font-medium text-[#3730a3]/70 underline underline-offset-2 transition hover:text-[#3730a3]"
      @click="dismissed = true"
    >
      Tutup
    </button>
  </div>
</template>
