<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import {
  PhArrowRight,
  PhBooks,
  PhCaretDown,
  PhMegaphone,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { getSubjectClassesByClass } from "../../services/classWorkspace";
import { useActiveClassStore } from "../../stores/activeClass";
import { useAuthStore } from "../../stores/auth";
import type { SubjectClassItem } from "../../types/classWorkspace";
import { getSubjectColor } from "../../utils/color";

const auth = useAuthStore();
const activeClassStore = useActiveClassStore();
const router = useRouter();

const subjects = ref<SubjectClassItem[]>([]);
const isLoading = ref(true);
const errorMessage = ref("");

const activeMembership = computed(() => auth.activeMembership);

const schoolName = computed(
  () => activeMembership.value?.school.name ?? "Sekolah aktif",
);
const schoolUserId = computed(() => auth.activeSchoolUserId);
const classes = computed(() => activeClassStore.classes);
const activeClass = computed(() => activeClassStore.activeClass);

async function loadSubjects() {
  if (!schoolUserId.value) {
    isLoading.value = false;
    errorMessage.value =
      "Konteks sekolah belum tersedia. Silakan login ulang atau pilih sekolah aktif terlebih dahulu.";
    return;
  }

  isLoading.value = true;
  errorMessage.value = "";

  try {
    await activeClassStore.loadClasses(schoolUserId.value);
    if (activeClassStore.errorMessage) {
      errorMessage.value = activeClassStore.errorMessage;
      subjects.value = [];
      return;
    }

    if (!activeClassStore.activeClassId) {
      subjects.value = [];
      return;
    }

    const data = await getSubjectClassesByClass(activeClassStore.activeClassId);
    subjects.value = data.subjects;
  } catch {
    errorMessage.value =
      activeClassStore.errorMessage ||
      "Mata pelajaran belum bisa dimuat. Periksa koneksi atau coba lagi nanti.";
  } finally {
    isLoading.value = false;
  }
}

async function changeActiveClass(classId: string) {
  activeClassStore.setActiveClass(classId);
  subjects.value = [];
  await loadSubjects();
}

function openSubject(subject: SubjectClassItem) {
  router.push({
    path: `/student/subjects/${subject.subjectClassId}`,
    query: subject.subjectName ? { title: subject.subjectName } : undefined,
  });
}

onMounted(loadSubjects);
</script>

<template>
  <main class="min-h-screen flex-1 bg-[#f8f7f4]">
    <section
      class="border-b border-[#ebe7df] bg-white px-5 py-3 sm:px-6 lg:px-8"
    >
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="flex items-center gap-2">
          <span class="text-xs text-[#9a95a3]">Kelas aktif:</span>
          <div
            class="flex items-center gap-2 rounded-xl border border-[#ebe7df] bg-[#f9fafb] px-3 py-2"
          >
            <div
              class="flex h-5 w-5 items-center justify-center rounded-md bg-[#4f46e5] text-[10px] text-white"
            >
              {{ activeClass?.classTitle?.slice(0, 2).toUpperCase() || "EV" }}
            </div>
            <div>
              <p class="text-xs font-medium text-[#171322]">
                {{ activeClass?.classTitle || "Belum ada kelas aktif" }}
              </p>
              <p class="text-[11px] text-[#7a7385]">{{ schoolName }}</p>
            </div>
            <PhCaretDown :size="14" class="text-[#a09aa8]" />
          </div>
          <select
            v-if="classes.length > 1"
            class="rounded-xl border border-[#ebe7df] bg-white px-3 py-2 text-xs text-[#3f3a4a] outline-none transition focus:border-[#4f46e5]"
            :value="activeClassStore.activeClassId ?? ''"
            @change="
              changeActiveClass(($event.target as HTMLSelectElement).value)
            "
          >
            <option
              v-for="item in classes"
              :key="item.enrollmentId"
              :value="item.classId"
            >
              {{ item.classTitle || "Kelas" }}
            </option>
          </select>
        </div>

        <RouterLink
          class="flex items-center gap-2 rounded-xl border border-[#ebe7df] bg-white px-3 py-2 text-xs font-medium text-[#3f3a4a] transition hover:border-[#4f46e5] hover:text-[#4f46e5]"
          to="/student/feed"
        >
          <PhMegaphone :size="16" />
          Feed kelas
        </RouterLink>
      </div>
    </section>

    <section class="px-5 py-5 sm:px-6 lg:px-8">
      <header class="mb-5 flex items-center justify-between gap-4">
        <div>
          <h1 class="text-2xl font-medium tracking-normal text-[#171322]">
            Mata pelajaran
          </h1>
          <p class="mt-1 text-sm text-[#7a7385]">
            {{ subjects.length }} subject class tersedia
            <span v-if="activeClass?.classTitle">
              · {{ activeClass.classTitle }}</span
            >
          </p>
        </div>
      </header>

      <section
        v-if="isLoading || activeClassStore.isLoading"
        class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3"
      >
        <div
          v-for="item in 6"
          :key="item"
          class="h-44 animate-pulse rounded-[18px] border border-[#ebe7df] bg-white"
        />
      </section>

      <section
        v-else-if="errorMessage"
        class="soft-card max-w-2xl rounded-[22px] p-5"
      >
        <div
          class="mb-4 flex h-11 w-11 items-center justify-center rounded-2xl bg-[#fff1f0] text-[#f2756a]"
        >
          <PhWarningCircle :size="24" weight="duotone" />
        </div>
        <p class="text-sm font-medium text-[#171322]">
          Tidak bisa memuat mata pelajaran
        </p>
        <p class="mt-2 text-sm leading-6 text-[#7a7385]">{{ errorMessage }}</p>
        <button
          class="mt-5 rounded-2xl bg-[#4f46e5] px-4 py-2 text-sm font-medium text-white"
          type="button"
          @click="loadSubjects"
        >
          Coba lagi
        </button>
      </section>

      <section
        v-else-if="!activeClass"
        class="soft-card max-w-2xl rounded-[22px] p-5"
      >
        <div
          class="mb-4 flex h-11 w-11 items-center justify-center rounded-2xl bg-[#eef2ff] text-[#4f46e5]"
        >
          <PhBooks :size="24" weight="duotone" />
        </div>
        <p class="text-sm font-medium text-[#171322]">Belum ada kelas aktif</p>
        <p class="mt-2 text-sm leading-6 text-[#7a7385]">
          Mata pelajaran akan tampil setelah akunmu terdaftar pada kelas di
          sekolah aktif.
        </p>
      </section>

      <section
        v-else-if="subjects.length === 0"
        class="soft-card max-w-2xl rounded-[22px] p-5"
      >
        <div
          class="mb-4 flex h-11 w-11 items-center justify-center rounded-2xl bg-[#eef2ff] text-[#4f46e5]"
        >
          <PhBooks :size="24" weight="duotone" />
        </div>
        <p class="text-sm font-medium text-[#171322]">
          Subject class belum tersedia
        </p>
        <p class="mt-2 text-sm leading-6 text-[#7a7385]">
          Kelas aktif belum memiliki mata pelajaran yang bisa ditampilkan.
        </p>
      </section>

      <section v-else class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
        <article
          v-for="subject in subjects"
          :key="subject.subjectClassId"
          class="group overflow-hidden rounded-[18px] border border-[#ebe7df] bg-white transition hover:-translate-y-0.5 hover:shadow-[0_18px_40px_rgba(66,55,40,0.08)]"
        >
          <button
            class="block w-full text-left"
            type="button"
            @click="openSubject(subject)"
          >
            <div
              class="relative flex h-24 flex-col justify-end px-4 pb-4 text-white"
              :style="{
                backgroundColor: getSubjectColor(
                  subject.subjectClassId ||
                    subject.subjectName ||
                    subject.subjectCode,
                ),
              }"
            >
              <span class="text-base font-medium">
                {{
                  subject.subjectName || subject.subjectCode || "Mata pelajaran"
                }}
              </span>
              <p class="mt-0.5 text-xs text-white/80">
                {{ subject.teacherName || "Guru belum tersedia" }}
              </p>
            </div>
            <div class="space-y-3 px-4 py-4">
              <div class="flex items-center justify-between gap-3">
                <div>
                  <p class="text-xs text-[#9a95a3]">Kode Mapel</p>
                  <p class="mt-1 text-sm font-medium text-[#3f3a4a]">
                    {{ subject.subjectCode || "Kode belum tersedia" }}
                  </p>
                </div>
                <PhArrowRight
                  :size="18"
                  class="text-[#a09aa8] transition group-hover:translate-x-0.5 group-hover:text-[#4f46e5]"
                />
              </div>
              <div class="flex flex-wrap gap-2 border-t border-[#f3f1ec] pt-3">
                <span
                  class="rounded-full bg-[#f3f1ec] px-2 py-1 text-[11px] text-[#6b6475]"
                >
                  Materi
                </span>
                <span
                  class="rounded-full bg-[#fff7ed] px-2 py-1 text-[11px] text-[#b45309]"
                >
                  Tugas
                </span>
                <span
                  class="rounded-full bg-[#eef2ff] px-2 py-1 text-[11px] text-[#4f46e5]"
                >
                  Catatan
                </span>
              </div>
            </div>
          </button>
        </article>
      </section>

      <!-- <section v-if="classes.length > 1" class="mt-6 rounded-2xl border border-[#ebe7df] bg-white p-4">
        <p class="text-xs font-medium uppercase tracking-[0.08em] text-[#9a95a3]">Class context</p>
        <div class="mt-3 grid gap-2 sm:grid-cols-2 xl:grid-cols-3">
          <div
            v-for="item in classes"
            :key="item.enrollmentId"
            class="flex items-center gap-3 rounded-xl px-3 py-2"
            :class="item.classId === activeClass?.classId ? 'bg-[#eef2ff]' : 'bg-[#f9fafb]'"
          >
            <div class="h-2 w-2 rounded-full bg-[#4f46e5]" />
            <div class="min-w-0 flex-1">
              <p class="truncate text-xs font-medium text-[#171322]">{{ item.classTitle || 'Kelas' }}</p>
              <p class="text-[11px] text-[#7a7385]">Dipakai sebagai konteks akademik aktif</p>
            </div>
            <PhCheck v-if="item.classId === activeClass?.classId" :size="15" class="text-[#4f46e5]" />
          </div>
        </div>
      </section> -->
    </section>
  </main>
</template>
