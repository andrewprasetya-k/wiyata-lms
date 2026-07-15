<script setup lang="ts">
import { ref } from "vue";
import { motion, useScroll, useTransform } from "motion-v";

interface PreviewStep {
  label: string;
  caption: string;
  url: string;
  image: { src: string; width: number; height: number };
}

const props = defineProps<{
  step: PreviewStep;
  index: number;
  maxW: string;
}>();

// Progress mengikuti perjalanan step ini sendiri melintasi viewport —
// bukan progress halaman, supaya tiap step punya entrance/recede sendiri
// yang selaras dengan durasi "dwell"-nya di posisi sticky.
const rootEl = ref<HTMLElement | null>(null);
const { scrollYProgress } = useScroll({
  target: rootEl,
  offset: ["start end", "end start"],
});

// 0 → 0.15: masuk (fade + scale up tipis)
// 0.15 → 0.85: dwell, diam di posisi natural
// 0.85 → 1: mulai tertutup step berikutnya — mengecil & meredup sedikit
const visualOpacity = useTransform(
  scrollYProgress,
  [0, 0.15, 0.85, 1],
  [0, 1, 1, 0.85],
);
const visualScale = useTransform(
  scrollYProgress,
  [0, 0.15, 0.85, 1],
  [0.96, 1, 1, 0.96],
);
</script>

<template>
  <div ref="rootEl" class="mx-auto w-full px-6 lg:px-8">
    <!-- Teks -->
    <motion.div
      class="mx-auto text-center"
      :class="maxW"
      :initial="{ opacity: 0, y: 16 }"
      :while-in-view="{ opacity: 1, y: 0 }"
      :viewport="{ once: true, margin: '-80px' }"
      :transition="{ duration: 0.5, ease: 'easeOut' }"
    >
      <p class="text-xs font-semibold uppercase tracking-widest text-brand">
        {{ String(props.index + 1).padStart(2, "0") }} — {{ props.step.label }}
      </p>
      <p class="mx-auto mt-3 max-w-xl text-base leading-7 text-muted">
        {{ props.step.caption }}
      </p>
    </motion.div>

    <!-- Visual -->
    <motion.div
      class="mx-auto mt-8"
      :class="maxW"
      :style="{ opacity: visualOpacity, scale: visualScale }"
    >
      <div
        class="overflow-hidden rounded-2xl border border-border bg-surface shadow-[0_28px_90px_-30px_rgba(79,70,229,0.28)]"
      >
        <div class="flex items-center gap-2 bg-[#faf9f7] px-5 py-3">
          <span class="h-3 w-3 rounded-full bg-[#fca5a5]" />
          <span class="h-3 w-3 rounded-full bg-[#fcd34d]" />
          <span class="h-3 w-3 rounded-full bg-[#86efac]" />
          <div
            class="ml-3 flex h-6 max-w-xs flex-1 items-center rounded-md bg-[#f0ece5] px-3 text-xs text-muted"
          >
            {{ props.step.url }}
          </div>
        </div>

        <img
          :src="props.step.image.src"
          :alt="props.step.caption"
          class="h-auto w-full"
          :width="props.step.image.width"
          :height="props.step.image.height"
        />
      </div>
    </motion.div>
  </div>
</template>
