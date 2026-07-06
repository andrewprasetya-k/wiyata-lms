<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import {
  PhArrowRight,
  PhBookOpen,
  PhCalendarBlank,
  PhChalkboardTeacher,
  PhClipboardText,
  PhMegaphone,
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

const auth = useAuthStore();

const loading = ref(false);
const errorMessage = ref("");
const summary = ref<TeacherDashboardSummary | null>(null);
const activities = ref<AcademicActivityItem[]>([]);
const activitiesLoading = ref(false);
const activitiesError = ref("");

const activeMembership = computed<MembershipInfo | undefined>(() => {
  const activeSchoolMembership = auth.memberships.find(
    (membership) => membership.school.id === auth.activeSchoolId,
  );
  const defaultMembership = auth.memberships.find(
    (membership) => membership.isDefault,
  );
  return activeSchoolMembership ?? defaultMembership ?? auth.memberships[0];
});

const schoolUserId = computed(() => {
  return (
    activeMembership.value?.schoolUserId ??
    auth.defaultContext?.schoolUserId ??
    ""
  );
});

const activeSchoolName = computed(
  () => activeMembership.value?.school.name ?? "Sekolah aktif",
);
const teacherName = computed(() => auth.user?.fullName ?? "Guru");

function formatPercentage(value: number | null | undefined) {
  if (typeof value !== "number" || !Number.isFinite(value)) {
    return "0.0%";
  }
  return `${Math.min(100, Math.max(0, value)).toFixed(1)}%`;
}

const stats = computed(() => [
  {
    label: "Menunggu Penilaian",
    value: summary.value?.pendingReviews,
    helper: "Pengumpulan yang belum dinilai",
    tone: "#e58f86",
  },
  {
    label: "Total siswa",
    value: summary.value?.totalStudents,
    helper: "Siswa unik dari kelas yang diajar",
    tone: "#74bfa5",
  },
  {
    label: "Pengumpulan tugas",
    value: summary.value
      ? formatPercentage(summary.value.submissionRate)
      : undefined,
    helper: "Pengumpulan dibanding siswa aktif per tugas",
    tone: "#7aa7d9",
  },
]);

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
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-[#f8f7f4]">
    <header class="border-b border-[#ebe7df] bg-white">
      <div
        class="flex min-w-0 flex-col gap-3 px-5 py-5 sm:px-6 lg:flex-row lg:items-center lg:justify-between lg:px-8"
      >
        <div class="min-w-0">
          <h1 class="text-2xl font-semibold text-[#171322] sm:text-3xl">
            Selamat mengajar, {{ teacherName }}
          </h1>
          <p class="mt-2 max-w-3xl text-sm leading-6 text-[#6b7280]">
            Pantau kelas, pengumpulan, dan tugas dari ruang kerja guru.
          </p>
        </div>
        <div class="flex min-w-0 flex-col gap-2 sm:flex-row sm:items-center">
          <div
            class="min-w-0 rounded-lg border border-[#ebe7df] bg-[#f9fafb] px-3 py-2 text-xs text-[#6b7280]"
          >
            <span class="block truncate font-medium text-[#171322]">
              {{ activeSchoolName }}
            </span>
          </div>
        </div>
      </div>
    </header>

    <section class="space-y-5 px-5 py-5 sm:px-6 lg:px-8 lg:py-6">
      <section
        v-if="!schoolUserId"
        class="flex min-h-[55vh] items-center justify-center"
      >
        <article
          class="w-full max-w-xl rounded-xl border border-[#ebe7df] bg-white p-8 text-center"
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
        <section class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          <article
            v-for="item in stats"
            :key="item.label"
            class="rounded-xl border border-[#ebe7df] bg-white p-4"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <p class="text-xs text-[#7a7385]">{{ item.label }}</p>
                <p class="mt-2 text-2xl font-medium text-[#171322]">
                  {{ item.value ?? (loading ? "..." : "-") }}
                </p>
              </div>
              <div
                class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg text-white"
                :style="{ backgroundColor: item.tone }"
              >
                <PhClipboardText :size="18" weight="duotone" />
              </div>
            </div>
            <p class="mt-2 text-xs leading-5 text-[#8a8494]">
              {{ item.helper }}
            </p>
          </article>
        </section>

        <section v-if="loading" class="grid gap-4 xl:grid-cols-[1.35fr_0.75fr]">
          <div
            class="h-80 animate-pulse rounded-xl border border-[#ebe7df] bg-white"
          />
          <div class="space-y-3">
            <div
              v-for="item in 3"
              :key="item"
              class="h-28 animate-pulse rounded-xl border border-[#ebe7df] bg-white"
            />
          </div>
        </section>

        <section
          v-else-if="errorMessage"
          class="rounded-xl border border-[#f1d6d3] bg-white p-6"
        >
          <div
            class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"
          >
            <div class="flex items-start gap-3">
              <PhWarningCircle
                :size="22"
                class="mt-0.5 shrink-0 text-[#dc2626]"
                weight="duotone"
              />
              <div>
                <h2 class="text-base font-medium text-[#171322]">
                  Dashboard tidak dapat dimuat
                </h2>
                <p class="mt-1 text-sm leading-6 text-[#7a7385]">
                  {{ errorMessage }}
                </p>
              </div>
            </div>
            <button
              class="rounded-lg bg-[#4f46e5] px-4 py-2 text-sm font-medium text-white transition hover:bg-[#4338ca]"
              type="button"
              @click="loadDashboard"
            >
              Coba lagi
            </button>
          </div>
        </section>

        <section
          v-else
          class="grid min-w-0 gap-4 xl:grid-cols-[minmax(0,1.35fr)_300px]"
        >
          <article
            class="min-w-0 rounded-xl border border-[#ebe7df] bg-white p-4 sm:p-5"
          >
            <div class="flex min-w-0 items-center justify-between gap-3">
              <div class="min-w-0">
                <h2 class="text-sm font-medium text-[#171322]">
                  Ringkasan kelas dan mata pelajaran
                </h2>
                <p class="mt-1 text-xs text-[#7a7385]">
                  Performa dari kelas yang sedang Anda ajar.
                </p>
              </div>
              <RouterLink
                to="/teacher/subjects"
                class="inline-flex shrink-0 items-center gap-1 text-xs font-medium text-[#4f46e5] sm:text-sm"
              >
                Lihat semua
                <PhArrowRight :size="15" />
              </RouterLink>
            </div>

            <div
              v-if="summary?.classPerformance?.length"
              class="mt-4 grid gap-3 md:grid-cols-2"
            >
              <article
                v-for="item in summary.classPerformance"
                :key="`${item.classId}-${item.subjectName}`"
                class="min-w-0 rounded-lg border border-[#ebe7df] bg-[#fbfaf8] p-4"
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
                  <div class="min-w-0">
                    <p class="truncate text-sm font-medium text-[#171322]">
                      {{ item.subjectName }}
                    </p>
                    <p class="mt-1 truncate text-xs text-[#7a7385]">
                      {{ item.className }}
                    </p>
                  </div>
                </div>
                <dl class="mt-4 grid grid-cols-3 gap-2 text-xs">
                  <div>
                    <dt class="text-[#8a8494]">Siswa</dt>
                    <dd class="mt-1 font-medium text-[#171322]">
                      {{ item.totalStudents }}
                    </dd>
                  </div>
                  <div>
                    <dt class="text-[#8a8494]">Terkumpul</dt>
                    <dd class="mt-1 font-medium text-[#171322]">
                      {{ formatPercentage(item.submissionRate) }}
                    </dd>
                  </div>
                  <div>
                    <dt class="text-[#8a8494]">Rata-rata</dt>
                    <dd class="mt-1 font-medium text-[#171322]">
                      {{ item.averageScore.toFixed(1) }}
                    </dd>
                  </div>
                </dl>
              </article>
            </div>

            <div
              v-else
              class="mt-4 rounded-lg border border-[#ebe7df] bg-[#fbfaf8] p-5"
            >
              <h3 class="text-sm font-semibold text-[#171322]">
                Belum ada ringkasan mata pelajaran
              </h3>
              <p class="mt-1 text-sm leading-6 text-[#6b7280]">
                Ringkasan akan tampil setelah data kelas dan aktivitas belajar
                tersedia.
              </p>
            </div>
          </article>

          <aside class="grid min-w-0 gap-3">
            <AcademicActivityCard
              :activities="activities"
              :loading="activitiesLoading"
              :error="activitiesError"
              role="teacher"
              :max-items="5"
            />

            <LatestChatCard to="/teacher/chat" :limit="4" />

            <RouterLink
              to="/teacher/submissions"
              class="group rounded-xl border border-[#ebe7df] bg-white p-4 transition hover:border-[#c7d2fe]"
            >
              <div class="flex items-start gap-3">
                <div
                  class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-[#eef2ff] text-[#4f46e5]"
                >
                  <PhChalkboardTeacher :size="19" weight="duotone" />
                </div>
                <div class="min-w-0 flex-1">
                  <p class="text-sm font-medium text-[#171322]">Pengumpulan</p>
                  <p class="mt-1 text-xs leading-5 text-[#7a7385]">
                    Pantau dan lanjutkan penilaian pengumpulan siswa.
                  </p>
                </div>
                <PhArrowRight
                  :size="15"
                  class="mt-1 text-[#a09aa8] group-hover:text-[#4f46e5]"
                />
              </div>
            </RouterLink>

            <RouterLink
              to="/teacher/assignments"
              class="group rounded-xl border border-[#ebe7df] bg-white p-4 transition hover:border-[#bbf7d0]"
            >
              <div class="flex items-start gap-3">
                <div
                  class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-[#ecfdf5] text-[#059669]"
                >
                  <PhCalendarBlank :size="19" weight="duotone" />
                </div>
                <div class="min-w-0 flex-1">
                  <p class="text-sm font-medium text-[#171322]">Tugas Saya</p>
                  <p class="mt-1 text-xs leading-5 text-[#7a7385]">
                    Kelola tugas dari semua mata pelajaran yang diajar.
                  </p>
                </div>
                <PhArrowRight
                  :size="15"
                  class="mt-1 text-[#a09aa8] group-hover:text-[#059669]"
                />
              </div>
            </RouterLink>

            <RouterLink
              to="/teacher/feed"
              class="group rounded-xl border border-[#ebe7df] bg-white p-4 transition hover:border-[#fed7aa]"
            >
              <div class="flex items-start gap-3">
                <div
                  class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-[#fff7ed] text-[#ea580c]"
                >
                  <PhMegaphone :size="19" weight="duotone" />
                </div>
                <div class="min-w-0 flex-1">
                  <p class="text-sm font-medium text-[#171322]">Feed Kelas</p>
                  <p class="mt-1 text-xs leading-5 text-[#7a7385]">
                    Kirim pengumuman untuk kelas yang Anda ajar.
                  </p>
                </div>
                <PhArrowRight
                  :size="15"
                  class="mt-1 text-[#a09aa8] group-hover:text-[#ea580c]"
                />
              </div>
            </RouterLink>
          </aside>
        </section>
      </template>
    </section>
  </main>
</template>
