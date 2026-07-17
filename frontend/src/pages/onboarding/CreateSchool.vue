<script setup lang="ts">
import { computed, reactive, ref } from "vue";
import { RouterLink, useRouter } from "vue-router";
import { PhArrowRight } from "@phosphor-icons/vue";
import { createSchool } from "../../services/school";
import { useAuthStore } from "../../stores/auth";
import { getApiError } from "../../utils/error";

const router = useRouter();
const auth = useAuthStore();

const form = reactive({
  schoolName: "",
});

const loading = ref(false);
const errorMessage = ref("");

const canSubmit = computed(() => form.schoolName.trim() !== "");

async function submit() {
  if (!canSubmit.value || loading.value) {
    errorMessage.value = "Isi nama sekolah terlebih dahulu.";
    return;
  }

  loading.value = true;
  errorMessage.value = "";

  try {
    const result = await createSchool({ schoolName: form.schoolName.trim() });

    // Server already committed School + SchoolUser + Admin role atomically —
    // refresh memberships first so switchContext() can find this school.
    await auth.refreshUserContext();
    const landingRoute = auth.switchContext({
      type: "school",
      schoolId: result.school.schoolId,
      schoolUserId: result.schoolUserId,
      role: "admin",
    });

    router.push(landingRoute ?? "/admin/dashboard");
  } catch (error) {
    errorMessage.value = getApiError(error);
  } finally {
    loading.value = false;
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

    <section class="mx-auto mt-16 max-w-xl">
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
          <span class="mb-2 block text-sm font-medium text-foreground-secondary">
            Nama sekolah
          </span>
          <input
            v-model="form.schoolName"
            class="h-11 w-full rounded-lg border border-border bg-surface-subtle px-3 text-sm outline-none transition focus:border-brand focus:bg-surface"
            type="text"
            autocomplete="organization"
            placeholder="SMA Wiyata Mandala"
            autofocus
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
          class="flex h-11 w-full items-center justify-center gap-1.5 rounded-lg bg-brand px-5 text-sm font-medium text-white transition hover:bg-brand-hover disabled:cursor-not-allowed disabled:bg-[#bab7d8]"
        >
          {{ loading ? "Membuat sekolah..." : "Buat sekolah" }}
          <PhArrowRight v-if="!loading" :size="16" />
        </button>
      </form>
    </section>
  </main>
</template>
