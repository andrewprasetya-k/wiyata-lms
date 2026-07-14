<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink, useRoute } from "vue-router";
import {
  PhArrowLeft,
  PhCheckCircle,
  PhClipboardText,
  PhClock,
  PhSealCheck,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { getStudentGradeDetail } from "../../services/teacherGrades";
import type {
  StudentGradeAssignment,
  StudentGradeDetailResponse,
} from "../../types/teacherGrades";
import { getSubjectColor } from "../../utils/color";
import { formatDateTime } from "../../utils/date";

const route = useRoute();
const classId = computed(() => String(route.params.classId ?? ""));
const subjectId = computed(() => String(route.params.subjectId ?? ""));
const studentId = computed(() => String(route.params.studentId ?? ""));

const detail = ref<StudentGradeDetailResponse | null>(null);
const loading = ref(false);
const errorMessage = ref("");

const breakdown = computed(() => detail.value?.breakdown ?? []);
const assignments = computed(() => detail.value?.assignments ?? []);
const subjectAccentColor = computed(() =>
  getSubjectColor(
    detail.value?.subject.subjectId || detail.value?.subject.subjectName,
  ),
);

async function loadDetail() {
  if (!classId.value || !subjectId.value || !studentId.value) {
    errorMessage.value = "Kelas, mata pelajaran, atau siswa tidak valid.";
    return;
  }

  loading.value = true;
  errorMessage.value = "";
  detail.value = null;

  try {
    detail.value = await getStudentGradeDetail(
      classId.value,
      subjectId.value,
      studentId.value,
    );
  } catch {
    errorMessage.value =
      "Detail nilai siswa belum bisa dimuat. Coba lagi beberapa saat.";
  } finally {
    loading.value = false;
  }
}

function formatScore(value?: number | null) {
  if (value === null || value === undefined) return "Belum tersedia";
  return new Intl.NumberFormat("id-ID", { maximumFractionDigits: 2 }).format(
    value,
  );
}

function statusLabel(assignment: StudentGradeAssignment) {
  if (assignment.status === "graded") return "Sudah dinilai";
  if (assignment.status === "submitted") return "Sudah dikumpulkan";
  return "Belum dikumpulkan";
}

function statusClasses(assignment: StudentGradeAssignment) {
  if (assignment.status === "graded") return "bg-success-soft text-success";
  if (assignment.status === "submitted") return "bg-brand-soft text-brand";
  return "bg-warning-soft text-warning";
}

onMounted(loadDetail);
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 bg-background">
    <header class="border-b border-border bg-surface">
      <div class="px-5 py-5 sm:px-6 lg:px-8">
        <div class="flex min-w-0 items-center gap-2 text-xs text-muted">
          <RouterLink
            :to="{
              name: 'teacher-class-grade-report',
              params: { classId: classId, subjectId: subjectId },
            }"
            class="inline-flex shrink-0 items-center gap-1.5 transition hover:text-brand"
          >
            <PhArrowLeft :size="15" />
            Kembali
          </RouterLink>
          <span class="text-border-strong">/</span>
          <span class="min-w-0 truncate font-medium text-foreground">
            Detail Nilai Siswa
          </span>
        </div>

        <div
          class="mt-4 flex min-w-0 flex-col gap-4 lg:flex-row lg:items-center lg:justify-between"
        >
          <div class="flex min-w-0 items-start gap-3">
            <span
              class="h-10 w-1 shrink-0 rounded-sm"
              :style="{ backgroundColor: subjectAccentColor }"
              aria-hidden="true"
            />
            <div class="min-w-0">
              <h1
                class="truncate text-xl font-semibold text-foreground sm:text-2xl"
              >
                {{
                  detail?.studentName ??
                  (loading ? "Memuat detail nilai..." : "Detail Nilai Siswa")
                }}
              </h1>
              <p class="mt-1 truncate text-xs text-muted sm:text-sm">
                {{ detail?.studentEmail ?? "" }}
              </p>
              <p class="mt-0.5 truncate text-xs text-muted sm:text-sm">
                {{
                  detail
                    ? [detail.subject.subjectName, detail.subject.subjectCode]
                        .filter(Boolean)
                        .join(" · ")
                    : "Breakdown nilai berbobot dan daftar tugas siswa."
                }}
              </p>
            </div>
          </div>

          <RouterLink
            v-if="detail"
            :to="{
              name: 'teacher-student-report',
              params: { classId: classId, studentId: studentId },
            }"
            class="inline-flex w-full items-center justify-center gap-2 rounded-lg border border-border bg-surface px-4 py-2.5 text-sm font-medium text-foreground transition hover:border-brand hover:text-brand lg:w-auto"
          >
            <PhClipboardText :size="17" weight="duotone" />
            View Full Report
          </RouterLink>
        </div>
      </div>
    </header>

    <section class="mx-auto max-w-screen px-5 py-5 sm:px-6 lg:px-8 lg:py-6">
      <section v-if="loading" class="space-y-5">
        <div class="grid gap-3 sm:grid-cols-2">
          <div
            v-for="item in 2"
            :key="item"
            class="h-24 animate-pulse rounded-xl border border-border bg-surface"
          />
        </div>
        <div class="space-y-3">
          <div
            v-for="item in 3"
            :key="item"
            class="h-24 animate-pulse rounded-xl border border-border bg-surface"
          />
        </div>
      </section>

      <section
        v-else-if="errorMessage"
        class="flex min-h-[55vh] items-center justify-center"
      >
        <article
          class="w-full max-w-xl rounded-xl border border-danger-line bg-danger-soft p-6"
        >
          <div class="flex items-start gap-3">
            <div
              class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-danger-soft text-danger"
            >
              <PhWarningCircle :size="22" weight="duotone" />
            </div>
            <div>
              <h2 class="text-base font-semibold text-foreground">
                Detail nilai tidak dapat dimuat
              </h2>
              <p class="mt-1 text-sm leading-6 text-muted">
                {{ errorMessage }}
              </p>
              <button
                class="mt-4 rounded-lg bg-brand px-4 py-2 text-sm font-medium text-white transition hover:bg-brand-hover"
                type="button"
                @click="loadDetail"
              >
                Coba lagi
              </button>
            </div>
          </div>
        </article>
      </section>

      <section v-else-if="detail" class="space-y-5">
        <!-- Card Ringkasan -->
        <div class="grid gap-3 sm:grid-cols-2">
          <article
            class="rounded-xl border border-border bg-surface shadow-sm px-4 py-3"
          >
            <p class="text-xs text-muted">Final Grade</p>
            <p class="mt-2 text-2xl font-medium text-foreground">
              {{ formatScore(detail.finalGrade) }}
            </p>
          </article>
        </div>

        <!-- Category Breakdown -->
        <article
          class="rounded-xl border border-border bg-surface shadow-sm p-4 sm:p-5"
        >
          <h2 class="text-sm font-semibold text-foreground">
            Rincian per Kategori
          </h2>

          <div v-if="breakdown.length === 0" class="mt-4 text-sm text-muted">
            Belum ada bobot kategori yang dikonfigurasi untuk mata pelajaran
            ini.
          </div>

          <div v-else class="mt-4 space-y-3">
            <div
              v-for="category in breakdown"
              :key="category.categoryId"
              class="grid gap-2 rounded-lg bg-surface-subtle px-4 py-3 sm:grid-cols-[minmax(0,1fr)_repeat(3,auto)] sm:items-center"
            >
              <p class="text-sm font-medium text-foreground">
                {{ category.categoryName }}
              </p>
              <p class="text-xs text-muted">
                Rata-rata
                <span class="font-medium text-foreground">{{
                  formatScore(category.averageScore)
                }}</span>
              </p>
              <p class="text-xs text-muted">
                Bobot
                <span class="font-medium text-foreground"
                  >{{ category.weight }}%</span
                >
              </p>
              <p class="text-xs text-muted">
                Skor berbobot
                <span class="font-medium text-foreground">{{
                  formatScore(category.weightedScore)
                }}</span>
                <span class="text-muted"
                  >&nbsp;· {{ category.assignmentCount }} tugas</span
                >
              </p>
            </div>
          </div>
        </article>

        <!-- Assignment List -->
        <article class="rounded-xl border border-border bg-surface shadow-sm">
          <h2
            class="px-4 pt-4 text-sm font-semibold text-foreground sm:px-5 sm:pt-5"
          >
            Daftar Tugas
          </h2>

          <div
            v-if="assignments.length === 0"
            class="px-4 py-8 text-center sm:px-5"
          >
            <div
              class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-brand-soft text-brand"
            >
              <PhClipboardText class="h-6 w-6" weight="duotone" />
            </div>
            <h3 class="mt-3 text-base font-semibold text-foreground">
              Belum ada tugas
            </h3>
            <p class="mt-1 text-sm leading-6 text-muted">
              Nilai akan muncul setelah guru membuat dan menilai tugas pada mata
              pelajaran ini.
            </p>
          </div>

          <div v-else class="mt-4 divide-y divide-surface-strong">
            <div
              v-for="assignment in assignments"
              :key="assignment.assignmentId"
              class="grid min-w-0 gap-3 px-4 py-4 sm:px-5 md:grid-cols-[minmax(0,1fr)_140px] md:items-start"
            >
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <h3 class="text-sm font-medium text-foreground">
                    {{ assignment.assignmentTitle }}
                  </h3>
                  <span
                    class="rounded-full px-2 py-1 text-[10px] font-medium"
                    :class="statusClasses(assignment)"
                  >
                    {{ statusLabel(assignment) }}
                  </span>
                </div>
                <p class="mt-1 text-xs text-muted">
                  {{ assignment.categoryName || "Kategori belum tersedia" }}
                </p>

                <div
                  class="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-[11px] text-muted"
                >
                  <span class="inline-flex items-center gap-1.5">
                    <PhClock :size="13" />
                    {{
                      assignment.deadline
                        ? `Tenggat ${formatDateTime(assignment.deadline)}`
                        : "Tanpa tenggat"
                    }}
                  </span>
                  <span
                    v-if="assignment.submittedAt"
                    class="inline-flex items-center gap-1.5"
                  >
                    <PhCheckCircle :size="13" />
                    Dikumpulkan {{ formatDateTime(assignment.submittedAt) }}
                  </span>
                  <span
                    v-if="assignment.assessedAt"
                    class="inline-flex items-center gap-1.5"
                  >
                    <PhSealCheck :size="13" />
                    Dinilai {{ formatDateTime(assignment.assessedAt) }}
                  </span>
                </div>

                <div
                  v-if="assignment.feedback"
                  class="mt-3 rounded-lg bg-surface-subtle px-3 py-2.5"
                >
                  <p class="text-[10px] font-medium text-muted">Feedback</p>
                  <p
                    class="mt-1 whitespace-pre-line wrap-break-word text-xs leading-5 text-foreground"
                  >
                    {{ assignment.feedback }}
                  </p>
                </div>
              </div>

              <div
                class="flex items-center justify-between rounded-lg bg-surface-subtle px-3 py-3 md:block md:text-right"
              >
                <p class="text-[10px] uppercase tracking-wide text-muted">
                  Skor
                </p>
                <p class="text-lg font-medium text-foreground md:mt-1">
                  {{ formatScore(assignment.score) }}
                </p>
                <p
                  v-if="assignment.assessorName"
                  class="text-[11px] text-muted md:mt-1"
                >
                  oleh {{ assignment.assessorName }}
                </p>
              </div>
            </div>
          </div>
        </article>
      </section>
    </section>
  </main>
</template>
