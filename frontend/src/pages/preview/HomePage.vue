<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
import { RouterLink } from "vue-router";
import {
  PhList,
  PhX,
  PhArrowRight,
  PhArrowDown,
  PhMonitor,
  PhPlay,
  PhCheckCircle,
  PhDot,
  PhChalkboardTeacher,
  PhShieldCheck,
  PhChatCircle,
} from "@phosphor-icons/vue";
import Lenis from "lenis";

// ── Mobile menu state
const mobileOpen = ref(false);

// ── Lenis smooth scroll — scoped to this page only
let lenis: Lenis | null = null;
let rafId: number | null = null;

onMounted(() => {
  // Respect prefers-reduced-motion
  const prefersReduced = window.matchMedia(
    "(prefers-reduced-motion: reduce)",
  ).matches;
  if (prefersReduced) return;

  lenis = new Lenis({
    duration: 1.1,
    easing: (t: number) => Math.min(1, 1.001 - Math.pow(2, -10 * t)),
    smoothWheel: true,
  });

  function raf(time: number) {
    lenis!.raf(time);
    rafId = requestAnimationFrame(raf);
  }
  rafId = requestAnimationFrame(raf);
});

onUnmounted(() => {
  if (rafId !== null) cancelAnimationFrame(rafId);
  lenis?.destroy();
  lenis = null;
  rafId = null;
});

// ── Features — editorial narrative list
const features = [
  {
    title: "Kelas & Materi",
    description:
      "Guru membuat dan membagikan materi dalam konteks mata pelajaran yang jelas. Siswa membaca, membuat catatan pribadi, dan melacak progres dari satu tempat yang tidak membingungkan.",
    points: [
      "Materi per mata pelajaran & kelas",
      "Catatan pribadi siswa per materi",
      "Alur yang sama untuk guru dan siswa",
    ],
  },
  {
    title: "Tugas & Penilaian",
    description:
      "Buat tugas dengan tenggat waktu, terima pengumpulan, dan berikan nilai dan catatan umpan balik. Semua dalam satu alur yang terorganisir, tanpa berpindah halaman.",
    points: [
      "Pembuatan tugas dengan deadline",
      "Pengumpulan dan review oleh guru",
      "Nilai dan umpan balik per siswa",
    ],
  },
  {
    title: "Chat Akademik",
    description:
      "Percakapan tidak lepas dari konteks sekolah. Feed kelas, chat antar warga sekolah, dan diskusi hadir sebagai bagian dari workspace — bukan aplikasi terpisah.",
    points: [
      "Chat per sekolah dan per grup",
      "Feed pengumuman kelas",
      "Terhubung dengan konteks akademik",
    ],
  },
  {
    title: "Notifikasi Aktivitas",
    description:
      "Setiap aktivitas penting — materi baru, pengumpulan tugas, nilai keluar, pengumuman — tersimpan dalam notifikasi yang bisa dibaca kapan saja tanpa harus aktif memantau.",
    points: [
      "Notifikasi per aktivitas penting",
      "Tidak perlu refresh atau cek manual",
      "Terpusat dalam satu panel",
    ],
  },
  {
    title: "Multi-role Workspace",
    description:
      "Satu platform untuk semua peran sekolah. Setiap pengguna mendapat tampilan dan akses yang sesuai dengan peran aktifnya — tanpa fitur yang tidak relevan menghalangi.",
    points: [
      "Siswa, guru, admin, super admin",
      "Akses terbatas sesuai konteks peran",
      "Konteks sekolah yang terisolasi dan aman",
    ],
  },
];

