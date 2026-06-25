<script></script>
<template>
  <main class="max-w-screen bg-[#fbfaf8] text-[#171322]">
    <!-- Navbar -->
    <header class="sticky top-0 z-50 bg-white/50 backdrop-blur-xl">
      <div
        class="mx-auto flex max-w-7xl items-center justify-between px-5 py-4 sm:px-6 lg:px-8"
      >
        <RouterLink to="/" class="flex items-center gap-3">
          <img
            src="/logo_fix.svg"
            alt="Eduverse"
            class="h-11 w-11 rounded-2xl object-contain"
          />
          <div>
            <p class="text-sm font-medium text-[#4f46e5]">Eduverse</p>
            <p class="text-xs text-[#6b7280]">Academic workspace</p>
          </div>
        </RouterLink>

        <nav class="hidden items-center gap-7 text-sm text-[#6b6475] md:flex">
          <a href="#fitur" class="transition hover:text-[#171322]">Fitur</a>
          <a href="#siswa" class="transition hover:text-[#171322]">Siswa</a>
          <a href="#guru" class="transition hover:text-[#171322]">Guru</a>
          <a href="#admin" class="transition hover:text-[#171322]">Admin</a>
        </nav>

        <RouterLink
          to="/auth/login"
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
        <h1
          class="max-w-3xl text-4xl font-medium leading-tight tracking-tight text-[#171322] sm:text-5xl lg:text-6xl"
        >
          Ruang belajar digital yang lebih tenang.
        </h1>

        <p class="mt-6 max-w-2xl text-base leading-8 text-[#6b6475] sm:text-lg">
          Eduverse membantu siswa, guru, dan admin sekolah mengelola materi,
          tugas, catatan, nilai, pengumuman, dan kelas dalam satu workspace
          akademik.
        </p>

        <div class="mt-8 flex flex-col gap-3 sm:flex-row">
          <RouterLink
            to="/auth/login"
            class="inline-flex h-12 items-center justify-center rounded-2xl bg-[#4f46e5] px-6 text-sm font-medium text-white transition hover:bg-[#4338ca]"
          >
            Masuk ke Eduverse
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
        class="rounded-4xl border border-[#e7e2da] bg-white p-3 shadow-xl shadow-[#d8d2c6]/30"
      >
        <div class="rounded-3xl bg-[#eef2ff] p-4">
          <div class="rounded-[1.25rem] border border-white/80 bg-white p-5">
            <div class="mb-5 flex items-center justify-between gap-4">
              <div>
                <p class="text-xs font-medium text-[#7a7385]">
                  Dashboard siswa
                </p>
                <h2 class="mt-1 text-xl font-medium text-[#171322]">
                  Selamat datang kembali
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

    <!-- Product Principles -->
    <section id="fitur" class="mx-auto max-w-7xl px-5 pb-20 sm:px-6 lg:px-8">
      <div class="grid gap-4 md:grid-cols-3">
        <article
          v-for="principle in principles"
          :key="principle.title"
          class="rounded-3xl border border-[#e7e2da] bg-white p-6"
        >
          <div
            class="mb-5 flex h-11 w-11 items-center justify-center rounded-2xl bg-[#eef2ff] text-lg"
          >
            {{ principle.icon }}
          </div>
          <h2 class="text-lg font-medium text-[#171322]">
            {{ principle.title }}
          </h2>
          <p class="mt-3 text-sm leading-7 text-[#6b6475]">
            {{ principle.description }}
          </p>
        </article>
      </div>
    </section>

    <!-- Role Sections -->
    <section class="mx-auto max-w-7xl space-y-6 px-5 pb-20 sm:px-6 lg:px-8">
      <section
        id="siswa"
        class="grid gap-8 rounded-4xl border border-[#e7e2da] bg-white p-6 lg:grid-cols-[0.9fr_1.1fr] lg:p-8"
      >
        <RoleIntro
          eyebrow="Untuk siswa"
          title="Belajar lebih terarah dari satu tempat."
          description="Siswa dapat membuka materi, mengumpulkan tugas, menulis catatan pribadi, membaca pengumuman, dan melihat nilai tanpa berpindah platform."
        />

        <FeatureGrid :features="studentFeatures" />
      </section>

      <section
        id="guru"
        class="grid gap-8 rounded-4xl border border-[#e7e2da] bg-white p-6 lg:grid-cols-[0.9fr_1.1fr] lg:p-8"
      >
        <RoleIntro
          eyebrow="Untuk guru"
          title="Ruang kerja mengajar yang rapi."
          description="Guru dapat mengelola materi, membuat tugas, memantau pengumpulan, memberi nilai, dan membagikan pengumuman kelas."
        />

        <FeatureGrid :features="teacherFeatures" />
      </section>

      <section
        id="admin"
        class="grid gap-8 rounded-4xl border border-[#e7e2da] bg-white p-6 lg:grid-cols-[0.9fr_1.1fr] lg:p-8"
      >
        <RoleIntro
          eyebrow="Untuk admin sekolah"
          title="Administrasi akademik lebih terstruktur."
          description="Admin sekolah dapat mengatur tahun ajaran, kelas, warga sekolah, penempatan siswa, dan penugasan guru."
        />

        <FeatureGrid :features="adminFeatures" />
      </section>
    </section>

    <!-- Showcase -->
    <section class="border-y border-[#e7e2da] bg-white">
      <div class="mx-auto max-w-7xl px-5 py-20 sm:px-6 lg:px-8">
        <div class="max-w-2xl">
          <p class="text-sm font-medium text-[#4f46e5]">Tampilan produk</p>
          <h2
            class="mt-3 text-3xl font-medium tracking-tight text-[#171322] sm:text-4xl"
          >
            Dibuat untuk aktivitas akademik sehari-hari.
          </h2>
          <p class="mt-4 text-base leading-8 text-[#6b6475]">
            Eduverse menyatukan pengalaman siswa, guru, dan admin dalam visual
            yang konsisten, tenang, dan mudah dipindai.
          </p>
        </div>

        <div class="mt-10 grid gap-5 lg:grid-cols-3">
          <ShowcaseCard
            title="Dashboard siswa"
            description="Ringkasan mata pelajaran, tugas, catatan, nilai, feed, dan notifikasi."
          />

          <ShowcaseCard
            title="Workspace guru"
            description="Kelola materi, tugas, pengumpulan, penilaian, dan pengumuman kelas."
          />

          <ShowcaseCard
            title="Admin sekolah"
            description="Atur struktur akademik, kelas, warga sekolah, dan penugasan mengajar."
          />
        </div>
      </div>
    </section>

    <!-- CTA -->
    <section class="mx-auto max-w-4xl px-5 py-20 text-center sm:px-6 lg:px-8">
      <p class="text-sm font-medium text-[#4f46e5]">
        Eduverse Academic Workspace
      </p>
      <h2
        class="mt-4 text-3xl font-medium tracking-tight text-[#171322] sm:text-4xl"
      >
        Masuk ke ruang belajar yang lebih tenang.
      </h2>
      <p class="mx-auto mt-4 max-w-2xl text-base leading-8 text-[#6b6475]">
        Gunakan akun Eduverse yang sudah terdaftar untuk masuk ke workspace
        sesuai role dan konteks sekolah.
      </p>

      <RouterLink
        to="/auth/login"
        class="mt-8 inline-flex h-12 items-center justify-center rounded-2xl bg-[#4f46e5] px-7 text-sm font-medium text-white transition hover:bg-[#4338ca]"
      >
        Masuk ke Eduverse
      </RouterLink>
    </section>

    <!-- Footer -->
    <footer class="border-t border-[#e7e2da] bg-white">
      <div
        class="mx-auto flex max-w-7xl flex-col gap-3 px-5 py-6 text-sm text-[#7a7385] sm:px-6 md:flex-row md:items-center md:justify-between lg:px-8"
      >
        <p>© 2026 Eduverse. All rights reserved.</p>
        <p>Learning Management System untuk sekolah modern.</p>
      </div>
    </footer>
  </main>
