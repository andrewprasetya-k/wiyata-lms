<script setup lang="ts">
import { PhBuildings, PhHouse, PhUsers } from '@phosphor-icons/vue'
import SlimSidebar from '../components/layout/Sidebar.vue'

const items = [
  { label: 'Dashboard', icon: PhHouse, to: '/superadmin/dashboard' },
  { label: 'Schools', icon: PhBuildings, to: '/superadmin/schools' },
  { label: 'Users', icon: PhUsers, to: '/superadmin/users' },
]
</script>

<template>
  <div class="min-h-screen bg-[#f8f7f4] text-[#2f2b3a]">
    <div class="mx-auto flex min-h-screen max-w-360">
      <SlimSidebar
        class="sticky top-0 h-screen shrink-0"
        label="Super admin navigation"
        :items="items"
        profile-to="/superadmin/profile"
      />
      <RouterView />
    </div>
  </div>
</template>
