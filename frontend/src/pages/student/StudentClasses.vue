<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { PhArrowRight, PhBooks, PhWarningCircle } from '@phosphor-icons/vue'
import { useAuthStore } from '../../stores/auth'
import { getStudentClasses } from '../../services/studentClasses'
import type { StudentClassEnrollment } from '../../types/studentClasses'

const auth = useAuthStore()
const router = useRouter()

const palette = ['#4f8ef7', '#f2756a', '#c673d8', '#f0a05a', '#4f46e5']
const classes = ref<StudentClassEnrollment[]>([])
const isLoading = ref(true)
const errorMessage = ref('')

const activeMembership = computed(() => {
  const activeSchoolId = auth.activeSchoolId
  return (
    auth.memberships.find((membership) => membership.school.id === activeSchoolId) ??
    auth.memberships[0]
  )
})

const schoolName = computed(() => activeMembership.value?.school.name ?? 'Sekolah aktif')
const schoolUserId = computed(
  () => activeMembership.value?.schoolUserId ?? auth.defaultContext?.schoolUserId ?? '',
)

async function loadClasses() {
  if (!schoolUserId.value) {
    isLoading.value = false
    errorMessage.value =
      'Konteks sekolah belum tersedia. Silakan login ulang atau pilih sekolah aktif terlebih dahulu.'
    return
  }

  isLoading.value = true
  errorMessage.value = ''

  try {
    classes.value = await getStudentClasses(schoolUserId.value)
  } catch {
    errorMessage.value = 'Daftar kelas belum bisa dimuat. Periksa koneksi atau coba lagi nanti.'
  } finally {
    isLoading.value = false
  }
}

function openClass(item: StudentClassEnrollment) {
  if (!item.classId) return
  router.push(`/student/classes/${item.classId}`)
}

function classTitle(item: StudentClassEnrollment) {
  return item.classTitle || 'Kelas tanpa judul'
}

onMounted(loadClasses)
</script>

<template>
  <main class="min-h-screen flex-1 px-5 py-6 sm:px-8 lg:px-10">
    <header class="mb-6 flex flex-col gap-2">
      <p class="text-sm text-[#7a7385]">{{ schoolName }}</p>
      <h1 class="text-2xl font-medium tracking-normal text-[#171322]">Kelas saya</h1>
      <p class="max-w-2xl text-sm leading-6 text-[#7a7385]">
        Daftar kelas diambil dari enrollment akunmu pada sekolah aktif.
      </p>
    </header>

    <section v-if="isLoading" class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
      <div
        v-for="item in 6"
        :key="item"
        class="h-44 animate-pulse rounded-[18px] border border-[#ebe7df] bg-white"
      />
    </section>

    <section v-else-if="errorMessage" class="soft-card max-w-2xl rounded-3xl p-6">
      <div class="mb-4 flex h-11 w-11 items-center justify-center rounded-2xl bg-[#fff1f0] text-[#f2756a]">
        <PhWarningCircle :size="24" weight="duotone" />
      </div>
      <p class="text-sm font-medium text-[#171322]">Tidak bisa memuat kelas</p>
      <p class="mt-2 text-sm leading-6 text-[#7a7385]">{{ errorMessage }}</p>
      <button
        class="mt-5 rounded-2xl bg-[#4f46e5] px-4 py-2 text-sm font-medium text-white"
        type="button"
        @click="loadClasses"
      >
        Coba lagi
      </button>
    </section>

    <section v-else-if="classes.length > 0" class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
      <article
        v-for="(item, index) in classes"
        :key="item.enrollmentId"
        class="group overflow-hidden rounded-[18px] border border-[#ebe7df] bg-white transition hover:-translate-y-0.5 hover:shadow-[0_18px_40px_rgba(66,55,40,0.08)]"
      >
        <button class="block w-full text-left" type="button" @click="openClass(item)">
          <div
            class="flex h-24 flex-col justify-end px-4 pb-4 text-white"
            :style="{ backgroundColor: palette[index % palette.length] }"
          >
            <h2 class="text-base font-medium">{{ classTitle(item) }}</h2>
          </div>
          <div class="space-y-3 px-4 py-4">
            <div class="flex items-center justify-between gap-3">
              <div>
                <p class="text-xs text-[#9a95a3]">Role kelas</p>
                <p class="mt-1 text-sm font-medium capitalize text-[#3f3a4a]">{{ item.role }}</p>
              </div>
              <PhArrowRight
                :size="18"
                class="text-[#a09aa8] transition group-hover:translate-x-0.5 group-hover:text-[#4f46e5]"
              />
            </div>
            <p class="text-xs leading-5 text-[#7a7385]">
              <span v-if="item.joinedAt">Bergabung {{ item.joinedAt }}</span>
              <span v-else>Detail segera tersedia</span>
            </p>
          </div>
        </button>
      </article>
    </section>

    <section v-else class="soft-card max-w-2xl rounded-3xl p-6">
      <div class="mb-4 flex h-11 w-11 items-center justify-center rounded-2xl bg-[#eef2ff] text-[#4f46e5]">
        <PhBooks :size="24" weight="duotone" />
      </div>
      <p class="text-sm font-medium text-[#171322]">Belum ada kelas aktif</p>
      <p class="mt-2 text-sm leading-6 text-[#7a7385]">
        Kelas akan tampil setelah akunmu terdaftar sebagai student pada kelas di sekolah aktif.
      </p>
    </section>
  </main>
</template>
