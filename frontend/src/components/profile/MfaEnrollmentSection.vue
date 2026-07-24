<script setup lang="ts">
import { onMounted, ref } from "vue";
import QRCode from "qrcode";
import {
  PhCheckCircle,
  PhCopy,
  PhDownloadSimple,
  PhShieldCheck,
} from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import { useToastStore } from "../../stores/toast";
import { confirmMfa, enrollMfa, getMfaStatus } from "../../services/mfa";
import { getApiError } from "../../utils/error";

const auth = useAuthStore();
const toast = useToastStore();

// The 409 branch in startEnroll() below is kept as a fallback (e.g. a race
// with another tab enabling MFA between the status check and the click),
// not the primary way this card learns its state.
type Step = "loading" | "idle" | "confirm" | "recovery" | "already-enabled";
const step = ref<Step>("loading");
const secret = ref("");
const qrDataUrl = ref("");
const code = ref("");
const isSubmitting = ref(false);
const errorMessage = ref("");
const recoveryCodes = ref<string[]>([]);
const copied = ref(false);

onMounted(async () => {
  try {
    const { enabled } = await getMfaStatus();
    step.value = enabled ? "already-enabled" : "idle";
  } catch {
    step.value = "idle";
  }
});

function mapError(error: unknown): string {
  const message = getApiError(error).toLowerCase();
  if (message.includes("invalid verification code")) {
    return "Kode tidak valid. Pastikan waktu di perangkatmu sudah tepat.";
  }
  return getApiError(error);
}

async function startEnroll() {
  isSubmitting.value = true;
  errorMessage.value = "";
  try {
    const result = await enrollMfa();
    secret.value = result.secret;
    qrDataUrl.value = await QRCode.toDataURL(result.otpauthUrl, {
      margin: 1,
      width: 220,
    });
    step.value = "confirm";
  } catch (error) {
    const message = getApiError(error).toLowerCase();
    if (message.includes("already enabled")) {
      step.value = "already-enabled";
    } else {
      errorMessage.value = mapError(error);
    }
  } finally {
    isSubmitting.value = false;
  }
}

