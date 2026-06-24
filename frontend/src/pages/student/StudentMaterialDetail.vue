<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import {
  PhArrowLeft,
  PhBookOpen,
  PhWarningCircle,
} from '@phosphor-icons/vue'
import AttachmentPreviewList from '../../components/common/AttachmentPreviewList.vue'
import StudentNoteCard from '../../components/student/StudentNoteCard.vue'
import { getMaterialById } from '../../services/classWorkspace'
import type { MaterialItem } from '../../types/classWorkspace'
import { formatDateTime } from '../../utils/date'

const route = useRoute()
const subjectClassId = computed(() => String(route.params.sclId ?? ''))
const materialId = computed(() => String(route.params.matId ?? ''))
const material = ref<MaterialItem | null>(null)
const isLoading = ref(true)
const errorMessage = ref('')
const didLoad = ref(false)

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

onMounted(loadMaterial)
</script>

<template>
  <main class="min-h-screen flex-1 px-5 py-5 sm:px-6 lg:px-8">
    <RouterLink
      class="mb-5 inline-flex items-center gap-2 rounded-md bg-white px-4 py-2 text-sm font-medium text-[#4f46e5] transition hover:bg-[#eef2ff]"
      :to="`/student/subjects/${subjectClassId}`"
    >
      <PhArrowLeft :size="18" />
      Kembali ke subject
    </RouterLink>

    <section v-if="isLoading" class="max-w-4xl space-y-3">
      <div class="h-40 animate-pulse rounded-3xl border border-[#ebe7df] bg-white" />
      <div class="h-28 animate-pulse rounded-3xl border border-[#ebe7df] bg-white" />
      <div class="h-24 animate-pulse rounded-3xl border border-[#ebe7df] bg-white" />
    </section>

    <section v-else-if="errorMessage" class="soft-card max-w-4xl rounded-[22px] p-5">
      <div class="mb-4 flex h-11 w-11 items-center justify-center rounded-2xl bg-[#fff1f0] text-[#f2756a]">
        <PhWarningCircle :size="24" weight="duotone" />
      </div>
      <p class="text-sm font-medium text-[#171322]">Tidak bisa memuat materi</p>
      <p class="mt-2 text-sm leading-6 text-[#7a7385]">{{ errorMessage }}</p>
      <button
        class="mt-5 rounded-2xl bg-[#4f46e5] px-4 py-2 text-sm font-medium text-white"
        type="button"
        @click="loadMaterial"
      >
        Coba lagi
      </button>
    </section>

    <section v-else-if="didLoad && !material" class="soft-card max-w-4xl rounded-[22px] p-5">
      <div class="mb-4 flex h-11 w-11 items-center justify-center rounded-2xl bg-[#eef2ff] text-[#4f46e5]">
        <PhBookOpen :size="24" weight="duotone" />
      </div>
      <p class="text-sm font-medium text-[#171322]">Materi tidak ditemukan</p>
      <p class="mt-2 text-sm leading-6 text-[#7a7385]">
        Material ID ini tidak ditemukan atau belum tersedia untuk subject yang sedang dibuka.
      </p>
    </section>

    <section v-else-if="material" class="max-w-4xl space-y-4">
      <article class="soft-card rounded-[22px] p-5">
        <div class="mb-5 flex items-start gap-4">
          <div
            class="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-[#eef2ff] text-[#4f46e5]"
          >
            <PhBookOpen :size="24" weight="duotone" />
          </div>
          <div class="min-w-0">
            <p class="text-sm text-[#7a7385]">
              {{ material.subjectName || 'Subject material' }}
            </p>
            <h1 class="mt-2 text-3xl font-medium tracking-normal text-[#171322]">
              {{ material.materialTitle }}
            </h1>
            <p v-if="material.materialType" class="mt-2 text-sm uppercase text-[#4f46e5]">
              {{ material.materialType }}
            </p>
          </div>
        </div>

        <div class="grid gap-3 sm:grid-cols-2">
          <div class="rounded-2xl bg-[#fbfaf8] p-4">
            <p class="text-xs font-medium text-[#7a7385]">Dibuat oleh</p>
            <p class="mt-2 text-sm text-[#3f3a4a]">
              {{ material.creatorName || 'Creator tidak tersedia' }}
            </p>
          </div>
          <div class="rounded-2xl bg-[#fbfaf8] p-4">
            <p class="text-xs font-medium text-[#7a7385]">Dibuat</p>
            <p class="mt-2 text-sm text-[#3f3a4a]">
              {{ formatDateTime(material.createdAt) }}
            </p>
          </div>
        </div>

        <div class="mt-5 rounded-2xl bg-white p-4">
          <p class="text-sm font-medium text-[#171322]">Deskripsi</p>
          <p
            v-if="material.materialDesc"
            class="mt-3 whitespace-pre-line text-sm leading-6 text-[#6b6475]"
          >
            {{ material.materialDesc }}
          </p>
          <p v-else class="mt-3 text-sm leading-6 text-[#7a7385]">
            Deskripsi materi belum tersedia.
          </p>
        </div>
      </article>

      <article class="rounded-[22px] border border-[#ebe7df] bg-white p-5">
        <p class="text-sm font-medium text-[#171322]">Lampiran</p>
        <AttachmentPreviewList
          class="mt-3"
          :attachments="material.attachments"
          empty-text="Materi ini tidak memiliki lampiran."
        />
      </article>

      <StudentNoteCard :material-id="material.materialId" />

      <article class="rounded-[22px] border border-[#ebe7df] bg-white p-5">
        <p class="text-sm font-medium text-[#171322]">Progress materi</p>
        <p class="mt-2 text-sm leading-6 text-[#7a7385]">
          Progress materi direncanakan setelah MVP sekolah. Membuka materi belum menandai progres selesai.
        </p>
      </article>
    </section>
  </main>
</template>
