<template>
  <main class="max-w-screen bg-[#fbfaf8] text-[#171322]">
    <!-- Navbar -->
    <header class="sticky top-0 z-50 bg-white/50 backdrop-blur-xl">
      <div
        class="mx-auto flex max-w-7xl items-center justify-between px-5 py-4 sm:px-6 lg:px-8"
      >
        <RouterLink to="/home" class="flex min-w-0 items-center gap-3">
          <img
            src="/logo_fix.svg"
            alt="Wiyata"
            class="h-11 w-11 rounded-2xl object-contain"
          />
          <div class="min-w-0">
            <p class="text-sm font-medium text-[#4f46e5]">Wiyata</p>
            <p class="truncate text-xs text-[#6b7280]">
              Workspace akademik sekolah
            </p>
          </div>
        </RouterLink>

        <nav class="hidden items-center gap-7 text-sm text-[#6b6475] md:flex">
          <a href="#fitur" class="transition hover:text-[#171322]">Fitur</a>
          <a href="#siswa" class="transition hover:text-[#171322]">Siswa</a>
          <a href="#guru" class="transition hover:text-[#171322]">Guru</a>
          <a href="#admin" class="transition hover:text-[#171322]">Admin</a>
        </nav>

        <RouterLink
          to="/login"
          class="rounded-2xl bg-[#4f46e5] px-5 py-2.5 text-sm font-medium text-white transition hover:bg-[#4338ca]"
        >
          Masuk
        </RouterLink>
      </div>
    </header>

    <!-- Hero -->
    <section
      class="mx-auto grid max-w-7xl gap-12 px-5 py-16 sm:px-6 lg:grid-cols-[0.95fr_1.05fr] lg:px-8 lg:py-24"
    >
      <div class="flex flex-col justify-center">
        <p class="text-sm font-medium text-[#4f46e5]">
          Wiyata Academic Workspace
        </p>
        <h1
          class="mt-4 max-w-3xl text-4xl font-medium leading-tight tracking-tight text-[#171322] sm:text-5xl lg:text-6xl"
        >
          Workspace akademik sekolah yang menyatukan belajar, mengajar, dan
          operasional.
        </h1>

        <p class="mt-6 max-w-2xl text-base leading-8 text-[#6b6475] sm:text-lg">
          Wiyata dirancang untuk sekolah yang membutuhkan alur kerja akademik
          yang rapi: dari setup tahun ajaran, pengelolaan warga sekolah, materi
          dan tugas, sampai chat sekolah dan grup diskusi.
        </p>

        <div class="mt-8 flex flex-col gap-3 sm:flex-row">
          <RouterLink
            to="/login"
            class="inline-flex h-12 items-center justify-center rounded-2xl bg-[#4f46e5] px-6 text-sm font-medium text-white transition hover:bg-[#4338ca]"
          >
            Masuk ke Wiyata
          </RouterLink>

          <a
            href="#fitur"
            class="inline-flex h-12 items-center justify-center rounded-2xl border border-[#e7e2da] bg-white px-6 text-sm font-medium text-[#171322] transition hover:bg-[#fbfaf8]"
          >
            Lihat fitur
          </a>
        </div>
      </div>

      <!-- Product Preview -->
      <div
        class="rounded-4xl border border-[#e7e2da] bg-white p-3 shadow-xl shadow-[#d8d2c6]/25"
      >
        <div class="rounded-[1.75rem] bg-[#eef2ff] p-4">
          <div class="rounded-[1.25rem] border border-white/80 bg-white p-5">
            <div class="mb-5 flex items-center justify-between gap-4">
              <div class="min-w-0">
                <p class="text-xs font-medium text-[#7a7385]">
                  Dashboard siswa
                </p>
                <h2 class="mt-1 truncate text-xl font-medium text-[#171322]">
                  Belajar lebih terarah
                </h2>
              </div>
              <div
                class="flex h-11 w-11 items-center justify-center rounded-2xl bg-[#4f46e5] text-sm font-medium text-white"
              >
                Ev
              </div>
            </div>

            <div class="grid gap-3 sm:grid-cols-3">
              <div
                v-for="item in previewStats"
                :key="item.label"
                class="rounded-2xl border border-[#e7e2da] bg-[#fbfaf8] p-4"
              >
                <p class="text-2xl font-medium text-[#171322]">
                  {{ item.value }}
                </p>
                <p class="mt-1 text-xs text-[#7a7385]">{{ item.label }}</p>
              </div>
            </div>

            <div class="mt-4 grid gap-4 lg:grid-cols-[1fr_0.85fr]">
              <div class="rounded-2xl border border-[#e7e2da] bg-white p-4">
                <div class="mb-3 flex items-center justify-between">
                  <h3 class="text-sm font-medium text-[#171322]">
                    Aktivitas hari ini
                  </h3>
                  <span
                    class="rounded-full bg-[#eef2ff] px-3 py-1 text-xs font-medium text-[#4f46e5]"
                  >
                    Aktif
                  </span>
                </div>

                <div class="space-y-3">
                  <div
                    v-for="activity in activities"
                    :key="activity.title"
                    class="rounded-2xl bg-[#fbfaf8] p-3"
                  >
                    <p class="text-sm font-medium text-[#171322]">
                      {{ activity.title }}
                    </p>
                    <p class="mt-1 text-xs text-[#7a7385]">
                      {{ activity.meta }}
                    </p>
                  </div>
                </div>
              </div>

              <div class="rounded-2xl border border-[#e7e2da] bg-white p-4">
                <h3 class="text-sm font-medium text-[#171322]">
                  Catatan pribadi
                </h3>
                <p
                  class="mt-3 rounded-2xl bg-[#eef2ff] p-4 text-sm leading-7 text-[#5f5968]"
                >
                  Teks eksposisi bertujuan menjelaskan informasi secara objektif
                  dengan struktur tesis, argumen, dan penegasan ulang.
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Workspace Story -->
    <section id="fitur" class="mx-auto max-w-7xl px-5 pb-20 sm:px-6 lg:px-8">
      <div
        class="grid gap-8 rounded-4xl border border-[#e7e2da] bg-white p-6 lg:grid-cols-[0.9fr_1.1fr] lg:p-8"
      >
        <div class="max-w-xl">
          <p class="text-sm font-medium text-[#4f46e5]">Cara kerja Wiyata</p>
          <h2
            class="mt-3 text-3xl font-medium tracking-tight text-[#171322] sm:text-4xl"
          >
            Aktivitas sekolah bergerak dalam satu alur yang sama.
          </h2>
          <p class="mt-4 text-base leading-8 text-[#6b6475]">
            Wiyata dimulai dari struktur sekolah yang rapi, lalu mengalir ke
            kelas, materi, tugas, pengumpulan, nilai, dan percakapan harian.
            Hasilnya bukan halaman yang berdiri sendiri, tetapi satu workspace
            yang terasa nyambung.
          </p>
          <p class="mt-4 text-base leading-8 text-[#6b6475]">
            Siswa melihat apa yang perlu dipelajari. Guru mengajar dan menilai
            dari tempat yang sama. Admin sekolah menjaga semuanya tetap teratur.
            Super Admin menangani struktur platform tanpa ikut masuk ke
            operasional akademik sekolah.
          </p>
        </div>

        <div class="grid gap-3 sm:grid-cols-2">
          <article
            v-for="story in workspaceStory"
            :key="story.title"
            class="rounded-3xl border border-[#e7e2da] bg-[#fbfaf8] p-5"
          >
            <p class="text-sm font-medium text-[#171322]">{{ story.title }}</p>
            <p class="mt-2 text-sm leading-7 text-[#6b6475]">
              {{ story.description }}
            </p>
          </article>
        </div>
      </div>
    </section>

    <!-- Roles -->
    <section class="mx-auto max-w-7xl px-5 pb-20 sm:px-6 lg:px-8">
      <div class="max-w-2xl">
        <p class="text-sm font-medium text-[#4f46e5]">Untuk setiap peran</p>
        <h2
          class="mt-3 text-3xl font-medium tracking-tight text-[#171322] sm:text-4xl"
        >
          Satu produk, dengan ruang kerja yang terasa tepat untuk masing-masing
          pengguna.
        </h2>
      </div>

      <div class="mt-8 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <article
          v-for="role in roleSummaries"
          :key="role.title"
          :id="role.anchor"
          class="rounded-4xl border border-[#e7e2da] bg-white p-6"
        >
          <p class="text-sm font-medium text-[#4f46e5]">{{ role.eyebrow }}</p>
          <h3 class="mt-3 text-xl font-medium text-[#171322]">
            {{ role.title }}
          </h3>
          <p class="mt-3 text-sm leading-7 text-[#6b6475]">
            {{ role.description }}
          </p>
        </article>
      </div>
    </section>

    <!-- Core Workflows -->
    <section class="border-y border-[#e7e2da] bg-white">
      <div class="mx-auto max-w-7xl px-5 py-20 sm:px-6 lg:px-8">
        <div class="max-w-2xl">
          <p class="text-sm font-medium text-[#4f46e5]">Workflow inti</p>
          <h2
            class="mt-3 text-3xl font-medium tracking-tight text-[#171322] sm:text-4xl"
          >
            Empat hal yang paling sering dikerjakan sekolah, disusun agar lebih
            mudah diikuti.
          </h2>
        </div>

        <div class="mt-10 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <article
            v-for="workflow in workflows"
            :key="workflow.title"
            class="rounded-4xl border border-[#e7e2da] bg-[#fbfaf8] p-6"
          >
            <p class="text-sm font-medium text-[#4f46e5]">
              {{ workflow.eyebrow }}
            </p>
            <h3 class="mt-3 text-lg font-medium text-[#171322]">
              {{ workflow.title }}
            </h3>
            <p class="mt-3 text-sm leading-7 text-[#6b6475]">
              {{ workflow.description }}
            </p>
          </article>
        </div>
      </div>
    </section>

    <!-- CTA -->
    <section class="mx-auto max-w-4xl px-5 py-20 text-center sm:px-6 lg:px-8">
      <p class="text-sm font-medium text-[#4f46e5]">
        Wiyata Academic Workspace
      </p>
      <h2
        class="mt-4 text-3xl font-medium tracking-tight text-[#171322] sm:text-4xl"
      >
        Masuk ke workspace sekolah yang sudah siap dipakai.
      </h2>
      <p class="mx-auto mt-4 max-w-2xl text-base leading-8 text-[#6b6475]">
        Gunakan akun Wiyata yang sudah terdaftar untuk melanjutkan pekerjaan
        belajar, mengajar, atau operasional sekolah dari tempat yang sama.
      </p>

      <RouterLink
        to="/login"
        class="mt-8 inline-flex h-12 items-center justify-center rounded-2xl bg-[#4f46e5] px-7 text-sm font-medium text-white transition hover:bg-[#4338ca]"
      >
        Masuk ke Wiyata
      </RouterLink>
    </section>

    <!-- Footer -->
    <footer class="border-t border-[#e7e2da] bg-white">
      <div
        class="mx-auto flex max-w-7xl flex-col gap-3 px-5 py-6 text-sm text-[#7a7385] sm:px-6 md:flex-row md:items-center md:justify-between lg:px-8"
      >
        <p>© 2026 Wiyata. All rights reserved.</p>
        <p>Learning Management System by Loka Wiyata</p>
      </div>
    </footer>
  </main>
