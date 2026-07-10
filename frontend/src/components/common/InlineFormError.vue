<script setup lang="ts">
defineProps<{
  message: string;
}>();
</script>

<template>
  <p v-if="message" class="text-xs font-medium text-[#b42318]">
    {{ message }}
  </p>
</template>
