<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import {
  PhArrowLeft,
  PhCalendarBlank,
  PhCheckCircle,
  PhClipboardText,
  PhFileText,
  PhPaperclip,
  PhTrash,
  PhWarningCircle,
} from '@phosphor-icons/vue'
import { useAuthStore } from '../../stores/auth'
import {
  getMySubmissionByAssignment,
  getSubjectAssignmentDetail,
  submitAssignment,
} from '../../services/assignment'
import { deleteMedia, uploadMediaFile } from '../../services/media'
import type { AssignmentItem, MySubmissionResponse, SubjectClassHeader } from '../../types/assignment'

const route = useRoute()
const auth = useAuthStore()
const subjectClassId = computed(() => String(route.params.sclId ?? ''))
const assignmentId = computed(() => String(route.params.asgId ?? ''))
const assignment = ref<AssignmentItem | null>(null)
const subjectClass = ref<SubjectClassHeader | null>(null)
const isLoading = ref(true)
const errorMessage = ref('')
const didLoad = ref(false)
const selectedFiles = ref<File[]>([])
const submitError = ref('')
const submitSuccess = ref('')
const isSubmitting = ref(false)
const submissionStatus = ref<MySubmissionResponse | null>(null)
const isSubmissionLoading = ref(false)
const submissionError = ref('')

const schoolId = computed(() => auth.activeSchoolId ?? auth.defaultContext?.schoolId ?? '')

async function loadAssignment() {
  if (!subjectClassId.value || !assignmentId.value) {
    isLoading.value = false
    errorMessage.value = 'Konteks tugas tidak lengkap.'
    return
  }

  isLoading.value = true
  errorMessage.value = ''
  didLoad.value = false

  try {
    const data = await getSubjectAssignmentDetail(subjectClassId.value, assignmentId.value)
    subjectClass.value = data.subjectClass
    assignment.value = data.assignment
    didLoad.value = true
    loadMySubmissionStatus()
  } catch {
    errorMessage.value = 'Detail tugas belum bisa dimuat. Periksa koneksi atau coba lagi nanti.'
  } finally {
    isLoading.value = false
  }
}

onMounted(loadAssignment)

async function loadMySubmissionStatus() {
  if (!assignmentId.value) return

  isSubmissionLoading.value = true
  submissionError.value = ''

  try {
    submissionStatus.value = await getMySubmissionByAssignment(assignmentId.value)
  } catch {
    submissionError.value =
      'Status pengumpulan belum bisa dimuat. Detail tugas tetap bisa dibaca.'
  } finally {
    isSubmissionLoading.value = false
  }
}

function handleFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  const files = Array.from(input.files ?? [])
  selectedFiles.value = [...selectedFiles.value, ...files]
  submitError.value = ''
  submitSuccess.value = ''
  input.value = ''
}

function removeFile(index: number) {
  selectedFiles.value = selectedFiles.value.filter((_, itemIndex) => itemIndex !== index)
}

