<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import {
  PhArrowRight,
  PhBell,
  PhBookOpen,
  PhCaretLeft,
  PhCaretRight,
  PhChatCircleText,
  PhClipboardText,
  PhNotebook,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import { useActiveClassStore } from "../../stores/activeClass";
import { getSubjectClassesByClass } from "../../services/classWorkspace";
import { getClassFeed } from "../../services/feed";
import { getStudentAssignmentInbox } from "../../services/assignment";
import {
  getRecentNotifications,
  getUnreadNotificationCount,
} from "../../services/studentDashboard";
import type { SubjectClassItem } from "../../types/classWorkspace";
import type { FeedPost } from "../../types/feed";
import type { NotificationItem } from "../../types/dashboard";
import type { StudentAssignmentInboxItem } from "../../types/assignment";
import { formatDate, formatDateTime } from "../../utils/date";
import { getSubjectColor } from "../../utils/color";

const auth = useAuthStore();
const activeClassStore = useActiveClassStore();

const subjects = ref<SubjectClassItem[]>([]);
const feedPosts = ref<FeedPost[]>([]);
const assignmentPreviewItems = ref<StudentAssignmentInboxItem[]>([]);
const notifications = ref<NotificationItem[]>([]);
const unreadCount = ref(0);
const isLoading = ref(true);
const assignmentsLoading = ref(false);
const assignmentsError = ref("");
const errorMessage = ref("");
const viewDate = ref(new Date());

const activeMembership = computed(() => auth.activeMembership);
const schoolUserId = computed(() => auth.activeSchoolUserId);
const schoolName = computed(
  () => activeMembership.value?.school.name ?? "Eduverse",
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
const assignmentPreview = computed(() =>
  [...assignmentPreviewItems.value].sort(compareAssignments).slice(0, 4),
);

function changeMonth(step: number) {
  const newDate = new Date(viewDate.value);
  newDate.setMonth(newDate.getMonth() + step);
  viewDate.value = newDate;
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
    });
  }

  for (let day = 1; day <= daysInMonth; day += 1) {
    days.push({
      key: String(day),
      label: String(day),
      isToday: isCurrentMonth && day === realToday.getDate(),
    });
  }

  return days;
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

  const deadlineDiff = getDeadlineTime(a.deadline) - getDeadlineTime(b.deadline);
  if (deadlineDiff !== 0) return deadlineDiff;

  return (a.assignmentTitle || "").localeCompare(b.assignmentTitle || "");
}

