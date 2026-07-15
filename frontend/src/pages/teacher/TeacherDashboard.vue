<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import {
  PhArrowRight,
  PhBookOpen,
  PhChalkboardTeacher,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import { getTeacherDashboard } from "../../services/teacherDashboard";
import { getAcademicActivities } from "../../services/activity";
import { useNotificationUnreadCount } from "../../composables/useNotificationUnreadCount";
import type { MembershipInfo } from "../../types/auth";
import type { TeacherDashboardSummary } from "../../types/teacherDashboard";
import type { AcademicActivityItem } from "../../types/activity";
import { resolveSubjectColor } from "../../utils/color";
import LatestChatCard from "../../components/chat/LatestChatCard.vue";
import AcademicActivityCard from "../../components/activity/AcademicActivityCard.vue";
import ActivityCalendarCard from "../../components/dashboard/ActivityCalendarCard.vue";
import NotificationsPanel from "../../components/dashboard/NotificationsPanel.vue";
import DashboardUpdatesPanel from "../../components/dashboard/DashboardUpdatesPanel.vue";
import ContextSwitcher from "../../components/layout/ContextSwitcher.vue";

const auth = useAuthStore();
const notificationUnread = useNotificationUnreadCount();
const chatPanelUnreadCount = ref(0);

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
    activeMembership.value?.schoolUserId ??
    auth.defaultContext?.schoolUserId ??
    "",
);

const teacherName = computed(() => auth.user?.fullName ?? "Guru");
const firstName = computed(() => teacherName.value.split(" ")[0]);

const pendingReviews = computed(() => summary.value?.pendingReviews ?? 0);
const hasPendingReviews = computed(() => pendingReviews.value > 0);

function formatPercentage(value: number | null | undefined) {
  if (typeof value !== "number" || !Number.isFinite(value)) return "0%";
  return `${Math.min(100, Math.max(0, value)).toFixed(1)}%`;
}

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

