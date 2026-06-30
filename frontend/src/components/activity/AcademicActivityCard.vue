<script setup lang="ts">
import { computed } from "vue";
import { RouterLink } from "vue-router";
import { PhArrowRight, PhCalendarCheck } from "@phosphor-icons/vue";
import type { AcademicActivityItem } from "../../types/activity";
import { getSubjectColor } from "../../utils/color";

const props = withDefaults(
  defineProps<{
    activities: AcademicActivityItem[];
    loading: boolean;
    error?: string;
    role: "student" | "teacher";
    maxItems?: number;
  }>(),
  {
    error: "",
    maxItems: 5,
  },
);

const sortedActivities = computed(() =>
  [...props.activities].sort((left, right) => {
    const priorityDiff =
      priorityWeight(right.priority) - priorityWeight(left.priority);
    if (priorityDiff !== 0) return priorityDiff;

    return getActivityTime(left) - getActivityTime(right);
  }),
);

const visibleActivities = computed(() =>
  sortedActivities.value.slice(0, props.maxItems),
);

const hiddenActivityCount = computed(() =>
  Math.max(0, sortedActivities.value.length - visibleActivities.value.length),
);

const subtitle = computed(() =>
  props.role === "teacher"
    ? "Aktivitas mengajar yang perlu ditindaklanjuti."
    : "Aktivitas akademik yang perlu diperhatikan.",
);

const emptyMessage = computed(() =>
  props.role === "teacher"
    ? "Tidak ada aktivitas mengajar yang perlu ditindaklanjuti hari ini."
    : "Tidak ada aktivitas akademik hari ini.",
);

const activityRoute = computed(() =>
  props.role === "teacher" ? "/teacher/activity" : "/student/activity",
);

function priorityWeight(priority?: string | null) {
  return priority === "high" ? 1 : 0;
}

function getActivityTime(item: AcademicActivityItem) {
  const dateTime = `${item.date || ""}T${item.time || "00:00"}`;
  const parsed = new Date(dateTime).getTime();
  return Number.isNaN(parsed) ? Number.MAX_SAFE_INTEGER : parsed;
}

function typeLabel(type: string) {
  if (props.role === "teacher") {
    const labels: Record<string, string> = {
      submission_received: "Pengumpulan",
      submission_pending_review: "Perlu dinilai",
      assignment_due: "Tugas",
      feed_comment: "Komentar",
    };
    return labels[type] ?? "Aktivitas";
  }

  const labels: Record<string, string> = {
    assignment_due: "Tugas",
    material_created: "Materi",
    feed_posted: "Pengumuman",
    assignment_graded: "Nilai",
  };
  return labels[type] ?? "Aktivitas";
}

function subjectColor(item: AcademicActivityItem) {
  return (
    item.subject?.color ||
    getSubjectColor(
      item.subject?.id ||
        item.subject?.name ||
        item.subject?.code ||
        item.class?.id ||
        item.class?.name,
    )
  );
}

function timelineLabel(item: AcademicActivityItem) {
  const relative = relativeDateLabel(item.date);
  if (item.time) return `${relative}, ${item.time}`;
  return relative;
}

function relativeDateLabel(value?: string | null) {
  if (!value) return "Tanggal belum tersedia";

  const date = parseDate(value);
  if (!date) return "Tanggal belum tersedia";

  const today = startOfDay(new Date());
  const target = startOfDay(date);
  const diffDays = Math.round(
    (target.getTime() - today.getTime()) / 86_400_000,
  );

  if (diffDays === 0) return "Hari ini";
  if (diffDays === 1) return "Besok";

  return new Intl.DateTimeFormat("id-ID", {
    day: "2-digit",
    month: "short",
  }).format(target);
}

function parseDate(value: string) {
  const parts = value.split("-").map(Number);
  if (parts.length !== 3 || parts.some((part) => Number.isNaN(part))) {
    return null;
  }

  const [year, month, day] = parts;
  return new Date(year, month - 1, day);
}

function startOfDay(date: Date) {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate());
}

function isInternalLink(link?: string | null) {
  return Boolean(link && link.startsWith("/") && !link.startsWith("//"));
}
</script>

