<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink, useRoute } from "vue-router";
import {
  PhArrowLeft,
  PhBookOpen,
  PhPaperclip,
  PhUserCircle,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import AttachmentPreviewList from "../../components/common/AttachmentPreviewList.vue";
import CommentThread from "../../components/comments/CommentThread.vue";
import StudentNoteCard from "../../components/student/StudentNoteCard.vue";
import { getMaterialById } from "../../services/classWorkspace";
import type { MaterialItem } from "../../types/classWorkspace";
import { formatDateTime } from "../../utils/date";

const route = useRoute();
const subjectClassId = computed(() => String(route.params.sclId ?? ""));
const materialId = computed(() => String(route.params.matId ?? ""));
const material = ref<MaterialItem | null>(null);
const isLoading = ref(true);
const errorMessage = ref("");
const didLoad = ref(false);

async function loadMaterial() {
  if (!subjectClassId.value || !materialId.value) {
    isLoading.value = false;
    errorMessage.value = "Konteks materi tidak lengkap.";
    return;
  }

  isLoading.value = true;
  errorMessage.value = "";
  didLoad.value = false;

  try {
    material.value = await getMaterialById(materialId.value);
    didLoad.value = true;
  } catch {
    errorMessage.value =
      "Detail materi belum bisa dimuat. Periksa koneksi atau coba lagi nanti.";
  } finally {
    isLoading.value = false;
  }
}

onMounted(loadMaterial);
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 bg-background">
    <header class="border-b border-border bg-surface">
      <div
        class="flex min-w-0 items-center gap-2 px-5 py-5 text-xs text-muted sm:px-6 lg:px-8"
      >
        <RouterLink
          class="inline-flex shrink-0 items-center gap-1.5 transition hover:text-brand"
          :to="`/student/subjects/${subjectClassId}`"
        >
          <PhArrowLeft :size="15" />
          Mata pelajaran
        </RouterLink>
        <span class="text-[#d1d5db]">/</span>
        <span class="shrink-0">Materi</span>
        <span class="text-[#d1d5db]">/</span>
        <span class="min-w-0 truncate font-medium text-foreground">
          {{ material?.materialTitle || "Detail materi" }}
        </span>
      </div>
    </header>

    <section
      v-if="isLoading"
      class="mx-auto grid max-w-screen gap-5 px-5 py-5 sm:px-6 lg:grid-cols-[minmax(0,1fr)_360px] lg:px-8 lg:py-6"
    >
      <div class="space-y-4">
        <div
          class="h-52 animate-pulse rounded-xl border border-border bg-surface"
        />
        <div
          class="h-80 animate-pulse rounded-xl border border-border bg-surface"
        />
      </div>
      <div
        class="h-112 animate-pulse rounded-xl border border-border bg-surface"
      />
    </section>

    <section
      v-else-if="errorMessage"
      class="flex min-h-[calc(100vh-49px)] items-center justify-center px-5 py-10"
    >
      <article
        class="w-full max-w-xl rounded-xl border border-danger-line bg-danger-soft p-6"
      >
        <div class="flex items-start gap-3">
          <div
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-danger-soft text-danger"
          >
            <PhWarningCircle :size="22" weight="duotone" />
          </div>
          <div>
            <h1 class="text-base font-semibold text-foreground">
              Tidak bisa memuat materi
            </h1>
            <p class="mt-1 text-sm leading-6 text-muted">
              {{ errorMessage }}
            </p>
            <button
              class="mt-4 rounded-lg bg-brand px-4 py-2 text-sm font-medium text-white transition hover:bg-brand-hover"
              type="button"
              @click="loadMaterial"
            >
              Coba lagi
            </button>
          </div>
        </div>
      </article>
    </section>

    <section
      v-else-if="didLoad && !material"
      class="flex min-h-[calc(100vh-49px)] items-center justify-center px-5 py-10"
    >
      <article
        class="w-full max-w-xl rounded-xl border border-border bg-surface p-8 text-center"
      >
        <div
          class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-brand-soft text-brand"
        >
          <PhBookOpen :size="24" weight="duotone" />
        </div>
        <h1 class="mt-4 text-base font-semibold text-foreground">
          Materi tidak ditemukan
        </h1>
        <p class="mx-auto mt-1 max-w-md text-sm leading-6 text-muted">
          Materi ini tidak tersedia atau sudah tidak dapat diakses.
        </p>
        <RouterLink
          class="mt-5 inline-flex items-center gap-2 rounded-lg bg-brand px-4 py-2 text-sm font-medium text-white transition hover:bg-brand-hover"
          :to="`/student/subjects/${subjectClassId}`"
        >
          <PhArrowLeft :size="16" />
          Kembali ke mata pelajaran
        </RouterLink>
      </article>
    </section>

    <section
      v-else-if="material"
      class="mx-auto grid w-full max-w-screen min-w-0 items-start gap-5 px-5 py-5 sm:px-6 lg:grid-cols-[minmax(0,1fr)_360px] lg:px-8 lg:py-6"
    >
      <div class="min-w-0 space-y-4">
        <article class="rounded-xl border border-border bg-surface shadow-sm p-5 sm:p-6">
          <div class="flex min-w-0 items-start gap-4">
            <div
              class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-brand-soft text-brand"
            >
              <PhBookOpen :size="22" weight="duotone" />
            </div>
            <div class="min-w-0 flex-1">
              <div class="flex flex-wrap items-center gap-2">
                <span
                  v-if="material.materialType"
                  class="rounded-full bg-brand-soft px-2.5 py-1 text-[11px] font-medium uppercase text-brand"
                >
                  {{ material.materialType }}
                </span>
                <span
                  v-if="material.subjectName"
                  class="rounded-full bg-background px-2.5 py-1 text-[11px] text-muted"
                >
                  {{ material.subjectName }}
                </span>
              </div>
              <h1
                class="mt-3 wrap-break-word text-xl font-semibold leading-7 text-foreground sm:text-2xl"
              >
                {{ material.materialTitle }}
              </h1>
              <div
                class="mt-3 flex flex-wrap items-center gap-x-4 gap-y-2 text-xs text-muted"
              >
                <span class="inline-flex items-center gap-1.5">
                  <PhUserCircle :size="15" />
                  {{ material.creatorName || "Guru belum tersedia" }}
                </span>
                <span>{{ formatDateTime(material.createdAt) }}</span>
              </div>
            </div>
          </div>

          <div class="mt-6 border-t border-[#f0ede8] pt-5">
            <h2 class="text-sm font-semibold text-foreground">Deskripsi materi</h2>
            <p
              v-if="material.materialDesc"
              class="mt-3 whitespace-pre-line wrap-break-word text-sm leading-7 text-foreground"
            >
              {{ material.materialDesc }}
            </p>
            <p v-else class="mt-3 text-sm leading-6 text-muted">
              Deskripsi materi belum tersedia.
            </p>
          </div>
        </article>

        <article class="rounded-xl border border-border bg-surface shadow-sm p-5 sm:p-6">
          <div class="flex items-center justify-between gap-3">
            <div class="flex items-center gap-2">
              <PhPaperclip :size="18" class="text-brand" />
              <h2 class="text-sm font-semibold text-foreground">
                Lampiran materi
              </h2>
            </div>
            <span
              v-if="material.attachments?.length"
              class="shrink-0 rounded-full bg-background px-2.5 py-1 text-[11px] text-muted"
            >
              {{ material.attachments.length }} file
            </span>
          </div>
          <p class="mt-1 text-xs leading-5 text-muted">
            Buka atau pelajari file pembelajaran yang dibagikan guru.
          </p>
          <AttachmentPreviewList
            class="mt-4"
            :attachments="material.attachments"
            enable-ai-summary
            :material-id="material.materialId"
            empty-text="Materi ini tidak memiliki lampiran."
          />
        </article>

        <!-- <article class="rounded-xl border border-border bg-surface p-5">
          <h2 class="text-sm font-semibold text-foreground">Progress materi</h2>
          <p class="mt-1 text-sm leading-6 text-muted">
            Progress materi direncanakan setelah MVP sekolah. Membuka materi
            belum menandai progres selesai.
          </p>
        </article> -->

        <CommentThread
          source-type="material"
          :source-id="material.materialId"
          title="Diskusi materi"
          placeholder="Tulis pertanyaan atau komentar tentang materi ini..."
          empty-text="Belum ada diskusi untuk materi ini."
        />
      </div>

      <aside class="min-w-0 lg:sticky lg:top-6 lg:h-[calc(100vh-3rem)]">
        <StudentNoteCard
          :material-id="material.materialId"
          :subject-class-id="subjectClassId"
        />
      </aside>
    </section>
  </main>
</template>
