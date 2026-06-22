<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import {
  PhArrowLeft,
  PhBookOpen,
  PhFileText,
  PhWarningCircle,
  PhPencilSimple,
  PhTrash,
} from '@phosphor-icons/vue'
import { getMaterialById } from '../../services/classWorkspace'
import { deleteMaterial } from '../../services/teacherMaterial'
import type { MaterialItem } from '../../types/classWorkspace'
import { formatDateTime } from '../../utils/date'
import { useToastStore } from '../../stores/toast'

const route = useRoute()
const router = useRouter()
const toast = useToastStore()
const subjectClassId = computed(() => String(route.params.subjectClassId ?? ''))
const materialId = computed(() => String(route.params.matId ?? ''))
const material = ref<MaterialItem | null>(null)
const isLoading = ref(true)
const errorMessage = ref('')
const didLoad = ref(false)
const isDeleting = ref(false)

async function loadMaterial() {
  if (!subjectClassId.value || !materialId.value) {
    isLoading.value = false
    errorMessage.value = 'Konteks materi tidak lengkap.'
    return
  }

  isLoading.value = true
  errorMessage.value = ''
  didLoad.value = false

  try {
    material.value = await getMaterialById(materialId.value)
    didLoad.value = true
  } catch {
    errorMessage.value = 'Detail materi belum bisa dimuat. Periksa koneksi atau coba lagi nanti.'
  } finally {
    isLoading.value = false
  }
}

async function handleDelete() {
  if (!window.confirm('Apakah anda yakin ingin menghapus materi ini? Tindakan ini tidak dapat dibatalkan.')) return
  isDeleting.value = true
  try {
    await deleteMaterial(materialId.value)
    toast.success('Materi berhasil dihapus.')
    router.push(`/teacher/subjects/${subjectClassId.value}`)
  } catch (err: any) {
    const msg = err.response?.data?.error || err.response?.data?.message || 'Gagal menghapus materi.'
    toast.error(msg)
  } finally {
    isDeleting.value = false
  }
}

onMounted(loadMaterial)
</script>

