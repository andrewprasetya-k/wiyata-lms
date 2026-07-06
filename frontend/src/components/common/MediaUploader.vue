<script setup lang="ts">
import { ref } from 'vue'
import { PhFile, PhTrash, PhUploadSimple } from '@phosphor-icons/vue'
import { deleteMedia, uploadMediaFile } from '../../services/media'
import { useToastStore } from '../../stores/toast'

interface InitialMedia {
  mediaId: string
  mediaName: string
  fileSize?: number
  fileUrl?: string
}

interface Props {
  schoolId: string
  ownerType?: string
  maxSizeMb?: number
  limit?: number
  cleanupOnRemove?: boolean
  initialMedia?: InitialMedia[]
}

const props = withDefaults(defineProps<Props>(), {
  ownerType: 'material',
  maxSizeMb: 10,
  limit: 5,
  cleanupOnRemove: false,
  initialMedia: () => [],
})

const toast = useToastStore()

const emit = defineEmits<{
  (e: 'update:mediaIds', ids: string[]): void
  (e: 'update:isUploading', value: boolean): void
  (e: 'update:hasUploadError', value: boolean): void
}>()

const files = ref<{
  id?: string
  name: string
  size: number
  progress: number
  status: 'pending' | 'uploading' | 'success' | 'error'
  errorMessage?: string
  mediaId?: string
  fileUrl?: string
  isInitial?: boolean
}[]>([])

const mediaIds = ref<string[]>([])

import { watch } from 'vue'

watch(() => props.initialMedia, (newVal) => {
  if (newVal && newVal.length > 0 && files.value.length === 0) {
    files.value = newVal.map(m => ({
      name: m.mediaName || 'Lampiran',
      size: m.fileSize || 0,
      progress: 100,
      status: 'success' as const,
      mediaId: m.mediaId,
      fileUrl: m.fileUrl,
      isInitial: true
    }))
    mediaIds.value = newVal.map(m => m.mediaId)
    // Emit initial mediaIds so parent state is synced
    emit('update:mediaIds', [...mediaIds.value])
  }
}, { immediate: true })

function emitUploadState() {
  emit('update:isUploading', files.value.some((file) => file.status === 'uploading'))
  emit('update:hasUploadError', files.value.some((file) => file.status === 'error'))
}

async function handleFileChange(event: Event) {
  const target = event.target as HTMLInputElement
  if (!target.files) return

  const newFiles = Array.from(target.files)
  
  if (files.value.length + newFiles.length > props.limit) {
    toast.error(`Maksimal ${props.limit} file yang dapat diunggah.`)
    return
  }

  for (const file of newFiles) {
    if (file.size > props.maxSizeMb * 1024 * 1024) {
      toast.error(`File ${file.name} melebihi batas ${props.maxSizeMb}MB.`)
      continue
    }

    const fileItem = {
      name: file.name,
      size: file.size,
      progress: 0,
      status: 'uploading' as const,
    }
    
    const index = files.value.push(fileItem) - 1
    emitUploadState()

    try {
      const response = await uploadMediaFile(file, props.schoolId, props.ownerType)
      files.value[index].status = 'success'
      files.value[index].mediaId = response.mediaId
      mediaIds.value.push(response.mediaId)
      emit('update:mediaIds', [...mediaIds.value])
    } catch (error) {
      files.value[index].status = 'error'
      files.value[index].errorMessage = 'Gagal mengunggah file'
    } finally {
      emitUploadState()
    }
  }

  // Reset input
  target.value = ''
}

async function removeFile(index: number) {
  const file = files.value[index]
  if (file.mediaId) {
    try {
      if (props.cleanupOnRemove && !file.isInitial) {
        await deleteMedia(file.mediaId)
      }
      mediaIds.value = mediaIds.value.filter(id => id !== file.mediaId)
      emit('update:mediaIds', [...mediaIds.value])
    } catch (error) {
      console.error('Failed to delete media', error)
      toast.error('File belum berhasil dihapus. Coba lagi.')
    }
  }
  files.value.splice(index, 1)
  emitUploadState()
}

function formatSize(bytes: number) {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}
</script>

<template>
  <div class="media-uploader">
    <label class="flex h-32 w-full cursor-pointer flex-col items-center justify-center rounded-xl border border-dashed border-[#ebe7df] bg-[#fbfaf8] transition hover:border-[#c7d2fe] hover:bg-white focus-within:border-[#4f46e5] focus-within:bg-white focus-within:ring-2 focus-within:ring-[#4f46e5]/15">
      <div class="flex flex-col items-center justify-center px-4 py-5 text-center">
        <PhUploadSimple :size="30" class="mb-2 text-[#8b8592]" />
        <p class="mb-1 text-sm font-medium text-[#171322]">Klik untuk unggah atau seret file</p>
        <p class="text-xs text-[#7a7385]">PDF, video, atau dokumen (maks. {{ maxSizeMb }}MB)</p>
      </div>
      <input type="file" class="hidden" multiple @change="handleFileChange" />
    </label>

    <div v-if="files.length > 0" class="mt-4 space-y-2">
      <div v-for="(file, index) in files" :key="index" class="flex max-w-full items-center justify-between gap-3 overflow-hidden rounded-xl border border-[#ebe7df] bg-white p-3">
        <div class="flex min-w-0 flex-1 items-center gap-3 overflow-hidden">
          <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-[#f3f0ea]">
            <PhFile :size="20" class="text-[#6b6475]" />
          </div>
          <div class="min-w-0 flex-1 overflow-hidden">
            <p class="truncate text-sm font-medium text-[#171322]">{{ file.name }}</p>
            <p class="text-xs text-[#7a7385]">{{ formatSize(file.size) }}</p>
          </div>
        </div>

        <div class="flex shrink-0 items-center gap-3">
          <span v-if="file.status === 'uploading'" class="max-w-28 animate-pulse truncate text-xs text-[#4f46e5]">Mengunggah...</span>
          <span v-else-if="file.status === 'error'" class="max-w-36 truncate text-xs text-[#dc2626]">{{ file.errorMessage }}</span>
          
          <button 
            type="button" 
            @click="removeFile(index)" 
            class="rounded-lg p-1.5 text-[#8b8592] transition hover:bg-[#fff1f0] hover:text-[#dc2626] focus:outline-none focus:ring-2 focus:ring-[#dc2626]/15"
            title="Hapus"
          >
            <PhTrash :size="18" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
