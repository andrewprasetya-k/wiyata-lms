<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  PhArrowClockwise,
  PhFloppyDisk,
  PhNotebook,
  PhTrash,
} from '@phosphor-icons/vue'
import {
  deleteStudentMaterialNote,
  getStudentMaterialNote,
  saveStudentMaterialNote,
} from '../../services/studentNotes'
import { useToastStore } from '../../stores/toast'
import type { StudentMaterialNote } from '../../types/studentNotes'
import { formatDateTime } from '../../utils/date'

const props = defineProps<{
  materialId: string
}>()

const maxLength = 10000
const toast = useToastStore()
const note = ref<StudentMaterialNote | null>(null)
const content = ref('')
const savedContent = ref('')
const isLoading = ref(false)
const isSaving = ref(false)
const isDeleting = ref(false)
const hasLoaded = ref(false)
const errorMessage = ref('')

const normalizedContent = computed(() => content.value.trim())
const isTooLong = computed(() => Array.from(content.value).length > maxLength)
const hasChanges = computed(() => normalizedContent.value !== savedContent.value)
const canSave = computed(
  () =>
    hasLoaded.value &&
    normalizedContent.value.length > 0 &&
    !isTooLong.value &&
    hasChanges.value &&
    !isSaving.value &&
    !isDeleting.value,
)

function getNoteErrorMessage(error: unknown, fallback: string) {
  if (typeof error === 'object' && error !== null && 'response' in error) {
    const response = (
      error as {
        response?: {
          status?: number
          data?: { error?: unknown; message?: unknown }
        }
      }
    ).response

    if (response?.status === 403) {
      return 'Catatan tidak dapat diakses karena materi ini tidak lagi tersedia di kelas aktifmu.'
    }
    if (response?.status === 404) {
      return 'Materi untuk catatan ini tidak ditemukan.'
    }
    if (typeof response?.data?.error === 'string') {
      return response.data.error
    }
    if (typeof response?.data?.message === 'string') {
      return response.data.message
    }
  }
  return fallback
}

async function loadNote() {
  if (!props.materialId) return

  isLoading.value = true
  hasLoaded.value = false
  errorMessage.value = ''

  try {
    const response = await getStudentMaterialNote(props.materialId)
    note.value = response.note
    content.value = response.note?.content ?? ''
    savedContent.value = response.note?.content.trim() ?? ''
    hasLoaded.value = true
  } catch (error) {
    errorMessage.value = getNoteErrorMessage(
      error,
      'Catatan belum bisa dimuat. Coba lagi beberapa saat.',
    )
  } finally {
    isLoading.value = false
  }
}

async function saveNote() {
  if (!canSave.value) {
    if (isTooLong.value) {
      errorMessage.value = 'Catatan maksimal 10.000 karakter.'
    } else if (!normalizedContent.value) {
      errorMessage.value = 'Isi catatan wajib diisi.'
    }
    return
  }

  isSaving.value = true
  errorMessage.value = ''

  try {
    const response = await saveStudentMaterialNote(props.materialId, {
      content: normalizedContent.value,
    })
    note.value = response.note
    content.value = response.note?.content ?? normalizedContent.value
    savedContent.value = content.value.trim()
    toast.success('Catatan berhasil disimpan.')
  } catch (error) {
    errorMessage.value = getNoteErrorMessage(
      error,
      'Catatan belum bisa disimpan. Teks yang kamu tulis tetap tersedia.',
    )
  } finally {
    isSaving.value = false
  }
}

async function deleteNote() {
  if (!note.value || isDeleting.value) return
  if (!window.confirm('Hapus catatan pribadi untuk materi ini?')) return

  isDeleting.value = true
  errorMessage.value = ''

  try {
    await deleteStudentMaterialNote(props.materialId)
    note.value = null
    content.value = ''
    savedContent.value = ''
    toast.success('Catatan berhasil dihapus.')
  } catch (error) {
    errorMessage.value = getNoteErrorMessage(
      error,
      'Catatan belum bisa dihapus.',
    )
  } finally {
    isDeleting.value = false
  }
}

watch(
  () => props.materialId,
  () => {
    void loadNote()
  },
  { immediate: true },
)
</script>

