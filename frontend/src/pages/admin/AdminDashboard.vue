<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import {
  PhArrowRight,
  PhBookOpen,
  PhChalkboardTeacher,
  PhCheckCircle,
  PhClipboardText,
  PhClockCountdown,
  PhEnvelopeSimple,
  PhScales,
  PhUsers,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import { getAdminDashboard } from "../../services/adminDashboard";
import {
  getAcademicYearsBySchool,
  getTermsByAcademicYear,
} from "../../services/adminAcademic";
import { listSchoolMemberInvitations } from "../../services/adminSchoolMemberInvitation";
import type { AdminDashboardSummary } from "../../types/adminDashboard";
import type { AcademicYearItem, TermItem } from "../../types/adminAcademic";
import type { SchoolMemberInvitationItem } from "../../types/adminSchoolMemberInvitation";
import { formatDate, formatDateTime } from "../../utils/date";
import LatestChatCard from "../../components/chat/LatestChatCard.vue";
import ContextSwitcher from "../../components/layout/ContextSwitcher.vue";

const auth = useAuthStore();

const loading = ref(true);
const errorMessage = ref("");
const stats = ref<AdminDashboardSummary | null>(null);
const activeYear = ref<AcademicYearItem | null>(null);
const activeTerm = ref<TermItem | null>(null);

const pendingInvitations = ref<SchoolMemberInvitationItem[]>([]);
const invitationsLoading = ref(true);
const invitationsError = ref("");

const schoolId = computed(() => auth.activeSchoolId ?? "");
const schoolCode = computed(() => auth.activeMembership?.school.code ?? "");
const firstName = computed(() => auth.user?.fullName?.split(" ")[0] ?? "Admin");

const academicSetupIncomplete = computed(
  () =>
    !loading.value &&
    !errorMessage.value &&
    (!activeYear.value || !activeTerm.value),
);

const sortedPendingInvitations = computed(() =>
  [...pendingInvitations.value].sort(
    (a, b) => new Date(a.expiresAt).getTime() - new Date(b.expiresAt).getTime(),
  ),
);

const invitationsExpiringSoon = computed(() =>
  sortedPendingInvitations.value.filter(
    (invitation) =>
      new Date(invitation.expiresAt).getTime() - Date.now() <
      48 * 60 * 60 * 1000,
  ),
);

const classesWithoutTeacher = computed(
  () => stats.value?.classesWithoutTeacher ?? [],
);
const classesWithoutTeacherTotal = computed(
  () => stats.value?.classesWithoutTeacherTotal ?? 0,
);

const contentLessSubjectClasses = computed(
  () => stats.value?.contentLessSubjectClasses ?? [],
);
const contentLessSubjectClassesTotal = computed(
  () => stats.value?.contentLessSubjectClassesTotal ?? 0,
);

const subjectsWithoutAssessmentWeight = computed(
  () => stats.value?.subjectsWithoutAssessmentWeight ?? [],
);
const subjectsWithoutAssessmentWeightTotal = computed(
  () => stats.value?.subjectsWithoutAssessmentWeightTotal ?? 0,
);

const needsAttentionLoading = computed(
  () => loading.value || invitationsLoading.value,
);
const needsAttentionPartialError = computed(
  () => Boolean(errorMessage.value || invitationsError.value),
);
const hasAttentionItems = computed(
  () =>
    academicSetupIncomplete.value ||
    pendingInvitations.value.length > 0 ||
    classesWithoutTeacherTotal.value > 0,
);

function invitationRoleLabel(role: string) {
  if (role === "student") return "Siswa";
  if (role === "teacher") return "Guru";
  return role;
}

const statCards = computed(() => [
  {
    label: "Total Siswa",
    value: stats.value?.totalStudents ?? 0,
    icon: PhUsers,
    color: "bg-brand-soft text-brand",
    to: "/admin/users",
  },
  {
    label: "Total Guru",
    value: stats.value?.totalTeachers ?? 0,
    icon: PhChalkboardTeacher,
    color: "bg-[#f0fdf4] text-[#059669]",
    to: "/admin/users",
  },
  {
    label: "Total Kelas",
    value: stats.value?.totalClasses ?? 0,
    icon: PhBookOpen,
    color: "bg-warning-soft text-[#ea580c]",
    to: "/admin/classes",
  },
  {
    label: "Kelas Aktif",
    value: stats.value?.activeClasses ?? 0,
    icon: PhClipboardText,
    color: "bg-[#f0fdf4] text-[#16a34a]",
    to: "/admin/classes",
  },
]);

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

    // Defensive: never assume a collection response is an array. A new
    // school has zero academic years, and a bare/never-appended-to Go slice
    // serializes as JSON null rather than [] — never trust the network
    // response shape alone.
    const years = yearsData.data ?? [];
    const foundActiveYear = years.find((y) => y.isActive) ?? null;
    activeYear.value = foundActiveYear;

    if (foundActiveYear) {
      const terms = await getTermsByAcademicYear(
        foundActiveYear.academicYearId,
      );
      const termsList = terms ?? [];
      activeTerm.value = termsList.find((t) => t.isActive) ?? null;
    }
  } catch {
    errorMessage.value =
      "Dashboard belum bisa dimuat. Periksa koneksi atau coba lagi nanti.";
  } finally {
    loading.value = false;
  }
}

