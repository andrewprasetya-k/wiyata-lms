<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import {
  PhArrowRight,
  PhCheckCircle,
  PhClipboardText,
  PhClock,
  PhSealCheck,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { getStudentAssignmentInbox } from "../../services/assignment";
import type {
  StudentAssignmentInboxItem,
  StudentAssignmentInboxSummary,
} from "../../types/assignment";
import {
  formatDate,
  formatDateTime,
  parseBackendTimestamp,
} from "../../utils/date";

type AssignmentFilter =
  | "all"
  | "not_submitted"
  | "submitted"
  | "graded"
  | "overdue";

const isLoading = ref(true);
const errorMessage = ref("");
const items = ref<StudentAssignmentInboxItem[]>([]);
const inboxSummary = ref<StudentAssignmentInboxSummary>({
  totalAssignments: 0,
  notSubmittedCount: 0,
  submittedCount: 0,
  gradedCount: 0,
  overdueCount: 0,
});
const activeFilter = ref<AssignmentFilter>("all");
const summary = computed(() => inboxSummary.value);

const filterTabs = computed(() => [
  { id: "all" as const, label: "Semua", count: items.value.length },
  {
    id: "not_submitted" as const,
    label: "Belum dikumpulkan",
    count: items.value.filter((item) => !item.isSubmitted).length,
  },
  {
    id: "submitted" as const,
    label: "Sudah dikumpulkan",
    count: items.value.filter((item) => item.isSubmitted).length,
  },
  {
    id: "graded" as const,
    label: "Sudah dinilai",
    count: items.value.filter((item) => item.isGraded).length,
  },
  {
    id: "overdue" as const,
    label: "Lewat deadline",
    count: items.value.filter((item) => item.isOverdue).length,
  },
]);

const filteredItems = computed(() => {
  const filtered = items.value.filter((item) => {
    if (activeFilter.value === "not_submitted") return !item.isSubmitted;
    if (activeFilter.value === "submitted") return item.isSubmitted;
    if (activeFilter.value === "graded") return item.isGraded;
    if (activeFilter.value === "overdue") return item.isOverdue;
    return true;
  });

  return [...filtered].sort(compareAssignments);
});

async function loadAssignments() {
  isLoading.value = true;
  errorMessage.value = "";
  items.value = [];
  inboxSummary.value = {
    totalAssignments: 0,
    notSubmittedCount: 0,
    submittedCount: 0,
    gradedCount: 0,
    overdueCount: 0,
  };

  try {
    const response = await getStudentAssignmentInbox();
    items.value = response.items ?? [];
    inboxSummary.value = response.summary ?? inboxSummary.value;
  } catch {
    errorMessage.value =
      "Tugas belum bisa dimuat. Periksa koneksi atau coba lagi nanti.";
  } finally {
    isLoading.value = false;
  }
}

function compareAssignments(
  a: StudentAssignmentInboxItem,
  b: StudentAssignmentInboxItem,
) {
  const overdueDiff = Number(b.isOverdue) - Number(a.isOverdue);
  if (overdueDiff !== 0) return overdueDiff;

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

function statusLabel(item: StudentAssignmentInboxItem) {
  if (item.isGraded) return "Sudah dinilai";
  if (item.isSubmitted) return "Sudah dikumpulkan";
  if (item.isOverdue) return "Lewat deadline";
  return "Belum dikumpulkan";
}

function statusClasses(item: StudentAssignmentInboxItem) {
  if (item.isGraded) return "bg-[#ecfdf3] text-[#027a48]";
  if (item.isSubmitted) return "bg-[#eef2ff] text-[#4f46e5]";
  if (item.isOverdue) return "bg-[#fef2f2] text-[#dc2626]";
  return "bg-[#fff7ed] text-[#b45309]";
}

function formatScore(value?: number | null) {
  if (value === null || value === undefined) return "Belum tersedia";
  return new Intl.NumberFormat("id-ID", { maximumFractionDigits: 2 }).format(
    value,
  );
}

onMounted(loadAssignments);
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-[#f8f7f4]">
    <header class="border-b border-[#ebe7df] bg-white">
      <div
        class="flex min-w-0 flex-col gap-3 px-8 py-5 sm:px-6 lg:flex-row lg:items-center lg:justify-between lg:px-8"
      >
        <div class="min-w-0">
          <h1 class="text-xl font-medium text-[#171322] sm:text-2xl">
            Tugas Saya
          </h1>
          <p class="mt-1 max-w-2xl text-xs leading-5 text-[#6b7280] sm:text-sm">
            Pantau tugas dari semua kelas aktif. Pengumpulan tetap dilakukan
            dari halaman detail tugas.
          </p>
        </div>
      </div>
    </header>

    <section
      class="mx-auto max-w-screen min-w-0 px-5 py-5 sm:px-6 lg:px-8 lg:py-6"
    >
      <section v-if="isLoading" class="space-y-5">
        <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
          <div
            v-for="item in 5"
            :key="item"
            class="h-28 animate-pulse rounded-xl border border-[#ebe7df] bg-white"
          />
        </div>
        <div
          class="h-64 animate-pulse rounded-xl border border-[#ebe7df] bg-white"
        />
      </section>

      <section
        v-else-if="errorMessage"
        class="flex min-h-[55vh] items-center justify-center"
      >
        <article
          class="w-full max-w-xl rounded-xl border border-[#f1d6d3] bg-white p-6"
        >
          <div class="flex items-start gap-3">
            <div
              class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-[#fff1f0] text-[#dc2626]"
            >
              <PhWarningCircle :size="22" weight="duotone" />
            </div>
            <div class="min-w-0">
              <h2 class="text-base font-medium text-[#171322]">
                Tugas tidak dapat dimuat
              </h2>
              <p class="mt-1 text-sm leading-6 text-[#7a7385]">
                {{ errorMessage }}
              </p>
              <button
                class="mt-4 rounded-lg bg-[#4f46e5] px-4 py-2 text-sm font-medium text-white transition hover:bg-[#4338ca]"
                type="button"
                @click="loadAssignments"
              >
                Coba lagi
              </button>
            </div>
          </div>
        </article>
      </section>

      <template v-else>
        <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
          <article class="rounded-xl border border-[#ebe7df] bg-white p-4">
            <PhClipboardText
              :size="22"
              class="text-[#4f46e5]"
              weight="duotone"
            />
            <p class="mt-3 text-xs text-[#8b8592]">Total tugas</p>
            <p class="mt-1 text-2xl font-medium text-[#171322]">
              {{ summary.totalAssignments }}
            </p>
          </article>
          <article class="rounded-xl border border-[#ebe7df] bg-white p-4">
            <PhClock :size="22" class="text-[#ea580c]" weight="duotone" />
            <p class="mt-3 text-xs text-[#8b8592]">Belum dikumpulkan</p>
            <p class="mt-1 text-2xl font-medium text-[#b45309]">
              {{ summary.notSubmittedCount }}
            </p>
          </article>
          <article class="rounded-xl border border-[#ebe7df] bg-white p-4">
            <PhCheckCircle :size="22" class="text-[#4f46e5]" weight="duotone" />
            <p class="mt-3 text-xs text-[#8b8592]">Sudah dikumpulkan</p>
            <p class="mt-1 text-2xl font-medium text-[#4f46e5]">
              {{ summary.submittedCount }}
            </p>
          </article>
          <article class="rounded-xl border border-[#ebe7df] bg-white p-4">
            <PhSealCheck :size="22" class="text-[#059669]" weight="duotone" />
            <p class="mt-3 text-xs text-[#8b8592]">Sudah dinilai</p>
            <p class="mt-1 text-2xl font-medium text-[#027a48]">
              {{ summary.gradedCount }}
            </p>
          </article>
          <article class="rounded-xl border border-[#ebe7df] bg-white p-4">
            <PhWarningCircle
              :size="22"
              class="text-[#dc2626]"
              weight="duotone"
            />
            <p class="mt-3 text-xs text-[#8b8592]">Lewat deadline</p>
            <p class="mt-1 text-2xl font-medium text-[#dc2626]">
              {{ summary.overdueCount }}
            </p>
          </article>
        </section>

        <section
          class="mt-5 min-w-0 rounded-xl border border-[#ebe7df] bg-white p-4 sm:p-5"
        >
          <div
            class="flex flex-col gap-4 border-b border-[#ebe7df] pb-4 lg:flex-row lg:items-end lg:justify-between"
          >
            <div class="min-w-0">
              <p class="text-sm font-medium text-[#171322]">Daftar tugas</p>
              <p class="mt-1 text-sm text-[#7a7385]">
                {{ items.length }} tugas dari kelas aktif di sekolah ini.
              </p>
            </div>
            <div class="flex min-w-0 flex-wrap gap-2">
              <button
                v-for="tab in filterTabs"
                :key="tab.id"
                type="button"
                class="max-w-full rounded-lg px-3 py-2 text-xs font-medium transition sm:px-4 sm:py-2.5 sm:text-sm"
                :class="
                  activeFilter === tab.id
                    ? 'bg-[#4f46e5] text-white'
                    : 'bg-[#f9fafb] text-[#6b7280] hover:bg-[#eef2ff] hover:text-[#4f46e5]'
                "
                @click="activeFilter = tab.id"
              >
                {{ tab.label }}
                <span class="ml-2 opacity-75">{{ tab.count }}</span>
              </button>
            </div>
          </div>

          <div v-if="items.length === 0" class="py-12 text-center">
            <div
              class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-[#eef2ff] text-[#4f46e5]"
            >
              <PhClipboardText :size="25" weight="duotone" />
            </div>
            <h2 class="mt-3 text-lg font-medium text-[#171322]">
              Belum ada tugas
            </h2>
            <p class="mx-auto mt-2 max-w-xl text-sm leading-6 text-[#7a7385]">
              Tugas akan tampil setelah guru menerbitkan tugas pada mata
              pelajaran di kelas aktifmu.
            </p>
          </div>

          <div v-else-if="filteredItems.length === 0" class="py-12 text-center">
            <div
              class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-[#eef2ff] text-[#4f46e5]"
            >
              <PhCheckCircle :size="25" weight="duotone" />
            </div>
            <h2 class="mt-3 text-lg font-medium text-[#171322]">
              Tidak ada tugas pada filter ini
            </h2>
            <p class="mx-auto mt-2 max-w-xl text-sm leading-6 text-[#7a7385]">
              Pilih filter lain untuk melihat tugas yang tersedia.
            </p>
          </div>

          <div v-else class="space-y-3 pt-5">
            <article
              v-for="item in filteredItems"
              :key="`${item.subjectClassId}-${item.assignmentId}`"
              class="min-w-0 rounded-lg border border-[#ebe7df] bg-[#fbfaf8] p-4 sm:p-5"
            >
              <div
                class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between"
              >
                <div class="min-w-0">
                  <div class="flex flex-wrap gap-2 text-xs font-medium">
                    <span
                      class="rounded-full bg-[#eef2ff] px-3 py-1.5 text-[#4f46e5]"
                    >
                      {{ item.subjectName || "Mata pelajaran" }}
                    </span>
                    <span
                      v-if="item.subjectCode"
                      class="rounded-full bg-white px-3 py-1.5 text-[#6b7280]"
                    >
                      {{ item.subjectCode }}
                    </span>
                    <span
                      class="rounded-full bg-white px-3 py-1.5 text-[#6b7280]"
                    >
                      {{ item.className || item.classCode || "Kelas" }}
                    </span>
                  </div>

                  <h2 class="mt-4 text-lg font-medium text-[#171322]">
                    {{ item.assignmentTitle }}
                  </h2>
                  <p class="mt-2 text-sm text-[#7a7385]">
                    <span v-if="item.categoryName">{{
                      item.categoryName
                    }}</span>
                    <span v-if="item.categoryName && item.deadline"> · </span>
                    <span v-if="item.deadline">
                      Deadline {{ formatDate(item.deadline) }}
                    </span>
                    <span v-if="!item.categoryName && !item.deadline">
                      Detail tugas tersedia di halaman tugas.
                    </span>
                  </p>

                  <div class="mt-3 flex flex-wrap items-center gap-2">
                    <span
                      class="rounded-full px-3 py-1.5 text-xs font-medium"
                      :class="statusClasses(item)"
                    >
                      {{ statusLabel(item) }}
                    </span>
                    <span
                      v-if="item.isSubmittedLate"
                      class="rounded-full bg-[#fff7ed] px-3 py-1.5 text-xs font-medium text-[#b45309]"
                    >
                      Dikumpulkan terlambat
                    </span>
                    <span
                      v-if="item.isGraded"
                      class="rounded-full bg-[#ecfdf3] px-3 py-1.5 text-xs font-medium text-[#027a48]"
                    >
                      Nilai {{ formatScore(item.score) }}
                    </span>
                  </div>

                  <p
                    v-if="item.submittedAt"
                    class="mt-3 text-xs text-[#8b8592]"
                  >
                    Dikumpulkan {{ formatDateTime(item.submittedAt) }}
                  </p>
                </div>

                <RouterLink
                  :to="{
                    name: 'student-assignment-detail',
                    params: {
                      sclId: item.subjectClassId,
                      asgId: item.assignmentId,
                    },
                  }"
                  class="inline-flex w-full shrink-0 items-center justify-center gap-2 rounded-lg bg-[#4f46e5] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#4338ca] sm:w-auto"
                >
                  Buka tugas
                  <PhArrowRight :size="16" />
                </RouterLink>
              </div>
            </article>
          </div>
        </section>
      </template>
    </section>
  </main>
</template>