function updateChatPanelUnreadCount(count: number) {
  chatPanelUnreadCount.value = Math.max(0, count);
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
    class="grid min-h-screen min-w-0 flex-1 grid-cols-1 overflow-x-hidden bg-background xl:grid-cols-[minmax(0,1fr)_300px]"
  >
    <!-- Main content -->
    <section
      class="min-w-0 xl:flex xl:h-dvh xl:min-h-0 xl:flex-col xl:overflow-hidden"
    >
      <!-- Header -->
      <header class="border-b border-border bg-surface xl:shrink-0">
        <div
          class="flex min-w-0 flex-col gap-3 px-5 py-5 sm:px-6 lg:flex-row lg:items-center lg:justify-between lg:px-8"
        >
          <div class="min-w-0">
            <h1 class="mt-1 text-2xl font-semibold text-foreground sm:text-3xl">
              Selamat mengajar, {{ firstName }}
            </h1>
            <p class="mt-2 text-sm leading-6 text-muted">
              <template v-if="loading">Memuat status pengumpulan...</template>
              <template v-else-if="!hasPendingReviews">
                Semua pengumpulan sudah selesai dinilai. 🎉
              </template>
              <template v-else>
                Ada {{ pendingReviews }} pengumpulan yang perlu Anda nilai hari
                ini.
                <RouterLink
                  to="/teacher/submissions"
                  class="font-medium text-brand hover:text-brand-hover"
                >
                  Lihat detail →
                </RouterLink>
              </template>
            </p>
          </div>
          <ContextSwitcher />
        </div>
      </header>

      <div
        class="space-y-5 px-5 py-5 sm:px-6 xl:flex xl:min-h-0 xl:flex-1 xl:flex-col xl:overflow-hidden xl:px-8 xl:py-6"
      >
        <!-- No school context -->
        <section
          v-if="!schoolUserId"
          class="flex min-h-[55vh] items-center justify-center"
        >
          <article
            class="w-full max-w-xl rounded-xl border border-border bg-surface shadow-sm p-8 text-center"
          >
            <div
              class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-warning-soft text-[#ea580c]"
            >
              <PhWarningCircle :size="24" weight="duotone" />
            </div>
            <h2 class="mt-3 text-lg font-medium text-foreground">
              Konteks guru belum tersedia
            </h2>
            <p class="mx-auto mt-2 max-w-md text-sm leading-6 text-muted">
              Pastikan akun guru memiliki akses aktif pada sekolah yang sedang
              digunakan.
            </p>
          </article>
        </section>

        <template v-else>
          <AcademicActivityCard
            class="xl:shrink-0"
            :activities="activities"
            :loading="activitiesLoading"
            :error="activitiesError"
            role="teacher"
            :max-items="5"
          />

          <!-- Error state -->
          <section
            v-if="errorMessage"
            class="rounded-xl border border-danger-line bg-danger-soft p-5 xl:shrink-0"
          >
            <div class="flex items-start gap-3">
              <PhWarningCircle
                :size="20"
                class="mt-0.5 shrink-0 text-danger"
                weight="duotone"
              />
              <div>
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

          <!-- Class performance -->
          <section
            class="rounded-xl border border-border bg-surface shadow-sm p-5 xl:flex xl:min-h-0 xl:flex-1 xl:flex-col xl:overflow-hidden"
          >
            <div class="mb-4 flex shrink-0 items-center justify-between gap-3">
              <div>
                <h2 class="text-sm font-semibold text-foreground">
                  Kelas dan Mata Pelajaran
                </h2>
                <p class="mt-0.5 text-xs text-muted">
                  Performa dari kelas yang sedang Anda ajar.
                </p>
              </div>
              <RouterLink
                to="/teacher/subjects"
                class="inline-flex shrink-0 items-center gap-1 text-xs font-medium text-brand transition hover:text-brand-hover"
              >
                Lihat semua
                <PhArrowRight :size="13" />
              </RouterLink>
            </div>

            <div v-if="loading" class="grid gap-3 md:grid-cols-2">
              <div
                v-for="i in 4"
                :key="i"
                class="h-28 animate-pulse rounded-lg bg-surface-strong"
              />
            </div>

            <div
              v-else-if="summary?.classPerformance?.length"
              class="grid gap-3 md:grid-cols-2 max-h-96 overflow-y-auto pr-1 xl:max-h-none xl:min-h-0 xl:flex-1"
            >
              <article
                v-for="item in summary.classPerformance"
                :key="`${item.classId}-${item.subjectName}`"
                class="min-w-0 rounded-lg bg-surface-subtle p-4"
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
                    <p class="truncate text-sm font-semibold text-foreground">
                      {{ item.subjectName }}
                    </p>
                    <p class="mt-0.5 truncate text-xs text-muted">
                      {{ item.className }}
                    </p>
                  </div>
                </div>

                <dl
                  class="mt-4 grid grid-cols-3 gap-2 border-t border-border pt-3 text-xs"
                >
                  <div>
                    <dt class="text-muted">Siswa</dt>
                    <dd class="mt-1 font-semibold text-foreground">
                      {{ item.totalStudents }}
                    </dd>
                  </div>
                  <div>
                    <dt class="text-muted">Terkumpul</dt>
                    <dd
                      class="mt-1 font-semibold"
                      :class="
                        item.submissionRate >= 80
                          ? 'text-success'
                          : item.submissionRate >= 50
                            ? 'text-[#ea580c]'
                            : 'text-danger'
                      "
                    >
                      {{ formatPercentage(item.submissionRate) }}
                    </dd>
                  </div>
                  <div>
                    <dt class="text-muted">Rata-rata</dt>
                    <dd class="mt-1 font-semibold text-foreground">
                      {{ item.averageScore.toFixed(1) }}
                    </dd>
                  </div>
                </dl>
              </article>
            </div>

            <div
              v-else
              class="rounded-lg bg-surface-subtle px-4 py-8 text-center"
            >
              <PhChalkboardTeacher
                class="mx-auto h-7 w-7 text-border-strong"
                weight="duotone"
              />
              <p class="mt-3 text-sm font-semibold text-foreground">
                Belum ada ringkasan kelas
              </p>
              <p class="mt-1 text-xs leading-5 text-muted">
                Ringkasan akan tampil setelah kamu dihubungkan ke kelas aktif
                dan siswa mulai beraktivitas.
              </p>
              <RouterLink
                to="/teacher/subjects"
                class="mt-3 inline-flex items-center gap-1.5 rounded-lg border border-border bg-surface px-3 py-2 text-xs font-medium text-brand transition hover:border-brand hover:bg-brand-soft"
              >
                Lihat mata pelajaran
                <PhArrowRight :size="13" />
              </RouterLink>
            </div>
          </section>
        </template>
      </div>
    </section>

    <!-- Right sidebar -->
    <aside
      class="min-w-0 border-t border-border bg-background xl:sticky xl:top-0 xl:h-dvh xl:min-h-0 xl:overflow-hidden xl:border-l xl:border-t-0 xl:bg-surface"
    >
      <div
        class="flex flex-col gap-3 p-4 xl:h-full xl:min-h-0 xl:overflow-hidden"
      >
        <DashboardUpdatesPanel
          class="xl:min-h-0 xl:flex-1 xl:overflow-hidden"
          :tabs="['notifications', 'chat']"
          :notification-badge="notificationUnread.unreadCount.value"
          :chat-badge="chatPanelUnreadCount"
        >
          <template #notifications>
            <NotificationsPanel embedded to="/teacher/notifications" />
          </template>
          <template #chat>
            <LatestChatCard
              to="/teacher/chat"
              :limit="4"
              embedded
              @unread-change="updateChatPanelUnreadCount"
            />
          </template>
        </DashboardUpdatesPanel>

        <div class="max-h-96 shrink-0 overflow-y-auto">
          <ActivityCalendarCard role="teacher" />
        </div>
      </div>
    </aside>
  </main>
</template>
