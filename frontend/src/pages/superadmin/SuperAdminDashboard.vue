<script setup lang="ts">
import { onMounted, ref } from "vue";
import {
  PhArrowRight,
  PhBuildings,
  PhCompass,
  PhIdentificationBadge,
  PhShieldCheck,
  PhUsers,
} from "@phosphor-icons/vue";
import {
  getSuperAdminSchools,
  getSuperAdminSchoolSummary,
} from "../../services/superAdminSchool";
import { getAdminUsers } from "../../services/adminUser";
import type {
  SuperAdminSchoolItem,
  SuperAdminSchoolSummary,
} from "../../types/superAdminSchool";
import { formatDate } from "../../utils/date";

const summary = ref<SuperAdminSchoolSummary | null>(null);
const summaryLoading = ref(true);
const summaryError = ref("");

const recentSchools = ref<SuperAdminSchoolItem[]>([]);
const recentSchoolsLoading = ref(true);
const recentSchoolsError = ref("");

const totalUsers = ref<number | null>(null);
const usersLoading = ref(true);
const usersError = ref("");

async function loadSummary() {
  summaryLoading.value = true;
  summaryError.value = "";
  try {
    summary.value = await getSuperAdminSchoolSummary();
  } catch {
    summary.value = null;
    summaryError.value = "Ringkasan sekolah belum bisa dimuat.";
  } finally {
    summaryLoading.value = false;
  }
}

async function loadRecentSchools() {
  recentSchoolsLoading.value = true;
  recentSchoolsError.value = "";
  try {
    const response = await getSuperAdminSchools({
      page: 1,
      limit: 5,
      status: "active",
    });
    recentSchools.value = response.data ?? [];
  } catch {
    recentSchools.value = [];
    recentSchoolsError.value = "Sekolah terbaru belum bisa dimuat.";
  } finally {
    recentSchoolsLoading.value = false;
  }
}

async function loadTotalUsers() {
  usersLoading.value = true;
  usersError.value = "";
  try {
    const response = await getAdminUsers({ page: 1, limit: 1 });
    totalUsers.value = Number(response.totalItems ?? 0);
  } catch {
    totalUsers.value = null;
    usersError.value = "Jumlah akun belum bisa dimuat.";
  } finally {
    usersLoading.value = false;
  }
}

const overviewCards = [
  {
    title: "Sekolah",
    description:
      "Kelola tenant sekolah Wiyata dari tingkat platform tanpa mengambil alih operasional akademik harian.",
    icon: PhBuildings,
    tone: "bg-[#fff4ee] text-[#ea580c]",
  },
  {
    title: "Akun Global",
    description:
      "Pantau identitas pengguna lintas sekolah dan siapkan akun awal untuk kebutuhan platform.",
    icon: PhUsers,
    tone: "bg-brand-soft text-brand",
  },
  {
    title: "Pengaturan Tenant",
    description:
      "Bantu proses awal agar sekolah punya admin sekolah yang dapat melanjutkan pengaturan akademik.",
    icon: PhCompass,
    tone: "bg-success-soft text-success",
  },
  {
    title: "Peran Platform",
    description:
      "Super Admin menjaga akses platform. Operasional akademik tetap berada di area Admin Sekolah.",
    icon: PhShieldCheck,
    tone: "bg-surface-strong text-muted",
  },
];

const setupFlow = [
  "Buat sekolah atau tenant sekolah.",
  "Buat atau pilih akun Admin Sekolah.",
  "Hubungkan akun ke sekolah yang tepat.",
  "Beri peran admin sekolah.",
  "Admin Sekolah melanjutkan pengaturan akademik.",
];

const quickActions = [
  {
    title: "Kelola sekolah",
    description: "Buka area sekolah platform yang saat ini masih disiapkan.",
    to: "/superadmin/schools",
    icon: PhBuildings,
  },
  {
    title: "Akun global",
    description: "Buka area akun global yang saat ini masih disiapkan.",
    to: "/superadmin/users",
    icon: PhUsers,
  },
  {
    title: "Profil platform",
    description: "Lihat informasi akun Super Admin yang sedang digunakan.",
    to: "/superadmin/profile",
    icon: PhIdentificationBadge,
  },
];

