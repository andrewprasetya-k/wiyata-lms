<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import { PhCaretLeft, PhCaretRight } from "@phosphor-icons/vue";
import { getAcademicActivities } from "../../services/activity";
import type { AcademicActivityItem } from "../../types/activity";
import {
  activitySubjectColor,
  activityTypeLabel,
  compareActivities,
  formatActivityDate,
  formatApiDate,
  isInternalActivityLink,
  parseActivityDate,
} from "../activity/activityView";

const calendarActivityCache = ref<Record<string, AcademicActivityItem[]>>({});
const calendarActivitiesLoading = ref(false);
const calendarActivitiesError = ref("");
const selectedDate = ref(formatApiDate(new Date()));
const viewDate = ref(new Date());

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

function changeMonth(step: number) {
  const newDate = new Date(viewDate.value);
  newDate.setMonth(newDate.getMonth() + step);
  viewDate.value = newDate;
  selectedDate.value = defaultSelectedDateForMonth(newDate);
  loadCalendarActivities();
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

function buildCalendarDays(date: Date) {
  const year = date.getFullYear();
  const month = date.getMonth();
  const firstDay = new Date(year, month, 1);
  const startOffset = firstDay.getDay();
  const calendarStart = new Date(year, month, 1 - startOffset);
  const days = [];

  const realToday = new Date();

  for (let index = 0; index < 42; index += 1) {
    const currentDate = new Date(
      calendarStart.getFullYear(),
      calendarStart.getMonth(),
      calendarStart.getDate() + index,
    );
    const dateKey = formatApiDate(currentDate);
    const dayActivities = activitiesForDate(dateKey);
    const isCurrentMonth =
      currentDate.getMonth() === month && currentDate.getFullYear() === year;
    days.push({
      key: dateKey,
      label: String(currentDate.getDate()),
      isCurrentMonth,
      isToday:
        currentDate.getFullYear() === realToday.getFullYear() &&
        currentDate.getMonth() === realToday.getMonth() &&
        currentDate.getDate() === realToday.getDate(),
      dateKey,
      activities: dayActivities.slice(0, 3),
      extraCount: Math.max(0, dayActivities.length - 3),
    });
  }

  return days;
}

function activitiesForDate(dateKey: string) {
  return calendarActivities.value
    .filter((item) => item.type === "assignment_due")
    .filter((item) => item.date === dateKey)
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
  isCurrentMonth?: boolean;
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
  const monthContext = day.isCurrentMonth ? "" : ", di luar bulan ini";
  return `${label}${monthContext}, ${count} aktivitas`;
}

function calendarActivityTime(activity: AcademicActivityItem) {
  return activity.time || "Sepanjang hari";
}

onMounted(() => {
  loadCalendarActivities();
});
</script>

<template>
  <section class="shrink-0 rounded-xl bg-surface p-3">
    <div class="mb-2 flex items-center justify-between">
      <p class="text-sm font-medium text-foreground">{{ currentMonth }}</p>
      <div class="flex gap-1">
        <button
          class="rounded-lg border border-border p-1.5 text-muted transition hover:bg-surface-subtle"
          type="button"
          @click="changeMonth(-1)"
        >
          <PhCaretLeft :size="14" />
        </button>
        <button
          class="rounded-lg border border-border p-1.5 text-muted transition hover:bg-surface-subtle"
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
        class="py-0.5 text-[10px] text-[#a09aa8]"
      >
        {{ day }}
      </span>
      <span v-for="day in calendarDays" :key="day.key" class="min-h-8">
        <button
          v-if="day.dateKey"
          type="button"
          class="flex h-full min-h-8 w-full flex-col items-center justify-center rounded-lg px-1 py-0.5 text-xs transition focus:outline-none focus-visible:ring-2 focus-visible:ring-brand focus-visible:ring-offset-2"
          :class="[
            day.isToday
              ? 'bg-brand font-medium text-white'
              : day.isCurrentMonth
                ? 'text-[#4a4356] hover:bg-surface-subtle'
                : 'text-[#c0bac8] hover:bg-surface-subtle',
            selectedDate === day.dateKey
              ? 'ring-2 ring-brand ring-offset-1'
              : '',
          ]"
          :aria-label="calendarDateAriaLabel(day)"
          :aria-pressed="selectedDate === day.dateKey"
          @click="selectCalendarDate(day.dateKey)"
        >
          <span>{{ day.label }}</span>
          <span
            v-if="day.activities.length || day.extraCount"
            class="mt-0.5 flex h-2 items-center justify-center gap-0.5"
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
              :class="day.isToday ? 'text-white' : 'text-muted'"
            >
              +{{ day.extraCount }}
            </span>
          </span>
        </button>
        <span v-else class="block min-h-8" />
      </span>
    </div>

    <div class="mt-3 border-t border-border pt-3">
      <div class="mb-2 flex items-center justify-between gap-3">
        <p class="text-sm font-medium text-foreground">Deadline Tugas</p>
        <p class="shrink-0 text-xs text-[#8b8592]">
          {{ formatActivityDate(selectedDate) }}
        </p>
      </div>

      <div class="h-32 overflow-y-auto pr-1">
        <div
          v-if="calendarActivitiesLoading"
          class="rounded-lg bg-surface-subtle p-3 text-xs leading-5 text-muted"
        >
          Memuat deadline...
        </div>

        <div
          v-else-if="calendarActivitiesError"
          class="rounded-lg bg-surface-subtle p-3 text-xs leading-5 text-muted"
        >
          {{ calendarActivitiesError }}
        </div>

        <div
          v-else-if="selectedDatePreview.length === 0"
          class="border-b border-border bg-surface-subtle p-3 text-xs leading-5 text-muted"
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
              class="group flex min-w-0 items-start gap-2 border-b border-border bg-surface-subtle p-3 transition hover:border-brand-line hover:bg-surface focus:outline-none focus-visible:ring-2 focus-visible:ring-brand focus-visible:ring-offset-2"
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
                  <span class="text-[11px] font-medium text-brand">
                    {{ activityTypeLabel(activity.type, "student") }}
                  </span>
                  <span class="shrink-0 text-[10px] text-muted">
                    {{ calendarActivityTime(activity) }}
                  </span>
                </span>
                <span
                  class="mt-1 block truncate text-xs font-medium text-foreground transition group-hover:text-brand"
                >
                  {{ activity.title }}
                </span>
              </span>
            </RouterLink>

            <article
              v-else
              class="flex min-w-0 items-start gap-2 rounded-lg bg-surface-subtle p-3"
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
                  <span class="text-[11px] font-medium text-brand">
                    {{ activityTypeLabel(activity.type, "student") }}
                  </span>
                  <span class="shrink-0 text-[10px] text-muted">
                    {{ calendarActivityTime(activity) }}
                  </span>
                </div>
                <p
                  class="mt-1 truncate text-xs font-medium text-foreground"
                >
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
    </div>
  </section>
</template>
