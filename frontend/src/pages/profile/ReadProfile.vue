<script setup lang="ts">
import { computed } from "vue";
import {
  PhBuildings,
  PhIdentificationCard,
  PhShieldCheck,
  PhUserCircle,
} from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import type { RoleName } from "../../types/auth";

const props = defineProps<{
  eyebrow: string;
  title: string;
  helper: string;
}>();

const auth = useAuthStore();

const activeMembership = computed(() => auth.activeMembership);
const roleLabels: Record<RoleName, string> = {
  student: "Student",
  teacher: "Teacher",
  admin: "Admin",
  super_admin: "Superadmin",
};

const currentRoles = computed(() => {
  const roles = auth.activeRoles.length > 0 ? auth.activeRoles : auth.allRoles;
  return roles.map((role) => roleLabels[role] ?? role).join(", ") || "-";
});

const globalRoles = computed(() =>
  auth.globalRoles.map((role) => roleLabels[role] ?? role).join(", ") || "-",
);

const initials = computed(() => {
  const name = auth.user?.fullName ?? "";
  const value = name
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0])
    .join("")
    .toUpperCase();
  return value || "EV";
});

const profileRows = computed(() => [
  { label: "Nama lengkap", value: auth.user?.fullName || "-" },
  { label: "Email", value: auth.user?.email || "-" },
  { label: "User ID", value: auth.user?.id || "-" },
  { label: "School User ID", value: activeMembership.value?.schoolUserId || "-" },
  { label: "Role aktif", value: currentRoles.value },
  { label: "Role global", value: globalRoles.value },
]);

const schoolRows = computed(() => [
  { label: "Sekolah aktif", value: activeMembership.value?.school.name || "-" },
  { label: "Kode sekolah", value: activeMembership.value?.school.code || "-" },
  { label: "School ID", value: activeMembership.value?.school.id || "-" },
  {
    label: "Default context",
    value: activeMembership.value?.isDefault ? "Ya" : "Tidak",
  },
]);
</script>

