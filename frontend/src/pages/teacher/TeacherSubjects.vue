<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import {
  PhArrowRight,
  PhBookOpen,
  PhClipboardText,
  PhFileText,
  PhUsersThree,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { getMyTeachingSubjectClasses } from "../../services/teacherSubjects";
import type { TeacherSubjectClass } from "../../types/teacherSubjects";
import { getSubjectColor } from "../../utils/color";

const loading = ref(false);
const errorMessage = ref("");
const subjects = ref<TeacherSubjectClass[]>([]);

const hasSubjects = computed(() => subjects.value.length > 0);

async function loadSubjects() {
  loading.value = true;
  errorMessage.value = "";
  try {
    subjects.value = await getMyTeachingSubjectClasses();
  } catch {
    errorMessage.value =
      "Subject yang diajar belum bisa dimuat. Coba lagi beberapa saat.";
  } finally {
    loading.value = false;
  }
}

onMounted(loadSubjects);
</script>

<template>
  <main class="min-h-screen flex-1 px-5 py-8 md:px-8 lg:px-10">
    <section class="mx-auto flex max-w-7xl flex-col gap-6">
      <header
        class="rounded-4xl bg-[#f0e9dd] px-6 py-7 shadow-sm ring-1 ring-black/5 md:px-8"
      >
        <p class="text-sm font-medium text-[#8a6d3b]">Teaching workspace</p>
        <h1 class="mt-3 text-3xl font-medium text-[#171322] md:text-4xl">
          Subject yang diajar
        </h1>
        <p class="mt-3 max-w-2xl text-sm leading-6 text-[#6b6475]">
          Pilih subject class untuk mengelola materi, tugas, dan melihat status
          submission siswa. Data di halaman ini berasal dari subject class yang
          terhubung dengan akun guru saat ini.
        </p>
      </header>

      <section
        v-if="loading"
        class="rounded-[28px] bg-white p-6 shadow-sm ring-1 ring-black/5"
      >
        <p class="text-sm text-[#6b6475]">Memuat subject yang diajar...</p>
      </section>

      <section
        v-else-if="errorMessage"
        class="rounded-[28px] bg-white p-6 shadow-sm ring-1 ring-black/5"
      >
        <div
          class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"
        >
          <div class="flex items-start gap-3">
            <PhWarningCircle
              :size="24"
              class="mt-0.5 text-[#e58f86]"
              weight="duotone"
            />
            <div>
              <h2 class="text-lg font-medium text-[#171322]">
                Gagal memuat subject
              </h2>
              <p class="mt-2 text-sm leading-6 text-[#6b6475]">
                {{ errorMessage }}
              </p>
            </div>
          </div>
          <button
            class="rounded-2xl bg-[#171322] px-4 py-3 text-sm font-medium text-white"
            type="button"
            @click="loadSubjects"
          >
            Coba lagi
          </button>
        </div>
      </section>

      <section
        v-else-if="!hasSubjects"
        class="rounded-[28px] bg-white p-6 shadow-sm ring-1 ring-black/5"
      >
        <h2 class="text-lg font-medium text-[#171322]">
          Belum ada subject yang diajar
        </h2>
        <p class="mt-2 max-w-2xl text-sm leading-6 text-[#6b6475]">
          Akun guru ini belum terhubung ke subject class pada school aktif.
          Admin sekolah perlu menambahkan subject class dan mengassign guru
          terlebih dahulu.
        </p>
      </section>

      <section v-else class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        <RouterLink
          v-for="subject in subjects"
          :key="subject.subjectClassId"
          :to="`/teacher/subjects/${subject.subjectClassId}`"
          class="group rounded-[30px] bg-white p-5 shadow-sm ring-1 ring-black/5 transition hover:-translate-y-0.5 hover:shadow-md"
        >
          <div class="flex items-start justify-between gap-4">
            <div
              class="flex h-13 w-13 items-center justify-center rounded-2xl text-white"
              :style="{
                backgroundColor: getSubjectColor(subject.subjectClassId),
              }"
            >
              <PhBookOpen :size="26" weight="duotone" />
            </div>
            <PhArrowRight
              :size="20"
              class="text-[#b5afbf] transition group-hover:translate-x-1 group-hover:text-[#4f46e5]"
            />
          </div>

          <p class="mt-5 text-sm text-[#8a8494]">{{ subject.className }}</p>
          <h2 class="mt-1 text-2xl font-medium text-[#171322]">
            {{ subject.subjectName }}
          </h2>
          <p
            v-if="subject.subjectCode || subject.classCode"
            class="mt-2 text-sm text-[#6b6475]"
          >
            {{
              [subject.subjectCode, subject.classCode]
                .filter(Boolean)
                .join(" / ")
            }}
          </p>

          <div class="mt-6 grid grid-cols-2 gap-3 text-sm">
            <div class="rounded-2xl bg-[#faf8f4] p-3">
              <div class="flex items-center gap-2 text-[#8a8494]">
                <PhUsersThree :size="16" weight="duotone" />
                Siswa
              </div>
              <p class="mt-1 font-medium text-[#171322]">
                {{ subject.studentCount }}
              </p>
            </div>
            <div class="rounded-2xl bg-[#faf8f4] p-3">
              <div class="flex items-center gap-2 text-[#8a8494]">
                <PhFileText :size="16" weight="duotone" />
                Materi
              </div>
              <p class="mt-1 font-medium text-[#171322]">
                {{ subject.materialCount }}
              </p>
            </div>
            <div class="rounded-2xl bg-[#faf8f4] p-3">
              <div class="flex items-center gap-2 text-[#8a8494]">
                <PhClipboardText :size="16" weight="duotone" />
                Tugas
              </div>
              <p class="mt-1 font-medium text-[#171322]">
                {{ subject.assignmentCount }}
              </p>
            </div>
            <div class="rounded-2xl bg-[#faf8f4] p-3">
              <div class="flex items-center gap-2 text-[#8a8494]">
                <PhWarningCircle :size="16" weight="duotone" />
                Review
              </div>
              <p class="mt-1 font-medium text-[#171322]">
                {{ subject.pendingSubmissions }}
              </p>
            </div>
          </div>
        </RouterLink>
      </section>
    </section>
  </main>
</template>