async function submitConfirm() {
  if (code.value.trim().length !== 6 || isSubmitting.value) return;
  isSubmitting.value = true;
  errorMessage.value = "";
  try {
    const result = await confirmMfa(code.value.trim());
    recoveryCodes.value = result.recoveryCodes;
    auth.clearMfaGraceReminder();
    step.value = "recovery";
    toast.success("MFA berhasil diaktifkan.");
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

function cancel() {
  step.value = "idle";
  code.value = "";
  errorMessage.value = "";
}

function finish() {
  step.value = "already-enabled";
  recoveryCodes.value = [];
}
</script>

<template>
  <section class="min-w-0 rounded-xl border border-border bg-surface p-5">
    <div class="mb-4 flex min-w-0 items-center gap-3">
      <div
        class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-brand-soft text-brand"
      >
        <PhShieldCheck :size="21" weight="duotone" />
      </div>
      <div class="min-w-0">
        <h2 class="text-sm font-medium text-foreground">
          Verifikasi dua langkah
        </h2>
        <p class="mt-1 text-xs text-muted">
          Tambahkan lapisan keamanan ekstra memakai aplikasi autentikator.
        </p>
      </div>
    </div>

    <p v-if="step === 'loading'" class="text-sm text-muted">Memeriksa status...</p>

    <div v-else-if="step === 'idle'">
      <p
        v-if="errorMessage"
        class="mb-3 rounded-lg bg-danger-soft px-3 py-2.5 text-sm text-danger"
      >
        {{ errorMessage }}
      </p>
      <button
        type="button"
        class="inline-flex h-10 items-center justify-center rounded-lg bg-brand px-4 text-sm font-medium text-white transition hover:bg-brand-hover disabled:cursor-not-allowed disabled:bg-[#bab7d8]"
        :disabled="isSubmitting"
        @click="startEnroll"
      >
        {{ isSubmitting ? "Menyiapkan..." : "Aktifkan MFA" }}
      </button>
    </div>

    <p
      v-else-if="step === 'already-enabled'"
      class="flex items-center gap-2 text-sm text-success"
    >
      <PhCheckCircle :size="18" weight="fill" />
      MFA sudah aktif untuk akun ini.
    </p>

    <div v-else-if="step === 'confirm'" class="space-y-4">
      <div class="flex justify-center rounded-lg bg-surface-subtle p-5">
        <img :src="qrDataUrl" alt="QR code MFA" class="h-44 w-44" />
      </div>
      <div class="rounded-lg border border-border p-3">
        <p class="text-xs text-muted">
          Tidak bisa pindai? Masukkan kode ini secara manual:
        </p>
        <p
          class="mt-1.5 break-all font-mono text-sm font-medium text-foreground"
        >
          {{ secret }}
        </p>
      </div>

      <form class="space-y-3" @submit.prevent="submitConfirm">
        <label class="block max-w-55">
          <span
            class="mb-1.5 block text-xs font-medium text-foreground-secondary"
          >
            Kode konfirmasi
          </span>
          <input
            v-model="code"
            class="h-11 w-full rounded-lg border border-border bg-surface-subtle px-3 text-center text-base tracking-[0.4em] outline-none transition focus:border-brand focus:bg-surface"
            inputmode="numeric"
            autocomplete="one-time-code"
            maxlength="6"
            placeholder="000000"
          />
        </label>

        <p
          v-if="errorMessage"
          class="rounded-lg bg-danger-soft px-3 py-2.5 text-sm text-danger"
        >
          {{ errorMessage }}
        </p>

        <div class="flex gap-3">
          <button
            class="inline-flex h-10 items-center justify-center rounded-lg bg-brand px-4 text-sm font-medium text-white transition hover:bg-brand-hover disabled:cursor-not-allowed disabled:bg-[#bab7d8]"
            type="submit"
            :disabled="code.trim().length !== 6 || isSubmitting"
          >
            {{ isSubmitting ? "Memverifikasi..." : "Konfirmasi" }}
          </button>
          <button
            type="button"
            class="inline-flex h-10 items-center justify-center rounded-lg border border-border px-4 text-sm font-medium text-foreground-secondary transition hover:text-foreground"
            @click="cancel"
          >
            Batal
          </button>
        </div>
      </form>
    </div>

    <div v-else-if="step === 'recovery'" class="space-y-4">
      <p class="text-sm leading-6 text-muted">
        Simpan 10 recovery code berikut di tempat aman. Setiap kode hanya bisa
        dipakai <strong>satu kali</strong> jika kamu kehilangan akses ke
        aplikasi autentikator. Kode ini
        <strong>tidak akan ditampilkan lagi</strong>.
      </p>
      <div
        class="grid grid-cols-2 gap-2 rounded-lg bg-surface-subtle p-4 font-mono text-sm sm:grid-cols-3"
      >
        <span v-for="rc in recoveryCodes" :key="rc" class="text-foreground">
          {{ rc }}
        </span>
      </div>
      <div class="flex flex-wrap gap-3">
        <button
          type="button"
          class="inline-flex h-10 items-center gap-2 rounded-lg border border-border px-4 text-sm font-medium text-foreground-secondary transition hover:text-foreground"
          @click="copyRecoveryCodes"
        >
          <PhCopy :size="16" />
          {{ copied ? "Tersalin!" : "Salin" }}
        </button>
        <button
          type="button"
          class="inline-flex h-10 items-center gap-2 rounded-lg border border-border px-4 text-sm font-medium text-foreground-secondary transition hover:text-foreground"
          @click="downloadRecoveryCodes"
        >
          <PhDownloadSimple :size="16" />
          Unduh .txt
        </button>
        <button
          type="button"
          class="inline-flex h-10 items-center rounded-lg bg-brand px-4 text-sm font-medium text-white transition hover:bg-brand-hover"
          @click="finish"
        >
          Selesai
        </button>
      </div>
    </div>
  </section>
</template>
