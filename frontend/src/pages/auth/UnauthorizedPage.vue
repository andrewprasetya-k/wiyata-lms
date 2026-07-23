<script setup lang="ts">
import router from "../../router";
import { useAuthStore } from "../../stores/auth";

async function logout() {
  const auth = useAuthStore();
  await auth.logout();
  router.push("/login");
}
</script>
<template>
  <section class="soft-card w-full max-w-lg rounded-[28px] p-8 text-center">
    <p class="text-sm font-medium text-[#f2756a]">Akses ditolak</p>
    <h1 class="mt-3 text-3xl font-medium text-foreground">
      Autorisasi tidak dikenali.
    </h1>
    <p class="mt-4 text-sm leading-6 text-muted">
      Pilih konteks sekolah atau role yang sesuai, lalu coba kembali ke halaman
      tujuan.
    </p>
    <RouterLink
      class="mt-7 inline-flex h-11 items-center justify-center rounded-xl bg-brand px-5 text-sm font-medium text-white"
      to="/login"
      @click.prevent="logout"
    >
      Kembali ke login
    </RouterLink>
  </section>
</template>