<template>
  <main class="min-h-screen flex-1 px-5 py-5 sm:px-6 lg:px-8">
    <div class="mb-5 flex flex-col sm:flex-row sm:items-center justify-between gap-4 w-full max-w-none">
      <RouterLink
        class="inline-flex items-center gap-2 rounded-md bg-white px-4 py-2 text-sm font-medium text-[#6b6475] transition hover:text-[#171322]"
        :to="`/teacher/subjects/${subjectClassId}`"
      >
        <PhArrowLeft :size="18" />
        Kembali ke workspace
      </RouterLink>

      <div v-if="material" class="flex items-center gap-3">
        <RouterLink
          :to="`/teacher/subjects/${subjectClassId}/materials/${materialId}/edit`"
          class="flex items-center gap-2 rounded-xl bg-white px-4 py-2 text-sm font-medium text-[#4f46e5] border border-[#eef2ff] hover:bg-[#eef2ff] transition"
        >
          <PhPencilSimple :size="16" />
          Edit Materi
        </RouterLink>
        <button
          @click="handleDelete"
          :disabled="isDeleting"
          class="flex items-center gap-2 rounded-xl bg-white px-4 py-2 text-sm font-medium text-[#dc2626] border border-[#fef2f2] hover:bg-[#fef2f2] transition disabled:opacity-50"
        >
          <PhTrash :size="16" />
          {{ isDeleting ? 'Menghapus...' : 'Hapus' }}
        </button>
      </div>
    </div>

    <section v-if="isLoading" class="w-full max-w-none space-y-3">
      <div class="h-40 animate-pulse rounded-3xl border border-[#ebe7df] bg-white" />
      <div class="h-28 animate-pulse rounded-3xl border border-[#ebe7df] bg-white" />
      <div class="h-24 animate-pulse rounded-3xl border border-[#ebe7df] bg-white" />
    </section>

    <section v-else-if="errorMessage" class="w-full max-w-none rounded-[22px] bg-[#faf8f4] p-5">
      <div class="mb-4 flex h-11 w-11 items-center justify-center rounded-2xl bg-[#fff1f0] text-[#f2756a]">
        <PhWarningCircle :size="24" weight="duotone" />
      </div>
      <p class="text-sm font-medium text-[#171322]">Tidak bisa memuat materi</p>
      <p class="mt-2 text-sm leading-6 text-[#6b6475]">{{ errorMessage }}</p>
      <button
        class="mt-5 rounded-2xl bg-[#171322] px-4 py-2 text-sm font-medium text-white"
        type="button"
        @click="loadMaterial"
      >
        Coba lagi
      </button>
    </section>

    <section v-else-if="didLoad && !material" class="w-full max-w-none rounded-[22px] bg-[#faf8f4] p-5">
      <div class="mb-4 flex h-11 w-11 items-center justify-center rounded-2xl bg-[#eef2ff] text-[#4f46e5]">
        <PhBookOpen :size="24" weight="duotone" />
      </div>
      <p class="text-sm font-medium text-[#171322]">Materi tidak ditemukan</p>
      <p class="mt-2 text-sm leading-6 text-[#6b6475]">
        Material ID ini tidak ditemukan atau bukan bagian dari workspace ini.
      </p>
    </section>

    <section v-else-if="material" class="w-full max-w-none space-y-4">
      <article class="rounded-[22px] bg-[#faf8f4] p-5 ring-1 ring-black/5">
        <div class="mb-5 flex items-start gap-4">
          <div
            class="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-[#eef2ff] text-[#4f46e5]"
          >
            <PhBookOpen :size="24" weight="duotone" />
          </div>
          <div class="min-w-0">
            <p v-if="material.materialType" class="text-sm uppercase text-[#7aa7d9]">
              {{ material.materialType }}
            </p>
            <h1 class="mt-2 text-2xl font-medium tracking-normal text-[#171322]">
              {{ material.materialTitle }}
            </h1>
            <p v-if="material.subjectName" class="mt-1 text-sm text-[#6b6475]">
              {{ material.subjectName }}
            </p>
          </div>
        </div>

        <div class="grid gap-3 sm:grid-cols-2">
          <div class="rounded-2xl bg-white p-4">
            <p class="text-xs font-medium text-[#8a8494]">Dibuat oleh</p>
            <p class="mt-2 text-sm text-[#374151]">
              {{ material.creatorName || 'Creator tidak tersedia' }}
            </p>
          </div>
          <div class="rounded-2xl bg-white p-4">
            <p class="text-xs font-medium text-[#8a8494]">Dibuat</p>
            <p class="mt-2 text-sm text-[#374151]">
              {{ formatDateTime(material.createdAt) }}
            </p>
          </div>
        </div>

        <div class="mt-4 rounded-2xl bg-white p-4">
          <p class="text-sm font-medium text-[#171322]">Deskripsi</p>
          <p
            v-if="material.materialDesc"
            class="mt-3 whitespace-pre-line text-sm leading-6 text-[#6b6475]"
          >
            {{ material.materialDesc }}
          </p>
          <p v-else class="mt-3 text-sm leading-6 text-[#8a8494]">
            Deskripsi materi belum tersedia.
          </p>
        </div>
      </article>

      <article
        v-if="material.attachments?.length"
        class="rounded-[22px] border border-[#ebe7df] bg-white p-5"
      >
        <p class="text-sm font-medium text-[#171322]">Lampiran</p>
        <div class="mt-3 space-y-2">
          <a
            v-for="attachment in material.attachments"
            :key="attachment.mediaId"
            class="flex max-w-full items-center gap-3 overflow-hidden rounded-2xl bg-[#faf8f4] px-4 py-3 text-sm text-[#4a4356] transition hover:bg-[#f0e9dd]"
            :href="attachment.fileUrl"
            rel="noreferrer"
            target="_blank"
          >
            <PhFileText :size="18" class="shrink-0 text-[#7aa7d9]" />
            <span class="min-w-0 flex-1 truncate">{{ attachment.mediaName || 'Lampiran materi' }}</span>
          </a>
        </div>
      </article>

      <article v-else class="rounded-[22px] border border-[#ebe7df] bg-white p-5">
        <p class="text-sm font-medium text-[#171322]">Lampiran</p>
        <p class="mt-2 text-sm leading-6 text-[#8a8494]">Materi ini tidak memiliki lampiran.</p>
      </article>
    </section>
  </main>
</template>
