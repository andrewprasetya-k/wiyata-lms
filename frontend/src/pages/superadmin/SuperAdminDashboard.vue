<script setup lang="ts">
import {
  PhArrowRight,
  PhBuildings,
  PhCompass,
  PhIdentificationBadge,
  PhShieldCheck,
  PhUsers,
} from "@phosphor-icons/vue";

const overviewCards = [
  {
    title: "Sekolah",
    description:
      "Kelola tenant sekolah Wiyata dari tingkat platform tanpa mengambil alih operasional akademik harian.",
    icon: PhBuildings,
    tone: "bg-[#fff4ee] text-[#ea580c]",
  },
  {
    title: "Akun Global",
    description:
      "Pantau identitas pengguna lintas sekolah dan siapkan akun awal untuk kebutuhan platform.",
    icon: PhUsers,
    tone: "bg-brand-soft text-brand",
  },
  {
    title: "Pengaturan Tenant",
    description:
      "Bantu proses awal agar sekolah punya admin sekolah yang dapat melanjutkan pengaturan akademik.",
    icon: PhCompass,
    tone: "bg-success-soft text-success",
  },
  {
    title: "Peran Platform",
    description:
      "Super Admin menjaga akses platform. Operasional akademik tetap berada di area Admin Sekolah.",
    icon: PhShieldCheck,
    tone: "bg-surface-strong text-muted",
  },
];

const setupFlow = [
  "Buat sekolah atau tenant sekolah.",
  "Buat atau pilih akun Admin Sekolah.",
  "Hubungkan akun ke sekolah yang tepat.",
  "Beri peran admin sekolah.",
  "Admin Sekolah melanjutkan pengaturan akademik.",
];

