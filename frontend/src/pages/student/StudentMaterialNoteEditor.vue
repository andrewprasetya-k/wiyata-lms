<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink, useRoute } from "vue-router";
import {
  PhArrowLeft,
  PhBookOpen,
  PhCheckCircle,
  PhClock,
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
import { useActiveClassStore } from "../../stores/activeClass";
import { useToastStore } from "../../stores/toast";
import { useConfirmStore } from "../../stores/confirm";
import type { MaterialItem } from "../../types/classWorkspace";
import type { StudentMaterialNote } from "../../types/studentNotes";
import { formatDateTime } from "../../utils/date";

const maxLength = 10000;
const route = useRoute();
const toast = useToastStore();
const confirm = useConfirmStore();
const activeClassStore = useActiveClassStore();
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
const saveStatus = computed(() => {
  if (isSaving.value) return "Menyimpan...";
  if (hasChanges.value) return "Belum disimpan";
  if (note.value) return "Tersimpan";
  return "Belum ada catatan";
});
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
  const ok = await confirm.confirm({
    title: "Hapus catatan?",
    description: "Catatan pribadi untuk materi ini akan dihapus permanen.",
    confirmLabel: "Hapus",
    variant: "danger",
  });
  if (!ok) return;

  isDeleting.value = true;
  errorMessage.value = "";

  try {
    await deleteStudentMaterialNote(materialId.value);
    note.value = null;
    content.value = "";
    savedContent.value = "";
    toast.success("Catatan berhasil dihapus.");
  } catch (error) {
    errorMessage.value = getErrorMessage(error, "Catatan belum bisa dihapus.");
  } finally {
    isDeleting.value = false;
  }
}

onMounted(loadPage);
</script>