function getDeadlineTime(deadline?: string | null) {
  if (!deadline) return Number.MAX_SAFE_INTEGER;
  const value = new Date(deadline).getTime();
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

onMounted(loadDashboard);
</script>

<template>
  <main
    class="grid min-h-screen flex-1 grid-cols-1 overflow-hidden lg:grid-cols-[1fr_320px]"
  >
    <section class="flex flex-col gap-6 px-5 py-6 sm:px-8 lg:px-10">
      <header class="flex flex-col gap-2">
        <p class="text-sm text-[#7a7385]">
          {{ schoolName }} -
          {{
            activeClassTitle ? `Kelas ${activeClassTitle}` : "Tanpa kelas aktif"
          }}
        </p>
        <h1 class="text-2xl font-medium tracking-normal text-[#171322]">
          Selamat datang, {{ firstName }}
        </h1>
        <p class="text-sm text-[#7a7385]">
          Ruang belajar hari ini mengikuti kelas aktif dan subject di dalamnya.
        </p>
      </header>

      <div v-if="errorMessage" class="soft-card rounded-[22px] p-5">
        <div class="flex items-start gap-3 text-sm text-[#b42318]">
          <PhWarningCircle :size="20" class="mt-0.5" weight="duotone" />
          <p>{{ errorMessage }}</p>
        </div>
      </div>

      <!-- <section class="soft-card rounded-[22px] p-5">
        <div
          class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between"
        >
          <select
            v-if="activeClassStore.classes.length > 1"
            class="rounded-2xl border border-[#ebe7df] bg-white px-4 py-2 text-sm text-[#3f3a4a] outline-none"
            :value="activeClassStore.activeClassId ?? ''"
            @change="handleActiveClassChange"
          >
            <option
              v-for="item in activeClassStore.classes"
              :key="item.classId"
              :value="item.classId"
            >
              {{ item.classTitle || item.classId }}
            </option>
          </select>
        </div>
      </section> -->

      <section
        v-if="isLoading"
        class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3"
      >
        <div
          v-for="item in 3"
          :key="item"
          class="h-36 animate-pulse rounded-[18px] border border-[#ebe7df] bg-white"
        />
      </section>

      <section
        v-else-if="!activeClassStore.activeClassId"
        class="soft-card rounded-[22px] p-6"
      >
        <p class="text-sm font-medium text-[#171322]">Belum ada kelas aktif</p>
        <p class="mt-2 text-sm text-[#7a7385]">
          Kelas akan muncul setelah akunmu terdaftar sebagai member kelas di
          sekolah aktif.
        </p>
      </section>

      <section v-else class="grid gap-4 xl:grid-cols-[1.15fr_0.85fr]">
        <article class="soft-card rounded-[22px] p-5 pl-0">
          <div class="mb-5 flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-[#171322]">List subject</p>
              <p class="mt-1 text-xs text-[#8b8592]">
                Buka subject untuk melihat materi dan tugas.
              </p>
            </div>
            <RouterLink
              class="text-sm font-medium text-[#4f46e5]"
              to="/student/subjects"
            >
              Lihat semua
            </RouterLink>
          </div>

          <div v-if="subjects.length > 0" class="grid gap-3 sm:grid-cols-2">
            <RouterLink
              v-for="subject in subjects.slice(0, 4)"
              :key="subject.subjectClassId"
              class="rounded-[18px] border border-[#ebe7df] bg-white p-4 transition hover:-translate-y-0.5 hover:shadow-[0_18px_40px_rgba(66,55,40,0.08)]"
              :to="`/student/subjects/${subject.subjectClassId}`"
            >
              <div
                class="mb-4 flex h-11 w-11 items-center justify-center rounded-2xl text-white"
                :style="{
                  backgroundColor: getSubjectColor(
                    subject.subjectClassId ||
                      subject.subjectName ||
                      subject.subjectCode,
                  ),
                }"
              >
                <PhBookOpen :size="21" weight="duotone" />
              </div>
              <p class="text-sm font-medium text-[#171322]">
                {{ subject.subjectName || subject.subjectCode || "Subject" }}
              </p>
              <p class="mt-2 text-xs leading-5 text-[#7a7385]">
                {{ subject.teacherName || "Guru belum tersedia" }}
              </p>
            </RouterLink>
          </div>

          <div v-else class="rounded-2xl bg-[#fbfaf8] p-5 pl-0">
            <p class="text-sm font-medium text-[#171322]">Belum ada subject</p>
            <p class="mt-2 text-sm leading-6 text-[#7a7385]">
              Subject akan tampil setelah kelas aktif memiliki subject class.
            </p>
          </div>
        </article>

        <article class="soft-card rounded-[22px] p-5 pl-0">
          <div class="mb-5 flex items-center justify-between gap-3">
            <div class="flex min-w-0 items-start gap-3">
              <div
                class="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-[#eef2ff] text-[#4f46e5]"
              >
                <PhClipboardText :size="21" weight="duotone" />
              </div>
              <div class="min-w-0">
                <p class="text-sm font-medium text-[#171322]">Tugas Saya</p>
                <p class="mt-1 text-xs text-[#8b8592]">
                  Lihat tugas dari subject yang kamu ikuti.
                </p>
              </div>
            </div>
            <RouterLink
              to="/student/assignments"
              class="shrink-0 text-sm font-medium text-[#4f46e5]"
            >
              Lihat semua
            </RouterLink>
          </div>

          <div v-if="assignmentsLoading" class="space-y-2">
            <div
              v-for="item in 3"
              :key="item"
              class="h-16 animate-pulse rounded-2xl bg-[#fbfaf8]"
            />
          </div>

          <div
            v-else-if="assignmentsError"
            class="rounded-2xl bg-[#fbfaf8] p-4 text-sm leading-6 text-[#7a7385]"
          >
            {{ assignmentsError }}
          </div>

          <div
            v-else-if="assignmentPreview.length === 0"
            class="rounded-2xl bg-[#fbfaf8] p-4"
          >
            <p class="text-sm font-medium text-[#171322]">Belum ada tugas</p>
            <p class="mt-2 text-sm leading-6 text-[#7a7385]">
              Tugas akan muncul setelah guru membuat tugas untuk subject di
              kelasmu.
            </p>
          </div>

          <div v-else class="space-y-2">
            <RouterLink
              v-for="assignment in assignmentPreview"
              :key="`${assignment.subjectClassId}-${assignment.assignmentId}`"
              :to="`/student/subjects/${assignment.subjectClassId}/assignments/${assignment.assignmentId}`"
              class="block rounded-2xl bg-[#fbfaf8] p-4 transition hover:bg-white hover:shadow-[0_12px_28px_rgba(66,55,40,0.08)]"
            >
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0">
                  <p class="truncate text-sm font-medium text-[#171322]">
                    {{ assignment.assignmentTitle }}
                  </p>
                  <p class="mt-1 truncate text-xs text-[#7a7385]">
                    {{ assignment.subjectName }}
                    <span v-if="assignment.subjectCode">
                      · {{ assignment.subjectCode }}
                    </span>
                  </p>
                </div>
                <span
                  class="shrink-0 rounded-full px-2.5 py-1 text-[10px] font-medium"
                  :class="assignmentStatusClasses(assignment)"
                >
                  {{ assignmentStatusLabel(assignment) }}
                </span>
              </div>
              <div
                class="mt-3 flex items-center justify-between gap-3 text-xs text-[#8b8592]"
              >
                <span>Deadline {{ formatDate(assignment.deadline) }}</span>
                <PhArrowRight :size="14" class="shrink-0 text-[#4f46e5]" />
              </div>
            </RouterLink>
          </div>
        </article>
      </section>

      <section class="grid gap-4 xl:grid-cols-[1.1fr_0.9fr]">
        <article class="soft-card rounded-[22px] p-5">
          <div class="mb-4 flex items-center justify-between">
            <p class="text-sm font-medium text-[#171322]">Feed kelas</p>
            <RouterLink
              class="text-sm font-medium text-[#4f46e5]"
              to="/student/feed"
            >
              Buka feed
            </RouterLink>
          </div>

          <div v-if="feedPosts.length > 0" class="space-y-3">
            <article
              v-for="post in feedPosts.slice(0, 3)"
              :key="post.feedId"
              class="rounded-2xl bg-[#fbfaf8] p-5 pl-0"
            >
              <p class="text-sm leading-6 text-[#3f3a4a]">{{ post.content }}</p>
              <p class="mt-2 text-xs text-[#a09aa8]">
                {{ post.creatorName || "Creator tidak tersedia" }} ·
                {{ formatDateTime(post.createdAt) }}
              </p>
            </article>
          </div>

          <p
            v-else
            class="rounded-2xl bg-[#fbfaf8] p-5 text-sm leading-6 text-[#7a7385]"
          >
            Belum ada posting feed untuk kelas aktif.
          </p>
        </article>

        <article class="soft-card rounded-[22px] p-5 pl-0 opacity-90">
          <p class="text-sm font-medium text-[#171322]">Fitur berikutnya</p>
          <p class="mt-1 text-xs text-[#8b8592]">
            Chat dan notes belum menjadi bagian dari flow utama saat ini.
          </p>
          <div class="mt-4 space-y-3">
            <div class="flex gap-3 rounded-2xl bg-[#eef2ff] p-4">
              <PhChatCircleText
                :size="20"
                class="mt-0.5 shrink-0 text-[#4f46e5]"
              />
              <p class="text-sm leading-6 text-[#6b6475]">
                Chat realtime masih ditunda dan belum menampilkan data
                percakapan.
              </p>
            </div>
            <div class="flex gap-3 rounded-2xl bg-[#f3ecff] p-4">
              <PhNotebook :size="20" class="mt-0.5 shrink-0 text-[#7c3aed]" />
              <p class="text-sm leading-6 text-[#6b6475]">
                Notes per materi dan autosave juga belum diimplementasikan.
              </p>
            </div>
          </div>
        </article>
      </section>
    </section>

    <aside class="border-l border-[#ebe7df] bg-white/95">
      <div class="flex border-b border-[#ebe7df] px-5">
        <button
          class="flex items-center gap-2 border-b-2 border-[#4f46e5] px-1 py-4 text-sm font-medium text-[#4f46e5]"
          type="button"
        >
          <PhBell :size="18" />
          Notifikasi
        </button>
        <button class="px-5 py-4 text-sm text-[#a09aa8]" type="button">
          {{ unreadCount }} belum dibaca
        </button>
      </div>

      <div v-if="isLoading" class="space-y-2 p-4">
        <div
          v-for="item in 3"
          :key="item"
          class="h-16 animate-pulse rounded-2xl bg-[#f0ede8]"
        />
      </div>
      <div v-else-if="notifications.length > 0" class="space-y-1 p-4">
        <article
          v-for="item in notifications"
          :key="item.notificationId"
          class="flex gap-3 rounded-2xl p-3 transition hover:bg-[#f8f7f4]"
          :class="!item.isRead ? 'bg-[#f5f7ff]' : ''"
        >
          <div
            class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full text-[11px] font-medium text-white"
            :style="{ backgroundColor: getSubjectColor(item.notificationId || item.title) }"
          >
            {{ initials(item.title) }}
          </div>
          <div class="min-w-0 flex-1">
            <div class="flex items-baseline justify-between gap-2">
              <p class="truncate text-sm font-medium text-[#171322]">
                {{ item.title }}
              </p>
              <span class="shrink-0 text-[10px] text-[#a09aa8]">{{
                formatDateTime(item.createdAt)
              }}</span>
            </div>
            <p class="truncate text-xs text-[#7a7385]">{{ item.message }}</p>
            <span
              v-if="!item.isRead"
              class="mt-1 inline-flex rounded-full bg-[#4f46e5] px-2 py-0.5 text-[10px] font-medium text-white"
            >
              baru
            </span>
          </div>
        </article>
      </div>
      <div v-else class="p-4">
        <div class="rounded-2xl bg-[#fbfaf8] p-4 text-sm text-[#7a7385]">
          Belum ada notifikasi terbaru.
        </div>
      </div>

      <section class="border-t border-[#ebe7df] p-5">
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
          <span
            v-for="day in calendarDays"
            :key="day.key"
            class="rounded-lg py-1.5 text-xs text-[#4a4356]"
            :class="day.isToday ? 'bg-[#4f46e5] font-medium text-white' : ''"
          >
            {{ day.label }}
          </span>
        </div>
      </section>
    </aside>
  </main>
</template>
