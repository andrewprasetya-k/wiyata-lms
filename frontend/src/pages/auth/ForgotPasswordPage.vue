<script setup lang="ts">
import { ref } from "vue";
import { RouterLink } from "vue-router";
import { PhArrowRight } from "@phosphor-icons/vue";
import { requestPasswordReset } from "../../services/passwordReset";
import { getApiError } from "../../utils/error";

const email = ref("");
const isSubmitting = ref(false);
const errorMessage = ref("");
const submitted = ref(false);

const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

async function submit() {
  if (isSubmitting.value) return;
  errorMessage.value = "";

  if (!emailPattern.test(email.value.trim())) {
    errorMessage.value = "Masukkan alamat email yang valid.";
    return;
  }

  isSubmitting.value = true;
  try {
    // The backend always returns the same generic message regardless of
    // whether this email is actually registered — this page just displays
    // whatever it gets back, it never infers existence itself either.
    await requestPasswordReset(email.value.trim());
    submitted.value = true;
  } catch (error) {
    errorMessage.value = getApiError(error);
  } finally {
    isSubmitting.value = false;
  }
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
          <p class="text-xs text-muted">Academic workspace</p>
        </div>
      </div>

      <div class="max-w-xl">
        <h1
          class="mt-4 text-4xl font-medium leading-tight text-foreground lg:text-6xl"
        >
          Lupa kata sandi? Tidak masalah.
        </h1>
        <p class="mt-6 text-base leading-7 text-muted">
          Masukkan email akun Wiyata-mu, kami kirimkan tautan untuk membuat
          password baru.
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
        <div class="mb-8">
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

          <h2 class="text-3xl font-medium text-foreground">
            Lupa kata sandi
          </h2>
          <p class="mt-3 text-sm text-muted">
            Masukkan email akun Wiyata-mu untuk menerima tautan reset
            password.
          </p>
        </div>

        <div v-if="submitted" class="space-y-5">
          <div class="rounded-2xl bg-[#f5fbf2] p-5">
            <p class="text-base font-medium text-[#1f3d25]">
              Periksa email kamu.
            </p>
            <p class="mt-2 text-sm leading-6 text-[#48614b]">
              Jika email tersebut terdaftar, tautan reset password sudah
              dikirim. Tautan berlaku selama 20 menit.
            </p>
          </div>
          <RouterLink
            to="/login"
            class="inline-flex h-11 items-center justify-center rounded-2xl border border-border px-5 text-sm font-medium text-foreground-secondary transition hover:text-foreground"
          >
            Kembali ke Login
          </RouterLink>
        </div>

        <form v-else class="space-y-5" @submit.prevent="submit">
          <label class="block">
            <span
              class="mb-2 block text-sm font-medium text-foreground-secondary"
            >
              Email
            </span>
            <input
              v-model="email"
              class="h-12 w-full rounded-2xl border border-border bg-surface-subtle px-4 text-sm outline-none transition focus:border-brand focus:bg-surface"
              type="email"
              autocomplete="email"
              placeholder="nama@sekolah.sch.id"
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
            :disabled="isSubmitting"
          >
            {{ isSubmitting ? "Mengirim..." : "Kirim tautan reset" }}
            <PhArrowRight v-if="!isSubmitting" :size="18" />
          </button>
        </form>

        <p class="mt-6 text-center text-sm text-muted">
          Sudah ingat password?
          <RouterLink
            to="/login"
            class="font-medium text-brand hover:text-brand-hover"
          >
            Masuk
          </RouterLink>
        </p>
      </div>
    </section>
  </main>
</template>
