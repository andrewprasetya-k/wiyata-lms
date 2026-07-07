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
import DiscussionComments from "../../components/discussion/DiscussionComments.vue";
import { useAuthStore } from "../../stores/auth";
import { useToastStore } from "../../stores/toast";
import {
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
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 bg-[#f8f7f4]">
    <header class="border-b border-[#ebe7df] bg-white">
      <div
        class="flex min-w-0 items-center gap-2 px-5 py-5 text-xs text-[#6b7280] sm:px-6 lg:px-8"
      >
        <RouterLink
          class="inline-flex shrink-0 items-center gap-1.5 transition hover:text-[#4f46e5]"
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
        <span class="min-w-0 truncate font-medium text-[#171322]">
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
          class="h-52 animate-pulse rounded-xl border border-[#ebe7df] bg-white"
        />
        <div
          class="h-72 animate-pulse rounded-xl border border-[#ebe7df] bg-white"
        />
      </div>
      <div
        class="h-96 animate-pulse rounded-xl border border-[#ebe7df] bg-white"
      />
    </section>

    <section
      v-else-if="errorMessage"
      class="flex min-h-[calc(100vh-49px)] items-center justify-center px-5 py-10"
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
          <div>
            <h1 class="text-base font-medium text-[#171322]">
              Tidak bisa memuat tugas
            </h1>
            <p class="mt-1 text-sm leading-6 text-[#7a7385]">
              {{ errorMessage }}
            </p>
            <button
              class="mt-4 rounded-lg bg-[#4f46e5] px-4 py-2 text-sm font-medium text-white transition hover:bg-[#4338ca]"
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
        class="w-full max-w-xl rounded-xl border border-[#ebe7df] bg-white p-6 text-center"
      >
        <div
          class="mx-auto flex h-11 w-11 items-center justify-center rounded-xl bg-[#eef2ff] text-[#4f46e5]"
        >
          <PhClipboardText :size="22" weight="duotone" />
        </div>
        <h1 class="mt-4 text-base font-medium text-[#171322]">
          Tugas tidak ditemukan
        </h1>
        <p class="mt-1 text-sm leading-6 text-[#7a7385]">
          Tugas ini tidak tersedia atau sudah tidak dapat diakses.
        </p>
        <RouterLink
          class="mt-5 inline-flex items-center gap-2 rounded-lg bg-[#4f46e5] px-4 py-2 text-sm font-medium text-white transition hover:bg-[#4338ca]"
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
        <article class="rounded-xl border border-[#ebe7df] bg-white p-5 sm:p-6">
          <div class="flex min-w-0 items-start gap-4">
            <div
              class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-[#eef2ff] text-[#4f46e5]"
            >
              <PhClipboardText :size="22" weight="duotone" />
            </div>
            <div class="min-w-0 flex-1">
              <div class="flex flex-wrap items-center gap-2">
                <span
                  v-if="assignment.categoryName"
                  class="rounded-full bg-[#eef2ff] px-2.5 py-1 text-[11px] font-medium text-[#4f46e5]"
                >
                  {{ assignment.categoryName }}
                </span>
                <span
                  v-if="assignment.subjectName || assignment.subjectCode"
                  class="rounded-full bg-[#f8f7f4] px-2.5 py-1 text-[11px] text-[#6b7280]"
                >
                  {{ assignment.subjectName || assignment.subjectCode }}
                </span>
              </div>
              <h1
                class="mt-3 wrap-break-word text-xl font-medium leading-7 text-[#171322] sm:text-2xl"
              >
                {{ assignment.assignmentTitle }}
              </h1>
              <div
                class="mt-3 flex flex-wrap items-center gap-x-4 gap-y-2 text-xs text-[#6b7280]"
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
            <h2 class="text-sm font-medium text-[#171322]">Instruksi tugas</h2>
            <p
              v-if="assignment.assignmentDescription"
              class="mt-3 whitespace-pre-line wrap-break-word text-sm leading-7 text-[#4a4356]"
            >
              {{ assignment.assignmentDescription }}
            </p>
            <p v-else class="mt-3 text-sm leading-6 text-[#7a7385]">
              Instruksi tugas belum tersedia.
            </p>
          </div>
        </article>

        <article class="rounded-xl border border-[#ebe7df] bg-white p-5 sm:p-6">
          <div class="flex items-center gap-2">
            <PhPaperclip :size="18" class="text-[#4f46e5]" />
            <h2 class="text-sm font-medium text-[#171322]">
              Lampiran dari guru
            </h2>
          </div>
          <p class="mt-1 text-xs leading-5 text-[#7a7385]">
            Buka file pendukung yang disertakan pada tugas ini.
          </p>
          <AttachmentPreviewList
            class="mt-4"
            :attachments="assignment.attachments"
            empty-text="Tugas ini tidak memiliki lampiran."
          />
        </article>

        <DiscussionComments
          source-type="assignment"
          :source-id="assignment.assignmentId"
          title="Diskusi tugas"
          placeholder="Tulis pertanyaan atau komentar tentang tugas ini..."
          empty-text="Belum ada diskusi untuk tugas ini."
        />
      </div>

      <aside class="min-w-0 lg:sticky lg:top-6">
        <article class="rounded-xl border border-[#ebe7df] bg-white p-5">
          <div class="flex items-start justify-between gap-3">
            <div>
              <p class="text-sm font-medium text-[#171322]">
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
                  ? 'bg-[#eef2ff] text-[#4f46e5]'
                  : submissionStatus.status === 'submitted'
                    ? 'bg-[#ecfdf3] text-[#027a48]'
                    : 'bg-[#fff7ed] text-[#ea580c]'
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
            class="mt-4 divide-y divide-[#f0ede8] rounded-lg bg-[#fbfaf8] px-3"
          >
            <div class="flex items-start justify-between gap-4 py-3">
              <dt class="text-xs text-[#7a7385]">Tenggat</dt>
              <dd class="text-right text-xs font-medium text-[#171322]">
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
              <dt class="text-xs text-[#7a7385]">Dibuat</dt>
              <dd class="text-right text-xs font-medium text-[#171322]">
                {{ formatDateTime(assignment.createdAt) }}
              </dd>
            </div>
          </dl>

          <div
            v-if="isSubmissionLoading"
            class="mt-4 h-28 animate-pulse rounded-xl bg-[#fbfaf8]"
          />

          <div
            v-else-if="submissionError"
            class="mt-4 rounded-xl bg-[#fff1f0] p-4"
          >
            <p class="text-sm leading-6 text-[#b42318]">
              {{ submissionError }}
            </p>
            <button
              class="mt-3 rounded-lg bg-white px-3 py-2 text-xs font-medium text-[#4f46e5] transition hover:bg-[#eef2ff]"
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
            <div class="rounded-xl border border-[#d1fae5] bg-[#ecfdf3] p-4">
              <div class="flex items-start gap-3">
                <PhCheckCircle
                  :size="20"
                  class="mt-0.5 shrink-0 text-[#027a48]"
                  weight="duotone"
                />
                <div>
                  <p class="text-sm font-medium text-[#171322]">
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
                    <template v-else>
                      Status pengumpulan tersimpan.
                    </template>
                  </p>
                </div>
              </div>
            </div>

            <div
              v-if="
                submissionStatus.status === 'graded' &&
                submissionStatus.submission?.assessment
              "
              class="rounded-xl border border-[#c7d2fe] bg-[#eef2ff] p-4"
            >
              <p class="text-xs font-medium text-[#4f46e5]">
                Nilai dan feedback
              </p>
              <p class="mt-2 text-3xl font-medium text-[#171322]">
                {{ submissionStatus.submission.assessment.score }}
              </p>
              <p
                v-if="submissionStatus.submission.assessment.feedback"
                class="mt-3 border-t border-[#c7d2fe] pt-3 whitespace-pre-line wrap-break-word text-sm leading-6 text-[#4a4356]"
              >
                {{ submissionStatus.submission.assessment.feedback }}
              </p>
              <p class="mt-3 text-[11px] leading-5 text-[#7a7385]">
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
              <p class="text-xs font-medium text-[#171322]">
                File yang dikumpulkan
              </p>
              <AttachmentPreviewList
                class="mt-3"
                :attachments="submissionStatus.submission.attachments"
              />
            </div>
          </div>

          <template v-else>
            <div
              class="mt-4 rounded-xl border border-dashed border-[#d8d2c8] bg-[#fbfaf8] p-4 text-center"
            >
              <PhPaperclip
                :size="24"
                class="mx-auto text-[#9ca3af]"
                weight="duotone"
              />
              <p class="mt-2 text-sm font-medium text-[#3f3a4a]">
                Pilih file jawaban
              </p>
              <p class="mt-1 text-xs leading-5 text-[#8b8592]">
                Kamu dapat memilih lebih dari satu file.
              </p>
              <label
                class="mt-3 inline-flex cursor-pointer items-center gap-2 rounded-lg border border-[#ddd8e4] bg-white px-3 py-2 text-xs font-medium text-[#4f46e5] transition hover:bg-[#eef2ff]"
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
                class="flex max-w-full items-center justify-between gap-3 overflow-hidden rounded-lg border border-[#ebe7df] bg-white px-3 py-3"
              >
                <div class="min-w-0 flex-1 overflow-hidden">
                  <p class="truncate text-xs font-medium text-[#3f3a4a]">
                    {{ file.name }}
                  </p>
                  <p class="mt-1 text-[11px] text-[#8b8592]">
                    {{ formatFileSize(file.size) }}
                  </p>
                </div>
                <button
                  class="shrink-0 rounded-lg p-2 text-[#dc2626] transition hover:bg-[#fff1f0]"
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
              class="mt-4 rounded-lg bg-[#fff1f0] p-3 text-sm leading-5 text-[#b42318]"
            >
              {{ submitError }}
            </p>

            <button
              class="mt-4 w-full rounded-lg px-4 py-2.5 text-sm font-medium text-white transition"
              :class="
                isSubmitting || selectedFiles.length === 0
                  ? 'cursor-not-allowed bg-[#d8d5dd]'
                  : 'bg-[#4f46e5] hover:bg-[#4338ca]'
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
