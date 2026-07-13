<script setup lang="ts">
import { PhCaretLeft, PhCaretRight } from "@phosphor-icons/vue";

const props = defineProps<{
  page: number;
  totalPages: number;
  totalItems?: number;
  limit?: number;
}>();

const emit = defineEmits<{
  (e: "change", page: number): void;
}>();

function from() {
  if (!props.totalItems || !props.limit) return null;
  return (props.page - 1) * props.limit + 1;
}

function to() {
  if (!props.totalItems || !props.limit) return null;
  return Math.min(props.page * props.limit, props.totalItems);
}
</script>

<template>
  <div
    v-if="totalPages > 1"
    class="flex flex-wrap items-center justify-between gap-3 border-t border-border pt-4"
  >
    <p v-if="totalItems && limit" class="text-xs text-muted">
      Menampilkan {{ from() }}–{{ to() }} dari {{ totalItems }} data
    </p>
    <p v-else class="text-xs text-muted">
      Halaman {{ page }} dari {{ totalPages }}
    </p>

    <div class="flex items-center gap-2">
      <button
        type="button"
        class="inline-flex h-8 w-8 items-center justify-center rounded-lg border border-border bg-surface text-foreground-secondary transition hover:border-brand hover:text-brand disabled:cursor-not-allowed disabled:opacity-40"
        :disabled="page <= 1"
        aria-label="Halaman sebelumnya"
        @click="emit('change', page - 1)"
      >
        <PhCaretLeft :size="14" weight="bold" />
      </button>
      <span class="text-xs font-medium text-foreground-secondary">
        {{ page }} / {{ totalPages }}
      </span>
      <button
        type="button"
        class="inline-flex h-8 w-8 items-center justify-center rounded-lg border border-border bg-surface text-foreground-secondary transition hover:border-brand hover:text-brand disabled:cursor-not-allowed disabled:opacity-40"
        :disabled="page >= totalPages"
        aria-label="Halaman berikutnya"
        @click="emit('change', page + 1)"
      >
        <PhCaretRight :size="14" weight="bold" />
      </button>
    </div>
  </div>
</template>
