<script setup lang="ts">
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { PhArrowLeft, PhClipboardText } from "@phosphor-icons/vue";

const route = useRoute();
const router = useRouter();
const subjectClassId = computed(() => String(route.params.sclId ?? ""));
const assignmentId = computed(() => String(route.params.asgId ?? ""));
</script>

<template>
  <main class="min-h-screen flex-1 px-5 py-6 sm:px-8 lg:px-10">
    <button
      class="mb-6 inline-flex items-center gap-2 rounded-md bg-white px-4 py-2 text-sm font-medium text-[#4f46e5] transition hover:bg-[#eef2ff]"
      type="button"
      @click="router.back()"
    >
      <PhArrowLeft :size="18" />
      Kembali
    </button>

    <section class="soft-card max-w-3xl rounded-md p-6">
      <div class="mb-6 flex items-center gap-4">
        <div
          class="flex h-12 w-12 items-center justify-center rounded-md bg-[#eef2ff] text-[#4f46e5]"
        >
          <PhClipboardText :size="24" weight="duotone" />
        </div>
        <div>
          <p class="text-sm text-[#7a7385]">Assignment</p>
          <h1 class="mt-2 text-3xl font-medium tracking-normal text-[#171322]">
            Detail tugas akan dimuat di sini
          </h1>
        </div>
      </div>

      <div class="mt-6 space-y-4 rounded-md bg-[#fbfaf8] p-4">
        <div>
          <p class="text-sm font-medium text-[#7a7385]">Status Implementasi</p>
          <p class="mt-2 text-sm leading-6 text-[#6b6475]">
            Halaman detail tugas adalah placeholder untuk fase MVP. Fitur yang
            belum diimplementasikan:
          </p>
          <ul
            class="mt-3 list-inside list-disc space-y-1 text-sm text-[#6b6475]"
          >
            <li>Fetch detail tugas dari API</li>
            <li>Tampilkan deskripsi lengkap</li>
            <li>Form submit tugas</li>
            <li>Upload file</li>
            <li>Lihat feedback/penilaian</li>
            <li>Riwayat submission</li>
          </ul>
        </div>

        <div class="mt-4 flex gap-2">
          <span
            class="rounded-full bg-[#e0f2fe] px-3 py-1 text-xs font-medium text-[#0369a1]"
          >
            Assignment ID: {{ assignmentId }}
          </span>
          <span
            class="rounded-full bg-[#f3e8ff] px-3 py-1 text-xs font-medium text-[#7c3aed]"
          >
            SubjectClass ID: {{ subjectClassId }}
          </span>
        </div>
      </div>
    </section>
  </main>
</template>
