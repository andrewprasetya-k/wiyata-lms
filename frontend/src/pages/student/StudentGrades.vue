<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import {
  PhBookOpen,
  PhChartBar,
  PhCheckCircle,
  PhClock,
  PhSealCheck,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { getMyGradebookByClass } from "../../services/studentGrades";
import { useActiveClassStore } from "../../stores/activeClass";
import { useAuthStore } from "../../stores/auth";
import type {
  GradebookAssignment,
  MyGradebookResponse,
} from "../../types/studentGrades";
import { getSubjectColor } from "../../utils/color";
import { formatDateTime } from "../../utils/date";

const auth = useAuthStore();
const activeClassStore = useActiveClassStore();
const gradebook = ref<MyGradebookResponse | null>(null);
const isLoading = ref(true);
const errorMessage = ref("");

const schoolUserId = computed(() => auth.activeSchoolUserId);
const subjects = computed(() => gradebook.value?.subjects ?? []);
const hasAssignments = computed(() =>
  subjects.value.some((subject) => subject.assignments.length > 0),
);

async function loadGrades(selectedClassId?: string) {
  if (!schoolUserId.value) {
    isLoading.value = false;
    errorMessage.value =
      "Konteks sekolah belum tersedia. Silakan login ulang atau pilih sekolah aktif terlebih dahulu.";
    return;
  }

  isLoading.value = true;
  errorMessage.value = "";
  gradebook.value = null;

  try {
    await activeClassStore.loadClasses(schoolUserId.value);
    if (activeClassStore.errorMessage) {
      errorMessage.value = activeClassStore.errorMessage;
      return;
    }

    const classId = selectedClassId ?? activeClassStore.activeClassId;
    if (!classId) {
      return;
    }

    gradebook.value = await getMyGradebookByClass(classId);
  } catch (error) {
    errorMessage.value = getGradebookErrorMessage(error);
  } finally {
    isLoading.value = false;
  }
}

function getGradebookErrorMessage(error: unknown) {
  if (typeof error === "object" && error !== null && "response" in error) {
    const response = (error as { response?: { status?: number } }).response;
    if (response?.status === 403) {
      return "Kamu belum terdaftar sebagai siswa pada kelas ini.";
    }
    if (response?.status === 404) {
      return "Kelas tidak ditemukan.";
    }
  }

  return "Nilai belum bisa dimuat. Periksa koneksi atau coba lagi nanti.";
}

function statusLabel(assignment: GradebookAssignment) {
  if (assignment.status === "graded") return "Sudah dinilai";
  if (assignment.status === "submitted") return "Sudah dikumpulkan";
  return "Belum dikumpulkan";
}

function statusClasses(assignment: GradebookAssignment) {
  if (assignment.status === "graded") return "bg-success-soft text-success";
  if (assignment.status === "submitted") return "bg-brand-soft text-brand";
  return "bg-warning-soft text-warning";
}

function formatScore(value?: number | null) {
  if (value === null || value === undefined) return "Belum tersedia";
  return new Intl.NumberFormat("id-ID", { maximumFractionDigits: 2 }).format(
    value,
  );
}

onMounted(loadGrades);
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 bg-background">
    <header class="border-b border-border bg-surface">
      <div
        class="flex flex-col gap-4 px-5 py-4 sm:flex-row sm:items-center sm:justify-between sm:px-6 lg:px-8"
      >
        <div>
          <h1 class="text-2xl font-semibold text-foreground sm:text-3xl">
            Nilai Saya
          </h1>
          <p class="mt-2 max-w-3xl text-sm leading-6 text-muted">
            Rekap nilai dan feedback untuk kelas aktif.
          </p>
        </div>
      </div>
    </header>

    <section class="mx-auto max-w-screen px-5 py-5 sm:px-6 lg:px-8 lg:py-6">
      <section v-if="isLoading || activeClassStore.isLoading" class="space-y-5">
        <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          <div
            v-for="item in 4"
            :key="item"
            class="h-24 animate-pulse rounded-xl border border-border bg-surface"
          />
        </div>
        <div class="space-y-3">
          <div
            v-for="item in 3"
            :key="item"
            class="h-40 animate-pulse rounded-xl border border-border bg-surface"
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
                Nilai tidak dapat dimuat
              </h2>
              <p class="mt-1 text-sm leading-6 text-muted">
                {{ errorMessage }}
              </p>
              <button
                class="mt-4 rounded-lg bg-brand px-4 py-2 text-sm font-medium text-white transition hover:bg-brand-hover"
                type="button"
                @click="loadGrades()"
              >
                Coba lagi
              </button>
            </div>
          </div>
        </article>
      </section>

      <section
        v-else-if="!activeClassStore.activeClassId"
        class="flex min-h-[55vh] items-center justify-center"
      >
        <article
          class="w-full max-w-xl rounded-xl border border-border bg-surface p-8 text-center"
        >
          <div
            class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-brand-soft text-brand"
          >
            <PhBookOpen class="h-6 w-6" weight="duotone" />
          </div>
          <h2 class="mt-3 text-base font-semibold text-foreground">
            Belum ada kelas aktif
          </h2>
          <p class="mx-auto mt-2 max-w-md text-sm leading-6 text-muted">
            Nilai akan tampil setelah kamu ditempatkan pada kelas aktif.
          </p>
        </article>
      </section>

      <section
        v-else-if="subjects.length === 0"
        class="flex min-h-[55vh] items-center justify-center"
      >
        <article
          class="w-full max-w-xl rounded-xl border border-border bg-surface p-8 text-center"
        >
          <div
            class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-brand-soft text-brand"
          >
            <PhChartBar class="h-6 w-6" weight="duotone" />
          </div>
          <h2 class="mt-3 text-base font-semibold text-foreground">
            Nilai belum tersedia
          </h2>
          <p class="mx-auto mt-2 max-w-md text-sm leading-6 text-muted">
            Belum ada mata pelajaran dengan data pengumpulan atau penilaian pada
            kelas ini.
          </p>
        </article>
      </section>

      <section v-else class="space-y-5">
        <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          <article class="rounded-xl bg-surface px-4 py-3">
            <div class="flex items-center justify-between gap-3">
              <p class="text-xs text-muted">Mata pelajaran</p>
              <PhBookOpen :size="17" class="text-brand" weight="duotone" />
            </div>
            <p class="mt-2 text-2xl font-medium text-foreground">
              {{ gradebook?.summary.subjectCount ?? subjects.length }}
            </p>
          </article>
          <article class="rounded-xl bg-surface px-4 py-3">
            <div class="flex items-center justify-between gap-3">
              <p class="text-xs text-muted">Sudah dinilai</p>
              <PhSealCheck :size="17" class="text-success" weight="duotone" />
            </div>
            <p class="mt-2 text-2xl font-medium text-success">
              {{ gradebook?.summary.gradedAssignmentCount ?? 0 }}
            </p>
          </article>
          <article class="rounded-xl bg-surface px-4 py-3">
            <div class="flex items-center justify-between gap-3">
              <p class="text-xs text-muted">Sudah dikumpulkan</p>
              <PhCheckCircle :size="17" class="text-brand" weight="duotone" />
            </div>
            <p class="mt-2 text-2xl font-medium text-brand">
              {{ gradebook?.summary.submittedAssignmentCount ?? 0 }}
            </p>
          </article>
          <article
            class="rounded-xl border border-border bg-surface shadow-sm px-4 py-3"
          >
            <div class="flex items-center justify-between gap-3">
              <p class="text-xs text-muted">Menunggu nilai</p>
              <PhClock :size="17" class="text-warning" weight="duotone" />
            </div>
            <p class="mt-2 text-2xl font-semibold text-warning">
              {{ gradebook?.summary.pendingAssessmentCount ?? 0 }}
            </p>
          </article>
        </div>

        <article
          v-if="!hasAssignments"
          class="rounded-xl border border-border bg-surface p-6 text-center"
        >
          <div
            class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-brand-soft text-brand"
          >
            <PhChartBar class="h-6 w-6" weight="duotone" />
          </div>
          <h2 class="mt-3 text-base font-semibold text-foreground">
            Belum ada tugas untuk dinilai
          </h2>
          <p class="mt-2 text-sm leading-6 text-muted">
            Mata pelajaran sudah tersedia, tetapi belum memiliki tugas pada
            kelas aktif.
          </p>
        </article>

        <div class="space-y-3">
          <article
            v-for="subject in subjects"
            :key="subject.subjectClassId"
            class="overflow-hidden rounded-xl border border-border bg-surface shadow-sm"
          >
            <header
              class="flex min-w-0 flex-col gap-4 px-4 py-4 sm:flex-row sm:items-center sm:justify-between"
            >
              <div class="flex min-w-0 items-center gap-3">
                <span
                  class="h-10 w-1 shrink-0 rounded-sm"
                  :style="{
                    backgroundColor: getSubjectColor(
                      subject.subjectClassId ||
                        subject.subjectName ||
                        subject.subjectCode,
                    ),
                  }"
                  aria-hidden="true"
                />
                <div class="min-w-0">
                  <h2 class="truncate text-sm font-semibold text-foreground">
                    {{ subject.subjectName || "Mata pelajaran" }}
                  </h2>
                  <p class="mt-0.5 text-[11px] text-muted">
                    {{ subject.subjectCode || "Kode belum tersedia" }}
                    · {{ subject.assignments.length }} tugas
                  </p>
                </div>
              </div>

              <div
                class="flex shrink-0 items-center justify-between gap-4 sm:justify-end"
              >
                <div class="text-right">
                  <p class="text-[10px] uppercase tracking-wide text-muted">
                    Rata-rata berbobot
                  </p>
                  <p class="mt-0.5 text-xl font-medium text-foreground">
                    {{ formatScore(subject.finalGrade) }}
                  </p>
                </div>
              </div>
            </header>

            <div
              class="grid gap-2 border-y border-surface-strong bg-surface-subtle px-4 py-3 sm:grid-cols-3"
            >
              <span class="text-xs text-success">
                <strong class="font-medium">{{ subject.gradedCount }}</strong>
                sudah dinilai
              </span>
              <span class="text-xs text-brand">
                <strong class="font-medium">{{
                  subject.submittedCount
                }}</strong>
                sudah dikumpulkan
              </span>
              <span class="text-xs text-warning">
                <strong class="font-medium">{{ subject.pendingCount }}</strong>
                menunggu nilai
              </span>
            </div>

            <div
              class="border-surface-strong bg-surface px-4 py-3 text-xs leading-5 text-muted"
            >
              <p
                v-if="
                  subject.finalGrade !== null &&
                  subject.finalGrade !== undefined
                "
              >
                Dihitung dari tugas yang sudah dinilai dan bobot kategori yang
                tersedia. Nilai ini masih bersifat sementara.
              </p>
              <p v-else>
                Rata-rata berbobot belum tersedia karena bobot atau nilai belum
                lengkap.
              </p>
            </div>

            <div v-if="subject.assignments.length === 0" class="px-4 py-5">
              <p class="text-sm font-medium text-foreground">Belum ada tugas</p>
              <p class="mt-1 text-sm leading-6 text-muted">
                Nilai akan muncul setelah guru membuat dan menilai tugas pada
                mata pelajaran ini.
              </p>
            </div>

            <div v-else class="divide-y divide-surface-strong">
              <div
                v-for="assignment in subject.assignments"
                :key="assignment.assignmentId"
                class="grid min-w-0 gap-3 px-4 py-4 md:grid-cols-[minmax(0,1fr)_140px] md:items-start"
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
                      Dikumpulkan
                      {{ formatDateTime(assignment.submittedAt) }}
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
        </div>
      </section>
    </section>
  </main>
</template>
