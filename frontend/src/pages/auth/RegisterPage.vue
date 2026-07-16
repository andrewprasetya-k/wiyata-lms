<script setup lang="ts">
import { computed, ref } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import { PhArrowRight, PhEye, PhEyeSlash } from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import { usePasswordVisibility } from "../../composables/usePasswordVisibility";

const auth = useAuthStore();
const route = useRoute();
const router = useRouter();

const fullName = ref("");
const email = ref("");
const password = ref("");
const confirmPassword = ref("");
const isSubmitting = ref(false);
const errorMessage = ref("");

const {
  visible: passwordVisible,
  inputType: passwordInputType,
  toggle: togglePasswordVisibility,
} = usePasswordVisibility();
const {
  visible: confirmPasswordVisible,
  inputType: confirmPasswordInputType,
  toggle: toggleConfirmPasswordVisibility,
} = usePasswordVisibility();

const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

const validationError = computed(() => {
  if (fullName.value.trim() === "") return "Nama lengkap wajib diisi.";
  if (!emailPattern.test(email.value.trim())) return "Email belum valid.";
  if (password.value.length < 6) return "Password minimal 6 karakter.";
  if (password.value !== confirmPassword.value)
    return "Konfirmasi password belum sama.";
  return "";
});

const canSubmit = computed(
  () =>
    fullName.value.trim() !== "" &&
    email.value.trim() !== "" &&
    password.value !== "" &&
    confirmPassword.value !== "",
);

function errorFromResponse(error: unknown) {
  const maybeError = error as { response?: { data?: { error?: string } } };
  const message = maybeError.response?.data?.error;
  if (!message) return "Registrasi belum bisa diproses. Coba lagi sebentar lagi.";
  if (message.toLowerCase().includes("already registered")) {
    return "Email ini sudah terdaftar. Silakan masuk dengan akun tersebut.";
  }
  return message;
}

async function submit() {
  if (isSubmitting.value) return;

  if (validationError.value) {
    errorMessage.value = validationError.value;
    return;
  }

  isSubmitting.value = true;
  errorMessage.value = "";

  try {
    await auth.register({
      fullName: fullName.value.trim(),
      email: email.value.trim(),
      password: password.value,
    });
    await router.push(
      (route.query.redirect as string | undefined) ?? auth.landingRoute(),
    );
  } catch (error) {
    errorMessage.value = errorFromResponse(error);
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
          Satu akun untuk seluruh ruang kerja sekolahmu.
        </h1>
        <p class="mt-6 text-base leading-7 text-muted">
          Buat akun sekali, lalu daftarkan sekolah baru atau bergabung memakai
          undangan yang diberikan administrator sekolahmu.
        </p>
      </div>

      <div class="text-xs text-muted">
        &copy; 2026 Wiyata. All rights reserved.
      </div>
    </section>

    <!-- Right Side: Register Form -->
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

          <h2 class="text-3xl font-medium text-foreground">Buat akun</h2>
          <p class="mt-3 text-sm text-muted">
            Mulai dengan akun Wiyata-mu sendiri.
          </p>
        </div>

        <form class="space-y-5" @submit.prevent="submit">
          <label class="block">
            <span class="mb-2 block text-sm font-medium text-foreground-secondary">
              Nama lengkap
            </span>
            <input
              v-model="fullName"
              class="h-12 w-full rounded-2xl border border-border bg-surface-subtle px-4 text-sm outline-none transition focus:border-brand focus:bg-surface"
              type="text"
              autocomplete="name"
              placeholder="Budi Santoso"
            />
          </label>

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
                autocomplete="new-password"
                placeholder="Minimal 6 karakter"
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

          <label class="block">
            <span class="mb-2 block text-sm font-medium text-foreground-secondary">
              Konfirmasi password
            </span>
            <div class="relative">
              <input
                v-model="confirmPassword"
                class="h-12 w-full rounded-2xl border border-border bg-surface-subtle px-4 pr-12 text-sm outline-none transition focus:border-brand focus:bg-surface"
                :type="confirmPasswordInputType"
                autocomplete="new-password"
                placeholder="Ulangi password"
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
            :disabled="!canSubmit || isSubmitting"
          >
            {{ isSubmitting ? "Memproses..." : "Buat akun" }}
            <PhArrowRight v-if="!isSubmitting" :size="18" />
          </button>
        </form>

        <p class="mt-6 text-center text-sm text-muted">
          Sudah punya akun?
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
