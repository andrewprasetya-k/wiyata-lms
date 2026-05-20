<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";
import {
  PhBookOpen,
  PhChatCircleText,
  PhClipboardText,
  PhFileText,
  PhNewspaper,
  PhNotebook,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { getClassMaterials } from "../../services/classWorkspace";
import type {
  ClassHeader,
  MaterialItem,
  SubjectClassItem,
} from "../../types/classWorkspace";

const route = useRoute();
const classId = computed(() => String(route.params.classId ?? ""));
const classInfo = ref<ClassHeader | null>(null);
const subjects = ref<SubjectClassItem[]>([]);
const materials = ref<MaterialItem[]>([]);
const isMaterialsLoading = ref(true);
const materialsError = ref("");
const activeTab = ref("materials");

const classTitle = computed(
  () => classInfo.value?.classTitle || String(route.query.title ?? "Detail Kelas"),
);

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

async function loadMaterials() {
  if (!classId.value) {
    isMaterialsLoading.value = false;
    materialsError.value = "Class ID tidak tersedia.";
    return;
  }

  isMaterialsLoading.value = true;
  materialsError.value = "";

  try {
    const data = await getClassMaterials(classId.value);
    classInfo.value = data.classInfo;
    subjects.value = data.subjects;
    materials.value = data.materials;
  } catch (error) {
    console.error('[StudentClassDetail] Error loading materials:', error)
    materialsError.value =
      "Materi kelas belum bisa dimuat. Periksa koneksi atau coba lagi nanti.";
  } finally {
    isMaterialsLoading.value = false;
  }
}

onMounted(loadMaterials);
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
          <template v-if="activeTab === 'materials'">
            <p class="text-sm font-medium text-[#4f46e5]">Materials</p>
            <h2 class="mt-3 text-2xl font-medium text-[#171322]">
              Materi kelas
            </h2>

            <div v-if="isMaterialsLoading" class="mt-6 space-y-3">
              <div
                v-for="item in 3"
                :key="item"
                class="h-20 animate-pulse rounded-md bg-white"
              />
            </div>

            <div v-else-if="materialsError" class="mt-6 rounded-md bg-white p-4">
              <div class="flex items-start gap-3">
                <PhWarningCircle
                  :size="22"
                  class="mt-0.5 text-[#f2756a]"
                  weight="duotone"
                />
                <div>
                  <p class="text-sm font-medium text-[#171322]">
                    Tidak bisa memuat materi
                  </p>
                  <p class="mt-1 text-sm leading-6 text-[#7a7385]">
                    {{ materialsError }}
                  </p>
                  <button
                    class="mt-4 rounded-md bg-[#4f46e5] px-4 py-2 text-sm font-medium text-white"
                    type="button"
                    @click="loadMaterials"
                  >
                    Coba lagi
                  </button>
                </div>
              </div>
            </div>

            <div v-else-if="subjects.length === 0" class="mt-6 rounded-md bg-white p-4">
              <p class="text-sm font-medium text-[#171322]">
                Subject class belum tersedia
              </p>
              <p class="mt-2 text-sm leading-6 text-[#7a7385]">
                Materi membutuhkan subjectClassId. Kelas ini belum memiliki subject assignment yang
                bisa dipakai untuk mengambil materi.
              </p>
            </div>

            <div v-else-if="materials.length === 0" class="mt-6 rounded-md bg-white p-4">
              <p class="text-sm font-medium text-[#171322]">Belum ada materi</p>
              <p class="mt-2 text-sm leading-6 text-[#7a7385]">
                Materi akan tampil setelah guru menambahkan konten pada subject class kelas ini.
              </p>
            </div>

            <div v-else class="mt-6 space-y-3">
              <article
                v-for="material in materials"
                :key="material.materialId"
                class="rounded-md bg-white p-4"
              >
                <div class="flex items-start gap-3">
                  <div
                    class="flex h-10 w-10 shrink-0 items-center justify-center rounded-md bg-[#eef2ff] text-[#4f46e5]"
                  >
                    <PhFileText :size="20" weight="duotone" />
                  </div>
                  <div class="min-w-0 flex-1">
                    <div class="flex flex-wrap items-center gap-2">
                      <p class="text-sm font-medium text-[#171322]">
                        {{ material.materialTitle }}
                      </p>
                      <span
                        class="rounded-full bg-[#f3ecff] px-2 py-0.5 text-[10px] uppercase tracking-wide text-[#9d5bd2]"
                      >
                        {{ material.materialType }}
                      </span>
                    </div>
                    <p v-if="material.subjectName" class="mt-1 text-xs text-[#8b8592]">
                      {{ material.subjectName }}
                    </p>
                    <p
                      v-if="material.materialDesc"
                      class="mt-2 line-clamp-2 text-sm leading-6 text-[#6b6475]"
                    >
                      {{ material.materialDesc }}
                    </p>
                    <p class="mt-2 text-xs text-[#a09aa8]">
                      {{ material.creatorName || "Creator tidak tersedia" }} ·
                      {{ material.createdAt }}
                    </p>
                  </div>
                </div>
              </article>
            </div>
          </template>

          <template v-else>
            <p class="text-sm font-medium text-[#4f46e5]">
              {{ currentTab.label }}
            </p>
            <h2 class="mt-3 text-2xl font-medium text-[#171322]">
              {{ currentTab.title }}
            </h2>
            <p class="mt-4 max-w-xl text-sm leading-6 text-[#6b6475]">
              {{ currentTab.description }}
            </p>
          </template>
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
              <p class="text-xs text-[#8b8592]">
                {{ subjects.length }} subject class tersedia
              </p>
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