</template>

<script setup lang="ts">
import { defineComponent } from "vue";
import { RouterLink } from "vue-router";

type Feature = {
  title: string;
  description: string;
};

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

const principles = [
  {
    icon: "✦",
    title: "Tenang",
    description:
      "Antarmuka dibuat bersih dan fokus agar aktivitas belajar tidak terasa ramai.",
  },
  {
    icon: "◇",
    title: "Terstruktur",
    description:
      "Materi, tugas, pengumpulan, nilai, dan administrasi sekolah berada dalam alur yang jelas.",
  },
  {
    icon: "◌",
    title: "Satu workspace",
    description:
      "Siswa, guru, dan admin sekolah bekerja di platform yang sama dengan ruang kerja sesuai peran.",
  },
];

const studentFeatures: Feature[] = [
  {
    title: "Materi & lampiran",
    description: "Akses materi dengan preview gambar, PDF, dan file pendukung.",
  },
  {
    title: "Tugas & pengumpulan",
    description: "Kumpulkan tugas dan pantau status penilaian dari guru.",
  },
  {
    title: "Catatan pribadi",
    description: "Tulis catatan per materi yang hanya bisa dilihat oleh siswa.",
  },
  {
    title: "Nilai & umpan balik",
    description: "Lihat nilai, status tugas, dan umpan balik pembelajaran.",
  },
];