const quickActions = [
  {
    title: "Kelola sekolah",
    description: "Buka area sekolah platform yang saat ini masih disiapkan.",
    to: "/superadmin/schools",
    icon: PhBuildings,
  },
  {
    title: "Akun global",
    description: "Buka area akun global yang saat ini masih disiapkan.",
    to: "/superadmin/users",
    icon: PhUsers,
  },
  {
    title: "Profil platform",
    description: "Lihat informasi akun Super Admin yang sedang digunakan.",
    to: "/superadmin/profile",
    icon: PhIdentificationBadge,
  },
];
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-background">
    <header class="border-b border-border bg-surface">
      <div
        class="flex min-w-0 flex-col gap-3 px-5 py-5 sm:px-6 lg:flex-row lg:items-end lg:justify-between lg:px-8"
      >
        <div class="min-w-0">
          <p
            class="eyebrow"
          >
            Super Admin
          </p>
          <h1 class="mt-2 text-2xl font-semibold text-foreground sm:text-3xl">
            Pusat Platform Wiyata
          </h1>
          <p class="mt-2 max-w-3xl text-sm leading-6 text-muted">
            Kelola kesiapan tenant sekolah dan akun global. Pengaturan akademik,
            kelas, konten, dan penilaian tetap berada di area Admin Sekolah.
          </p>
        </div>
        <RouterLink
          to="/superadmin/schools"
          class="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-[#ea580c] px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-[#c2410c] sm:w-auto"
        >
          Buka sekolah
          <PhArrowRight :size="16" weight="bold" />
        </RouterLink>
      </div>
    </header>

    <section
      class="grid w-full max-w-none gap-6 px-5 py-6 sm:px-6 lg:px-8 xl:grid-cols-[minmax(0,1fr)_340px]"
    >
      <div class="flex min-w-0 flex-col gap-6">
        <section
          class="rounded-xl border border-border bg-surface p-5 shadow-sm"
        >
          <p
            class="eyebrow"
          >
            Pusat platform
          </p>
          <h2 class="mt-2 text-xl font-semibold text-foreground">
            Ruang kerja tingkat platform
          </h2>
          <p class="mt-2 max-w-3xl text-sm leading-6 text-muted">
            Halaman ini adalah ringkasan statis untuk orientasi Super Admin.
            Tidak ada statistik tanpa sumber data operasional yang ditampilkan.
          </p>
        </section>

        <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
          <article
            v-for="item in overviewCards"
            :key="item.title"
            class="rounded-xl border border-border bg-surface p-4 shadow-sm"
          >
            <div
              class="flex h-10 w-10 items-center justify-center rounded-xl"
              :class="item.tone"
            >
              <component :is="item.icon" :size="20" weight="duotone" />
            </div>
            <h3 class="mt-4 text-sm font-semibold text-foreground">
              {{ item.title }}
            </h3>
            <p class="mt-2 text-xs leading-5 text-muted">
              {{ item.description }}
            </p>
          </article>
        </section>

        <section
          class="grid gap-5 lg:grid-cols-[minmax(0,1.1fr)_minmax(0,0.9fr)]"
        >
          <article
            class="rounded-xl border border-border bg-surface p-5 shadow-sm"
          >
            <div class="flex items-start justify-between gap-4">
              <div class="min-w-0">
                <p class="text-sm font-semibold text-foreground">
                  Alur pengaturan awal tenant
                </p>
                <p class="mt-1 text-xs leading-5 text-muted">
                  Urutan kerja aman sebelum Admin Sekolah mulai mengelola data
                  akademik.
                </p>
              </div>
              <PhCompass
                :size="20"
                class="shrink-0 text-[#ea580c]"
                weight="duotone"
              />
            </div>

            <div class="mt-4 divide-y divide-[#f3f4f6]">
              <div
                v-for="(item, index) in setupFlow"
                :key="item"
                class="flex items-start gap-3 py-3"
              >
                <span
                  class="mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-[#fff4ee] text-xs font-semibold text-[#ea580c]"
                >
                  {{ index + 1 }}
                </span>
                <p class="min-w-0 text-sm leading-6 text-foreground-secondary">
                  {{ item }}
                </p>
              </div>
            </div>
          </article>

          <article
            class="rounded-xl border border-border bg-surface p-5 shadow-sm"
          >
            <p class="text-sm font-semibold text-foreground">
              Batas tanggung jawab
            </p>
            <p class="mt-2 text-sm leading-6 text-muted">
              Super Admin mengatur platform, tenant sekolah, dan akun global.
              Pengaturan dan operasional akademik setiap sekolah tetap berada di
              area Admin Sekolah.
            </p>
            <div
              class="mt-4 rounded-lg border border-warning-line bg-warning-soft px-4 py-3"
            >
              <p class="text-xs leading-5 text-warning">
                Halaman sekolah dan akun global masih menggunakan navigasi yang
                sudah ada. Implementasi operasionalnya akan dilakukan terpisah.
              </p>
            </div>
          </article>
        </section>
      </div>

      <aside class="flex min-w-0 flex-col gap-5">
        <section
          class="rounded-xl border border-border bg-surface p-5 shadow-sm"
        >
          <p class="text-sm font-semibold text-foreground">Aksi cepat</p>
          <p class="mt-1 text-xs leading-5 text-muted">
            Semua tautan memakai halaman Super Admin yang sudah tersedia.
          </p>

          <div class="mt-4 space-y-2">
            <RouterLink
              v-for="item in quickActions"
              :key="item.to"
              :to="item.to"
              class="flex items-center gap-3 rounded-lg p-3 transition hover:bg-surface-strong"
            >
              <span
                class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-[#fff4ee] text-[#ea580c]"
              >
                <component :is="item.icon" :size="20" weight="duotone" />
              </span>
              <span class="min-w-0 flex-1">
                <span class="block text-sm font-semibold text-foreground">
                  {{ item.title }}
                </span>
                <span class="mt-1 block text-xs leading-5 text-muted">
                  {{ item.description }}
                </span>
              </span>
              <PhArrowRight :size="16" class="shrink-0 text-muted" />
            </RouterLink>
          </div>
        </section>

        <section
          class="rounded-xl border border-border bg-surface p-5 shadow-sm"
        >
          <p class="text-sm font-semibold text-foreground">
            Prinsip akses platform
          </p>
          <div class="mt-4 space-y-3">
            <div class="flex gap-3">
              <span class="mt-1.5 h-2 w-2 shrink-0 rounded-full bg-[#ea580c]" />
              <p class="text-xs leading-5 text-muted">
                Tidak ada operasi akademik sekolah dari dashboard ini.
              </p>
            </div>
            <div class="flex gap-3">
              <span class="mt-1.5 h-2 w-2 shrink-0 rounded-full bg-brand" />
              <p class="text-xs leading-5 text-muted">
                Tidak ada angka atau status langsung yang dibuat tanpa sumber
                data nyata.
              </p>
            </div>
            <div class="flex gap-3">
              <span class="mt-1.5 h-2 w-2 shrink-0 rounded-full bg-success" />
              <p class="text-xs leading-5 text-muted">
                Tindakan destruktif platform tidak ditampilkan pada ringkasan
                ini.
              </p>
            </div>
          </div>
        </section>
      </aside>
    </section>
  </main>
</template>
