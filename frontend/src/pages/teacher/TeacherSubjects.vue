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
import { resolveSubjectColor } from "../../utils/color";

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
      "Mata pelajaran yang diajar belum bisa dimuat. Coba lagi beberapa saat.";
  } finally {
    loading.value = false;
  }
}

onMounted(loadSubjects);
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-[#f8f7f4]">
    <header class="border-b border-[#ebe7df] bg-white">
      <div class="px-5 py-5 sm:px-6 lg:px-8">
        <h1 class="text-2xl font-semibold text-[#171322] sm:text-3xl">
          Mata Pelajaran yang Diajar
        </h1>
        <p class="mt-2 max-w-3xl text-sm leading-6 text-[#6b7280]">
          Pilih mata pelajaran untuk mengelola materi, tugas, dan pengumpulan
          siswa.
        </p>
      </div>
    </header>

    <section
      class="mx-auto max-w-screen min-w-0 px-5 py-5 sm:px-6 lg:px-8 lg:py-6"
    >
      <section v-if="loading" class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
        <div
          v-for="item in 6"
          :key="item"
          class="h-56 animate-pulse rounded-xl border border-[#ebe7df] bg-white"
        />
      </section>

      <section
        v-else-if="errorMessage"
        class="flex min-h-[55vh] items-center justify-center"
      >
        <article
          class="w-full max-w-xl rounded-xl border border-[#f1d6d3] bg-white p-6"
        >
          <div class="flex items-start gap-3">
            <div
              class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-[#fff1f0] text-[#dc2626]"
            >
              <PhWarningCircle :size="22" weight="duotone" />
            </div>
            <div class="min-w-0">
              <h2 class="text-base font-medium text-[#171322]">
                Mata pelajaran tidak dapat dimuat
              </h2>
              <p class="mt-1 text-sm leading-6 text-[#7a7385]">
                {{ errorMessage }}
              </p>
              <button
                class="mt-4 rounded-lg bg-[#4f46e5] px-4 py-2 text-sm font-medium text-white transition hover:bg-[#4338ca]"
                type="button"
                @click="loadSubjects"
              >
                Coba lagi
              </button>
            </div>
          </div>
        </article>
      </section>

      <section
        v-else-if="!hasSubjects"
        class="flex min-h-[55vh] items-center justify-center"
      >
        <article
          class="w-full max-w-xl rounded-xl border border-[#ebe7df] bg-white p-8 text-center"
        >
          <div
            class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-[#eef2ff] text-[#4f46e5]"
          >
            <PhBookOpen class="h-6 w-6" weight="duotone" />
          </div>
          <h2 class="mt-3 text-base font-semibold text-[#171322]">
            Belum ada mata pelajaran yang diajar
          </h2>
          <p class="mx-auto mt-2 max-w-md text-sm leading-6 text-[#6b7280]">
            Admin sekolah perlu menugaskan guru ke mata pelajaran dan kelas
            aktif terlebih dahulu.
          </p>
        </article>
      </section>

      <section v-else>
        <div class="mb-4 min-w-0">
          <h2 class="text-sm font-medium text-[#171322]">
            Daftar mata pelajaran
          </h2>
          <p class="mt-1 text-xs text-[#7a7385] sm:text-sm">
            {{ subjects.length }} mata pelajaran tersedia.
          </p>
        </div>

        <div class="grid min-w-0 gap-3 md:grid-cols-2 xl:grid-cols-3">
          <RouterLink
            v-for="subject in subjects"
            :key="subject.subjectClassId"
            :to="`/teacher/subjects/${subject.subjectClassId}`"
            class="group min-w-0 overflow-hidden rounded-xl border border-[#ebe7df] bg-white transition hover:-translate-y-0.5 hover:shadow-[0_18px_40px_rgba(66,55,40,0.08)]"
          >
            <div
              class="h-1.5 w-full"
              :style="{
                backgroundColor: resolveSubjectColor(subject),
              }"
            />
            <div class="p-4 sm:p-5">
              <div class="flex min-w-0 items-start gap-3">
                <div
                  class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg text-white"
                  :style="{
                    backgroundColor: resolveSubjectColor(subject),
                  }"
                >
                  <PhBookOpen :size="20" weight="duotone" />
                </div>
                <div class="min-w-0 flex-1">
                  <p class="truncate text-xs text-[#7a7385]">
                    {{ subject.className || subject.classCode || "Kelas" }}
                  </p>
                  <h2
                    class="mt-1 line-clamp-2 wrap-break-word text-base font-medium text-[#171322]"
                  >
                    {{ subject.subjectName || "Mata pelajaran" }}
                  </h2>
                  <p
                    v-if="subject.subjectCode || subject.classCode"
                    class="mt-1 truncate text-xs text-[#8a8494]"
                  >
                    {{
                      [subject.subjectCode, subject.classCode]
                        .filter(Boolean)
                        .join(" · ")
                    }}
                  </p>
                </div>
                <PhArrowRight
                  :size="17"
                  class="mt-1 shrink-0 text-[#a09aa8] transition group-hover:translate-x-0.5 group-hover:text-[#4f46e5]"
                />
              </div>

              <dl class="mt-4 grid grid-cols-2 gap-2 text-xs">
                <div class="rounded-lg bg-[#fbfaf8] p-3">
                  <dt class="flex items-center gap-1.5 text-[#8a8494]">
                    <PhUsersThree :size="15" weight="duotone" />
                    Siswa
                  </dt>
                  <dd class="mt-1 text-base font-medium text-[#171322]">
                    {{ subject.studentCount }}
                  </dd>
                </div>
                <div class="rounded-lg bg-[#fbfaf8] p-3">
                  <dt class="flex items-center gap-1.5 text-[#8a8494]">
                    <PhFileText :size="15" weight="duotone" />
                    Materi
                  </dt>
                  <dd class="mt-1 text-base font-medium text-[#171322]">
                    {{ subject.materialCount }}
                  </dd>
                </div>
                <div class="rounded-lg bg-[#fbfaf8] p-3">
                  <dt class="flex items-center gap-1.5 text-[#8a8494]">
                    <PhClipboardText :size="15" weight="duotone" />
                    Tugas
                  </dt>
                  <dd class="mt-1 text-base font-medium text-[#171322]">
                    {{ subject.assignmentCount }}
                  </dd>
                </div>
                <div class="rounded-lg bg-[#fff7ed] p-3">
                  <dt class="flex items-center gap-1.5 text-[#b45309]">
                    <PhWarningCircle :size="15" weight="duotone" />
                    Perlu dinilai
                  </dt>
                  <dd class="mt-1 text-base font-medium text-[#b45309]">
                    {{ subject.pendingSubmissions }}
                  </dd>
                </div>
              </dl>
            </div>
          </RouterLink>
        </div>
      </section>
    </section>
  </main>
</template>