// ── Roles
const mainRoles = [
  {
    anchor: "siswa",
    eyebrow: "Untuk Siswa",
    title: "Belajar dengan arah yang jelas",
    description:
      "Buka materi, kirim tugas, simpan catatan pribadi, dan pantau nilai dari satu tempat yang ringan dipakai setiap hari.",
    points: [
      "Akses materi per mata pelajaran",
      "Pengumpulan tugas & pantau nilai",
      "Catatan pribadi per materi",
    ],
  },
  {
    anchor: "guru",
    eyebrow: "Untuk Guru",
    title: "Mengajar tanpa alat yang tercerai",
    description:
      "Materi, tugas, pengumpulan, penilaian, dan pengumuman tersusun dalam satu workspace — proses mengajar lebih mudah dipantau.",
    points: [
      "Kelola materi & tugas per kelas",
      "Tinjau pengumpulan & beri nilai",
      "Pengumuman kelas terintegrasi",
    ],
  },
  {
    anchor: "admin",
    eyebrow: "Untuk Admin Sekolah",
    title: "Menjaga sekolah tetap terstruktur",
    description:
      "Kelola struktur akademik, kelas, penugasan guru, dan warga sekolah dalam konteks sekolah aktif yang jelas dan aman.",
    points: [
      "Kelola tahun ajaran & kelas",
      "Atur penugasan guru ke mata pelajaran",
      "Manajemen warga sekolah",
    ],
  },
];

const adminRoles = [
  {
    anchor: "super-admin",
    eyebrow: "Super Admin",
    title: "Mengelola struktur platform dari lapisan teratas",
    description:
      "Sekolah dan akun global bisa diatur dari lapisan platform tanpa harus ikut masuk ke operasional akademik harian masing-masing sekolah.",
  },
];

