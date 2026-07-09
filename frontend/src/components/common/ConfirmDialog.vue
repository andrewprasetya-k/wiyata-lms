<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { PhInfo, PhWarning, PhWarningCircle } from "@phosphor-icons/vue";
import { useConfirmStore } from "../../stores/confirm";

const store = useConfirmStore();
const cancelRef = ref<HTMLButtonElement | null>(null);
const confirmRef = ref<HTMLButtonElement | null>(null);
let previousFocus: Element | null = null;

const variantConfig = computed(() => {
  const v = store.options.variant ?? "default";
  if (v === "danger")
    return {
      icon: PhWarningCircle,
      iconBg: "bg-[#fef2f2]",
      iconColor: "text-[#dc2626]",
      confirmClass:
        "bg-[#dc2626] text-white hover:bg-[#b91c1c] focus-visible:ring-[#dc2626]",
    };
  if (v === "warning")
    return {
      icon: PhWarning,
      iconBg: "bg-[#fff7ed]",
      iconColor: "text-[#ea580c]",
      confirmClass:
        "bg-[#4f46e5] text-white hover:bg-[#4338ca] focus-visible:ring-[#4f46e5]",
    };
  return {
    icon: PhInfo,
    iconBg: "bg-[#eef2ff]",
    iconColor: "text-[#4f46e5]",
    confirmClass:
      "bg-[#4f46e5] text-white hover:bg-[#4338ca] focus-visible:ring-[#4f46e5]",
  };
});

watch(
  () => store.open,
  (isOpen) => {
    if (isOpen) {
      previousFocus = document.activeElement;
      setTimeout(() => cancelRef.value?.focus(), 60);
    } else {
      const el = previousFocus as HTMLElement | null;
      previousFocus = null;
      setTimeout(() => el?.focus?.(), 160);
    }
  },
);

function handleKeydown(e: KeyboardEvent) {
  if (!store.open) return;

  if (e.key === "Escape") {
    e.preventDefault();
    store.dismiss();
    return;
  }

  if (e.key === "Tab") {
    const focusable = [cancelRef.value, confirmRef.value].filter(
      Boolean,
    ) as HTMLButtonElement[];
    if (focusable.length < 2) return;
    const [first, last] = [focusable[0], focusable[focusable.length - 1]];
    if (e.shiftKey && document.activeElement === first) {
      e.preventDefault();
      last.focus();
    } else if (!e.shiftKey && document.activeElement === last) {
      e.preventDefault();
      first.focus();
    }
  }
}

onMounted(() => window.addEventListener("keydown", handleKeydown));
onUnmounted(() => window.removeEventListener("keydown", handleKeydown));
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="store.open"
        class="fixed inset-0 z-60 flex items-end justify-center p-4 sm:items-center"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="
          store.options.title ? 'confirm-dialog-title' : undefined
        "
        :aria-describedby="
          store.options.description ? 'confirm-dialog-desc' : undefined
        "
      >
        <!-- Overlay -->
        <div
          class="absolute inset-0 bg-black/40 backdrop-blur-[2px]"
          aria-hidden="true"
          @click="store.dismiss()"
        />

        <!-- Dialog box -->
        <Transition
          appear
          enter-active-class="transition duration-200 ease-out"
          enter-from-class="opacity-0 translate-y-2 scale-[0.97] sm:translate-y-0"
          enter-to-class="opacity-100 translate-y-0 scale-100"
          leave-active-class="transition duration-150 ease-in"
          leave-from-class="opacity-100 translate-y-0 scale-100"
          leave-to-class="opacity-0 translate-y-2 scale-[0.97] sm:translate-y-0"
        >
          <div
            v-if="store.open"
            class="relative w-full max-w-sm rounded-2xl bg-white p-6 shadow-2xl shadow-black/20 ring-1 ring-black/5"
          >
            <!-- Icon -->
            <div
              class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl"
              :class="variantConfig.iconBg"
            >
              <component
                :is="variantConfig.icon"
                :size="24"
                :class="variantConfig.iconColor"
                weight="duotone"
              />
            </div>

            <!-- Title -->
            <h2
              id="confirm-dialog-title"
              class="mt-4 text-center text-base font-semibold text-[#171322]"
            >
              {{ store.options.title }}
            </h2>

            <!-- Description -->
            <p
              v-if="store.options.description"
              id="confirm-dialog-desc"
              class="mt-2 text-center text-sm leading-6 text-[#6b7280]"
            >
              {{ store.options.description }}
            </p>

            <!-- Buttons -->
            <div
              class="mt-6 flex flex-col-reverse gap-2.5 sm:flex-row sm:justify-center"
            >
              <button
                ref="cancelRef"
                type="button"
                class="flex-1 rounded-xl border border-[#ebe7df] bg-white px-4 py-2.5 text-sm font-medium text-[#374151] transition hover:border-[#d1d5db] hover:bg-[#f9f8f7] focus:outline-none focus-visible:ring-2 focus-visible:ring-[#4f46e5] focus-visible:ring-offset-2 sm:flex-none sm:min-w-22"
                @click="store.dismiss()"
              >
                {{ store.options.cancelLabel ?? "Batal" }}
              </button>
              <button
                ref="confirmRef"
                type="button"
                class="flex-1 rounded-xl px-4 py-2.5 text-sm font-medium transition focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 sm:flex-none sm:min-w-22"
                :class="variantConfig.confirmClass"
                @click="store.accept()"
              >
                {{ store.options.confirmLabel ?? "Konfirmasi" }}
              </button>
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>
