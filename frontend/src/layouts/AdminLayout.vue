<script setup lang="ts">
import { computed } from "vue";
import {
  PhBookOpen,
  PhCalendarBlank,
  PhChalkboardTeacher,
  PhChatCircle,
  PhHouse,
  PhStudent,
  PhUsers,
} from "@phosphor-icons/vue";
import { useRoute } from "vue-router";
import Sidebar from "../components/layout/Sidebar.vue";
import { useChatUnreadCount } from "../composables/useChatUnreadCount";

const route = useRoute();
const { unreadCount, badgeLabel } = useChatUnreadCount();

const items = computed(() => [
  { label: "Dashboard", icon: PhHouse, to: "/admin/dashboard" },
  {
    label: "Struktur Akademik",
    icon: PhCalendarBlank,
    to: "/admin/academic-years",
  },
  { label: "Kelas", icon: PhBookOpen, to: "/admin/classes" },
  { label: "Warga Sekolah", icon: PhUsers, to: "/admin/users" },
  { label: "Penempatan Kelas", icon: PhStudent, to: "/admin/enrollments" },
  {
    label: "Penugasan Mengajar",
    icon: PhChalkboardTeacher,
    to: "/admin/subject-classes",
  },
  {
    label: "Chat",
    icon: PhChatCircle,
    to: "/admin/chat",
    badgeCount: unreadCount.value,
    badgeLabel: badgeLabel.value,
    badgeAriaLabel: `${unreadCount.value} chat belum dibaca`,
    emphasized: unreadCount.value > 0 && !route.path.startsWith("/admin/chat"),
  },
]);
</script>

<template>
  <div class="min-h-screen bg-[#f8f7f4] text-[#2f2b3a]">
    <div class="mx-auto flex min-h-screen max-w-360">
      <Sidebar
        class="sticky top-0 h-screen shrink-0"
        label="Navigasi admin sekolah"
        :items="items"
        profile-to="/admin/profile"
      />
      <RouterView />
    </div>
  </div>
</template>
