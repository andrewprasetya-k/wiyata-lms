<script setup lang="ts">
import { computed, ref } from "vue";
import { useRoute } from "vue-router";
import {
  PhBookOpen,
  PhChatCircleText,
  PhClipboardText,
  PhNewspaper,
  PhNotebook,
} from "@phosphor-icons/vue";

const route = useRoute();
const classId = computed(() => String(route.params.classId ?? ""));
const classTitle = computed(() => String(route.query.title ?? "Detail Kelas"));
const activeTab = ref("materials");

const tabs = [
  {
    key: "materials",
    label: "Materials",
    icon: PhBookOpen,
    title: "Materi kelas akan tampil di sini",
    description:
      "Nantinya bagian ini menampilkan materi belajar berdasarkan subject class dan konteks kelas.",
  },
  {
    key: "assignments",
    label: "Assignments",
    icon: PhClipboardText,
    title: "Tugas kelas akan tampil di sini",
    description:
      "Nantinya bagian ini menampilkan tugas aktif, status submit, deadline, dan hasil penilaian.",
  },
  {
    key: "feed",
    label: "Feed",
    icon: PhNewspaper,
    title: "Diskusi dan pengumuman kelas akan tampil di sini",
    description:
      "Nantinya bagian ini menjadi ruang pengumuman dan aktivitas kelas yang terhubung ke feed backend.",
  },
  {
    key: "notes",
    label: "Notes",
    icon: PhNotebook,
    title: "Catatan pribadi akan tersedia di halaman materi",
    description:
      "Notes belum memakai autosave pada tahap ini dan akan dikaitkan dengan detail material.",
  },
];

const currentTab = computed(
  () => tabs.find((tab) => tab.key === activeTab.value) ?? tabs[0],
);
</script>

<template>
  <main class="min-h-screen flex-1 px-5 py-6 sm:px-8 lg:px-10">
    <header class="mb-6">
      <p class="text-sm text-[#7a7385]">Workspace kelas</p>
      <h1 class="mt-2 text-3xl font-medium tracking-normal text-[#171322]">
        {{ classTitle }}
      </h1>
      <p class="mt-3 max-w-2xl text-sm leading-6 text-[#7a7385]">
        Ruang kelas ini akan menjadi titik masuk untuk materi, tugas, feed, dan
        catatan belajar.
      </p>
    </header>

    <section class="soft-card overflow-hidden rounded-md">
      <div class="bg-white/80 px-4 py-3">
        <div class="flex flex-wrap gap-2">
          <button
            v-for="tab in tabs"
            :key="tab.key"
            class="flex h-10 items-center gap-2 px-4 text-sm transition"
            :class="
              activeTab === tab.key
                ? 'bg-[#eef2ff]/30 font-medium text-[#4f46e5] border-b-2 border-[#4f46e5]'
                : 'text-[#7a7385] hover:bg-[#f8f7f4] hover:text-[#3f3a4a]'
            "
            type="button"
            @click="activeTab = tab.key"
          >
            <component :is="tab.icon" :size="17" />
            {{ tab.label }}
          </button>
        </div>
      </div>

      <div class="grid gap-6 p-6 lg:grid-cols-[1fr_260px]">
        <article class="rounded-md bg-[#fbfaf8] p-6">
          <div
            class="mb-5 flex h-12 w-12 items-center justify-center rounded-md bg-[#eef2ff] text-[#4f46e5]"
          >
            <component :is="currentTab.icon" :size="24" weight="duotone" />
          </div>
          <p class="text-sm font-medium text-[#4f46e5]">
            {{ currentTab.label }}
          </p>
          <h2 class="mt-3 text-2xl font-medium text-[#171322]">
            {{ currentTab.title }}
          </h2>
          <p class="mt-4 max-w-xl text-sm leading-6 text-[#6b6475]">
            {{ currentTab.description }}
          </p>
        </article>

        <aside class="rounded-md border border-[#ebe7df] bg-white p-5">
          <div class="flex items-center gap-3">
            <div
              class="flex h-10 w-10 items-center justify-center rounded-md bg-[#f3ecff] text-[#9d5bd2]"
            >
              <PhChatCircleText :size="20" weight="duotone" />
            </div>
            <div>
              <p class="text-sm font-medium text-[#171322]">Class context</p>
              <p class="text-xs text-[#8b8592]">Data detail belum dimuat</p>
            </div>
          </div>
          <p
            class="mt-5 rounded-md bg-[#fbfaf8] px-4 py-3 text-xs text-[#7a7385]"
          >
            Class ID: {{ classId }}
          </p>
        </aside>
      </div>
    </section>
  </main>
</template>
