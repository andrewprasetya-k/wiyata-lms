<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { RouterLink, useRoute } from "vue-router";
import {
  acceptInvitation,
  getInvitation,
  type AcceptInvitationResponse,
  type InvitationMetadata,
} from "../../services/onboarding";

const route = useRoute();
const token = computed(() => String(route.params.token ?? ""));

const invitation = ref<InvitationMetadata | null>(null);
const accepted = ref<AcceptInvitationResponse | null>(null);
const loading = ref(true);
const submitting = ref(false);
const errorMessage = ref("");

const form = reactive({
  name: "",
  password: "",
  confirmPassword: "",
});

const canSubmit = computed(
  () =>
    form.name.trim() !== "" &&
    form.password.length >= 6 &&
    form.password === form.confirmPassword,
);

function formatDateTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat("id-ID", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

function errorFromResponse(error: unknown) {
  const maybeError = error as { response?: { data?: { error?: string } } };
  return maybeError.response?.data?.error ?? "Undangan belum bisa diproses.";
}

async function loadInvitation() {
  loading.value = true;
  errorMessage.value = "";
  try {
    invitation.value = await getInvitation(token.value);
  } catch (error) {
    errorMessage.value =
      errorFromResponse(error) ||
      "Undangan tidak valid, sudah kedaluwarsa, atau sudah digunakan.";
  } finally {
    loading.value = false;
  }
}

async function submit() {
  if (!canSubmit.value || submitting.value) {
    errorMessage.value =
      form.password !== form.confirmPassword
        ? "Konfirmasi password belum sama."
        : "Lengkapi nama dan password minimal 6 karakter.";
    return;
  }

  submitting.value = true;
  errorMessage.value = "";
  try {
    accepted.value = await acceptInvitation(token.value, {
      name: form.name.trim(),
      password: form.password,
      confirmPassword: form.confirmPassword,
    });
  } catch (error) {
    errorMessage.value = errorFromResponse(error);
  } finally {
    submitting.value = false;
  }
}

onMounted(loadInvitation);
</script>

<template>
  <main class="min-h-screen bg-[#fbfaf8] px-6 py-8 text-foreground">
    <div class="mx-auto flex w-full max-w-4xl items-center justify-between">
      <RouterLink to="/home" class="flex items-center gap-3">
        <img src="/logo_fix.svg" alt="Wiyata" class="h-9 w-9 rounded-lg" />
        <span class="text-sm font-semibold">Wiyata Academic Workspace</span>
      </RouterLink>
      <RouterLink
        to="/login"
        class="rounded-lg border border-border bg-white px-4 py-2 text-sm font-medium text-[#5f5968] transition hover:text-foreground"
      >
        Masuk
      </RouterLink>
    </div>

    <section class="mx-auto mt-12 max-w-4xl">
      <div
        class="rounded-xl border border-border bg-white p-6 shadow-sm md:p-8"
      >
        <div v-if="loading" class="space-y-5">
          <div class="h-5 w-36 animate-pulse rounded bg-border" />
          <div class="h-9 w-2/3 animate-pulse rounded bg-border" />
          <div class="space-y-3">
            <div class="h-4 w-full animate-pulse rounded bg-[#f0ece5]" />
            <div class="h-4 w-4/5 animate-pulse rounded bg-[#f0ece5]" />
          </div>
        </div>

        <div v-else-if="accepted" class="space-y-5">
          <div class="rounded-xl border border-[#dbe7d5] bg-[#f5fbf2] p-5">
            <p class="text-lg font-semibold text-[#1f3d25]">
              Undangan berhasil diterima.
            </p>
            <p class="mt-2 text-sm leading-6 text-[#48614b]">
              Akun {{ accepted.user.email }} sekarang terhubung dengan
              {{ accepted.school.schoolName }} sebagai {{ accepted.role }}.
              Silakan login memakai email undangan dan password yang baru
              dibuat.
            </p>
          </div>
          <RouterLink
            to="/login"
            class="inline-flex h-10 items-center justify-center rounded-lg bg-brand px-5 text-sm font-medium text-white transition hover:bg-brand-hover"
          >
            Login ke Wiyata
          </RouterLink>
        </div>

        <div v-else-if="errorMessage && !invitation" class="space-y-5">
          <div class="rounded-xl bg-[#fff7f5] p-5">
            <p class="text-lg font-semibold text-[#9f2a1d]">
              Undangan tidak bisa dibuka.
            </p>
            <p class="mt-2 text-sm leading-6 text-[#b42318]">
              {{ errorMessage }}
            </p>
          </div>
          <RouterLink
            to="/home"
            class="inline-flex h-10 items-center justify-center rounded-lg border border-border px-4 text-sm font-medium text-[#5f5968] transition hover:text-foreground"
          >
            Kembali ke beranda
          </RouterLink>
        </div>

        <div
          v-else-if="invitation"
          class="grid gap-8 lg:grid-cols-[0.9fr_1.1fr]"
        >
          <div>
            <p class="text-sm font-medium text-brand">Undangan sekolah</p>
            <h1 class="mt-3 text-3xl font-semibold leading-tight">
              Selesaikan akun admin sekolah.
            </h1>
            <p class="mt-4 text-sm leading-6 text-[#6b6475]">
              Undangan ini terhubung dengan sekolah berikut. Buat password untuk
              mulai memakai Wiyata setelah login.
            </p>

            <dl
              class="mt-6 space-y-4 rounded-xl border border-border bg-[#fbfaf8] p-5 text-sm"
            >
              <div>
                <dt class="text-[#8a8394]">Sekolah</dt>
                <dd class="mt-1 font-medium text-foreground">
                  {{ invitation.school.schoolName }}
                </dd>
              </div>
              <div>
                <dt class="text-[#8a8394]">Email undangan</dt>
                <dd class="mt-1 font-medium text-foreground">
                  {{ invitation.email }}
                </dd>
              </div>
              <div>
                <dt class="text-[#8a8394]">Role</dt>
                <dd class="mt-1 font-medium capitalize text-foreground">
                  {{ invitation.role }}
                </dd>
              </div>
              <div>
                <dt class="text-[#8a8394]">Berlaku sampai</dt>
                <dd class="mt-1 font-medium text-foreground">
                  {{ formatDateTime(invitation.expiresAt) }}
                </dd>
              </div>
            </dl>
          </div>

          <form class="space-y-5" @submit.prevent="submit">
            <label class="block">
              <span class="mb-2 block text-sm font-medium text-[#5f5968]">
                Nama lengkap
              </span>
              <input
                v-model="form.name"
                class="h-11 w-full rounded-lg border border-border bg-[#fbfaf8] px-3 text-sm outline-none transition focus:border-brand focus:bg-white"
                type="text"
                autocomplete="name"
                placeholder="Budi Santoso"
              />
            </label>

            <label class="block">
              <span class="mb-2 block text-sm font-medium text-[#5f5968]">
                Password
              </span>
              <input
                v-model="form.password"
                class="h-11 w-full rounded-lg border border-border bg-[#fbfaf8] px-3 text-sm outline-none transition focus:border-brand focus:bg-white"
                type="password"
                autocomplete="new-password"
                placeholder="Minimal 6 karakter"
              />
            </label>

            <label class="block">
              <span class="mb-2 block text-sm font-medium text-[#5f5968]">
                Konfirmasi password
              </span>
              <input
                v-model="form.confirmPassword"
                class="h-11 w-full rounded-lg border border-border bg-[#fbfaf8] px-3 text-sm outline-none transition focus:border-brand focus:bg-white"
                type="password"
                autocomplete="new-password"
                placeholder="Ulangi password"
              />
            </label>

            <p
              v-if="errorMessage"
              class="rounded-lg border border-[#ffd7d2] bg-[#fff7f5] px-4 py-3 text-sm text-[#b42318]"
            >
              {{ errorMessage }}
            </p>

            <button
              type="submit"
              :disabled="submitting || !canSubmit"
              class="flex h-11 w-full items-center justify-center rounded-lg bg-brand px-5 text-sm font-medium text-white transition hover:bg-brand-hover disabled:cursor-not-allowed disabled:bg-[#bab7d8]"
            >
              {{ submitting ? "Memproses..." : "Terima undangan" }}
            </button>
          </form>
        </div>
      </div>
    </section>
  </main>
</template>
