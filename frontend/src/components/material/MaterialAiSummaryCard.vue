<script setup lang="ts">
import { computed } from "vue";
import DOMPurify from "dompurify";
import MarkdownIt from "markdown-it";
import { PhSparkle } from "@phosphor-icons/vue";

const props = defineProps<{
  summary?: string;
}>();

const markdown = new MarkdownIt({
  html: false,
  linkify: false,
  typographer: false,
  breaks: false,
});

const sanitizedHtml = computed(() => {
  const source = props.summary?.trim() ?? "";
  if (!source) return "";

  const rendered = markdown.render(source);
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
</script>

<template>
  <div
    v-if="sanitizedHtml"
    class="rounded-xl border border-[#ebe7df] bg-[#fbfaf8] p-4"
  >
    <div class="flex items-center gap-2 text-sm font-medium text-[#171322]">
      <PhSparkle :size="17" class="text-[#4f46e5]" weight="duotone" />
      Rangkuman AI
    </div>

    <div
      class="ai-summary-content mt-4 text-sm leading-7 text-[#4a4356]"
      v-html="sanitizedHtml"
    />
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