</template>

<script setup lang="ts">
import { RouterLink } from "vue-router";

const previewStats = [
  { value: "8", label: "Mata pelajaran" },
  { value: "5", label: "Tugas aktif" },
  { value: "12", label: "Catatan" },
];

const activities = [
  {
    title: "Materi baru tersedia",
    meta: "Bahasa Indonesia · 10 menit lalu",
  },
  {
    title: "Tugas menunggu pengumpulan",
    meta: "Matematika · Tenggat minggu ini",
  },
  {
    title: "Pengumuman kelas",
    meta: "12 IPA 1 · Hari ini",
  },
];

const workspaceStory = [
  {
    title: "Sekolah disiapkan lebih dulu",
    description:
      "Tahun ajaran, semester, kelas, mata pelajaran, dan struktur sekolah dibangun lebih dulu agar operasional berikutnya tidak berantakan.",
  },
  {
    title: "Pembelajaran berjalan dari konteks yang jelas",
    description:
      "Materi, tugas, pengumpulan, dan nilai hadir di tempat yang sama sehingga guru dan siswa tidak perlu berpindah alur.",
  },
  {
    title: "Warga sekolah dikelola dengan tenang",
    description:
      "Penambahan warga sekolah, import CSV, dan pengelolaan membership tetap berada dalam konteks sekolah aktif.",
  },
  {
    title: "Percakapan ikut tinggal di dalam workspace",
    description:
      "Feed kelas, chat sekolah, dan grup diskusi hadir sebagai bagian dari kerja harian, bukan produk terpisah.",
  },
];

