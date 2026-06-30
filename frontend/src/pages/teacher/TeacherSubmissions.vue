<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import {
  PhArrowRight,
  PhCheckCircle,
  PhClipboardText,
  PhClock,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { getTeacherSubmissionInbox } from "../../services/teacherAssignment";
import type {
  TeacherSubmissionInboxItem,
  TeacherSubmissionInboxSummary,
} from "../../types/teacherAssignment";
import { formatDate, parseBackendTimestamp } from "../../utils/date";

type InboxFilter = "all" | "pending" | "graded";

const loading = ref(false);
const errorMessage = ref("");
const inboxItems = ref<TeacherSubmissionInboxItem[]>([]);
const inboxSummary = ref<TeacherSubmissionInboxSummary>({
  totalSubmissions: 0,
  pendingCount: 0,
  gradedCount: 0,
  lateCount: 0,
});
const activeFilter = ref<InboxFilter>("all");

const summary = computed(() => ({
  submissions: inboxSummary.value.totalSubmissions,
  pending: inboxSummary.value.pendingCount,
  graded: inboxSummary.value.gradedCount,
  late: inboxSummary.value.lateCount,
}));

const filterTabs = computed(() => [
  { id: "all" as const, label: "Semua", count: inboxItems.value.length },
  {
    id: "pending" as const,
    label: "Perlu dinilai",
    count: inboxItems.value.filter((item) => item.pendingCount > 0).length,
  },
  {
    id: "graded" as const,
    label: "Sudah dinilai",
    count: inboxItems.value.filter(
      (item) => item.submissionCount > 0 && item.pendingCount === 0,
    ).length,
  },
]);

const filteredItems = computed(() => {
  const items = inboxItems.value.filter((item) => {
    if (activeFilter.value === "pending") return item.pendingCount > 0;
    if (activeFilter.value === "graded") {
      return item.submissionCount > 0 && item.pendingCount === 0;
    }
    return true;
  });

  return [...items].sort(compareInboxItems);
});

function compareInboxItems(
  a: TeacherSubmissionInboxItem,
  b: TeacherSubmissionInboxItem,
) {
  const pendingDiff = Number(b.pendingCount > 0) - Number(a.pendingCount > 0);
  if (pendingDiff !== 0) return pendingDiff;

  const aDeadline = getDeadlineTime(a.deadline);
  const bDeadline = getDeadlineTime(b.deadline);
  if (aDeadline !== bDeadline) return aDeadline - bDeadline;

  return (a.assignmentTitle || "").localeCompare(b.assignmentTitle || "");
}

function getDeadlineTime(deadline?: string | null) {
  if (!deadline) return Number.MAX_SAFE_INTEGER;
  const value = parseBackendTimestamp(deadline)?.getTime() ?? Number.NaN;
  return Number.isNaN(value) ? Number.MAX_SAFE_INTEGER : value;
}

async function loadInbox() {
  loading.value = true;
  errorMessage.value = "";
  inboxItems.value = [];
  inboxSummary.value = {
    totalSubmissions: 0,
    pendingCount: 0,
    gradedCount: 0,
    lateCount: 0,
  };

  try {
    const response = await getTeacherSubmissionInbox();
    inboxItems.value = response.items ?? [];
    inboxSummary.value = response.summary ?? inboxSummary.value;
  } catch {
    errorMessage.value =
      "Inbox pengumpulan belum bisa dimuat. Coba lagi beberapa saat.";
  } finally {
    loading.value = false;
  }
}

onMounted(loadInbox);
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-[#f8f7f4]">
    <header class="border-b border-[#ebe7df] bg-white">
      <div class="px-5 py-5 sm:px-6 lg:px-8">
        <div
          class="mt-2 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between"
        >
          <div class="min-w-0">
            <h1 class="text-2xl font-semibold text-[#171322] sm:text-3xl">
              Inbox Pengumpulan
            </h1>
            <p class="mt-2 max-w-2xl text-sm leading-6 text-[#6b6475]">
              Pantau pengumpulan dari semua mata pelajaran yang Anda ajar.
              Penilaian dan umpan balik tetap dilakukan di halaman nilai tugas.
            </p>
          </div>
          <p class="shrink-0 text-sm text-[#8a8494]">
            {{ inboxItems.length }} tugas
          </p>
        </div>
      </div>
    </header>

    <section class="space-y-5 px-5 py-5 sm:px-6 lg:px-8">
      <template v-if="loading">
        <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
          <div
            v-for="index in 4"
            :key="index"
            class="h-28 animate-pulse rounded-xl border border-[#ebe7df] bg-white"
          />
        </section>
        <section
          class="h-56 animate-pulse rounded-xl border border-[#ebe7df] bg-white"
        />
      </template>

      <section
        v-else-if="errorMessage"
        class="rounded-xl border border-[#f0d8d2] bg-white px-5 py-8 text-center"
      >
        <PhWarningCircle
          :size="30"
          class="mx-auto text-[#d97757]"
          weight="duotone"
        />
        <h2 class="mt-3 text-lg font-semibold text-[#171322]">
          Pengumpulan belum bisa dimuat
        </h2>
        <p class="mx-auto mt-2 max-w-xl text-sm leading-6 text-[#6b6475]">
          {{ errorMessage }}
        </p>
        <button
          type="button"
          class="mt-5 inline-flex items-center justify-center rounded-lg bg-[#171322] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#2f2b3a]"
          @click="loadInbox"
        >
          Coba lagi
        </button>
      </section>

      <template v-else>
        <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
          <article class="rounded-xl border border-[#ebe7df] bg-white p-4">
            <div class="flex items-center justify-between gap-3">
              <p class="text-sm text-[#6b6475]">Total pengumpulan</p>
              <PhClipboardText
                :size="21"
                class="text-[#7aa7d9]"
                weight="duotone"
              />
            </div>
            <p class="mt-3 text-2xl font-semibold text-[#171322]">
              {{ summary.submissions }}
            </p>
          </article>
          <article class="rounded-xl border border-[#ebe7df] bg-white p-4">
            <div class="flex items-center justify-between gap-3">
              <p class="text-sm text-[#6b6475]">Perlu dinilai</p>
              <PhWarningCircle
                :size="21"
                class="text-[#e58f86]"
                weight="duotone"
              />
            </div>
            <p class="mt-3 text-2xl font-semibold text-[#171322]">
              {{ summary.pending }}
            </p>
          </article>
          <article class="rounded-xl border border-[#ebe7df] bg-white p-4">
            <div class="flex items-center justify-between gap-3">
              <p class="text-sm text-[#6b6475]">Sudah dinilai</p>
              <PhCheckCircle
                :size="21"
                class="text-[#74bfa5]"
                weight="duotone"
              />
            </div>
            <p class="mt-3 text-2xl font-semibold text-[#171322]">
              {{ summary.graded }}
            </p>
          </article>
          <article class="rounded-xl border border-[#ebe7df] bg-white p-4">
            <div class="flex items-center justify-between gap-3">
              <p class="text-sm text-[#6b6475]">Terlambat</p>
              <PhClock :size="21" class="text-[#b889c9]" weight="duotone" />
            </div>
            <p class="mt-3 text-2xl font-semibold text-[#171322]">
              {{ summary.late }}
            </p>
          </article>
        </section>

        <section class="rounded-xl border border-[#ebe7df] bg-white">
          <div
            class="flex flex-col gap-4 border-b border-[#ebe7df] px-4 py-4 sm:px-5 lg:flex-row lg:items-end lg:justify-between"
          >
            <div>
              <h2 class="text-base font-semibold text-[#171322]">
                Daftar tugas
              </h2>
              <p class="mt-1 text-sm text-[#8a8494]">
                {{ inboxItems.length }} tugas memiliki data pengumpulan di
                sekolah aktif.
              </p>
            </div>
            <div class="flex max-w-full gap-2 overflow-x-auto pb-1">
              <button
                v-for="tab in filterTabs"
                :key="tab.id"
                type="button"
                class="shrink-0 rounded-lg px-3.5 py-2 text-sm font-medium transition"
                :class="
                  activeFilter === tab.id
                    ? 'bg-[#171322] text-white'
                    : 'bg-[#faf8f4] text-[#6b6475] hover:bg-[#f0e9dd] hover:text-[#171322]'
                "
                @click="activeFilter = tab.id"
              >
                {{ tab.label }}
                <span class="ml-1.5 opacity-70">{{ tab.count }}</span>
              </button>
            </div>
          </div>

          <div v-if="inboxItems.length === 0" class="px-5 py-12 text-center">
            <PhClipboardText
              :size="34"
              class="mx-auto text-[#b5afbf]"
              weight="duotone"
            />
            <h2 class="mt-3 text-lg font-semibold text-[#171322]">
              Belum ada pengumpulan
            </h2>
            <p class="mx-auto mt-2 max-w-xl text-sm leading-6 text-[#6b6475]">
              Pengumpulan akan tampil setelah siswa mengumpulkan tugas pada mata
              pelajaran yang Anda ajar.
            </p>
          </div>

          <div
            v-else-if="filteredItems.length === 0"
            class="px-5 py-12 text-center"
          >
            <PhCheckCircle
              :size="34"
              class="mx-auto text-[#b5afbf]"
              weight="duotone"
            />
            <h2 class="mt-3 text-lg font-semibold text-[#171322]">
              Tidak ada hasil
            </h2>
            <p class="mx-auto mt-2 max-w-xl text-sm leading-6 text-[#6b6475]">
              Tidak ada tugas yang sesuai dengan filter saat ini.
            </p>
          </div>

          <div v-else class="divide-y divide-[#ebe7df]">
            <article
              v-for="item in filteredItems"
              :key="`${item.subjectClassId}-${item.assignmentId}`"
              class="px-4 py-5 sm:px-5"
            >
              <div
                class="flex min-w-0 flex-col gap-4 lg:flex-row lg:items-start lg:justify-between"
              >
                <div class="min-w-0">
                  <div class="flex flex-wrap gap-2 text-xs font-medium">
                    <span
                      class="rounded-lg bg-[#eef0ff] px-2.5 py-1 text-[#4f46e5]"
                    >
                      {{ item.subjectName }}
                      <template v-if="item.subjectCode">
                        · {{ item.subjectCode }}
                      </template>
                    </span>
                    <span
                      class="rounded-lg bg-[#faf8f4] px-2.5 py-1 text-[#6b6475]"
                    >
                      {{ item.className || item.classCode || "Kelas" }}
                    </span>
                  </div>

                  <h3
                    class="mt-3 wrap-break-word text-base font-semibold text-[#171322] sm:text-lg"
                  >
                    {{ item.assignmentTitle }}
                  </h3>
                  <p
                    v-if="item.deadline"
                    class="mt-1.5 inline-flex items-center gap-1.5 text-sm text-[#6b6475]"
                  >
                    <PhClock :size="15" weight="duotone" />
                    Tenggat {{ formatDate(item.deadline) }}
                  </p>
                </div>

                <RouterLink
                  :to="{
                    name: 'teacher-assignment-review',
                    params: { assignmentId: item.assignmentId },
                  }"
                  class="inline-flex shrink-0 items-center justify-center gap-2 rounded-lg bg-[#171322] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#2f2b3a]"
                >
                  Nilai pengumpulan
                  <PhArrowRight :size="16" />
                </RouterLink>
              </div>

              <div class="mt-4 grid grid-cols-2 gap-2 sm:grid-cols-4">
                <div class="rounded-lg bg-[#faf8f4] px-3 py-2.5">
                  <p class="text-xs text-[#8a8494]">Pengumpulan</p>
                  <p class="mt-1 text-lg font-semibold text-[#171322]">
                    {{ item.submissionCount }}
                  </p>
                </div>
                <div class="rounded-lg bg-[#fff7e8] px-3 py-2.5">
                  <p class="text-xs text-[#9f6b1d]">Perlu dinilai</p>
                  <p class="mt-1 text-lg font-semibold text-[#171322]">
                    {{ item.pendingCount }}
                  </p>
                </div>
                <div class="rounded-lg bg-[#eef7f2] px-3 py-2.5">
                  <p class="text-xs text-[#2f7d5c]">Sudah dinilai</p>
                  <p class="mt-1 text-lg font-semibold text-[#171322]">
                    {{ item.gradedCount }}
                  </p>
                </div>
                <div class="rounded-lg bg-[#fff1ed] px-3 py-2.5">
                  <p class="text-xs text-[#b86845]">Terlambat</p>
                  <p class="mt-1 text-lg font-semibold text-[#171322]">
                    {{ item.lateCount }}
                  </p>
                </div>
              </div>
            </article>
          </div>
        </section>
      </template>
    </section>
  </main>
</template>
