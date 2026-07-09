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
import DiscussionComments from "../../components/discussion/DiscussionComments.vue";
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
  <main class="min-h-screen min-w-0 flex-1 bg-[#f8f7f4]">
    <header class="border-b border-[#ebe7df] bg-white">
      <div
        class="flex min-w-0 items-center gap-2 px-5 py-5 text-xs text-[#6b7280] sm:px-6 lg:px-8"
      >
        <RouterLink
          class="inline-flex shrink-0 items-center gap-1.5 transition hover:text-[#4f46e5]"
          :to="`/student/subjects/${subjectClassId}`"
        >
          <PhArrowLeft :size="15" />
          Mata pelajaran
        </RouterLink>
        <span class="text-[#d1d5db]">/</span>
        <span class="shrink-0">Materi</span>
        <span class="text-[#d1d5db]">/</span>
        <span class="min-w-0 truncate font-medium text-[#171322]">
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
          class="h-52 animate-pulse rounded-xl border border-[#ebe7df] bg-white"
        />
        <div
          class="h-80 animate-pulse rounded-xl border border-[#ebe7df] bg-white"
        />
      </div>
      <div
        class="h-112 animate-pulse rounded-xl border border-[#ebe7df] bg-white"
      />
    </section>

    <section
      v-else-if="errorMessage"
      class="flex min-h-[calc(100vh-49px)] items-center justify-center px-5 py-10"
    >
      <article
        class="w-full max-w-xl rounded-xl border border-[#fecaca] bg-white p-6"
      >
        <div class="flex items-start gap-3">
          <div
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-[#fef2f2] text-[#dc2626]"
          >
            <PhWarningCircle :size="22" weight="duotone" />
          </div>
          <div>
            <h1 class="text-base font-medium text-[#171322]">
              Tidak bisa memuat materi
            </h1>
            <p class="mt-1 text-sm leading-6 text-[#7a7385]">
              {{ errorMessage }}
            </p>
            <button
              class="mt-4 rounded-lg bg-[#4f46e5] px-4 py-2 text-sm font-medium text-white transition hover:bg-[#4338ca]"
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
        class="w-full max-w-xl rounded-xl border border-[#ebe7df] bg-white p-8 text-center"
      >
        <div
          class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-[#eef2ff] text-[#4f46e5]"
        >
          <PhBookOpen :size="24" weight="duotone" />
        </div>
        <h1 class="mt-4 text-base font-medium text-[#171322]">
          Materi tidak ditemukan
        </h1>
        <p class="mx-auto mt-1 max-w-md text-sm leading-6 text-[#7a7385]">
          Materi ini tidak tersedia atau sudah tidak dapat diakses.
        </p>
        <RouterLink
          class="mt-5 inline-flex items-center gap-2 rounded-lg bg-[#4f46e5] px-4 py-2 text-sm font-medium text-white transition hover:bg-[#4338ca]"
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
        <article class="rounded-xl border border-[#ebe7df] bg-white p-5 sm:p-6">
          <div class="flex min-w-0 items-start gap-4">
            <div
              class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-[#eef2ff] text-[#4f46e5]"
            >
              <PhBookOpen :size="22" weight="duotone" />
            </div>
            <div class="min-w-0 flex-1">
              <div class="flex flex-wrap items-center gap-2">
                <span
                  v-if="material.materialType"
                  class="rounded-full bg-[#eef2ff] px-2.5 py-1 text-[11px] font-medium uppercase text-[#4f46e5]"
                >
                  {{ material.materialType }}
                </span>
                <span
                  v-if="material.subjectName"
                  class="rounded-full bg-[#f8f7f4] px-2.5 py-1 text-[11px] text-[#6b7280]"
                >
                  {{ material.subjectName }}
                </span>
              </div>
              <h1
                class="mt-3 wrap-break-word text-xl font-medium leading-7 text-[#171322] sm:text-2xl"
              >
                {{ material.materialTitle }}
              </h1>
              <div
                class="mt-3 flex flex-wrap items-center gap-x-4 gap-y-2 text-xs text-[#6b7280]"
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
            <h2 class="text-sm font-medium text-[#171322]">Deskripsi materi</h2>
            <p
              v-if="material.materialDesc"
              class="mt-3 whitespace-pre-line wrap-break-word text-sm leading-7 text-[#4a4356]"
            >
              {{ material.materialDesc }}
            </p>
            <p v-else class="mt-3 text-sm leading-6 text-[#7a7385]">
              Deskripsi materi belum tersedia.
            </p>
          </div>
        </article>

        <article class="rounded-xl border border-[#ebe7df] bg-white p-5 sm:p-6">
          <div class="flex items-center justify-between gap-3">
            <div class="flex items-center gap-2">
              <PhPaperclip :size="18" class="text-[#4f46e5]" />
              <h2 class="text-sm font-medium text-[#171322]">
                Lampiran materi
              </h2>
            </div>
            <span
              v-if="material.attachments?.length"
              class="shrink-0 rounded-full bg-[#f8f7f4] px-2.5 py-1 text-[11px] text-[#6b7280]"
            >
              {{ material.attachments.length }} file
            </span>
          </div>
          <p class="mt-1 text-xs leading-5 text-[#7a7385]">
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

        <!-- <article class="rounded-xl border border-[#ebe7df] bg-white p-5">
          <h2 class="text-sm font-medium text-[#171322]">Progress materi</h2>
          <p class="mt-1 text-sm leading-6 text-[#7a7385]">
            Progress materi direncanakan setelah MVP sekolah. Membuka materi
            belum menandai progres selesai.
          </p>
        </article> -->

        <DiscussionComments
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
