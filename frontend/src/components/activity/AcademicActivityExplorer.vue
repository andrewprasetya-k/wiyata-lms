<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { PhCalendarCheck, PhWarningCircle } from "@phosphor-icons/vue";
import { getAcademicActivities } from "../../services/activity";
import type { AcademicActivityItem } from "../../types/activity";
import AcademicActivityList from "./AcademicActivityList.vue";
import {
  activityFilters,
  activityGroupLabel,
  activityRanges,
  addDays,
  compareActivities,
  formatActivityDate,
  formatApiDate,
  type ActivityRole,
} from "./activityView";

const props = defineProps<{
  role: ActivityRole;
}>();

const activities = ref<AcademicActivityItem[]>([]);
const loading = ref(false);
const errorMessage = ref("");
const selectedFilter = ref("all");
const selectedRange = ref<"today" | "7d" | "30d">("7d");
let activityRequestVersion = 0;

const filters = computed(() => activityFilters(props.role));
const rangeOptions = activityRanges;

const helperText = computed(() =>
  props.role === "teacher"
    ? "Ringkasan pengumpulan, penilaian, tenggat, dan diskusi kelas."
    : "Ringkasan tugas, materi, pengumuman, dan nilai yang perlu diperhatikan.",
);

const emptyMessage = computed(() =>
  props.role === "teacher"
    ? "Tidak ada aktivitas mengajar pada rentang ini."
    : "Tidak ada aktivitas akademik pada rentang ini.",
);

const activeRange = computed(
  () =>
    rangeOptions.find((range) => range.value === selectedRange.value) ??
    rangeOptions[1],
);

const dateRange = computed(() => {
  const today = new Date();
  const to = addDays(today, activeRange.value.days);
  return {
    from: formatApiDate(today),
    to: formatApiDate(to),
    label:
      activeRange.value.days === 0
        ? formatActivityDate(formatApiDate(today))
        : `${formatActivityDate(formatApiDate(today))} - ${formatActivityDate(
            formatApiDate(to),
          )}`,
  };
});

const filteredActivities = computed(() => {
  const filter = filters.value.find(
    (item) => item.value === selectedFilter.value,
  );
  const types = filter?.types ?? [];

  return [...activities.value]
    .filter((item) => types.length === 0 || types.includes(item.type))
    .sort(compareActivities);
});

const groupedActivities = computed(() => {
  const groupOrder = ["Hari Ini", "Besok", "Minggu Ini", "Nanti", "Sebelumnya"];
  const groups = new Map<string, AcademicActivityItem[]>();

  for (const item of filteredActivities.value) {
    const label = activityGroupLabel(item.date);
    groups.set(label, [...(groups.get(label) ?? []), item]);
  }

  return groupOrder
    .filter((label) => groups.has(label))
    .map((label) => ({
      label,
      items: groups.get(label) ?? [],
    }));
});

const highPriorityCount = computed(
  () =>
    filteredActivities.value.filter((item) => item.priority === "high").length,
);

onMounted(loadActivities);

async function loadActivities() {
  const requestVersion = ++activityRequestVersion;
  const requestRange = dateRange.value;
  loading.value = true;
  errorMessage.value = "";

  try {
    const response = await getAcademicActivities({
      from: requestRange.from,
      to: requestRange.to,
    });
    if (requestVersion !== activityRequestVersion) return;
    activities.value = response.items ?? [];
  } catch {
    if (requestVersion !== activityRequestVersion) return;
    activities.value = [];
    errorMessage.value = "Aktivitas akademik belum bisa dimuat.";
  } finally {
    if (requestVersion === activityRequestVersion) {
      loading.value = false;
    }
  }
}