onMounted(() => {
  loadSummary();
  loadRecentSchools();
  loadTotalUsers();
});
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-background">
    <header class="border-b border-border bg-surface">
      <div
        class="flex min-w-0 flex-col gap-3 px-5 py-5 sm:px-6 lg:flex-row lg:items-end lg:justify-between lg:px-8"
      >
        <div class="min-w-0">
          <p
            class="eyebrow"
          >
            Super Admin
          </p>
          <h1 class="mt-2 text-2xl font-semibold text-foreground sm:text-3xl">
            Pusat Platform Wiyata
          </h1>
          <p class="mt-2 max-w-3xl text-sm leading-6 text-muted">
            Kelola kesiapan tenant sekolah dan akun global. Pengaturan akademik,
            kelas, konten, dan penilaian tetap berada di area Admin Sekolah.
          </p>
        </div>
        <RouterLink
          to="/superadmin/schools"
          class="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-[#ea580c] px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-[#c2410c] sm:w-auto"
        >
          Buka sekolah
          <PhArrowRight :size="16" weight="bold" />
        </RouterLink>
      </div>
    </header>

    <section
      class="grid w-full max-w-none gap-6 px-5 py-6 sm:px-6 lg:px-8 xl:grid-cols-[minmax(0,1fr)_340px]"
    >
      <div class="flex min-w-0 flex-col gap-6">
        <!-- Overview: Total Schools + Total Users -->
        <section class="grid gap-3 sm:grid-cols-2">
          <RouterLink
            to="/superadmin/schools"
            class="rounded-xl border border-border bg-surface p-4 shadow-sm transition hover:-translate-y-0.5 hover:shadow-md"
          >
            <div
              class="flex h-10 w-10 items-center justify-center rounded-xl bg-[#fff4ee] text-[#ea580c]"
            >
              <PhBuildings :size="20" weight="duotone" />
            </div>
            <template v-if="summaryLoading">
              <div class="mt-4 h-8 w-16 animate-pulse rounded bg-surface-strong" />
            </template>
            <template v-else-if="summaryError">
              <p class="mt-4 text-sm leading-6 text-muted">{{ summaryError }}</p>
            </template>
            <template v-else>
              <p class="mt-4 text-2xl font-semibold text-foreground">
                {{ summary?.totalSchools ?? 0 }}
              </p>
              <p class="mt-1 text-xs text-muted">
                Total Sekolah · {{ summary?.totalActive ?? 0 }} aktif,
                {{ summary?.totalDeleted ?? 0 }} dihapus
              </p>
            </template>
          </RouterLink>

          <RouterLink
            to="/superadmin/users"
            class="rounded-xl border border-border bg-surface p-4 shadow-sm transition hover:-translate-y-0.5 hover:shadow-md"
          >
            <div
              class="flex h-10 w-10 items-center justify-center rounded-xl bg-brand-soft text-brand"
            >
              <PhUsers :size="20" weight="duotone" />
            </div>
            <template v-if="usersLoading">
              <div class="mt-4 h-8 w-16 animate-pulse rounded bg-surface-strong" />
            </template>
            <template v-else-if="usersError">
              <p class="mt-4 text-sm leading-6 text-muted">{{ usersError }}</p>
            </template>
            <template v-else>
              <p class="mt-4 text-2xl font-semibold text-foreground">
                {{ totalUsers ?? 0 }}
              </p>
              <p class="mt-1 text-xs text-muted">Total Akun Pengguna</p>
            </template>
          </RouterLink>
        </section>

        <!-- Recently Created Schools -->
        <section
          v-if="recentSchoolsLoading || recentSchools.length > 0 || recentSchoolsError"
          class="rounded-xl border border-border bg-surface p-5 shadow-sm"
        >
          <div class="mb-4 flex items-center justify-between gap-3">
            <div>
              <p class="text-sm font-semibold text-foreground">
                Sekolah Terbaru
              </p>
              <p class="mt-0.5 text-xs text-muted">
                Sekolah yang paling baru dibuat di platform.
              </p>
            </div>
            <RouterLink
              v-if="recentSchools.length > 0"
              to="/superadmin/schools"
              class="shrink-0 text-xs font-medium text-brand transition hover:text-brand-hover"
            >
              Lihat semua
            </RouterLink>
          </div>

          <div v-if="recentSchoolsLoading" class="space-y-2">
            <div
              v-for="i in 3"
              :key="i"
              class="h-12 animate-pulse rounded-lg bg-surface-strong"
            />
          </div>

          <div
            v-else-if="recentSchoolsError"
            class="rounded-lg bg-surface-subtle p-4 text-sm leading-6 text-muted"
          >
            {{ recentSchoolsError }}
          </div>

          <div v-else class="divide-y divide-[#f3f4f6]">
            <RouterLink
              v-for="school in recentSchools"
              :key="school.schoolId"
              to="/superadmin/schools"
              class="flex items-center justify-between gap-3 py-2.5 first:pt-0 last:pb-0"
            >
              <div class="min-w-0 flex-1">
                <p class="truncate text-sm font-medium text-foreground">
                  {{ school.schoolName }}
                </p>
                <p class="mt-0.5 truncate text-xs text-muted">
                  {{ school.schoolCode }}
                </p>
              </div>
              <span class="shrink-0 text-xs text-muted">
                Dibuat {{ formatDate(school.createdAt) }}
              </span>
            </RouterLink>
          </div>
        </section>

        <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
          <article
            v-for="item in overviewCards"
            :key="item.title"
            class="rounded-xl border border-border bg-surface p-4 shadow-sm"
          >
            <div
              class="flex h-10 w-10 items-center justify-center rounded-xl"
              :class="item.tone"
            >
              <component :is="item.icon" :size="20" weight="duotone" />
            </div>
            <h3 class="mt-4 text-sm font-semibold text-foreground">
              {{ item.title }}
            </h3>
            <p class="mt-2 text-xs leading-5 text-muted">
              {{ item.description }}
            </p>
          </article>
        </section>

        <article
          class="rounded-xl border border-border bg-surface p-5 shadow-sm"
        >
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <p class="text-sm font-semibold text-foreground">
                Alur pengaturan awal tenant
              </p>
              <p class="mt-1 text-xs leading-5 text-muted">
                Urutan kerja aman sebelum Admin Sekolah mulai mengelola data
                akademik.
              </p>
            </div>
            <PhCompass
              :size="20"
              class="shrink-0 text-[#ea580c]"
              weight="duotone"
            />
          </div>

          <div class="mt-4 divide-y divide-[#f3f4f6]">
            <div
              v-for="(item, index) in setupFlow"
              :key="item"
              class="flex items-start gap-3 py-3"
            >
              <span
                class="mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-[#fff4ee] text-xs font-semibold text-[#ea580c]"
              >
                {{ index + 1 }}
              </span>
              <p class="min-w-0 text-sm leading-6 text-foreground-secondary">
                {{ item }}
              </p>
            </div>
          </div>
        </article>
      </div>

      <aside class="flex min-w-0 flex-col gap-5">
        <section
          class="rounded-xl border border-border bg-surface p-5 shadow-sm"
        >
          <p class="text-sm font-semibold text-foreground">Aksi cepat</p>
          <p class="mt-1 text-xs leading-5 text-muted">
            Semua tautan memakai halaman Super Admin yang sudah tersedia.
          </p>

          <div class="mt-4 space-y-2">
            <RouterLink
              v-for="item in quickActions"
              :key="item.to"
              :to="item.to"
              class="flex items-center gap-3 rounded-lg p-3 transition hover:bg-surface-strong"
            >
              <span
                class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-[#fff4ee] text-[#ea580c]"
              >
                <component :is="item.icon" :size="20" weight="duotone" />
              </span>
              <span class="min-w-0 flex-1">
                <span class="block text-sm font-semibold text-foreground">
                  {{ item.title }}
                </span>
                <span class="mt-1 block text-xs leading-5 text-muted">
                  {{ item.description }}
                </span>
              </span>
              <PhArrowRight :size="16" class="shrink-0 text-muted" />
            </RouterLink>
          </div>
        </section>
      </aside>
    </section>
  </main>
</template>
