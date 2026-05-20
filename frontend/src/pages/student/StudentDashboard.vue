<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import {
  PhBell,
  PhCaretLeft,
  PhCaretRight,
  PhChatCircleText,
} from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import {
  getMemberClasses,
  getRecentNotifications,
  getStudentDashboard,
  getUnreadNotificationCount,
} from "../../services/studentDashboard";
import type {
  EnrollmentClass,
  NotificationItem,
  StudentDashboardSummary,
} from "../../types/dashboard";

const auth = useAuthStore();

const palette = ["#4f8ef7", "#f2756a", "#c673d8", "#f0a05a", "#4f46e5"];
const summary = ref<StudentDashboardSummary | null>(null);
const classes = ref<EnrollmentClass[]>([]);
const notifications = ref<NotificationItem[]>([]);
const unreadCount = ref(0);
const isLoading = ref(true);
const errorMessage = ref("");

const activeMembership = computed(() => {
  const activeSchoolId = auth.activeSchoolId;
  return (
    auth.memberships.find(
      (membership) => membership.school.id === activeSchoolId,
    ) ?? auth.memberships[0]
  );
});

const schoolName = computed(
  () => activeMembership.value?.school.name ?? "Eduverse",
);
const firstName = computed(() => auth.user?.fullName?.split(" ")[0] ?? "Siswa");
const materialProgress = computed(() => {
  if (!summary.value?.totalMaterials) return 0;
  return Math.round(
    (summary.value.completedMaterials / summary.value.totalMaterials) * 100,
  );
});
const currentMonth = computed(() =>
  new Intl.DateTimeFormat("id-ID", { month: "long", year: "numeric" }).format(
    new Date(),
  ),
);
const calendarDays = computed(() => buildCalendarDays(new Date()));

async function loadDashboard() {
  if (!auth.user?.id) {
    errorMessage.value = "Sesi login belum lengkap. Silakan login ulang.";
    isLoading.value = false;
    return;
  }

  isLoading.value = true;
  errorMessage.value = "";

  try {
    const schoolUserId =
      auth.defaultContext?.schoolUserId ?? activeMembership.value?.schoolUserId;
    const [dashboardData, notificationData, unreadData, classData] =
      await Promise.all([
        getStudentDashboard(auth.user.id),
        getRecentNotifications(),
        getUnreadNotificationCount(),
        schoolUserId ? getMemberClasses(schoolUserId) : Promise.resolve([]),
      ]);

    summary.value = dashboardData;
    notifications.value = notificationData.data ?? [];
    unreadCount.value =
      unreadData.unreadCount ?? notificationData.unreadCount ?? 0;
    classes.value = classData ?? [];
  } catch {
    errorMessage.value =
      "Dashboard belum bisa dimuat. Periksa koneksi atau coba lagi nanti.";
  } finally {
    isLoading.value = false;
  }
}

function classTitle(item: EnrollmentClass) {
  return item.subjectName || item.classTitle || item.classCode || "Kelas";
}

function classSubtitle(item: EnrollmentClass) {
  return [item.classCode, item.classTitle].filter(Boolean).join(" - ");
}

function nearestDeadlineFor(subjectName?: string) {
  const deadlines = summary.value?.upcomingDeadlines ?? [];
  if (!subjectName) return deadlines[0];
  return (
    deadlines.find((deadline) => deadline.subjectName === subjectName) ??
    deadlines[0]
  );
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

  for (let i = 0; i < startOffset; i += 1) {
    days.push({
      key: `empty-${i}`,
      label: "",
      isToday: false,
      hasEvent: false,
    });
  }

  const eventDays = new Set(
    (summary.value?.upcomingDeadlines ?? [])
      .map((deadline) => parseDeadlineDay(deadline.deadline))
      .filter((day): day is number => Boolean(day)),
  );

  for (let day = 1; day <= daysInMonth; day += 1) {
    days.push({
      key: String(day),
      label: String(day),
      isToday: day === date.getDate(),
      hasEvent: eventDays.has(day),
    });
  }

  return days;
}