function selectRange(value: "today" | "7d" | "30d") {
  if (loading.value || selectedRange.value === value) return;
  selectedRange.value = value;
  loadActivities();
}
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-[#f8f7f4]">
    <header class="border-b border-border bg-white">
      <div class="px-5 py-5 sm:px-6 lg:px-8">
        <div class="max-w-5xl">
          <h1 class="mt-2 text-xl font-medium text-foreground sm:text-2xl">
            Aktivitas Akademik
          </h1>
          <p class="mt-1 max-w-2xl text-sm leading-6 text-muted">
            {{ helperText }}
          </p>
        </div>
      </div>
    </header>

    <section
      class="grid gap-4 px-5 py-5 sm:px-6 lg:px-8 lg:py-6 xl:grid-cols-[minmax(0,1fr)_300px]"
    >
      <div class="min-w-0 space-y-4">
        <section class="rounded-xl border border-border bg-white p-4 sm:p-5">
          <div
            class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between"
          >
            <div class="min-w-0">
              <p class="text-sm font-medium text-foreground">Filter aktivitas</p>
              <p class="mt-1 text-xs leading-5 text-[#8b8592]">
                Pilih jenis aktivitas dan rentang waktu yang ingin dilihat.
              </p>
            </div>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="range in rangeOptions"
                :key="range.value"
                type="button"
                class="rounded-lg border px-3 py-1.5 text-xs font-medium transition disabled:cursor-not-allowed disabled:opacity-60 focus:outline-none focus-visible:ring-2 focus-visible:ring-brand focus-visible:ring-offset-2"
                :class="
                  selectedRange === range.value
                    ? 'border-brand bg-[#eef2ff] text-brand'
                    : 'border-border bg-white text-muted hover:bg-[#fbfaf8]'
                "
                :aria-pressed="selectedRange === range.value"
                :disabled="loading"
                @click="selectRange(range.value)"
              >
                {{ range.label }}
              </button>
            </div>
          </div>

          <div class="mt-4 flex flex-wrap gap-2">
            <button
              v-for="filter in filters"
              :key="filter.value"
              type="button"
              class="rounded-full border px-3 py-1.5 text-xs font-medium transition focus:outline-none focus-visible:ring-2 focus-visible:ring-brand focus-visible:ring-offset-2"
              :class="
                selectedFilter === filter.value
                  ? 'border-brand bg-[#eef2ff] text-brand'
                  : 'border-border bg-white text-muted hover:bg-[#fbfaf8]'
              "
              :aria-pressed="selectedFilter === filter.value"
              @click="selectedFilter = filter.value"
            >
              {{ filter.label }}
            </button>
          </div>
        </section>

        <section
          v-if="loading"
          class="space-y-3 rounded-xl border border-border bg-white p-4 sm:p-5"
          aria-label="Memuat aktivitas akademik"
        >
          <div
            v-for="item in 5"
            :key="item"
            class="flex gap-3 rounded-lg border border-border bg-[#fbfaf8] p-4"
          >
            <div class="mt-1 h-2.5 w-2.5 shrink-0 rounded-full bg-[#e5e7eb]" />
            <div class="min-w-0 flex-1 space-y-2">
              <div class="h-3 w-24 animate-pulse rounded bg-[#e9e5dd]" />
              <div class="h-4 w-3/4 animate-pulse rounded bg-[#e9e5dd]" />
              <div class="h-3 w-1/2 animate-pulse rounded bg-[#eeeae3]" />
            </div>
          </div>
        </section>

        <section
          v-else-if="errorMessage"
          class="rounded-xl border border-[#fecaca] bg-white p-5"
        >
          <div
            class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"
          >
            <div class="flex min-w-0 gap-3">
              <PhWarningCircle
                :size="22"
                class="mt-0.5 shrink-0 text-[#dc2626]"
                weight="duotone"
              />
              <div class="min-w-0">
                <h2 class="text-sm font-medium text-foreground">
                  Aktivitas tidak dapat dimuat
                </h2>
                <p class="mt-1 text-sm leading-6 text-[#7a7385]">
                  {{ errorMessage }}
                </p>
              </div>
            </div>
            <button
              class="rounded-lg bg-brand px-4 py-2 text-sm font-medium text-white transition hover:bg-[#4338ca] disabled:cursor-not-allowed disabled:opacity-60 focus:outline-none focus-visible:ring-2 focus-visible:ring-brand focus-visible:ring-offset-2"
              type="button"
              :disabled="loading"
              @click="loadActivities"
            >
              Coba lagi
            </button>
          </div>
        </section>

        <section
          v-else-if="groupedActivities.length === 0"
          class="rounded-lg bg-[#fbfaf8] px-5 py-8 text-center"
        >
          <PhCalendarCheck
            class="mx-auto h-7 w-7 text-[#9ca3af]"
            weight="duotone"
          />
          <h2 class="mt-3 text-sm font-semibold text-foreground">
            {{ emptyMessage }}
          </h2>
          <p class="mt-2 text-sm leading-6 text-muted">
            Coba ubah rentang waktu atau filter untuk melihat aktivitas lain.
          </p>
        </section>

        <AcademicActivityList v-else :groups="groupedActivities" :role="role" />
      </div>

      <aside class="min-w-0 space-y-3">
        <section class="rounded-xl border border-border bg-white p-4 sm:p-5">
          <h2 class="text-sm font-medium text-foreground">Ringkasan</h2>
          <dl class="mt-4 space-y-3 text-sm">
            <div class="flex items-center justify-between gap-3">
              <dt class="text-[#7a7385]">Total aktivitas</dt>
              <dd class="font-medium text-foreground">
                {{ filteredActivities.length }}
              </dd>
            </div>
            <div class="flex items-center justify-between gap-3">
              <dt class="text-[#7a7385]">Prioritas</dt>
              <dd class="font-medium text-foreground">
                {{ highPriorityCount }}
              </dd>
            </div>
            <div class="border-t border-border pt-3">
              <dt class="text-xs text-[#8b8592]">Rentang</dt>
              <dd class="mt-1 text-sm font-medium text-foreground">
                {{ dateRange.label }}
              </dd>
            </div>
          </dl>
        </section>
      </aside>
    </section>
  </main>
</template>
