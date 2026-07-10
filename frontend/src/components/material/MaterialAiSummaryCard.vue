<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from "vue";
import DOMPurify from "dompurify";
import MarkdownIt from "markdown-it";
import { PhSparkle } from "@phosphor-icons/vue";

const props = defineProps<{
  summary?: string;
  sourceName?: string;
}>();

const markdown = new MarkdownIt({
  html: false,
  linkify: false,
  typographer: false,
  breaks: false,
});

const sanitizedHtml = computed(() => {
  if (!trimmedSummary.value) return "";

  const rendered = markdown.render(trimmedSummary.value);
  return DOMPurify.sanitize(rendered, {
    ALLOWED_TAGS: [
      "p",
      "h1",
      "h2",
      "h3",
      "strong",
      "em",
      "ul",
      "ol",
      "li",
      "hr",
      "br",
    ],
    ALLOWED_ATTR: [],
  });
});

const trimmedSummary = computed(() => props.summary?.trim() ?? "");
const copied = ref(false);
let copiedTimer: number | undefined;

async function copySummary() {
  if (
    !trimmedSummary.value ||
    typeof navigator === "undefined" ||
    !navigator.clipboard
  ) {
    return;
  }

  try {
    await navigator.clipboard.writeText(trimmedSummary.value);
  } catch {
    return;
  }

  copied.value = true;

  if (copiedTimer) {
    window.clearTimeout(copiedTimer);
  }
  copiedTimer = window.setTimeout(() => {
    copied.value = false;
  }, 1800);
}

onBeforeUnmount(() => {
  if (copiedTimer) {
    window.clearTimeout(copiedTimer);
  }
});
</script>

<template>
  <div
    class="rounded-xl border border-border bg-[#fbfaf8] p-4"
  >
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div class="min-w-0">
        <div class="flex items-center gap-2">
          <span
            class="inline-flex items-center gap-1.5 rounded-full border border-[#ddd7ee] bg-white px-2.5 py-1 text-xs font-semibold text-brand"
          >
            <PhSparkle :size="15" class="text-brand" weight="duotone" />
            AI Summary
          </span>
        </div>
        <p
          v-if="sourceName"
          class="mt-2 truncate text-xs leading-5 text-[#7a7385]"
          :title="sourceName"
        >
          Dirangkum dari: {{ sourceName }}
        </p>
      </div>

      <button
        class="inline-flex items-center rounded-lg border border-border bg-white px-3 py-1.5 text-xs font-medium text-[#5b4b7a] transition hover:border-[#d8d1c5] hover:bg-[#f8f7f4] focus:outline-none focus:ring-2 focus:ring-brand/25 disabled:cursor-not-allowed disabled:opacity-50"
        type="button"
        :disabled="!trimmedSummary"
        @click="copySummary"
      >
        {{ copied ? "Tersalin" : "Salin" }}
      </button>
    </div>

    <div
      v-if="sanitizedHtml"
      class="ai-summary-content mt-4 text-sm leading-7 text-[#4a4356]"
      v-html="sanitizedHtml"
    />
    <p v-else class="mt-4 text-sm leading-6 text-[#7a7385]">
      Rangkuman belum tersedia.
    </p>
  </div>
</template>

<style scoped>
.ai-summary-content :deep(h1) {
  margin: 0.75rem 0 0.5rem;
  color: #171322;
  font-size: 1rem;
  font-weight: 600;
  line-height: 1.5rem;
}

.ai-summary-content :deep(h2),
.ai-summary-content :deep(h3) {
  margin: 0.75rem 0 0.5rem;
  color: #171322;
  font-size: 0.875rem;
  font-weight: 600;
  line-height: 1.5rem;
}

.ai-summary-content :deep(p) {
  margin: 0.5rem 0;
}

.ai-summary-content :deep(strong) {
  color: #2f2938;
  font-weight: 600;
}

.ai-summary-content :deep(em) {
  font-style: italic;
}

.ai-summary-content :deep(ul),
.ai-summary-content :deep(ol) {
  margin: 0.5rem 0;
  padding-left: 1.25rem;
}

.ai-summary-content :deep(ul) {
  list-style-type: disc;
}

.ai-summary-content :deep(ol) {
  list-style-type: decimal;
}

.ai-summary-content :deep(li) {
  margin: 0.25rem 0;
  padding-left: 0.125rem;
}

.ai-summary-content :deep(hr) {
  margin: 0.875rem 0;
  border: 0;
  border-top: 1px solid #ebe7df;
}

.ai-summary-content :deep(:first-child) {
  margin-top: 0;
}

.ai-summary-content :deep(:last-child) {
  margin-bottom: 0;
}
</style>
