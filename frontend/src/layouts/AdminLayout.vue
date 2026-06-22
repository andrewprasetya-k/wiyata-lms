<script setup lang="ts">
import {
  PhBookOpen,
  PhCalendarBlank,
  PhChalkboardTeacher,
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
  { label: "Student Assignment", icon: PhStudent, to: "/admin/enrollments" },
  {
    label: "Teacher Assignment",
    icon: PhChalkboardTeacher,
    to: "/admin/subject-classes",
  },
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
