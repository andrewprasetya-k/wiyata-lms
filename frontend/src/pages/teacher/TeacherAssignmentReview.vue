<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  PhArrowLeft,
  PhCaretLeft,
  PhCaretRight,
  PhCheckCircle,
  PhClock,
  PhFile,
  PhPaperPlaneTilt,
  PhUser,
} from "@phosphor-icons/vue";
import AttachmentPreviewList from "../../components/common/AttachmentPreviewList.vue";
import InlineFormError from "../../components/common/InlineFormError.vue";
import {
  getAssignmentDetailWithSubmissions,
  assessSubmission,
} from "../../services/teacherAssignment";
import type {
  AssignmentWithSubmissionsResponse,
  TeacherSubmission,
} from "../../types/teacherAssignment";
import { formatDateTime } from "../../utils/date";
import { useToastStore } from "../../stores/toast";

const route = useRoute();
const router = useRouter();
const toast = useToastStore();

const assignmentId = computed(() => String(route.params.assignmentId ?? ""));
const assignment = ref<AssignmentWithSubmissionsResponse["assignment"] | null>(
  null,
);
const submissions = ref<TeacherSubmission[]>([]);
const loading = ref(false);
const submitting = ref(false);
const errorMessage = ref("");
const gradeFormError = ref("");
const activeIndex = ref(0);

const currentSubmission = computed<TeacherSubmission | null>(
  () => submissions.value[activeIndex.value] ?? null,
);

// Grading form state
const score = ref<number | string>("");
const feedback = ref("");

async function loadData() {
  loading.value = true;
  errorMessage.value = "";
  try {
    const data = await getAssignmentDetailWithSubmissions(assignmentId.value);
    assignment.value = data.assignment;
    submissions.value = data.submissions ?? [];

    const targetSubmissionId = route.query.submission as string | undefined;
    if (targetSubmissionId) {
      const idx = submissions.value.findIndex(
        (s) => s.submissionId === targetSubmissionId,
      );
      activeIndex.value = idx >= 0 ? idx : 0;
    } else if (activeIndex.value >= submissions.value.length) {
      activeIndex.value = Math.max(submissions.value.length - 1, 0);
    }

    updateGradingForm();
  } catch (err) {
    console.error("Failed to load assignment review", err);
    errorMessage.value = getLoadErrorMessage(err);
  } finally {
    loading.value = false;
  }
}

function getLoadErrorMessage(error: unknown) {
  if (typeof error === "object" && error !== null && "response" in error) {
    const status = (error as { response?: { status?: number } }).response
      ?.status;
    if (status === 403) return "Anda tidak memiliki akses ke tugas ini.";
    if (status === 404) return "Tugas tidak ditemukan.";
  }
  return "Data tinjauan belum bisa dimuat.";
}

function updateGradingForm() {
  gradeFormError.value = "";
  if (currentSubmission.value?.assessment) {
    score.value = currentSubmission.value.assessment.score;
    feedback.value = currentSubmission.value.assessment.feedback;
  } else {
    score.value = "";
    feedback.value = "";
  }
}

watch(activeIndex, updateGradingForm);

function patchCurrentSubmissionAssessment(
  scoreValue: number,
  feedbackValue: string,
) {
  const submissionId = currentSubmission.value?.submissionId;
  if (!submissionId) return;

  submissions.value = submissions.value.map((submission) =>
    submission.submissionId === submissionId
      ? {
          ...submission,
          assessment: {
            ...submission.assessment,
            score: scoreValue,
            feedback: feedbackValue,
          },
        }
      : submission,
  );
}

async function handleGrade() {
  if (!currentSubmission.value) return;
  gradeFormError.value = "";
  if (score.value === "") {
    gradeFormError.value = "Nilai wajib diisi.";
    return;
  }

  const nextScore = Number(score.value);
  const nextFeedback = feedback.value;

  submitting.value = true;
  try {
    await assessSubmission(currentSubmission.value.submissionId, {
      score: nextScore,
      feedback: nextFeedback,
    });

    patchCurrentSubmissionAssessment(nextScore, nextFeedback);
    toast.success("Nilai berhasil disimpan.");
  } catch (err) {
    toast.error("Gagal menyimpan nilai.");
  } finally {
    submitting.value = false;
  }
}

