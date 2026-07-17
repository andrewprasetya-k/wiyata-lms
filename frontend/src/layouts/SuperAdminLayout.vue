<script setup lang="ts">
import { PhBuildings, PhHouse, PhUsers } from "@phosphor-icons/vue";
import SlimSidebar from "../components/layout/Sidebar.vue";
import { useAuthStore } from "../stores/auth";

const auth = useAuthStore();
const items = [
  { label: "Dashboard", icon: PhHouse, to: "/superadmin/dashboard" },
  { label: "Sekolah", icon: PhBuildings, to: "/superadmin/schools" },
  { label: "Akun Global", icon: PhUsers, to: "/superadmin/users" },
];
</script>

<template>
  <div class="min-h-screen bg-background text-foreground">
    <div class="mx-auto flex min-h-screen max-w-360">
      <SlimSidebar
        class="sticky top-0 h-screen shrink-0"
        label="Navigasi Super Admin Platform Wiyata"
        :items="items"
        profile-to="/superadmin/profile"
      />
      <RouterView :key="auth.contextVersion" />
    </div>
  </div>
</template>
