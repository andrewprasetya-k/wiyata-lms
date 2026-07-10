<script setup lang="ts">
import {
  PhCheckCircle,
  PhInfo,
  PhWarningCircle,
  PhX,
} from "@phosphor-icons/vue";
import { computed } from "vue";
import { useToastStore, type ToastItem } from "../../stores/toast";

const toastStore = useToastStore();

const variantMeta = computed(() => ({
  success: {
    icon: PhCheckCircle,
    className: "border-[#BBF7D0] bg-[#ECFDF5] text-[#065F46]",
    iconClass: "text-[#059669]",
  },
  error: {
    icon: PhWarningCircle,
    className: "border-[#FECACA] bg-[#FEF2F2] text-[#991B1B]",
    iconClass: "text-[#DC2626]",
  },
  info: {
    icon: PhInfo,
    className: "border-[#C7D2FE] bg-[#EEF2FF] text-[#3730A3]",
    iconClass: "text-brand",
  },
}));

function getMeta(toast: ToastItem) {
  return variantMeta.value[toast.variant];
}
</script>

<template>
  <div
    aria-live="polite"
    aria-relevant="additions removals"
    class="pointer-events-none fixed right-4 top-4 z-50 flex w-[min(24rem,calc(100vw-2rem))] flex-col gap-3"
  >
    <div
      v-for="toast in toastStore.toasts"
      :key="toast.id"
      role="status"
      class="pointer-events-auto flex items-start gap-3 rounded-[14px] border px-4 py-3 text-sm shadow-lg shadow-black/5"
      :class="getMeta(toast).className"
    >
      <component
        :is="getMeta(toast).icon"
        :size="20"
        class="mt-0.5 shrink-0"
        :class="getMeta(toast).iconClass"
        weight="duotone"
      />
      <p class="min-w-0 flex-1 leading-6">
        {{ toast.message }}
      </p>
      <button
        type="button"
        class="mt-0.5 rounded-md p-1 text-current opacity-70 transition hover:bg-white/60 hover:opacity-100"
        aria-label="Tutup notifikasi"
        @click="toastStore.remove(toast.id)"
      >
        <PhX :size="14" />
      </button>
    </div>
  </div>
</template>
