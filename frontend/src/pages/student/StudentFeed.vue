<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { PhMegaphone, PhWarningCircle } from '@phosphor-icons/vue'
import { useActiveClassStore } from '../../stores/activeClass'
import { useAuthStore } from '../../stores/auth'

const auth = useAuthStore()
const activeClassStore = useActiveClassStore()
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
    }
  } catch {
    errorMessage.value = 'Konteks kelas belum bisa dimuat. Periksa koneksi atau coba lagi nanti.'
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

    <section v-else-if="activeClass" class="soft-card max-w-3xl rounded-3xl p-6">
      <div class="mb-4 flex h-12 w-12 items-center justify-center rounded-2xl bg-[#eef2ff] text-[#4f46e5]">
        <PhMegaphone :size="26" weight="duotone" />
      </div>
      <p class="text-sm font-medium text-[#171322]">
        {{ activeClass?.classTitle || 'Belum ada kelas aktif' }}
      </p>
      <p class="mt-2 text-sm leading-6 text-[#7a7385]">
        Integrasi feed class belum diaktifkan di frontend tahap ini. Setelah active class context
        final, halaman ini dapat memakai endpoint feed class tanpa membuat data palsu.
      </p>
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
