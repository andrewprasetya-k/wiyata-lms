<script setup lang="ts">
import { computed, reactive, ref } from "vue";
import { RouterLink } from "vue-router";
import { submitSchoolRegistrationRequest } from "../../services/onboarding";

const form = reactive({
  schoolName: "",
  npsn: "",
  picName: "",
  picEmail: "",
  picPhone: "",
  picRole: "",
  message: "",
});

const loading = ref(false);
const errorMessage = ref("");
const submittedEmail = ref("");
const submittedSchool = ref("");

const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

const canSubmit = computed(
  () =>
    form.schoolName.trim() !== "" &&
    form.picName.trim() !== "" &&
    emailPattern.test(form.picEmail.trim()),
);

function optional(value: string) {
  const trimmed = value.trim();
  return trimmed === "" ? undefined : trimmed;
}

function errorFromResponse(error: unknown) {
  const maybeError = error as { response?: { data?: { error?: string } } };
  const message = maybeError.response?.data?.error;
  if (!message) return "Request belum bisa dikirim. Coba lagi sebentar lagi.";
  if (message.includes("pending registration request")) {
    return "Request pendaftaran untuk sekolah atau email ini masih menunggu review.";
  }
  return message;
}

async function submit() {
  if (!canSubmit.value || loading.value) {
    errorMessage.value =
      "Lengkapi nama sekolah, nama PIC, dan email yang valid.";
    return;
  }

  loading.value = true;
  errorMessage.value = "";

  try {
    await submitSchoolRegistrationRequest({
      schoolName: form.schoolName.trim(),
      npsn: optional(form.npsn),
      picName: form.picName.trim(),
      picEmail: form.picEmail.trim(),
      picPhone: optional(form.picPhone),
      picRole: optional(form.picRole),
      message: optional(form.message),
    });
    submittedEmail.value = form.picEmail.trim();
    submittedSchool.value = form.schoolName.trim();
  } catch (error) {
    errorMessage.value = errorFromResponse(error);
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <main class="min-h-screen bg-surface-subtle px-6 py-8 text-foreground">
    <div class="mx-auto flex w-full max-w-5xl items-center justify-between">
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

    <section
      class="mx-auto mt-12 grid max-w-5xl gap-8 lg:grid-cols-[0.9fr_1.1fr]"
    >
      <div class="pt-4">
        <p class="text-sm font-medium text-brand">Pendaftaran sekolah</p>
        <h1 class="mt-4 text-4xl font-semibold leading-tight lg:text-5xl">
          Daftarkan sekolah untuk memakai Wiyata.
        </h1>
        <p class="mt-5 max-w-xl text-base leading-7 text-muted">
          Kirim data awal sekolah dan kontak PIC. Tim Wiyata akan meninjau
          request sebelum akun admin sekolah dibuat melalui undangan.
        </p>
      </div>

      <div class="rounded-xl border border-border bg-surface p-6 shadow-sm">
        <div v-if="submittedEmail" class="space-y-5">
          <div class="rounded-xl border border-[#dbe7d5] bg-[#f5fbf2] p-5">
            <p class="text-base font-semibold text-[#1f3d25]">
              Request berhasil dikirim.
            </p>
            <p class="mt-2 text-sm leading-6 text-[#48614b]">
              Pendaftaran {{ submittedSchool }} sudah masuk untuk direview admin
              Wiyata. Kontak PIC yang tercatat: {{ submittedEmail }}.
            </p>
          </div>
          <RouterLink
            to="/home"
            class="inline-flex h-10 items-center justify-center rounded-lg border border-border px-4 text-sm font-medium text-foreground-secondary transition hover:text-foreground"
          >
            Kembali ke beranda
          </RouterLink>
        </div>

        <form v-else class="space-y-5" @submit.prevent="submit">
          <div>
            <h2 class="text-xl font-semibold">Data sekolah</h2>
            <p class="mt-1 text-sm text-muted">
              Isi informasi dasar untuk proses review.
            </p>
          </div>

          <label class="block">
            <span class="mb-2 block text-sm font-medium text-foreground-secondary">
              Nama sekolah
            </span>
            <input
              v-model="form.schoolName"
              class="h-11 w-full rounded-lg border border-border bg-surface-subtle px-3 text-sm outline-none transition focus:border-brand focus:bg-surface"
              type="text"
              autocomplete="organization"
              placeholder="SMA Wiyata Mandala"
            />
          </label>

          <label class="block">
            <span class="mb-2 block text-sm font-medium text-foreground-secondary">
              NPSN
              <span class="font-normal text-muted">opsional</span>
            </span>
            <input
              v-model="form.npsn"
              class="h-11 w-full rounded-lg border border-border bg-surface-subtle px-3 text-sm outline-none transition focus:border-brand focus:bg-surface"
              type="text"
              placeholder="12345678"
            />
          </label>

          <div class="grid gap-4 sm:grid-cols-2">
            <label class="block">
              <span class="mb-2 block text-sm font-medium text-foreground-secondary">
                Nama PIC
              </span>
              <input
                v-model="form.picName"
                class="h-11 w-full rounded-lg border border-border bg-surface-subtle px-3 text-sm outline-none transition focus:border-brand focus:bg-surface"
                type="text"
                autocomplete="name"
                placeholder="Budi Santoso"
              />
            </label>

            <label class="block">
              <span class="mb-2 block text-sm font-medium text-foreground-secondary">
                Email PIC
              </span>
              <input
                v-model="form.picEmail"
                class="h-11 w-full rounded-lg border border-border bg-surface-subtle px-3 text-sm outline-none transition focus:border-brand focus:bg-surface"
                type="email"
                autocomplete="email"
                placeholder="budi@sekolah.sch.id"
              />
            </label>
          </div>

          <div class="grid gap-4 sm:grid-cols-2">
            <label class="block">
              <span class="mb-2 block text-sm font-medium text-foreground-secondary">
                Nomor HP
                <span class="font-normal text-muted">opsional</span>
              </span>
              <input
                v-model="form.picPhone"
                class="h-11 w-full rounded-lg border border-border bg-surface-subtle px-3 text-sm outline-none transition focus:border-brand focus:bg-surface"
                type="tel"
                autocomplete="tel"
                placeholder="081234567890"
              />
            </label>

            <label class="block">
              <span class="mb-2 block text-sm font-medium text-foreground-secondary">
                Peran PIC
                <span class="font-normal text-muted">opsional</span>
              </span>
              <input
                v-model="form.picRole"
                class="h-11 w-full rounded-lg border border-border bg-surface-subtle px-3 text-sm outline-none transition focus:border-brand focus:bg-surface"
                type="text"
                placeholder="Kepala sekolah"
              />
            </label>
          </div>

          <label class="block">
            <span class="mb-2 block text-sm font-medium text-foreground-secondary">
              Catatan
              <span class="font-normal text-muted">opsional</span>
            </span>
            <textarea
              v-model="form.message"
              class="min-h-28 w-full resize-y rounded-lg border border-border bg-surface-subtle px-3 py-3 text-sm outline-none transition focus:border-brand focus:bg-surface"
              placeholder="Ceritakan kebutuhan sekolah secara singkat."
            />
          </label>

          <p
            v-if="errorMessage"
            class="rounded-lg border border-[#ffd7d2] bg-[#fff7f5] px-4 py-3 text-sm text-danger"
          >
            {{ errorMessage }}
          </p>

          <button
            type="submit"
            :disabled="loading || !canSubmit"
            class="flex h-11 w-full items-center justify-center rounded-lg bg-brand px-5 text-sm font-medium text-white transition hover:bg-brand-hover disabled:cursor-not-allowed disabled:bg-[#bab7d8]"
          >
            {{ loading ? "Mengirim request..." : "Kirim request pendaftaran" }}
          </button>
        </form>
      </div>
    </section>
  </main>
</template>
