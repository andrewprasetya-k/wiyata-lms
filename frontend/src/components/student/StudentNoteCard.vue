<script setup lang="ts">
import { ref, watch } from "vue";
import { RouterLink } from "vue-router";
import {
  PhArrowClockwise,
  PhArrowRight,
  PhNotebook,
} from "@phosphor-icons/vue";
import { getStudentMaterialNote } from "../../services/studentNotes";
import type { StudentMaterialNote } from "../../types/studentNotes";
import { formatDateTime } from "../../utils/date";

const props = defineProps<{
  materialId: string;
  subjectClassId: string;
}>();

const note = ref<StudentMaterialNote | null>(null);
const isLoading = ref(false);
const hasLoaded = ref(false);
const errorMessage = ref("");

function getNoteErrorMessage(error: unknown) {
  if (typeof error === "object" && error !== null && "response" in error) {
    const response = (
      error as {
        response?: {
          status?: number;
          data?: { error?: unknown; message?: unknown };
        };
      }
    ).response;

    if (response?.status === 403) {
      return "Catatan tidak dapat diakses karena materi ini tidak lagi tersedia di kelas aktifmu.";
    }
    if (response?.status === 404) {
      return "Materi untuk catatan ini tidak ditemukan.";
    }
    if (typeof response?.data?.error === "string") {
      return response.data.error;
    }
    if (typeof response?.data?.message === "string") {
      return response.data.message;
    }
  }
  return "Preview catatan belum bisa dimuat.";
}

async function loadNote() {
  if (!props.materialId) return;

  isLoading.value = true;
  hasLoaded.value = false;
  errorMessage.value = "";

  try {
    const response = await getStudentMaterialNote(props.materialId);
    note.value = response.note;
    hasLoaded.value = true;
  } catch (error) {
    errorMessage.value = getNoteErrorMessage(error);
  } finally {
    isLoading.value = false;
  }
}

watch(
  () => props.materialId,
  () => {
    void loadNote();
  },
  { immediate: true },
);
</script>

<template>
  <article
    class="flex rounded-[22px] border border-[#ebe7df] bg-[#fbfaf8] p-5 lg:h-full lg:min-h-0 lg:flex-col"
  >
    <div class="flex items-start gap-3">
      <div
        class="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-[#f3ecff] text-[#7c3aed]"
      >
        <PhNotebook :size="20" weight="duotone" />
      </div>
      <div class="min-w-0">
        <h2 class="text-base font-medium text-[#171322]">Catatan Saya</h2>
        <p class="mt-1 text-sm leading-6 text-[#7a7385]">
          Catatan pribadi untuk materi ini.
        </p>
      </div>
    </div>

    <div v-if="isLoading" class="mt-5 flex-1 space-y-2">
      <div class="h-4 animate-pulse rounded bg-white" />
      <div class="h-4 w-4/5 animate-pulse rounded bg-white" />
      <div class="h-4 w-2/3 animate-pulse rounded bg-white" />
    </div>

    <div
      v-else-if="!hasLoaded"
      class="mt-5 flex-1 rounded-[18px] bg-white p-4"
    >
      <p class="text-sm leading-6 text-[#b42318]">{{ errorMessage }}</p>
      <button
        class="mt-3 inline-flex items-center gap-2 text-sm font-medium text-[#4f46e5]"
        type="button"
        @click="loadNote"
      >
        <PhArrowClockwise :size="16" />
        Coba lagi
      </button>
    </div>

    <div v-else class="mt-5 flex min-h-0 flex-1 flex-col">
      <template v-if="note">
        <p
          class="line-clamp-5 whitespace-pre-line text-sm leading-6 text-[#4f4858] lg:overflow-y-auto"
        >
          {{ note.content }}
        </p>
        <p class="mt-3 text-xs text-[#a09aa8]">
          Disimpan {{ formatDateTime(note.updatedAt) }}
        </p>
      </template>
      <p v-else class="text-sm leading-6 text-[#7a7385]">
        Belum ada catatan. Buka editor untuk menulis ringkasan atau poin penting
        dari materi ini.
      </p>

      <RouterLink
        class="mt-5 inline-flex w-full shrink-0 items-center justify-center gap-2 rounded-2xl bg-[#4f46e5] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#4338ca] lg:mt-auto"
        :to="{
          name: 'student-material-note',
          params: {
            sclId: subjectClassId,
            matId: materialId,
          },
        }"
      >
        Buka editor catatan
        <PhArrowRight :size="17" />
      </RouterLink>
    </div>
  </article>
</template>
