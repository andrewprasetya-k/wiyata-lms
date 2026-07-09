<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import {
  PhArrowRight,
  PhBookOpen,
  PhCalendarBlank,
  PhChartLineUp,
  PhChalkboardTeacher,
  PhClipboardText,
  PhMegaphone,
  PhUsers,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import { getTeacherDashboard } from "../../services/teacherDashboard";
import { getAcademicActivities } from "../../services/activity";
import type { MembershipInfo } from "../../types/auth";
import type { TeacherDashboardSummary } from "../../types/teacherDashboard";
import type { AcademicActivityItem } from "../../types/activity";
import { resolveSubjectColor } from "../../utils/color";
import LatestChatCard from "../../components/chat/LatestChatCard.vue";
import AcademicActivityCard from "../../components/activity/AcademicActivityCard.vue";
import ContextSwitcher from "../../components/layout/ContextSwitcher.vue";

const auth = useAuthStore();

const loading = ref(false);
const errorMessage = ref("");
const summary = ref<TeacherDashboardSummary | null>(null);
const activities = ref<AcademicActivityItem[]>([]);
const activitiesLoading = ref(false);
const activitiesError = ref("");

const activeMembership = computed<MembershipInfo | undefined>(() => {
  return (
    auth.memberships.find((m) => m.school.id === auth.activeSchoolId) ??
    auth.memberships.find((m) => m.isDefault) ??
    auth.memberships[0]
  );
});

const schoolUserId = computed(
  () =>
    activeMembership.value?.schoolUserId ?? auth.defaultContext?.schoolUserId ?? "",
);

const teacherName = computed(() => auth.user?.fullName ?? "Guru");
const firstName = computed(() => teacherName.value.split(" ")[0]);

const pendingReviews = computed(() => summary.value?.pendingReviews ?? 0);
const hasPendingReviews = computed(() => pendingReviews.value > 0);

function formatPercentage(value: number | null | undefined) {
  if (typeof value !== "number" || !Number.isFinite(value)) return "0%";
  return `${Math.min(100, Math.max(0, value)).toFixed(1)}%`;
}

const statCards = computed(() => [
  {
    label: "Menunggu Penilaian",
    value: loading.value ? "..." : String(pendingReviews.value),
    helper: "Pengumpulan yang belum dinilai",
    icon: PhClipboardText,
    urgent: hasPendingReviews.value,
    to: "/teacher/submissions",
    colorIcon: hasPendingReviews.value
      ? "bg-[#fff7ed] text-[#ea580c]"
      : "bg-[#f0fdf4] text-[#059669]",
    colorValue: hasPendingReviews.value ? "text-[#ea580c]" : "text-[#171322]",
    border: hasPendingReviews.value
      ? "border-[#fed7aa] hover:border-[#fb923c]"
      : "border-[#ebe7df] hover:border-[#bbf7d0]",
  },
  {
    label: "Total Siswa",
    value: loading.value ? "..." : String(summary.value?.totalStudents ?? 0),
    helper: "Siswa dari semua kelas yang diajar",
    icon: PhUsers,
    urgent: false,
    to: null,
    colorIcon: "bg-[#eef2ff] text-[#4f46e5]",
    colorValue: "text-[#171322]",
    border: "border-[#ebe7df]",
  },
  {
    label: "Pengumpulan Tugas",
    value: loading.value
      ? "..."
      : formatPercentage(summary.value?.submissionRate),
    helper: "Rata-rata pengumpulan dibanding siswa aktif",
    icon: PhChartLineUp,
    urgent: false,
    to: null,
    colorIcon: "bg-[#f0fdf4] text-[#16a34a]",
    colorValue: "text-[#171322]",
    border: "border-[#ebe7df]",
  },
]);

const quickActions = [
  {
    label: "Nilai Pengumpulan",
    icon: PhClipboardText,
    to: "/teacher/submissions",
    color: "bg-[#fff7ed] text-[#ea580c]",
    border: "hover:border-[#fed7aa]",
  },
  {
    label: "Buat Tugas",
    icon: PhCalendarBlank,
    to: "/teacher/assignments",
    color: "bg-[#f0fdf4] text-[#059669]",
    border: "hover:border-[#bbf7d0]",
  },
  {
    label: "Feed Kelas",
    icon: PhMegaphone,
    to: "/teacher/feed",
    color: "bg-[#eef2ff] text-[#4f46e5]",
    border: "hover:border-[#c7d2fe]",
  },
  {
    label: "Ruang Mengajar",
    icon: PhChalkboardTeacher,
    to: "/teacher/subjects",
    color: "bg-[#f3f1ec] text-[#6b7280]",
    border: "hover:border-[#d1d5db]",
  },
];

async function loadDashboard() {
  if (!schoolUserId.value) {
    summary.value = null;
    errorMessage.value = "";
    return;
  }

  loading.value = true;
  errorMessage.value = "";
  try {
    summary.value = await getTeacherDashboard(schoolUserId.value);
  } catch {
    errorMessage.value =
      "Dashboard guru belum bisa dimuat. Coba lagi beberapa saat.";
  } finally {
    loading.value = false;
  }
}

async function loadActivities() {
  activitiesLoading.value = true;
  activitiesError.value = "";
  try {
    const response = await getAcademicActivities();
    activities.value = response.items ?? [];
  } catch {
    activities.value = [];
    activitiesError.value = "Aktivitas mengajar belum bisa dimuat.";
  } finally {
    activitiesLoading.value = false;
  }
}

onMounted(() => {
  loadDashboard();
  loadActivities();
});
</script>

<template>
  <main
    class="grid min-h-screen min-w-0 flex-1 grid-cols-1 overflow-x-hidden bg-[#f8f7f4] xl:grid-cols-[minmax(0,1fr)_300px]"
  >
    <!-- Main content -->
    <section class="min-w-0">
      <!-- Header -->
      <header class="border-b border-[#ebe7df] bg-white">
        <div
          class="flex min-w-0 flex-col gap-3 px-5 py-5 sm:px-6 lg:flex-row lg:items-center lg:justify-between lg:px-8"
        >
          <div class="min-w-0">
            <p
              class="text-xs font-semibold uppercase tracking-[0.06em] text-[#9ca3af]"
            >
              Guru
            </p>
            <h1 class="mt-1 text-2xl font-semibold text-[#171322] sm:text-3xl">
              Selamat mengajar, {{ firstName }}
            </h1>
          </div>
          <ContextSwitcher />
        </div>
      </header>

      <div class="space-y-5 px-5 py-5 sm:px-6 lg:px-8 lg:py-6">
        <!-- No school context -->
        <section
          v-if="!schoolUserId"
          class="flex min-h-[55vh] items-center justify-center"
        >
          <article
            class="w-full max-w-xl rounded-xl border border-[#ebe7df] bg-white shadow-sm p-8 text-center"
          >
            <div
              class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-[#fff7ed] text-[#ea580c]"
            >
              <PhWarningCircle :size="24" weight="duotone" />
            </div>
            <h2 class="mt-3 text-lg font-medium text-[#171322]">
              Konteks guru belum tersedia
            </h2>
            <p class="mx-auto mt-2 max-w-md text-sm leading-6 text-[#7a7385]">
              Pastikan akun guru memiliki akses aktif pada sekolah yang sedang
              digunakan.
            </p>
          </article>
        </section>

        <template v-else>
          <!-- Stat cards -->
          <section class="grid gap-3 sm:grid-cols-3">
            <template v-if="loading">
              <div
                v-for="i in 3"
                :key="i"
                class="h-28 animate-pulse rounded-xl border border-[#ebe7df] bg-white shadow-sm"
              />
            </template>
            <template v-else>
              <component
                :is="card.to ? RouterLink : 'article'"
                v-for="card in statCards"
                :key="card.label"
                v-bind="card.to ? { to: card.to } : {}"
                class="group rounded-xl border bg-white shadow-sm p-4 transition"
                :class="[
                  card.border,
                  card.to ? 'hover:-translate-y-0.5 hover:shadow-md' : '',
                ]"
              >
                <div class="flex items-start justify-between gap-3">
                  <div
                    class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl"
                    :class="card.colorIcon"
                  >
                    <component :is="card.icon" :size="20" weight="duotone" />
                  </div>
                  <PhArrowRight
                    v-if="card.to"
                    :size="15"
                    class="mt-1 shrink-0 text-[#d1d5db] transition group-hover:translate-x-0.5 group-hover:text-[#4f46e5]"
                  />
                </div>
                <p
                  class="mt-3 text-2xl font-semibold"
                  :class="card.colorValue"
                >
                  {{ card.value }}
                </p>
                <p class="mt-0.5 text-sm font-medium text-[#171322]">
                  {{ card.label }}
                </p>
                <p class="mt-1 text-xs leading-5 text-[#8a8494]">
                  {{ card.helper }}
                </p>
              </component>
            </template>
          </section>

          <!-- Urgent pending reviews banner -->
          <RouterLink
            v-if="!loading && hasPendingReviews"
            to="/teacher/submissions"
            class="group flex items-center gap-4 rounded-xl border border-[#fed7aa] bg-[#fff7ed] px-5 py-4 transition hover:border-[#fb923c] hover:bg-[#ffedd5]"
          >
            <div
              class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-[#ea580c] text-white"
            >
              <PhClipboardText :size="20" weight="duotone" />
            </div>
            <div class="min-w-0 flex-1">
              <p class="text-sm font-semibold text-[#9a3412]">
                {{ pendingReviews }} pengumpulan menunggu penilaian
              </p>
              <p class="mt-0.5 text-xs text-[#c2410c]">
                Nilai sekarang agar siswa mendapat feedback lebih cepat.
              </p>
            </div>
            <PhArrowRight
              :size="16"
              class="shrink-0 text-[#ea580c] transition group-hover:translate-x-0.5"
            />
          </RouterLink>

          <!-- Error state -->
          <section
            v-if="errorMessage"
            class="rounded-xl border border-[#fecaca] bg-white p-5"
          >
            <div class="flex items-start gap-3">
              <PhWarningCircle
                :size="20"
                class="mt-0.5 shrink-0 text-[#dc2626]"
                weight="duotone"
              />
              <div>
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

          <!-- Class performance -->
          <section class="rounded-xl border border-[#ebe7df] bg-white shadow-sm p-5">
            <div class="mb-4 flex items-center justify-between gap-3">
              <div>
                <h2 class="text-sm font-semibold text-[#171322]">
                  Kelas dan Mata Pelajaran
                </h2>
                <p class="mt-0.5 text-xs text-[#6b7280]">
                  Performa dari kelas yang sedang Anda ajar.
                </p>
              </div>
              <RouterLink
                to="/teacher/subjects"
                class="inline-flex shrink-0 items-center gap-1 text-xs font-medium text-[#4f46e5] transition hover:text-[#4338ca]"
              >
                Lihat semua
                <PhArrowRight :size="13" />
              </RouterLink>
            </div>

            <div v-if="loading" class="grid gap-3 md:grid-cols-2">
              <div
                v-for="i in 4"
                :key="i"
                class="h-28 animate-pulse rounded-lg bg-[#f3f1ec]"
              />
            </div>

            <div
              v-else-if="summary?.classPerformance?.length"
              class="grid gap-3 md:grid-cols-2"
            >
              <article
                v-for="item in summary.classPerformance"
                :key="`${item.classId}-${item.subjectName}`"
                class="min-w-0 rounded-lg bg-[#fbfaf8] p-4"
              >
                <div class="flex min-w-0 items-start gap-3">
                  <div
                    class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg text-white"
                    :style="{
                      backgroundColor: resolveSubjectColor({
                        subjectColor: item.subjectColor,
                        subjectId: item.classId,
                        subjectName: item.subjectName,
                      }),
                    }"
                  >
                    <PhBookOpen :size="18" weight="duotone" />
                  </div>
                  <div class="min-w-0 flex-1">
                    <p class="truncate text-sm font-semibold text-[#171322]">
                      {{ item.subjectName }}
                    </p>
                    <p class="mt-0.5 truncate text-xs text-[#6b7280]">
                      {{ item.className }}
                    </p>
                  </div>
                </div>

                <dl class="mt-4 grid grid-cols-3 gap-2 border-t border-[#ebe7df] pt-3 text-xs">
                  <div>
                    <dt class="text-[#9ca3af]">Siswa</dt>
                    <dd class="mt-1 font-semibold text-[#171322]">
                      {{ item.totalStudents }}
                    </dd>
                  </div>
                  <div>
                    <dt class="text-[#9ca3af]">Terkumpul</dt>
                    <dd
                      class="mt-1 font-semibold"
                      :class="
                        item.submissionRate >= 80
                          ? 'text-[#059669]'
                          : item.submissionRate >= 50
                            ? 'text-[#ea580c]'
                            : 'text-[#dc2626]'
                      "
                    >
                      {{ formatPercentage(item.submissionRate) }}
                    </dd>
                  </div>
                  <div>
                    <dt class="text-[#9ca3af]">Rata-rata</dt>
                    <dd class="mt-1 font-semibold text-[#171322]">
                      {{ item.averageScore.toFixed(1) }}
                    </dd>
                  </div>
                </dl>
              </article>
            </div>

            <div
              v-else
              class="rounded-lg bg-[#fbfaf8] px-4 py-8 text-center"
            >
              <PhChalkboardTeacher
                class="mx-auto h-7 w-7 text-[#d1d5db]"
                weight="duotone"
              />
              <p class="mt-3 text-sm font-semibold text-[#171322]">
                Belum ada ringkasan kelas
              </p>
              <p class="mt-1 text-xs leading-5 text-[#6b7280]">
                Ringkasan akan tampil setelah data kelas dan aktivitas belajar
                tersedia.
              </p>
            </div>
          </section>

          <!-- Quick Actions -->
          <section>
            <h2 class="mb-3 text-sm font-semibold text-[#171322]">
              Aksi Cepat
            </h2>
            <div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
              <RouterLink
                v-for="action in quickActions"
                :key="action.label"
                :to="action.to"
                class="group flex items-center gap-3 rounded-xl border border-[#ebe7df] bg-white shadow-sm p-3.5 transition hover:-translate-y-0.5 hover:shadow-sm"
                :class="action.border"
              >
                <div
                  class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg"
                  :class="action.color"
                >
                  <component :is="action.icon" :size="18" weight="duotone" />
                </div>
                <span class="min-w-0 flex-1 text-sm font-medium text-[#171322]">
                  {{ action.label }}
                </span>
                <PhArrowRight
                  :size="14"
                  class="shrink-0 text-[#d1d5db] transition group-hover:translate-x-0.5 group-hover:text-[#4f46e5]"
                />
              </RouterLink>
            </div>
          </section>
        </template>
      </div>
    </section>

    <!-- Right sidebar -->
    <aside
      class="min-w-0 border-t border-[#ebe7df] bg-[#f8f7f4] xl:sticky xl:top-0 xl:h-dvh xl:min-h-0 xl:overflow-y-auto xl:border-l xl:border-t-0 xl:bg-white"
    >
      <div class="flex flex-col gap-4 p-5">
        <AcademicActivityCard
          :activities="activities"
          :loading="activitiesLoading"
          :error="activitiesError"
          role="teacher"
          :max-items="5"
        />
        <LatestChatCard to="/teacher/chat" :limit="4" />
      </div>
    </aside>
  </main>
</template>
