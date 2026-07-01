<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink, useRouter } from "vue-router";
import {
  PhArrowRight,
  PhBookOpen,
  PhCaretLeft,
  PhCaretRight,
  PhClipboardText,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import { useActiveClassStore } from "../../stores/activeClass";
import { getSubjectClassesByClass } from "../../services/classWorkspace";
import { getClassFeed } from "../../services/feed";
import { getStudentAssignmentInbox } from "../../services/assignment";
import { useFeedUnreadCount } from "../../composables/useFeedUnreadCount";
import {
  getRecentNotifications,
  getUnreadNotificationCount,
  markAllNotificationsAsRead,
  markNotificationAsRead,
} from "../../services/studentDashboard";
import { getAcademicActivities } from "../../services/activity";
import type { SubjectClassItem } from "../../types/classWorkspace";
import type { FeedPost } from "../../types/feed";
import type { NotificationItem } from "../../types/dashboard";
import type { StudentAssignmentInboxItem } from "../../types/assignment";
import type { AcademicActivityItem } from "../../types/activity";
import {
  formatDate,
  formatDateTime,
  parseBackendTimestamp,
} from "../../utils/date";
import { getSubjectColor, resolveSubjectColor } from "../../utils/color";
import { useToastStore } from "../../stores/toast";
import LatestChatCard from "../../components/chat/LatestChatCard.vue";
import AcademicActivityCard from "../../components/activity/AcademicActivityCard.vue";
import DashboardUpdatesPanel from "../../components/dashboard/DashboardUpdatesPanel.vue";
import {
  activitySubjectColor,
  activityTypeLabel,
  compareActivities,
  formatActivityDate,
  formatApiDate,
  isInternalActivityLink,
  parseActivityDate,
} from "../../components/activity/activityView";

const auth = useAuthStore();
const activeClassStore = useActiveClassStore();
const toast = useToastStore();
const router = useRouter();
const { unreadCount: feedPanelUnreadCount } = useFeedUnreadCount();

const subjects = ref<SubjectClassItem[]>([]);
const feedPosts = ref<FeedPost[]>([]);
const assignmentPreviewItems = ref<StudentAssignmentInboxItem[]>([]);
const notifications = ref<NotificationItem[]>([]);
const unreadCount = ref(0);
const isLoading = ref(true);
const assignmentsLoading = ref(false);
const assignmentsError = ref("");
const activities = ref<AcademicActivityItem[]>([]);
const activitiesLoading = ref(false);
const activitiesError = ref("");
const calendarActivityCache = ref<Record<string, AcademicActivityItem[]>>({});
const calendarActivitiesLoading = ref(false);
const calendarActivitiesError = ref("");
const selectedDate = ref(formatApiDate(new Date()));
const chatPanelUnreadCount = ref(0);
const markingNotificationIds = ref<Set<string>>(new Set());
const markingAllNotifications = ref(false);
const errorMessage = ref("");
const viewDate = ref(new Date());

const activeMembership = computed(() => auth.activeMembership);
const schoolUserId = computed(() => auth.activeSchoolUserId);
const schoolName = computed(
  () => activeMembership.value?.school.name ?? "Wiyata",
);
const firstName = computed(() => auth.user?.fullName?.split(" ")[0] ?? "Siswa");
const activeClassTitle = computed(
  () =>
    activeClassStore.activeClassTitle ||
    activeClassStore.activeClass?.classTitle ||
    "",
);
const currentMonth = computed(() =>
  new Intl.DateTimeFormat("id-ID", { month: "long", year: "numeric" }).format(
    viewDate.value,
  ),
);
const calendarDays = computed(() => buildCalendarDays(viewDate.value));
const currentMonthKey = computed(() => monthKey(viewDate.value));
const calendarActivities = computed(
  () => calendarActivityCache.value[currentMonthKey.value] ?? [],
);
const selectedDateActivities = computed(() =>
  calendarActivities.value
    .filter((item) => item.date === selectedDate.value)
    .sort(compareActivities),
);
const selectedDateDeadlineActivities = computed(() =>
  selectedDateActivities.value.filter((item) => item.type === "assignment_due"),
);
const selectedDatePreview = computed(() =>
  selectedDateDeadlineActivities.value.slice(0, 3),
);
const assignmentPreview = computed(() =>
  [...assignmentPreviewItems.value].sort(compareAssignments),
);

function changeMonth(step: number) {
  const newDate = new Date(viewDate.value);
  newDate.setMonth(newDate.getMonth() + step);
  viewDate.value = newDate;
  selectedDate.value = defaultSelectedDateForMonth(newDate);
  loadCalendarActivities();
}

async function loadDashboard(selectedClassId?: string) {
  if (!auth.user?.id) {
    errorMessage.value = "Sesi login belum lengkap. Silakan login ulang.";
    isLoading.value = false;
    return;
  }

  if (!schoolUserId.value) {
    errorMessage.value = "Konteks sekolah belum tersedia.";
    isLoading.value = false;
    return;
  }

  isLoading.value = true;
  errorMessage.value = "";

  try {
    await activeClassStore.loadClasses(schoolUserId.value);

    const activeClassId = selectedClassId ?? activeClassStore.activeClassId;
    const [notificationData, unreadData] = await Promise.all([
      getRecentNotifications(),
      getUnreadNotificationCount(),
    ]);

    notifications.value = notificationData.data ?? [];
    unreadCount.value =
      unreadData.unreadCount ?? notificationData.unreadCount ?? 0;
    await loadAssignmentPreview();

    if (!activeClassId) {
      subjects.value = [];
      feedPosts.value = [];
      return;
    }

    const [subjectData, feedData] = await Promise.all([
      getSubjectClassesByClass(activeClassId),
      getClassFeed(activeClassId),
    ]);

    subjects.value = subjectData.subjects ?? [];
    feedPosts.value = feedData.data.data ?? [];
  } catch {
    errorMessage.value =
      "Dashboard belum bisa dimuat. Periksa koneksi atau coba lagi nanti.";
  } finally {
    isLoading.value = false;
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

async function loadCalendarActivities() {
  const key = currentMonthKey.value;
  if (calendarActivityCache.value[key]) {
    calendarActivitiesError.value = "";
    return;
  }

  calendarActivitiesLoading.value = true;
  calendarActivitiesError.value = "";

  const from = new Date(
    viewDate.value.getFullYear(),
    viewDate.value.getMonth(),
    1,
  );
  const to = new Date(
    viewDate.value.getFullYear(),
    viewDate.value.getMonth() + 1,
    0,
  );

  try {
    const response = await getAcademicActivities({
      from: formatApiDate(from),
      to: formatApiDate(to),
    });
    calendarActivityCache.value = {
      ...calendarActivityCache.value,
      [key]: response.items ?? [],
    };
  } catch {
    calendarActivitiesError.value = "Tidak dapat memuat deadline tugas.";
  } finally {
    calendarActivitiesLoading.value = false;
  }
}

function initials(value: string) {
  return value
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0])
    .join("")
    .toUpperCase();
}

function buildCalendarDays(date: Date) {
  const year = date.getFullYear();
  const month = date.getMonth();
  const firstDay = new Date(year, month, 1);
  const startOffset = firstDay.getDay();
  const daysInMonth = new Date(year, month + 1, 0).getDate();
  const days = [];

  const realToday = new Date();
  const isCurrentMonth =
    realToday.getMonth() === month && realToday.getFullYear() === year;

  for (let i = 0; i < startOffset; i += 1) {
    days.push({
      key: `empty-${i}`,
      label: "",
      isToday: false,
      dateKey: "",
      activities: [] as AcademicActivityItem[],
      extraCount: 0,
    });
  }

  for (let day = 1; day <= daysInMonth; day += 1) {
    const dateKey = formatApiDate(new Date(year, month, day));
    const dayActivities = activitiesForDate(dateKey);
    days.push({
      key: String(day),
      label: String(day),
      isToday: isCurrentMonth && day === realToday.getDate(),
      dateKey,
      activities: dayActivities.slice(0, 3),
      extraCount: Math.max(0, dayActivities.length - 3),
    });
  }

  return days;
}

function activitiesForDate(dateKey: string) {
  return calendarActivities.value
    .filter((item) => item.date === dateKey)
    .filter((item) => item.type === "assignment_due")
    .sort(compareActivities);
}

function monthKey(date: Date) {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}`;
}

function defaultSelectedDateForMonth(date: Date) {
  const today = new Date();
  if (
    today.getFullYear() === date.getFullYear() &&
    today.getMonth() === date.getMonth()
  ) {
    return formatApiDate(today);
  }
  return formatApiDate(new Date(date.getFullYear(), date.getMonth(), 1));
}

function selectCalendarDate(dateKey: string) {
  if (!dateKey) return;
  selectedDate.value = dateKey;
}

function calendarDateAriaLabel(day: {
  label: string;
  dateKey: string;
  activities: AcademicActivityItem[];
  extraCount: number;
}) {
  if (!day.dateKey) return "Tanggal kosong";
  const date = parseActivityDate(day.dateKey);
  const label = date
    ? new Intl.DateTimeFormat("id-ID", {
        day: "numeric",
        month: "long",
        year: "numeric",
      }).format(date)
    : day.label;
  const count = day.activities.length + day.extraCount;
  return `${label}, ${count} aktivitas`;
}

function calendarActivityTime(activity: AcademicActivityItem) {
  return activity.time || "Sepanjang hari";
}

function updateChatPanelUnreadCount(count: number) {
  chatPanelUnreadCount.value = Math.max(0, count);
}

function compareAssignments(
  a: StudentAssignmentInboxItem,
  b: StudentAssignmentInboxItem,
) {
  const overdueNotSubmittedDiff =
    Number(b.isOverdue && !b.isSubmitted) -
    Number(a.isOverdue && !a.isSubmitted);
  if (overdueNotSubmittedDiff !== 0) return overdueNotSubmittedDiff;

  const notSubmittedDiff = Number(!b.isSubmitted) - Number(!a.isSubmitted);
  if (notSubmittedDiff !== 0) return notSubmittedDiff;

  const deadlineDiff =
    getDeadlineTime(a.deadline) - getDeadlineTime(b.deadline);
  if (deadlineDiff !== 0) return deadlineDiff;

  return (a.assignmentTitle || "").localeCompare(b.assignmentTitle || "");
}

function getDeadlineTime(deadline?: string | null) {
  if (!deadline) return Number.MAX_SAFE_INTEGER;
  const value = parseBackendTimestamp(deadline)?.getTime() ?? Number.NaN;
  return Number.isNaN(value) ? Number.MAX_SAFE_INTEGER : value;
}

function assignmentStatusLabel(item: StudentAssignmentInboxItem) {
  if (item.isGraded) return "Sudah dinilai";
  if (item.isSubmitted) return "Sudah dikumpulkan";
  if (item.isOverdue) return "Lewat deadline";
  return "Belum dikumpulkan";
}

function assignmentStatusClasses(item: StudentAssignmentInboxItem) {
  if (item.isGraded) return "bg-[#ecfdf3] text-[#027a48]";
  if (item.isSubmitted) return "bg-[#eef2ff] text-[#4f46e5]";
  if (item.isOverdue) return "bg-[#fef2f2] text-[#dc2626]";
  return "bg-[#fff7ed] text-[#b45309]";
}

function notificationErrorMessage(error: unknown) {
  if (
    typeof error === "object" &&
    error !== null &&
    "response" in error &&
    typeof (error as { response?: { data?: { error?: unknown } } }).response
      ?.data?.error === "string"
  ) {
    return (error as { response: { data: { error: string } } }).response.data
      .error;
  }

  return "Status notifikasi belum bisa diperbarui.";
}

function notificationTitle(item: NotificationItem) {
  if (item.type === "assignment_created") return "Tugas baru";
  if (item.type === "feed_posted") return "Pengumuman kelas baru";
  if (item.type === "assignment_graded") return "Tugas sudah dinilai";
  if (item.type === "material_added") return "Materi baru";
  return item.title || "Notifikasi";
}

function notificationBadge(item: NotificationItem) {
  if (item.type === "assignment_created") return "TB";
  if (item.type === "feed_posted") return "PG";
  if (item.type === "assignment_graded") return "AG";
  if (item.type === "material_added") return "MT";
  return initials(notificationTitle(item));
}

function notificationMessage(item: NotificationItem) {
  return item.message || "Buka notifikasi untuk melihat informasi terbaru.";
}

function notificationAriaLabel(item: NotificationItem) {
  const action = item.link ? "Buka notifikasi" : "Tandai notifikasi dibaca";
  return `${action}: ${notificationTitle(item)}`;
}

function isInternalNotificationLink(link?: string) {
  return Boolean(link && link.startsWith("/") && !link.startsWith("//"));
}

async function markNotificationRead(item: NotificationItem) {
  if (markingNotificationIds.value.has(item.notificationId)) {
    return false;
  }

  if (item.isRead) {
    return true;
  }

  markingNotificationIds.value = new Set([
    ...markingNotificationIds.value,
    item.notificationId,
  ]);

  try {
    await markNotificationAsRead(item.notificationId);
    notifications.value = notifications.value.map((notification) =>
      notification.notificationId === item.notificationId
        ? { ...notification, isRead: true }
        : notification,
    );
    unreadCount.value = Math.max(0, unreadCount.value - 1);
    return true;
  } catch (error) {
    toast.error(notificationErrorMessage(error));
    return false;
  } finally {
    const next = new Set(markingNotificationIds.value);
    next.delete(item.notificationId);
    markingNotificationIds.value = next;
  }
}

async function handleNotificationClick(item: NotificationItem) {
  if (markingNotificationIds.value.has(item.notificationId)) {
    return;
  }

  const didMark = await markNotificationRead(item);
  if (didMark && isInternalNotificationLink(item.link)) {
    await router.push(item.link as string);
  }
}

async function markAllNotificationsRead() {
  if (unreadCount.value <= 0 || markingAllNotifications.value) return;

  markingAllNotifications.value = true;

  try {
    await markAllNotificationsAsRead();
    notifications.value = notifications.value.map((notification) => ({
      ...notification,
      isRead: true,
    }));
    unreadCount.value = 0;
    toast.success("Semua notifikasi ditandai sudah dibaca.");
  } catch (error) {
    toast.error(notificationErrorMessage(error));
  } finally {
    markingAllNotifications.value = false;
  }
}

onMounted(() => {
  loadDashboard();
  loadActivities();
  loadCalendarActivities();
});
</script>

<template>
  <main
    class="grid min-h-screen min-w-0 flex-1 grid-cols-1 overflow-x-hidden bg-[#f8f7f4] lg:grid-cols-[minmax(0,1fr)_320px]"
  >
    <section class="min-w-0">
      <header class="border-b border-[#ebe7df] bg-white">
        <div
          class="flex min-w-0 flex-col gap-3 px-5 py-4 sm:px-6 lg:flex-row lg:items-center lg:justify-between lg:px-8"
        >
          <div class="min-w-0">
            <h1 class="text-xl font-medium text-[#171322] sm:text-2xl">
              Selamat datang, {{ firstName }}
            </h1>
            <p
              class="mt-1 max-w-2xl text-xs leading-5 text-[#6b7280] sm:text-sm"
            >
              Mau belajar apa hari ini?
            </p>
          </div>
          <div
            class="inline-flex min-w-0 max-w-full items-center self-start rounded-lg border border-[#ebe7df] bg-[#f9fafb] px-3 py-2 text-xs text-[#6b7280] lg:self-auto"
          >
            <span class="min-w-0 truncate font-medium text-[#171322]">
              {{ schoolName }}
            </span>
            <span class="mx-2 shrink-0 text-[#d1d5db]">·</span>
            <span class="min-w-0 truncate">
              {{
                activeClassTitle
                  ? `Kelas ${activeClassTitle}`
                  : "Tanpa kelas aktif"
              }}
            </span>
          </div>
        </div>
      </header>

      <div class="space-y-5 px-5 py-5 sm:px-6 lg:px-8 lg:py-6">
        <AcademicActivityCard
          :activities="activities"
          :loading="activitiesLoading"
          :error="activitiesError"
          role="student"
          :max-items="5"
        />

        <section
          v-if="errorMessage"
          class="rounded-xl border border-[#f1d6d3] bg-white p-5 sm:p-6"
        >
          <div class="flex items-start gap-3">
            <div
              class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-[#fff1f0] text-[#dc2626]"
            >
              <PhWarningCircle :size="22" weight="duotone" />
            </div>
            <div class="min-w-0">
              <h2 class="text-base font-medium text-[#171322]">
                Dashboard tidak dapat dimuat
              </h2>
              <p class="mt-1 text-sm leading-6 text-[#7a7385]">
                {{ errorMessage }}
              </p>
              <button
                class="mt-4 rounded-lg bg-[#4f46e5] px-4 py-2 text-sm font-medium text-white transition hover:bg-[#4338ca]"
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
              class="h-36 animate-pulse rounded-xl border border-[#ebe7df] bg-white"
            />
          </div>
          <div class="grid gap-4 xl:grid-cols-2">
            <div
              v-for="item in 2"
              :key="`panel-${item}`"
              class="h-64 animate-pulse rounded-xl border border-[#ebe7df] bg-white"
            />
          </div>
        </section>

        <section
          v-else-if="!activeClassStore.activeClassId"
          class="flex min-h-[50vh] items-center justify-center"
        >
          <article
            class="w-full max-w-xl rounded-xl border border-[#ebe7df] bg-white p-8 text-center"
          >
            <div
              class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-[#eef2ff] text-[#4f46e5]"
            >
              <PhBookOpen :size="25" weight="duotone" />
            </div>
            <h2 class="mt-3 text-lg font-medium text-[#171322]">
              Belum ada kelas aktif
            </h2>
            <p class="mx-auto mt-2 max-w-md text-sm leading-6 text-[#7a7385]">
              Kelas akan muncul setelah akunmu terdaftar sebagai anggota kelas
              di sekolah aktif.
            </p>
          </article>
        </section>

        <section v-else class="grid min-w-0 gap-4 xl:grid-cols-[1.15fr_0.85fr]">
          <article
            class="min-w-0 rounded-xl border border-[#ebe7df] bg-white p-4 sm:p-5"
          >
            <div class="mb-4 flex min-w-0 items-center justify-between gap-3">
              <div class="min-w-0">
                <p class="text-sm font-medium text-[#171322]">
                  Daftar mata pelajaran
                </p>
                <p class="mt-1 text-xs text-[#8b8592]">
                  Buka materi dan tugas dari kelas aktif.
                </p>
              </div>
              <RouterLink
                class="shrink-0 text-xs font-medium text-[#4f46e5] transition hover:text-[#4338ca] sm:text-sm"
                to="/student/subjects"
              >
                Lihat semua
              </RouterLink>
            </div>

            <div v-if="subjects.length > 0" class="grid gap-3 sm:grid-cols-2">
              <RouterLink
                v-for="subject in subjects.slice(0, 4)"
                :key="subject.subjectClassId"
                class="group min-w-0 overflow-hidden rounded-lg border border-[#ebe7df] bg-[#fbfaf8] transition hover:-translate-y-0.5 hover:bg-white hover:shadow-[0_18px_40px_rgba(66,55,40,0.08)]"
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
                      class="line-clamp-2 wrap-break-word text-sm font-medium text-[#171322]"
                    >
                      {{
                        subject.subjectName ||
                        subject.subjectCode ||
                        "Mata pelajaran"
                      }}
                    </p>
                    <p class="mt-1 truncate text-xs leading-5 text-[#7a7385]">
                      {{ subject.teacherName || "Guru belum tersedia" }}
                    </p>
                  </div>
                  <PhArrowRight
                    :size="15"
                    class="mt-1 shrink-0 text-[#a09aa8] transition group-hover:translate-x-0.5 group-hover:text-[#4f46e5]"
                  />
                </div>
              </RouterLink>
            </div>

            <div
              v-else
              class="rounded-lg border border-[#ebe7df] bg-[#fbfaf8] p-5"
            >
              <p class="text-sm font-medium text-[#171322]">
                Belum ada mata pelajaran
              </p>
              <p class="mt-2 text-sm leading-6 text-[#7a7385]">
                Mata pelajaran akan tampil setelah tersedia pada kelas aktif.
              </p>
            </div>
          </article>

          <article
            class="flex min-h-90 min-w-0 flex-col rounded-xl border border-[#ebe7df] bg-white p-4 sm:p-5 lg:max-h-130"
          >
            <div class="mb-4 flex shrink-0 items-center justify-between gap-3">
              <div class="flex min-w-0 items-start gap-3">
                <div
                  class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-[#eef2ff] text-[#4f46e5]"
                >
                  <PhClipboardText :size="19" weight="duotone" />
                </div>
                <div class="min-w-0">
                  <p class="text-sm font-medium text-[#171322]">Tugas Saya</p>
                  <p class="mt-1 text-xs text-[#8b8592]">
                    Prioritas tugas dari semua mata pelajaran.
                  </p>
                </div>
              </div>
              <RouterLink
                to="/student/assignments"
                class="shrink-0 text-xs font-medium text-[#4f46e5] transition hover:text-[#4338ca] sm:text-sm"
              >
                Lihat semua
              </RouterLink>
            </div>

            <div v-if="assignmentsLoading" class="shrink-0 space-y-2">
              <div
                v-for="item in 3"
                :key="item"
                class="h-16 animate-pulse rounded-lg bg-[#fbfaf8]"
              />
            </div>

            <div
              v-else-if="assignmentsError"
              class="shrink-0 rounded-lg border border-[#ebe7df] bg-[#fbfaf8] p-4 text-sm leading-6 text-[#7a7385]"
            >
              {{ assignmentsError }}
            </div>

            <div
              v-else-if="assignmentPreview.length === 0"
              class="shrink-0 rounded-lg border border-[#ebe7df] bg-[#fbfaf8] p-4"
            >
              <p class="text-sm font-medium text-[#171322]">Belum ada tugas</p>
              <p class="mt-2 text-sm leading-6 text-[#7a7385]">
                Tugas akan muncul setelah guru membuat tugas untuk mata
                pelajaran di kelasmu.
              </p>
            </div>

            <div
              v-else
              class="min-h-0 flex-1 divide-y divide-[#ebe7df] overflow-y-auto pr-1"
            >
              <RouterLink
                v-for="assignment in assignmentPreview"
                :key="`${assignment.subjectClassId}-${assignment.assignmentId}`"
                :to="`/student/subjects/${assignment.subjectClassId}/assignments/${assignment.assignmentId}`"
                class="group block min-w-0 py-3 first:pt-0 last:pb-0"
              >
                <div class="flex min-w-0 items-start justify-between gap-3">
                  <div class="min-w-0 flex-1">
                    <p
                      class="truncate text-sm font-medium text-[#171322] transition group-hover:text-[#4f46e5]"
                    >
                      {{ assignment.assignmentTitle }}
                    </p>
                    <p class="mt-1 truncate text-xs text-[#7a7385]">
                      {{ assignment.subjectName }}
                      <span v-if="assignment.subjectCode">
                        · {{ assignment.subjectCode }}
                      </span>
                    </p>
                    <p class="mt-1.5 text-[11px] text-[#8b8592]">
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
                      class="text-[#a09aa8] transition group-hover:translate-x-0.5 group-hover:text-[#4f46e5]"
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
      class="min-w-0 border-t border-[#ebe7df] bg-white lg:sticky lg:top-0 lg:h-screen lg:min-h-0 lg:overflow-hidden lg:border-l lg:border-t-0"
    >
      <div class="flex flex-col gap-4 p-4 lg:h-full lg:min-h-0 lg:overflow-hidden">
        <DashboardUpdatesPanel
          class="lg:min-h-0 lg:flex-1 lg:overflow-hidden"
          :notification-badge="unreadCount"
          :chat-badge="chatPanelUnreadCount"
          :feed-badge="feedPanelUnreadCount"
        >
          <template #notifications>
            <div class="mb-3 flex items-center justify-between gap-3">
              <p class="text-xs text-[#a09aa8]">
                {{ unreadCount }} belum dibaca
              </p>
              <button
                v-if="unreadCount > 0"
                class="rounded-lg bg-[#eef2ff] px-3 py-1 text-xs font-medium text-[#4f46e5] transition hover:bg-[#e0e7ff] disabled:cursor-not-allowed disabled:opacity-60"
                type="button"
                :disabled="markingAllNotifications"
                @click="markAllNotificationsRead"
              >
                {{
                  markingAllNotifications
                    ? "Menyimpan..."
                    : "Tandai semua dibaca"
                }}
              </button>
            </div>

            <div v-if="isLoading" class="space-y-2">
              <div
                v-for="item in 3"
                :key="item"
                class="h-16 animate-pulse rounded-lg bg-[#f0ede8]"
              />
            </div>
            <div v-else-if="notifications.length > 0" class="space-y-1">
              <button
                v-for="item in notifications"
                :key="item.notificationId"
                class="flex min-w-0 w-full gap-3 p-3 text-left transition hover:bg-[#f8f7f4] disabled:cursor-wait disabled:opacity-75 border-b border-[#ebe7df]"
                :class="!item.isRead ? 'bg-[#f5f7ff]' : ''"
                type="button"
                :disabled="markingNotificationIds.has(item.notificationId)"
                :aria-label="notificationAriaLabel(item)"
                @click="handleNotificationClick(item)"
              >
                <div
                  class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full text-[11px] font-medium text-white"
                  :style="{
                    backgroundColor: getSubjectColor(
                      item.type || item.notificationId,
                    ),
                  }"
                >
                  {{ notificationBadge(item) }}
                </div>
                <div class="min-w-0 flex-1">
                  <div class="flex items-baseline justify-between gap-2">
                    <p class="line-clamp-1 text-sm font-medium text-[#171322]">
                      {{ notificationTitle(item) }}
                    </p>
                    <span class="shrink-0 text-[10px] text-[#a09aa8]">{{
                      formatDateTime(item.createdAt)
                    }}</span>
                  </div>
                  <p class="line-clamp-2 text-xs leading-5 text-[#7a7385]">
                    {{ notificationMessage(item) }}
                  </p>
                  <span
                    v-if="!item.isRead"
                    class="mt-1 inline-flex rounded-full bg-[#4f46e5] px-2 py-0.5 text-[10px] font-medium text-white"
                  >
                    baru
                  </span>
                </div>
              </button>
            </div>
            <div
              v-else
              class="rounded-lg border border-[#ebe7df] bg-[#fbfaf8] p-4 text-sm text-[#7a7385]"
            >
              Belum ada notifikasi terbaru.
            </div>
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
                class="shrink-0 text-xs font-medium text-[#4f46e5] transition hover:text-[#4338ca] inline-flex gap-1 pt-1"
                to="/student/feed"
              >
                Buka feed
                <PhArrowRight :size="14" />
              </RouterLink>
            </div>

            <div v-if="isLoading" class="space-y-2">
              <div
                v-for="item in 3"
                :key="item"
                class="h-16 animate-pulse rounded-lg bg-[#f0ede8]"
              />
            </div>
            <div
              v-else-if="feedPosts.length > 0"
              class="divide-y divide-[#ebe7df]"
            >
              <article
                v-for="post in feedPosts.slice(0, 5)"
                :key="post.feedId"
                class="min-w-0 py-3 first:pt-0 last:pb-0"
              >
                <p class="line-clamp-2 text-sm leading-6 text-[#3f3a4a]">
                  {{ post.content }}
                </p>
                <p class="mt-1 text-xs text-[#a09aa8]">
                  {{ post.creatorName || "Pengirim tidak tersedia" }} ·
                  {{ formatDateTime(post.createdAt) }}
                </p>
              </article>
            </div>
            <div
              v-else
              class="rounded-lg border border-[#ebe7df] bg-[#fbfaf8] p-4 text-sm leading-6 text-[#7a7385]"
            >
              Belum ada pengumuman untuk kelas aktif.
            </div>
          </template>
        </DashboardUpdatesPanel>

        <section class="shrink-0 rounded-xl bg-white p-4">
          <div class="mb-4 flex items-center justify-between">
            <p class="text-sm font-medium text-[#171322]">{{ currentMonth }}</p>
            <div class="flex gap-1">
              <button
                class="rounded-lg border border-[#ebe7df] p-1.5 text-[#7a7385] transition hover:bg-[#fbfaf8]"
                type="button"
                @click="changeMonth(-1)"
              >
                <PhCaretLeft :size="14" />
              </button>
              <button
                class="rounded-lg border border-[#ebe7df] p-1.5 text-[#7a7385] transition hover:bg-[#fbfaf8]"
                type="button"
                @click="changeMonth(1)"
              >
                <PhCaretRight :size="14" />
              </button>
            </div>
          </div>

          <div class="grid grid-cols-7 gap-1 text-center">
            <span
              v-for="day in ['Min', 'Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab']"
              :key="day"
              class="py-1 text-[10px] text-[#a09aa8]"
            >
              {{ day }}
            </span>
            <span v-for="day in calendarDays" :key="day.key" class="min-h-10.5">
              <button
                v-if="day.dateKey"
                type="button"
                class="flex h-full min-h-10.5 w-full flex-col items-center justify-center rounded-lg px-1 py-1 text-xs transition focus:outline-none focus-visible:ring-2 focus-visible:ring-[#4f46e5] focus-visible:ring-offset-2"
                :class="[
                  day.isToday
                    ? 'bg-[#4f46e5] font-medium text-white'
                    : 'text-[#4a4356] hover:bg-[#fbfaf8]',
                  selectedDate === day.dateKey
                    ? 'ring-2 ring-[#4f46e5] ring-offset-1'
                    : '',
                ]"
                :aria-label="calendarDateAriaLabel(day)"
                :aria-pressed="selectedDate === day.dateKey"
                @click="selectCalendarDate(day.dateKey)"
              >
                <span>{{ day.label }}</span>
                <span
                  v-if="day.activities.length || day.extraCount"
                  class="mt-1 flex h-2 items-center justify-center gap-0.5"
                  aria-hidden="true"
                >
                  <span
                    v-for="activity in day.activities"
                    :key="activity.id"
                    class="h-1.5 w-1.5 rounded-full"
                    :style="{
                      backgroundColor: activity.subject
                        ? activitySubjectColor(activity)
                        : '#a09aa8',
                    }"
                  />
                  <span
                    v-if="day.extraCount"
                    class="ml-0.5 text-[9px] font-medium"
                    :class="day.isToday ? 'text-white' : 'text-[#7a7385]'"
                  >
                    +{{ day.extraCount }}
                  </span>
                </span>
              </button>
              <span v-else class="block min-h-10.5" />
            </span>
          </div>

          <div class="mt-5 border-t border-[#ebe7df] pt-4">
            <div class="mb-3 flex items-center justify-between gap-3">
              <p class="text-sm font-medium text-[#171322]">Deadline Tugas</p>
              <p class="shrink-0 text-xs text-[#8b8592]">
                {{ formatActivityDate(selectedDate) }}
              </p>
            </div>

            <div v-if="calendarActivitiesLoading" class="space-y-2">
              <div
                v-for="item in 3"
                :key="item"
                class="h-12 animate-pulse rounded-lg bg-[#fbfaf8]"
              />
            </div>

            <div
              v-else-if="calendarActivitiesError"
              class="rounded-lg border border-[#ebe7df] bg-[#fbfaf8] p-3 text-xs leading-5 text-[#7a7385]"
            >
              {{ calendarActivitiesError }}
            </div>

            <div
              v-else-if="selectedDatePreview.length === 0"
              class="border-b border-[#ebe7df] bg-[#fbfaf8] p-3 text-xs leading-5 text-[#7a7385]"
            >
              Tidak ada deadline tugas pada tanggal ini.
            </div>

            <ul
              v-else
              class="space-y-2"
              aria-label="Deadline tugas pada tanggal ini"
            >
              <li
                v-for="activity in selectedDatePreview"
                :key="activity.id"
                class="min-w-0"
              >
                <RouterLink
                  v-if="isInternalActivityLink(activity.link)"
                  :to="activity.link || ''"
                  class="group flex min-w-0 items-start gap-2 border-b border-[#ebe7df] bg-[#fbfaf8] p-3 transition hover:border-[#c7d2fe] hover:bg-white focus:outline-none focus-visible:ring-2 focus-visible:ring-[#4f46e5] focus-visible:ring-offset-2"
                  :aria-label="`${activityTypeLabel(activity.type, 'student')}: ${activity.title}`"
                >
                  <span
                    class="mt-1.5 h-2 w-2 shrink-0 rounded-full"
                    :style="{
                      backgroundColor: activity.subject
                        ? activitySubjectColor(activity)
                        : '#a09aa8',
                    }"
                    aria-hidden="true"
                  />
                  <span class="min-w-0 flex-1">
                    <span
                      class="flex min-w-0 items-center justify-between gap-2"
                    >
                      <span class="text-[11px] font-medium text-[#4f46e5]">
                        {{ activityTypeLabel(activity.type, "student") }}
                      </span>
                      <span class="shrink-0 text-[10px] text-[#9ca3af]">
                        {{ calendarActivityTime(activity) }}
                      </span>
                    </span>
                    <span
                      class="mt-1 block truncate text-xs font-medium text-[#171322] transition group-hover:text-[#4f46e5]"
                    >
                      {{ activity.title }}
                    </span>
                  </span>
                </RouterLink>

                <article
                  v-else
                  class="flex min-w-0 items-start gap-2 rounded-lg border border-[#ebe7df] bg-[#fbfaf8] p-3"
                >
                  <span
                    class="mt-1.5 h-2 w-2 shrink-0 rounded-full"
                    :style="{
                      backgroundColor: activity.subject
                        ? activitySubjectColor(activity)
                        : '#a09aa8',
                    }"
                    aria-hidden="true"
                  />
                  <div class="min-w-0 flex-1">
                    <div
                      class="flex min-w-0 items-center justify-between gap-2"
                    >
                      <span class="text-[11px] font-medium text-[#4f46e5]">
                        {{ activityTypeLabel(activity.type, "student") }}
                      </span>
                      <span class="shrink-0 text-[10px] text-[#9ca3af]">
                        {{ calendarActivityTime(activity) }}
                      </span>
                    </div>
                    <p class="mt-1 truncate text-xs font-medium text-[#171322]">
                      {{ activity.title }}
                    </p>
                  </div>
                </article>
              </li>
            </ul>

            <p
              v-if="
                selectedDateDeadlineActivities.length >
                selectedDatePreview.length
              "
              class="mt-3 text-xs text-[#8b8592]"
            >
              +{{
                selectedDateDeadlineActivities.length -
                selectedDatePreview.length
              }}
              deadline lainnya
            </p>
          </div>
        </section>
      </div>
    </aside>
  </main>
</template>
