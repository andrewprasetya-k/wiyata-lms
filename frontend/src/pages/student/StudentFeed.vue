<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { PhChatCircleText, PhFileText, PhMegaphone, PhWarningCircle } from '@phosphor-icons/vue'
import { getClassFeed } from '../../services/feed'
import { useActiveClassStore } from '../../stores/activeClass'
import { useAuthStore } from '../../stores/auth'
import type { FeedClassHeader, FeedPost } from '../../types/feed'

const auth = useAuthStore()
const activeClassStore = useActiveClassStore()
const classHeader = ref<FeedClassHeader | null>(null)
const posts = ref<FeedPost[]>([])
const isLoading = ref(true)
const errorMessage = ref('')

const activeMembership = computed(() => {
  const activeSchoolId = auth.activeSchoolId
  return (
    auth.memberships.find((membership) => membership.school.id === activeSchoolId) ??
    auth.memberships[0]
  )
})

const schoolUserId = computed(
  () => activeMembership.value?.schoolUserId ?? auth.defaultContext?.schoolUserId ?? '',
)
const activeClass = computed(() => activeClassStore.activeClass)

async function loadContext() {
  if (!schoolUserId.value) {
    isLoading.value = false
    errorMessage.value =
      'Konteks sekolah belum tersedia. Silakan login ulang atau pilih sekolah aktif terlebih dahulu.'
    return
  }

  isLoading.value = true
  errorMessage.value = ''

  try {
    await activeClassStore.loadClasses(schoolUserId.value)
    if (activeClassStore.errorMessage) {
      errorMessage.value = activeClassStore.errorMessage
      posts.value = []
      return
    }

    if (!activeClassStore.activeClassId) {
      posts.value = []
      return
    }

    const feed = await getClassFeed(activeClassStore.activeClassId)
    classHeader.value = feed.class
    posts.value = feed.data.data || []
  } catch {
    errorMessage.value = 'Feed kelas belum bisa dimuat. Periksa koneksi atau coba lagi nanti.'
  } finally {
    isLoading.value = false
  }
}

onMounted(loadContext)
</script>

<template>
  <main class="min-h-screen flex-1 px-5 py-6 sm:px-8 lg:px-10">
    <header class="mb-6">
      <p class="text-sm text-[#7a7385]">Class feed</p>
      <h1 class="mt-2 text-3xl font-medium tracking-normal text-[#171322]">Feed kelas</h1>
      <p class="mt-3 max-w-2xl text-sm leading-6 text-[#7a7385]">
        Feed adalah komunikasi level class untuk pengumuman dan aktivitas kelas. Chat tetap fitur
        realtime terpisah yang sedang dikembangkan.
      </p>
    </header>

    <section v-if="isLoading || activeClassStore.isLoading" class="soft-card max-w-3xl rounded-3xl p-6">
      <div class="h-24 animate-pulse rounded-2xl bg-white" />
    </section>

    <section v-else-if="errorMessage" class="soft-card max-w-3xl rounded-3xl p-6">
      <div class="mb-4 flex h-11 w-11 items-center justify-center rounded-2xl bg-[#fff1f0] text-[#f2756a]">
        <PhWarningCircle :size="24" weight="duotone" />
      </div>
      <p class="text-sm font-medium text-[#171322]">Tidak bisa memuat konteks feed</p>
      <p class="mt-2 text-sm leading-6 text-[#7a7385]">{{ errorMessage }}</p>
      <button
        class="mt-5 rounded-2xl bg-[#4f46e5] px-4 py-2 text-sm font-medium text-white"
        type="button"
        @click="loadContext"
      >
        Coba lagi
      </button>
    </section>

    <section v-else-if="activeClass" class="max-w-3xl space-y-3">
      <div class="soft-card rounded-3xl p-6">
        <div class="mb-4 flex h-12 w-12 items-center justify-center rounded-2xl bg-[#eef2ff] text-[#4f46e5]">
          <PhMegaphone :size="26" weight="duotone" />
        </div>
        <p class="text-sm font-medium text-[#171322]">
          {{ classHeader?.classTitle || activeClass.classTitle || 'Kelas aktif' }}
        </p>
        <p class="mt-2 text-sm leading-6 text-[#7a7385]">
          Feed kelas menampilkan komunikasi dan pengumuman level class. Komentar dan pembuatan post
          belum diimplementasikan di frontend tahap ini.
        </p>
      </div>

      <div v-if="posts.length === 0" class="soft-card rounded-3xl p-6">
        <p class="text-sm font-medium text-[#171322]">Belum ada feed</p>
        <p class="mt-2 text-sm leading-6 text-[#7a7385]">
          Post akan tampil setelah guru atau admin membuat pengumuman untuk kelas ini.
        </p>
      </div>

      <article
        v-for="post in posts"
        v-else
        :key="post.feedId"
        class="rounded-3xl border border-[#ebe7df] bg-white p-5"
      >
        <div class="flex items-start gap-3">
          <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-[#eef2ff] text-[#4f46e5]">
            <PhMegaphone :size="20" weight="duotone" />
          </div>
          <div class="min-w-0 flex-1">
            <div class="flex flex-wrap items-baseline justify-between gap-2">
              <p class="text-sm font-medium text-[#171322]">
                {{ post.creatorName || 'Pengirim tidak tersedia' }}
              </p>
              <span class="text-xs text-[#a09aa8]">{{ post.createdAt }}</span>
            </div>
            <p class="mt-3 whitespace-pre-line text-sm leading-6 text-[#4a4356]">
              {{ post.content }}
            </p>
            <div class="mt-4 flex flex-wrap gap-2 text-xs text-[#7a7385]">
              <span
                v-if="post.commentCount !== undefined"
                class="inline-flex items-center gap-1 rounded-full bg-[#fbfaf8] px-3 py-1"
              >
                <PhChatCircleText :size="14" />
                {{ post.commentCount }} komentar
              </span>
              <span
                v-if="post.attachments?.length"
                class="inline-flex items-center gap-1 rounded-full bg-[#fbfaf8] px-3 py-1"
              >
                <PhFileText :size="14" />
                {{ post.attachments.length }} lampiran
              </span>
            </div>
          </div>
        </div>
      </article>
    </section>

    <section v-else class="soft-card max-w-3xl rounded-3xl p-6">
      <div class="mb-4 flex h-12 w-12 items-center justify-center rounded-2xl bg-[#eef2ff] text-[#4f46e5]">
        <PhMegaphone :size="26" weight="duotone" />
      </div>
      <p class="text-sm font-medium text-[#171322]">Belum ada kelas aktif</p>
      <p class="mt-2 text-sm leading-6 text-[#7a7385]">
        Feed kelas akan tersedia setelah akunmu memiliki enrollment dan active class context.
      </p>
    </section>
  </main>
</template>
