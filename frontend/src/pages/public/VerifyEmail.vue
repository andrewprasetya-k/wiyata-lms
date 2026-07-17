<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink, useRoute } from "vue-router";
import { verifyEmail } from "../../services/emailVerification";
import { getApiError } from "../../utils/error";

const route = useRoute();
const token = computed(() => String(route.query.token ?? ""));

const loading = ref(true);
const verifiedAt = ref("");
const errorMessage = ref("");

function formatDateTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat("id-ID", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

async function runVerification() {
  loading.value = true;
  errorMessage.value = "";

  if (!token.value) {
    errorMessage.value = "Link verifikasi tidak lengkap.";
    loading.value = false;
    return;
  }

  try {
    const response = await verifyEmail(token.value);
    verifiedAt.value = response.emailVerifiedAt;
  } catch (error) {
    errorMessage.value = getApiError(error);
  } finally {
    loading.value = false;
  }
}

onMounted(runVerification);
</script>

<template>
  <main class="min-h-screen bg-surface-subtle px-6 py-8 text-foreground">
    <div class="mx-auto flex w-full max-w-screen items-center justify-between">
      <RouterLink to="/home" class="flex items-center gap-3">
        <img src="/logo_fix.svg" alt="Wiyata" class="h-9 w-9 rounded-lg" />
        <span class="text-sm font-semibold">Wiyata Academic Workspace</span>
      </RouterLink>
      <RouterLink
        to="/login"
        class="rounded-lg border border-border bg-surface px-4 py-2 text-sm font-medium text-foreground-secondary transition hover:text-foreground"
      >
        Masuk
      </RouterLink>
    </div>

    <section class="mx-auto mt-12 max-w-screen">
      <div
        class="rounded-xl border border-border bg-surface p-6 shadow-sm md:p-8"
      >
        <div v-if="loading" class="space-y-5">
          <div class="h-5 w-36 animate-pulse rounded bg-border" />
          <div class="h-9 w-2/3 animate-pulse rounded bg-border" />
          <div class="space-y-3">
            <div class="h-4 w-full animate-pulse rounded bg-[#f0ece5]" />
            <div class="h-4 w-4/5 animate-pulse rounded bg-[#f0ece5]" />
          </div>
        </div>

        <div v-else-if="verifiedAt" class="space-y-5">
          <div class="rounded-xl bg-[#f5fbf2] p-5">
            <p class="text-lg font-semibold text-[#1f3d25]">
              Email berhasil diverifikasi.
            </p>
            <p class="mt-2 text-sm leading-6 text-[#48614b]">
              Email kamu terverifikasi pada {{ formatDateTime(verifiedAt) }}.
              Silakan login untuk melanjutkan.
            </p>
          </div>
          <RouterLink
            to="/login"
            class="inline-flex h-10 items-center justify-center rounded-lg bg-brand px-5 text-sm font-medium text-white transition hover:bg-brand-hover"
          >
            Login ke Wiyata
          </RouterLink>
        </div>

        <div v-else class="space-y-5">
          <div class="rounded-xl bg-[#fff7f5] p-5">
            <p class="text-lg font-semibold text-[#9f2a1d]">
              Verifikasi belum berhasil.
            </p>
            <p class="mt-2 text-sm leading-6 text-danger">
              {{
                errorMessage ||
                "Link verifikasi tidak valid atau sudah kedaluwarsa."
              }}
            </p>
          </div>
          <RouterLink
            to="/login"
            class="inline-flex h-10 items-center justify-center rounded-lg border border-border px-4 text-sm font-medium text-foreground-secondary transition hover:text-foreground"
          >
            Kembali ke Login
          </RouterLink>
        </div>
      </div>
    </section>
  </main>
</template>
