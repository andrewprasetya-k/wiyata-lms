<script setup lang="ts">
import { computed } from "vue";
import {
  PhCalendarBlank,
  PhCalendarCheck,
  PhBell,
  PhBookOpen,
  PhChartBar,
  PhChatCircle,
  PhHouse,
  PhMegaphone,
  PhNotebook,
} from "@phosphor-icons/vue";
import { useRoute } from "vue-router";
import SlimSidebar from "../components/layout/Sidebar.vue";
import { useChatUnreadCount } from "../composables/useChatUnreadCount";
import { useFeedUnreadCount } from "../composables/useFeedUnreadCount";
import { useNotificationUnreadCount } from "../composables/useNotificationUnreadCount";
import { useAuthStore } from "../stores/auth";

const auth = useAuthStore();
const route = useRoute();
const { unreadCount, badgeLabel } = useChatUnreadCount();
const { unreadCount: feedUnreadCount, badgeLabel: feedBadgeLabel } =
  useFeedUnreadCount();
const {
  unreadCount: notificationUnreadCount,
  badgeLabel: notificationBadgeLabel,
} = useNotificationUnreadCount();

const items = computed(() => [
  { label: "Dashboard", icon: PhHouse, to: "/student/dashboard" },
  { label: "Mata Pelajaran", icon: PhBookOpen, to: "/student/subjects" },
  {
    label: "Feed",
    icon: PhMegaphone,
    to: "/student/feed",
    badgeCount: feedUnreadCount.value,
    badgeLabel: feedBadgeLabel.value,
    badgeAriaLabel: `${feedUnreadCount.value} feed belum dibaca`,
    emphasized: feedUnreadCount.value > 0 && !route.path.startsWith("/student/feed"),
  },
  { label: "Tugas", icon: PhCalendarBlank, to: "/student/assignments" },
  { label: "Aktivitas", icon: PhCalendarCheck, to: "/student/activity" },
  {
    label: "Notifikasi",
    icon: PhBell,
    to: "/student/notifications",
    badgeCount: notificationUnreadCount.value,
    badgeLabel: notificationBadgeLabel.value,
    badgeAriaLabel: `${notificationUnreadCount.value} notifikasi belum dibaca`,
    emphasized:
      notificationUnreadCount.value > 0 &&
      !route.path.startsWith("/student/notifications"),
  },
  { label: "Nilai", icon: PhChartBar, to: "/student/grades" },
  {
    label: "Chat",
    icon: PhChatCircle,
    to: "/student/chat",
    badgeCount: unreadCount.value,
    badgeLabel: badgeLabel.value,
    badgeAriaLabel: `${unreadCount.value} chat belum dibaca`,
    emphasized: unreadCount.value > 0 && !route.path.startsWith("/student/chat"),
  },
  { label: "Catatan", icon: PhNotebook, to: "/student/notes" },
]);
</script>

<template>
  <div class="min-h-screen bg-background text-[#2f2b3a]">
    <div class="mx-auto flex min-h-screen max-w-360">
      <SlimSidebar
        class="sticky top-0 h-screen shrink-0"
        label="Navigasi siswa"
        :items="items"
        profile-to="/student/profile"
      />
      <RouterView :key="auth.contextVersion" />
    </div>
  </div>
</template>
