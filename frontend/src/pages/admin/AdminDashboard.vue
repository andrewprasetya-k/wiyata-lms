<script setup lang="ts">
import {
  PhBookOpen,
  PhCalendarBlank,
  PhChalkboardTeacher,
  PhClipboardText,
  PhUsers,
} from "@phosphor-icons/vue";
import LatestChatCard from "../../components/chat/LatestChatCard.vue";

const quickActions = [
  {
    title: "Struktur Akademik",
    description:
      "Siapkan tahun ajaran, semester, mata pelajaran, dan kategori tugas.",
    icon: PhCalendarBlank,
    tone: "bg-[#fff4ee] text-[#ea580c]",
    to: "/admin/academic-years",
  },
  {
    title: "Kelas",
    description: "Buat dan kelola kelas untuk semester aktif.",
    icon: PhBookOpen,
    tone: "bg-[#eef2ff] text-[#4f46e5]",
    to: "/admin/classes",
  },
  {
    title: "Penempatan Kelas",
    description: "Masukkan siswa dan guru ke kelas yang sesuai.",
    icon: PhClipboardText,
    tone: "bg-[#fff7ed] text-[#b45309]",
    to: "/admin/enrollments",
  },
  {
    title: "Warga Sekolah",
    description: "Tambahkan pengguna ke sekolah dan atur peran mereka.",
    icon: PhUsers,
    tone: "bg-[#ecfdf5] text-[#059669]",
    to: "/admin/users",
  },
  {
    title: "Ruang Mengajar",
    description: "Hubungkan guru, kelas, dan mata pelajaran untuk membuka akses.",
    icon: PhChalkboardTeacher,
    tone: "bg-[#f3f4f6] text-[#6b7280]",
    to: "/admin/subject-classes",
  },
];
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-[#f8f7f4]">
    <header class="border-b border-[#ebe7df] bg-white">
      <div
        class="flex min-w-0 flex-col gap-3 px-5 py-5 sm:px-6 lg:flex-row lg:items-end lg:justify-between lg:px-8"
      >
        <div class="min-w-0">
          <h1 class="mt-2 text-2xl font-semibold text-[#171322] sm:text-3xl">
            Dashboard admin sekolah
          </h1>
          <p class="mt-2 max-w-3xl text-sm leading-6 text-[#6b7280]">
            Pusat pengaturan sekolah. Pilih modul di bawah untuk mulai
            mengelola struktur akademik, kelas, warga sekolah, dan ruang mengajar.
          </p>
        </div>
      </div>
    </header>

    <section
      class="grid w-full max-w-none gap-6 px-5 py-6 sm:px-6 lg:px-8 xl:grid-cols-[minmax(0,1fr)_340px]"
    >
      <div class="flex min-w-0 flex-col gap-6">
        <section>
          <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5">
            <RouterLink
              v-for="item in quickActions"
              :key="item.to"
              :to="item.to"
              class="flex min-h-36 flex-col rounded-xl border border-[#ebe7df] bg-white p-4 shadow-sm transition hover:-translate-y-0.5 hover:border-[#ea580c] hover:shadow-md"
            >
              <div
                class="flex h-10 w-10 items-center justify-center rounded-xl"
                :class="item.tone"
              >
                <component :is="item.icon" :size="20" weight="duotone" />
              </div>
              <p class="mt-4 text-sm font-semibold text-[#171322]">
                {{ item.title }}
              </p>
              <p class="mt-2 text-xs leading-5 text-[#6b7280]">
                {{ item.description }}
              </p>
              <p class="mt-auto pt-4 text-xs font-semibold text-[#ea580c]">
                Buka halaman
              </p>
            </RouterLink>
          </div>
        </section>
      </div>

      <aside class="flex min-w-0 flex-col gap-5">
        <LatestChatCard to="/admin/chat" :limit="5" />
      </aside>
    </section>
  </main>
</template>