<template>
  <article class="rounded-[22px] border border-[#ebe7df] bg-[#fbfaf8] p-5">
    <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
      <div class="flex items-start gap-3">
        <div
          class="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-[#f3ecff] text-[#7c3aed]"
        >
          <PhNotebook :size="20" weight="duotone" />
        </div>
        <div>
          <h2 class="text-base font-medium text-[#171322]">Catatan Saya</h2>
          <p class="mt-1 max-w-2xl text-sm leading-6 text-[#7a7385]">
            Tulis ringkasan, poin penting, atau hal yang ingin kamu ingat dari materi ini.
          </p>
        </div>
      </div>

      <p
        v-if="note?.updatedAt"
        class="shrink-0 text-xs text-[#a09aa8]"
      >
        Disimpan {{ formatDateTime(note.updatedAt) }}
      </p>
    </div>

    <div v-if="isLoading" class="mt-5 space-y-3">
      <div class="h-40 animate-pulse rounded-[18px] bg-white" />
      <div class="h-10 w-36 animate-pulse rounded-2xl bg-white" />
    </div>

    <div v-else-if="!hasLoaded" class="mt-5 rounded-[18px] bg-white p-4">
      <p class="text-sm leading-6 text-[#b42318]">{{ errorMessage }}</p>
      <button
        class="mt-3 inline-flex items-center gap-2 rounded-2xl border border-[#ebe7df] bg-white px-4 py-2 text-sm font-medium text-[#4f46e5] transition hover:border-[#4f46e5]"
        type="button"
        @click="loadNote"
      >
        <PhArrowClockwise :size="16" />
        Coba lagi
      </button>
    </div>

    <form v-else class="mt-5" @submit.prevent="saveNote">
      <textarea
        v-model="content"
        class="min-h-52 w-full resize-y rounded-[18px] border border-[#ebe7df] bg-white px-5 py-4 text-sm leading-7 text-[#374151] outline-none transition placeholder:text-[#a09aa8] focus:border-[#4f46e5] focus:ring-2 focus:ring-[#4f46e5]/10"
        maxlength="10001"
        placeholder="Mulai tulis catatanmu di sini..."
        aria-label="Catatan pribadi untuk materi"
      />

      <div class="mt-2 flex flex-wrap items-center justify-between gap-2">
        <p
          class="text-xs"
          :class="isTooLong ? 'text-[#b42318]' : 'text-[#a09aa8]'"
        >
          {{ Array.from(content).length.toLocaleString('id-ID') }} / 10.000 karakter
        </p>
        <p v-if="hasChanges && !isTooLong" class="text-xs text-[#8b8592]">
          Perubahan belum disimpan
        </p>
      </div>

      <p
        v-if="errorMessage"
        class="mt-3 rounded-2xl bg-[#fff1f0] px-4 py-3 text-sm leading-6 text-[#b42318]"
      >
        {{ errorMessage }}
      </p>

      <div class="mt-4 flex flex-wrap items-center gap-3">
        <button
          class="inline-flex items-center gap-2 rounded-2xl px-4 py-2 text-sm font-medium text-white transition disabled:cursor-not-allowed disabled:bg-[#d8d5dd]"
          :class="canSave ? 'bg-[#4f46e5] hover:bg-[#4338ca]' : 'bg-[#d8d5dd]'"
          :disabled="!canSave"
          type="submit"
        >
          <PhFloppyDisk :size="17" weight="duotone" />
          {{ isSaving ? 'Menyimpan...' : 'Simpan catatan' }}
        </button>

        <button
          v-if="note"
          class="inline-flex items-center gap-2 rounded-2xl border border-[#ebe7df] bg-white px-4 py-2 text-sm font-medium text-[#b42318] transition hover:border-[#fda29b] hover:bg-[#fff1f0] disabled:cursor-not-allowed disabled:opacity-60"
          :disabled="isSaving || isDeleting"
          type="button"
          @click="deleteNote"
        >
          <PhTrash :size="17" weight="duotone" />
          {{ isDeleting ? 'Menghapus...' : 'Hapus' }}
        </button>
      </div>
    </form>
  </article>
</template>
