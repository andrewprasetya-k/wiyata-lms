<script setup lang="ts">
withDefaults(defineProps<{
  label: string
  overdue?: boolean
}>(), { overdue: false })
</script>

<template>
  <span
    class="inline-flex rounded-full px-2 py-0.5 text-[11px] font-medium"
    :class="overdue ? 'bg-danger-soft text-danger' : 'bg-brand-soft text-brand'"
  >
    {{ label }}
  </span>
</template>
