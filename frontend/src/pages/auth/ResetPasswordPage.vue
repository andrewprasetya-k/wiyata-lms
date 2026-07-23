<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink, useRoute } from "vue-router";
import { PhArrowRight, PhEye, PhEyeSlash } from "@phosphor-icons/vue";
import {
  getPasswordResetMetadata,
  resetPassword,
} from "../../services/passwordReset";
import { usePasswordVisibility } from "../../composables/usePasswordVisibility";
import { passwordPolicyError } from "../../utils/passwordPolicy";
import { getApiError } from "../../utils/error";

const route = useRoute();
const token = computed(() => String(route.params.token ?? ""));

const checkingToken = ref(true);
const tokenValid = ref(false);
const tokenErrorMessage = ref("");

const newPassword = ref("");
const confirmPassword = ref("");
const isSubmitting = ref(false);
const errorMessage = ref("");
const resetDone = ref(false);

const {
  visible: newPasswordVisible,
  inputType: newPasswordInputType,
  toggle: toggleNewPasswordVisibility,
} = usePasswordVisibility();
const {
  visible: confirmPasswordVisible,
  inputType: confirmPasswordInputType,
  toggle: toggleConfirmPasswordVisibility,
} = usePasswordVisibility();

const canSubmit = computed(
  () =>
    newPassword.value !== "" &&
    confirmPassword.value !== "" &&
    !isSubmitting.value,
);

async function checkToken() {
  checkingToken.value = true;
  tokenErrorMessage.value = "";

  if (!token.value) {
    tokenValid.value = false;
    tokenErrorMessage.value = "Link reset password tidak lengkap.";
    checkingToken.value = false;
    return;
  }

  try {
    await getPasswordResetMetadata(token.value);
    tokenValid.value = true;
  } catch (error) {
    tokenValid.value = false;
    tokenErrorMessage.value = getApiError(error);
  } finally {
    checkingToken.value = false;
  }
}

async function submit() {
  if (isSubmitting.value) return;
  errorMessage.value = "";

  if (newPassword.value !== confirmPassword.value) {
    errorMessage.value = "Konfirmasi password belum sama.";
    return;
  }
  const policyError = passwordPolicyError(newPassword.value);
  if (policyError) {
    errorMessage.value = policyError;
    return;
  }

  isSubmitting.value = true;
  try {
    await resetPassword(token.value, newPassword.value);
    resetDone.value = true;
  } catch (error) {
    errorMessage.value = getApiError(error);
  } finally {
    isSubmitting.value = false;
  }
}