async function loadInvitations() {
  if (!schoolId.value) {
    invitationsLoading.value = false;
    return;
  }

  invitationsLoading.value = true;
  invitationsError.value = "";

  try {
    const response = await listSchoolMemberInvitations({ limit: 20 });
    pendingInvitations.value = response.data ?? [];
  } catch {
    pendingInvitations.value = [];
    invitationsError.value = "Undangan belum bisa dimuat.";
  } finally {
    invitationsLoading.value = false;
  }
}

onMounted(() => {
  loadDashboard();
  loadInvitations();
});
</script>

<template>
  <main
    class="grid min-h-screen min-w-0 flex-1 grid-cols-1 overflow-x-hidden bg-background xl:grid-cols-[minmax(0,1fr)_320px]"
  >
    <!-- Main content -->
    <section class="min-w-0">
      <!-- SECTION 1 — Greeting header -->
      <header class="border-b border-border bg-surface">
        <div
          class="flex min-w-0 flex-col gap-3 px-5 py-5 sm:px-6 lg:flex-row lg:items-center lg:justify-between lg:px-8"
        >
          <div class="min-w-0">
            <h1 class="mt-1 text-2xl font-semibold text-foreground sm:text-3xl">
              Halo, {{ firstName }}
            </h1>
          </div>
          <ContextSwitcher />
        </div>
      </header>

      <div class="space-y-6 px-5 py-6 sm:px-6 lg:px-8">
        <!-- Error state -->
        <section
          v-if="errorMessage && !loading"
          class="rounded-xl border border-danger-line bg-danger-soft p-5"
        >
          <div class="flex items-start gap-3">
            <PhWarningCircle
              class="mt-0.5 h-5 w-5 shrink-0 text-danger"
              weight="duotone"
            />
            <div class="min-w-0">
              <p class="text-sm font-medium text-foreground">
                Dashboard tidak dapat dimuat
              </p>
              <p class="mt-1 text-sm leading-6 text-muted">
                {{ errorMessage }}
              </p>
              <button
                class="mt-3 rounded-lg bg-brand px-4 py-2 text-sm font-medium text-white transition hover:bg-brand-hover"
                type="button"
                @click="loadDashboard"
              >
                Coba lagi
              </button>
            </div>
          </div>
        </section>

        <!-- SECTION 1B — Needs Attention -->
        <section class="space-y-2.5">
          <h2 class="text-sm font-semibold text-foreground">Perlu Perhatian</h2>

          <div
            v-if="needsAttentionLoading"
            class="rounded-xl border border-border bg-surface shadow-sm p-4"
          >
            <div class="h-14 animate-pulse rounded-lg bg-surface-strong" />
          </div>

          <template v-else>
          <RouterLink
            v-if="academicSetupIncomplete"
            to="/admin/academic-years"
            class="flex items-start gap-3 rounded-xl border border-warning-line bg-warning-soft p-4 transition hover:brightness-95"
          >
            <PhWarningCircle
              :size="20"
              weight="duotone"
              class="mt-0.5 shrink-0 text-[#ea580c]"
            />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-foreground">
                Tahun ajaran atau semester aktif belum diatur
              </p>
              <p class="mt-0.5 text-xs text-muted">
                Sebagian fitur akademik memerlukan tahun ajaran dan semester
                aktif. Atur sekarang.
              </p>
            </div>
            <PhArrowRight :size="14" class="mt-1 shrink-0 text-[#ea580c]" />
          </RouterLink>

          <RouterLink
            v-if="pendingInvitations.length > 0"
            to="/admin/users"
            class="flex items-start gap-3 rounded-xl border p-4 transition hover:brightness-95"
            :class="
              invitationsExpiringSoon.length > 0
                ? 'border-danger-line bg-danger-soft'
                : 'border-warning-line bg-warning-soft'
            "
          >
            <PhEnvelopeSimple
              :size="20"
              weight="duotone"
              class="mt-0.5 shrink-0"
              :class="
                invitationsExpiringSoon.length > 0
                  ? 'text-danger'
                  : 'text-[#ea580c]'
              "
            />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-foreground">
                {{ pendingInvitations.length }} undangan menunggu diterima
              </p>
              <p class="mt-0.5 text-xs text-muted">
                <template v-if="invitationsExpiringSoon.length > 0">
                  Termasuk {{ invitationsExpiringSoon.length }} undangan yang
                  akan kedaluwarsa dalam 48 jam.
                </template>
                <template v-else>
                  Undangan akan kedaluwarsa otomatis jika tidak diterima.
                </template>
              </p>
            </div>
            <PhArrowRight
              :size="14"
              class="mt-1 shrink-0"
              :class="
                invitationsExpiringSoon.length > 0
                  ? 'text-danger'
                  : 'text-[#ea580c]'
              "
            />
          </RouterLink>

          <RouterLink
            v-if="classesWithoutTeacherTotal > 0"
            to="/admin/classes"
            class="flex items-start gap-3 rounded-xl border border-warning-line bg-warning-soft p-4 transition hover:brightness-95"
          >
            <PhChalkboardTeacher
              :size="20"
              weight="duotone"
              class="mt-0.5 shrink-0 text-[#ea580c]"
            />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-foreground">
                {{ classesWithoutTeacherTotal }} kelas belum memiliki guru
              </p>
              <p class="mt-0.5 text-xs text-muted">
                Kelas ini belum bisa memulai pembelajaran sampai guru
                ditugaskan.
              </p>
            </div>
            <PhArrowRight :size="14" class="mt-1 shrink-0 text-[#ea580c]" />
          </RouterLink>

          <div
            v-if="!hasAttentionItems && !needsAttentionPartialError"
            class="flex items-center gap-3 rounded-xl border border-success-line bg-success-soft p-4"
          >
            <PhCheckCircle :size="20" weight="duotone" class="shrink-0 text-success" />
            <p class="text-sm font-medium text-foreground">
              Semua beres. Tidak ada yang perlu ditindaklanjuti saat ini.
            </p>
          </div>

          <div
            v-else-if="!hasAttentionItems && needsAttentionPartialError"
            class="rounded-xl border border-border bg-surface-subtle p-4 text-sm leading-6 text-muted"
          >
            Sebagian data belum bisa dimuat, jadi status di atas mungkin belum
            lengkap.
          </div>
          </template>
        </section>

        <!-- SECTION 1C — Work Queue: Pending Invitations -->
        <section
          class="rounded-xl border border-border bg-surface shadow-sm p-5"
        >
          <div class="mb-4 flex items-center justify-between gap-3">
            <div>
              <h2 class="text-sm font-semibold text-foreground">
                Undangan Tertunda
              </h2>
              <p class="mt-0.5 text-xs text-muted">
                Anggota yang diundang tapi belum menerima undangan.
              </p>
            </div>
            <RouterLink
              v-if="pendingInvitations.length > 0"
              to="/admin/users"
              class="shrink-0 text-xs font-medium text-brand transition hover:text-brand-hover"
            >
              Lihat semua
            </RouterLink>
          </div>

          <div v-if="invitationsLoading" class="space-y-2">
            <div
              v-for="i in 3"
              :key="i"
              class="h-14 animate-pulse rounded-lg bg-surface-strong"
            />
          </div>

          <div
            v-else-if="invitationsError"
            class="rounded-lg bg-surface-subtle p-4 text-sm leading-6 text-muted"
          >
            {{ invitationsError }}
          </div>

          <div
            v-else-if="pendingInvitations.length === 0"
            class="rounded-lg bg-surface-subtle px-4 py-8 text-center"
          >
            <PhEnvelopeSimple
              class="mx-auto h-7 w-7 text-border-strong"
              weight="duotone"
            />
            <p class="mt-3 text-sm font-medium text-foreground">
              Tidak ada undangan yang tertunda.
            </p>
          </div>

          <div v-else class="divide-y divide-[#f3f4f6]">
            <div
              v-for="invitation in sortedPendingInvitations.slice(0, 5)"
              :key="invitation.invitationId"
              class="flex items-center justify-between gap-3 py-2.5 first:pt-0 last:pb-0"
            >
              <div class="min-w-0 flex-1">
                <p class="truncate text-sm font-medium text-foreground">
                  {{ invitation.fullName }}
                </p>
                <p class="mt-0.5 truncate text-xs text-muted">
                  {{ invitationRoleLabel(invitation.role) }}
                  <span v-if="invitation.class">
                    · {{ invitation.class.classTitle }}</span
                  >
                </p>
              </div>
              <span class="shrink-0 text-xs text-muted">
                Kedaluwarsa {{ formatDate(invitation.expiresAt) }}
              </span>
            </div>
          </div>
        </section>

        <!-- SECTION 1D — Work Queue: Classes without teacher -->
        <section
          class="rounded-xl border border-border bg-surface shadow-sm p-5"
        >
          <div class="mb-4 flex items-center justify-between gap-3">
            <div>
              <h2 class="text-sm font-semibold text-foreground">
                Kelas Tanpa Guru
              </h2>
              <p class="mt-0.5 text-xs text-muted">
                Kelas aktif yang belum memiliki mata pelajaran dengan guru
                ditugaskan.
              </p>
            </div>
            <RouterLink
              v-if="classesWithoutTeacherTotal > 0"
              to="/admin/classes"
              class="shrink-0 text-xs font-medium text-brand transition hover:text-brand-hover"
            >
              Lihat semua
            </RouterLink>
          </div>

          <div v-if="loading" class="space-y-2">
            <div
              v-for="i in 3"
              :key="i"
              class="h-14 animate-pulse rounded-lg bg-surface-strong"
            />
          </div>

          <div
            v-else-if="errorMessage"
            class="rounded-lg bg-surface-subtle p-4 text-sm leading-6 text-muted"
          >
            Kelas tanpa guru belum bisa dimuat.
          </div>

          <div
            v-else-if="classesWithoutTeacher.length === 0"
            class="rounded-lg bg-surface-subtle px-4 py-8 text-center"
          >
            <PhChalkboardTeacher
              class="mx-auto h-7 w-7 text-border-strong"
              weight="duotone"
            />
            <p class="mt-3 text-sm font-medium text-foreground">
              Semua kelas sudah memiliki guru.
            </p>
          </div>

          <div v-else class="divide-y divide-[#f3f4f6]">
            <RouterLink
              v-for="klass in classesWithoutTeacher"
              :key="klass.classId"
              to="/admin/classes"
              class="flex items-center justify-between gap-3 py-2.5 first:pt-0 last:pb-0"
            >
              <p
                class="min-w-0 flex-1 truncate text-sm font-medium text-foreground"
              >
                {{ klass.className }}
              </p>
              <PhArrowRight :size="14" class="shrink-0 text-border-strong" />
            </RouterLink>
          </div>
        </section>

        <!-- SECTION 1E — Work Queue: Subject-classes without content -->
        <section class="rounded-xl border border-border bg-surface shadow-sm p-5">
          <div class="mb-4 flex items-center justify-between gap-3">
            <div>
              <h2 class="text-sm font-semibold text-foreground">
                Subject-Class Tanpa Konten
              </h2>
              <p class="mt-0.5 text-xs text-muted">
                Kelas dan mata pelajaran yang belum memiliki materi atau tugas.
              </p>
            </div>
            <RouterLink
              v-if="contentLessSubjectClassesTotal > 0"
              to="/admin/subject-classes"
              class="shrink-0 text-xs font-medium text-brand transition hover:text-brand-hover"
            >
              Lihat semua
            </RouterLink>
          </div>

          <div v-if="loading" class="space-y-2">
            <div
              v-for="i in 3"
              :key="i"
              class="h-14 animate-pulse rounded-lg bg-surface-strong"
            />
          </div>

          <div
            v-else-if="errorMessage"
            class="rounded-lg bg-surface-subtle p-4 text-sm leading-6 text-muted"
          >
            Subject-class tanpa konten belum bisa dimuat.
          </div>

          <div
            v-else-if="contentLessSubjectClasses.length === 0"
            class="rounded-lg bg-surface-subtle px-4 py-8 text-center"
          >
            <PhBookOpen
              class="mx-auto h-7 w-7 text-border-strong"
              weight="duotone"
            />
            <p class="mt-3 text-sm font-medium text-foreground">
              Semua subject-class sudah memiliki konten.
            </p>
          </div>

          <div v-else class="divide-y divide-[#f3f4f6]">
            <RouterLink
              v-for="item in contentLessSubjectClasses"
              :key="item.subjectClassId"
              to="/admin/subject-classes"
              class="flex items-center justify-between gap-3 py-2.5 first:pt-0 last:pb-0"
            >
              <div class="min-w-0 flex-1">
                <p class="truncate text-sm font-medium text-foreground">
                  {{ item.subjectName }}
                </p>
                <p class="mt-0.5 truncate text-xs text-muted">
                  {{ item.className }}
                </p>
              </div>
              <PhArrowRight :size="14" class="shrink-0 text-border-strong" />
            </RouterLink>
          </div>
        </section>

        <!-- SECTION 1F — Work Queue: Subjects without assessment weight -->
        <section class="rounded-xl border border-border bg-surface shadow-sm p-5">
          <div class="mb-4 flex items-center justify-between gap-3">
            <div>
              <h2 class="text-sm font-semibold text-foreground">
                Mata Pelajaran Belum Dikonfigurasi
              </h2>
              <p class="mt-0.5 text-xs text-muted">
                Mata pelajaran yang belum memiliki bobot penilaian.
              </p>
            </div>
            <RouterLink
              v-if="subjectsWithoutAssessmentWeightTotal > 0"
              to="/admin/subject-classes"
              class="shrink-0 text-xs font-medium text-brand transition hover:text-brand-hover"
            >
              Lihat semua
            </RouterLink>
          </div>

          <div v-if="loading" class="space-y-2">
            <div
              v-for="i in 3"
              :key="i"
              class="h-14 animate-pulse rounded-lg bg-surface-strong"
            />
          </div>

          <div
            v-else-if="errorMessage"
            class="rounded-lg bg-surface-subtle p-4 text-sm leading-6 text-muted"
          >
            Mata pelajaran belum dikonfigurasi belum bisa dimuat.
          </div>

          <div
            v-else-if="subjectsWithoutAssessmentWeight.length === 0"
            class="rounded-lg bg-surface-subtle px-4 py-8 text-center"
          >
            <PhScales
              class="mx-auto h-7 w-7 text-border-strong"
              weight="duotone"
            />
            <p class="mt-3 text-sm font-medium text-foreground">
              Semua mata pelajaran sudah dikonfigurasi.
            </p>
          </div>

          <div v-else class="divide-y divide-[#f3f4f6]">
            <RouterLink
              v-for="subject in subjectsWithoutAssessmentWeight"
              :key="subject.subjectId"
              to="/admin/subject-classes"
              class="flex items-center justify-between gap-3 py-2.5 first:pt-0 last:pb-0"
            >
              <p
                class="min-w-0 flex-1 truncate text-sm font-medium text-foreground"
              >
                {{ subject.subjectName }}
              </p>
              <PhArrowRight :size="14" class="shrink-0 text-border-strong" />
            </RouterLink>
          </div>
        </section>

        <!-- SECTION 2 — Quick Stats -->
        <section>
          <div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
            <template v-if="loading">
              <div
                v-for="i in 4"
                :key="i"
                class="h-28 animate-pulse rounded-xl border border-border bg-surface shadow-sm"
              />
            </template>
            <div
              v-else-if="errorMessage"
              class="col-span-2 rounded-xl border border-border bg-surface shadow-sm p-4 text-center sm:col-span-4"
            >
              <p class="text-sm text-muted">Statistik belum bisa dimuat.</p>
            </div>
            <template v-else>
              <RouterLink
                v-for="card in statCards"
                :key="card.label"
                :to="card.to"
                class="rounded-xl border border-border bg-surface shadow-sm p-4 transition hover:-translate-y-0.5 hover:shadow-md"
              >
                <div
                  class="flex h-10 w-10 items-center justify-center rounded-xl"
                  :class="card.color"
                >
                  <component :is="card.icon" :size="20" weight="duotone" />
                </div>
                <p class="mt-4 text-2xl font-semibold text-foreground">
                  {{ card.value }}
                </p>
                <p class="mt-1 text-xs text-muted">{{ card.label }}</p>
              </RouterLink>
            </template>
          </div>
        </section>

        <!-- SECTION 5 — Recent Activity -->
        <section
          class="rounded-xl border border-border bg-surface shadow-sm p-5"
        >
          <div class="mb-4 flex items-center gap-2">
            <PhClockCountdown :size="17" weight="duotone" class="text-muted" />
            <h2 class="text-sm font-semibold text-foreground">
              Aktivitas Terbaru
            </h2>
          </div>

          <div v-if="loading" class="space-y-3">
            <div
              v-for="i in 5"
              :key="i"
              class="h-10 animate-pulse rounded-lg bg-surface-strong"
            />
          </div>

          <div
            v-else-if="errorMessage"
            class="rounded-lg bg-surface-subtle p-4 text-sm leading-6 text-muted"
          >
            Aktivitas terbaru belum bisa dimuat.
          </div>

          <div
            v-else-if="!stats?.recentActivities?.length"
            class="rounded-lg bg-surface-subtle px-4 py-8 text-center"
          >
            <PhClockCountdown
              class="mx-auto h-7 w-7 text-border-strong"
              weight="duotone"
            />
            <p class="mt-3 text-sm font-medium text-foreground">
              Belum ada aktivitas
            </p>
            <p class="mt-1 text-xs leading-5 text-muted">
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
                class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-surface-strong text-[10px] font-semibold text-muted"
              >
                {{ activity.userName?.charAt(0)?.toUpperCase() ?? "?" }}
              </div>
              <div class="min-w-0 flex-1">
                <p class="text-sm leading-5 text-foreground">
                  <span class="font-medium">{{ activity.userName }}</span>
                  {{ " " }}
                  <span class="text-muted">{{ activity.action }}</span>
                </p>
                <p class="mt-0.5 text-xs text-muted">
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
      class="min-w-0 border-t border-border bg-surface xl:sticky xl:top-0 xl:h-dvh xl:min-h-0 xl:overflow-y-auto xl:border-l xl:border-t-0"
    >
      <div class="flex flex-col gap-5 p-5">
        <!-- Chat -->
        <LatestChatCard to="/admin/chat" :limit="5" />

        <!-- Enrollment distribution -->
        <section
          class="rounded-xl border border-border bg-surface shadow-sm p-4"
        >
          <h3 class="mb-3 text-sm font-semibold text-foreground">
            Distribusi Kelas
          </h3>

          <div v-if="loading" class="space-y-2">
            <div
              v-for="i in 4"
              :key="i"
              class="h-8 animate-pulse rounded-lg bg-surface-strong"
            />
          </div>

          <div
            v-else-if="errorMessage"
            class="rounded-lg bg-surface-subtle px-3 py-5 text-center"
          >
            <p class="text-xs text-muted">Distribusi kelas belum bisa dimuat.</p>
          </div>

          <div
            v-else-if="!stats?.enrollmentTrends?.length"
            class="rounded-lg bg-surface-subtle px-3 py-5 text-center"
          >
            <p class="text-xs text-muted">Belum ada data distribusi kelas.</p>
          </div>

          <div v-else class="space-y-2.5">
            <div
              v-for="trend in stats.enrollmentTrends.slice(0, 6)"
              :key="trend.className"
              class="min-w-0"
            >
              <div class="flex items-center justify-between gap-2 text-xs">
                <span class="truncate font-medium text-foreground-secondary">
                  {{ trend.className }}
                </span>
                <span class="shrink-0 text-muted">
                  {{ trend.students }} siswa
                </span>
              </div>
              <div
                class="mt-1.5 h-1.5 w-full overflow-hidden rounded-full bg-surface-strong"
              >
                <div
                  class="h-full rounded-full bg-brand transition-all"
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