// ── Screenshot slots for media section
const screenshotSlots = [
  {
    label: "Tampilan Guru",
    note: "Ganti dengan screenshot halaman guru",
    icon: PhChalkboardTeacher,
  },
  {
    label: "Tampilan Admin",
    note: "Ganti dengan screenshot halaman admin",
    icon: PhShieldCheck,
  },
  {
    label: "Tampilan Chat & Feed",
    note: "Ganti dengan screenshot halaman komunikasi",
    icon: PhChatCircle,
  },
];
</script>
<template>
  <main class="overflow-x-hidden bg-[#fbfaf8] text-[#171322]">
    <!-- ───────────── NAVBAR ───────────── -->
    <header
      class="sticky top-0 z-50 border-b border-[#e7e2da]/70 bg-[#fbfaf8]/95 backdrop-blur-md"
    >
      <div
        class="mx-auto flex max-w-7xl items-center justify-between px-6 py-4 lg:px-8"
      >
        <!-- Wordmark -->
        <RouterLink to="/home" class="flex items-center gap-2.5">
          <img
            src="/logo_fix.svg"
            alt="Wiyata"
            class="h-7 w-7 rounded-lg object-contain"
          />
          <span class="text-[15px] font-semibold tracking-tight text-[#171322]"
            >Wiyata</span
          >
        </RouterLink>

        <!-- Desktop nav links -->
        <nav class="hidden items-center gap-8 text-sm text-[#6b7280] md:flex">
          <a href="#fitur" class="transition-colors hover:text-[#171322]"
            >Fitur</a
          >
          <a href="#peran" class="transition-colors hover:text-[#171322]"
            >Untuk Siapa</a
          >
          <a href="#preview" class="transition-colors hover:text-[#171322]"
            >Preview</a
          >
        </nav>

        <!-- CTA + mobile -->
        <div class="flex items-center gap-3">
          <RouterLink
            to="/login"
            class="rounded-lg bg-[#4f46e5] px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-[#4338ca]"
          >
            Masuk
          </RouterLink>
          <button
            id="nav-mobile-toggle"
            class="flex h-9 w-9 items-center justify-center rounded-lg border border-[#e7e2da] bg-white md:hidden"
            aria-label="Buka menu"
            @click="mobileOpen = !mobileOpen"
          >
            <PhList v-if="!mobileOpen" :size="18" class="text-[#6b7280]" />
            <PhX v-else :size="18" class="text-[#6b7280]" />
          </button>
        </div>
      </div>

      <!-- Mobile drawer -->
      <Transition
        enter-active-class="transition-all duration-200 ease-out"
        enter-from-class="opacity-0 -translate-y-1"
        enter-to-class="opacity-100 translate-y-0"
        leave-active-class="transition-all duration-150 ease-in"
        leave-from-class="opacity-100 translate-y-0"
        leave-to-class="opacity-0 -translate-y-1"
      >
        <nav
          v-if="mobileOpen"
          class="border-t border-[#e7e2da] bg-white md:hidden"
        >
          <div class="flex flex-col gap-1 px-6 py-4">
            <a
              href="#fitur"
              @click="mobileOpen = false"
              class="rounded-md px-3 py-2.5 text-sm text-[#6b7280] hover:bg-[#f8f7f4] hover:text-[#171322]"
              >Fitur</a
            >
            <a
              href="#peran"
              @click="mobileOpen = false"
              class="rounded-md px-3 py-2.5 text-sm text-[#6b7280] hover:bg-[#f8f7f4] hover:text-[#171322]"
              >Untuk Siapa</a
            >
            <a
              href="#preview"
              @click="mobileOpen = false"
              class="rounded-md px-3 py-2.5 text-sm text-[#6b7280] hover:bg-[#f8f7f4] hover:text-[#171322]"
              >Preview</a
            >
          </div>
        </nav>
      </Transition>
    </header>

    <!-- ───────────── HERO ───────────── -->
    <section class="mx-auto max-w-7xl px-6 pb-0 pt-24 lg:px-8 lg:pt-32">
      <!-- Eyebrow -->
      <p class="text-sm font-medium tracking-wide text-[#4f46e5]">
        Wiyata Academic Workspace
      </p>

      <!-- Headline -->
      <h1
        class="mt-5 max-w-3xl text-5xl font-semibold leading-[1.1] tracking-tight text-[#171322] sm:text-6xl lg:text-[68px]"
      >
        Satu ruang kerja untuk aktivitas akademik sekolah.
      </h1>

      <!-- Sub-headline -->
      <p class="mt-6 max-w-2xl text-lg leading-8 text-[#6b7280]">
        Kelola materi, tugas, penilaian, komunikasi, dan aktivitas kelas dalam
        workspace yang rapi dan terhubung — untuk siswa, guru, dan admin
        sekolah.
      </p>

      <!-- CTAs -->
      <div
        class="mt-9 flex flex-col items-start gap-3 sm:flex-row sm:items-center"
      >
        <RouterLink
          to="/login"
          id="hero-cta-masuk"
          class="inline-flex h-11 items-center justify-center rounded-lg bg-[#4f46e5] px-6 text-sm font-medium text-white transition-colors hover:bg-[#4338ca]"
        >
          Masuk ke Wiyata
        </RouterLink>
        <a
          href="#preview"
          id="hero-cta-preview"
          class="inline-flex h-11 items-center justify-center gap-1.5 rounded-lg border border-[#e7e2da] bg-white px-6 text-sm font-medium text-[#171322] transition-colors hover:bg-[#f8f7f4]"
        >
          Lihat preview
          <PhArrowDown :size="14" class="text-[#6b7280]" />
        </a>
      </div>

      <!-- ── LARGE PRODUCT PREVIEW PLACEHOLDER ── -->
      <!-- TODO: Replace this placeholder with actual Wiyata product screenshot or demo video. -->
      <div class="mt-16">
        <div
          class="relative overflow-hidden rounded-2xl border border-[#e7e2da] bg-white shadow-[0_4px_48px_-8px_rgba(23,19,34,0.10)]"
        >
          <!-- Browser chrome -->
          <div
            class="flex items-center gap-2 border-b border-[#f0ece5] bg-[#faf9f7] px-5 py-3"
          >
            <span class="h-3 w-3 rounded-full bg-[#fca5a5]" />
            <span class="h-3 w-3 rounded-full bg-[#fcd34d]" />
            <span class="h-3 w-3 rounded-full bg-[#86efac]" />
            <div
              class="ml-3 flex h-6 flex-1 max-w-xs items-center rounded-md bg-[#f0ece5] px-3 text-xs text-[#9ca3af]"
            >
              app.wiyata.id/student/dashboard
            </div>
          </div>

          <!-- Placeholder content area -->
          <div
            class="flex aspect-[16/9] w-full flex-col items-center justify-center gap-4 bg-[#f8f7f4] px-8 text-center"
            style="min-height: 380px"
          >
            <div
              class="flex h-14 w-14 items-center justify-center rounded-xl border border-[#e7e2da] bg-white shadow-sm"
            >
              <PhMonitor :size="26" class="text-[#4f46e5]" />
            </div>
            <div>
              <p class="text-base font-medium text-[#171322]">Preview Wiyata</p>
              <p class="mt-1.5 max-w-sm text-sm leading-relaxed text-[#9ca3af]">
                Tempatkan screenshot dashboard atau video demo di sini.
              </p>
            </div>
            <!-- Skeleton rows that signal "real UI is coming" -->
            <div class="mt-4 w-full max-w-lg space-y-2.5 opacity-40">
              <div class="h-2.5 rounded-full bg-[#e7e2da]" />
              <div class="h-2.5 w-4/5 rounded-full bg-[#e7e2da]" />
              <div class="h-2.5 w-3/5 rounded-full bg-[#e7e2da]" />
            </div>
          </div>
        </div>

        <p class="mt-3 text-center text-xs text-[#9ca3af]">
          Screenshot produk nyata akan ditambahkan di sini.
        </p>
      </div>
    </section>

    <!-- ───────────── FEATURE STORY ───────────── -->
    <section id="fitur" class="mx-auto max-w-7xl px-6 py-32 lg:px-8">
      <!-- Section label -->
      <div class="max-w-2xl">
        <p class="text-sm font-medium text-[#4f46e5]">Fitur utama</p>
        <h2
          class="mt-4 text-4xl font-semibold tracking-tight text-[#171322] sm:text-5xl"
        >
          Semua yang dibutuhkan sekolah dalam satu tempat.
        </h2>
        <p class="mt-5 text-lg leading-8 text-[#6b7280]">
          Wiyata tidak mencoba menjadi segalanya. Platform ini fokus pada alur
          kerja akademik yang paling sering dijalankan sekolah setiap hari.
        </p>
      </div>

      <!-- Feature rows — editorial split layout -->
      <div
        class="mt-20 space-y-0 divide-y divide-[#e7e2da] border-t border-b border-[#e7e2da]"
      >
        <div
          v-for="(feature, i) in features"
          :key="feature.title"
          class="grid grid-cols-1 gap-8 py-10 md:grid-cols-[280px_1fr] md:gap-16 lg:gap-24"
        >
          <!-- Label column -->
          <div class="flex flex-col justify-start pt-0.5">
            <span
              class="mb-2 text-xs font-semibold uppercase tracking-widest text-[#4f46e5]"
            >
              {{ String(i + 1).padStart(2, "0") }}
            </span>
            <h3 class="text-xl font-semibold text-[#171322]">
              {{ feature.title }}
            </h3>
          </div>
          <!-- Description column -->
          <div class="flex flex-col justify-center">
            <p class="text-base leading-8 text-[#6b7280]">
              {{ feature.description }}
            </p>
            <ul class="mt-5 space-y-2">
              <li
                v-for="point in feature.points"
                :key="point"
                class="flex items-start gap-2.5 text-sm text-[#6b7280]"
              >
                <PhCheckCircle
                  :size="16"
                  class="mt-0.5 flex-shrink-0 text-[#4f46e5]"
                />
                {{ point }}
              </li>
            </ul>
          </div>
        </div>
      </div>
    </section>

    <!-- ───────────── ROLE SECTION ───────────── -->
    <section id="peran" class="border-t border-[#e7e2da] bg-white">
      <div class="mx-auto max-w-7xl px-6 py-32 lg:px-8">
        <div class="max-w-2xl">
          <p class="text-sm font-medium text-[#4f46e5]">Untuk setiap peran</p>
          <h2
            class="mt-4 text-4xl font-semibold tracking-tight text-[#171322] sm:text-5xl"
          >
            Satu workspace. Perspektif berbeda untuk tiap pengguna.
          </h2>
          <p class="mt-5 text-lg leading-8 text-[#6b7280]">
            Wiyata mengenali bahwa siswa, guru, dan admin punya ritme kerja yang
            berbeda. Setiap peran mendapat tampilan dan akses yang relevan.
          </p>
        </div>

        <!-- Three primary role columns -->
        <div
          class="mt-16 grid gap-px border border-[#e7e2da] bg-[#e7e2da] sm:grid-cols-3"
        >
          <article
            v-for="role in mainRoles"
            :key="role.title"
            :id="role.anchor"
            class="bg-white px-8 py-10"
          >
            <p
              class="text-xs font-semibold uppercase tracking-widest text-[#4f46e5]"
            >
              {{ role.eyebrow }}
            </p>
            <h3 class="mt-4 text-xl font-semibold text-[#171322]">
              {{ role.title }}
            </h3>
            <p class="mt-3 text-sm leading-7 text-[#6b7280]">
              {{ role.description }}
            </p>
            <ul class="mt-6 space-y-2">
              <li
                v-for="point in role.points"
                :key="point"
                class="flex items-start gap-2 text-sm text-[#6b7280]"
              >
                <PhDot :size="16" class="mt-1 flex-shrink-0 text-[#4f46e5]" />
                {{ point }}
              </li>
            </ul>
          </article>
        </div>

        <!-- Admin + Super Admin — horizontal card below -->
        <div
          class="mt-px grid gap-px border-x border-b border-[#e7e2da] bg-[#e7e2da] sm:grid-cols-2"
        >
          <article
            v-for="role in adminRoles"
            :key="role.title"
            :id="role.anchor"
            class="bg-white px-8 py-8"
          >
            <p
              class="text-xs font-semibold uppercase tracking-widest text-[#6b7280]"
            >
              {{ role.eyebrow }}
            </p>
            <h3 class="mt-3 text-lg font-semibold text-[#171322]">
              {{ role.title }}
            </h3>
            <p class="mt-2.5 text-sm leading-7 text-[#6b7280]">
              {{ role.description }}
            </p>
          </article>
        </div>
      </div>
    </section>

    <!-- ───────────── PRODUCT VIDEO / MEDIA PLACEHOLDER ───────────── -->
    <!-- TODO: Replace with product video or feature screenshot carousel later. -->
    <section id="preview" class="border-t border-[#e7e2da] bg-[#fbfaf8]">
      <div class="mx-auto max-w-7xl px-6 py-32 lg:px-8">
        <div class="mx-auto max-w-2xl text-center">
          <p class="text-sm font-medium text-[#4f46e5]">Demo produk</p>
          <h2
            class="mt-4 text-4xl font-semibold tracking-tight text-[#171322] sm:text-5xl"
          >
            Siapkan ruang untuk demo produk.
          </h2>
          <p class="mt-5 text-lg leading-8 text-[#6b7280]">
            Bagian ini dapat diisi dengan video singkat, walkthrough, atau
            kumpulan screenshot fitur Wiyata.
          </p>
        </div>

        <!-- Video placeholder -->
        <div class="mx-auto mt-14 max-w-4xl">
          <div
            class="relative flex aspect-video w-full flex-col items-center justify-center overflow-hidden rounded-2xl border border-[#e7e2da] bg-[#1a1830]"
          >
            <!-- Subtle dot grid -->
            <div
              class="pointer-events-none absolute inset-0 opacity-[0.06]"
              style="
                background-image: radial-gradient(
                  circle,
                  #ffffff 1px,
                  transparent 1px
                );
                background-size: 28px 28px;
              "
              aria-hidden="true"
            />

            <!-- Play button placeholder -->
            <div class="relative flex flex-col items-center gap-5 text-center">
              <button
                disabled
                aria-label="Video belum tersedia"
                class="flex h-16 w-16 cursor-default items-center justify-center rounded-full border border-white/20 bg-white/10"
              >
                <PhPlay :size="22" weight="fill" class="ml-1 text-white/70" />
              </button>
              <div>
                <p class="text-sm font-medium text-white/80">
                  Video demo akan ditambahkan di sini
                </p>
                <p class="mt-1 text-xs text-white/40">
                  Walkthrough produk Wiyata · Segera hadir
                </p>
              </div>
            </div>

            <!-- Duration badge -->
            <div
              class="absolute bottom-4 right-4 rounded-md bg-white/10 px-2.5 py-1 text-xs text-white/50"
            >
              ~2 menit
            </div>
          </div>

          <p class="mt-4 text-center text-xs text-[#9ca3af]">
            <!-- TODO: Replace this area with the actual Wiyata product demo video. -->
            Placeholder video. Ganti dengan demo produk nyata saat sudah siap.
          </p>
        </div>

        <!-- Three feature screenshot placeholders -->
        <div class="mt-10 grid gap-4 sm:grid-cols-3">
          <div
            v-for="slot in screenshotSlots"
            :key="slot.label"
            class="flex aspect-[4/3] flex-col items-center justify-center gap-3 rounded-xl border border-[#e7e2da] bg-white text-center"
          >
            <component :is="slot.icon" :size="22" class="text-[#d1cde5]" />
            <div>
              <p class="text-xs font-medium text-[#9ca3af]">{{ slot.label }}</p>
              <p class="mt-0.5 text-[10px] text-[#c4bfcc]">{{ slot.note }}</p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ───────────── FINAL CTA ───────────── -->
    <section class="border-t border-[#e7e2da] bg-white">
      <div class="mx-auto max-w-7xl px-6 py-32 lg:px-8">
        <div class="grid gap-12 lg:grid-cols-[1fr_auto] lg:items-center">
          <div class="max-w-2xl">
            <p class="text-sm font-medium text-[#4f46e5]">
              Wiyata Academic Workspace
            </p>
            <h2
              class="mt-4 text-4xl font-semibold tracking-tight text-[#171322] sm:text-5xl"
            >
              Workspace sekolah yang sudah siap dipakai.
            </h2>
            <p class="mt-5 text-lg leading-8 text-[#6b7280]">
              Gunakan akun Wiyata yang sudah terdaftar untuk melanjutkan
              aktivitas belajar, mengajar, atau operasional sekolah dari satu
              tempat yang sama.
            </p>
            <p class="mt-3 text-sm text-[#9ca3af]">
              Untuk pengguna yang sudah memiliki akun Wiyata.
            </p>
          </div>

          <div class="flex flex-col items-start gap-3 lg:items-end">
            <RouterLink
              to="/login"
              id="final-cta-masuk"
              class="inline-flex h-12 items-center justify-center gap-2 rounded-lg bg-[#4f46e5] px-7 text-sm font-medium text-white transition-colors hover:bg-[#4338ca]"
            >
              Masuk ke Wiyata
              <PhArrowRight :size="15" />
            </RouterLink>
            <a
              href="#fitur"
              class="inline-flex h-12 items-center justify-center rounded-lg border border-[#e7e2da] bg-transparent px-7 text-sm font-medium text-[#6b7280] transition-colors hover:border-[#d1cde5] hover:text-[#171322]"
            >
              Pelajari fitur
            </a>
          </div>
        </div>
      </div>
    </section>

    <!-- ───────────── FOOTER ───────────── -->
    <footer class="border-t border-[#e7e2da] bg-[#fbfaf8]">
      <div class="mx-auto max-w-7xl px-6 py-8 lg:px-8">
        <div
          class="flex flex-col gap-6 sm:flex-row sm:items-center sm:justify-between"
        >
          <!-- Brand -->
          <div class="flex items-center gap-2.5">
            <img
              src="/logo_fix.svg"
              alt="Wiyata"
              class="h-6 w-6 rounded-md object-contain"
            />
            <span class="text-sm font-semibold text-[#171322]">Wiyata</span>
          </div>

          <!-- Footer links -->
          <nav class="flex flex-wrap gap-x-7 gap-y-2 text-sm text-[#9ca3af]">
            <a href="#fitur" class="hover:text-[#6b7280]">Fitur</a>
            <a href="#peran" class="hover:text-[#6b7280]">Untuk Siapa</a>
            <a href="#preview" class="hover:text-[#6b7280]">Preview</a>
            <RouterLink to="/login" class="hover:text-[#6b7280]"
              >Masuk</RouterLink
            >
          </nav>
        </div>

        <div
          class="mt-8 flex flex-col gap-1 border-t border-[#e7e2da] pt-6 text-xs text-[#c4bfcc] sm:flex-row sm:justify-between"
        >
          <p>© 2026 Wiyata. All rights reserved.</p>
          <p>Learning Management System by Loka Wiyata</p>
        </div>
      </div>
    </footer>
  </main>
</template>
