<script setup lang="ts">
import { ref } from "vue";
import {
  PhArrowSquareOut,
  PhFile,
  PhFilePdf,
  PhImage,
} from "@phosphor-icons/vue";

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
  emptyText?: string;
  initiallyExpanded?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  attachments: () => [],
  emptyText: "Tidak ada lampiran.",
  initiallyExpanded: true,
});

const failedImages = ref<Record<string, boolean>>({});
const expandedPreviews = ref<Record<string, boolean>>({});

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
        class="overflow-hidden rounded-[18px] border border-[#ebe7df] bg-[#fbfaf8]"
      >
        <div
          v-if="
            isImage(attachment) &&
            previewURL(attachment) &&
            !failedImages[attachment.mediaId] &&
            isPreviewExpanded(attachment)
          "
          class="border-b border-[#ebe7df] bg-white"
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
          class="border-b border-[#ebe7df] bg-white p-2 sm:p-3"
        >
          <iframe
            class="h-105 w-full rounded-xl border border-[#ebe7df] bg-white sm:h-130"
            :src="attachment.fileUrl"
            :title="`Preview ${attachment.mediaName || 'PDF'}`"
            loading="lazy"
          />
        </div>

        <div
          class="flex min-w-0 flex-col gap-3 p-4 sm:flex-row sm:items-center"
        >
          <div
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-white text-[#4f46e5]"
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
            <p class="truncate text-sm font-medium text-[#171322]">
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
              class="mt-1 text-xs text-[#b45309]"
            >
              URL file tidak valid.
            </p>
          </div>

          <div
            v-if="isSafeURL(attachment.fileUrl)"
            class="flex shrink-0 flex-wrap gap-2"
          >
            <a
              class="inline-flex items-center gap-2 rounded-xl bg-white px-3 py-2 text-xs font-medium text-[#4f46e5] transition hover:bg-[#eef2ff] focus:outline-none focus:ring-2 focus:ring-[#4f46e5]/30"
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
      </article>
    </div>
  </div>
</template>
