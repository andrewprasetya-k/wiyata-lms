<script setup lang="ts">
import {
  PhArrowRight,
  PhBookOpen,
  PhCalendarBlank,
  PhChalkboardTeacher,
  PhClipboardText,
  PhHouse,
  PhLinkSimple,
  PhStudent,
  PhUsers,
} from "@phosphor-icons/vue";

const quickActions = [
  {
    title: "Atur Struktur Akademik",
    description:
      "Siapkan tahun ajaran, semester, mata pelajaran, dan kategori tugas.",
    icon: PhCalendarBlank,
    tone: "bg-[#fff4ee] text-[#ea580c]",
    to: "/admin/academic-years",
  },
  {
    title: "Buat Kelas",
    description: "Tambahkan kelas untuk periode akademik yang sudah aktif.",
    icon: PhBookOpen,
    tone: "bg-[#eef2ff] text-[#4f46e5]",
    to: "/admin/classes",
  },
  {
    title: "Kelola Warga Sekolah",
    description: "Tambahkan akun pengguna ke sekolah dan atur perannya.",
    icon: PhUsers,
    tone: "bg-[#ecfdf5] text-[#059669]",
    to: "/admin/users",
  },
  {
    title: "Tempatkan ke Kelas",
    description: "Masukkan siswa atau guru ke kelas yang sesuai.",
    icon: PhClipboardText,
    tone: "bg-[#fff7ed] text-[#ea580c]",
    to: "/admin/enrollments",
  },
];

const setupFlow = [
  {
    title: "Struktur Akademik",
    description: "Buat periode akademik dan data mata pelajaran lebih dulu.",
    to: "/admin/academic-years",
  },
  {
    title: "Kelas",
    description: "Buat kelas di semester yang akan digunakan.",
    to: "/admin/classes",
  },
  {
    title: "Warga Sekolah",
    description: "Pastikan siswa, guru, dan admin sudah menjadi warga sekolah.",
    to: "/admin/users",
  },
  {
    title: "Penempatan Kelas",
    description: "Tempatkan siswa dan guru ke kelas aktif.",
    to: "/admin/enrollments",
  },
  {
    title: "Penugasan Mengajar",
    description: "Hubungkan guru, kelas, dan mata pelajaran.",
    to: "/admin/subject-classes",
  },
];

