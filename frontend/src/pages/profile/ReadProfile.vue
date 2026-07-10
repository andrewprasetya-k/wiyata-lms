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
  student: "Siswa",
  teacher: "Guru",
  admin: "Admin",
  super_admin: "Superadmin",
};

const titleLabels: Record<string, string> = {
  "Profil Student": "Profil Siswa",
  "Profil Teacher": "Profil Guru",
  "Profil Admin": "Profil Admin",
  "Profil Superadmin": "Profil Superadmin",
};

const eyebrowLabels: Record<string, string> = {
  "Student profile": "Akun siswa",
  "Teacher profile": "Akun guru",
  "Admin profile": "Akun admin",
  "Super admin profile": "Akun superadmin",
};

const helperLabels: Record<string, string> = {
  "Student profile":
    "Lihat informasi akun, peran, dan sekolah yang sedang aktif.",
  "Teacher profile":
    "Lihat informasi akun guru, peran aktif, dan akses sekolah.",
  "Admin profile":
    "Lihat informasi akun admin dan sekolah yang sedang dikelola.",
  "Super admin profile": "Lihat informasi akun dan peran pengelola platform.",
};

const pageTitle = computed(() => titleLabels[props.title] ?? props.title);
const pageEyebrow = computed(
  () => eyebrowLabels[props.eyebrow] ?? "Informasi akun",
);
const pageHelper = computed(
  () =>
    helperLabels[props.eyebrow] ||
    props.helper ||
    "Lihat informasi akun dan konteks aktif.",
);

const currentRoles = computed(() => {
  const roles = auth.activeRoles.length > 0 ? auth.activeRoles : auth.allRoles;
  return roles.map((role) => roleLabels[role] ?? role).join(", ") || "-";
});

const globalRoles = computed(
  () =>
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
  { label: "Peran aktif", value: currentRoles.value },
  { label: "Peran platform", value: globalRoles.value },
]);

