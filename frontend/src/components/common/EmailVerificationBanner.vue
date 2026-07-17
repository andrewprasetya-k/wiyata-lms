<script setup lang="ts">
import { ref } from "vue";
import { useAuthStore } from "../../stores/auth";
import { useToastStore } from "../../stores/toast";
import { resendVerificationEmail } from "../../services/emailVerification";
import { getApiError } from "../../utils/error";

const auth = useAuthStore();
const toast = useToastStore();
const sending = ref(false);

async function resend() {
  if (sending.value) return;
  sending.value = true;
  try {
    await resendVerificationEmail();
    toast.success("Email verifikasi sudah dikirim ulang.");
  } catch (error) {
    toast.error(getApiError(error));
  } finally {
    sending.value = false;
  }
}
</script>

<template>
  <div
    v-if="auth.isAuthenticated && auth.isContextReady && !auth.emailVerified"
    class="flex flex-col items-center justify-center gap-2 bg-[#fff4e5] px-4 py-2.5 text-center text-sm text-[#7a4a12] sm:flex-row"
  >
    <span>Verifikasi email kamu untuk membuka semua fitur Wiyata.</span>
    <button
      type="button"
      class="font-medium underline underline-offset-2 transition hover:text-[#4f3009] disabled:cursor-not-allowed disabled:opacity-60"
      :disabled="sending"
      @click="resend"
    >
      {{ sending ? "Mengirim..." : "Kirim ulang email verifikasi" }}
    </button>
  </div>
</template>