const teacherFeatures: Feature[] = [
  {
    title: "Workspace mengajar",
    description: "Kelola materi, tugas, dan pengumpulan per mata pelajaran.",
  },
  {
    title: "Penilaian tugas",
    description: "Tinjau jawaban siswa, beri nilai, dan tulis umpan balik.",
  },
  {
    title: "Pengumuman kelas",
    description: "Bagikan informasi kelas dan kelola komentar siswa.",
  },
  {
    title: "Lampiran pembelajaran",
    description: "Tambahkan file pendukung untuk materi maupun tugas.",
  },
];

const adminFeatures: Feature[] = [
  {
    title: "Struktur akademik",
    description:
      "Atur tahun ajaran, semester, mata pelajaran, dan kategori tugas.",
  },
  {
    title: "Manajemen kelas",
    description: "Buat kelas dan hubungkan dengan periode akademik.",
  },
  {
    title: "Warga sekolah",
    description: "Kelola siswa, guru, dan admin sekolah dalam satu konteks.",
  },
  {
    title: "Penugasan mengajar",
    description: "Hubungkan guru dengan kelas dan mata pelajaran.",
  },
];

const RoleIntro = defineComponent({
  props: {
    eyebrow: {
      type: String,
      required: true,
    },
    title: {
      type: String,
      required: true,
    },
    description: {
      type: String,
      required: true,
    },
  },
  template: `
    <div>
      <p class="text-sm font-medium text-[#4f46e5]">{{ eyebrow }}</p>
      <h2 class="mt-3 text-3xl font-medium tracking-tight text-[#171322]">
        {{ title }}
      </h2>
      <p class="mt-4 text-base leading-8 text-[#6b6475]">
        {{ description }}
      </p>
    </div>
  `,
});

const FeatureGrid = defineComponent({
  props: {
    features: {
      type: Array as () => Feature[],
      required: true,
    },
  },
  template: `
    <div class="grid gap-3 sm:grid-cols-2">
      <article
        v-for="feature in features"
        :key="feature.title"
        class="rounded-2xl border border-[#e7e2da] bg-[#fbfaf8] p-5"
      >
        <h3 class="text-sm font-medium text-[#171322]">{{ feature.title }}</h3>
        <p class="mt-2 text-sm leading-6 text-[#6b6475]">
          {{ feature.description }}
        </p>
      </article>
    </div>
  `,
});

const ShowcaseCard = defineComponent({
  props: {
    title: {
      type: String,
      required: true,
    },
    description: {
      type: String,
      required: true,
    },
  },
  template: `
    <article class="rounded-[2rem] border border-[#e7e2da] bg-[#fbfaf8] p-4">
      <div class="aspect-[4/3] rounded-[1.5rem] border border-[#e7e2da] bg-white p-4">
        <div class="flex items-center justify-between">
          <div>
            <div class="h-3 w-24 rounded-full bg-[#e7e2da]"></div>
            <div class="mt-2 h-3 w-16 rounded-full bg-[#eef2ff]"></div>
          </div>
          <div class="h-9 w-9 rounded-2xl bg-[#4f46e5]"></div>
        </div>

        <div class="mt-5 grid gap-3">
          <div class="h-14 rounded-2xl bg-[#fbfaf8]"></div>
          <div class="h-14 rounded-2xl bg-[#fbfaf8]"></div>
          <div class="h-14 rounded-2xl bg-[#fbfaf8]"></div>
        </div>
      </div>

      <h3 class="mt-5 text-base font-medium text-[#171322]">{{ title }}</h3>
      <p class="mt-2 text-sm leading-6 text-[#6b6475]">{{ description }}</p>
    </article>
  `,
});
</script>
