<script setup lang="ts">
import { computed } from "vue";
import {
  PhBookOpen,
  PhCalendarBlank,
  PhChatCircle,
  PhHouse,
  PhMegaphone,
  PhTray,
} from "@phosphor-icons/vue";
import { useRoute } from "vue-router";
import SlimSidebar from "../components/layout/Sidebar.vue";
import { useChatUnreadCount } from "../composables/useChatUnreadCount";

const route = useRoute();
const { unreadCount, badgeLabel } = useChatUnreadCount();

const items = computed(() => [
  { label: "Dashboard", icon: PhHouse, to: "/teacher/dashboard" },
  { label: "Mata Pelajaran", icon: PhBookOpen, to: "/teacher/subjects" },
  { label: "Tugas", icon: PhCalendarBlank, to: "/teacher/assignments" },
  { label: "Pengumpulan", icon: PhTray, to: "/teacher/submissions" },
  { label: "Feed", icon: PhMegaphone, to: "/teacher/feed" },
  {
    label: "Chat",
    icon: PhChatCircle,
    to: "/teacher/chat",
    badgeCount: unreadCount.value,
    badgeLabel: badgeLabel.value,
    badgeAriaLabel: `${unreadCount.value} chat belum dibaca`,
    emphasized: unreadCount.value > 0 && !route.path.startsWith("/teacher/chat"),
  },
]);
</script>

<template>
  <div class="min-h-screen bg-[#f8f7f4] text-[#2f2b3a]">
    <div class="mx-auto flex min-h-screen max-w-360">
      <SlimSidebar
        class="sticky top-0 h-screen shrink-0"
        label="Navigasi guru"
        :items="items"
        profile-to="/teacher/profile"
      />
      <RouterView />
    </div>
  </div>
</template>
