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
import CommentThread from '../../components/comments/CommentThread.vue'
import { getMaterialById } from '../../services/classWorkspace'
import { deleteMaterial } from '../../services/teacherMaterial'
import type { MaterialItem } from '../../types/classWorkspace'
import { formatDateTime } from '../../utils/date'
import { useToastStore } from '../../stores/toast'
import { useConfirmStore } from '../../stores/confirm'

const route = useRoute()
const router = useRouter()
const toast = useToastStore()
const confirm = useConfirmStore()
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
  const ok = await confirm.confirm({
    title: 'Hapus materi?',
    description: 'Tindakan ini tidak dapat dibatalkan.',
    confirmLabel: 'Hapus',
    variant: 'danger',
  })
  if (!ok) return
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
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-background">
    <header class="border-b border-border bg-surface">
      <div
        class="flex min-w-0 items-center gap-2 px-5 py-3 text-xs text-muted sm:px-6 lg:px-8"
      >
        <RouterLink
          :to="`/teacher/subjects/${subjectClassId}`"
          class="inline-flex shrink-0 items-center gap-1.5 transition hover:text-brand"
        >
          <PhArrowLeft :size="15" />
          Mata pelajaran
        </RouterLink>
        <span class="text-border-strong">/</span>
        <span class="shrink-0">Materi</span>
        <span class="text-border-strong">/</span>
        <span class="min-w-0 truncate font-medium text-foreground">
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
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-brand-soft text-brand"
          >
            <PhBookOpen :size="21" weight="duotone" />
          </div>
          <div class="min-w-0">
            <h1
              class="wrap-break-word text-xl font-semibold text-foreground sm:text-2xl"
            >
              {{ material.materialTitle }}
            </h1>
            <p class="mt-1 text-sm text-muted">
              {{ material.subjectName || 'Materi mata pelajaran' }}
            </p>
          </div>
        </div>
        <span
          v-if="material.materialType"
          class="inline-flex self-start rounded-lg bg-brand-soft px-2.5 py-1.5 text-xs font-medium uppercase tracking-wide text-brand lg:self-auto"
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
          class="h-80 animate-pulse rounded-xl border border-border bg-surface"
        />
        <div
          class="h-64 animate-pulse rounded-xl border border-border bg-surface"
        />
      </section>

      <section
        v-else-if="errorMessage"
        class="mx-auto max-w-xl rounded-xl border border-danger-line bg-danger-soft px-5 py-8 text-center"
      >
        <PhWarningCircle
          :size="30"
          class="mx-auto text-[#d97757]"
          weight="duotone"
        />
        <h2 class="mt-3 text-lg font-semibold text-foreground">
          Materi belum bisa dimuat
        </h2>
        <p class="mt-2 text-sm leading-6 text-muted">
          {{ errorMessage }}
        </p>
        <button
          class="mt-5 rounded-lg bg-foreground px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#2f2b3a]"
          type="button"
          @click="loadMaterial"
        >
          Coba lagi
        </button>
      </section>

      <section
        v-else-if="didLoad && !material"
        class="mx-auto max-w-xl rounded-xl border border-border bg-surface px-5 py-10 text-center"
      >
        <div
          class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-brand-soft text-brand"
        >
          <PhBookOpen :size="24" weight="duotone" />
        </div>
        <h2 class="mt-3 text-lg font-semibold text-foreground">
          Materi tidak ditemukan
        </h2>
        <p class="mt-2 text-sm leading-6 text-muted">
          Materi ini tidak tersedia atau bukan bagian dari mata pelajaran yang
          Anda ajar.
        </p>
        <RouterLink
          :to="`/teacher/subjects/${subjectClassId}`"
          class="mt-5 inline-flex items-center gap-2 rounded-lg border border-border bg-surface px-4 py-2.5 text-sm font-medium text-brand transition hover:border-brand hover:bg-brand-soft"
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
          <article class="rounded-xl border border-border bg-surface shadow-sm p-5 sm:p-6">
            <p
              class="text-[10px] font-medium uppercase tracking-[0.08em] text-muted"
            >
              Deskripsi materi
            </p>
            <p
              v-if="material.materialDesc"
              class="mt-3 whitespace-pre-line wrap-break-word text-sm leading-7 text-foreground"
            >
              {{ material.materialDesc }}
            </p>
            <div
              v-else
              class="mt-3 rounded-lg bg-surface-subtle px-4 py-5 text-sm leading-6 text-muted"
            >
              Deskripsi materi belum tersedia.
            </div>
          </article>

          <article class="rounded-xl border border-border bg-surface shadow-sm p-5 sm:p-6">
            <div class="flex items-center justify-between gap-3">
              <div>
                <h2 class="text-base font-semibold text-foreground">Lampiran</h2>
                <p class="mt-1 text-xs text-muted">
                  {{ material.attachments?.length ?? 0 }} lampiran terhubung
                </p>
              </div>
              <PhPaperclip :size="20" class="text-brand" weight="duotone" />
            </div>
            <AttachmentPreviewList
              class="mt-4"
              :attachments="material.attachments"
              :material-id="material.materialId"
              empty-text="Materi ini tidak memiliki lampiran."
            />
          </article>

          <CommentThread
            source-type="material"
            :source-id="material.materialId"
            title="Diskusi materi"
            placeholder="Tulis tanggapan atau jawab pertanyaan tentang materi ini..."
            empty-text="Belum ada diskusi untuk materi ini."
          />
        </div>

        <aside class="min-w-0">
          <div class="space-y-4 lg:sticky lg:top-6">
            <article class="rounded-xl border border-border bg-surface shadow-sm p-5">
              <h2 class="text-sm font-semibold text-foreground">
                Informasi materi
              </h2>
              <dl class="mt-4 space-y-4">
                <div class="flex items-start gap-3">
                  <PhUserCircle
                    :size="18"
                    class="mt-0.5 shrink-0 text-muted"
                    weight="duotone"
                  />
                  <div class="min-w-0">
                    <dt class="text-xs text-muted">Dibuat oleh</dt>
                    <dd class="mt-1 wrap-break-word text-sm text-foreground-secondary">
                      {{ material.creatorName || 'Pengirim tidak tersedia' }}
                    </dd>
                  </div>
                </div>
                <div class="flex items-start gap-3">
                  <PhCalendarBlank
                    :size="18"
                    class="mt-0.5 shrink-0 text-muted"
                    weight="duotone"
                  />
                  <div class="min-w-0">
                    <dt class="text-xs text-muted">Dibuat</dt>
                    <dd class="mt-1 text-sm text-foreground-secondary">
                      {{ formatDateTime(material.createdAt) }}
                    </dd>
                  </div>
                </div>
                <div class="flex items-start gap-3">
                  <PhPaperclip
                    :size="18"
                    class="mt-0.5 shrink-0 text-muted"
                    weight="duotone"
                  />
                  <div class="min-w-0">
                    <dt class="text-xs text-muted">Lampiran</dt>
                    <dd class="mt-1 text-sm text-foreground-secondary">
                      {{ material.attachments?.length ?? 0 }} lampiran
                    </dd>
                  </div>
                </div>
              </dl>
            </article>

            <article class="rounded-xl border border-border bg-surface p-4">
              <p class="text-xs font-medium uppercase tracking-wide text-muted">
                Kelola materi
              </p>
              <div class="mt-3 grid gap-2">
                <RouterLink
                  :to="`/teacher/subjects/${subjectClassId}/materials/${materialId}/edit`"
                  class="inline-flex items-center justify-center gap-2 rounded-lg bg-brand px-4 py-2.5 text-sm font-medium text-white transition hover:bg-brand-hover"
                >
                  <PhPencilSimple :size="16" />
                  Edit materi
                </RouterLink>
                <button
                  type="button"
                  class="inline-flex items-center justify-center gap-2 rounded-lg border border-danger-line bg-surface px-4 py-2.5 text-sm font-medium text-danger transition hover:bg-danger-soft disabled:opacity-50"
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
