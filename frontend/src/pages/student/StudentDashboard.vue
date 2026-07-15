<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import {
  PhArrowClockwise,
  PhArrowRight,
  PhBookOpen,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import { useActiveClassStore } from "../../stores/activeClass";
import { getSubjectClassesByClass } from "../../services/classWorkspace";
import { getClassFeed } from "../../services/feed";
import { getStudentAssignmentInbox } from "../../services/assignment";
import { useFeedUnreadCount } from "../../composables/useFeedUnreadCount";
import { useNotificationUnreadCount } from "../../composables/useNotificationUnreadCount";
import { getNotifications } from "../../services/notifications";
import { getAcademicActivities } from "../../services/activity";
import type { SubjectClassItem } from "../../types/classWorkspace";
import type { FeedPost } from "../../types/feed";
import type { NotificationItem } from "../../types/dashboard";
import type { StudentAssignmentInboxItem } from "../../types/assignment";
import type { AcademicActivityItem } from "../../types/activity";
import { formatDate, formatDateTime } from "../../utils/date";
import { resolveSubjectColor } from "../../utils/color";
import LatestChatCard from "../../components/chat/LatestChatCard.vue";
import AcademicActivityCard from "../../components/activity/AcademicActivityCard.vue";
import DashboardUpdatesPanel from "../../components/dashboard/DashboardUpdatesPanel.vue";

import ActivityCalendarCard from "../../components/dashboard/ActivityCalendarCard.vue";

import ContextSwitcher from "../../components/layout/ContextSwitcher.vue";
import {
  compareAssignments,
  assignmentStatusClasses,
  assignmentStatusLabel,
} from "../../utils/studentAssignmentPreview";

import NotificationsPanel from "../../components/dashboard/NotificationsPanel.vue";

const auth = useAuthStore();
const activeClassStore = useActiveClassStore();

const { unreadCount: feedPanelUnreadCount } = useFeedUnreadCount();
const notificationUnread = useNotificationUnreadCount();

const subjects = ref<SubjectClassItem[]>([]);
const feedPosts = ref<FeedPost[]>([]);
const assignmentPreviewItems = ref<StudentAssignmentInboxItem[]>([]);
const notifications = ref<NotificationItem[]>([]);
const isLoading = ref(true);
const notificationsLoading = ref(false);
const notificationsError = ref("");
const assignmentsLoading = ref(false);
const assignmentsError = ref("");
const subjectPreviewLoading = ref(false);
const subjectPreviewError = ref("");
const feedPreviewLoading = ref(false);
const feedPreviewError = ref("");
const activities = ref<AcademicActivityItem[]>([]);
const activitiesLoading = ref(false);
const activitiesError = ref("");
const chatPanelUnreadCount = ref(0);

const errorMessage = ref("");
let subjectPreviewRequestId = 0;
let feedPreviewRequestId = 0;

interface ClassContextLoadResult {
  canLoadAreas: boolean;
  activeClassId: string | null;
}

const schoolUserId = computed(() => auth.activeSchoolUserId);
const firstName = computed(() => auth.user?.fullName?.split(" ")[0] ?? "Siswa");
const assignmentPreview = computed(() =>
  [...assignmentPreviewItems.value].sort(compareAssignments),
);

async function loadDashboard(selectedClassId?: string) {
  const classContext = await loadClassContext(selectedClassId);
  if (!classContext.canLoadAreas) return;

  await Promise.allSettled([
    loadNotifications(),
    loadAssignmentPreview(),
    classContext.activeClassId
      ? loadSubjectPreview(classContext.activeClassId)
      : Promise.resolve(),
    classContext.activeClassId
      ? loadFeedPreview(classContext.activeClassId)
      : Promise.resolve(),
  ]);
}

async function loadClassContext(
  selectedClassId?: string,
): Promise<ClassContextLoadResult> {
  if (!auth.user?.id) {
    errorMessage.value = "Sesi login belum lengkap. Silakan login ulang.";
    isLoading.value = false;
    return { canLoadAreas: false, activeClassId: null };
  }

  if (!schoolUserId.value) {
    errorMessage.value = "Konteks sekolah belum tersedia.";
    isLoading.value = false;
    return { canLoadAreas: false, activeClassId: null };
  }

  isLoading.value = true;
  errorMessage.value = "";
  subjects.value = [];
  feedPosts.value = [];
  subjectPreviewError.value = "";
  feedPreviewError.value = "";
  subjectPreviewRequestId += 1;
  feedPreviewRequestId += 1;
  subjectPreviewLoading.value = false;
  feedPreviewLoading.value = false;

  try {
    await activeClassStore.loadClasses(schoolUserId.value);

    const activeClassId = selectedClassId ?? activeClassStore.activeClassId;
    if (!activeClassId) {
      subjects.value = [];
      feedPosts.value = [];
      return { canLoadAreas: true, activeClassId: null };
    }

    return { canLoadAreas: true, activeClassId };
  } catch {
    errorMessage.value =
      "Dashboard belum bisa dimuat. Periksa koneksi atau coba lagi nanti.";
    return { canLoadAreas: false, activeClassId: null };
  } finally {
    isLoading.value = false;
  }
}

async function loadNotifications() {
  notificationsLoading.value = true;
  notificationsError.value = "";

  try {
    const notificationData = await getNotifications({ page: 1, limit: 5 });
    notifications.value = notificationData.data ?? [];
    notificationUnread.set(notificationData.unreadCount ?? 0);
  } catch {
    notifications.value = [];
    notificationsError.value = "Notifikasi belum bisa dimuat.";
  } finally {
    notificationsLoading.value = false;
  }
}

async function loadAssignmentPreview() {
  assignmentsLoading.value = true;
  assignmentsError.value = "";

  try {
    const response = await getStudentAssignmentInbox();
    assignmentPreviewItems.value = response.items ?? [];
  } catch {
    assignmentPreviewItems.value = [];
    assignmentsError.value = "Preview tugas belum bisa dimuat.";
  } finally {
    assignmentsLoading.value = false;
  }
}

async function loadSubjectPreview(activeClassId: string) {
  const requestId = (subjectPreviewRequestId += 1);
  subjectPreviewLoading.value = true;
  subjectPreviewError.value = "";

  try {
    const subjectData = await getSubjectClassesByClass(activeClassId);
    if (
      requestId !== subjectPreviewRequestId ||
      activeClassId !== activeClassStore.activeClassId
    )
      return;
    subjects.value = subjectData.subjects ?? [];
  } catch {
    if (
      requestId !== subjectPreviewRequestId ||
      activeClassId !== activeClassStore.activeClassId
    )
      return;
    subjects.value = [];
    subjectPreviewError.value = "Mata pelajaran belum bisa dimuat.";
  } finally {
    if (
      requestId === subjectPreviewRequestId &&
      activeClassId === activeClassStore.activeClassId
    ) {
      subjectPreviewLoading.value = false;
    }
  }
}

async function loadFeedPreview(activeClassId: string) {
  const requestId = (feedPreviewRequestId += 1);
  feedPreviewLoading.value = true;
  feedPreviewError.value = "";

  try {
    const feedData = await getClassFeed(activeClassId);
    if (
      requestId !== feedPreviewRequestId ||
      activeClassId !== activeClassStore.activeClassId
    )
      return;
    feedPosts.value = feedData.data.data ?? [];
  } catch {
    if (
      requestId !== feedPreviewRequestId ||
      activeClassId !== activeClassStore.activeClassId
    )
      return;
    feedPosts.value = [];
    feedPreviewError.value = "Feed kelas belum bisa dimuat.";
  } finally {
    if (
      requestId === feedPreviewRequestId &&
      activeClassId === activeClassStore.activeClassId
    ) {
      feedPreviewLoading.value = false;
    }
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
    activitiesError.value = "Aktivitas akademik belum bisa dimuat.";
  } finally {
    activitiesLoading.value = false;
  }
}

function updateChatPanelUnreadCount(count: number) {
  chatPanelUnreadCount.value = Math.max(0, count);
}

function retryFeedPreview() {
  const activeClassId = activeClassStore.activeClassId;
  if (!activeClassId) return;
  void loadFeedPreview(activeClassId);
}

onMounted(() => {
  loadDashboard();
  loadActivities();
});
</script>

<template>
  <main
    class="grid min-h-screen min-w-0 flex-1 grid-cols-1 overflow-x-hidden bg-background lg:grid-cols-[minmax(0,1fr)_320px]"
  >
    <section
      class="min-w-0 lg:flex lg:h-dvh lg:min-h-0 lg:flex-col lg:overflow-hidden"
    >
      <header class="border-b border-border bg-surface lg:shrink-0">
        <div
          class="flex min-w-0 flex-col gap-3 px-5 py-4 sm:px-6 lg:flex-row lg:items-center lg:justify-between lg:px-8"
        >
          <div class="min-w-0">
            <h1 class="mt-1 text-2xl font-semibold text-foreground sm:text-3xl">
              Selamat datang, {{ firstName }}
            </h1>
            <p class="mt-2 max-w-3xl text-sm leading-6 text-muted">
              Mau belajar apa hari ini?
            </p>
          </div>
          <ContextSwitcher />
        </div>
      </header>

      <div
        class="space-y-5 px-5 py-5 sm:px-6 lg:flex lg:min-h-0 lg:flex-1 lg:flex-col lg:overflow-hidden lg:px-8 lg:py-6"
      >
        <AcademicActivityCard
          class="lg:shrink-0"
          :activities="activities"
          :loading="activitiesLoading"
          :error="activitiesError"
          role="student"
          :max-items="3"
        />

        <section
          v-if="errorMessage"
          class="rounded-xl border border-danger-line bg-danger-soft p-5 sm:p-6"
        >
          <div class="flex items-start gap-3">
            <div
              class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-danger-soft text-danger"
            >
              <PhWarningCircle :size="22" weight="duotone" />
            </div>
            <div class="min-w-0">
              <h2 class="text-base font-medium text-foreground">
                Dashboard tidak dapat dimuat
              </h2>
              <p class="mt-1 text-sm leading-6 text-muted">
                {{ errorMessage }}
              </p>
              <button
                class="mt-4 rounded-lg bg-brand px-4 py-2 text-sm font-medium text-white transition hover:bg-brand-hover"
                type="button"
                @click="loadDashboard()"
              >
                Coba lagi
              </button>
            </div>
          </div>
        </section>

        <section v-if="isLoading" class="space-y-4">
          <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
            <div
              v-for="item in 3"
              :key="item"
              class="h-36 animate-pulse rounded-xl border border-border bg-surface shadow-sm"
            />
          </div>
          <div class="grid gap-4 xl:grid-cols-2">
            <div
              v-for="item in 2"
              :key="`panel-${item}`"
              class="h-64 animate-pulse rounded-xl border border-border bg-surface shadow-sm"
            />
          </div>
        </section>

        <section
          v-else-if="!activeClassStore.activeClassId"
          class="flex min-h-[50vh] items-center justify-center"
        >
          <article
            class="w-full max-w-xl rounded-xl border border-border bg-surface shadow-sm p-8 text-center"
          >
            <div
              class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-brand-soft text-brand"
            >
              <PhBookOpen :size="25" weight="duotone" />
            </div>
            <h2 class="mt-3 text-lg font-medium text-foreground">
              Belum ada kelas aktif
            </h2>
            <p class="mx-auto mt-2 max-w-md text-sm leading-6 text-muted">
              Kelas akan muncul setelah akunmu terdaftar sebagai anggota kelas
              di sekolah aktif.
            </p>
          </article>
        </section>

        <section
          v-else
          class="flex min-w-0 flex-col gap-4 lg:min-h-0 lg:flex-1 xl:grid xl:grid-cols-[1.15fr_0.85fr]"
        >
          <article
            class="min-w-0 rounded-xl border border-border bg-surface shadow-sm p-4 sm:p-5 lg:flex lg:min-h-0 lg:flex-1 lg:flex-col lg:overflow-hidden"
          >
            <div
              class="mb-4 flex min-w-0 shrink-0 items-center justify-between gap-3"
            >
              <div class="min-w-0">
                <p class="text-sm font-medium text-foreground">
                  Daftar mata pelajaran
                </p>
                <p class="mt-1 text-xs text-muted">
                  Buka materi dan tugas dari kelas aktif.
                </p>
              </div>
              <RouterLink
                class="shrink-0 text-xs font-medium text-brand transition hover:text-brand-hover sm:text-sm"
                to="/student/subjects"
              >
                Lihat semua
              </RouterLink>
            </div>

            <div v-if="subjectPreviewLoading" class="grid gap-3 sm:grid-cols-2">
              <div
                v-for="item in 4"
                :key="item"
                class="h-24 animate-pulse rounded-lg bg-surface-subtle"
              />
            </div>

            <div
              v-else-if="subjectPreviewError"
              class="rounded-lg bg-surface-subtle p-5"
            >
              <p class="mt-2 text-sm leading-6 text-muted">
                {{ subjectPreviewError }}
              </p>
            </div>

            <div
              v-else-if="subjects.length > 0"
              class="grid gap-3 sm:grid-cols-2 max-h-80 overflow-y-auto pr-1 lg:max-h-none lg:min-h-0 lg:flex-1"
            >
              <RouterLink
                v-for="subject in subjects"
                :key="subject.subjectClassId"
                class="group min-w-0 overflow-hidden rounded-lg bg-surface-subtle transition hover:-translate-y-0.5 hover:bg-surface hover:shadow-[0_18px_40px_rgba(66,55,40,0.08)]"
                :to="`/student/subjects/${subject.subjectClassId}`"
              >
                <div
                  class="h-1.5 w-full"
                  :style="{
                    backgroundColor: resolveSubjectColor(subject),
                  }"
                />
                <div class="flex min-w-0 items-start gap-3 p-4">
                  <div
                    class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg text-white"
                    :style="{
                      backgroundColor: resolveSubjectColor(subject),
                    }"
                  >
                    <PhBookOpen :size="18" weight="duotone" />
                  </div>
                  <div class="min-w-0 flex-1">
                    <p
                      class="line-clamp-2 wrap-break-word text-sm font-medium text-foreground"
                    >
                      {{
                        subject.subjectName ||
                        subject.subjectCode ||
                        "Mata pelajaran"
                      }}
                    </p>
                    <p class="mt-1 truncate text-xs leading-5 text-muted">
                      {{ subject.teacherName || "Guru belum tersedia" }}
                    </p>
                  </div>
                  <PhArrowRight
                    :size="15"
                    class="mt-1 shrink-0 text-muted transition group-hover:translate-x-0.5 group-hover:text-brand"
                  />
                </div>
              </RouterLink>
            </div>

            <div v-else class="rounded-lg bg-surface-subtle p-5">
              <p class="text-sm font-medium text-foreground">
                Belum ada mata pelajaran
              </p>
              <p class="mt-2 text-sm leading-6 text-muted">
                Mata pelajaran akan tampil setelah tersedia pada kelas aktif.
              </p>
            </div>
          </article>

          <article
            class="flex min-h-90 min-w-0 flex-col rounded-xl border border-border bg-surface shadow-sm p-4 sm:p-5 lg:min-h-0 lg:flex-1 lg:overflow-hidden"
          >
            <div class="mb-4 flex shrink-0 items-center justify-between gap-3">
              <div class="flex min-w-0 items-start gap-3">
                <div class="min-w-0">
                  <p class="text-sm font-medium text-foreground">Tugas Saya</p>
                  <p class="mt-1 text-xs text-muted">
                    Prioritas tugas dari semua mata pelajaran.
                  </p>
                </div>
              </div>
              <RouterLink
                to="/student/assignments"
                class="shrink-0 text-xs font-medium text-brand transition hover:text-brand-hover sm:text-sm"
              >
                Lihat semua
              </RouterLink>
            </div>

            <div v-if="assignmentsLoading" class="shrink-0 space-y-2">
              <div
                v-for="item in 3"
                :key="item"
                class="h-16 animate-pulse rounded-lg bg-surface-subtle"
              />
            </div>

            <div
              v-else-if="assignmentsError"
              class="shrink-0 rounded-lg bg-surface-subtle p-4 text-sm leading-6 text-muted"
            >
              {{ assignmentsError }}
            </div>

            <div
              v-else-if="assignmentPreview.length === 0"
              class="shrink-0 rounded-lg bg-surface-subtle p-4"
            >
              <p class="text-sm font-semibold text-foreground">
                Belum ada tugas
              </p>
              <p class="mt-1 text-sm leading-6 text-muted">
                Tugas akan muncul setelah guru membuat tugas untuk mata
                pelajaran di kelasmu.
              </p>
            </div>

            <div
              v-else
              class="min-h-0 flex-1 divide-y divide-border overflow-y-auto pr-1"
            >
              <RouterLink
                v-for="assignment in assignmentPreview"
                :key="`${assignment.subjectClassId}-${assignment.assignmentId}`"
                :to="`/student/subjects/${assignment.subjectClassId}/assignments/${assignment.assignmentId}`"
                class="group block min-w-0 py-3 first:pt-0 last:pb-0 border-b border-border transition hover:bg-background"
              >
                <div class="flex min-w-0 items-start justify-between gap-3">
                  <div class="min-w-0 flex-1">
                    <p
                      class="truncate text-sm font-medium text-foreground transition group-hover:text-brand"
                    >
                      {{ assignment.assignmentTitle }}
                    </p>
                    <p class="mt-1 truncate text-xs text-muted">
                      {{ assignment.subjectName }}
                      <span v-if="assignment.subjectCode">
                        · {{ assignment.subjectCode }}
                      </span>
                    </p>
                    <p class="mt-1.5 text-[11px] text-muted mb-2">
                      Tenggat {{ formatDate(assignment.deadline) }}
                    </p>
                  </div>
                  <div class="flex shrink-0 flex-col items-end gap-2">
                    <span
                      class="rounded-full px-2.5 py-1 text-[10px] font-medium"
                      :class="assignmentStatusClasses(assignment)"
                    >
                      {{ assignmentStatusLabel(assignment) }}
                    </span>
                    <PhArrowRight
                      :size="14"
                      class="text-muted transition group-hover:translate-x-0.5 group-hover:text-brand"
                    />
                  </div>
                </div>
              </RouterLink>
            </div>
          </article>
        </section>
      </div>
    </section>

    <aside
      class="min-w-0 border-t border-border bg-surface lg:sticky lg:top-0 lg:h-dvh lg:min-h-0 lg:overflow-hidden lg:border-l lg:border-t-0"
    >
      <div
        class="flex flex-col gap-3 p-4 lg:h-full lg:min-h-0 lg:overflow-hidden"
      >
        <DashboardUpdatesPanel
          class="lg:min-h-0 lg:flex-1 lg:overflow-hidden"
          :notification-badge="notificationUnread.unreadCount.value"
          :chat-badge="chatPanelUnreadCount"
          :feed-badge="feedPanelUnreadCount"
        >
          <template #notifications>
            <NotificationsPanel embedded to="/student/notifications" />
          </template>

          <template #chat>
            <LatestChatCard
              to="/student/chat"
              :limit="4"
              embedded
              @unread-change="updateChatPanelUnreadCount"
            />
          </template>

          <template #feed>
            <div class="mb-3 flex items-center justify-between gap-3">
              <RouterLink
                class="shrink-0 text-xs font-medium text-brand transition hover:text-brand-hover inline-flex gap-1 pt-1"
                to="/student/feed"
              >
                Buka feed
                <PhArrowRight :size="14" />
              </RouterLink>
            </div>

            <div v-if="feedPreviewLoading" class="space-y-2">
              <div
                v-for="item in 3"
                :key="item"
                class="h-16 animate-pulse rounded-lg bg-surface-strong"
              />
            </div>
            <div
              v-else-if="feedPreviewError"
              class="rounded-lg bg-surface-subtle p-4 text-sm leading-6 text-muted"
            >
              <p>{{ feedPreviewError }}</p>
              <button
                v-if="activeClassStore.activeClassId"
                type="button"
                class="mt-3 inline-flex items-center gap-1.5 rounded-lg border border-border bg-surface px-3 py-1.5 text-xs font-medium text-brand transition hover:border-brand hover:bg-brand-soft disabled:cursor-not-allowed disabled:opacity-60"
                :disabled="feedPreviewLoading"
                @click="retryFeedPreview"
              >
                <PhArrowClockwise :size="14" />
                Coba lagi
              </button>
            </div>
            <div
              v-else-if="feedPosts.length > 0"
              class="divide-y divide-border"
            >
              <article
                v-for="post in feedPosts.slice(0, 5)"
                :key="post.feedId"
                class="min-w-0 py-3 first:pt-0 last:pb-0"
              >
                <p class="line-clamp-2 text-sm leading-6 text-foreground">
                  {{ post.content }}
                </p>
                <p class="mt-1 text-xs text-muted">
                  {{ post.creatorName || "Pengirim tidak tersedia" }} ·
                  {{ formatDateTime(post.createdAt) }}
                </p>
              </article>
            </div>
            <div v-else class="rounded-lg bg-surface-subtle p-4">
              <p class="text-sm font-semibold text-foreground">
                Belum ada pengumuman
              </p>
              <p class="mt-1 text-sm leading-6 text-muted">
                Pengumuman untuk kelas aktif akan tampil di sini.
              </p>
            </div>
          </template>
        </DashboardUpdatesPanel>

        <ActivityCalendarCard role="student" />
      </div>
    </aside>
  </main>
</template>