<template>
  <main class="min-h-screen flex-1 px-5 py-5 sm:px-6 lg:px-8">
    <section class="flex w-full max-w-none flex-col gap-5">
      <header
        class="rounded-[22px] bg-[#f0e9dd] px-5 py-5 shadow-sm ring-1 ring-black/5 md:px-6"
      >
        <p class="text-sm font-medium text-[#8a6d3b]">{{ props.eyebrow }}</p>
        <h1 class="mt-3 text-3xl font-medium text-[#171322] md:text-4xl">
          {{ props.title }}
        </h1>
        <p class="mt-3 max-w-2xl text-sm leading-6 text-[#6b6475]">
          {{ props.helper }}
        </p>
      </header>

      <section class="grid gap-5 xl:grid-cols-[0.8fr_1.2fr]">
        <article class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5">
          <div class="flex items-start gap-4">
            <div
              class="flex h-16 w-16 shrink-0 items-center justify-center rounded-2xl bg-[#4f46e5] text-lg font-medium text-white"
            >
              {{ initials }}
            </div>
            <div class="min-w-0">
              <p class="text-sm text-[#8a8494]">Akun EduVerse</p>
              <h2 class="mt-1 truncate text-2xl font-medium text-[#171322]">
                {{ auth.user?.fullName || "Nama tidak tersedia" }}
              </h2>
              <p class="mt-2 truncate text-sm text-[#6b6475]">
                {{ auth.user?.email || "Email tidak tersedia" }}
              </p>
            </div>
          </div>

          <div class="mt-6 rounded-[18px] bg-[#faf8f4] p-4">
            <div class="flex items-start gap-3">
              <PhShieldCheck
                :size="22"
                class="mt-0.5 shrink-0 text-[#4f46e5]"
                weight="duotone"
              />
              <p class="text-sm leading-6 text-[#6b6475]">
                Halaman ini bersifat read-only untuk MVP. Perubahan profil,
                password, dan avatar belum tersedia dari halaman ini.
              </p>
            </div>
          </div>
        </article>

        <div class="grid gap-5 lg:grid-cols-2">
          <article class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5">
            <div class="mb-5 flex items-center gap-3">
              <div
                class="flex h-11 w-11 items-center justify-center rounded-2xl bg-[#eef2ff] text-[#4f46e5]"
              >
                <PhIdentificationCard :size="22" weight="duotone" />
              </div>
              <div>
                <p class="text-sm font-medium text-[#171322]">Identitas akun</p>
                <p class="mt-1 text-xs text-[#8a8494]">
                  Data berasal dari sesi login aktif.
                </p>
              </div>
            </div>

            <dl class="space-y-3">
              <div
                v-for="row in profileRows"
                :key="row.label"
                class="rounded-2xl bg-[#faf8f4] p-4"
              >
                <dt class="text-xs font-medium uppercase tracking-wide text-[#9ca3af]">
                  {{ row.label }}
                </dt>
                <dd class="mt-1 wrap-break-word text-sm font-medium text-[#171322]">
                  {{ row.value }}
                </dd>
              </div>
            </dl>
          </article>

          <article class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5">
            <div class="mb-5 flex items-center gap-3">
              <div
                class="flex h-11 w-11 items-center justify-center rounded-2xl bg-[#ecfdf5] text-[#059669]"
              >
                <PhBuildings :size="22" weight="duotone" />
              </div>
              <div>
                <p class="text-sm font-medium text-[#171322]">Konteks sekolah</p>
                <p class="mt-1 text-xs text-[#8a8494]">
                  Mengikuti active school dari sesi saat ini.
                </p>
              </div>
            </div>

            <dl class="space-y-3">
              <div
                v-for="row in schoolRows"
                :key="row.label"
                class="rounded-2xl bg-[#faf8f4] p-4"
              >
                <dt class="text-xs font-medium uppercase tracking-wide text-[#9ca3af]">
                  {{ row.label }}
                </dt>
                <dd class="mt-1 wrap-break-word text-sm font-medium text-[#171322]">
                  {{ row.value }}
                </dd>
              </div>
            </dl>
          </article>
        </div>
      </section>

      <section
        v-if="auth.memberships.length > 0"
        class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5"
      >
        <div class="flex items-center gap-3">
          <div
            class="flex h-11 w-11 items-center justify-center rounded-2xl bg-[#fff7ed] text-[#ea580c]"
          >
            <PhUserCircle :size="22" weight="duotone" />
          </div>
          <div>
            <p class="text-sm font-medium text-[#171322]">Membership sekolah</p>
            <p class="mt-1 text-xs text-[#8a8494]">
              Daftar school membership yang tersedia di sesi login.
            </p>
          </div>
        </div>

        <div class="mt-5 grid gap-3 md:grid-cols-2 xl:grid-cols-3">
          <article
            v-for="membership in auth.memberships"
            :key="membership.schoolUserId"
            class="rounded-[18px] border border-[#ebe7df] bg-white p-4"
          >
            <p class="text-sm font-medium text-[#171322]">
              {{ membership.school.name || "Sekolah tidak tersedia" }}
            </p>
            <p class="mt-1 text-xs text-[#8a8494]">
              {{ membership.school.code || "Kode tidak tersedia" }}
            </p>
            <div class="mt-4 flex flex-wrap gap-2">
              <span
                v-for="role in membership.roles"
                :key="role"
                class="rounded-full bg-[#eef2ff] px-3 py-1 text-xs font-medium text-[#4f46e5]"
              >
                {{ roleLabels[role] ?? role }}
              </span>
              <span
                v-if="membership.isDefault"
                class="rounded-full bg-[#ecfdf5] px-3 py-1 text-xs font-medium text-[#059669]"
              >
                Default
              </span>
            </div>
          </article>
        </div>
      </section>
    </section>
  </main>
</template>
