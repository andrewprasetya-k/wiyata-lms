<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink, useRoute } from "vue-router";
import {
  PhArrowLeft,
  PhBookOpen,
  PhFloppyDisk,
  PhNotebook,
  PhTrash,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import AttachmentPreviewList from "../../components/common/AttachmentPreviewList.vue";
import { getMaterialById } from "../../services/classWorkspace";
import {
  deleteStudentMaterialNote,
  getStudentMaterialNote,
  saveStudentMaterialNote,
} from "../../services/studentNotes";
import { useToastStore } from "../../stores/toast";
import type { MaterialItem } from "../../types/classWorkspace";
import type { StudentMaterialNote } from "../../types/studentNotes";
import { formatDateTime } from "../../utils/date";

const maxLength = 10000;
const route = useRoute();
const toast = useToastStore();
const subjectClassId = computed(() => String(route.params.sclId ?? ""));
const materialId = computed(() => String(route.params.matId ?? ""));
const material = ref<MaterialItem | null>(null);
const note = ref<StudentMaterialNote | null>(null);
const content = ref("");
const savedContent = ref("");
const isLoading = ref(true);
const isSaving = ref(false);
const isDeleting = ref(false);
const errorMessage = ref("");

const normalizedContent = computed(() => content.value.trim());
const isTooLong = computed(() => Array.from(content.value).length > maxLength);
const hasChanges = computed(
  () => normalizedContent.value !== savedContent.value,
);
const canSave = computed(
  () =>
    normalizedContent.value.length > 0 &&
    !isTooLong.value &&
    hasChanges.value &&
    !isSaving.value &&
    !isDeleting.value,
);

function getErrorMessage(error: unknown, fallback: string) {
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
  return fallback;
}

async function loadPage() {
  if (!materialId.value || !subjectClassId.value) {
    isLoading.value = false;
    errorMessage.value = "Konteks materi tidak lengkap.";
    return;
  }

  isLoading.value = true;
  errorMessage.value = "";

  try {
    const [materialResponse, noteResponse] = await Promise.all([
      getMaterialById(materialId.value),
      getStudentMaterialNote(materialId.value),
    ]);
    material.value = materialResponse;
    note.value = noteResponse.note;
    content.value = noteResponse.note?.content ?? "";
    savedContent.value = noteResponse.note?.content.trim() ?? "";
  } catch (error) {
    errorMessage.value = getErrorMessage(
      error,
      "Editor catatan belum bisa dimuat. Coba lagi beberapa saat.",
    );
  } finally {
    isLoading.value = false;
  }
}

async function saveNote() {
  if (isTooLong.value) {
    errorMessage.value = "Catatan maksimal 10.000 karakter.";
    return;
  }
  if (!normalizedContent.value) {
    errorMessage.value = "Isi catatan wajib diisi.";
    return;
  }
  if (!canSave.value) return;

  isSaving.value = true;
  errorMessage.value = "";

  try {
    const response = await saveStudentMaterialNote(materialId.value, {
      content: normalizedContent.value,
    });
    note.value = response.note;
    content.value = response.note?.content ?? normalizedContent.value;
    savedContent.value = content.value.trim();
    toast.success("Catatan berhasil disimpan.");
  } catch (error) {
    errorMessage.value = getErrorMessage(
      error,
      "Catatan belum bisa disimpan. Teks yang kamu tulis tetap tersedia.",
    );
  } finally {
    isSaving.value = false;
  }
}

async function deleteNote() {
  if (!note.value || isDeleting.value) return;
  if (!window.confirm("Hapus catatan pribadi untuk materi ini?")) return;

  isDeleting.value = true;
  errorMessage.value = "";

  try {
    await deleteStudentMaterialNote(materialId.value);
    note.value = null;
    content.value = "";
    savedContent.value = "";
    toast.success("Catatan berhasil dihapus.");
  } catch (error) {
    errorMessage.value = getErrorMessage(
      error,
      "Catatan belum bisa dihapus.",
    );
  } finally {
    isDeleting.value = false;
  }
}

onMounted(loadPage);
</script>

<template>
  <main class="min-h-screen flex-1 px-5 py-5 sm:px-6 lg:px-8">
    <RouterLink
      class="mb-5 inline-flex items-center gap-2 rounded-md bg-white px-4 py-2 text-sm font-medium text-[#4f46e5] transition hover:bg-[#eef2ff]"
      :to="{
        name: 'student-material-detail',
        params: { sclId: subjectClassId, matId: materialId },
      }"
    >
      <PhArrowLeft :size="18" />
      Kembali ke materi
    </RouterLink>

    <section
      v-if="isLoading"
      class="grid w-full max-w-7xl gap-5 lg:grid-cols-[minmax(0,2fr)_minmax(0,3fr)]"
    >
      <div
        class="h-80 animate-pulse rounded-[22px] border border-[#ebe7df] bg-white"
      />
      <div
        class="h-150 animate-pulse rounded-[22px] border border-[#ebe7df] bg-white"
      />
    </section>

    <section
      v-else-if="errorMessage && !material"
      class="soft-card max-w-3xl rounded-[22px] p-5"
    >
      <div
        class="mb-4 flex h-11 w-11 items-center justify-center rounded-2xl bg-[#fff1f0] text-[#f2756a]"
      >
        <PhWarningCircle :size="24" weight="duotone" />
      </div>
      <p class="text-sm font-medium text-[#171322]">
        Tidak bisa membuka editor catatan
      </p>
      <p class="mt-2 text-sm leading-6 text-[#7a7385]">{{ errorMessage }}</p>
      <button
        class="mt-5 rounded-2xl bg-[#4f46e5] px-4 py-2 text-sm font-medium text-white transition hover:bg-[#4338ca]"
        type="button"
        @click="loadPage"
      >
        Coba lagi
      </button>
    </section>

    <section
      v-else-if="material"
      class="grid w-full max-w-7xl items-start gap-5 lg:grid-cols-[minmax(0,2fr)_minmax(0,3fr)]"
    >
      <aside class="min-w-0 space-y-4 lg:sticky lg:top-6">
        <article class="soft-card rounded-[22px] p-5">
          <div
            class="flex h-11 w-11 items-center justify-center rounded-2xl bg-[#eef2ff] text-[#4f46e5]"
          >
            <PhBookOpen :size="22" weight="duotone" />
          </div>
          <p class="mt-4 text-xs font-medium uppercase text-[#4f46e5]">
            {{ material.subjectName || "Materi kelas" }}
          </p>
          <h1 class="mt-2 text-2xl font-medium text-[#171322]">
            {{ material.materialTitle }}
          </h1>
          <p
            v-if="material.materialDesc"
            class="mt-4 line-clamp-6 whitespace-pre-line text-sm leading-6 text-[#6b6475]"
          >
            {{ material.materialDesc }}
          </p>
          <p v-else class="mt-4 text-sm leading-6 text-[#7a7385]">
            Deskripsi materi belum tersedia.
          </p>
          <p class="mt-4 text-xs text-[#a09aa8]">
            Dibuat {{ formatDateTime(material.createdAt) }}
          </p>
        </article>

        <article class="rounded-[22px] border border-[#ebe7df] bg-white p-5">
          <p class="text-sm font-medium text-[#171322]">Lampiran materi</p>
          <AttachmentPreviewList
            class="mt-3"
            :attachments="material.attachments"
            empty-text="Materi ini tidak memiliki lampiran."
            :initially-expanded="false"
          />
        </article>
      </aside>

      <article
        class="min-w-0 rounded-[22px] border border-[#ebe7df] bg-[#fbfaf8] p-5 sm:p-6"
      >
        <div class="flex items-start gap-3">
          <div
            class="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-[#f3ecff] text-[#7c3aed]"
          >
            <PhNotebook :size="22" weight="duotone" />
          </div>
          <div>
            <h2 class="text-xl font-medium text-[#171322]">Catatan Saya</h2>
            <p class="mt-1 text-sm leading-6 text-[#7a7385]">
              Tulis ringkasan, poin penting, atau hal yang ingin kamu ingat dari
              materi ini.
            </p>
            <p v-if="note?.updatedAt" class="mt-1 text-xs text-[#a09aa8]">
              Disimpan {{ formatDateTime(note.updatedAt) }}
            </p>
          </div>
        </div>

        <form class="mt-6" @submit.prevent="saveNote">
          <textarea
            v-model="content"
            class="min-h-128 w-full resize-y rounded-[18px] border border-[#ebe7df] bg-white px-5 py-5 text-sm leading-7 text-[#374151] outline-none transition placeholder:text-[#b3adb9] focus:border-[#c7d2fe] focus:ring-4 focus:ring-[#4f46e5]/10"
            placeholder="Mulai tulis catatanmu di sini..."
          />

          <div class="mt-2 flex flex-wrap items-center justify-between gap-2">
            <p
              class="text-xs"
              :class="isTooLong ? 'text-[#b42318]' : 'text-[#a09aa8]'"
            >
              {{ Array.from(content).length.toLocaleString("id-ID") }} / 10.000
              karakter
            </p>
            <p v-if="hasChanges && !isTooLong" class="text-xs text-[#8b8592]">
              Perubahan belum disimpan
            </p>
          </div>

          <p
            v-if="errorMessage"
            class="mt-3 rounded-2xl bg-[#fff1f0] px-4 py-3 text-sm leading-6 text-[#b42318]"
          >
            {{ errorMessage }}
          </p>

          <div class="mt-5 flex flex-wrap items-center gap-3">
            <button
              class="inline-flex items-center gap-2 rounded-2xl px-4 py-2 text-sm font-medium text-white transition disabled:cursor-not-allowed disabled:bg-[#d8d5dd]"
              :class="
                canSave ? 'bg-[#4f46e5] hover:bg-[#4338ca]' : 'bg-[#d8d5dd]'
              "
              :disabled="!canSave"
              type="submit"
            >
              <PhFloppyDisk :size="17" weight="duotone" />
              {{ isSaving ? "Menyimpan..." : "Simpan catatan" }}
            </button>

            <button
              v-if="note"
              class="inline-flex items-center gap-2 rounded-2xl border border-[#ebe7df] bg-white px-4 py-2 text-sm font-medium text-[#b42318] transition hover:border-[#fda29b] hover:bg-[#fff1f0] disabled:cursor-not-allowed disabled:opacity-60"
              :disabled="isSaving || isDeleting"
              type="button"
              @click="deleteNote"
            >
              <PhTrash :size="17" weight="duotone" />
              {{ isDeleting ? "Menghapus..." : "Hapus" }}
            </button>
          </div>
        </form>
      </article>
    </section>
  </main>
</template>
