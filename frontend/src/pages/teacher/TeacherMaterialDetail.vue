<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import {
  PhArrowLeft,
  PhBookOpen,
  PhWarningCircle,
  PhPencilSimple,
  PhTrash,
  PhCalendarBlank,
  PhUserCircle,
  PhPaperclip,
} from '@phosphor-icons/vue'
import AttachmentPreviewList from '../../components/common/AttachmentPreviewList.vue'
import DiscussionComments from '../../components/discussion/DiscussionComments.vue'
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
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-[#f8f7f4]">
    <header class="border-b border-[#ebe7df] bg-white">
      <div
        class="flex min-w-0 items-center gap-2 px-5 py-3 text-xs text-[#6b7280] sm:px-6 lg:px-8"
      >
        <RouterLink
          :to="`/teacher/subjects/${subjectClassId}`"
          class="inline-flex shrink-0 items-center gap-1.5 transition hover:text-[#4f46e5]"
        >
          <PhArrowLeft :size="15" />
          Mata pelajaran
        </RouterLink>
        <span class="text-[#d1d5db]">/</span>
        <span class="shrink-0">Materi</span>
        <span class="text-[#d1d5db]">/</span>
        <span class="min-w-0 truncate font-medium text-[#171322]">
          {{
            material?.materialTitle ??
            (isLoading ? 'Memuat...' : 'Detail materi')
          }}
        </span>
      </div>

      <div
        v-if="material"
        class="flex min-w-0 flex-col gap-4 border-t border-[#f3f1ec] px-5 py-4 sm:px-6 lg:flex-row lg:items-center lg:justify-between lg:px-8"
      >
        <div class="flex min-w-0 items-start gap-3">
          <div
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-[#eef2ff] text-[#4f46e5]"
          >
            <PhBookOpen :size="21" weight="duotone" />
          </div>
          <div class="min-w-0">
            <h1
              class="wrap-break-word text-xl font-semibold text-[#171322] sm:text-2xl"
            >
              {{ material.materialTitle }}
            </h1>
            <p class="mt-1 text-sm text-[#6b7280]">
              {{ material.subjectName || 'Materi mata pelajaran' }}
            </p>
          </div>
        </div>
        <span
          v-if="material.materialType"
          class="inline-flex self-start rounded-lg bg-[#eef2ff] px-2.5 py-1.5 text-xs font-medium uppercase tracking-wide text-[#4f46e5] lg:self-auto"
        >
          {{ material.materialType }}
        </span>
      </div>
    </header>

    <section class="px-5 py-5 sm:px-6 lg:px-8">
      <section
        v-if="isLoading"
        class="grid gap-5 lg:grid-cols-[minmax(0,1fr)_300px]"
      >
        <div
          class="h-80 animate-pulse rounded-xl border border-[#ebe7df] bg-white"
        />
        <div
          class="h-64 animate-pulse rounded-xl border border-[#ebe7df] bg-white"
        />
      </section>

      <section
        v-else-if="errorMessage"
        class="mx-auto max-w-xl rounded-xl border border-[#f0d8d2] bg-white px-5 py-8 text-center"
      >
        <PhWarningCircle
          :size="30"
          class="mx-auto text-[#d97757]"
          weight="duotone"
        />
        <h2 class="mt-3 text-lg font-semibold text-[#171322]">
          Materi belum bisa dimuat
        </h2>
        <p class="mt-2 text-sm leading-6 text-[#6b6475]">
          {{ errorMessage }}
        </p>
        <button
          class="mt-5 rounded-lg bg-[#171322] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#2f2b3a]"
          type="button"
          @click="loadMaterial"
        >
          Coba lagi
        </button>
      </section>

      <section
        v-else-if="didLoad && !material"
        class="mx-auto max-w-xl rounded-xl border border-[#ebe7df] bg-white px-5 py-10 text-center"
      >
        <div
          class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-[#eef2ff] text-[#4f46e5]"
        >
          <PhBookOpen :size="24" weight="duotone" />
        </div>
        <h2 class="mt-3 text-lg font-semibold text-[#171322]">
          Materi tidak ditemukan
        </h2>
        <p class="mt-2 text-sm leading-6 text-[#6b6475]">
          Materi ini tidak tersedia atau bukan bagian dari mata pelajaran yang
          Anda ajar.
        </p>
        <RouterLink
          :to="`/teacher/subjects/${subjectClassId}`"
          class="mt-5 inline-flex items-center gap-2 rounded-lg border border-[#ebe7df] bg-white px-4 py-2.5 text-sm font-medium text-[#4f46e5] transition hover:border-[#4f46e5] hover:bg-[#eef2ff]"
        >
          <PhArrowLeft :size="16" />
          Kembali ke mata pelajaran
        </RouterLink>
      </section>

      <section
        v-else-if="material"
        class="grid min-w-0 gap-5 lg:grid-cols-[minmax(0,1fr)_300px]"
      >
        <div class="min-w-0 space-y-5">
          <article class="rounded-xl border border-[#ebe7df] bg-white p-5 sm:p-6">
            <p
              class="text-[10px] font-medium uppercase tracking-[0.08em] text-[#9ca3af]"
            >
              Deskripsi materi
            </p>
            <p
              v-if="material.materialDesc"
              class="mt-3 whitespace-pre-line wrap-break-word text-sm leading-7 text-[#4a4356]"
            >
              {{ material.materialDesc }}
            </p>
            <div
              v-else
              class="mt-3 rounded-lg bg-[#fbfaf8] px-4 py-5 text-sm leading-6 text-[#8a8494]"
            >
              Deskripsi materi belum tersedia.
            </div>
          </article>

          <article class="rounded-xl border border-[#ebe7df] bg-white p-5 sm:p-6">
            <div class="flex items-center justify-between gap-3">
              <div>
                <h2 class="text-base font-semibold text-[#171322]">Lampiran</h2>
                <p class="mt-1 text-xs text-[#8a8494]">
                  {{ material.attachments?.length ?? 0 }} lampiran terhubung
                </p>
              </div>
              <PhPaperclip :size="20" class="text-[#4f46e5]" weight="duotone" />
            </div>
            <AttachmentPreviewList
              class="mt-4"
              :attachments="material.attachments"
              :material-id="material.materialId"
              empty-text="Materi ini tidak memiliki lampiran."
            />
          </article>

          <DiscussionComments
            source-type="material"
            :source-id="material.materialId"
            title="Diskusi materi"
            placeholder="Tulis tanggapan atau jawab pertanyaan tentang materi ini..."
            empty-text="Belum ada diskusi untuk materi ini."
          />
        </div>

        <aside class="min-w-0">
          <div class="space-y-4 lg:sticky lg:top-6">
            <article class="rounded-xl border border-[#ebe7df] bg-white p-5">
              <h2 class="text-sm font-semibold text-[#171322]">
                Informasi materi
              </h2>
              <dl class="mt-4 space-y-4">
                <div class="flex items-start gap-3">
                  <PhUserCircle
                    :size="18"
                    class="mt-0.5 shrink-0 text-[#8a8494]"
                    weight="duotone"
                  />
                  <div class="min-w-0">
                    <dt class="text-xs text-[#8a8494]">Dibuat oleh</dt>
                    <dd class="mt-1 wrap-break-word text-sm text-[#374151]">
                      {{ material.creatorName || 'Pengirim tidak tersedia' }}
                    </dd>
                  </div>
                </div>
                <div class="flex items-start gap-3">
                  <PhCalendarBlank
                    :size="18"
                    class="mt-0.5 shrink-0 text-[#8a8494]"
                    weight="duotone"
                  />
                  <div class="min-w-0">
                    <dt class="text-xs text-[#8a8494]">Dibuat</dt>
                    <dd class="mt-1 text-sm text-[#374151]">
                      {{ formatDateTime(material.createdAt) }}
                    </dd>
                  </div>
                </div>
                <div class="flex items-start gap-3">
                  <PhPaperclip
                    :size="18"
                    class="mt-0.5 shrink-0 text-[#8a8494]"
                    weight="duotone"
                  />
                  <div class="min-w-0">
                    <dt class="text-xs text-[#8a8494]">Lampiran</dt>
                    <dd class="mt-1 text-sm text-[#374151]">
                      {{ material.attachments?.length ?? 0 }} lampiran
                    </dd>
                  </div>
                </div>
              </dl>
            </article>

            <article class="rounded-xl border border-[#ebe7df] bg-white p-4">
              <p class="text-xs font-medium uppercase tracking-wide text-[#9ca3af]">
                Kelola materi
              </p>
              <div class="mt-3 grid gap-2">
                <RouterLink
                  :to="`/teacher/subjects/${subjectClassId}/materials/${materialId}/edit`"
                  class="inline-flex items-center justify-center gap-2 rounded-lg bg-[#4f46e5] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#4338ca]"
                >
                  <PhPencilSimple :size="16" />
                  Edit materi
                </RouterLink>
                <button
                  type="button"
                  class="inline-flex items-center justify-center gap-2 rounded-lg border border-[#f1d6d3] bg-white px-4 py-2.5 text-sm font-medium text-[#dc2626] transition hover:bg-[#fef2f2] disabled:opacity-50"
                  :disabled="isDeleting"
                  @click="handleDelete"
                >
                  <PhTrash :size="16" />
                  {{ isDeleting ? 'Menghapus...' : 'Hapus materi' }}
                </button>
              </div>
            </article>
          </div>
        </aside>
      </section>
    </section>
  </main>
</template>
