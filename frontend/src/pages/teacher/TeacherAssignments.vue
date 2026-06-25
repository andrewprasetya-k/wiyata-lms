<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import {
  PhArrowRight,
  PhBookOpen,
  PhCalendarBlank,
  PhCheckCircle,
  PhClipboardText,
  PhPencilSimple,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { getTeacherAssignmentInbox } from "../../services/teacherAssignment";
import type {
  TeacherAssignmentInboxItem,
  TeacherAssignmentInboxSummary,
} from "../../types/teacherAssignment";
import { getSubjectColor } from "../../utils/color";
import { formatDate } from "../../utils/date";

type AssignmentFilter = "all" | "active" | "overdue" | "pending";

interface TeacherAssignmentRow {
  item: TeacherAssignmentInboxItem;
  submissionCount: number;
  pendingCount: number;
  gradedCount: number;
  lateCount: number;
  isOverdue: boolean;
}

const loading = ref(false);
const errorMessage = ref("");
const assignments = ref<TeacherAssignmentRow[]>([]);
const activeFilter = ref<AssignmentFilter>("all");
const summary = ref<TeacherAssignmentInboxSummary>({
  totalAssignments: 0,
  activeAssignments: 0,
  overdueAssignments: 0,
  pendingReviewCount: 0,
  totalSubmissions: 0,
});

const filterTabs = computed(() => [
  { id: "all" as const, label: "Semua", count: assignments.value.length },
  {
    id: "active" as const,
    label: "Aktif",
    count: summary.value.activeAssignments,
  },
  {
    id: "overdue" as const,
    label: "Lewat deadline",
    count: summary.value.overdueAssignments,
  },
  {
    id: "pending" as const,
    label: "Perlu dinilai",
    count: assignments.value.filter((item) => item.pendingCount > 0).length,
  },
]);

const filteredAssignments = computed(() => {
  const filtered = assignments.value.filter((item) => {
    if (activeFilter.value === "active") return !item.isOverdue;
    if (activeFilter.value === "overdue") return item.isOverdue;
    if (activeFilter.value === "pending") return item.pendingCount > 0;
    return true;
  });

  return [...filtered].sort(compareAssignments);
});

async function loadAssignments() {
  loading.value = true;
  errorMessage.value = "";
  assignments.value = [];
  summary.value = {
    totalAssignments: 0,
    activeAssignments: 0,
    overdueAssignments: 0,
    pendingReviewCount: 0,
    totalSubmissions: 0,
  };

  try {
    const response = await getTeacherAssignmentInbox();
    summary.value = response.summary;
    assignments.value = (response.items ?? []).map(mapAssignmentRow);
  } catch {
    errorMessage.value =
      "Daftar tugas belum bisa dimuat. Coba lagi beberapa saat.";
  } finally {
    loading.value = false;
  }
}

function mapAssignmentRow(
  item: TeacherAssignmentInboxItem,
): TeacherAssignmentRow {
  return {
    item,
    submissionCount: item.submissionCount,
    pendingCount: item.pendingCount,
    gradedCount: item.gradedCount,
    lateCount: item.lateCount,
    isOverdue: isPastDeadline(item.deadline),
  };
}

function compareAssignments(a: TeacherAssignmentRow, b: TeacherAssignmentRow) {
  const pendingDiff = Number(b.pendingCount > 0) - Number(a.pendingCount > 0);
  if (pendingDiff !== 0) return pendingDiff;

  const deadlineDiff =
    getDeadlineTime(a.item.deadline) - getDeadlineTime(b.item.deadline);
  if (deadlineDiff !== 0) return deadlineDiff;

  return (a.item.assignmentTitle || "").localeCompare(
    b.item.assignmentTitle || "",
  );
}

function getDeadlineTime(deadline?: string | null) {
  if (!deadline) return Number.MAX_SAFE_INTEGER;
  const value = new Date(deadline).getTime();
  return Number.isNaN(value) ? Number.MAX_SAFE_INTEGER : value;
}

function isPastDeadline(deadline?: string | null) {
  const deadlineTime = getDeadlineTime(deadline);
  if (deadlineTime === Number.MAX_SAFE_INTEGER) return false;
  return deadlineTime < Date.now();
}

function statusLabel(item: TeacherAssignmentRow) {
  if (item.pendingCount > 0) return "Perlu dinilai";
  if (item.submissionCount > 0) return "Sudah dinilai";
  if (item.isOverdue) return "Lewat deadline";
  return "Aktif";
}

function statusClasses(item: TeacherAssignmentRow) {
  if (item.pendingCount > 0) return "bg-[#fff7ed] text-[#b45309]";
  if (item.submissionCount > 0) return "bg-[#ecfdf3] text-[#027a48]";
  if (item.isOverdue) return "bg-[#fef2f2] text-[#dc2626]";
  return "bg-[#eef2ff] text-[#4f46e5]";
}

onMounted(loadAssignments);
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-[#f8f7f4]">
    <header class="border-b border-[#ebe7df] bg-white">
      <div class="px-5 py-5 sm:px-6 lg:px-8">
        <h1 class="text-xl font-medium text-[#171322] sm:text-2xl">
          Tugas Saya
        </h1>
        <p class="mt-1 max-w-2xl text-xs leading-5 text-[#6b7280] sm:text-sm">
          Kelola tugas dari semua mata pelajaran. Penilaian tetap dilakukan
          melalui halaman peninjauan pengumpulan.
        </p>
      </div>
    </header>

    <section
      class="mx-auto max-w-screen min-w-0 px-5 py-5 sm:px-6 lg:px-8 lg:py-6"
    >
      <section v-if="loading" class="space-y-5">
        <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
          <div
            v-for="item in 4"
            :key="item"
            class="h-24 animate-pulse rounded-xl border border-[#ebe7df] bg-white"
          />
        </div>
        <div
          class="h-72 animate-pulse rounded-xl border border-[#ebe7df] bg-white"
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
                type="button"
                class="mt-4 rounded-lg bg-[#4f46e5] px-4 py-2 text-sm font-medium text-white transition hover:bg-[#4338ca]"
                @click="loadAssignments"
              >
                Coba lagi
              </button>
            </div>
          </div>
        </article>
      </section>

      <template v-else>
        <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
          <article class="rounded-xl border border-[#ebe7df] bg-white p-4">
            <PhClipboardText
              :size="21"
              class="text-[#4f46e5]"
              weight="duotone"
            />
            <p class="mt-3 text-xs text-[#7a7385]">Total tugas</p>
            <p class="mt-1 text-2xl font-medium text-[#171322]">
              {{ summary.totalAssignments }}
            </p>
          </article>
          <article class="rounded-xl border border-[#ebe7df] bg-white p-4">
            <PhCalendarBlank
              :size="21"
              class="text-[#059669]"
              weight="duotone"
            />
            <p class="mt-3 text-xs text-[#7a7385]">Tugas aktif</p>
            <p class="mt-1 text-2xl font-medium text-[#171322]">
              {{ summary.activeAssignments }}
            </p>
          </article>
          <article class="rounded-xl border border-[#ebe7df] bg-white p-4">
            <PhWarningCircle
              :size="21"
              class="text-[#ea580c]"
              weight="duotone"
            />
            <p class="mt-3 text-xs text-[#7a7385]">Perlu dinilai</p>
            <p class="mt-1 text-2xl font-medium text-[#171322]">
              {{ summary.pendingReviewCount }}
            </p>
          </article>
          <article class="rounded-xl border border-[#ebe7df] bg-white p-4">
            <PhCheckCircle :size="21" class="text-[#4f8ef7]" weight="duotone" />
            <p class="mt-3 text-xs text-[#7a7385]">Total pengumpulan</p>
            <p class="mt-1 text-2xl font-medium text-[#171322]">
              {{ summary.totalSubmissions }}
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
              <h2 class="text-sm font-medium text-[#171322]">Daftar tugas</h2>
              <p class="mt-1 text-xs text-[#7a7385] sm:text-sm">
                {{ assignments.length }} tugas dari mata pelajaran yang Anda
                ajar.
              </p>
            </div>
            <div class="flex min-w-0 flex-wrap gap-2">
              <button
                v-for="tab in filterTabs"
                :key="tab.id"
                type="button"
                class="rounded-lg px-3 py-2 text-xs font-medium transition sm:px-4 sm:text-sm"
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

          <div v-if="assignments.length === 0" class="py-12 text-center">
            <div
              class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-[#eef2ff] text-[#4f46e5]"
            >
              <PhClipboardText :size="25" weight="duotone" />
            </div>
            <h2 class="mt-3 text-lg font-medium text-[#171322]">
              Belum ada tugas
            </h2>
            <p class="mx-auto mt-2 max-w-xl text-sm leading-6 text-[#7a7385]">
              Pastikan Anda masih aktif di Penempatan Kelas untuk mata pelajaran
              yang diajar.
            </p>
            <RouterLink
              to="/teacher/create?type=assignment"
              class="mt-5 inline-flex items-center justify-center gap-2 rounded-lg bg-[#4f46e5] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#4338ca]"
            >
              Pilih mata pelajaran
              <PhArrowRight :size="16" />
            </RouterLink>
          </div>

          <div
            v-else-if="filteredAssignments.length === 0"
            class="py-12 text-center"
          >
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

          <div v-else class="divide-y divide-[#ebe7df] pt-1">
            <article
              v-for="item in filteredAssignments"
              :key="`${item.item.subjectClassId}-${item.item.assignmentId}`"
              class="min-w-0 py-5 first:pt-4 last:pb-0"
            >
              <div
                class="flex min-w-0 flex-col gap-4 lg:flex-row lg:items-start lg:justify-between"
              >
                <div class="min-w-0 flex-1">
                  <div class="flex min-w-0 items-start gap-3">
                    <div
                      class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg text-white"
                      :style="{
                        backgroundColor: getSubjectColor(
                          item.item.subjectClassId ||
                            item.item.subjectName ||
                            item.item.subjectCode,
                        ),
                      }"
                    >
                      <PhBookOpen :size="20" weight="duotone" />
                    </div>
                    <div class="min-w-0">
                      <div class="flex flex-wrap gap-2 text-[11px]">
                        <span class="font-medium text-[#4f46e5]">
                          {{ item.item.subjectName }}
                        </span>
                        <span
                          v-if="item.item.subjectCode"
                          class="text-[#8a8494]"
                        >
                          {{ item.item.subjectCode }}
                        </span>
                        <span class="text-[#8a8494]">
                          {{
                            item.item.className ||
                            item.item.classCode ||
                            "Kelas"
                          }}
                        </span>
                      </div>
                      <h3
                        class="mt-1 line-clamp-2 wrap-break-word text-base font-medium text-[#171322]"
                      >
                        {{ item.item.assignmentTitle }}
                      </h3>
                      <p class="mt-1 text-xs text-[#7a7385]">
                        {{ item.item.categoryName || "Tanpa kategori" }} ·
                        Tenggat {{ formatDate(item.item.deadline) }}
                      </p>
                    </div>
                  </div>
                </div>

                <span
                  class="self-start rounded-full px-3 py-1.5 text-xs font-medium"
                  :class="statusClasses(item)"
                >
                  {{ statusLabel(item) }}
                </span>
              </div>

              <dl class="mt-4 grid grid-cols-3 gap-2 text-xs sm:max-w-lg">
                <div class="rounded-lg bg-[#fbfaf8] p-3">
                  <dt class="text-[#8a8494]">Pengumpulan</dt>
                  <dd class="mt-1 text-base font-medium text-[#171322]">
                    {{ item.submissionCount }}
                  </dd>
                </div>
                <div class="rounded-lg bg-[#fff7ed] p-3">
                  <dt class="text-[#b45309]">Perlu dinilai</dt>
                  <dd class="mt-1 text-base font-medium text-[#b45309]">
                    {{ item.pendingCount }}
                  </dd>
                </div>
                <div class="rounded-lg bg-[#ecfdf5] p-3">
                  <dt class="text-[#027a48]">Sudah dinilai</dt>
                  <dd class="mt-1 text-base font-medium text-[#027a48]">
                    {{ item.gradedCount }}
                  </dd>
                </div>
              </dl>

              <div class="mt-4 flex flex-col gap-2 sm:flex-row sm:flex-wrap">
                <RouterLink
                  :to="{
                    name: 'teacher-assignment-review',
                    params: { assignmentId: item.item.assignmentId },
                  }"
                  class="inline-flex items-center justify-center gap-2 rounded-lg bg-[#4f46e5] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#4338ca]"
                >
                  nilai pengumpulan
                  <PhArrowRight :size="16" />
                </RouterLink>
                <RouterLink
                  :to="{
                    name: 'teacher-assignment-edit',
                    params: {
                      subjectClassId: item.item.subjectClassId,
                      asgId: item.item.assignmentId,
                    },
                  }"
                  class="inline-flex items-center justify-center gap-2 rounded-lg border border-[#ebe7df] bg-white px-4 py-2.5 text-sm font-medium text-[#374151] transition hover:border-[#4f46e5] hover:text-[#4f46e5]"
                >
                  <PhPencilSimple :size="16" />
                  Edit
                </RouterLink>
                <RouterLink
                  :to="{
                    name: 'teacher-subject-detail',
                    params: { subjectClassId: item.item.subjectClassId },
                  }"
                  class="inline-flex items-center justify-center rounded-lg px-4 py-2.5 text-sm font-medium text-[#6b7280] transition hover:bg-[#f9fafb] hover:text-[#171322]"
                >
                  Lihat di ruang kerja
                </RouterLink>
              </div>
            </article>
          </div>
        </section>
      </template>
    </section>
  </main>
</template>
