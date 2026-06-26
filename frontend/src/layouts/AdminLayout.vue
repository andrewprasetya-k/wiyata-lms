<script setup lang="ts">
import {
  PhBookOpen,
  PhCalendarBlank,
  PhChalkboardTeacher,
  PhChatCircle,
  PhHouse,
  PhStudent,
  PhUsers,
} from "@phosphor-icons/vue";
import Sidebar from "../components/layout/Sidebar.vue";

const items = [
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
  { label: "Chat", icon: PhChatCircle, to: "/admin/chat" },
];
</script>

<template>
  <div class="min-h-screen bg-[#f8f7f4] text-[#2f2b3a]">
    <div class="mx-auto flex min-h-screen max-w-360">
      <Sidebar
        class="sticky top-0 h-screen shrink-0"
        label="Admin navigation"
        :items="items"
        profile-to="/admin/profile"
      />
      <RouterView />
    </div>
  </div>
</template>
