<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink, useRoute } from "vue-router";
import {
  PhArrowLeft,
  PhCalendarBlank,
  PhCheckCircle,
  PhClipboardText,
  PhPaperclip,
  PhTrash,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import AttachmentPreviewList from "../../components/common/AttachmentPreviewList.vue";
import CommentThread from "../../components/comments/CommentThread.vue";
import { useAuthStore } from "../../stores/auth";
import { useConfirmStore } from "../../stores/confirm";
import { useToastStore } from "../../stores/toast";
import {
  deleteSubmission,
  getMySubmissionByAssignment,
  getStudentAssignmentDetail,
  submitAssignment,
} from "../../services/assignment";
import { deleteMedia, uploadMediaFile } from "../../services/media";
import type {
  AssignmentItem,
  MySubmissionResponse,
} from "../../types/assignment";
import type { MediaUploadResponse } from "../../types/media";
import { formatDateTime } from "../../utils/date";
import { getApiError } from "../../utils/error";

const route = useRoute();
const auth = useAuthStore();
const toast = useToastStore();
const confirm = useConfirmStore();
const subjectClassId = computed(() => String(route.params.sclId ?? ""));
const assignmentId = computed(() => String(route.params.asgId ?? ""));
const assignment = ref<AssignmentItem | null>(null);
const isLoading = ref(true);
const errorMessage = ref("");
const didLoad = ref(false);
const selectedFiles = ref<File[]>([]);
const submitError = ref("");
const isSubmitting = ref(false);
const submissionStatus = ref<MySubmissionResponse | null>(null);
const isSubmissionLoading = ref(false);
const submissionError = ref("");
const isWithdrawing = ref(false);

const isAssignmentOpen = computed(() => {
  if (!assignment.value?.deadline) return true;
  if (assignment.value.allowLateSubmission) return true;
  return new Date(assignment.value.deadline).getTime() > Date.now();
});

const canWithdraw = computed(
  () =>
    submissionStatus.value?.status === "submitted" && isAssignmentOpen.value,
);

const schoolId = computed(
  () => auth.activeSchoolId ?? auth.defaultContext?.schoolId ?? "",
);

async function loadAssignment() {
  if (!subjectClassId.value || !assignmentId.value) {
    isLoading.value = false;
    errorMessage.value = "Konteks tugas tidak lengkap.";
    return;
  }

  isLoading.value = true;
  errorMessage.value = "";
  didLoad.value = false;

  try {
    assignment.value = await getStudentAssignmentDetail(assignmentId.value);
    didLoad.value = true;
    await loadMySubmissionStatus();
  } catch {
    errorMessage.value =
      "Detail tugas belum bisa dimuat. Periksa koneksi atau coba lagi nanti.";
  } finally {
    isLoading.value = false;
  }
}

onMounted(loadAssignment);

async function loadMySubmissionStatus() {
  if (!assignmentId.value) return;

  isSubmissionLoading.value = true;
  submissionError.value = "";

  try {
    submissionStatus.value = await getMySubmissionByAssignment(
      assignmentId.value,
    );
  } catch {
    submissionError.value =
      "Status pengumpulan belum bisa dimuat. Detail tugas tetap bisa dibaca.";
  } finally {
    isSubmissionLoading.value = false;
  }
}

function handleFileChange(event: Event) {
  const input = event.target as HTMLInputElement;
  const files = Array.from(input.files ?? []);
  selectedFiles.value = [...selectedFiles.value, ...files];
  submitError.value = "";
  input.value = "";
}

function removeFile(index: number) {
  selectedFiles.value = selectedFiles.value.filter(
    (_, itemIndex) => itemIndex !== index,
  );
}

function formatFileSize(size: number) {
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${Math.round(size / 1024)} KB`;
  return `${(size / (1024 * 1024)).toFixed(1)} MB`;
}

function patchSubmittedStatus(uploaded: MediaUploadResponse[]) {
  if (!assignment.value) return;

  submissionStatus.value = {
    status: "submitted",
    submission: {
      submissionId: submissionStatus.value?.submission?.submissionId ?? "",
      assignmentId: assignment.value.assignmentId,
      submittedAt: submissionStatus.value?.submission?.submittedAt ?? "",
      assessment: null,
      attachments: uploaded.map((media) => ({
        mediaId: media.mediaId,
        mediaName: media.fileName,
        fileSize: media.fileSize,
        mimeType: media.mimeType,
        fileUrl: media.fileUrl,
      })),
    },
  };
  submissionError.value = "";
  isSubmissionLoading.value = false;
}

async function handleSubmit() {
  if (!assignment.value) return;
  if (!schoolId.value) {
    submitError.value = "Konteks sekolah belum tersedia. Silakan login ulang.";
    return;
  }
  if (selectedFiles.value.length === 0) {
    submitError.value = "Pilih minimal satu file untuk dikumpulkan.";
    return;
  }

  isSubmitting.value = true;
  submitError.value = "";
  const uploadedMediaIds: string[] = [];

  try {
    const uploaded = [];
    for (const file of selectedFiles.value) {
      const media = await uploadMediaFile(file, schoolId.value, "submission");
      uploaded.push(media);
      uploadedMediaIds.push(media.mediaId);
    }

    await submitAssignment(assignment.value.assignmentId, {
      schoolId: schoolId.value,
      mediaIds: uploaded.map((item) => item.mediaId),
    });

    patchSubmittedStatus(uploaded);
    selectedFiles.value = [];
    toast.success("Tugas berhasil dikumpulkan.");
  } catch (error) {
    if (uploadedMediaIds.length > 0) {
      await Promise.allSettled(
        uploadedMediaIds.map(async (mediaId) => {
          try {
            await deleteMedia(mediaId);
          } catch (cleanupError) {
            console.warn(
              "Failed to cleanup uploaded submission media",
              mediaId,
              cleanupError,
            );
          }
        }),
      );
    }

    submitError.value = getApiError(error);
  } finally {
    isSubmitting.value = false;
  }
}

async function handleWithdraw() {
  const submissionId = submissionStatus.value?.submission?.submissionId;
  if (!submissionId) return;

  const ok = await confirm.confirm({
    title: "Tarik pengumpulan?",
    description: "Anda dapat mengumpulkan ulang sebelum tenggat.",
    confirmLabel: "Ya, tarik kembali",
    variant: "warning",
  });
  if (!ok) return;

  isWithdrawing.value = true;
  try {
    await deleteSubmission(submissionId);
    submissionStatus.value = { status: "not_submitted", submission: null };
    toast.success("Submission berhasil ditarik kembali.");
  } catch (error) {
    toast.error(getApiError(error));
  } finally {
    isWithdrawing.value = false;
  }
}
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 bg-background">
    <header class="border-b border-border bg-surface">
      <div
        class="flex min-w-0 items-center gap-2 px-5 py-5 text-xs text-muted sm:px-6 lg:px-8"
      >
        <RouterLink
          class="inline-flex shrink-0 items-center gap-1.5 transition hover:text-brand"
          :to="`/student/subjects/${subjectClassId}`"
        >
          <PhArrowLeft :size="15" />
          Mata pelajaran
        </RouterLink>
        <span class="text-[#d1d5db]">/</span>
        <span
          v-if="assignment?.subjectName || assignment?.subjectCode"
          class="hidden min-w-0 truncate sm:inline"
        >
          {{ assignment.subjectName || assignment.subjectCode }}
        </span>
        <span
          v-if="assignment?.subjectName || assignment?.subjectCode"
          class="hidden text-[#d1d5db] sm:inline"
        >
          /
        </span>
        <span class="min-w-0 truncate font-medium text-foreground">
          {{ assignment?.assignmentTitle || "Detail tugas" }}
        </span>
      </div>
    </header>

    <section
      v-if="isLoading"
      class="grid gap-5 px-5 py-5 sm:px-6 lg:grid-cols-[minmax(0,1fr)_360px] lg:px-8 lg:py-6"
    >
      <div class="space-y-4">
        <div
          class="h-52 animate-pulse rounded-xl border border-border bg-surface"
        />
        <div
          class="h-72 animate-pulse rounded-xl border border-border bg-surface"
        />
      </div>
      <div
        class="h-96 animate-pulse rounded-xl border border-border bg-surface"
      />
    </section>

    <section
      v-else-if="errorMessage"
      class="flex min-h-[calc(100vh-49px)] items-center justify-center px-5 py-10"
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
            <h1 class="text-base font-semibold text-foreground">
              Tidak bisa memuat tugas
            </h1>
            <p class="mt-1 text-sm leading-6 text-muted">
              {{ errorMessage }}
            </p>
            <button
              class="mt-4 rounded-lg bg-brand px-4 py-2 text-sm font-medium text-white transition hover:bg-brand-hover"
              type="button"
              @click="loadAssignment"
            >
              Coba lagi
            </button>
          </div>
        </div>
      </article>
    </section>

    <section
      v-else-if="didLoad && !assignment"
      class="flex min-h-[calc(100vh-49px)] items-center justify-center px-5 py-10"
    >
      <article
        class="w-full max-w-xl rounded-xl border border-border bg-surface p-6 text-center"
      >
        <div
          class="mx-auto flex h-11 w-11 items-center justify-center rounded-xl bg-brand-soft text-brand"
        >
          <PhClipboardText :size="22" weight="duotone" />
        </div>
        <h1 class="mt-4 text-base font-semibold text-foreground">
          Tugas tidak ditemukan
        </h1>
        <p class="mt-1 text-sm leading-6 text-muted">
          Tugas ini tidak tersedia atau sudah tidak dapat diakses.
        </p>
        <RouterLink
          class="mt-5 inline-flex items-center gap-2 rounded-lg bg-brand px-4 py-2 text-sm font-medium text-white transition hover:bg-brand-hover"
          :to="`/student/subjects/${subjectClassId}`"
        >
          <PhArrowLeft :size="16" />
          Kembali ke mata pelajaran
        </RouterLink>
      </article>
    </section>

    <section
      v-else-if="assignment"
      class="mx-auto grid max-w-screen min-w-0 gap-5 px-5 py-5 sm:px-6 lg:grid-cols-[minmax(0,1fr)_360px] lg:items-start lg:px-8 lg:py-6"
    >
      <div class="min-w-0 space-y-4">
        <article
          class="rounded-xl border border-border bg-surface shadow-sm p-5 sm:p-6"
        >
          <div class="flex min-w-0 items-start gap-4">
            <div
              class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-brand-soft text-brand"
            >
              <PhClipboardText :size="22" weight="duotone" />
            </div>
            <div class="min-w-0 flex-1">
              <div class="flex flex-wrap items-center gap-2">
                <span
                  v-if="assignment.categoryName"
                  class="rounded-full bg-brand-soft px-2.5 py-1 text-[11px] font-medium text-brand"
                >
                  {{ assignment.categoryName }}
                </span>
                <span
                  v-if="assignment.subjectName || assignment.subjectCode"
                  class="rounded-full bg-background px-2.5 py-1 text-[11px] text-muted"
                >
                  {{ assignment.subjectName || assignment.subjectCode }}
                </span>
              </div>
              <h1
                class="mt-3 wrap-break-word text-xl font-semibold leading-7 text-foreground sm:text-2xl"
              >
                {{ assignment.assignmentTitle }}
              </h1>
              <div
                class="mt-3 flex flex-wrap items-center gap-x-4 gap-y-2 text-xs text-muted"
              >
                <span class="inline-flex items-center gap-1.5">
                  <PhCalendarBlank :size="15" />
                  {{
                    assignment.deadline
                      ? `Tenggat ${formatDateTime(assignment.deadline)}`
                      : "Tanpa tenggat"
                  }}
                </span>
                <span>
                  {{
                    assignment.allowLateSubmission
                      ? "Dapat dikumpulkan setelah tanggal tenggat"
                      : "Tidak dapat dikumpulkan setelah tanggal tenggat"
                  }}
                </span>
              </div>
            </div>
          </div>

          <div class="mt-6 border-t border-[#f0ede8] pt-5">
            <h2 class="text-sm font-semibold text-foreground">
              Instruksi tugas
            </h2>
            <p
              v-if="assignment.assignmentDescription"
              class="mt-3 whitespace-pre-line wrap-break-word text-sm leading-7 text-foreground"
            >
              {{ assignment.assignmentDescription }}
            </p>
            <p v-else class="mt-3 text-sm leading-6 text-muted">
              Instruksi tugas belum tersedia.
            </p>
          </div>
        </article>

        <article
          class="rounded-xl border border-border bg-surface shadow-sm p-5 sm:p-6"
        >
          <div class="flex items-center gap-2">
            <PhPaperclip :size="18" class="text-brand" />
            <h2 class="text-sm font-semibold text-foreground">
              Lampiran dari guru
            </h2>
          </div>
          <p class="mt-1 text-xs leading-5 text-muted">
            Buka file pendukung yang disertakan pada tugas ini.
          </p>
          <AttachmentPreviewList
            class="mt-4"
            :attachments="assignment.attachments"
            empty-text="Tugas ini tidak memiliki lampiran."
          />
        </article>

        <CommentThread
          source-type="assignment"
          :source-id="assignment.assignmentId"
          title="Diskusi tugas"
          placeholder="Tulis pertanyaan atau komentar tentang tugas ini..."
          empty-text="Belum ada diskusi untuk tugas ini."
        />
      </div>

      <aside class="min-w-0 lg:sticky lg:top-6">
        <article class="rounded-xl border border-border bg-surface shadow-sm p-5">
          <div class="flex items-start justify-between gap-3">
            <div>
              <p class="text-sm font-semibold text-foreground">
                Pengumpulan tugas
              </p>
            </div>
            <span
              v-if="
                !isSubmissionLoading &&
                !submissionError &&
                submissionStatus?.status
              "
              class="shrink-0 rounded-full px-2.5 py-1 text-[10px] font-medium"
              :class="
                submissionStatus.status === 'graded'
                  ? 'bg-brand-soft text-brand'
                  : submissionStatus.status === 'submitted'
                    ? 'bg-success-soft text-success'
                    : 'bg-warning-soft text-[#ea580c]'
              "
            >
              {{
                submissionStatus.status === "graded"
                  ? "Sudah dinilai"
                  : submissionStatus.status === "submitted"
                    ? "Sudah dikumpulkan"
                    : "Belum dikumpulkan"
              }}
            </span>
          </div>

          <dl
            class="mt-4 divide-y divide-[#f0ede8] rounded-lg bg-surface-subtle px-3"
          >
            <div class="flex items-start justify-between gap-4 py-3">
              <dt class="text-xs text-muted">Tenggat</dt>
              <dd class="text-right text-xs font-medium text-foreground">
                {{
                  assignment.deadline
                    ? formatDateTime(assignment.deadline)
                    : "Tidak ada tenggat"
                }}
              </dd>
            </div>
            <div
              v-if="assignment.createdAt"
              class="flex items-start justify-between gap-4 py-3"
            >
              <dt class="text-xs text-muted">Dibuat</dt>
              <dd class="text-right text-xs font-medium text-foreground">
                {{ formatDateTime(assignment.createdAt) }}
              </dd>
            </div>
          </dl>

          <div
            v-if="isSubmissionLoading"
            class="mt-4 h-28 animate-pulse rounded-xl bg-surface-subtle"
          />

          <div
            v-else-if="submissionError"
            class="mt-4 rounded-xl bg-danger-soft p-4"
          >
            <p class="text-sm leading-6 text-danger">
              {{ submissionError }}
            </p>
            <button
              class="mt-3 rounded-lg bg-surface px-3 py-2 text-xs font-medium text-brand transition hover:bg-brand-soft"
              type="button"
              @click="loadMySubmissionStatus"
            >
              Coba lagi
            </button>
          </div>

          <div
            v-else-if="
              submissionStatus?.status === 'submitted' ||
              submissionStatus?.status === 'graded'
            "
            class="mt-4 space-y-4"
          >
            <div class="rounded-xl border border-[#d1fae5] bg-success-soft p-4">
              <div class="flex items-start gap-3">
                <PhCheckCircle
                  :size="20"
                  class="mt-0.5 shrink-0 text-success"
                  weight="duotone"
                />
                <div>
                  <p class="text-sm font-semibold text-foreground">
                    {{
                      submissionStatus.status === "graded"
                        ? "Tugas sudah dinilai"
                        : "Tugas sudah dikumpulkan"
                    }}
                  </p>
                  <p class="mt-1 text-xs leading-5 text-[#667085]">
                    <template v-if="submissionStatus.submission?.submittedAt">
                      Dikumpulkan
                      {{
                        formatDateTime(submissionStatus.submission.submittedAt)
                      }}
                    </template>
                    <template v-else> Status pengumpulan tersimpan. </template>
                  </p>
                </div>
              </div>
            </div>

            <div
              v-if="
                submissionStatus.status === 'graded' &&
                submissionStatus.submission?.assessment
              "
              class="rounded-xl border border-brand-line bg-brand-soft p-4"
            >
              <p class="text-xs font-medium text-brand">Nilai dan feedback</p>
              <p class="mt-2 text-3xl font-medium text-foreground">
                {{ submissionStatus.submission.assessment.score }}
              </p>
              <p
                v-if="submissionStatus.submission.assessment.feedback"
                class="mt-3 border-t border-brand-line pt-3 whitespace-pre-line wrap-break-word text-sm leading-6 text-foreground"
              >
                {{ submissionStatus.submission.assessment.feedback }}
              </p>
              <p class="mt-3 text-[11px] leading-5 text-muted">
                Dinilai oleh
                {{
                  submissionStatus.submission.assessment.assessorName || "Guru"
                }}
                ·
                {{
                  formatDateTime(
                    submissionStatus.submission.assessment.assessedAt,
                  )
                }}
              </p>
            </div>

            <div
              v-if="submissionStatus.submission?.attachments?.length"
              class="min-w-0"
            >
              <p class="text-xs font-medium text-foreground">
                File yang dikumpulkan
              </p>
              <AttachmentPreviewList
                class="mt-3"
                :attachments="submissionStatus.submission.attachments"
              />
            </div>

            <button
              v-if="canWithdraw"
              class="w-full rounded-lg bg-danger cursor-pointer px-3 py-2 text-xs font-medium text-white transition hover:bg-danger-hover disabled:cursor-not-allowed disabled:opacity-60"
              type="button"
              :disabled="isWithdrawing"
              @click="handleWithdraw"
            >
              {{ isWithdrawing ? "Menarik kembali..." : "Tarik kembali" }}
            </button>
          </div>

          <template v-else>
            <div
              class="mt-4 rounded-xl border border-dashed border-border bg-surface-subtle p-4 text-center"
            >
              <PhPaperclip
                :size="24"
                class="mx-auto text-muted"
                weight="duotone"
              />
              <p class="mt-2 text-sm font-medium text-foreground">
                Pilih file jawaban
              </p>
              <p class="mt-1 text-xs leading-5 text-muted">
                Kamu dapat memilih lebih dari satu file.
              </p>
              <label
                class="mt-3 inline-flex cursor-pointer items-center gap-2 rounded-lg border border-border bg-surface px-3 py-2 text-xs font-medium text-brand transition hover:bg-brand-soft"
              >
                <PhPaperclip :size="16" />
                Pilih file
                <input
                  class="hidden"
                  multiple
                  type="file"
                  @change="handleFileChange"
                />
              </label>
            </div>

            <div v-if="selectedFiles.length > 0" class="mt-4 space-y-2">
              <div
                v-for="(file, index) in selectedFiles"
                :key="`${file.name}-${file.size}-${index}`"
                class="flex max-w-full items-center justify-between gap-3 overflow-hidden rounded-lg border border-border bg-surface px-3 py-3"
              >
                <div class="min-w-0 flex-1 overflow-hidden">
                  <p class="truncate text-xs font-medium text-foreground">
                    {{ file.name }}
                  </p>
                  <p class="mt-1 text-[11px] text-muted">
                    {{ formatFileSize(file.size) }}
                  </p>
                </div>
                <button
                  class="shrink-0 rounded-lg p-2 text-danger transition hover:bg-danger-soft"
                  type="button"
                  title="Hapus file"
                  @click="removeFile(index)"
                >
                  <PhTrash :size="16" />
                </button>
              </div>
            </div>

            <p
              v-if="submitError"
              class="mt-4 rounded-lg bg-danger-soft p-3 text-sm leading-5 text-danger"
            >
              {{ submitError }}
            </p>

            <button
              class="mt-4 w-full rounded-lg px-4 py-2.5 text-sm font-medium text-white transition"
              :class="
                isSubmitting || selectedFiles.length === 0
                  ? 'cursor-not-allowed bg-[#d8d5dd]'
                  : 'bg-brand hover:bg-brand-hover'
              "
              :disabled="isSubmitting || selectedFiles.length === 0"
              type="button"
              @click="handleSubmit"
            >
              {{ isSubmitting ? "Mengumpulkan..." : "Kumpulkan tugas" }}
            </button>
          </template>
        </article>
      </aside>
    </section>
  </main>
</template>
