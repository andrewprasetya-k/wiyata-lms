<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import {
  PhArrowRight,
  PhBookOpen,
  PhCalendarBlank,
  PhChalkboardTeacher,
  PhCheckCircle,
  PhCircleDashed,
  PhClipboardText,
  PhClockCountdown,
  PhUsers,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import { getAdminDashboard } from "../../services/adminDashboard";
import {
  getAcademicYearsBySchool,
  getTermsByAcademicYear,
} from "../../services/adminAcademic";
import type { AdminDashboardSummary } from "../../types/adminDashboard";
import type { AcademicYearItem, TermItem } from "../../types/adminAcademic";
import { formatDateTime } from "../../utils/date";
import LatestChatCard from "../../components/chat/LatestChatCard.vue";

const auth = useAuthStore();

const loading = ref(true);
const errorMessage = ref("");
const stats = ref<AdminDashboardSummary | null>(null);
const activeYear = ref<AcademicYearItem | null>(null);
const activeTerm = ref<TermItem | null>(null);

const schoolId = computed(() => auth.activeSchoolId ?? "");
const schoolCode = computed(() => auth.activeMembership?.school.code ?? "");
const schoolName = computed(() => auth.activeMembership?.school.name ?? "");
const firstName = computed(
  () => auth.user?.fullName?.split(" ")[0] ?? "Admin",
);

const setupSteps = computed(() => [
  {
    label: "Tahun Ajaran",
    detail: activeYear.value
      ? activeYear.value.academicYearName
      : "Belum ada tahun ajaran aktif",
    done: Boolean(activeYear.value),
    link: "/admin/academic-years",
  },
  {
    label: "Semester",
    detail: activeTerm.value
      ? activeTerm.value.termName
      : "Belum ada semester aktif",
    done: Boolean(activeTerm.value),
    link: "/admin/academic-years",
  },
  {
    label: "Kelas",
    detail:
      (stats.value?.totalClasses ?? 0) > 0
        ? `${stats.value!.totalClasses} kelas dibuat`
        : "Belum ada kelas",
    done: (stats.value?.totalClasses ?? 0) > 0,
    link: "/admin/classes",
  },
  {
    label: "Kelas Aktif",
    detail:
      (stats.value?.activeClasses ?? 0) > 0
        ? `${stats.value!.activeClasses} kelas aktif`
        : "Belum ada kelas aktif",
    done: (stats.value?.activeClasses ?? 0) > 0,
    link: "/admin/classes",
  },
  {
    label: "Siswa Terdaftar",
    detail:
      (stats.value?.totalStudents ?? 0) > 0
        ? `${stats.value!.totalStudents} siswa`
        : "Belum ada siswa",
    done: (stats.value?.totalStudents ?? 0) > 0,
    link: "/admin/enrollments",
  },
  {
    label: "Guru Ditugaskan",
    detail:
      (stats.value?.totalTeachers ?? 0) > 0
        ? `${stats.value!.totalTeachers} guru`
        : "Belum ada guru",
    done: (stats.value?.totalTeachers ?? 0) > 0,
    link: "/admin/users",
  },
]);

const isSetupComplete = computed(() => setupSteps.value.every((s) => s.done));

const statCards = computed(() => [
  {
    label: "Total Siswa",
    value: stats.value?.totalStudents ?? 0,
    icon: PhUsers,
    color: "bg-[#eef2ff] text-[#4f46e5]",
  },
  {
    label: "Total Guru",
    value: stats.value?.totalTeachers ?? 0,
    icon: PhChalkboardTeacher,
    color: "bg-[#ecfdf5] text-[#059669]",
  },
  {
    label: "Total Kelas",
    value: stats.value?.totalClasses ?? 0,
    icon: PhBookOpen,
    color: "bg-[#fff7ed] text-[#ea580c]",
  },
  {
    label: "Kelas Aktif",
    value: stats.value?.activeClasses ?? 0,
    icon: PhClipboardText,
    color: "bg-[#f0fdf4] text-[#16a34a]",
  },
]);

const quickActions = [
  {
    label: "Tambah Siswa",
    description: "Undang atau impor siswa baru",
    icon: PhUsers,
    to: "/admin/users",
    color: "bg-[#eef2ff] text-[#4f46e5]",
    border: "hover:border-[#c7d2fe]",
  },
  {
    label: "Tambah Guru",
    description: "Undang guru ke sekolah",
    icon: PhChalkboardTeacher,
    to: "/admin/users",
    color: "bg-[#ecfdf5] text-[#059669]",
    border: "hover:border-[#bbf7d0]",
  },
  {
    label: "Buat Kelas",
    description: "Tambahkan kelas baru",
    icon: PhBookOpen,
    to: "/admin/classes",
    color: "bg-[#fff7ed] text-[#ea580c]",
    border: "hover:border-[#fed7aa]",
  },
  {
    label: "Penempatan Kelas",
    description: "Masukkan siswa ke kelas",
    icon: PhClipboardText,
    to: "/admin/enrollments",
    color: "bg-[#f3f4f6] text-[#6b7280]",
    border: "hover:border-[#d1d5db]",
  },
  {
    label: "Ruang Mengajar",
    description: "Hubungkan guru dengan kelas",
    icon: PhCalendarBlank,
    to: "/admin/subject-classes",
    color: "bg-[#fff4ee] text-[#c2410c]",
    border: "hover:border-[#fdba74]",
  },
];

async function loadDashboard() {
  if (!schoolId.value || !schoolCode.value) {
    errorMessage.value = "Konteks sekolah belum tersedia.";
    loading.value = false;
    return;
  }

  loading.value = true;
  errorMessage.value = "";

  try {
    const [dashboardData, yearsData] = await Promise.all([
      getAdminDashboard(schoolId.value),
      getAcademicYearsBySchool(schoolCode.value),
    ]);

    stats.value = dashboardData;

    const foundActiveYear = yearsData.data.find((y) => y.isActive) ?? null;
    activeYear.value = foundActiveYear;

    if (foundActiveYear) {
      const terms = await getTermsByAcademicYear(foundActiveYear.academicYearId);
      activeTerm.value = terms.find((t) => t.isActive) ?? null;
    }
  } catch {
    errorMessage.value =
      "Dashboard belum bisa dimuat. Periksa koneksi atau coba lagi nanti.";
  } finally {
    loading.value = false;
  }
}

onMounted(loadDashboard);
</script>

<template>
  <main
    class="grid min-h-screen min-w-0 flex-1 grid-cols-1 overflow-x-hidden bg-[#f8f7f4] xl:grid-cols-[minmax(0,1fr)_320px]"
  >
    <!-- Main content -->
    <section class="min-w-0">
      <!-- SECTION 1 — Greeting header -->
      <header class="border-b border-[#ebe7df] bg-white">
        <div
          class="flex min-w-0 flex-col gap-3 px-5 py-5 sm:px-6 lg:flex-row lg:items-center lg:justify-between lg:px-8"
        >
          <div class="min-w-0">
            <p class="text-xs font-semibold uppercase tracking-[0.06em] text-[#9ca3af]">
              Admin Sekolah
            </p>
            <h1 class="mt-1 text-2xl font-semibold text-[#171322] sm:text-3xl">
              Halo, {{ firstName }}
            </h1>
          </div>
          <div
            class="flex min-w-0 flex-wrap items-center gap-x-4 gap-y-1.5 rounded-lg border border-[#ebe7df] bg-[#f9fafb] px-4 py-2.5 text-sm"
          >
            <span class="font-semibold text-[#171322]">{{ schoolName }}</span>
            <span
              v-if="activeYear"
              class="flex items-center gap-1.5 text-[#6b7280]"
            >
              <span class="text-[#d1d5db]">·</span>
              {{ activeYear.academicYearName }}
            </span>
            <span
              v-if="activeTerm"
              class="flex items-center gap-1.5 text-[#6b7280]"
            >
              <span class="text-[#d1d5db]">·</span>
              {{ activeTerm.termName }}
            </span>
          </div>
        </div>
      </header>

      <div class="space-y-6 px-5 py-6 sm:px-6 lg:px-8">
        <!-- Error state -->
        <section
          v-if="errorMessage && !loading"
          class="rounded-xl border border-[#f1d6d3] bg-white p-5"
        >
          <div class="flex items-start gap-3">
            <PhWarningCircle
              class="mt-0.5 h-5 w-5 shrink-0 text-[#dc2626]"
              weight="duotone"
            />
            <div class="min-w-0">
              <p class="text-sm font-medium text-[#171322]">
                Dashboard tidak dapat dimuat
              </p>
              <p class="mt-1 text-sm leading-6 text-[#7a7385]">
                {{ errorMessage }}
              </p>
              <button
                class="mt-3 rounded-lg bg-[#4f46e5] px-4 py-2 text-sm font-medium text-white transition hover:bg-[#4338ca]"
                type="button"
                @click="loadDashboard"
              >
                Coba lagi
              </button>
            </div>
          </div>
        </section>

        <!-- SECTION 2 — Quick Stats -->
        <section>
          <div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
            <template v-if="loading">
              <div
                v-for="i in 4"
                :key="i"
                class="h-28 animate-pulse rounded-xl border border-[#ebe7df] bg-white"
              />
            </template>
            <template v-else>
              <article
                v-for="card in statCards"
                :key="card.label"
                class="rounded-xl border border-[#ebe7df] bg-white p-4"
              >
                <div
                  class="flex h-10 w-10 items-center justify-center rounded-xl"
                  :class="card.color"
                >
                  <component :is="card.icon" :size="20" weight="duotone" />
                </div>
                <p class="mt-4 text-2xl font-semibold text-[#171322]">
                  {{ card.value }}
                </p>
                <p class="mt-1 text-xs text-[#6b7280]">{{ card.label }}</p>
              </article>
            </template>
          </div>
        </section>

        <!-- SECTION 3 — Setup Progress -->
        <section
          class="rounded-xl border border-[#ebe7df] bg-white p-5"
          :class="{ 'animate-pulse': loading }"
        >
          <div class="mb-4 flex items-center justify-between gap-3">
            <div>
              <h2 class="text-sm font-semibold text-[#171322]">
                Setup Sekolah
              </h2>
              <p class="mt-0.5 text-xs text-[#6b7280]">
                Kelengkapan konfigurasi awal sebelum pembelajaran dimulai.
              </p>
            </div>
            <span
              v-if="!loading && isSetupComplete"
              class="shrink-0 rounded-full bg-[#ecfdf5] px-3 py-1 text-xs font-semibold text-[#059669]"
            >
              Lengkap
            </span>
            <span
              v-else-if="!loading"
              class="shrink-0 rounded-full bg-[#fff7ed] px-3 py-1 text-xs font-semibold text-[#ea580c]"
            >
              {{ setupSteps.filter((s) => s.done).length }}/{{ setupSteps.length }} selesai
            </span>
          </div>

          <div v-if="loading" class="space-y-2">
            <div
              v-for="i in 6"
              :key="i"
              class="h-9 animate-pulse rounded-lg bg-[#f3f4f6]"
            />
          </div>

          <div v-else class="divide-y divide-[#f3f4f6]">
            <RouterLink
              v-for="step in setupSteps"
              :key="step.label"
              :to="step.link"
              class="group flex items-center gap-3 py-2.5 transition hover:bg-transparent first:pt-0 last:pb-0"
            >
              <component
                :is="step.done ? PhCheckCircle : PhCircleDashed"
                :size="18"
                weight="duotone"
                class="shrink-0 transition"
                :class="
                  step.done ? 'text-[#059669]' : 'text-[#d1d5db] group-hover:text-[#9ca3af]'
                "
              />
              <span class="min-w-0 flex-1">
                <span
                  class="block text-sm font-medium"
                  :class="
                    step.done ? 'text-[#171322]' : 'text-[#6b7280]'
                  "
                >
                  {{ step.label }}
                </span>
                <span class="block text-xs text-[#9ca3af]">
                  {{ step.detail }}
                </span>
              </span>
              <PhArrowRight
                :size="14"
                class="shrink-0 text-[#d1d5db] transition group-hover:text-[#4f46e5]"
              />
            </RouterLink>
          </div>

          <div
            v-if="!loading && isSetupComplete"
            class="mt-4 rounded-lg bg-[#ecfdf5] px-4 py-3"
          >
            <p class="text-sm font-medium text-[#059669]">
              Sekolah siap digunakan.
            </p>
            <p class="mt-0.5 text-xs text-[#065f46]">
              Semua konfigurasi dasar sudah lengkap. Guru dan siswa dapat mulai
              menggunakan platform.
            </p>
          </div>
        </section>

        <!-- SECTION 4 — Quick Actions -->
        <section>
          <h2 class="mb-3 text-sm font-semibold text-[#171322]">
            Aksi Cepat
          </h2>
          <div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
            <RouterLink
              v-for="action in quickActions"
              :key="action.label"
              :to="action.to"
              class="group flex flex-col gap-3 rounded-xl border border-[#ebe7df] bg-white p-4 transition hover:-translate-y-0.5 hover:shadow-md"
              :class="action.border"
            >
              <div
                class="flex h-10 w-10 items-center justify-center rounded-xl"
                :class="action.color"
              >
                <component :is="action.icon" :size="20" weight="duotone" />
              </div>
              <div class="min-w-0">
                <p class="text-sm font-semibold text-[#171322]">
                  {{ action.label }}
                </p>
                <p class="mt-1 text-xs leading-4 text-[#6b7280]">
                  {{ action.description }}
                </p>
              </div>
              <PhArrowRight
                :size="14"
                class="mt-auto text-[#d1d5db] transition group-hover:translate-x-0.5 group-hover:text-[#4f46e5]"
              />
            </RouterLink>
          </div>
        </section>

        <!-- SECTION 5 — Recent Activity -->
        <section class="rounded-xl border border-[#ebe7df] bg-white p-5">
          <div class="mb-4 flex items-center gap-2">
            <PhClockCountdown
              :size="17"
              weight="duotone"
              class="text-[#6b7280]"
            />
            <h2 class="text-sm font-semibold text-[#171322]">
              Aktivitas Terbaru
            </h2>
          </div>

          <div v-if="loading" class="space-y-3">
            <div
              v-for="i in 5"
              :key="i"
              class="h-10 animate-pulse rounded-lg bg-[#f3f4f6]"
            />
          </div>

          <div
            v-else-if="
              !stats?.recentActivities?.length
            "
            class="rounded-lg bg-[#fbfaf8] px-4 py-8 text-center"
          >
            <PhClockCountdown
              class="mx-auto h-7 w-7 text-[#d1d5db]"
              weight="duotone"
            />
            <p class="mt-3 text-sm font-medium text-[#171322]">
              Belum ada aktivitas
            </p>
            <p class="mt-1 text-xs leading-5 text-[#6b7280]">
              Aktivitas admin dan guru di sekolah ini akan muncul di sini.
            </p>
          </div>

          <div v-else class="divide-y divide-[#f3f4f6]">
            <div
              v-for="(activity, index) in stats.recentActivities"
              :key="index"
              class="flex items-start gap-3 py-3 first:pt-0 last:pb-0"
            >
              <div
                class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-[#f3f4f6] text-[10px] font-semibold text-[#6b7280]"
              >
                {{ activity.userName?.charAt(0)?.toUpperCase() ?? "?" }}
              </div>
              <div class="min-w-0 flex-1">
                <p class="text-sm leading-5 text-[#171322]">
                  <span class="font-medium">{{ activity.userName }}</span>
                  {{ " " }}
                  <span class="text-[#6b7280]">{{ activity.action }}</span>
                </p>
                <p class="mt-0.5 text-xs text-[#9ca3af]">
                  {{ formatDateTime(activity.timestamp) }}
                </p>
              </div>
            </div>
          </div>
        </section>
      </div>
    </section>

    <!-- Right sidebar -->
    <aside
      class="min-w-0 border-t border-[#ebe7df] bg-white xl:sticky xl:top-0 xl:h-dvh xl:min-h-0 xl:overflow-y-auto xl:border-l xl:border-t-0"
    >
      <div class="flex flex-col gap-5 p-5">
        <!-- Chat -->
        <LatestChatCard to="/admin/chat" :limit="5" />

        <!-- Enrollment distribution -->
        <section class="rounded-xl border border-[#ebe7df] bg-white p-4">
          <h3 class="mb-3 text-sm font-semibold text-[#171322]">
            Distribusi Kelas
          </h3>

          <div v-if="loading" class="space-y-2">
            <div
              v-for="i in 4"
              :key="i"
              class="h-8 animate-pulse rounded-lg bg-[#f3f4f6]"
            />
          </div>

          <div
            v-else-if="!stats?.enrollmentTrends?.length"
            class="rounded-lg bg-[#fbfaf8] px-3 py-5 text-center"
          >
            <p class="text-xs text-[#6b7280]">Belum ada data distribusi kelas.</p>
          </div>

          <div v-else class="space-y-2.5">
            <div
              v-for="trend in stats.enrollmentTrends.slice(0, 6)"
              :key="trend.className"
              class="min-w-0"
            >
              <div class="flex items-center justify-between gap-2 text-xs">
                <span class="truncate font-medium text-[#374151]">
                  {{ trend.className }}
                </span>
                <span class="shrink-0 text-[#6b7280]">
                  {{ trend.students }} siswa
                </span>
              </div>
              <div class="mt-1.5 h-1.5 w-full overflow-hidden rounded-full bg-[#f3f4f6]">
                <div
                  class="h-full rounded-full bg-[#4f46e5] transition-all"
                  :style="{
                    width: `${Math.min(100, (trend.students / Math.max(1, ...stats!.enrollmentTrends.map((t) => t.students))) * 100)}%`,
                  }"
                />
              </div>
            </div>
          </div>
        </section>
      </div>
    </aside>
  </main>
</template>