<template>
  <main class="min-h-screen flex-1 bg-[#f8f7f4]">
    <header
      class="sticky top-0 z-10 border-b border-border bg-white/95 px-5 py-3 backdrop-blur sm:px-6 lg:px-8"
    >
      <div
        class="mx-auto flex w-full max-w-400 items-center justify-between gap-4"
      >
        <nav
          class="flex min-w-0 items-center gap-2 text-xs text-[#7a7385]"
          aria-label="Breadcrumb"
        >
          <RouterLink
            class="inline-flex shrink-0 items-center gap-1.5 font-medium transition hover:text-brand"
            :to="{
              name: 'student-material-detail',
              params: { sclId: subjectClassId, matId: materialId },
            }"
          >
            <PhArrowLeft :size="16" />
            Kembali
          </RouterLink>
          <span class="text-[#d1ccd5]">/</span>
          <span class="hidden truncate sm:inline">
            {{
              activeClassStore.activeClass?.classTitle ||
              material?.subjectName ||
              "Kelas aktif"
            }}
          </span>
          <span class="hidden text-[#d1ccd5] sm:inline">/</span>
          <span class="truncate font-medium text-foreground">
            {{ material?.materialTitle || "Editor catatan" }}
          </span>
        </nav>
      </div>
    </header>

    <section
      v-if="isLoading"
      class="mx-auto grid w-full max-w-400 gap-5 px-5 py-5 sm:px-6 lg:grid-cols-[minmax(0,3fr)_minmax(340px,2fr)] lg:px-8"
    >
      <div
        class="h-150 animate-pulse rounded-[22px] border border-border bg-white"
      />
      <div
        class="h-150 animate-pulse rounded-[22px] border border-border bg-white"
      />
    </section>

    <section
      v-else-if="errorMessage && !material"
      class="soft-card mx-5 mt-5 max-w-3xl rounded-[22px] p-5 sm:mx-6 lg:mx-8"
    >
      <div
        class="mb-4 flex h-11 w-11 items-center justify-center rounded-xl bg-danger-soft text-[#f2756a]"
      >
        <PhWarningCircle :size="24" weight="duotone" />
      </div>
      <p class="text-sm font-medium text-foreground">
        Tidak bisa membuka editor catatan
      </p>
      <p class="mt-2 text-sm leading-6 text-[#7a7385]">{{ errorMessage }}</p>
      <button
        class="mt-5 rounded-xl bg-brand px-4 py-2 text-sm font-medium text-white transition hover:bg-brand-hover"
        type="button"
        @click="loadPage"
      >
        Coba lagi
      </button>
    </section>

    <section
      v-else-if="material"
      class="mx-auto grid w-full max-w-400 items-start gap-5 px-5 py-5 sm:px-6 lg:grid-cols-[minmax(0,3fr)_minmax(340px,2fr)] lg:px-8"
    >
      <section
        class="min-w-0 overflow-hidden rounded-[22px] border border-border bg-white"
      >
        <header class="border-b border-border px-5 py-4 sm:px-6">
          <div class="flex flex-wrap items-center gap-2">
            <span
              class="inline-flex items-center gap-1.5 rounded-lg bg-brand-soft px-2.5 py-1 text-xs font-medium text-brand"
            >
              <PhBookOpen :size="14" weight="duotone" />
              {{ material.subjectName || "Materi kelas" }}
            </span>
            <span
              v-if="material.materialType"
              class="rounded-lg bg-[#f3f1ec] px-2.5 py-1 text-xs font-medium uppercase text-muted"
            >
              {{ material.materialType }}
            </span>
          </div>

          <h1 class="mt-3 text-2xl font-medium text-foreground">
            {{ material.materialTitle }}
          </h1>
          <p class="mt-1 text-xs text-[#8b8592]">
            <template v-if="material.creatorName">
              {{ material.creatorName }} ·
            </template>
            {{ formatDateTime(material.createdAt) }}
          </p>
        </header>

        <div class="space-y-5 bg-[#f4f3f1] p-4 sm:p-6">
          <article
            class="mx-auto w-full max-w-4xl rounded-[18px] border border-border bg-white p-5 shadow-[0_8px_28px_rgba(53,45,35,0.06)] sm:p-6"
          >
            <div class="flex items-center justify-between gap-3">
              <div>
                <p class="text-sm font-medium text-foreground">
                  Lampiran materi
                </p>
                <p class="mt-1 text-xs text-[#8b8592]">
                  Preview file tetap tersedia saat kamu menulis catatan.
                </p>
              </div>
              <span
                class="shrink-0 rounded-lg bg-[#f3f1ec] px-2.5 py-1 text-xs text-muted"
              >
                {{ material.attachments?.length || 0 }} file
              </span>
            </div>
            <AttachmentPreviewList
              class="mt-4"
              :attachments="material.attachments"
              empty-text="Materi ini tidak memiliki lampiran."
            />
          </article>
        </div>
      </section>

      <article
        class="flex min-h-168 min-w-0 flex-col overflow-hidden rounded-[22px] border border-border bg-[#fbfaf8] lg:sticky lg:top-17 lg:h-[calc(100vh-6rem)] lg:min-h-0"
      >
        <header
          class="flex shrink-0 items-start justify-between gap-3 border-b border-border bg-white px-5 py-4"
        >
          <div class="flex min-w-0 items-start gap-3">
            <div
              class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-brand-soft text-brand"
            >
              <PhNotebook :size="18" weight="duotone" />
            </div>
            <div class="min-w-0">
              <h2 class="truncate text-sm font-medium text-foreground">
                Catatan — {{ material.materialTitle }}
              </h2>
              <p class="mt-1 text-xs leading-5 text-[#8b8592]">
                Catatan ini hanya dapat dilihat olehmu.
              </p>
            </div>
          </div>

          <div
            class="inline-flex shrink-0 items-center gap-1.5 rounded-lg px-2.5 py-1 text-xs font-medium"
            :class="
              hasChanges
                ? 'bg-warning-soft text-warning'
                : 'bg-[#f0fdf4] text-[#059669]'
            "
          >
            <PhClock v-if="hasChanges" :size="14" />
            <PhCheckCircle v-else :size="14" />
            {{ saveStatus }}
          </div>
        </header>

        <form class="flex min-h-0 flex-1 flex-col" @submit.prevent="saveNote">
          <div class="flex min-h-0 flex-1 flex-col overflow-y-auto px-5 py-4">
            <p class="mb-3 text-xs leading-5 text-[#a09aa8]">
              Tulis ringkasan, poin penting, atau hal yang ingin kamu ingat.
            </p>
            <textarea
              v-model="content"
              class="min-h-120 w-full flex-1 resize-none border-0 bg-transparent p-0 text-sm leading-7 text-[#374151] outline-none placeholder:text-[#b3adb9] focus:ring-0 lg:min-h-0"
              placeholder="Mulai tulis catatanmu di sini..."
            />
          </div>

          <p
            v-if="errorMessage"
            class="mx-5 mb-3 rounded-xl bg-danger-soft px-4 py-3 text-sm leading-6 text-[#b42318]"
          >
            {{ errorMessage }}
          </p>

          <footer class="shrink-0 border-t border-border bg-white px-5 py-4">
            <div
              class="flex flex-wrap items-center justify-between gap-x-4 gap-y-2"
            >
              <div class="flex flex-wrap items-center gap-x-4 gap-y-1">
                <p
                  class="text-xs"
                  :class="isTooLong ? 'text-[#b42318]' : 'text-[#a09aa8]'"
                >
                  {{ Array.from(content).length.toLocaleString("id-ID") }} /
                  10.000 karakter
                </p>
                <p v-if="note?.updatedAt" class="text-xs text-[#a09aa8]">
                  Disimpan {{ formatDateTime(note.updatedAt) }}
                </p>
              </div>

              <div class="flex flex-wrap items-center gap-2">
                <button
                  v-if="note"
                  class="inline-flex items-center gap-2 rounded-xl border border-border bg-white px-3 py-2 text-xs font-medium text-[#b42318] transition hover:border-[#fda29b] hover:bg-danger-soft disabled:cursor-not-allowed disabled:opacity-60"
                  :disabled="isSaving || isDeleting"
                  type="button"
                  @click="deleteNote"
                >
                  <PhTrash :size="16" weight="duotone" />
                  {{ isDeleting ? "Menghapus..." : "Hapus" }}
                </button>

                <button
                  class="inline-flex items-center gap-2 rounded-xl px-3.5 py-2 text-xs font-medium text-white transition disabled:cursor-not-allowed disabled:bg-[#d8d5dd]"
                  :class="
                    canSave ? 'bg-brand hover:bg-brand-hover' : 'bg-[#d8d5dd]'
                  "
                  :disabled="!canSave"
                  type="submit"
                >
                  <PhFloppyDisk :size="16" weight="duotone" />
                  {{ isSaving ? "Menyimpan..." : "Simpan catatan" }}
                </button>
              </div>
            </div>
          </footer>
        </form>
      </article>
    </section>
  </main>
</template>
