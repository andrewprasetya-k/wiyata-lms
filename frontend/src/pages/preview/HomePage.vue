<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import { RouterLink } from "vue-router";
import { useAuthStore } from "../../stores/auth";
import {
  PhList,
  PhX,
  PhArrowRight,
  PhPlay,
  PhMonitor,
  PhCheckCircle,
  PhDot,
  PhChalkboardTeacher,
  PhShieldCheck,
  PhChatCircle,
} from "@phosphor-icons/vue";
import Lenis from "lenis";

const auth = useAuthStore();
const isSchoolless = computed(() => auth.isAuthenticated && !auth.activeContext);

// ── Mobile menu state
const mobileOpen = ref(false);

// ── Lenis smooth scroll — scoped to this page only
let lenis: Lenis | null = null;
let rafId: number | null = null;

const handleAnchorClick = (e: MouseEvent, selector: string) => {
  if (lenis) {
    e.preventDefault();
    lenis.scrollTo(selector, {
      offset: -80,
      duration: 1.1,
    });
  }
};

onMounted(() => {
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

// ── Features
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
      "Nilai pengumpulan tugas & berikan umpan balik",
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

// ── Screenshot slots
const screenshotSlots = [
  {
    label: "Tampilan Guru",
    note: "Preview halaman guru segera ditambahkan",
    icon: PhChalkboardTeacher,
  },
  {
    label: "Tampilan Admin",
    note: "Preview halaman admin segera ditambahkan",
    icon: PhShieldCheck,
  },
  {
    label: "Tampilan Chat & Feed",
    note: "Preview komunikasi kelas segera ditambahkan",
    icon: PhChatCircle,
  },
];
</script>

<template>
  <main class="relative isolate overflow-x-hidden bg-[#fbfaf8] text-foreground">
    <!-- ── Global decorative background -->
    <div class="pointer-events-none absolute inset-0 -z-10 overflow-hidden">
      <!-- Indigo glow — hero anchor -->
      <div
        class="absolute left-1/2 top-0 h-130 w-225 -translate-x-1/2 rounded-full bg-brand/18 blur-3xl"
        aria-hidden="true"
      />
      <!-- Warm amber glow — mid-page -->
      <div
        class="absolute -left-40 top-250 h-120 w-120 rounded-full bg-[#f59e0b]/12 blur-3xl"
        aria-hidden="true"
      />
      <!-- Subtle dot grid -->
      <div
        class="absolute inset-0 opacity-[0.05]"
        style="
          background-image:
            linear-gradient(#171322 1px, transparent 1px),
            linear-gradient(90deg, #171322 1px, transparent 1px);
          background-size: 44px 44px;
        "
        aria-hidden="true"
      />
    </div>

    <!-- ───────────── NAVBAR ───────────── -->
    <header
      class="fixed left-0 right-0 top-0 z-50 w-full backdrop-blur-xl backdrop-saturate-150 transition-colors duration-300"
    >
      <div
        class="mx-auto flex max-w-7xl items-center justify-between px-6 py-4 lg:px-8"
      >
        <RouterLink to="/home" class="flex items-center gap-2.5">
          <img
            src="/logo_fix.svg"
            alt="Wiyata"
            class="h-7 w-7 rounded-lg object-contain"
          />
          <span class="text-[15px] font-semibold tracking-tight text-foreground">
            Wiyata Academic Workspace
          </span>
        </RouterLink>

        <nav class="hidden items-center gap-8 text-sm text-muted md:flex">
          <a
            href="#fitur"
            class="transition-colors hover:text-foreground"
            @click="handleAnchorClick($event, '#fitur')"
            >Fitur</a
          >
          <a
            href="#peran"
            class="transition-colors hover:text-foreground"
            @click="handleAnchorClick($event, '#peran')"
            >Untuk Siapa</a
          >
          <a
            href="#preview"
            class="transition-colors hover:text-foreground"
            @click="handleAnchorClick($event, '#preview')"
            >Preview</a
          >
        </nav>

        <div class="flex items-center gap-3">
          <template v-if="isSchoolless">
            <span class="hidden text-sm text-muted sm:block">
              {{ auth.user?.fullName?.split(" ")[0] }}
            </span>
            <button
              type="button"
              class="rounded-lg border border-[#e7e2da] bg-white px-4 py-2 text-sm font-medium text-[#5f5968] transition-colors hover:text-foreground"
              @click="auth.logout()"
            >
              Keluar
            </button>
          </template>
          <template v-else>
            <RouterLink
              to="/school-registration"
              class="hidden rounded-lg border border-[#e7e2da] bg-white px-4 py-2 text-sm font-medium text-[#5f5968] transition-colors hover:text-foreground sm:inline-flex"
            >
              Daftarkan Sekolah
            </RouterLink>
            <RouterLink
              to="/login"
              class="rounded-lg bg-brand px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-brand-hover"
            >
              Masuk
            </RouterLink>
          </template>
          <button
            id="nav-mobile-toggle"
            class="flex h-9 w-9 items-center justify-center rounded-lg border border-[#e7e2da] bg-white md:hidden"
            aria-label="Buka menu"
            @click="mobileOpen = !mobileOpen"
          >
            <PhList v-if="!mobileOpen" :size="18" class="text-muted" />
            <PhX v-else :size="18" class="text-muted" />
          </button>
        </div>
      </div>

      <Transition
        enter-active-class="transition-all duration-200 ease-out"
        enter-from-class="opacity-0 -translate-y-1"
        enter-to-class="opacity-100 translate-y-0"
        leave-active-class="transition-all duration-150 ease-in"
        leave-from-class="opacity-100 translate-y-0"
        leave-to-class="opacity-0 -translate-y-1"
      >
        <nav v-if="mobileOpen" class="border-[#e7e2da] bg-white md:hidden">
          <div class="flex flex-col gap-1 px-6 py-4">
            <a
              href="#fitur"
              class="rounded-md px-3 py-2.5 text-sm text-muted hover:bg-[#f8f7f4] hover:text-foreground"
              @click="
                mobileOpen = false;
                handleAnchorClick($event, '#fitur');
              "
              >Fitur</a
            >
            <a
              href="#peran"
              class="rounded-md px-3 py-2.5 text-sm text-muted hover:bg-[#f8f7f4] hover:text-foreground"
              @click="
                mobileOpen = false;
                handleAnchorClick($event, '#peran');
              "
              >Untuk Siapa</a
            >
            <a
              href="#preview"
              class="rounded-md px-3 py-2.5 text-sm text-muted hover:bg-[#f8f7f4] hover:text-foreground"
              @click="
                mobileOpen = false;
                handleAnchorClick($event, '#preview');
              "
              >Preview</a
            >
            <template v-if="isSchoolless">
              <div class="mt-2 border-t border-[#f0ece5] pt-2">
                <p class="px-3 py-1.5 text-xs text-[#9ca3af]">
                  {{ auth.user?.fullName }}
                </p>
                <button
                  type="button"
                  class="w-full rounded-md px-3 py-2.5 text-left text-sm text-muted hover:bg-[#f8f7f4] hover:text-foreground"
                  @click="auth.logout(); mobileOpen = false"
                >
                  Keluar
                </button>
              </div>
            </template>
          </div>
        </nav>
      </Transition>
    </header>

    <!-- ───────────── HERO ───────────── -->
    <section
      class="relative mx-auto max-w-7xl px-6 pb-0 pt-24 lg:px-8 lg:pt-32"
    >
      <!-- Hero-local blobs -->
      <div
        class="pointer-events-none absolute -left-30 top-16 -z-10 h-72 w-72 rounded-full bg-brand/20 blur-3xl"
        aria-hidden="true"
      />
      <div
        class="pointer-events-none absolute -right-40 top-32 -z-10 h-80 w-80 rounded-full bg-[#f59e0b]/14 blur-3xl"
        aria-hidden="true"
      />

      <!-- School-less user: logged in but no school context -->
      <template v-if="isSchoolless">
        <h1
          class="mt-5 max-w-3xl text-5xl font-semibold leading-[1.1] tracking-tight text-foreground sm:text-6xl lg:text-[68px]"
        >
          Selamat datang di Wiyata
        </h1>
        <p class="mt-6 max-w-2xl text-lg leading-8 text-muted">
          Akunmu berhasil dibuat dan kamu sudah masuk ke Wiyata. Saat ini
          akunmu belum terhubung ke sekolah mana pun. Setelah bergabung ke
          sekolah, kamu dapat mengakses kelas, materi, tugas, nilai, dan
          aktivitas akademik.
        </p>
        <div class="mt-9 flex flex-wrap items-center gap-4">
          <RouterLink
            to="/school-registration"
            class="inline-flex h-12 items-center justify-center rounded-lg bg-brand px-8 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-hover"
          >
            Daftarkan Sekolah
          </RouterLink>
          <p class="max-w-sm text-sm leading-6 text-[#9ca3af]">
            Sudah mendapat undangan? Buka link undangan yang dikirim ke
            emailmu untuk bergabung ke sekolah.
          </p>
        </div>
      </template>

      <!-- Visitor: not logged in -->
      <template v-else>
        <h1
          class="mt-5 max-w-3xl text-5xl font-semibold leading-[1.1] tracking-tight text-foreground sm:text-6xl lg:text-[68px]"
        >
          Satu workspace untuk aktivitas akademik sekolah.
        </h1>
        <p class="mt-6 max-w-2xl text-lg leading-8 text-muted">
          Kelola materi, tugas, komunikasi, dan penilaian dalam satu workspace
          bagi murid, guru, dan sekolah.
        </p>
        <!-- CTAs — three-tier hierarchy -->
        <div class="mt-9 flex flex-wrap items-center gap-3">
          <RouterLink
            to="/school-registration"
            id="hero-cta-daftar-sekolah"
            class="inline-flex h-12 items-center justify-center rounded-lg bg-brand px-8 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-hover"
          >
            Daftarkan Sekolah
          </RouterLink>
          <RouterLink
            to="/login"
            id="hero-cta-masuk"
            class="inline-flex h-11 items-center justify-center rounded-lg border border-[#e7e2da] bg-white px-6 text-sm font-medium text-[#5f5968] transition-colors hover:bg-[#f8f7f4] hover:text-foreground"
          >
            Masuk ke Wiyata
          </RouterLink>
        </div>
      </template>

      <!-- ── Product mockup (dashboard UI) ── -->
      <div class="mt-16">
        <div
          class="relative overflow-hidden rounded-2xl border border-[#e7e2da] bg-white shadow-[0_28px_90px_-30px_rgba(79,70,229,0.32)] ring-1 ring-white/70"
        >
          <!-- Browser chrome -->
          <div
            class="flex items-center gap-2 border-b border-[#f0ece5] bg-[#faf9f7] px-5 py-3"
          >
            <span class="h-3 w-3 rounded-full bg-[#fca5a5]" />
            <span class="h-3 w-3 rounded-full bg-[#fcd34d]" />
            <span class="h-3 w-3 rounded-full bg-[#86efac]" />
            <div
              class="ml-3 flex h-6 max-w-xs flex-1 items-center rounded-md bg-[#f0ece5] px-3 text-xs text-[#9ca3af]"
            >
              app.wiyata.id/student/dashboard
            </div>
          </div>

          <!-- Placeholder content area -->
          <div
            class="flex aspect-video w-full flex-col items-center justify-center gap-4 bg-[#f8f7f4] px-8 text-center"
            style="min-height: 380px"
          >
            <div
              class="flex h-14 w-14 items-center justify-center rounded-xl border border-[#e7e2da] bg-white shadow-sm"
            >
              <PhMonitor :size="26" class="text-brand" />
            </div>
            <div>
              <p class="text-base font-medium text-foreground">Preview Wiyata</p>
              <p class="mt-1.5 max-w-sm text-sm leading-relaxed text-[#9ca3af]">
                Screenshot dashboard Wiyata akan segera ditambahkan.
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
          Tampilan dashboard siswa Wiyata.
        </p>
      </div>
    </section>

    <!-- ───────────── FEATURE STORY ───────────── -->
    <section
      id="fitur"
      class="relative mx-auto max-w-7xl scroll-mt-24 border-t border-[#e7e2da] px-6 py-32 lg:px-8"
    >
      <div class="max-w-2xl">
        <p class="text-sm font-medium text-brand">Fitur utama</p>
        <h2
          class="mt-4 text-4xl font-semibold tracking-tight text-foreground sm:text-5xl"
        >
          Semua yang dibutuhkan sekolah dalam satu tempat.
        </h2>
        <p class="mt-5 text-lg leading-8 text-muted">
          Wiyata tidak mencoba menjadi segalanya. Platform ini fokus pada alur
          kerja akademik yang paling sering dijalankan sekolah setiap hari.
        </p>
      </div>

      <div class="mt-20 divide-y divide-[#e7e2da]">
        <div
          v-for="(feature, i) in features"
          :key="feature.title"
          class="grid grid-cols-1 gap-8 py-10 transition-colors hover:bg-[#faf9f7] hover:backdrop-blur-3xl md:grid-cols-[280px_1fr] md:gap-16 lg:gap-24"
        >
          <div class="flex flex-col justify-start pt-0.5">
            <span
              class="mb-2 text-xs font-semibold uppercase tracking-widest text-brand"
            >
              {{ String(i + 1).padStart(2, "0") }}
            </span>
            <h3 class="text-xl font-semibold text-foreground">
              {{ feature.title }}
            </h3>
          </div>
          <div class="flex flex-col justify-center">
            <p class="text-base leading-8 text-muted">
              {{ feature.description }}
            </p>
            <ul class="mt-5 space-y-2">
              <li
                v-for="point in feature.points"
                :key="point"
                class="flex items-start gap-2.5 text-sm text-muted"
              >
                <PhCheckCircle
                  :size="16"
                  class="mt-0.5 shrink-0 text-brand"
                />
                {{ point }}
              </li>
            </ul>
          </div>
        </div>
      </div>
    </section>

    <!-- ───────────── ROLE SECTION ───────────── -->
    <section id="peran" class="relative bg-white scroll-mt-24">
      <div
        class="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_20%_10%,rgba(79,70,229,0.07),transparent_32%),radial-gradient(circle_at_80%_30%,rgba(245,158,11,0.06),transparent_28%)]"
        aria-hidden="true"
      />
      <div class="relative mx-auto max-w-7xl px-6 py-32 lg:px-8">
        <div class="max-w-2xl">
          <p class="text-sm font-medium text-brand">Untuk setiap peran</p>
          <h2
            class="mt-4 text-4xl font-semibold tracking-tight text-foreground sm:text-5xl"
          >
            Satu workspace. Perspektif berbeda untuk tiap pengguna.
          </h2>
          <p class="mt-5 text-lg leading-8 text-muted">
            Wiyata mengenali bahwa siswa, guru, dan admin punya ritme kerja yang
            berbeda. Setiap peran mendapat tampilan dan akses yang relevan.
          </p>
        </div>

        <!-- Individual role cards -->
        <div class="mt-16 grid gap-5 sm:grid-cols-3">
          <article
            v-for="role in mainRoles"
            :id="role.anchor"
            :key="role.title"
            class="rounded-2xl border border-[#e7e2da] bg-white px-8 py-9 transition hover:border-[#c7c3d7] hover:shadow-md"
          >
            <p
              class="text-xs font-semibold uppercase tracking-widest text-brand"
            >
              {{ role.eyebrow }}
            </p>
            <h3 class="mt-4 text-xl font-semibold text-foreground">
              {{ role.title }}
            </h3>
            <p class="mt-3 text-sm leading-7 text-muted">
              {{ role.description }}
            </p>
            <ul class="mt-6 space-y-2">
              <li
                v-for="point in role.points"
                :key="point"
                class="flex items-start gap-2 text-sm text-muted"
              >
                <PhDot :size="16" class="mt-1 shrink-0 text-brand" />
                {{ point }}
              </li>
            </ul>
          </article>
        </div>
      </div>
    </section>

    <!-- ───────────── PREVIEW / MEDIA ───────────── -->
    <section id="preview" class="relative bg-[#fbfaf8] scroll-mt-24">
      <div
        class="pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,rgba(251,250,248,1)_0%,rgba(245,243,255,0.80)_45%,rgba(251,250,248,1)_100%)]"
        aria-hidden="true"
      />
      <div class="relative mx-auto max-w-7xl px-6 py-32 lg:px-8">
        <div class="mx-auto max-w-2xl text-center">
          <p class="text-sm font-medium text-brand">Demo produk</p>
          <h2
            class="mt-4 text-4xl font-semibold tracking-tight text-foreground sm:text-5xl"
          >
            Preview fitur Wiyata sedang disiapkan.
          </h2>
          <p class="mt-5 text-lg leading-8 text-muted">
            Video showcase dan screenshot fitur utama Wiyata akan segera
            ditambahkan.
          </p>
        </div>

        <!-- Video placeholder -->
        <div class="mx-auto mt-14 max-w-4xl">
          <div
            class="relative flex aspect-video w-full flex-col items-center justify-center overflow-hidden rounded-2xl border border-white/10 bg-[#1a1830] shadow-[0_30px_90px_-40px_rgba(26,24,48,0.75)]"
          >
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
                  Video showcase Wiyata akan segera diunggah
                </p>
                <p class="mt-1 text-xs text-white/40">
                  Walkthrough singkat produk Wiyata sedang disiapkan
                </p>
              </div>
            </div>
            <div
              class="absolute bottom-4 right-4 rounded-md bg-white/10 px-2.5 py-1 text-xs text-white/50"
            >
              ~2 menit
            </div>
          </div>

          <p class="mt-4 text-center text-xs text-[#9ca3af]">
            Preview video produk sedang disiapkan dan akan ditampilkan di bagian
            ini.
          </p>
        </div>

        <!-- Screenshot placeholders -->
        <div class="mt-10 grid gap-4 sm:grid-cols-3">
          <div
            v-for="slot in screenshotSlots"
            :key="slot.label"
            class="flex aspect-4/3 flex-col items-center justify-center gap-3 rounded-xl border border-[#e7e2da] bg-white text-center shadow-sm"
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
    <section class="relative overflow-hidden bg-[#fbfaf8]">
      <!-- Backdrop blobs — exist in section behind the card so backdrop-blur has something to blur -->
      <div
        class="pointer-events-none absolute right-0 top-2/7 h-40 w-lg -translate-y-1/2 translate-x-1/3 rounded-full bg-brand/22 blur-3xl"
        aria-hidden="true"
      />
      <div
        class="pointer-events-none absolute left-0 top-2/3 h-70 w-80 -translate-x-1/3 -translate-y-1/2 rounded-full bg-[#f59e0b]/20 blur-3xl"
        aria-hidden="true"
      />

      <div class="mx-auto max-w-7xl px-6 py-20 lg:px-8">
        <!-- Glass card — backdrop-blur now blurs the section blobs behind it -->
        <div
          class="relative rounded-3xl bg-white/40 px-8 py-12 shadow-[0_2px_24px_-4px_rgba(79,70,229,0.09),inset_0_1px_0_rgba(255,255,255,0.85)] backdrop-blur-3xl backdrop-saturate-250 lg:px-14 lg:py-14"
        >
          <div class="grid gap-10 lg:grid-cols-[1fr_auto] lg:items-center">
            <div class="max-w-2xl">
              <p class="text-sm font-medium text-brand">
                Wiyata Academic Workspace
              </p>
              <h2
                class="mt-4 text-4xl font-semibold tracking-tight text-foreground sm:text-5xl"
              >
                Workspace sekolah yang sudah siap dipakai.
              </h2>
              <p class="mt-5 text-lg leading-8 text-muted">
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
                to="/school-registration"
                id="final-cta-daftar-sekolah"
                class="inline-flex h-12 items-center justify-center gap-2 rounded-lg bg-brand px-8 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-hover"
              >
                Daftarkan Sekolah
                <PhArrowRight :size="15" />
              </RouterLink>
              <RouterLink
                to="/login"
                id="final-cta-masuk"
                class="inline-flex h-11 items-center justify-center gap-2 rounded-lg border border-[#e0daf7] bg-white/70 px-7 text-sm font-medium text-muted transition-colors hover:border-[#c7c3d7] hover:text-foreground"
              >
                Masuk ke Wiyata
                <PhArrowRight :size="15" />
              </RouterLink>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ───────────── FOOTER ───────────── -->
    <footer class="border-t border-[#e7e2da] bg-[#fbfaf8]">
      <div class="mx-auto max-w-7xl px-6 py-10 lg:px-8">
        <div
          class="flex flex-col gap-6 sm:flex-row sm:items-center sm:justify-between"
        >
          <div class="flex items-center gap-2.5">
            <img
              src="/logo_fix.svg"
              alt="Wiyata"
              class="h-6 w-6 rounded-md object-contain"
            />
            <span class="text-sm font-semibold text-foreground">Wiyata</span>
          </div>

          <nav class="flex flex-wrap gap-x-7 gap-y-2 text-sm text-[#9ca3af]">
            <a
              href="#fitur"
              class="hover:text-muted"
              @click="handleAnchorClick($event, '#fitur')"
              >Fitur</a
            >
            <a
              href="#peran"
              class="hover:text-muted"
              @click="handleAnchorClick($event, '#peran')"
              >Untuk Siapa</a
            >
            <a
              href="#preview"
              class="hover:text-muted"
              @click="handleAnchorClick($event, '#preview')"
              >Preview</a
            >
            <RouterLink to="/login" class="hover:text-muted"
              >Masuk</RouterLink
            >
            <RouterLink to="/school-registration" class="hover:text-muted">
              Daftarkan Sekolah
            </RouterLink>
          </nav>
        </div>

        <div
          class="mt-8 flex flex-col gap-1 border-t border-[#e7e2da] pt-6 text-xs text-[#c4bfcc] sm:flex-row sm:justify-between"
        >
          <p>© 2026 Wiyata Academic Workspace. All rights reserved.</p>
          <p>Learning Management System by Loka Wiyata</p>
        </div>
      </div>
    </footer>
  </main>
</template>
