<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import {
  PhBookOpen,
  PhCaretDown,
  PhChartBar,
  PhCheckCircle,
  PhClock,
  PhSealCheck,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { getMyGradebookByClass } from "../../services/studentGrades";
import { useActiveClassStore } from "../../stores/activeClass";
import { useAuthStore } from "../../stores/auth";
import type {
  GradebookAssignment,
  MyGradebookResponse,
} from "../../types/studentGrades";
import { getSubjectColor } from "../../utils/color";
import { formatDateTime } from "../../utils/date";

const auth = useAuthStore();
const activeClassStore = useActiveClassStore();
const gradebook = ref<MyGradebookResponse | null>(null);
const isLoading = ref(true);
const errorMessage = ref("");

const activeMembership = computed(() => auth.activeMembership);

const schoolName = computed(
  () => activeMembership.value?.school.name ?? "Sekolah aktif",
);
const schoolUserId = computed(() => auth.activeSchoolUserId);
const activeClass = computed(() => activeClassStore.activeClass);
const subjects = computed(() => gradebook.value?.subjects ?? []);
const hasAssignments = computed(() =>
  subjects.value.some((subject) => subject.assignments.length > 0),
);

async function loadGrades(selectedClassId?: string) {
  if (!schoolUserId.value) {
    isLoading.value = false;
    errorMessage.value =
      "Konteks sekolah belum tersedia. Silakan login ulang atau pilih sekolah aktif terlebih dahulu.";
    return;
  }

  isLoading.value = true;
  errorMessage.value = "";
  gradebook.value = null;

  try {
    await activeClassStore.loadClasses(schoolUserId.value);
    if (activeClassStore.errorMessage) {
      errorMessage.value = activeClassStore.errorMessage;
      return;
    }

    const classId = selectedClassId ?? activeClassStore.activeClassId;
    if (!classId) {
      return;
    }

    gradebook.value = await getMyGradebookByClass(classId);
  } catch (error) {
    errorMessage.value = getGradebookErrorMessage(error);
  } finally {
    isLoading.value = false;
  }
}

async function changeActiveClass(classId: string) {
  activeClassStore.setActiveClass(classId);
  await loadGrades(classId);
}

function getGradebookErrorMessage(error: unknown) {
  if (typeof error === "object" && error !== null && "response" in error) {
    const response = (error as { response?: { status?: number } }).response;
    if (response?.status === 403) {
      return "Kamu belum terdaftar sebagai siswa pada kelas ini.";
    }
    if (response?.status === 404) {
      return "Kelas tidak ditemukan.";
    }
  }

  return "Nilai belum bisa dimuat. Periksa koneksi atau coba lagi nanti.";
}

function statusLabel(assignment: GradebookAssignment) {
  if (assignment.status === "graded") return "Sudah dinilai";
  if (assignment.status === "submitted") return "Sudah dikumpulkan";
  return "Belum dikumpulkan";
}

function statusClasses(assignment: GradebookAssignment) {
  if (assignment.status === "graded") return "bg-[#ecfdf3] text-[#027a48]";
  if (assignment.status === "submitted") return "bg-[#eef2ff] text-[#4f46e5]";
  return "bg-[#fff7ed] text-[#b45309]";
}

function formatScore(value?: number | null) {
  if (value === null || value === undefined) return "Belum tersedia";
  return new Intl.NumberFormat("id-ID", { maximumFractionDigits: 2 }).format(
    value,
  );
}

onMounted(loadGrades);
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
                {{
                  activeClass?.classTitle ||
                  gradebook?.class.className ||
                  "Belum ada kelas aktif"
                }}
              </p>
              <p class="text-[11px] text-[#7a7385]">{{ schoolName }}</p>
            </div>
            <PhCaretDown :size="14" class="text-[#a09aa8]" />
          </div>
          <select
            v-if="activeClassStore.classes.length > 1"
            class="rounded-xl border border-[#ebe7df] bg-white px-3 py-2 text-xs text-[#3f3a4a] outline-none transition focus:border-[#4f46e5]"
            :value="activeClassStore.activeClassId ?? ''"
            @change="
              changeActiveClass(($event.target as HTMLSelectElement).value)
            "
          >
            <option
              v-for="item in activeClassStore.classes"
              :key="item.enrollmentId"
              :value="item.classId"
            >
              {{ item.classTitle || "Kelas" }}
            </option>
          </select>
        </div>
      </div>
    </section>

    <section class="px-5 py-5 sm:px-6 lg:px-8">
      <header class="mb-5 flex flex-col gap-2">
        <p class="text-sm text-[#7a7385]">Gradebook siswa</p>
        <h1 class="text-2xl font-medium tracking-normal text-[#171322]">
          Nilai dan feedback
        </h1>
        <p class="max-w-2xl text-sm leading-6 text-[#7a7385]">
          Nilai dikelompokkan berdasarkan subject pada kelas aktif. Data hanya
          menampilkan pengumpulan dan penilaian milik akun login saat ini.
        </p>
      </header>

      <section
        v-if="isLoading || activeClassStore.isLoading"
        class="grid gap-3 lg:grid-cols-[0.9fr_1.1fr]"
      >
        <div
          class="h-40 animate-pulse rounded-[22px] border border-[#ebe7df] bg-white"
        />
        <div
          class="h-40 animate-pulse rounded-[22px] border border-[#ebe7df] bg-white"
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
          Tidak bisa memuat nilai
        </p>
        <p class="mt-2 text-sm leading-6 text-[#7a7385]">{{ errorMessage }}</p>
        <button
          class="mt-5 rounded-2xl bg-[#4f46e5] px-4 py-2 text-sm font-medium text-white"
          type="button"
          @click="loadGrades()"
        >
          Coba lagi
        </button>
      </section>

      <section
        v-else-if="!activeClassStore.activeClassId"
        class="soft-card max-w-2xl rounded-[22px] p-5"
      >
        <div
          class="mb-4 flex h-11 w-11 items-center justify-center rounded-2xl bg-[#eef2ff] text-[#4f46e5]"
        >
          <PhBookOpen :size="24" weight="duotone" />
        </div>
        <p class="text-sm font-medium text-[#171322]">Belum ada kelas aktif</p>
        <p class="mt-2 text-sm leading-6 text-[#7a7385]">
          Nilai akan tampil setelah akunmu terdaftar pada kelas di sekolah
          aktif.
        </p>
      </section>

      <section
        v-else-if="subjects.length === 0"
        class="soft-card max-w-2xl rounded-[22px] p-5"
      >
        <div
          class="mb-4 flex h-11 w-11 items-center justify-center rounded-2xl bg-[#eef2ff] text-[#4f46e5]"
        >
          <PhChartBar :size="24" weight="duotone" />
        </div>
        <p class="text-sm font-medium text-[#171322]">
          Gradebook belum tersedia
        </p>
        <p class="mt-2 text-sm leading-6 text-[#7a7385]">
          Subject class pada kelas aktif belum memiliki data gradebook.
        </p>
      </section>

      <section v-else class="space-y-5">
        <div class="grid gap-3 md:grid-cols-4">
          <div class="rounded-[20px] border border-[#ebe7df] bg-white p-4">
            <p class="text-xs text-[#8b8592]">Subject</p>
            <p class="mt-2 text-2xl font-medium text-[#171322]">
              {{ gradebook?.summary.subjectCount ?? subjects.length }}
            </p>
          </div>
          <div class="rounded-[20px] border border-[#ebe7df] bg-white p-4">
            <p class="text-xs text-[#8b8592]">Sudah dinilai</p>
            <p class="mt-2 text-2xl font-medium text-[#027a48]">
              {{ gradebook?.summary.gradedAssignmentCount ?? 0 }}
            </p>
          </div>
          <div class="rounded-[20px] border border-[#ebe7df] bg-white p-4">
            <p class="text-xs text-[#8b8592]">Sudah dikumpulkan</p>
            <p class="mt-2 text-2xl font-medium text-[#4f46e5]">
              {{ gradebook?.summary.submittedAssignmentCount ?? 0 }}
            </p>
          </div>
          <div class="rounded-[20px] border border-[#ebe7df] bg-white p-4">
            <p class="text-xs text-[#8b8592]">Menunggu nilai</p>
            <p class="mt-2 text-2xl font-medium text-[#b45309]">
              {{ gradebook?.summary.pendingAssessmentCount ?? 0 }}
            </p>
          </div>
        </div>

        <div v-if="!hasAssignments" class="soft-card rounded-[22px] p-5">
          <div
            class="mb-4 flex h-11 w-11 items-center justify-center rounded-2xl bg-[#eef2ff] text-[#4f46e5]"
          >
            <PhChartBar :size="24" weight="duotone" />
          </div>
          <p class="text-sm font-medium text-[#171322]">
            Belum ada tugas untuk dinilai
          </p>
          <p class="mt-2 text-sm leading-6 text-[#7a7385]">
            Subject sudah tersedia, tetapi belum ada assignment pada kelas
            aktif.
          </p>
        </div>

        <article
          v-for="subject in subjects"
          :key="subject.subjectClassId"
          class="overflow-hidden rounded-[22px] border border-[#ebe7df] bg-white"
        >
          <div
            class="flex flex-col gap-4 px-5 py-5 text-white md:flex-row md:items-center md:justify-between"
            :style="{
              backgroundColor: getSubjectColor(
                subject.subjectClassId ||
                  subject.subjectName ||
                  subject.subjectCode,
              ),
            }"
          >
            <div>
              <p class="text-sm text-white/80">
                {{ subject.subjectCode || "Kode belum tersedia" }}
              </p>
              <h2 class="mt-1 text-xl font-medium tracking-normal">
                {{ subject.subjectName || "Mata pelajaran" }}
              </h2>
            </div>
            <div class="rounded-2xl bg-white/15 px-4 py-3 backdrop-blur">
              <p class="text-xs text-white/80">Rata-rata berbobot</p>
              <p class="mt-1 text-2xl font-medium">
                {{ formatScore(subject.finalGrade) }}
                <span v-if="subject.letterGrade" class="text-sm text-white/80">
                  {{ subject.letterGrade }}
                </span>
              </p>
              <p
                v-if="subject.finalGrade !== null && subject.finalGrade !== undefined"
                class="mt-2 max-w-xs text-xs leading-5 text-white/75"
              >
                Dihitung dari tugas yang sudah dinilai dan bobot kategori yang tersedia.
              </p>
              <p v-else class="mt-2 max-w-xs text-xs leading-5 text-white/75">
                Belum tersedia karena bobot atau nilai belum lengkap.
              </p>
            </div>
          </div>

          <div
            class="grid gap-2 border-b border-[#f3f1ec] px-5 py-4 sm:grid-cols-3"
          >
            <span
              class="rounded-full bg-[#ecfdf3] px-3 py-1.5 text-xs text-[#027a48]"
            >
              {{ subject.gradedCount }} sudah dinilai
            </span>
            <span
              class="rounded-full bg-[#eef2ff] px-3 py-1.5 text-xs text-[#4f46e5]"
            >
              {{ subject.submittedCount }} sudah dikumpulkan
            </span>
            <span
              class="rounded-full bg-[#fff7ed] px-3 py-1.5 text-xs text-[#b45309]"
            >
              {{ subject.pendingCount }} menunggu nilai
            </span>
          </div>

          <div v-if="subject.assignments.length === 0" class="px-5 py-5">
            <p class="text-sm font-medium text-[#171322]">
              Belum ada assignment
            </p>
            <p class="mt-2 text-sm leading-6 text-[#7a7385]">
              Nilai akan muncul setelah guru membuat dan menilai tugas pada
              subject ini.
            </p>
          </div>

          <div v-else class="divide-y divide-[#f3f1ec]">
            <div
              v-for="assignment in subject.assignments"
              :key="assignment.assignmentId"
              class="grid gap-4 px-5 py-4 lg:grid-cols-[1fr_170px]"
            >
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <span
                    class="rounded-full px-3 py-1 text-xs font-medium"
                    :class="statusClasses(assignment)"
                  >
                    {{ statusLabel(assignment) }}
                  </span>
                  <span class="text-xs text-[#8b8592]">
                    {{ assignment.categoryName || "Kategori belum tersedia" }}
                  </span>
                </div>
                <p class="mt-3 text-sm font-medium text-[#171322]">
                  {{ assignment.assignmentTitle }}
                </p>
                <div
                  class="mt-3 grid gap-2 text-xs text-[#7a7385] sm:grid-cols-2"
                >
                  <span class="inline-flex items-center gap-1.5">
                    <PhClock :size="14" />
                    Deadline: {{ formatDateTime(assignment.deadline) }}
                  </span>
                  <span
                    v-if="assignment.submittedAt"
                    class="inline-flex items-center gap-1.5"
                  >
                    <PhCheckCircle :size="14" />
                    Dikumpulkan: {{ formatDateTime(assignment.submittedAt) }}
                  </span>
                  <span
                    v-if="assignment.assessedAt"
                    class="inline-flex items-center gap-1.5"
                  >
                    <PhSealCheck :size="14" />
                    Dinilai: {{ formatDateTime(assignment.assessedAt) }}
                  </span>
                </div>
                <p
                  v-if="assignment.feedback"
                  class="mt-3 rounded-2xl bg-[#fbfaf8] px-4 py-3 text-sm leading-6 text-[#4a4356]"
                >
                  {{ assignment.feedback }}
                </p>
              </div>

              <div class="rounded-2xl bg-[#fbfaf8] p-4">
                <p class="text-xs text-[#8b8592]">Skor</p>
                <p class="mt-2 text-2xl font-medium text-[#171322]">
                  {{ formatScore(assignment.score) }}
                </p>
                <p
                  v-if="assignment.assessorName"
                  class="mt-2 text-xs text-[#7a7385]"
                >
                  oleh {{ assignment.assessorName }}
                </p>
              </div>
            </div>
          </div>
        </article>
      </section>
    </section>
  </main>
</template>
