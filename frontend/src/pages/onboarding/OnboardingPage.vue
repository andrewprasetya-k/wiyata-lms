<script setup lang="ts">
import { RouterLink } from "vue-router";
import { PhArrowRight, PhBuildings } from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import { useLogoutConfirm } from "../../composables/useLogoutConfirm";

const auth = useAuthStore();
const { confirmLogout } = useLogoutConfirm({ redirectTo: "/login" });
</script>

<template>
  <main class="min-h-screen bg-surface-subtle px-6 py-10 sm:py-16">
    <div class="mx-auto flex w-full max-w-4xl items-center justify-between">
      <div class="flex items-center gap-3">
        <img src="/logo_fix.svg" alt="Wiyata" class="h-9 w-9 rounded-lg" />
        <span class="text-sm font-semibold text-foreground">
          Wiyata Academic Workspace
        </span>
      </div>
      <button
        type="button"
        class="rounded-lg border border-border bg-surface px-4 py-2 text-sm font-medium text-foreground-secondary transition hover:text-foreground"
        @click="confirmLogout"
      >
        Keluar
      </button>
    </div>

    <section class="mx-auto mt-16 max-w-xl">
      <p class="text-sm font-medium text-brand">
        Halo, {{ auth.user?.fullName?.split(" ")[0] ?? "" }}
      </p>
      <h1
        class="mt-3 text-4xl font-semibold leading-tight tracking-tight text-foreground sm:text-5xl"
      >
        Selamat datang di Wiyata 👋
      </h1>
      <p class="mt-4 text-base leading-7 text-muted">
        Akun Anda berhasil dibuat. Langkah berikutnya adalah mendaftarkan
        sekolah agar dapat mulai menggunakan Wiyata.
      </p>

      <RouterLink
        to="/school-registration"
        class="group mt-8 flex flex-col rounded-2xl border border-border bg-surface p-8 shadow-sm transition hover:border-brand hover:shadow-md"
      >
        <div
          class="flex h-12 w-12 items-center justify-center rounded-xl bg-brand-soft text-brand"
        >
          <PhBuildings :size="24" weight="duotone" />
        </div>
        <h2 class="mt-6 text-xl font-semibold text-foreground">
          Buat Sekolah Baru
        </h2>
        <p class="mt-2 text-sm leading-6 text-muted">
          Daftarkan sekolah Anda untuk mulai menggunakan Wiyata.
        </p>
        <span
          class="mt-6 inline-flex items-center gap-1.5 text-sm font-medium text-brand"
        >
          Daftarkan Sekolah
          <PhArrowRight
            :size="16"
            class="transition group-hover:translate-x-0.5"
          />
        </span>
      </RouterLink>

      <p class="mt-8 text-sm leading-6 text-muted">
        Sudah mendapat undangan dari sekolah? Gunakan tautan pada email
        undangan yang dikirim administrator sekolah.
      </p>
    </section>
  </main>
</template>
