<script setup lang="ts">
import { computed, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { PhArrowRight, PhEye, PhEyeSlash } from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import { usePasswordVisibility } from "../../composables/usePasswordVisibility";

const auth = useAuthStore();
const route = useRoute();
const router = useRouter();

const email = ref("");
const password = ref("");
const isSubmitting = ref(false);
const errorMessage = ref("");
const {
  visible: passwordVisible,
  inputType: passwordInputType,
  toggle: togglePasswordVisibility,
} = usePasswordVisibility();

const canSubmit = computed(
  () => email.value.trim() !== "" && password.value.trim() !== "",
);

async function submit() {
  if (!canSubmit.value || isSubmitting.value) return;
  isSubmitting.value = true;
  errorMessage.value = "";

  try {
    await auth.login({ email: email.value, password: password.value });
    await router.push((route.query.redirect as string | undefined) ?? auth.landingRoute());
  } catch {
    errorMessage.value = "Email atau password tidak valid.";
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
        <p class="text-sm font-medium text-muted">
          Learning Management System
        </p>
        <h1
          class="mt-4 text-4xl font-medium leading-tight text-foreground lg:text-6xl"
        >
          Masuk ke ruang belajar yang lebih tenang.
        </h1>
        <p class="mt-6 text-base leading-7 text-muted">
          Satu login untuk siswa, guru, admin sekolah, dan super admin. Wiyata
          akan memilih ruang kerja berdasarkan role dan konteks sekolah.
        </p>
      </div>

      <div class="text-xs text-muted">
        &copy; 2026 Wiyata. All rights reserved.
      </div>
    </section>

    <!-- Right Side: Login Form -->
    <section
      class="flex h-screen items-center justify-center bg-surface px-6 py-8 sm:px-12"
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

          <h2 class="text-3xl font-medium text-foreground">Login</h2>
          <p class="mt-3 text-sm text-muted">
            Gunakan akun Wiyata yang sudah terdaftar.
          </p>
        </div>

        <form class="space-y-5" @submit.prevent="submit">
          <label class="block">
            <span class="mb-2 block text-sm font-medium text-foreground-secondary">
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

          <label class="block">
            <span class="mb-2 block text-sm font-medium text-foreground-secondary">
              Password
            </span>
            <div class="relative">
              <input
                v-model="password"
                class="h-12 w-full rounded-2xl border border-border bg-surface-subtle px-4 pr-12 text-sm outline-none transition focus:border-brand focus:bg-surface"
                :type="passwordInputType"
                autocomplete="current-password"
                placeholder="••••••••"
              />
              <button
                type="button"
                class="absolute right-3 top-1/2 -translate-y-1/2 rounded-lg p-1.5 text-muted transition hover:text-foreground"
                :aria-label="
                  passwordVisible ? 'Sembunyikan password' : 'Tampilkan password'
                "
                :aria-pressed="passwordVisible"
                @click="togglePasswordVisibility"
              >
                <PhEyeSlash v-if="passwordVisible" :size="18" />
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
            :disabled="!canSubmit || isSubmitting"
          >
            {{ isSubmitting ? "Memproses..." : "Masuk" }}
            <PhArrowRight v-if="!isSubmitting" :size="18" />
          </button>
        </form>
      </div>
    </section>
  </main>
</template>