<template>
  <article
    class="min-w-0 rounded-xl border border-[#ebe7df] bg-white p-4 sm:p-5"
    aria-labelledby="academic-activity-title"
  >
    <div class="mb-4 flex min-w-0 items-start justify-between gap-3">
      <div class="min-w-0">
        <h2
          id="academic-activity-title"
          class="text-sm font-semibold text-[#171322]"
        >
          Hari Ini
        </h2>
        <p class="mt-1 text-xs leading-5 text-[#8b8592]">
          {{ subtitle }}
        </p>
      </div>
      <RouterLink
        :to="activityRoute"
        class="inline-flex shrink-0 items-center gap-1 rounded-lg px-2 py-1 text-xs font-medium text-[#4f46e5] transition hover:bg-[#eef2ff] hover:text-[#4338ca] focus:outline-none focus-visible:ring-2 focus-visible:ring-[#4f46e5] focus-visible:ring-offset-2"
        aria-label="Lihat semua aktivitas akademik"
      >
        Lihat semua aktivitas
        <PhArrowRight :size="14" />
      </RouterLink>
    </div>

    <div v-if="loading" class="space-y-3" aria-label="Memuat aktivitas">
      <div
        v-for="item in 4"
        :key="item"
        class="flex min-w-0 gap-3 rounded-lg border border-[#ebe7df] bg-[#fbfaf8] p-3"
      >
        <div class="mt-1 h-2 w-2 shrink-0 rounded-full bg-[#e5e7eb]" />
        <div class="min-w-0 flex-1 space-y-2">
          <div class="h-3 w-20 animate-pulse rounded bg-[#e9e5dd]" />
          <div class="h-4 w-3/4 animate-pulse rounded bg-[#e9e5dd]" />
          <div class="h-3 w-1/2 animate-pulse rounded bg-[#eeeae3]" />
        </div>
      </div>
    </div>

    <div
      v-else-if="error"
      class="rounded-lg border border-[#ebe7df] bg-[#fbfaf8] p-4 text-sm leading-6 text-[#7a7385]"
    >
      Aktivitas belum bisa dimuat. Dashboard lain tetap dapat digunakan.
    </div>

    <div
      v-else-if="visibleActivities.length === 0"
      class="rounded-lg border border-[#ebe7df] bg-[#fbfaf8] p-4"
    >
      <div
        class="mb-3 flex h-9 w-9 items-center justify-center rounded-lg bg-[#eef2ff] text-[#4f46e5]"
      >
        <PhCalendarCheck :size="18" weight="duotone" />
      </div>
      <p class="text-sm font-medium text-[#171322]">
        {{ emptyMessage }}
      </p>
      <p class="mt-2 text-xs leading-5 text-[#7a7385]">
        Aktivitas akan muncul saat ada tenggat, materi, pengumuman, atau
        pengumpulan yang relevan.
      </p>
    </div>

    <div v-else>
      <ul class="divide-y divide-[#ebe7df]" aria-label="Daftar aktivitas">
        <li
          v-for="activity in visibleActivities"
          :key="activity.id"
          class="min-w-0 py-3 first:pt-0 last:pb-0"
        >
          <RouterLink
            v-if="isInternalLink(activity.link)"
            :to="activity.link || ''"
            class="group flex min-w-0 gap-3 rounded-lg p-1 -m-1 transition hover:bg-[#fbfaf8] focus:outline-none focus-visible:ring-2 focus-visible:ring-[#4f46e5] focus-visible:ring-offset-2"
            :aria-label="`${typeLabel(activity.type)}: ${activity.title}`"
          >
            <span
              class="mt-2 h-2 w-2 shrink-0 rounded-full"
              :style="{ backgroundColor: subjectColor(activity) }"
              aria-hidden="true"
            />
            <span class="min-w-0 flex-1">
              <span class="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1">
                <span class="text-[11px] font-medium text-[#4f46e5]">
                  {{ typeLabel(activity.type) }}
                </span>
                <span class="text-[11px] text-[#9ca3af]">
                  {{ timelineLabel(activity) }}
                </span>
                <span
                  v-if="activity.priority === 'high'"
                  class="rounded-full bg-[#fff7ed] px-2 py-0.5 text-[10px] font-medium text-[#b45309]"
                >
                  Prioritas
                </span>
              </span>
              <span
                class="mt-1 line-clamp-2 text-sm font-medium leading-5 text-[#171322] transition group-hover:text-[#4f46e5]"
              >
                {{ activity.title }}
              </span>
              <span class="mt-1 line-clamp-2 text-xs leading-5 text-[#7a7385]">
                {{ activity.description }}
              </span>
            </span>
            <PhArrowRight
              :size="14"
              class="mt-2 shrink-0 text-[#a09aa8] transition group-hover:translate-x-0.5 group-hover:text-[#4f46e5]"
              aria-hidden="true"
            />
          </RouterLink>

          <div v-else class="flex min-w-0 gap-3">
            <span
              class="mt-2 h-2 w-2 shrink-0 rounded-full"
              :style="{ backgroundColor: subjectColor(activity) }"
              aria-hidden="true"
            />
            <div class="min-w-0 flex-1">
              <div class="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1">
                <span class="text-[11px] font-medium text-[#4f46e5]">
                  {{ typeLabel(activity.type) }}
                </span>
                <span class="text-[11px] text-[#9ca3af]">
                  {{ timelineLabel(activity) }}
                </span>
                <span
                  v-if="activity.priority === 'high'"
                  class="rounded-full bg-[#fff7ed] px-2 py-0.5 text-[10px] font-medium text-[#b45309]"
                >
                  Prioritas
                </span>
              </div>
              <p class="mt-1 line-clamp-2 text-sm font-medium text-[#171322]">
                {{ activity.title }}
              </p>
              <p class="mt-1 line-clamp-2 text-xs leading-5 text-[#7a7385]">
                {{ activity.description }}
              </p>
            </div>
          </div>
        </li>
      </ul>

      <p
        v-if="hiddenActivityCount > 0"
        class="mt-3 rounded-lg bg-[#fbfaf8] px-3 py-2 text-xs text-[#7a7385]"
      >
        +{{ hiddenActivityCount }} aktivitas lainnya
      </p>
    </div>
  </article>
</template>
