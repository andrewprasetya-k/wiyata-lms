<script setup lang="ts">
import { PhSparkle } from '@phosphor-icons/vue'

defineProps<{
  eyebrow: string
  title: string
  description: string
}>()
</script>

<template>
  <section class="flex min-h-screen flex-1 items-center justify-center px-5 py-8">
    <article class="soft-card w-full max-w-2xl rounded-[28px] p-8">
      <div
        class="mb-5 flex h-12 w-12 items-center justify-center rounded-2xl bg-brand-soft text-brand"
      >
        <PhSparkle :size="24" weight="duotone" />
      </div>
      <p class="text-sm font-medium text-brand">{{ eyebrow }}</p>
      <span
        class="mt-3 inline-flex rounded-full bg-brand-soft px-3 py-1 text-xs font-medium text-brand"
      >
        Post-MVP
      </span>
      <h1 class="mt-3 text-3xl font-medium text-foreground">Direncanakan setelah MVP sekolah</h1>
      <h2 class="mt-2 text-xl font-medium text-[#3f3a4a]">{{ title }}</h2>
      <p class="mt-4 max-w-xl text-sm leading-6 text-[#6b6475]">
        {{ description }}
      </p>
    </article>
  </section>
</template>