function parseDeadlineDay(value: string) {
  const match = value.match(/^(\d{1,2})[-/\s]/);
  if (!match) return null;
  return Number(match[1]);
}

onMounted(loadDashboard);
</script>

<template>
  <main
    class="grid min-h-screen flex-1 grid-cols-1 overflow-hidden lg:grid-cols-[1fr_320px]"
  >
    <section class="flex flex-col gap-6 px-5 py-6 sm:px-8 lg:px-10">
      <header class="flex flex-col gap-2">
        <p class="text-sm text-[#7a7385]">{{ schoolName }}</p>
        <h1 class="text-2xl font-medium tracking-normal text-[#171322]">
          Selamat datang, {{ firstName }}
        </h1>
        <p class="text-sm text-[#7a7385]">
          Ringkasan akademik dari kelas, tugas, materi, dan notifikasi terbaru.
        </p>
      </header>

      <div
        v-if="errorMessage"
        class="soft-card rounded-[22px] p-5 text-sm text-[#b42318]"
      >
        {{ errorMessage }}
      </div>

      <section
        v-if="isLoading"
        class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3"
      >
        <div
          v-for="item in 3"
          :key="item"
          class="h-44 animate-pulse rounded-[18px] border border-[#ebe7df] bg-white"
        />
      </section>

      <section
        v-else-if="classes.length > 0"
        class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3"
      >
        <article
          v-for="(item, index) in classes"
          :key="item.classId ?? item.classCode ?? index"
          class="overflow-hidden rounded-[18px] border border-[#ebe7df] bg-white transition hover:-translate-y-0.5 hover:shadow-[0_18px_40px_rgba(66,55,40,0.08)]"
        >
          <div
            class="flex h-24 flex-col justify-end px-4 pb-4 text-white"
            :style="{ backgroundColor: palette[index % palette.length] }"
          >
            <h2 class="text-base font-medium">{{ classTitle(item) }}</h2>
          </div>
          <div class="space-y-3 px-4 py-4">
            <p class="min-h-9 text-xs leading-5 text-[#7a7385]">
              <template v-if="nearestDeadlineFor(item.subjectName)">
                <strong class="font-medium text-[#3f3a4a]">
                  {{ nearestDeadlineFor(item.subjectName)?.deadline }}
                </strong>
                <span>
                  -
                  {{
                    nearestDeadlineFor(item.subjectName)?.assignmentTitle
                  }}</span
                >
              </template>
              <template v-else>
                {{ classSubtitle(item) || "Belum ada deadline terdekat" }}
              </template>
            </p>
            <div class="flex items-center gap-2">
              <div class="h-1 flex-1 overflow-hidden rounded-full bg-[#f0ede8]">
                <div
                  class="h-full rounded-full"
                  :style="{
                    width: `${materialProgress}%`,
                    backgroundColor: palette[index % palette.length],
                  }"
                />
              </div>
              <span class="w-8 text-right text-[11px] text-[#9a95a3]"
                >{{ materialProgress }}%</span
              >
            </div>
          </div>
        </article>
      </section>

      <section v-else class="soft-card rounded-[22px] p-6">
        <p class="text-sm font-medium text-[#171322]">Belum ada kelas aktif</p>
        <p class="mt-2 text-sm text-[#7a7385]">
          Kelas akan muncul setelah akunmu terdaftar sebagai member kelas di
          sekolah aktif.
        </p>
      </section>

      <section class="grid gap-4 xl:grid-cols-[1.1fr_0.9fr]">
        <article class="soft-card rounded-[22px] p-5">
          <div class="mb-4 flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-[#171322]">
                Aktivitas akademik
              </p>
              <p class="text-xs text-[#8b8592]">
                Diambil dari dashboard siswa backend
              </p>
            </div>
            <PhBell :size="20" class="text-[#4f46e5]" />
          </div>
          <div class="grid gap-3 sm:grid-cols-3">
            <div class="rounded-2xl bg-[#eef2ff] p-4">
              <p class="text-2xl font-medium text-[#4f46e5]">
                {{ summary?.pendingAssignments ?? 0 }}
              </p>
              <p class="mt-1 text-xs text-[#6b6475]">tugas pending</p>
            </div>
            <div class="rounded-2xl bg-[#fff1f0] p-4">
              <p class="text-2xl font-medium text-[#f2756a]">
                {{ summary?.averageScore ?? 0 }}
              </p>
              <p class="mt-1 text-xs text-[#6b6475]">rata-rata nilai</p>
            </div>
            <div class="rounded-2xl bg-[#f3ecff] p-4">
              <p class="text-2xl font-medium text-[#9d5bd2]">
                {{ materialProgress }}%
              </p>
              <p class="mt-1 text-xs text-[#6b6475]">progress materi</p>
            </div>
          </div>
        </article>

        <article class="soft-card rounded-[22px] p-5">
          <p class="text-sm font-medium text-[#171322]">Deadline terdekat</p>
          <div
            v-if="summary?.upcomingDeadlines?.length"
            class="mt-4 space-y-3 text-sm text-[#6b6475]"
          >
            <div
              v-for="deadline in summary.upcomingDeadlines.slice(0, 2)"
              :key="deadline.assignmentId"
              class="rounded-2xl bg-[#fbfaf8] p-3"
            >
              <p class="font-medium text-[#3f3a4a]">
                {{ deadline.subjectName }}
              </p>
              <p class="mt-1 text-xs">
                {{ deadline.assignmentTitle }} - {{ deadline.deadline }}
              </p>
            </div>
          </div>
          <p
            v-else
            class="mt-4 rounded-2xl bg-[#fbfaf8] p-3 text-sm text-[#7a7385]"
          >
            Tidak ada deadline terdekat.
          </p>
        </article>
      </section>
    </section>

    <aside class="border-l border-[#ebe7df] bg-white/95">
      <div class="flex border-b border-[#ebe7df] px-5">
        <button
          class="flex items-center gap-2 border-b-2 border-[#4f46e5] px-1 py-4 text-sm font-medium text-[#4f46e5]"
          type="button"
        >
          <PhChatCircleText :size="18" />
          Aktivitas
        </button>
        <button class="px-5 py-4 text-sm text-[#a09aa8]" type="button">
          {{ unreadCount }} belum dibaca
        </button>
      </div>

      <div v-if="isLoading" class="space-y-2 p-4">
        <div
          v-for="item in 3"
          :key="item"
          class="h-16 animate-pulse rounded-2xl bg-[#bdbdbd]"
        />
      </div>
      <div v-else-if="notifications.length > 0" class="space-y-1 p-4">
        <article
          v-for="(item, index) in notifications"
          :key="item.notificationId"
          class="flex gap-3 rounded-2xl p-3 transition hover:bg-[#f8f7f4]"
          :class="!item.isRead ? 'bg-[#f5f7ff]' : ''"
        >
          <div
            class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full text-[11px] font-medium text-white"
            :style="{ backgroundColor: palette[index % palette.length] }"
          >
            {{ initials(item.title) }}
          </div>
          <div class="min-w-0 flex-1">
            <div class="flex items-baseline justify-between gap-2">
              <p class="truncate text-sm font-medium text-[#171322]">
                {{ item.title }}
              </p>
              <span class="shrink-0 text-[10px] text-[#a09aa8]">{{
                item.createdAt
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
              class="rounded-lg border border-[#ebe7df] p-1.5 text-[#7a7385]"
              type="button"
            >
              <PhCaretLeft :size="14" />
            </button>
            <button
              class="rounded-lg border border-[#ebe7df] p-1.5 text-[#7a7385]"
              type="button"
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
            class="relative rounded-lg py-1.5 text-xs text-[#4a4356]"
            :class="day.isToday ? 'bg-[#4f46e5] font-medium text-white' : ''"
          >
            {{ day.label }}
            <span
              v-if="day.hasEvent"
              class="absolute bottom-0.5 left-1/2 h-1 w-1 -translate-x-1/2 rounded-full bg-[#f0a05a]"
            />
          </span>
        </div>
      </section>
    </aside>
  </main>
</template>
