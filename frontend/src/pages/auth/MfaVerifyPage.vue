<script setup lang="ts">
import { computed, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { PhArrowRight, PhKey } from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import { verifyMfaLogin } from "../../services/mfa";
import { getApiError } from "../../utils/error";

const auth = useAuthStore();
const route = useRoute();
const router = useRouter();

const mode = ref<"code" | "recovery">("code");
const code = ref("");
const recoveryCode = ref("");
const isSubmitting = ref(false);
const errorMessage = ref("");
const sessionExpired = ref(false);

const canSubmit = computed(() =>
  mode.value === "code"
    ? code.value.trim().length === 6
    : recoveryCode.value.trim() !== "",
);

function toggleMode() {
  mode.value = mode.value === "code" ? "recovery" : "code";
  errorMessage.value = "";
}

function mapError(error: unknown): string {
  const message = getApiError(error).toLowerCase();
  if (message.includes("invalid or expired")) {
    sessionExpired.value = true;
    return "Sesi verifikasi sudah kedaluwarsa. Silakan login kembali.";
  }
  if (message.includes("too many failed attempts")) {
    return "Terlalu banyak percobaan gagal. Coba lagi dalam beberapa menit.";
  }
  if (message.includes("invalid verification code")) {
    return mode.value === "recovery"
      ? "Recovery code tidak valid atau sudah pernah dipakai."
      : "Kode tidak valid. Periksa kembali kode dari aplikasi autentikator.";
  }
  return getApiError(error);
}

async function submit() {
  if (!canSubmit.value || isSubmitting.value) return;

  const preAuthToken = auth.pendingMfaToken;
  if (!preAuthToken) {
    sessionExpired.value = true;
    errorMessage.value = "Sesi verifikasi tidak ditemukan. Silakan login kembali.";
    return;
  }

  isSubmitting.value = true;
  errorMessage.value = "";
  try {
    const data = await verifyMfaLogin(
      mode.value === "code"
        ? { preAuthToken, code: code.value.trim() }
        : { preAuthToken, recoveryCode: recoveryCode.value.trim() },
    );
    auth.completeMfaLogin(data);
    const redirect = route.query.redirect as string | undefined;
    await router.push(redirect ?? auth.landingRoute());
  } catch (error) {
    errorMessage.value = mapError(error);
  } finally {
    isSubmitting.value = false;
  }
}

function backToLogin() {
  auth.clearPendingMfa();
  router.push("/login");
}
</script>

<template>
  <main class="fixed inset-0 grid overflow-hidden md:grid-cols-[1fr_1fr]">
    <!-- Left Side: Branding/Intro -->
    <section
      class="hidden flex-col justify-between bg-brand-soft px-8 py-8 sm:px-12 md:flex md:px-16 lg:px-20"
    >
      <div class="flex items-center gap-3">
        <div class="flex h-14 w-14 items-center justify-center rounded-2xl">
          <img
            src="/logo_fix.svg"
            alt="Wiyata"
            class="h-14 w-14 rounded-2xl object-contain"
          />
        </div>
        <div>
          <p class="text-sm font-medium text-brand">Wiyata</p>
          <p class="text-xs text-muted">Academic Workspace</p>
        </div>
      </div>

      <div class="max-w-xl">
        <h1
          class="mt-4 text-4xl font-medium leading-tight text-foreground lg:text-6xl"
        >
          Satu langkah lagi untuk masuk.
        </h1>
        <p class="mt-6 text-base leading-7 text-muted">
          Akun ini dilindungi verifikasi dua langkah. Masukkan kode dari
          aplikasi autentikatormu.
        </p>
      </div>

      <div class="text-xs text-muted">
        &copy; 2026 Wiyata. All rights reserved.
      </div>
    </section>

    <!-- Right Side: Verify Form -->
    <section
      class="flex h-screen items-center justify-center bg-surface px-6 py-8 sm:px-12"
    >
      <div class="w-full max-w-md">
        <div class="mb-8">
          <div
            class="mb-4 flex h-12 w-12 items-center justify-center rounded-2xl bg-brand-soft text-brand"
          >
            <PhKey :size="22" weight="duotone" />
          </div>
          <h2 class="text-3xl font-medium text-foreground">
            Verifikasi dua langkah
          </h2>
          <p class="mt-3 text-sm text-muted">
            {{
              mode === "code"
                ? "Masukkan kode 6 digit dari aplikasi autentikatormu."
                : "Masukkan salah satu recovery code yang sudah kamu simpan."
            }}
          </p>
        </div>

        <form v-if="!sessionExpired" class="space-y-5" @submit.prevent="submit">
          <label v-if="mode === 'code'" class="block">
            <span
              class="mb-2 block text-sm font-medium text-foreground-secondary"
            >
              Kode autentikator
            </span>
            <input
              v-model="code"
              class="h-12 w-full rounded-2xl border border-border bg-surface-subtle px-4 text-center text-lg tracking-[0.5em] outline-none transition focus:border-brand focus:bg-surface"
              inputmode="numeric"
              autocomplete="one-time-code"
              maxlength="6"
              placeholder="000000"
            />
          </label>

          <label v-else class="block">
            <span
              class="mb-2 block text-sm font-medium text-foreground-secondary"
            >
              Recovery code
            </span>
            <input
              v-model="recoveryCode"
              class="h-12 w-full rounded-2xl border border-border bg-surface-subtle px-4 text-center text-sm uppercase tracking-widest outline-none transition focus:border-brand focus:bg-surface"
              autocomplete="off"
              placeholder="XXXXX-XXXXX"
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
            :disabled="!canSubmit || isSubmitting"
          >
            {{ isSubmitting ? "Memverifikasi..." : "Verifikasi" }}
            <PhArrowRight v-if="!isSubmitting" :size="18" />
          </button>

          <button
            type="button"
            class="w-full text-center text-sm font-medium text-brand hover:text-brand-hover"
            @click="toggleMode"
          >
            {{
              mode === "code"
                ? "Pakai recovery code"
                : "Pakai kode autentikator"
            }}
          </button>
        </form>

        <div v-else class="space-y-5">
          <p
            class="rounded-2xl bg-danger-soft px-4 py-3 text-sm text-danger"
          >
            {{ errorMessage }}
          </p>
          <button
            type="button"
            class="flex h-12 w-full items-center justify-center rounded-2xl bg-brand text-sm font-medium text-white transition hover:bg-brand-hover"
            @click="backToLogin"
          >
            Kembali ke Login
          </button>
        </div>

        <p v-if="!sessionExpired" class="mt-6 text-center text-sm text-muted">
          Bukan kamu?
          <button
            type="button"
            class="font-medium text-brand hover:text-brand-hover"
            @click="backToLogin"
          >
            Kembali ke login
          </button>
        </p>
      </div>
    </section>
  </main>
</template>
