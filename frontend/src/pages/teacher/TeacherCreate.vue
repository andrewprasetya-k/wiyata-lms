<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink, useRoute } from "vue-router";
import {
  PhArrowRight,
  PhBookOpen,
  PhChalkboardTeacher,
  PhPlusCircle,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { getMyTeachingSubjectClasses } from "../../services/teacherSubjects";
import type { TeacherSubjectClass } from "../../types/teacherSubjects";
import { resolveSubjectColor } from "../../utils/color";

const route = useRoute();
const loading = ref(false);
const errorMessage = ref("");
const subjects = ref<TeacherSubjectClass[]>([]);

const hasSubjects = computed(() => subjects.value.length > 0);
const requestedType = computed(() => {
  if (route.query.type === "assignment") return "assignment";
  if (route.query.type === "material") return "material";
  return "";
});

function createContentTarget(subjectClassId: string) {
  return {
    name: "teacher-content-create",
    params: { subjectClassId },
    query: requestedType.value ? { type: requestedType.value } : {},
  };
}

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
  <main class="min-h-screen flex-1 px-5 py-5 sm:px-6 lg:px-8">
    <section class="flex w-full max-w-none flex-col gap-5">
      <header
        class="rounded-[22px] bg-[#f0e9dd] px-5 py-5 shadow-sm ring-1 ring-black/5 md:px-6"
      >
        <p class="text-sm font-medium text-[#8a6d3b]">Buat konten</p>
        <h1 class="mt-3 text-3xl font-medium text-foreground md:text-4xl">
          Pilih mata pelajaran terlebih dahulu
        </h1>
        <p class="mt-3 max-w-2xl text-sm leading-6 text-[#6b6475]">
          Materi dan tugas dibuat dari ruang mengajar agar setiap konten
          terhubung ke mata pelajaran dan kelas yang tepat.
        </p>
      </header>

      <section
        v-if="loading"
        class="rounded-[22px] bg-surface p-5 shadow-sm ring-1 ring-black/5"
      >
        <p class="text-sm text-[#6b6475]">Memuat mata pelajaran yang diajar...</p>
      </section>

      <section
        v-else-if="errorMessage"
        class="rounded-[22px] bg-surface p-5 shadow-sm ring-1 ring-black/5"
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
              <h2 class="text-lg font-medium text-foreground">
                Gagal memuat mata pelajaran
              </h2>
              <p class="mt-2 text-sm leading-6 text-[#6b6475]">
                {{ errorMessage }}
              </p>
            </div>
          </div>
          <button
            class="rounded-xl bg-foreground px-4 py-3 text-sm font-medium text-white"
            type="button"
            @click="loadSubjects"
          >
            Coba lagi
          </button>
        </div>
      </section>

      <section
        v-else-if="!hasSubjects"
        class="rounded-[22px] bg-surface p-5 shadow-sm ring-1 ring-black/5"
      >
        <div class="flex items-start gap-3">
          <div
            class="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-brand-soft text-brand"
          >
            <PhChalkboardTeacher :size="24" weight="duotone" />
          </div>
          <div>
            <h2 class="text-lg font-medium text-foreground">
              Belum ada mata pelajaran yang diajar
            </h2>
            <p class="mt-2 max-w-2xl text-sm leading-6 text-[#6b6475]">
              Admin sekolah perlu menugaskan guru ke kelas ajar terlebih
              dahulu sebelum materi atau tugas dapat dibuat.
            </p>
          </div>
        </div>
      </section>

      <section v-else class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        <article
          v-for="subject in subjects"
          :key="subject.subjectClassId"
          class="rounded-[22px] bg-surface p-5 shadow-sm ring-1 ring-black/5"
        >
          <div class="flex items-start justify-between gap-4">
            <div
              class="flex h-13 w-13 items-center justify-center rounded-xl text-white"
              :style="{
                backgroundColor: resolveSubjectColor(subject),
              }"
            >
              <PhBookOpen :size="26" weight="duotone" />
            </div>
            <span
              v-if="subject.subjectCode"
              class="rounded-full bg-brand-soft px-3 py-1 text-xs font-medium text-brand"
            >
              {{ subject.subjectCode }}
            </span>
          </div>

          <p class="mt-5 text-sm text-[#8a8494]">
            {{ subject.className || subject.classCode || "Kelas" }}
          </p>
          <h2 class="mt-1 text-2xl font-medium text-foreground">
            {{ subject.subjectName || "Mata pelajaran" }}
          </h2>
          <p
            v-if="subject.classCode"
            class="mt-2 text-sm text-[#6b6475]"
          >
            Kode kelas: {{ subject.classCode }}
          </p>
          <p class="mt-4 text-sm leading-6 text-[#6b6475]">
            Konten yang dibuat akan masuk ke ruang mengajar ini dan dapat
            dilihat oleh siswa di kelas terkait.
          </p>

          <RouterLink
            :to="createContentTarget(subject.subjectClassId)"
            class="mt-5 inline-flex w-full items-center justify-center gap-2 rounded-xl bg-foreground px-4 py-3 text-sm font-medium text-white transition hover:bg-[#2f2b3a]"
          >
            <PhPlusCircle :size="18" weight="duotone" />
            Buat Konten
            <PhArrowRight :size="16" />
          </RouterLink>
        </article>
      </section>
    </section>
  </main>
</template>
