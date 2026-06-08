<script setup lang="ts">
import {
  PhBookOpen,
  PhCalendarBlank,
  PhChatCircle,
  PhHouse,
  PhMegaphone,
  PhPlusCircle,
  PhTray,
  PhUserCircle,
} from '@phosphor-icons/vue'
import SlimSidebar from '../components/layout/Sidebar.vue'

const items = [
  { label: 'Dashboard', icon: PhHouse, to: '/teacher/dashboard' },
  { label: 'Subjects', icon: PhBookOpen, to: '/teacher/subjects' },
  { label: 'Assignments', icon: PhCalendarBlank, to: '/teacher/assignments' },
  { label: 'Submissions', icon: PhTray, to: '/teacher/submissions' },
  { label: 'Create', icon: PhPlusCircle, to: '/teacher/create' },
  { label: 'Feed', icon: PhMegaphone, to: '/teacher/feed' },
  { label: 'Chat', icon: PhChatCircle, to: '/teacher/chat', hasDot: true },
  { label: 'Profile', icon: PhUserCircle, to: '/teacher/profile' },
]
</script>

<template>
  <div class="min-h-screen bg-[#f8f7f4] text-[#2f2b3a]">
    <div class="mx-auto flex min-h-screen max-w-360">
      <SlimSidebar class="sticky top-0 h-screen shrink-0" label="Teacher navigation" :items="items" />
      <RouterView />
    </div>
  </div>
</template>
