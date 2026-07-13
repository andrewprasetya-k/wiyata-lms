<script setup lang="ts">
import { computed } from "vue";
import { useRouter } from "vue-router";
import { PhCompassTool } from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";

const router = useRouter();
const auth = useAuthStore();

const primaryTarget = computed(() => {
  if (!auth.isAuthenticated) {
    return "/home";
  }

  return auth.landingRoute();
});

const primaryLabel = computed(() =>
  auth.isAuthenticated ? "Kembali ke Dashboard" : "Kembali ke Beranda",
);

function goBack() {
  if (window.history.length > 1) {
    router.back();
    return;
  }

  router.push(primaryTarget.value);
}
</script>

<template>
  <section class="soft-card w-full max-w-screen rounded-[28px] p-8 text-center">
    <div
      class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-brand-soft text-brand"
    >
      <PhCompassTool :size="22" weight="duotone" />
    </div>
    <p class="mt-4 text-sm font-medium text-[#f2756a]">Error 404</p>
    <h1 class="mt-3 text-3xl font-medium text-foreground">
      Halaman tidak ditemukan
    </h1>
    <p class="mt-4 text-sm leading-6 text-[#6b6475]">
      Halaman yang kamu cari tidak tersedia atau mungkin sudah dipindahkan.
    </p>
    <div class="mt-7 flex flex-col items-center justify-center gap-3">
      <RouterLink
        class="inline-flex h-11 items-center justify-center rounded-xl bg-brand px-5 text-sm font-medium text-white"
        :to="primaryTarget"
      >
        {{ primaryLabel }}
      </RouterLink>
      <button
        type="button"
        class="inline-flex h-11 items-center justify-center rounded-xl border border-[#ddd6cb] bg-surface px-5 text-sm font-medium text-foreground transition hover:bg-background"
        @click="goBack"
      >
        Kembali ke halaman sebelumnya
      </button>
    </div>
  </section>
</template>
