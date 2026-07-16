<script setup lang="ts">
import { computed, ref } from "vue";
import { RouterLink, useRouter } from "vue-router";
import { PhArrowRight, PhTicket } from "@phosphor-icons/vue";

const router = useRouter();
const codeOrLink = ref("");
const errorMessage = ref("");

const canSubmit = computed(() => codeOrLink.value.trim() !== "");

function extractToken(input: string) {
  const trimmed = input.trim();
  const marker = "/invite/";
  const markerIndex = trimmed.indexOf(marker);
  if (markerIndex === -1) return trimmed;

  const afterMarker = trimmed.slice(markerIndex + marker.length);
  return afterMarker.split(/[/?#]/)[0];
}

function submit() {
  if (!canSubmit.value) {
    errorMessage.value = "Masukkan kode atau tautan undangan terlebih dahulu.";
    return;
  }

  const token = extractToken(codeOrLink.value);
  if (!token) {
    errorMessage.value = "Kode atau tautan undangan belum valid.";
    return;
  }

  errorMessage.value = "";
  router.push({ name: "accept-invitation", params: { token } });
}
</script>

<template>
  <main class="min-h-screen bg-surface-subtle px-6 py-10 sm:py-16">
    <div class="mx-auto flex w-full max-w-4xl items-center gap-3">
      <img src="/logo_fix.svg" alt="Wiyata" class="h-9 w-9 rounded-lg" />
      <span class="text-sm font-semibold text-foreground">
        Wiyata Academic Workspace
      </span>
    </div>

    <section class="mx-auto mt-16 max-w-lg">
      <RouterLink
        to="/onboarding"
        class="text-sm font-medium text-muted transition hover:text-foreground"
      >
        &larr; Kembali
      </RouterLink>

      <div
        class="mt-6 rounded-2xl border border-border bg-surface p-8 shadow-sm"
      >
        <div
          class="flex h-12 w-12 items-center justify-center rounded-xl bg-brand-soft text-brand"
        >
          <PhTicket :size="24" weight="duotone" />
        </div>
        <h1 class="mt-6 text-2xl font-semibold text-foreground">
          Masukkan Kode Undangan
        </h1>
        <p class="mt-2 text-sm leading-6 text-muted">
          Tempel tautan undangan lengkap, atau kode undangan yang diberikan
          administrator sekolah Anda.
        </p>

        <form class="mt-8 space-y-5" @submit.prevent="submit">
          <label class="block">
            <span
              class="mb-2 block text-sm font-medium text-foreground-secondary"
            >
              Kode atau tautan undangan
            </span>
            <input
              v-model="codeOrLink"
              class="h-12 w-full rounded-2xl border border-border bg-surface-subtle px-4 text-sm outline-none transition focus:border-brand focus:bg-surface"
              type="text"
              placeholder="https://app.wiyata.id/invite/... atau kode undangan"
            />
          </label>

          <p
            v-if="errorMessage"
            class="rounded-2xl bg-danger-soft px-4 py-3 text-sm text-danger"
          >
            {{ errorMessage }}
          </p>

          <button
            class="flex h-12 w-full items-center justify-center gap-2 rounded-2xl bg-brand text-sm font-medium text-white transition hover:bg-brand-hover disabled:cursor-not-allowed disabled:bg-[#bab7d8]"
            type="submit"
            :disabled="!canSubmit"
          >
            Lanjutkan
            <PhArrowRight :size="18" />
          </button>
        </form>
      </div>
    </section>
  </main>
</template>