const roleSummaries = [
  {
    anchor: "siswa",
    eyebrow: "Siswa",
    title: "Belajar dengan arah yang jelas",
    description:
      "Buka materi, kirim tugas, simpan catatan pribadi, lihat nilai, dan ikuti percakapan sekolah dari satu tempat yang terasa ringan dipakai.",
  },
  {
    anchor: "guru",
    eyebrow: "Guru",
    title: "Mengajar tanpa panel yang tercerai",
    description:
      "Materi, tugas, pengumpulan, penilaian, dan pengumuman tersusun dalam satu workspace sehingga proses mengajar lebih mudah dipantau.",
  },
  {
    anchor: "admin",
    eyebrow: "Admin Sekolah",
    title: "Menjaga sekolah tetap rapi",
    description:
      "Kelola struktur akademik, kelas, penugasan guru, dan warga sekolah dalam konteks sekolah aktif yang jelas dan aman.",
  },
  {
    anchor: "super-admin",
    eyebrow: "Super Admin",
    title: "Mengelola struktur platform",
    description:
      "Sekolah dan akun global tetap bisa diatur dari lapisan platform, tanpa mencampur operasional akademik harian sekolah.",
  },
];

const workflows = [
  {
    eyebrow: "Akademik",
    title: "Struktur sekolah dibentuk lebih dulu",
    description:
      "Tahun ajaran, semester, kelas, mata pelajaran, dan penugasan guru disusun sebagai fondasi sebelum pembelajaran berjalan.",
  },
  {
    eyebrow: "Pembelajaran",
    title: "Materi, tugas, dan nilai saling terhubung",
    description:
      "Guru membagikan materi, memberi tugas, meninjau pengumpulan, lalu siswa melihat hasil belajar dan umpan balik di alur yang sama.",
  },
  {
    eyebrow: "Komunikasi",
    title: "Percakapan tidak lepas dari konteks sekolah",
    description:
      "Feed kelas, chat sekolah, dan grup diskusi membantu informasi berjalan tanpa memisahkan aktivitas akademik dari komunikasi.",
  },
  {
    eyebrow: "Administrasi sekolah",
    title: "Warga sekolah dikelola seperlunya",
    description:
      "Penambahan warga sekolah, import CSV, dan perubahan membership tetap berada dalam kontrol admin sekolah yang sedang aktif.",
  },
];
</script>
