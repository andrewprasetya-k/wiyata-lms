<script setup lang="ts">
const setupLinks = [
  {
    title: "Struktur Akademik",
    description:
      "Kelola tahun ajaran, semester, mata pelajaran, dan kategori tugas.",
    to: "/admin/academic-years",
  },
  {
    title: "Kelas",
    description: "Buat kelas pada tahun ajaran dan semester yang aktif.",
    to: "/admin/classes",
  },
  {
    title: "Warga Sekolah",
    description: "Kelola membership sekolah dan role per warga sekolah.",
    to: "/admin/users",
  },
  {
    title: "Penempatan Kelas",
    description: "Tempatkan student atau teacher ke kelas aktif.",
    to: "/admin/enrollments",
  },
  {
    title: "Penugasan Mengajar",
    description:
      "Hubungkan teacher, kelas, dan mata pelajaran untuk workspace mengajar.",
    to: "/admin/subject-classes",
  },
];
</script>

<template>
  <main class="min-h-screen flex-1 px-5 py-5 sm:px-6 lg:px-8">
    <section class="flex w-full max-w-none flex-col gap-5">
      <header class="soft-card rounded-[22px] p-5 md:p-6">
        <p class="text-sm font-medium text-[#4f46e5]">Admin sekolah</p>
        <h1 class="mt-2 text-2xl font-medium text-[#171322]">
          Dashboard admin sekolah
        </h1>
        <p class="mt-3 max-w-3xl text-sm leading-6 text-[#6b6475]">
          Gunakan halaman ini sebagai hub setup sekolah. Mulai dari struktur
          akademik, kelas, warga sekolah, penempatan kelas, sampai penugasan
          mengajar.
        </p>
      </header>

      <section class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        <RouterLink
          v-for="item in setupLinks"
          :key="item.to"
          :to="item.to"
          class="rounded-[20px] border border-[#ebe7df] bg-white p-5 shadow-sm transition hover:border-[#4f46e5] hover:shadow-md"
        >
          <p class="text-base font-medium text-[#171322]">{{ item.title }}</p>
          <p class="mt-2 text-sm leading-6 text-[#6b6475]">
            {{ item.description }}
          </p>
          <p class="mt-4 text-sm font-medium text-[#4f46e5]">Buka halaman</p>
        </RouterLink>
      </section>
    </section>
  </main>
</template>