function nextStudent() {
  if (activeIndex.value < submissions.value.length - 1) {
    activeIndex.value++;
  }
}

function prevStudent() {
  if (activeIndex.value > 0) {
    activeIndex.value--;
  }
}

onMounted(loadData);
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-background">
    <header class="border-b border-border bg-surface">
      <div
        class="flex min-w-0 items-center gap-2 px-5 py-3 text-xs text-muted sm:px-6 lg:px-8"
      >
        <button
          type="button"
          class="inline-flex shrink-0 items-center gap-1.5 transition hover:text-brand"
          @click="router.back()"
        >
          <PhArrowLeft :size="15" />
          Mata pelajaran
        </button>
        <span class="text-border-strong">/</span>
        <span class="shrink-0">Tugas</span>
        <span class="text-border-strong">/</span>
        <span class="min-w-0 truncate font-medium text-foreground">
          Nilai pengumpulan
        </span>
      </div>

      <div
        class="flex min-w-0 flex-col gap-4 border-t border-[#f3f1ec] px-5 py-4 sm:px-6 lg:flex-row lg:items-center lg:justify-between lg:px-8"
      >
        <div class="min-w-0">
          <p class="text-xs font-medium uppercase tracking-wide text-[#7b61a8]">
            Nilai pengumpulan
          </p>
          <h1
            class="mt-1 wrap-break-word text-xl font-semibold text-foreground sm:text-2xl"
          >
            {{ assignment?.assignmentTitle || "Memuat tugas..." }}
          </h1>
          <p class="mt-1 text-sm text-muted">
            Periksa jawaban siswa dan simpan nilai serta umpan balik.
          </p>
        </div>

        <div class="flex min-w-0 items-center gap-2 self-start lg:self-center">
          <button
            type="button"
            class="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border border-border bg-surface text-muted transition hover:bg-surface-strong hover:text-foreground disabled:cursor-not-allowed disabled:opacity-30"
            :disabled="activeIndex === 0"
            title="Siswa sebelumnya"
            @click="prevStudent"
          >
            <PhCaretLeft :size="16" />
          </button>
          <span
            class="min-w-24 rounded-lg bg-[#faf8f4] px-3 py-2 text-center text-xs font-medium text-foreground-secondary"
          >
            {{ submissions.length > 0 ? activeIndex + 1 : 0 }} /
            {{ submissions.length }} siswa
          </span>
          <button
            type="button"
            class="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border border-border bg-surface text-muted transition hover:bg-surface-strong hover:text-foreground disabled:cursor-not-allowed disabled:opacity-30"
            :disabled="
              submissions.length === 0 || activeIndex === submissions.length - 1
            "
            title="Siswa berikutnya"
            @click="nextStudent"
          >
            <PhCaretRight :size="16" />
          </button>
        </div>
      </div>
    </header>

    <section class="px-5 py-5 sm:px-6 lg:px-8">
      <template v-if="loading">
        <div class="grid gap-5 lg:grid-cols-[minmax(0,1fr)_340px]">
          <div class="space-y-4">
            <div
              class="h-28 animate-pulse rounded-xl border border-border bg-surface"
            />
            <div
              class="h-80 animate-pulse rounded-xl border border-border bg-surface"
            />
          </div>
          <div
            class="h-96 animate-pulse rounded-xl border border-border bg-surface"
          />
        </div>
      </template>

      <section
        v-else-if="errorMessage"
        class="mx-auto max-w-xl rounded-xl border border-danger-line bg-danger-soft px-5 py-10 text-center"
      >
        <div
          class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-danger-soft text-danger"
        >
          <PhFile :size="24" weight="duotone" />
        </div>
        <h2 class="mt-3 text-lg font-semibold text-foreground">
          Tinjauan belum tersedia
        </h2>
        <p class="mx-auto mt-2 max-w-md text-sm leading-6 text-muted">
          {{ errorMessage }}
        </p>
        <button
          type="button"
          class="mt-5 rounded-lg bg-foreground px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#2f2b3a]"
          @click="loadData"
        >
          Coba lagi
        </button>
      </section>

      <section
        v-else-if="submissions.length === 0"
        class="mx-auto max-w-screen rounded-xl border border-border bg-surface px-5 py-10 text-center"
      >
        <div
          class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-brand-soft text-brand"
        >
          <PhUser :size="24" weight="duotone" />
        </div>
        <h2 class="mt-3 text-lg font-semibold text-foreground">
          Belum ada pengumpulan
        </h2>
        <p class="mx-auto mt-2 max-w-screen text-sm leading-6 text-muted">
          Belum ada siswa yang mengumpulkan tugas ini. Daftar pengumpulan akan
          muncul setelah siswa mengirim jawaban.
        </p>
      </section>

      <section
        v-else-if="currentSubmission"
        class="grid min-w-0 gap-5 lg:grid-cols-[minmax(0,1fr)_340px]"
      >
        <div class="min-w-0 space-y-5">
          <article
            class="rounded-xl border border-border bg-surface shadow-sm p-5"
          >
            <div
              class="flex min-w-0 flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"
            >
              <div class="flex min-w-0 items-center gap-3">
                <div
                  class="flex h-11 w-11 shrink-0 items-center justify-center rounded-full bg-brand text-sm font-semibold text-white"
                >
                  {{ currentSubmission.studentName?.charAt(0) }}
                </div>
                <div class="min-w-0">
                  <h2 class="truncate text-base font-semibold text-foreground">
                    {{ currentSubmission.studentName }}
                  </h2>
                  <p
                    class="mt-1 flex flex-wrap items-center gap-x-1.5 gap-y-1 text-xs text-muted"
                  >
                    <PhClock :size="14" weight="duotone" />
                    Dikumpulkan
                    {{ formatDateTime(currentSubmission.submittedAt) }}
                  </p>
                </div>
              </div>
              <div class="flex flex-wrap items-center gap-2">
                <span
                  class="rounded-lg px-2.5 py-1.5 text-[11px] font-medium"
                  :class="
                    currentSubmission.isLate
                      ? 'bg-danger-soft text-danger'
                      : 'bg-success-soft text-success'
                  "
                >
                  {{ currentSubmission.isLate ? "Terlambat" : "Tepat waktu" }}
                </span>
                <span
                  v-if="currentSubmission.assessment"
                  class="inline-flex items-center gap-1.5 rounded-lg bg-success-soft px-2.5 py-1.5 text-[11px] font-medium text-success"
                >
                  <PhCheckCircle :size="14" weight="bold" />
                  Sudah dinilai
                </span>
              </div>
            </div>
          </article>

          <article
            class="rounded-xl border border-border bg-surface shadow-sm p-5 sm:p-6"
          >
            <div
              class="flex flex-col gap-2 border-b border-[#f3f1ec] pb-4 sm:flex-row sm:items-center sm:justify-between"
            >
              <div>
                <h2 class="text-base font-semibold text-foreground">
                  Lampiran jawaban
                </h2>
                <p class="mt-1 text-xs text-muted">
                  {{ currentSubmission.attachments?.length || 0 }} lampiran
                  dikirim siswa
                </p>
              </div>
              <PhFile :size="20" class="text-brand" weight="duotone" />
            </div>
            <AttachmentPreviewList
              class="mt-4"
              :attachments="currentSubmission.attachments"
              empty-text="Siswa tidak mengirim lampiran."
            />
          </article>
        </div>

        <aside class="min-w-0">
          <div class="space-y-4 lg:sticky lg:top-6">
            <section
              class="rounded-xl border border-border bg-surface shadow-sm p-5"
            >
              <div class="flex items-center justify-between gap-3">
                <div>
                  <p
                    class="text-[10px] font-medium uppercase tracking-[0.08em] text-muted"
                  >
                    Penilaian
                  </p>
                  <h2 class="mt-1 text-base font-semibold text-foreground">
                    Nilai dan umpan balik
                  </h2>
                </div>
                <span
                  v-if="currentSubmission.assessment"
                  class="rounded-lg bg-success-soft px-2.5 py-1 text-[11px] font-medium text-success"
                >
                  Tersimpan
                </span>
              </div>

              <div class="mt-5">
                <label
                  class="block text-xs font-medium text-muted"
                  for="submission-score"
                >
                  Nilai (0–100)
                </label>
                <div class="mt-2 flex items-end gap-2">
                  <input
                    id="submission-score"
                    v-model="score"
                    type="number"
                    min="0"
                    max="100"
                    class="w-24 rounded-lg border border-border bg-surface-subtle px-3 py-2.5 text-center text-2xl font-semibold text-foreground outline-none transition focus:border-brand focus:bg-surface"
                    placeholder="0"
                  />
                  <span class="mb-2.5 text-sm text-muted">/ 100</span>
                </div>
              </div>

              <div class="mt-5">
                <label
                  class="block text-xs font-medium text-muted"
                  for="submission-feedback"
                >
                  Umpan balik untuk siswa
                </label>
                <textarea
                  id="submission-feedback"
                  v-model="feedback"
                  rows="6"
                  class="mt-2 w-full resize-none rounded-lg border border-border bg-surface-subtle px-3.5 py-3 text-sm leading-6 text-foreground-secondary outline-none transition placeholder:text-muted focus:border-brand focus:bg-surface"
                  placeholder="Tuliskan masukan untuk siswa..."
                />
              </div>

              <InlineFormError :message="gradeFormError" class="mt-3" />
              <button
                type="button"
                class="mt-4 inline-flex w-full items-center justify-center gap-2 rounded-lg bg-brand px-4 py-2.5 text-sm font-medium text-white transition hover:bg-brand-hover disabled:opacity-50"
                :disabled="submitting"
                @click="handleGrade"
              >
                <PhPaperPlaneTilt v-if="!submitting" :size="17" weight="bold" />
                {{
                  submitting
                    ? "Menyimpan..."
                    : currentSubmission.assessment
                      ? "Perbarui nilai"
                      : "Simpan nilai"
                }}
              </button>
            </section>

            <section
              class="rounded-xl border border-border bg-surface shadow-sm p-4"
            >
              <div class="flex items-center justify-between gap-3 px-1">
                <h2 class="text-sm font-semibold text-foreground">
                  Daftar siswa
                </h2>
                <span class="text-xs text-muted">
                  {{ submissions.length }} pengumpulan
                </span>
              </div>
              <div class="mt-3 max-h-80 space-y-2 overflow-y-auto">
                <button
                  v-for="(sub, index) in submissions"
                  :key="sub.submissionId"
                  type="button"
                  class="flex w-full min-w-0 items-center justify-between gap-3 rounded-lg border p-3 text-left transition"
                  :class="
                    activeIndex === index
                      ? 'border-brand-line bg-brand-soft'
                      : 'border-transparent hover:bg-surface-strong'
                  "
                  @click="activeIndex = index"
                >
                  <div class="flex min-w-0 items-center gap-3">
                    <div
                      class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-[11px] font-semibold text-white"
                      :class="
                        activeIndex === index ? 'bg-brand' : 'bg-[#9ca3af]'
                      "
                    >
                      {{ sub.studentName?.charAt(0) }}
                    </div>
                    <div class="min-w-0">
                      <p
                        class="truncate text-xs font-medium"
                        :class="
                          activeIndex === index
                            ? 'text-brand'
                            : 'text-foreground-secondary'
                        "
                      >
                        {{ sub.studentName }}
                      </p>
                      <p class="mt-0.5 text-[10px] text-muted">
                        {{
                          sub.assessment
                            ? `Nilai ${sub.assessment.score}`
                            : "Belum dinilai"
                        }}
                      </p>
                    </div>
                  </div>
                  <div class="flex items-center gap-2">
                    <span
                      :class="sub.assessment ? 'text-success' : 'text-red-500'"
                    >
                      {{ sub.assessment ? "Sudah dinilai" : "Belum dinilai" }}
                    </span>
                  </div>
                </button>
              </div>
            </section>
          </div>
        </aside>
      </section>

      <section
        v-else
        class="mx-auto max-w-xl rounded-xl border border-border bg-surface px-5 py-10 text-center"
      >
        <h2 class="text-lg font-semibold text-foreground">
          Pengumpulan belum dipilih
        </h2>
        <p class="mt-2 text-sm leading-6 text-muted">
          Pilih siswa dari daftar pengumpulan untuk mulai meninjau jawaban.
        </p>
      </section>
    </section>
  </main>
</template>

<style scoped>
/* Chrome, Safari, Edge, Opera */
input::-webkit-outer-spin-button,
input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

/* Firefox */
input[type="number"] {
  -moz-appearance: textfield;
}
</style>
