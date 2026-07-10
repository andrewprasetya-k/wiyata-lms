<script setup lang="ts">
import { ref } from "vue";
import {
  PhArrowSquareOut,
  PhFile,
  PhFilePdf,
  PhImage,
  PhSparkle,
  PhSpinnerGap,
} from "@phosphor-icons/vue";
import MaterialAiSummaryCard from "../material/MaterialAiSummaryCard.vue";
import { summarizeMaterialDocument } from "../../services/materialSummary";
import { useAuthStore } from "../../stores/auth";
import type { MaterialDocumentSummaryResponse } from "../../types/materialSummary";

interface AttachmentPreviewItem {
  mediaId: string;
  mediaName?: string;
  fileSize?: number;
  mimeType?: string;
  fileUrl?: string;
  thumbnailUrl?: string;
}

interface Props {
  attachments?: AttachmentPreviewItem[];
  materialId?: string;
  emptyText?: string;
  initiallyExpanded?: boolean;
  enableAiSummary?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  attachments: () => [],
  emptyText: "Tidak ada lampiran.",
  initiallyExpanded: true,
  enableAiSummary: false,
});

const auth = useAuthStore();
const failedImages = ref<Record<string, boolean>>({});
const expandedPreviews = ref<Record<string, boolean>>({});
const summaryLoading = ref<Record<string, boolean>>({});
const summaryErrors = ref<Record<string, string>>({});
const summaries = ref<Record<string, MaterialDocumentSummaryResponse>>({});
const summaryRequestIds = ref<Record<string, number>>({});
let summaryRequestSequence = 0;

function isSafeURL(value?: string) {
  if (!value || value.trim() !== value) return false;
  try {
    const parsed = new URL(value);
    return (
      (parsed.protocol === "http:" || parsed.protocol === "https:") &&
      Boolean(parsed.host)
    );
  } catch {
    return false;
  }
}

function isImage(attachment: AttachmentPreviewItem) {
  return attachment.mimeType?.toLowerCase().startsWith("image/") ?? false;
}

function isPDF(attachment: AttachmentPreviewItem) {
  const mimeType = attachment.mimeType?.toLowerCase() ?? "";
  const fileName = attachment.mediaName?.toLowerCase() ?? "";
  return mimeType === "application/pdf" || fileName.endsWith(".pdf");
}

function previewURL(attachment: AttachmentPreviewItem) {
  if (isSafeURL(attachment.thumbnailUrl)) return attachment.thumbnailUrl;
  if (isSafeURL(attachment.fileUrl)) return attachment.fileUrl;
  return "";
}

function isPreviewExpanded(attachment: AttachmentPreviewItem) {
  return expandedPreviews.value[attachment.mediaId] ?? props.initiallyExpanded;
}