const managementLinks = [
  {
    label: "Penugasan Mengajar",
    description: "Buat ruang mengajar guru untuk mata pelajaran di kelas.",
    icon: PhChalkboardTeacher,
    tone: "bg-[#f3f4f6] text-[#6b7280]",
    to: "/admin/subject-classes",
  },
  {
    label: "Warga Sekolah",
    description: "Periksa akses sekolah dan peran sebelum penempatan kelas.",
    icon: PhUsers,
    tone: "bg-[#ecfdf5] text-[#059669]",
    to: "/admin/users",
  },
  {
    label: "Penempatan Kelas",
    description:
      "Keluarkan atau tempatkan ulang warga sekolah tanpa menghapus histori.",
    icon: PhStudent,
    tone: "bg-[#fff7ed] text-[#ea580c]",
    to: "/admin/enrollments",
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
            Pusat pengaturan sekolah untuk membuka alur akademik, kelas, warga
            sekolah, penempatan kelas, dan penugasan mengajar.
          </p>
        </div>
      </div>
    </header>

    <section
      class="grid w-full max-w-none gap-6 px-5 py-6 sm:px-6 lg:px-8 xl:grid-cols-[minmax(0,1fr)_340px]"
    >
      <div class="flex min-w-0 flex-col gap-6">
        <section
          class="rounded-xl border border-[#ebe7df] bg-white p-5 shadow-sm"
        >
          <div
            class="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between"
          >
            <div class="min-w-0">
              <p
                class="text-xs font-semibold uppercase tracking-[0.16em] text-[#ea580c]"
              >
                Pusat pengaturan
              </p>
              <h2 class="mt-2 text-xl font-semibold text-[#171322]">
                Alur awal sekolah
              </h2>
              <p class="mt-2 max-w-3xl text-sm leading-6 text-[#6b7280]">
                Kartu berikut adalah pintasan operasional. Dashboard ini tidak
                menampilkan statistik langsung agar tidak memberi angka yang
                menyesatkan.
              </p>
            </div>
            <span
              class="inline-flex rounded-lg bg-[#fff4ee] px-3 py-2 text-xs font-semibold text-[#ea580c]"
            >
              MVP sekolah
            </span>
          </div>
        </section>

        <section>
          <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
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

        <section class="grid gap-5 lg:grid-cols-2">
          <article
            class="rounded-xl border border-[#ebe7df] bg-white p-5 shadow-sm"
          >
            <div class="flex items-center justify-between gap-4">
              <div>
                <p class="text-sm font-semibold text-[#171322]">
                  Alur pengaturan
                </p>
                <p class="mt-1 text-xs text-[#6b7280]">
                  Urutan kerja yang disarankan untuk admin sekolah.
                </p>
              </div>
              <PhLinkSimple
                :size="20"
                class="text-[#ea580c]"
                weight="duotone"
              />
            </div>

            <div class="mt-4 divide-y divide-[#f3f4f6]">
              <RouterLink
                v-for="(item, index) in setupFlow"
                :key="item.to"
                :to="item.to"
                class="flex items-start gap-3 py-3"
              >
                <span
                  class="mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-[#fff4ee] text-xs font-semibold text-[#ea580c]"
                >
                  {{ index + 1 }}
                </span>
                <span class="min-w-0">
                  <span class="block text-sm font-semibold text-[#171322]">
                    {{ item.title }}
                  </span>
                  <span class="mt-1 block text-xs leading-5 text-[#6b7280]">
                    {{ item.description }}
                  </span>
                </span>
              </RouterLink>
            </div>
          </article>

          <article
            class="rounded-xl border border-[#ebe7df] bg-white p-5 shadow-sm"
          >
            <div class="flex items-center justify-between gap-4">
              <div>
                <p class="text-sm font-semibold text-[#171322]">
                  Manajemen aktif
                </p>
                <p class="mt-1 text-xs text-[#6b7280]">
                  Area yang biasanya perlu dicek setelah pengaturan awal.
                </p>
              </div>
              <PhHouse :size="20" class="text-[#4f46e5]" weight="duotone" />
            </div>

            <div class="mt-4 space-y-2">
              <RouterLink
                v-for="item in managementLinks"
                :key="item.to"
                :to="item.to"
                class="flex items-center gap-3 rounded-lg p-3 transition hover:bg-[#f9fafb]"
              >
                <span
                  class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl"
                  :class="item.tone"
                >
                  <component :is="item.icon" :size="20" weight="duotone" />
                </span>
                <span class="min-w-0 flex-1">
                  <span class="block text-sm font-semibold text-[#171322]">
                    {{ item.label }}
                  </span>
                  <span class="mt-1 block text-xs leading-5 text-[#6b7280]">
                    {{ item.description }}
                  </span>
                </span>
                <PhArrowRight :size="16" class="shrink-0 text-[#9ca3af]" />
              </RouterLink>
            </div>
          </article>
        </section>
      </div>

      <aside class="flex min-w-0 flex-col gap-5">
        <section
          class="rounded-xl border border-[#ebe7df] bg-white p-5 shadow-sm"
        >
          <p class="text-sm font-semibold text-[#171322]">Manajemen sekolah</p>
          <p class="mt-1 text-xs leading-5 text-[#6b7280]">
            Modul utama mengikuti alur pengaturan MVP dan mengarah ke halaman
            admin yang sudah tersedia.
          </p>

          <div class="mt-4 space-y-2">
            <RouterLink
              v-for="item in setupFlow"
              :key="`module-${item.to}`"
              :to="item.to"
              class="block rounded-lg px-3 py-2.5 transition hover:bg-[#f9fafb]"
            >
              <p class="text-sm font-semibold text-[#171322]">
                {{ item.title }}
              </p>
              <p class="mt-1 text-xs leading-5 text-[#6b7280]">
                {{ item.description }}
              </p>
            </RouterLink>
          </div>
        </section>

        <!-- <section
          class="rounded-xl border border-[#ebe7df] bg-white p-5 shadow-sm"
        >
          <p class="text-sm font-semibold text-[#171322]">Catatan MVP</p>
          <div class="mt-4 space-y-3">
            <div class="flex gap-3">
              <span class="mt-1.5 h-2 w-2 shrink-0 rounded-full bg-[#059669]" />
              <p class="text-xs leading-5 text-[#6b7280]">
                Data langsung dan statistik operasional belum ditarik di
                dashboard ini agar tidak menampilkan angka yang menyesatkan.
              </p>
            </div>
            <div class="flex gap-3">
              <span class="mt-1.5 h-2 w-2 shrink-0 rounded-full bg-[#4f46e5]" />
              <p class="text-xs leading-5 text-[#6b7280]">
                Gunakan halaman Penempatan Kelas dan Penugasan Mengajar untuk
                membuka akses guru dan siswa ke ruang belajar.
              </p>
            </div>
            <div class="flex gap-3">
              <span class="mt-1.5 h-2 w-2 shrink-0 rounded-full bg-[#ea580c]" />
              <p class="text-xs leading-5 text-[#6b7280]">
                Histori akademik tidak dikelola dari dashboard; tindakan utama
                tetap berada di halaman pengaturan masing-masing.
              </p>
            </div>
          </div>
        </section> -->
      </aside>
    </section>
  </main>
</template>
