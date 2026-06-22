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
import { getSubjectAssignments } from "../../services/assignment";
import { getSubjectClassSubmissions } from "../../services/teacherAssignment";
import { getMyTeachingSubjectClasses } from "../../services/teacherSubjects";
import type { AssignmentItem } from "../../types/assignment";
import type { TeacherSubmissionGroup } from "../../types/teacherAssignment";
import type { TeacherSubjectClass } from "../../types/teacherSubjects";
import { getSubjectColor } from "../../utils/color";
import { formatDate } from "../../utils/date";

type AssignmentFilter = "all" | "active" | "overdue" | "pending";

interface TeacherAssignmentRow {
  assignment: AssignmentItem;
  subject: TeacherSubjectClass;
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

const summary = computed(() => ({
  total: assignments.value.length,
  active: assignments.value.filter((item) => !item.isOverdue).length,
  overdue: assignments.value.filter((item) => item.isOverdue).length,
  pending: assignments.value.filter((item) => item.pendingCount > 0).length,
  submissions: assignments.value.reduce(
    (total, item) => total + item.submissionCount,
    0,
  ),
}));

const filterTabs = computed(() => [
  { id: "all" as const, label: "Semua", count: assignments.value.length },
  { id: "active" as const, label: "Aktif", count: summary.value.active },
  {
    id: "overdue" as const,
    label: "Lewat deadline",
    count: summary.value.overdue,
  },
  {
    id: "pending" as const,
    label: "Perlu review",
    count: summary.value.pending,
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

  try {
    const subjects = await getMyTeachingSubjectClasses();
    const rows = await Promise.all(
      subjects.map(async (subject) => {
        const [assignmentResponse, submissionResponse] = await Promise.all([
          getSubjectAssignments(subject.subjectClassId, 1, 100),
          getSubjectClassSubmissions(subject.subjectClassId),
        ]);
        const submissionMap = new Map<string, TeacherSubmissionGroup>();
        for (const group of submissionResponse.assignments ?? []) {
          submissionMap.set(group.assignment.assignmentId, group);
        }

        return (assignmentResponse.data?.data ?? []).map((assignment) =>
          mapAssignmentRow(assignment, subject, submissionMap),
        );
      }),
    );

    assignments.value = rows.flat();
  } catch {
    errorMessage.value =
      "Daftar tugas belum bisa dimuat. Coba lagi beberapa saat.";
  } finally {
    loading.value = false;
  }
}

function mapAssignmentRow(
  assignment: AssignmentItem,
  subject: TeacherSubjectClass,
  submissionMap: Map<string, TeacherSubmissionGroup>,
): TeacherAssignmentRow {
  const submission = submissionMap.get(assignment.assignmentId);
  return {
    assignment,
    subject,
    submissionCount: submission?.submissionCount ?? 0,
    pendingCount: submission?.pendingCount ?? 0,
    gradedCount: submission?.gradedCount ?? 0,
    lateCount: submission?.submissions.filter((item) => item.isLate).length ?? 0,
    isOverdue: isPastDeadline(assignment.deadline),
  };
}

function compareAssignments(a: TeacherAssignmentRow, b: TeacherAssignmentRow) {
  const pendingDiff = Number(b.pendingCount > 0) - Number(a.pendingCount > 0);
  if (pendingDiff !== 0) return pendingDiff;

  const deadlineDiff =
    getDeadlineTime(a.assignment.deadline) -
    getDeadlineTime(b.assignment.deadline);
  if (deadlineDiff !== 0) return deadlineDiff;

  return (a.assignment.assignmentTitle || "").localeCompare(
    b.assignment.assignmentTitle || "",
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
  if (item.pendingCount > 0) return "Perlu review";
  if (item.submissionCount > 0) return "Sudah direview";
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
  <main class="min-h-screen flex-1 px-5 py-5 sm:px-6 lg:px-8">
    <section class="flex w-full max-w-none flex-col gap-5">
      <header
        class="rounded-[22px] bg-[#f0e9dd] px-5 py-5 shadow-sm ring-1 ring-black/5 md:px-6"
      >
        <p class="text-sm font-medium text-[#8a6d3b]">Manajemen tugas</p>
        <h1 class="mt-3 text-3xl font-medium text-[#171322] md:text-4xl">
          Tugas Saya
        </h1>
        <p class="mt-3 max-w-2xl text-sm leading-6 text-[#6b6475]">
          Kelola semua tugas dari subject yang Anda ajar. Review pengumpulan
          tetap dilakukan dari halaman review tugas.
        </p>
      </header>

      <section
        v-if="loading"
        class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5"
      >
        <p class="text-sm text-[#6b6475]">Memuat daftar tugas...</p>
      </section>

      <section
        v-else-if="errorMessage"
        class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5"
      >
        <div
          class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"
        >
          <div class="flex items-start gap-3">
            <PhWarningCircle
              :size="24"
              class="mt-0.5 text-[#e58f86]"
              weight="duotone"
            />
            <div>
              <h2 class="text-lg font-medium text-[#171322]">
                Gagal memuat tugas
              </h2>
              <p class="mt-2 text-sm leading-6 text-[#6b6475]">
                {{ errorMessage }}
              </p>
            </div>
          </div>
          <button
            type="button"
            class="rounded-2xl bg-[#171322] px-4 py-3 text-sm font-medium text-white"
            @click="loadAssignments"
          >
            Coba lagi
          </button>
        </div>
      </section>

      <template v-else>
        <section class="grid gap-4 md:grid-cols-4">
          <article
            class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5"
          >
            <PhClipboardText
              :size="24"
              class="text-[#7aa7d9]"
              weight="duotone"
            />
            <p class="mt-4 text-sm text-[#8a8494]">Total tugas</p>
            <p class="mt-1 text-2xl font-medium text-[#171322]">
              {{ summary.total }}
            </p>
          </article>
          <article
            class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5"
          >
            <PhCalendarBlank
              :size="24"
              class="text-[#4f46e5]"
              weight="duotone"
            />
            <p class="mt-4 text-sm text-[#8a8494]">Aktif</p>
            <p class="mt-1 text-2xl font-medium text-[#171322]">
              {{ summary.active }}
            </p>
          </article>
          <article
            class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5"
          >
            <PhWarningCircle
              :size="24"
              class="text-[#e58f86]"
              weight="duotone"
            />
            <p class="mt-4 text-sm text-[#8a8494]">Perlu review</p>
            <p class="mt-1 text-2xl font-medium text-[#171322]">
              {{ summary.pending }}
            </p>
          </article>
          <article
            class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5"
          >
            <PhCheckCircle
              :size="24"
              class="text-[#74bfa5]"
              weight="duotone"
            />
            <p class="mt-4 text-sm text-[#8a8494]">Submission</p>
            <p class="mt-1 text-2xl font-medium text-[#171322]">
              {{ summary.submissions }}
            </p>
          </article>
        </section>

        <section
          class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5"
        >
          <div
            class="flex flex-col gap-4 border-b border-[#ece8df] pb-4 lg:flex-row lg:items-end lg:justify-between"
          >
            <div>
              <p class="text-sm font-medium text-[#171322]">Daftar tugas</p>
              <p class="mt-1 text-sm text-[#8a8494]">
                {{ assignments.length }} tugas dari subject yang Anda ajar.
              </p>
            </div>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="tab in filterTabs"
                :key="tab.id"
                type="button"
                class="rounded-2xl px-4 py-2.5 text-sm font-medium transition"
                :class="
                  activeFilter === tab.id
                    ? 'bg-[#171322] text-white'
                    : 'bg-[#faf8f4] text-[#6b6475] hover:bg-[#f0e9dd] hover:text-[#171322]'
                "
                @click="activeFilter = tab.id"
              >
                {{ tab.label }}
                <span class="ml-2 opacity-70">{{ tab.count }}</span>
              </button>
            </div>
          </div>

          <div v-if="assignments.length === 0" class="py-10 text-center">
            <PhClipboardText
              :size="34"
              class="mx-auto text-[#b5afbf]"
              weight="duotone"
            />
            <h2 class="mt-3 text-lg font-medium text-[#171322]">
              Belum ada tugas
            </h2>
            <p class="mx-auto mt-2 max-w-xl text-sm leading-6 text-[#6b6475]">
              Tugas akan tampil setelah Anda membuat tugas dari subject
              workspace.
            </p>
            <RouterLink
              to="/teacher/create"
              class="mt-5 inline-flex items-center justify-center gap-2 rounded-2xl bg-[#171322] px-4 py-3 text-sm font-medium text-white transition hover:bg-[#2f2b3a]"
            >
              Pilih subject untuk membuat konten
              <PhArrowRight :size="16" />
            </RouterLink>
          </div>

          <div
            v-else-if="filteredAssignments.length === 0"
            class="py-10 text-center"
          >
            <PhCheckCircle
              :size="34"
              class="mx-auto text-[#b5afbf]"
              weight="duotone"
            />
            <h2 class="mt-3 text-lg font-medium text-[#171322]">
              Tidak ada tugas pada filter ini
            </h2>
            <p class="mx-auto mt-2 max-w-xl text-sm leading-6 text-[#6b6475]">
              Ganti filter untuk melihat tugas lain dari subject yang Anda
              ajar.
            </p>
          </div>

          <div v-else class="space-y-3 pt-5">
            <article
              v-for="item in filteredAssignments"
              :key="`${item.subject.subjectClassId}-${item.assignment.assignmentId}`"
              class="rounded-[18px] bg-[#faf8f4] p-5 ring-1 ring-black/5"
            >
              <div
                class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between"
              >
                <div class="min-w-0">
                  <div class="flex flex-wrap gap-2 text-xs font-medium">
                    <span
                      class="rounded-2xl bg-white px-3 py-1.5 text-[#4f46e5]"
                    >
                      {{ item.subject.subjectName }}
                    </span>
                    <span
                      v-if="item.subject.subjectCode"
                      class="rounded-2xl bg-white px-3 py-1.5 text-[#6b6475]"
                    >
                      {{ item.subject.subjectCode }}
                    </span>
                    <span
                      class="rounded-2xl bg-white px-3 py-1.5 text-[#6b6475]"
                    >
                      {{
                        item.subject.className ||
                        item.subject.classCode ||
                        "Kelas"
                      }}
                    </span>
                  </div>

                  <div class="mt-4 flex items-start gap-3">
                    <div
                      class="mt-0.5 flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl text-white"
                      :style="{
                        backgroundColor: getSubjectColor(
                          item.subject.subjectClassId ||
                            item.subject.subjectName ||
                            item.subject.subjectCode,
                        ),
                      }"
                    >
                      <PhBookOpen :size="22" weight="duotone" />
                    </div>
                    <div class="min-w-0">
                      <p
                        class="text-xs font-medium uppercase tracking-wide text-[#e58f86]"
                      >
                        {{ item.assignment.categoryName || "Tanpa kategori" }}
                      </p>
                      <h2 class="mt-1 text-lg font-medium text-[#171322]">
                        {{ item.assignment.assignmentTitle }}
                      </h2>
                      <p
                        v-if="item.assignment.assignmentDescription"
                        class="mt-2 line-clamp-2 text-sm leading-6 text-[#6b6475]"
                      >
                        {{ item.assignment.assignmentDescription }}
                      </p>
                    </div>
                  </div>
                </div>

                <span
                  class="self-start rounded-2xl px-3 py-2 text-xs font-medium"
                  :class="statusClasses(item)"
                >
                  {{ statusLabel(item) }}
                </span>
              </div>

              <div class="mt-5 grid gap-3 sm:grid-cols-4">
                <div class="rounded-2xl bg-white p-4">
                  <p class="text-xs text-[#8a8494]">Deadline</p>
                  <p class="mt-1 text-sm font-medium text-[#171322]">
                    {{ formatDate(item.assignment.deadline) }}
                  </p>
                </div>
                <div class="rounded-2xl bg-white p-4">
                  <p class="text-xs text-[#8a8494]">Submission</p>
                  <p class="mt-1 text-xl font-medium text-[#171322]">
                    {{ item.submissionCount }}
                  </p>
                </div>
                <div class="rounded-2xl bg-white p-4">
                  <p class="text-xs text-[#8a8494]">Perlu review</p>
                  <p class="mt-1 text-xl font-medium text-[#171322]">
                    {{ item.pendingCount }}
                  </p>
                </div>
                <div class="rounded-2xl bg-white p-4">
                  <p class="text-xs text-[#8a8494]">Sudah dinilai</p>
                  <p class="mt-1 text-xl font-medium text-[#171322]">
                    {{ item.gradedCount }}
                  </p>
                </div>
              </div>

              <div class="mt-5 flex flex-wrap gap-2">
                <RouterLink
                  :to="{
                    name: 'teacher-assignment-edit',
                    params: {
                      subjectClassId: item.subject.subjectClassId,
                      asgId: item.assignment.assignmentId,
                    },
                  }"
                  class="inline-flex items-center justify-center gap-2 rounded-2xl border border-[#ebe7df] bg-white px-4 py-3 text-sm font-medium text-[#171322] transition hover:border-[#4f46e5] hover:text-[#4f46e5]"
                >
                  <PhPencilSimple :size="16" weight="bold" />
                  Edit
                </RouterLink>
                <RouterLink
                  :to="{
                    name: 'teacher-assignment-review',
                    params: { assignmentId: item.assignment.assignmentId },
                  }"
                  class="inline-flex items-center justify-center gap-2 rounded-2xl bg-[#171322] px-4 py-3 text-sm font-medium text-white transition hover:bg-[#2f2b3a]"
                >
                  Review pengumpulan
                  <PhArrowRight :size="16" />
                </RouterLink>
                <RouterLink
                  :to="{
                    name: 'teacher-subject-detail',
                    params: { subjectClassId: item.subject.subjectClassId },
                  }"
                  class="inline-flex items-center justify-center gap-2 rounded-2xl bg-white px-4 py-3 text-sm font-medium text-[#6b6475] transition hover:text-[#171322]"
                >
                  Lihat di workspace
                </RouterLink>
              </div>
            </article>
          </div>
        </section>
      </template>
    </section>
  </main>
</template>