function formatFileSize(size?: number) {
  if (size === undefined || size === null || size < 0) return "";
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${Math.round(size / 1024)} KB`;
  return `${(size / (1024 * 1024)).toFixed(1)} MB`;
}

function fileTypeLabel(attachment: AttachmentPreviewItem) {
  if (isImage(attachment)) return "Gambar";
  if (isPDF(attachment)) return "PDF";
  return attachment.mimeType || "File";
}

function markImageFailed(mediaId: string) {
  failedImages.value = { ...failedImages.value, [mediaId]: true };
}

function canSummarize(attachment: AttachmentPreviewItem) {
  return Boolean(
    props.enableAiSummary &&
      props.materialId &&
      attachment.mediaId &&
      isPDF(attachment),
  );
}

function isSummaryLoading(mediaId: string) {
  return Boolean(summaryLoading.value[mediaId]);
}

function summaryError(mediaId: string) {
  return summaryErrors.value[mediaId] ?? "";
}

function summaryResult(mediaId: string) {
  return summaries.value[mediaId];
}

async function summarizeAttachment(attachment: AttachmentPreviewItem) {
  if (!props.materialId || !canSummarize(attachment) || isSummaryLoading(attachment.mediaId)) {
    return;
  }

  const mediaId = attachment.mediaId;
  const requestId = ++summaryRequestSequence;
  const contextVersion = auth.contextVersion;
  summaryRequestIds.value = { ...summaryRequestIds.value, [mediaId]: requestId };
  summaryLoading.value = { ...summaryLoading.value, [mediaId]: true };
  summaryErrors.value = { ...summaryErrors.value, [mediaId]: "" };

  try {
    const result = await summarizeMaterialDocument(props.materialId, mediaId);
    if (!isCurrentSummaryRequest(mediaId, requestId, contextVersion)) return;
    summaries.value = { ...summaries.value, [mediaId]: result };
  } catch (error) {
    if (!isCurrentSummaryRequest(mediaId, requestId, contextVersion)) return;
    summaryErrors.value = {
      ...summaryErrors.value,
      [mediaId]: materialSummaryErrorMessage(error),
    };
  } finally {
    if (isCurrentSummaryRequest(mediaId, requestId, contextVersion)) {
      summaryLoading.value = { ...summaryLoading.value, [mediaId]: false };
    }
  }
}

function isCurrentSummaryRequest(mediaId: string, requestId: number, contextVersion: number) {
  return (
    summaryRequestIds.value[mediaId] === requestId &&
    auth.contextVersion === contextVersion
  );
}

function materialSummaryErrorMessage(error: unknown) {
  const status = responseStatus(error);
  if (status === 415) return "Saat ini rangkuman AI hanya mendukung file PDF.";
  if (status === 413) return "File terlalu besar untuk dirangkum saat ini.";
  if (status === 422) {
    return "PDF ini tidak memiliki teks yang dapat dibaca. PDF hasil scan belum didukung.";
  }
  if (status === 503) {
    return "Fitur rangkuman AI belum aktif atau sedang tidak tersedia.";
  }
  if (status === 403 || status === 404) {
    return "Kamu tidak memiliki akses ke dokumen ini atau dokumen tidak ditemukan.";
  }
  return "Gagal membuat rangkuman. Coba lagi nanti.";
}

function responseStatus(error: unknown) {
  if (
    typeof error === "object" &&
    error !== null &&
    "response" in error &&
    typeof (error as { response?: { status?: unknown } }).response?.status === "number"
  ) {
    return (error as { response: { status: number } }).response.status;
  }
  return undefined;
}
</script>

<template>
  <div>
    <p v-if="attachments.length === 0" class="text-sm leading-6 text-[#7a7385]">
      {{ emptyText }}
    </p>

    <div v-else class="space-y-3">
      <article
        v-for="attachment in attachments"
        :key="attachment.mediaId"
        class="overflow-hidden rounded-[18px] border border-border bg-[#fbfaf8]"
      >
        <div
          v-if="
            isImage(attachment) &&
            previewURL(attachment) &&
            !failedImages[attachment.mediaId] &&
            isPreviewExpanded(attachment)
          "
          class="border-b border-border bg-white"
        >
          <a
            v-if="isSafeURL(attachment.fileUrl)"
            :href="attachment.fileUrl"
            rel="noopener noreferrer"
            target="_blank"
            title="Buka gambar penuh di tab baru"
          >
            <img
              :alt="attachment.mediaName || 'Preview lampiran gambar'"
              class="max-h-96 w-full object-contain"
              :src="previewURL(attachment)"
              @error="markImageFailed(attachment.mediaId)"
            />
          </a>
          <img
            v-else
            :alt="attachment.mediaName || 'Preview lampiran gambar'"
            class="max-h-96 w-full object-contain"
            :src="previewURL(attachment)"
            @error="markImageFailed(attachment.mediaId)"
          />
        </div>

        <div
          v-else-if="
            isPDF(attachment) &&
            isSafeURL(attachment.fileUrl) &&
            isPreviewExpanded(attachment)
          "
          class="border-b border-border bg-white p-2 sm:p-3"
        >
          <iframe
            class="h-105 w-full rounded-xl bg-white sm:h-130"
            :src="attachment.fileUrl"
            :title="`Preview ${attachment.mediaName || 'PDF'}`"
            loading="lazy"
          />
        </div>

        <div
          class="flex min-w-0 flex-col gap-3 p-4 sm:flex-row sm:items-center"
        >
          <div
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-white text-brand"
          >
            <PhImage v-if="isImage(attachment)" :size="20" weight="duotone" />
            <PhFilePdf
              v-else-if="isPDF(attachment)"
              :size="20"
              weight="duotone"
            />
            <PhFile v-else :size="20" weight="duotone" />
          </div>

          <div class="min-w-0 flex-1 overflow-hidden">
            <p class="truncate text-sm font-medium text-foreground">
              {{ attachment.mediaName || "Lampiran" }}
            </p>
            <p class="mt-1 truncate text-xs text-[#8b8592]">
              {{ fileTypeLabel(attachment) }}
              <template v-if="formatFileSize(attachment.fileSize)">
                · {{ formatFileSize(attachment.fileSize) }}
              </template>
            </p>
            <p
              v-if="!isSafeURL(attachment.fileUrl)"
              class="mt-1 text-xs text-warning"
            >
              URL file tidak valid.
            </p>
          </div>

          <div
            v-if="isSafeURL(attachment.fileUrl) || canSummarize(attachment)"
            class="flex shrink-0 flex-wrap gap-2"
          >
            <button
              v-if="canSummarize(attachment)"
              class="inline-flex items-center gap-2 rounded-xl border border-border bg-white px-3 py-2 text-xs font-medium text-[#5b4b7a] transition hover:border-[#d8d1c5] hover:bg-[#f8f7f4] focus:outline-none focus:ring-2 focus:ring-brand/25 disabled:cursor-not-allowed disabled:opacity-60"
              type="button"
              :disabled="isSummaryLoading(attachment.mediaId)"
              @click="summarizeAttachment(attachment)"
            >
              <PhSpinnerGap
                v-if="isSummaryLoading(attachment.mediaId)"
                class="h-4 w-4 animate-spin"
              />
              <PhSparkle v-else :size="16" weight="duotone" />
              {{
                isSummaryLoading(attachment.mediaId)
                  ? "Merangkum..."
                  : summaryResult(attachment.mediaId)
                    ? "Rangkum ulang"
                    : "Rangkum dokumen"
              }}
            </button>
            <a
              v-if="isSafeURL(attachment.fileUrl)"
              class="inline-flex items-center gap-2 rounded-xl bg-white px-3 py-2 text-xs font-medium text-brand transition hover:bg-brand-soft focus:outline-none focus:ring-2 focus:ring-brand/30"
              :href="attachment.fileUrl"
              rel="noopener noreferrer"
              target="_blank"
            >
              <PhArrowSquareOut :size="16" />
              {{
                isPDF(attachment) ? "Buka Attachment di tab baru" : "Buka file"
              }}
            </a>
          </div>
        </div>

        <div
          v-if="
            canSummarize(attachment) &&
            (isSummaryLoading(attachment.mediaId) ||
              summaryError(attachment.mediaId) ||
              summaryResult(attachment.mediaId))
          "
          class="border-t border-border bg-white px-4 py-4"
        >
          <div
            v-if="isSummaryLoading(attachment.mediaId)"
            class="flex items-center gap-2 text-sm text-[#6b6475]"
          >
            <PhSpinnerGap class="h-4 w-4 animate-spin text-brand" />
            AI sedang membaca dokumen...
          </div>

          <div
            v-else-if="summaryError(attachment.mediaId)"
            class="rounded-xl border border-danger-line bg-[#fffaf9] p-4"
          >
            <p class="text-sm leading-6 text-[#8a463f]">
              {{ summaryError(attachment.mediaId) }}
            </p>
            <button
              class="mt-3 inline-flex items-center gap-2 rounded-lg border border-[#ead5d0] bg-white px-3 py-2 text-xs font-medium text-[#8a463f] transition hover:bg-[#fff1ef] focus:outline-none focus:ring-2 focus:ring-[#d97757]/25"
              type="button"
              @click="summarizeAttachment(attachment)"
            >
              Coba lagi
            </button>
          </div>

          <div
            v-else-if="summaryResult(attachment.mediaId)"
          >
            <MaterialAiSummaryCard
              :summary="summaryResult(attachment.mediaId)?.summary"
              :source-name="
                summaryResult(attachment.mediaId)?.source.mediaName ||
                attachment.mediaName
              "
            />
          </div>
        </div>
      </article>
    </div>
  </div>
</template>