onMounted(checkToken);
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
          <p class="text-xs text-muted">Academic workspace</p>
        </div>
      </div>

      <div class="max-w-xl">
        <h1
          class="mt-4 text-4xl font-medium leading-tight text-foreground lg:text-6xl"
        >
          Buat password baru.
        </h1>
        <p class="mt-6 text-base leading-7 text-muted">
          Password baru minimal 8 karakter, kombinasi huruf besar, huruf
          kecil, dan angka.
        </p>
      </div>

      <div class="text-xs text-muted">
        &copy; 2026 Wiyata. All rights reserved.
      </div>
    </section>

    <!-- Right Side: Form -->
    <section
      class="flex h-screen items-center justify-center overflow-y-auto bg-surface px-6 py-8 sm:px-12"
    >
      <div class="w-full max-w-md">
        <div class="mb-8 flex items-center gap-3 md:hidden">
          <div class="flex h-12 w-12 items-center justify-center rounded-2xl">
            <img
              src="/logo_fix.svg"
              alt="Wiyata"
              class="h-12 w-12 rounded-2xl object-contain"
            />
          </div>
          <div>
            <p class="text-sm font-medium text-brand">Wiyata</p>
            <p class="text-xs text-muted">Academic workspace</p>
          </div>
        </div>

        <div v-if="checkingToken" class="space-y-3">
          <div class="h-8 w-2/3 animate-pulse rounded bg-border" />
          <div class="h-4 w-full animate-pulse rounded bg-[#f0ece5]" />
        </div>

        <div v-else-if="!tokenValid" class="space-y-5">
          <div class="rounded-2xl bg-[#fff7f5] p-5">
            <p class="text-lg font-medium text-[#9f2a1d]">
              Link tidak valid.
            </p>
            <p class="mt-2 text-sm leading-6 text-danger">
              {{
                tokenErrorMessage ||
                "Link reset password tidak valid atau sudah kedaluwarsa."
              }}
            </p>
          </div>
          <RouterLink
            to="/forgot-password"
            class="inline-flex h-11 items-center justify-center rounded-2xl bg-brand px-5 text-sm font-medium text-white transition hover:bg-brand-hover"
          >
            Minta tautan baru
          </RouterLink>
        </div>

        <div v-else-if="resetDone" class="space-y-5">
          <div class="rounded-2xl bg-[#f5fbf2] p-5">
            <p class="text-base font-medium text-[#1f3d25]">
              Password berhasil diubah.
            </p>
            <p class="mt-2 text-sm leading-6 text-[#48614b]">
              Silakan masuk menggunakan password baru kamu.
            </p>
          </div>
          <RouterLink
            to="/login"
            class="inline-flex h-11 items-center justify-center rounded-2xl bg-brand px-5 text-sm font-medium text-white transition hover:bg-brand-hover"
          >
            Login ke Wiyata
          </RouterLink>
        </div>

        <template v-else>
          <div class="mb-8">
            <h2 class="text-3xl font-medium text-foreground">
              Buat password baru
            </h2>
            <p class="mt-3 text-sm text-muted">
              Masukkan password baru untuk akun Wiyata-mu.
            </p>
          </div>

          <form class="space-y-5" @submit.prevent="submit">
            <label class="block">
              <span
                class="mb-2 block text-sm font-medium text-foreground-secondary"
              >
                Password baru
              </span>
              <div class="relative">
                <input
                  v-model="newPassword"
                  class="h-12 w-full rounded-2xl border border-border bg-surface-subtle px-4 pr-12 text-sm outline-none transition focus:border-brand focus:bg-surface"
                  :type="newPasswordInputType"
                  autocomplete="new-password"
                  placeholder="Minimal 8 karakter"
                />
                <button
                  type="button"
                  class="absolute right-3 top-1/2 -translate-y-1/2 rounded-lg p-1.5 text-muted transition hover:text-foreground"
                  :aria-label="
                    newPasswordVisible
                      ? 'Sembunyikan password'
                      : 'Tampilkan password'
                  "
                  :aria-pressed="newPasswordVisible"
                  @click="toggleNewPasswordVisibility"
                >
                  <PhEyeSlash v-if="newPasswordVisible" :size="18" />
                  <PhEye v-else :size="18" />
                </button>
              </div>
            </label>

            <label class="block">
              <span
                class="mb-2 block text-sm font-medium text-foreground-secondary"
              >
                Konfirmasi password baru
              </span>
              <div class="relative">
                <input
                  v-model="confirmPassword"
                  class="h-12 w-full rounded-2xl border border-border bg-surface-subtle px-4 pr-12 text-sm outline-none transition focus:border-brand focus:bg-surface"
                  :type="confirmPasswordInputType"
                  autocomplete="new-password"
                  placeholder="Ulangi password baru"
                />
                <button
                  type="button"
                  class="absolute right-3 top-1/2 -translate-y-1/2 rounded-lg p-1.5 text-muted transition hover:text-foreground"
                  :aria-label="
                    confirmPasswordVisible
                      ? 'Sembunyikan password'
                      : 'Tampilkan password'
                  "
                  :aria-pressed="confirmPasswordVisible"
                  @click="toggleConfirmPasswordVisibility"
                >
                  <PhEyeSlash v-if="confirmPasswordVisible" :size="18" />
                  <PhEye v-else :size="18" />
                </button>
              </div>
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
              {{ isSubmitting ? "Menyimpan..." : "Simpan password baru" }}
              <PhArrowRight v-if="!isSubmitting" :size="18" />
            </button>
          </form>
        </template>
      </div>
    </section>
  </main>
</template>
