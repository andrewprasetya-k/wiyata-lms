<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import QRCode from "qrcode";
import {
  PhArrowRight,
  PhCheckCircle,
  PhCopy,
  PhDownloadSimple,
} from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import { useToastStore } from "../../stores/toast";
import { confirmMfaSetup, enrollMfaSetup } from "../../services/mfa";
import { getApiError } from "../../utils/error";

const auth = useAuthStore();
const route = useRoute();
const router = useRouter();
const toast = useToastStore();

type Step = "loading" | "enroll" | "recovery" | "error";
const step = ref<Step>("loading");
const secret = ref("");
const qrDataUrl = ref("");
const code = ref("");
const isSubmitting = ref(false);
const errorMessage = ref("");
const recoveryCodes = ref<string[]>([]);
const copied = ref(false);

function mapError(error: unknown): string {
  const message = getApiError(error).toLowerCase();
  if (message.includes("invalid or expired")) {
    return "Sesi setup sudah kedaluwarsa. Silakan login kembali.";
  }
  if (message.includes("too many failed attempts")) {
    return "Terlalu banyak percobaan gagal. Coba lagi dalam beberapa menit.";
  }
  if (message.includes("invalid verification code")) {
    return "Kode tidak valid. Pastikan waktu di perangkatmu sudah tepat, lalu coba lagi.";
  }
  if (message.includes("already enabled")) {
    return "MFA sudah aktif untuk akun ini. Silakan login kembali.";
  }
  return getApiError(error);
}

async function startEnrollment() {
  const preAuthToken = auth.pendingMfaToken;
  if (!preAuthToken) {
    step.value = "error";
    errorMessage.value = "Sesi setup tidak ditemukan. Silakan login kembali.";
    return;
  }

  try {
    const result = await enrollMfaSetup(preAuthToken);
    secret.value = result.secret;
    qrDataUrl.value = await QRCode.toDataURL(result.otpauthUrl, {
      margin: 1,
      width: 220,
    });
    step.value = "enroll";
  } catch (error) {
    step.value = "error";
    errorMessage.value = mapError(error);
  }
}

async function submitConfirm() {
  if (isSubmitting.value || code.value.trim().length !== 6) return;

  const preAuthToken = auth.pendingMfaToken;
  if (!preAuthToken) {
    step.value = "error";
    errorMessage.value = "Sesi setup tidak ditemukan. Silakan login kembali.";
    return;
  }

  isSubmitting.value = true;
  errorMessage.value = "";
  try {
    const result = await confirmMfaSetup(preAuthToken, code.value.trim());
    auth.completeMfaLogin(result);
    recoveryCodes.value = result.recoveryCodes;
    step.value = "recovery";
  } catch (error) {
    errorMessage.value = mapError(error);
  } finally {
    isSubmitting.value = false;
  }
}

async function copyRecoveryCodes() {
  try {
    await navigator.clipboard.writeText(recoveryCodes.value.join("\n"));
    copied.value = true;
    window.setTimeout(() => (copied.value = false), 2000);
  } catch {
    toast.error("Gagal menyalin. Coba salin manual.");
  }
}

function downloadRecoveryCodes() {
  const blob = new Blob(
    [
      "Wiyata - Recovery codes MFA\n" +
        "Simpan daftar ini di tempat aman. Setiap kode hanya bisa dipakai sekali.\n\n" +
        recoveryCodes.value.join("\n") +
        "\n",
    ],
    { type: "text/plain" },
  );
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = "wiyata-recovery-codes.txt";
  link.click();
  URL.revokeObjectURL(url);
}

function backToLogin() {
  auth.clearPendingMfa();
  router.push("/login");
}

async function continueToDashboard() {
  const redirect = route.query.redirect as string | undefined;
  await router.push(redirect ?? auth.landingRoute());
}

