<template>
  <RouterView />
  <ToastViewport />
  <ConfirmDialog />
</template>

<script setup lang="ts">
import ToastViewport from "./components/common/ToastViewport.vue";
import ConfirmDialog from "./components/common/ConfirmDialog.vue";
</script>