function formatFileSize(size: number) {
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${Math.round(size / 1024)} KB`
  return `${(size / (1024 * 1024)).toFixed(1)} MB`
}

function getErrorMessage(error: unknown) {
  if (typeof error === 'object' && error !== null && 'response' in error) {
    const response = (error as { response?: { data?: { error?: string; message?: string } } })
      .response
    return response?.data?.error ?? response?.data?.message
  }
  return undefined
}

async function handleSubmit() {
  if (!assignment.value) return
  if (!schoolId.value) {
    submitError.value = 'Konteks sekolah belum tersedia. Silakan login ulang.'
    return
  }
  if (selectedFiles.value.length === 0) {
    submitError.value = 'Pilih minimal satu file untuk dikumpulkan.'
    return
  }

  isSubmitting.value = true
  submitError.value = ''
  submitSuccess.value = ''
  const uploadedMediaIds: string[] = []

  try {
    const uploaded = []
    for (const file of selectedFiles.value) {
      const media = await uploadMediaFile(file, schoolId.value, 'submission')
      uploaded.push(media)
      uploadedMediaIds.push(media.mediaId)
    }

    await submitAssignment(assignment.value.assignmentId, {
      schoolId: schoolId.value,
      mediaIds: uploaded.map((item) => item.mediaId),
    })

    selectedFiles.value = []
    submitSuccess.value = 'Tugas berhasil dikumpulkan.'
    await loadMySubmissionStatus()
  } catch (error) {
    if (uploadedMediaIds.length > 0) {
      await Promise.allSettled(
        uploadedMediaIds.map(async (mediaId) => {
          try {
            await deleteMedia(mediaId)
          } catch (cleanupError) {
            console.warn('Failed to cleanup uploaded submission media', mediaId, cleanupError)
          }
        }),
      )
    }

    submitError.value =
      getErrorMessage(error) ??
      'Pengumpulan tugas gagal. Pastikan file valid dan coba lagi nanti.'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <main class="min-h-screen flex-1 px-5 py-6 sm:px-8 lg:px-10">
    <RouterLink
      class="mb-6 inline-flex items-center gap-2 rounded-md bg-white px-4 py-2 text-sm font-medium text-[#4f46e5] transition hover:bg-[#eef2ff]"
      :to="`/student/subjects/${subjectClassId}`"
    >
      <PhArrowLeft :size="18" />
      Kembali ke subject
    </RouterLink>

    <section v-if="isLoading" class="max-w-3xl space-y-3">
      <div class="h-40 animate-pulse rounded-3xl border border-[#ebe7df] bg-white" />
      <div class="h-28 animate-pulse rounded-3xl border border-[#ebe7df] bg-white" />
    </section>

    <section v-else-if="errorMessage" class="soft-card max-w-3xl rounded-3xl p-6">
      <div class="mb-4 flex h-11 w-11 items-center justify-center rounded-2xl bg-[#fff1f0] text-[#f2756a]">
        <PhWarningCircle :size="24" weight="duotone" />
      </div>
      <p class="text-sm font-medium text-[#171322]">Tidak bisa memuat tugas</p>
      <p class="mt-2 text-sm leading-6 text-[#7a7385]">{{ errorMessage }}</p>
      <button
        class="mt-5 rounded-2xl bg-[#4f46e5] px-4 py-2 text-sm font-medium text-white"
        type="button"
        @click="loadAssignment"
      >
        Coba lagi
      </button>
    </section>

    <section v-else-if="didLoad && !assignment" class="soft-card max-w-3xl rounded-3xl p-6">
      <div class="mb-4 flex h-11 w-11 items-center justify-center rounded-2xl bg-[#eef2ff] text-[#4f46e5]">
        <PhClipboardText :size="24" weight="duotone" />
      </div>
      <p class="text-sm font-medium text-[#171322]">Tugas tidak ditemukan</p>
      <p class="mt-2 text-sm leading-6 text-[#7a7385]">
        Assignment ID ini tidak ditemukan pada subject class yang sedang dibuka.
      </p>
    </section>

    <section v-else-if="assignment" class="max-w-3xl space-y-4">
      <article class="soft-card rounded-3xl p-6">
        <div class="mb-6 flex items-start gap-4">
          <div
            class="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-[#eef2ff] text-[#4f46e5]"
          >
            <PhClipboardText :size="24" weight="duotone" />
          </div>
          <div class="min-w-0">
            <p class="text-sm text-[#7a7385]">
              {{ subjectClass?.subjectName || subjectClass?.subjectCode || 'Subject assignment' }}
            </p>
            <h1 class="mt-2 text-3xl font-medium tracking-normal text-[#171322]">
              {{ assignment.assignmentTitle }}
            </h1>
            <p v-if="assignment.categoryName" class="mt-2 text-sm text-[#4f46e5]">
              {{ assignment.categoryName }}
            </p>
          </div>
        </div>

        <div class="grid gap-3 sm:grid-cols-3">
          <div class="rounded-2xl bg-[#fbfaf8] p-4">
            <div class="mb-2 flex items-center gap-2 text-[#4f46e5]">
              <PhCalendarBlank :size="17" />
              <p class="text-xs font-medium">Deadline</p>
            </div>
            <p class="text-sm text-[#3f3a4a]">
              {{ assignment.deadline || 'Belum tersedia' }}
            </p>
          </div>
          <div class="rounded-2xl bg-[#fbfaf8] p-4">
            <p class="text-xs font-medium text-[#7a7385]">Late submission</p>
            <p class="mt-2 text-sm text-[#3f3a4a]">
              {{ assignment.allowLateSubmission ? 'Diizinkan' : 'Tidak diizinkan' }}
            </p>
          </div>
          <div class="rounded-2xl bg-[#fbfaf8] p-4">
            <p class="text-xs font-medium text-[#7a7385]">Dibuat</p>
            <p class="mt-2 text-sm text-[#3f3a4a]">
              {{ assignment.createdAt || 'Belum tersedia' }}
            </p>
          </div>
        </div>

        <div class="mt-6 rounded-2xl bg-white p-4">
          <p class="text-sm font-medium text-[#171322]">Deskripsi</p>
          <p
            v-if="assignment.assignmentDescription"
            class="mt-3 whitespace-pre-line text-sm leading-6 text-[#6b6475]"
          >
            {{ assignment.assignmentDescription }}
          </p>
          <p v-else class="mt-3 text-sm leading-6 text-[#7a7385]">
            Deskripsi tugas belum tersedia.
          </p>
        </div>
      </article>

      <article v-if="assignment.attachments?.length" class="rounded-3xl border border-[#ebe7df] bg-white p-5">
        <p class="text-sm font-medium text-[#171322]">Lampiran</p>
        <div class="mt-3 space-y-2">
          <a
            v-for="attachment in assignment.attachments"
            :key="attachment.mediaId"
            class="flex items-center gap-3 rounded-2xl bg-[#fbfaf8] px-4 py-3 text-sm text-[#4a4356]"
            :href="attachment.fileUrl"
            rel="noreferrer"
            target="_blank"
          >
            <PhFileText :size="18" class="text-[#4f46e5]" />
            <span class="truncate">{{ attachment.mediaName || 'Lampiran tugas' }}</span>
          </a>
        </div>
      </article>

      <article class="rounded-3xl border border-[#ebe7df] bg-white p-5">
        <p class="text-sm font-medium text-[#171322]">Pengumpulan tugas</p>
        <p class="mt-2 text-sm leading-6 text-[#7a7385]">
          Status pengumpulan diambil dari submission milik akun login saat ini.
        </p>

        <div v-if="isSubmissionLoading" class="mt-4 h-24 animate-pulse rounded-2xl bg-[#fbfaf8]" />

        <div v-else-if="submissionError" class="mt-4 rounded-2xl bg-[#fff1f0] p-4">
          <p class="text-sm text-[#b42318]">{{ submissionError }}</p>
          <button
            class="mt-3 rounded-xl bg-white px-3 py-1.5 text-sm font-medium text-[#4f46e5]"
            type="button"
            @click="loadMySubmissionStatus"
          >
            Coba lagi
          </button>
        </div>

        <div
          v-else-if="submissionStatus?.status === 'submitted' || submissionStatus?.status === 'graded'"
          class="mt-4 space-y-4"
        >
          <div class="rounded-2xl bg-[#ecfdf3] p-4">
            <div class="flex items-start gap-3">
              <PhCheckCircle :size="22" class="mt-0.5 shrink-0 text-[#027a48]" weight="duotone" />
              <div>
                <p class="text-sm font-medium text-[#171322]">
                  {{
                    submissionStatus.status === 'graded'
                      ? 'Tugas sudah dinilai'
                      : 'Tugas sudah dikumpulkan'
                  }}
                </p>
                <p class="mt-1 text-sm text-[#667085]">
                  Dikumpulkan: {{ submissionStatus.submission?.submittedAt || 'Waktu tidak tersedia' }}
                </p>
              </div>
            </div>
          </div>

          <div
            v-if="submissionStatus.submission?.attachments?.length"
            class="rounded-2xl bg-[#fbfaf8] p-4"
          >
            <p class="text-sm font-medium text-[#171322]">File yang dikumpulkan</p>
            <div class="mt-3 space-y-2">
              <a
                v-for="attachment in submissionStatus.submission.attachments"
                :key="attachment.mediaId"
                class="flex items-center gap-3 rounded-2xl bg-white px-4 py-3 text-sm text-[#4a4356]"
                :href="attachment.fileUrl"
                rel="noreferrer"
                target="_blank"
              >
                <PhFileText :size="18" class="text-[#4f46e5]" />
                <span class="min-w-0 flex-1 truncate">{{ attachment.mediaName || 'File submission' }}</span>
                <span class="shrink-0 text-xs text-[#8b8592]">{{ formatFileSize(attachment.fileSize) }}</span>
              </a>
            </div>
          </div>

          <div
            v-if="submissionStatus.status === 'graded' && submissionStatus.submission?.assessment"
            class="rounded-2xl bg-[#eef2ff] p-4"
          >
            <p class="text-sm font-medium text-[#171322]">Penilaian</p>
            <p class="mt-3 text-3xl font-medium text-[#4f46e5]">
              {{ submissionStatus.submission.assessment.score }}
            </p>
            <p
              v-if="submissionStatus.submission.assessment.feedback"
              class="mt-3 whitespace-pre-line text-sm leading-6 text-[#4a4356]"
            >
              {{ submissionStatus.submission.assessment.feedback }}
            </p>
            <p class="mt-3 text-xs text-[#8b8592]">
              Dinilai oleh {{ submissionStatus.submission.assessment.assessorName || 'Guru' }}
              · {{ submissionStatus.submission.assessment.assessedAt }}
            </p>
          </div>

          <p
            v-else
            class="rounded-2xl bg-[#fbfaf8] p-4 text-sm leading-6 text-[#7a7385]"
          >
            Menunggu penilaian dari guru.
          </p>
        </div>

        <template v-else>
          <p class="mt-4 text-sm leading-6 text-[#7a7385]">
            Upload file tugas, lalu kirim submission. Identitas siswa diambil dari token login.
          </p>

          <div class="mt-4 rounded-2xl border border-dashed border-[#d8d2c8] bg-[#fbfaf8] p-4">
            <label
              class="inline-flex cursor-pointer items-center gap-2 rounded-2xl bg-white px-4 py-2 text-sm font-medium text-[#4f46e5] transition hover:bg-[#eef2ff]"
            >
              <PhPaperclip :size="18" />
              Pilih file
              <input class="hidden" multiple type="file" @change="handleFileChange" />
            </label>

            <p class="mt-3 text-xs leading-5 text-[#8b8592]">
              File akan diupload ke storage backend terlebih dahulu, lalu media ID dikirim ke endpoint submission.
            </p>
          </div>

          <div v-if="selectedFiles.length > 0" class="mt-4 space-y-2">
            <div
              v-for="(file, index) in selectedFiles"
              :key="`${file.name}-${file.size}-${index}`"
              class="flex items-center justify-between gap-3 rounded-2xl bg-[#fbfaf8] px-4 py-3"
            >
              <div class="min-w-0">
                <p class="truncate text-sm font-medium text-[#3f3a4a]">{{ file.name }}</p>
                <p class="mt-1 text-xs text-[#8b8592]">{{ formatFileSize(file.size) }}</p>
              </div>
              <button
                class="shrink-0 rounded-xl p-2 text-[#f2756a] transition hover:bg-[#fff1f0]"
                type="button"
                @click="removeFile(index)"
              >
                <PhTrash :size="17" />
              </button>
            </div>
          </div>

          <p v-if="submitError" class="mt-4 rounded-2xl bg-[#fff1f0] p-3 text-sm text-[#b42318]">
            {{ submitError }}
          </p>
          <p v-if="submitSuccess" class="mt-4 rounded-2xl bg-[#ecfdf3] p-3 text-sm text-[#027a48]">
            {{ submitSuccess }}
          </p>

          <button
            class="mt-4 rounded-2xl px-4 py-2 text-sm font-medium text-white transition"
            :class="
              isSubmitting || selectedFiles.length === 0
                ? 'bg-[#d8d5dd]'
                : 'bg-[#4f46e5] hover:bg-[#4338ca]'
            "
            :disabled="isSubmitting || selectedFiles.length === 0"
            type="button"
            @click="handleSubmit"
          >
            {{ isSubmitting ? 'Mengumpulkan...' : 'Kumpulkan tugas' }}
          </button>
        </template>
      </article>
    </section>
  </main>
</template>