const schoolRows = computed(() => [
  { label: "Sekolah aktif", value: activeMembership.value?.school.name || "-" },
  { label: "Kode sekolah", value: activeMembership.value?.school.code || "-" },
  {
    label: "Sekolah utama",
    value: activeMembership.value?.isDefault ? "Ya" : "Bukan",
  },
]);
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-[#f8f7f4]">
    <header class="border-b border-border bg-white">
      <div class="px-5 py-4 sm:px-6 lg:px-8">
        <p class="text-[11px] font-medium text-brand">
          {{ pageEyebrow }}
        </p>
        <h1 class="mt-1 text-xl font-medium text-foreground sm:text-2xl">
          {{ pageTitle }}
        </h1>
        <p class="mt-1 max-w-2xl text-xs leading-5 text-muted sm:text-sm">
          {{ pageHelper }}
        </p>
      </div>
    </header>

    <section
      v-if="!auth.user"
      class="flex min-h-[65vh] items-center justify-center px-5 py-10 sm:px-6 lg:px-8"
    >
      <article
        class="w-full max-w-xl rounded-xl border border-border bg-white p-8 text-center"
      >
        <div
          class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-[#eef2ff] text-brand"
        >
          <PhUserCircle :size="25" weight="duotone" />
        </div>
        <h2 class="mt-3 text-lg font-medium text-foreground">
          Data profil belum tersedia
        </h2>
        <p class="mx-auto mt-2 max-w-md text-sm leading-6 text-[#7a7385]">
          Informasi akun belum ditemukan pada sesi aktif. Silakan masuk kembali
          untuk memuat profil.
        </p>
      </article>
    </section>

    <section
      v-else
      class="mx-auto grid max-w-screen min-w-0 gap-5 px-5 py-5 sm:px-6 lg:grid-cols-[320px_minmax(0,1fr)] lg:px-8 lg:py-6"
    >
      <aside class="min-w-0">
        <article
          class="rounded-xl border border-border bg-white p-5 lg:sticky lg:top-6"
        >
          <div
            class="flex h-14 w-14 items-center justify-center rounded-xl bg-brand text-base font-medium text-white"
          >
            {{ initials }}
          </div>
          <p class="mt-4 text-xs text-[#8a8494]">Akun Wiyata</p>
          <h2 class="mt-1 wrap-break-word text-xl font-medium text-foreground">
            {{ auth.user.fullName || "Nama tidak tersedia" }}
          </h2>
          <p class="mt-1 break-all text-sm leading-6 text-[#6b6475]">
            {{ auth.user.email || "Email tidak tersedia" }}
          </p>

          <div class="mt-5 rounded-lg bg-[#eef2ff] p-4">
            <div class="flex items-start gap-3">
              <PhShieldCheck
                :size="20"
                class="mt-0.5 shrink-0 text-brand"
                weight="duotone"
              />
              <p class="text-sm leading-6 text-[#5f5a70]">
                Informasi profil ini hanya dapat dilihat. Perubahan akun
                dilakukan melalui pengelola sekolah.
              </p>
            </div>
          </div>
        </article>
      </aside>

      <div class="min-w-0 space-y-5">
        <section class="grid min-w-0 gap-5 xl:grid-cols-2">
          <article
            class="min-w-0 rounded-xl border border-border bg-white p-5"
          >
            <div class="mb-4 flex min-w-0 items-center gap-3">
              <div
                class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-[#eef2ff] text-brand"
              >
                <PhIdentificationCard :size="21" weight="duotone" />
              </div>
              <div class="min-w-0">
                <h2 class="text-sm font-medium text-foreground">
                  Identitas akun
                </h2>
                <p class="mt-1 text-xs text-[#8a8494]">
                  Informasi dari sesi yang sedang digunakan.
                </p>
              </div>
            </div>

            <dl class="divide-y divide-border">
              <div
                v-for="row in profileRows"
                :key="row.label"
                class="grid min-w-0 gap-1 py-3 first:pt-0 last:pb-0 sm:grid-cols-[120px_minmax(0,1fr)] sm:gap-4"
              >
                <dt class="text-xs text-[#8a8494]">{{ row.label }}</dt>
                <dd
                  class="wrap-break-word text-sm font-medium text-foreground sm:text-right"
                >
                  {{ row.value }}
                </dd>
              </div>
            </dl>
          </article>

          <article
            class="min-w-0 rounded-xl border border-border bg-white p-5"
          >
            <div class="mb-4 flex min-w-0 items-center gap-3">
              <div
                class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-[#f0fdf4] text-[#059669]"
              >
                <PhBuildings :size="21" weight="duotone" />
              </div>
              <div class="min-w-0">
                <h2 class="text-sm font-medium text-foreground">
                  Konteks sekolah
                </h2>
                <p class="mt-1 text-xs text-[#8a8494]">
                  Sekolah yang sedang digunakan pada sesi ini.
                </p>
              </div>
            </div>

            <dl class="divide-y divide-border">
              <div
                v-for="row in schoolRows"
                :key="row.label"
                class="grid min-w-0 gap-1 py-3 first:pt-0 last:pb-0 sm:grid-cols-[120px_minmax(0,1fr)] sm:gap-4"
              >
                <dt class="text-xs text-[#8a8494]">{{ row.label }}</dt>
                <dd
                  class="wrap-break-word text-sm font-medium text-foreground sm:text-right"
                >
                  {{ row.value }}
                </dd>
              </div>
            </dl>
          </article>
        </section>

        <section
          v-if="auth.memberships.length > 0"
          class="min-w-0 rounded-xl border border-border bg-white p-5"
        >
          <div class="flex min-w-0 items-center gap-3">
            <div
              class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-[#fff7ed] text-[#ea580c]"
            >
              <PhUserCircle :size="21" weight="duotone" />
            </div>
            <div class="min-w-0">
              <h2 class="text-sm font-medium text-foreground">Akses sekolah</h2>
              <p class="mt-1 text-xs text-[#8a8494]">
                Sekolah dan peran yang tersedia untuk akun ini.
              </p>
            </div>
          </div>

          <div class="mt-4 grid min-w-0 gap-3 md:grid-cols-2 xl:grid-cols-3">
            <article
              v-for="membership in auth.memberships"
              :key="membership.schoolUserId"
              class="min-w-0 rounded-lg border border-border bg-[#fbfaf8] p-4"
            >
              <p class="truncate text-sm font-medium text-foreground">
                {{ membership.school.name || "Sekolah tidak tersedia" }}
              </p>
              <p class="mt-1 truncate text-xs text-[#8a8494]">
                {{ membership.school.code || "Kode tidak tersedia" }}
              </p>
              <div class="mt-3 flex flex-wrap gap-2">
                <span
                  v-for="role in membership.roles"
                  :key="role"
                  class="rounded-full bg-[#eef2ff] px-2.5 py-1 text-xs font-medium text-brand"
                >
                  {{ roleLabels[role] ?? role }}
                </span>
                <span
                  v-if="membership.isDefault"
                  class="rounded-full bg-[#f0fdf4] px-2.5 py-1 text-xs font-medium text-[#059669]"
                >
                  Utama
                </span>
              </div>
            </article>
          </div>
        </section>

        <section
          v-else
          class="rounded-xl border border-border bg-white p-6 text-center"
        >
          <h2 class="text-sm font-medium text-foreground">
            Akses sekolah belum tersedia
          </h2>
          <p class="mt-2 text-sm leading-6 text-[#7a7385]">
            Akun ini belum memiliki akses sekolah pada sesi aktif.
          </p>
        </section>
      </div>
    </section>
  </main>
</template>
