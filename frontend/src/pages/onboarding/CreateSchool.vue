<script setup lang="ts">
import { computed, reactive, ref } from "vue";
import { RouterLink, useRouter } from "vue-router";
import { PhArrowRight } from "@phosphor-icons/vue";
import { createSchool } from "../../services/school";
import { resendVerificationEmail } from "../../services/emailVerification";
import { useAuthStore } from "../../stores/auth";
import { useToastStore } from "../../stores/toast";
import { getApiError } from "../../utils/error";
import { classifyApiError } from "../../utils/errorPresentation";
import InlineFormError from "../../components/common/InlineFormError.vue";

const router = useRouter();
const auth = useAuthStore();
const toast = useToastStore();

const form = reactive({
  schoolName: "",
});

const loading = ref(false);
const validationError = ref("");
const verificationRequired = ref(false);
const resending = ref(false);

const canSubmit = computed(() => form.schoolName.trim() !== "");

async function submit() {
  validationError.value = "";
  verificationRequired.value = false;

  if (!canSubmit.value || loading.value) {
    validationError.value = "Isi nama sekolah terlebih dahulu.";
    return;
  }

  loading.value = true;

  try {
    const result = await createSchool({ schoolName: form.schoolName.trim() });

    await auth.refreshUserContext();
    const landingRoute = auth.switchContext({
      type: "school",
      schoolId: result.school.schoolId,
      schoolUserId: result.schoolUserId,
      role: "admin",
    });

    router.push(landingRoute ?? "/admin/dashboard");
  } catch (error) {
    const category = classifyApiError(error);
    if (category === "permission") {
      verificationRequired.value = true;
    } else if (category === "business_rule") {
      toast.error(
        "Sekolah belum bisa dibuat karena ada konflik data. Coba lagi sebentar lagi.",
      );
    } else {
      toast.error(getApiError(error));
    }
  } finally {
    loading.value = false;
  }
}

async function resendVerification() {
  if (resending.value) return;
  resending.value = true;
  try {
    await resendVerificationEmail();
    toast.success("Email verifikasi sudah dikirim ulang.");
  } catch (error) {
    toast.error(getApiError(error));
  } finally {
    resending.value = false;
  }
}
</script>

<template>
  <main class="min-h-screen bg-surface-subtle px-6 py-8 text-foreground">
    <div class="mx-auto flex w-full max-w-screen items-center justify-between">
      <RouterLink to="/home" class="flex items-center gap-3">
        <img src="/logo_fix.svg" alt="Wiyata" class="h-9 w-9 rounded-lg" />
        <span class="text-sm font-semibold">Wiyata Academic Workspace</span>
      </RouterLink>
      <RouterLink
        to="/onboarding"
        class="rounded-lg border border-border bg-surface px-4 py-2 text-sm font-medium text-foreground-secondary transition hover:text-foreground"
      >
        Kembali
      </RouterLink>
    </div>

    <section class="mx-auto mt-16 max-w-screen">
      <p class="text-sm font-medium text-brand">Buat sekolah</p>
      <h1 class="mt-3 text-3xl font-semibold leading-tight sm:text-4xl">
        Beri nama sekolahmu untuk mulai.
      </h1>
      <p class="mt-4 text-sm leading-6 text-muted">
        Sekolah langsung aktif setelah dibuat dan kamu otomatis menjadi Admin.
        Detail lain seperti alamat, email, dan telepon sekolah bisa dilengkapi
        kapan saja lewat halaman Edit Sekolah.
      </p>

      <form
        class="mt-8 space-y-5 rounded-xl border border-border bg-surface p-6 shadow-sm"
        @submit.prevent="submit"
      >
        <label class="block">
          <span
            class="mb-2 block text-sm font-medium text-foreground-secondary"
          >
            Nama sekolah
          </span>
          <input
            v-model="form.schoolName"
            class="h-11 w-full rounded-lg border border-border bg-surface-subtle px-3 text-sm outline-none transition focus:border-brand focus:bg-surface"
            type="text"
            autocomplete="organization"
            placeholder="SMA Wiyata Mandala"
            autofocus
            :disabled="loading"
          />
        </label>

        <InlineFormError :message="validationError" />

        <div
          v-if="verificationRequired"
          class="space-y-3 rounded-lg border border-[#ffd7d2] bg-[#fff7f5] px-4 py-3 text-sm text-danger"
        >
          <p>
            Silakan verifikasi email terlebih dahulu sebelum membuat sekolah.
          </p>
          <button
            type="button"
            class="font-medium underline underline-offset-2 transition hover:text-[#9f2a1d] disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="resending"
            @click="resendVerification"
          >
            {{ resending ? "Mengirim..." : "Kirim ulang email verifikasi" }}
          </button>
        </div>

        <button
          type="submit"
          :disabled="loading || !canSubmit"
          class="flex h-11 w-full items-center justify-center gap-1.5 rounded-lg bg-brand px-5 text-sm font-medium text-white transition hover:bg-brand-hover disabled:cursor-not-allowed disabled:bg-[#bab7d8]"
        >
          {{ loading ? "Membuat sekolah..." : "Buat sekolah" }}
          <PhArrowRight v-if="!loading" :size="16" />
        </button>
      </form>
    </section>
  </main>
</template>
