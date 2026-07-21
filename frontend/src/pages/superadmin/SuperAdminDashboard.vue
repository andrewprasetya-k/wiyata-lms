<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import {
  PhArrowRight,
  PhBuildings,
  PhCheckCircle,
  PhCompass,
  PhIdentificationBadge,
  PhShieldCheck,
  PhUsers,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import {
  getSuperAdminSchools,
  getSuperAdminSchoolSummary,
} from "../../services/superAdminSchool";
import { getAdminUsers } from "../../services/adminUser";
import { getSuperAdminDashboard } from "../../services/superAdminDashboard";
import type {
  SuperAdminSchoolItem,
  SuperAdminSchoolSummary,
} from "../../types/superAdminSchool";
import type { SuperAdminDashboardSummary } from "../../types/superAdminDashboard";
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

const attention = ref<SuperAdminDashboardSummary | null>(null);
const attentionLoading = ref(true);
const attentionError = ref("");

const schoolsWithoutAdmin = computed(() => attention.value?.schoolsWithoutAdmin ?? []);
const schoolsWithoutAdminTotal = computed(
  () => attention.value?.schoolsWithoutAdminTotal ?? 0,
);
const schoolsWithoutSetup = computed(() => attention.value?.schoolsWithoutSetup ?? []);
const schoolsWithoutSetupTotal = computed(
  () => attention.value?.schoolsWithoutSetupTotal ?? 0,
);

const hasAttentionItems = computed(
  () => schoolsWithoutAdminTotal.value > 0 || schoolsWithoutSetupTotal.value > 0,
);

const monthLabelFormatter = new Intl.DateTimeFormat("id-ID", { month: "short" });

function formatTrendMonth(period: string) {
  const [year, month] = period.split("-").map(Number);
  if (!year || !month) return period;
  return monthLabelFormatter.format(new Date(year, month - 1, 1));
}

function trendHasData(points: { count: number }[]) {
  return points.some((p) => p.count > 0);
}

function trendBarHeight(count: number, points: { count: number }[]) {
  const max = Math.max(1, ...points.map((p) => p.count));
  return `${Math.max(0, (count / max) * 100)}%`;
}

const schoolGrowthTrend = computed(() => attention.value?.schoolGrowthTrend ?? []);
const userGrowthTrend = computed(() => attention.value?.userGrowthTrend ?? []);

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

async function loadAttention() {
  attentionLoading.value = true;
  attentionError.value = "";
  try {
    attention.value = await getSuperAdminDashboard();
  } catch {
    attention.value = null;
    attentionError.value = "Sekolah yang perlu tindak lanjut belum bisa dimuat.";
  } finally {
    attentionLoading.value = false;
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
  loadAttention();
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
        <!-- Needs Attention -->
        <section class="space-y-2.5">
          <h2 class="text-sm font-semibold text-foreground">Perlu Perhatian</h2>

          <div
            v-if="attentionLoading"
            class="rounded-xl border border-border bg-surface shadow-sm p-4"
          >
            <div class="h-14 animate-pulse rounded-lg bg-surface-strong" />
          </div>

          <template v-else>
          <RouterLink
            v-if="schoolsWithoutAdminTotal > 0"
            to="/superadmin/schools"
            class="flex items-start gap-3 rounded-xl border border-warning-line bg-warning-soft p-4 transition hover:brightness-95"
          >
            <PhWarningCircle
              :size="20"
              weight="duotone"
              class="mt-0.5 shrink-0 text-[#ea580c]"
            />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-foreground">
                {{ schoolsWithoutAdminTotal }} sekolah belum memiliki admin
              </p>
              <p class="mt-0.5 text-xs text-muted">
                Sekolah ini belum bisa melanjutkan pengaturan akademik tanpa
                admin sekolah.
              </p>
            </div>
            <PhArrowRight :size="14" class="mt-1 shrink-0 text-[#ea580c]" />
          </RouterLink>

          <RouterLink
            v-if="schoolsWithoutSetupTotal > 0"
            to="/superadmin/schools"
            class="flex items-start gap-3 rounded-xl border border-warning-line bg-warning-soft p-4 transition hover:brightness-95"
          >
            <PhWarningCircle
              :size="20"
              weight="duotone"
              class="mt-0.5 shrink-0 text-[#ea580c]"
            />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-foreground">
                {{ schoolsWithoutSetupTotal }} sekolah belum memiliki tahun
                ajaran aktif
              </p>
              <p class="mt-0.5 text-xs text-muted">
                Admin sekolah belum menyelesaikan pengaturan tahun ajaran di
                sekolah ini.
              </p>
            </div>
            <PhArrowRight :size="14" class="mt-1 shrink-0 text-[#ea580c]" />
          </RouterLink>

          <div
            v-if="!hasAttentionItems && !attentionError"
            class="flex items-center gap-3 rounded-xl border border-success-line bg-success-soft p-4"
          >
            <PhCheckCircle :size="20" weight="duotone" class="shrink-0 text-success" />
            <p class="text-sm font-medium text-foreground">
              Semua beres. Tidak ada sekolah yang membutuhkan tindak lanjut.
            </p>
          </div>

          <div
            v-else-if="!hasAttentionItems && attentionError"
            class="rounded-xl border border-border bg-surface-subtle p-4 text-sm leading-6 text-muted"
          >
            Sebagian data belum bisa dimuat, jadi status di atas mungkin belum
            lengkap.
          </div>
          </template>
        </section>

        <!-- Work Queue: Schools without admin -->
        <section class="rounded-xl border border-border bg-surface p-5 shadow-sm">
          <div class="mb-4 flex items-center justify-between gap-3">
            <div>
              <p class="text-sm font-semibold text-foreground">
                Sekolah Tanpa Admin
              </p>
              <p class="mt-0.5 text-xs text-muted">
                Sekolah aktif yang belum memiliki akun dengan peran admin.
              </p>
            </div>
            <RouterLink
              v-if="schoolsWithoutAdminTotal > 0"
              to="/superadmin/schools"
              class="shrink-0 text-xs font-medium text-brand transition hover:text-brand-hover"
            >
              Lihat semua
            </RouterLink>
          </div>

          <div v-if="attentionLoading" class="space-y-2">
            <div
              v-for="i in 3"
              :key="i"
              class="h-12 animate-pulse rounded-lg bg-surface-strong"
            />
          </div>

          <div
            v-else-if="attentionError"
            class="rounded-lg bg-surface-subtle p-4 text-sm leading-6 text-muted"
          >
            {{ attentionError }}
          </div>

          <div
            v-else-if="schoolsWithoutAdmin.length === 0"
            class="rounded-lg bg-surface-subtle px-4 py-8 text-center"
          >
            <p class="text-sm font-medium text-foreground">
              Semua sekolah sudah memiliki admin.
            </p>
          </div>

          <div v-else class="divide-y divide-[#f3f4f6]">
            <RouterLink
              v-for="school in schoolsWithoutAdmin"
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

        <!-- Work Queue: Schools without academic setup -->
        <section class="rounded-xl border border-border bg-surface p-5 shadow-sm">
          <div class="mb-4 flex items-center justify-between gap-3">
            <div>
              <p class="text-sm font-semibold text-foreground">
                Sekolah Tanpa Setup Akademik
              </p>
              <p class="mt-0.5 text-xs text-muted">
                Sekolah aktif yang belum memiliki tahun ajaran aktif.
              </p>
            </div>
            <RouterLink
              v-if="schoolsWithoutSetupTotal > 0"
              to="/superadmin/schools"
              class="shrink-0 text-xs font-medium text-brand transition hover:text-brand-hover"
            >
              Lihat semua
            </RouterLink>
          </div>

          <div v-if="attentionLoading" class="space-y-2">
            <div
              v-for="i in 3"
              :key="i"
              class="h-12 animate-pulse rounded-lg bg-surface-strong"
            />
          </div>

          <div
            v-else-if="attentionError"
            class="rounded-lg bg-surface-subtle p-4 text-sm leading-6 text-muted"
          >
            {{ attentionError }}
          </div>

          <div
            v-else-if="schoolsWithoutSetup.length === 0"
            class="rounded-lg bg-surface-subtle px-4 py-8 text-center"
          >
            <p class="text-sm font-medium text-foreground">
              Semua sekolah sudah memiliki tahun ajaran aktif.
            </p>
          </div>

          <div v-else class="divide-y divide-[#f3f4f6]">
            <RouterLink
              v-for="school in schoolsWithoutSetup"
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

          <div
            v-else-if="recentSchools.length === 0"
            class="rounded-lg bg-surface-subtle px-4 py-8 text-center"
          >
            <p class="text-sm font-medium text-foreground">
              Belum ada sekolah yang dibuat.
            </p>
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

        <!-- Platform Trends -->
        <section class="grid gap-3 sm:grid-cols-2">
          <article class="rounded-xl border border-border bg-surface p-5 shadow-sm">
            <div class="mb-4">
              <p class="text-sm font-semibold text-foreground">
                Pertumbuhan Sekolah
              </p>
              <p class="mt-0.5 text-xs text-muted">
                Sekolah baru per bulan, 6 bulan terakhir.
              </p>
            </div>

            <div v-if="attentionLoading" class="h-[180px] animate-pulse rounded-lg bg-surface-strong" />

            <div
              v-else-if="attentionError"
              class="flex h-[180px] items-center justify-center rounded-lg bg-surface-subtle p-4 text-center text-sm leading-6 text-muted"
            >
              Tren pertumbuhan sekolah belum bisa dimuat.
            </div>

            <div
              v-else-if="!trendHasData(schoolGrowthTrend)"
              class="flex h-[180px] items-center justify-center rounded-lg bg-surface-subtle p-4 text-center"
            >
              <p class="text-sm font-medium text-foreground">Belum cukup data.</p>
            </div>

            <div v-else class="flex h-[180px] flex-col">
              <div class="flex flex-1 items-end gap-2">
                <div
                  v-for="point in schoolGrowthTrend"
                  :key="point.period"
                  class="flex min-w-0 flex-1 flex-col items-center justify-end gap-1.5"
                >
                  <span class="text-[10px] font-medium text-muted">{{ point.count }}</span>
                  <div class="w-full rounded-t bg-brand" :style="{ height: trendBarHeight(point.count, schoolGrowthTrend) }" />
                </div>
              </div>
              <div class="mt-2 flex gap-2">
                <span
                  v-for="point in schoolGrowthTrend"
                  :key="point.period"
                  class="min-w-0 flex-1 text-center text-[10px] text-muted"
                >
                  {{ formatTrendMonth(point.period) }}
                </span>
              </div>
            </div>
          </article>

          <article class="rounded-xl border border-border bg-surface p-5 shadow-sm">
            <div class="mb-4">
              <p class="text-sm font-semibold text-foreground">
                Pertumbuhan Akun Pengguna
              </p>
              <p class="mt-0.5 text-xs text-muted">
                Akun baru per bulan, 6 bulan terakhir.
              </p>
            </div>

            <div v-if="attentionLoading" class="h-[180px] animate-pulse rounded-lg bg-surface-strong" />

            <div
              v-else-if="attentionError"
              class="flex h-[180px] items-center justify-center rounded-lg bg-surface-subtle p-4 text-center text-sm leading-6 text-muted"
            >
              Tren pertumbuhan akun belum bisa dimuat.
            </div>

            <div
              v-else-if="!trendHasData(userGrowthTrend)"
              class="flex h-[180px] items-center justify-center rounded-lg bg-surface-subtle p-4 text-center"
            >
              <p class="text-sm font-medium text-foreground">Belum cukup data.</p>
            </div>

            <div v-else class="flex h-[180px] flex-col">
              <div class="flex flex-1 items-end gap-2">
                <div
                  v-for="point in userGrowthTrend"
                  :key="point.period"
                  class="flex min-w-0 flex-1 flex-col items-center justify-end gap-1.5"
                >
                  <span class="text-[10px] font-medium text-muted">{{ point.count }}</span>
                  <div class="w-full rounded-t bg-info" :style="{ height: trendBarHeight(point.count, userGrowthTrend) }" />
                </div>
              </div>
              <div class="mt-2 flex gap-2">
                <span
                  v-for="point in userGrowthTrend"
                  :key="point.period"
                  class="min-w-0 flex-1 text-center text-[10px] text-muted"
                >
                  {{ formatTrendMonth(point.period) }}
                </span>
              </div>
            </div>
          </article>
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
