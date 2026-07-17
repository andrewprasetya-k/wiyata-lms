<template>
  <EmailVerificationBanner />
  <RouterView />
  <ToastViewport />
  <ConfirmDialog />
</template>

<script setup lang="ts">
import ToastViewport from "./components/common/ToastViewport.vue";
import ConfirmDialog from "./components/common/ConfirmDialog.vue";
import EmailVerificationBanner from "./components/common/EmailVerificationBanner.vue";
import { useAuthContextSync } from "./composables/useAuthContextSync";

useAuthContextSync();
</script>