onMounted(startEnrollment);
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
          Aktifkan verifikasi dua langkah.
        </h1>
        <p class="mt-6 text-base leading-7 text-muted">
          Masa tenggang akunmu untuk mengaktifkan MFA sudah berakhir.
          Selesaikan setup ini untuk melanjutkan masuk.
        </p>
      </div>

      <div class="text-xs text-muted">
        &copy; 2026 Wiyata. All rights reserved.
      </div>
    </section>

    <!-- Right Side: Setup Flow -->
    <section
      class="flex h-screen items-center justify-center overflow-y-auto bg-surface px-6 py-8 sm:px-12"
    >
      <div class="w-full max-w-md">
        <div
          v-if="step === 'loading'"
          class="py-16 text-center text-sm text-muted"
        >
          Menyiapkan setup MFA...
        </div>

        <div v-else-if="step === 'error'" class="space-y-5">
          <p class="rounded-2xl bg-danger-soft px-4 py-3 text-sm text-danger">
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

        <div v-else-if="step === 'enroll'" class="space-y-5">
          <div>
            <h2 class="text-3xl font-medium text-foreground">
              Setup verifikasi dua langkah
            </h2>
            <p class="mt-3 text-sm text-muted">
              Wajib diaktifkan sebelum melanjutkan. Pindai kode QR memakai
              aplikasi autentikator (Google Authenticator, Authy, dsb).
            </p>
          </div>

          <div class="flex justify-center rounded-2xl bg-surface-subtle p-5">
            <img :src="qrDataUrl" alt="QR code MFA" class="h-48 w-48" />
          </div>

          <div class="rounded-2xl border border-border p-4">
            <p class="text-xs text-muted">
              Tidak bisa pindai? Masukkan kode ini secara manual:
            </p>
            <p
              class="mt-1.5 break-all font-mono text-sm font-medium text-foreground"
            >
              {{ secret }}
            </p>
          </div>

          <form class="space-y-4" @submit.prevent="submitConfirm">
            <label class="block">
              <span
                class="mb-2 block text-sm font-medium text-foreground-secondary"
              >
                Kode konfirmasi
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

            <p
              v-if="errorMessage"
              class="rounded-2xl bg-danger-soft px-4 py-3 text-sm text-danger"
            >
              {{ errorMessage }}
            </p>

            <button
              class="flex h-12 w-full items-center justify-center gap-2 rounded-2xl bg-brand text-sm font-medium text-white transition hover:bg-brand-hover disabled:cursor-not-allowed disabled:bg-[#bab7d8]"
              type="submit"
              :disabled="code.trim().length !== 6 || isSubmitting"
            >
              {{ isSubmitting ? "Memverifikasi..." : "Aktifkan MFA" }}
              <PhArrowRight v-if="!isSubmitting" :size="18" />
            </button>
          </form>
        </div>

        <div v-else class="space-y-5">
          <div
            class="flex h-12 w-12 items-center justify-center rounded-2xl bg-success-soft text-success"
          >
            <PhCheckCircle :size="24" weight="fill" />
          </div>
          <div>
            <h2 class="text-2xl font-medium text-foreground">
              MFA berhasil diaktifkan
            </h2>
            <p class="mt-3 text-sm leading-6 text-muted">
              Simpan 10 recovery code berikut di tempat aman. Setiap kode
              hanya bisa dipakai
              <strong>satu kali</strong> untuk masuk jika kamu kehilangan
              akses ke aplikasi autentikator. Kode ini
              <strong>tidak akan ditampilkan lagi</strong>.
            </p>
          </div>

          <div
            class="grid grid-cols-2 gap-2 rounded-2xl bg-surface-subtle p-4 font-mono text-sm"
          >
            <span v-for="rc in recoveryCodes" :key="rc" class="text-foreground">
              {{ rc }}
            </span>
          </div>

          <div class="flex gap-3">
            <button
              type="button"
              class="flex h-11 flex-1 items-center justify-center gap-2 rounded-2xl border border-border text-sm font-medium text-foreground-secondary transition hover:text-foreground"
              @click="copyRecoveryCodes"
            >
              <PhCopy :size="17" />
              {{ copied ? "Tersalin!" : "Salin" }}
            </button>
            <button
              type="button"
              class="flex h-11 flex-1 items-center justify-center gap-2 rounded-2xl border border-border text-sm font-medium text-foreground-secondary transition hover:text-foreground"
              @click="downloadRecoveryCodes"
            >
              <PhDownloadSimple :size="17" />
              Unduh .txt
            </button>
          </div>

          <button
            type="button"
            class="flex h-12 w-full items-center justify-center gap-2 rounded-2xl bg-brand text-sm font-medium text-white transition hover:bg-brand-hover"
            @click="continueToDashboard"
          >
            Lanjut ke Dashboard
            <PhArrowRight :size="18" />
          </button>
        </div>
      </div>
    </section>
  </main>
</template>
