<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  PhArrowLeft,
  PhCaretLeft,
  PhCaretRight,
  PhCheckCircle,
  PhClock,
  PhDownloadSimple,
  PhFile,
  PhPaperPlaneTilt,
  PhUser,
} from "@phosphor-icons/vue";
import {
  getAssignmentDetailWithSubmissions,
  assessSubmission,
} from "../../services/teacherAssignment";
import type {
  AssignmentWithSubmissionsResponse,
  TeacherSubmission,
} from "../../types/teacherAssignment";
import { formatDateTime } from "../../utils/date";
import { useToastStore } from "../../stores/toast";

const route = useRoute();
const router = useRouter();
const toast = useToastStore();

const assignmentId = computed(() => String(route.params.assignmentId ?? ""));
const assignment = ref<AssignmentWithSubmissionsResponse["assignment"] | null>(null);
const submissions = ref<TeacherSubmission[]>([]);
const loading = ref(false);
const submitting = ref(false);
const errorMessage = ref("");
const activeIndex = ref(0);

const currentSubmission = computed<TeacherSubmission | null>(
  () => submissions.value[activeIndex.value] ?? null,
);

// Grading form state
const score = ref<number | string>("");
const feedback = ref("");

async function loadData() {
  loading.value = true;
  errorMessage.value = "";
  try {
    const data = await getAssignmentDetailWithSubmissions(assignmentId.value);
    assignment.value = data.assignment;
    submissions.value = data.submissions ?? [];
    if (activeIndex.value >= submissions.value.length) {
      activeIndex.value = Math.max(submissions.value.length - 1, 0);
    }

    updateGradingForm();
  } catch (err) {
    console.error("Failed to load assignment review", err);
    errorMessage.value = getLoadErrorMessage(err);
  } finally {
    loading.value = false;
  }
}

function getLoadErrorMessage(error: unknown) {
  if (typeof error === "object" && error !== null && "response" in error) {
    const status = (error as { response?: { status?: number } }).response?.status;
    if (status === 403) return "Anda tidak memiliki akses ke tugas ini.";
    if (status === 404) return "Tugas tidak ditemukan.";
  }
  return "Data review belum bisa dimuat.";
}

function updateGradingForm() {
  if (currentSubmission.value?.assessment) {
    score.value = currentSubmission.value.assessment.score;
    feedback.value = currentSubmission.value.assessment.feedback;
  } else {
    score.value = "";
    feedback.value = "";
  }
}

watch(activeIndex, updateGradingForm);

function formatFileSize(size?: number) {
  if (!size || size <= 0) return "Ukuran tidak tersedia";
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${Math.round(size / 1024)} KB`;
  return `${(size / (1024 * 1024)).toFixed(2)} MB`;
}

async function handleGrade() {
  if (!currentSubmission.value) return;
  if (score.value === "") {
    toast.error("Nilai wajib diisi.");
    return;
  }

  submitting.value = true;
  try {
    await assessSubmission(currentSubmission.value.submissionId, {
      score: Number(score.value),
      feedback: feedback.value,
    });

    await loadData();
    toast.success("Nilai berhasil disimpan.");
  } catch (err) {
    toast.error("Gagal menyimpan nilai.");
  } finally {
    submitting.value = false;
  }
}

function nextStudent() {
  if (activeIndex.value < submissions.value.length - 1) {
    activeIndex.value++;
  }
}

function prevStudent() {
  if (activeIndex.value > 0) {
    activeIndex.value--;
  }
}

onMounted(loadData);
</script>

<template>
  <main class="h-[calc(100vh-64px)] overflow-hidden flex flex-col bg-[#F8F7F4]">
    <!-- Topbar -->
    <header
      class="h-14 bg-white border-b border-[#EBEBEB] flex items-center justify-between px-5 shrink-0"
    >
      <div class="flex items-center gap-4">
        <button
          @click="router.back()"
          class="flex items-center gap-2 text-xs font-semibold text-[#6B7280] hover:text-[#111827] transition"
        >
          <PhArrowLeft :size="16" />
          {{ assignment?.assignmentTitle || "Kembali" }}
        </button>
      </div>

      <div class="flex items-center gap-4">
        <div class="flex items-center gap-2">
          <button
            @click="prevStudent"
            :disabled="activeIndex === 0"
            class="p-2 border border-[#EBEBEB] rounded-lg hover:bg-[#F9FAFB] disabled:opacity-30 transition"
          >
            <PhCaretLeft :size="16" />
          </button>
          <span class="text-xs font-medium text-[#374151]">
            {{ submissions.length > 0 ? activeIndex + 1 : 0 }} /
            {{ submissions.length }} Siswa
          </span>
          <button
            @click="nextStudent"
            :disabled="activeIndex === submissions.length - 1"
            class="p-2 border border-[#EBEBEB] rounded-lg hover:bg-[#F9FAFB] disabled:opacity-30 transition"
          >
            <PhCaretRight :size="16" />
          </button>
        </div>
      </div>
    </header>

    <div v-if="loading" class="flex-1 flex items-center justify-center">
      <p class="text-sm text-[#6B7280]">Memuat pengumpulan...</p>
    </div>

    <div
      v-else-if="errorMessage"
      class="flex-1 flex flex-col items-center justify-center text-center p-8"
    >
      <div
        class="w-16 h-16 bg-white rounded-2xl shadow-sm flex items-center justify-center mb-4"
      >
        <PhFile :size="32" class="text-[#D1D5DB]" />
      </div>
      <h2 class="text-lg font-semibold text-[#111827]">Review belum tersedia</h2>
      <p class="text-sm text-[#6B7280] mt-2 max-w-xs">
        {{ errorMessage }}
      </p>
    </div>

    <div
      v-else-if="submissions.length === 0"
      class="flex-1 flex flex-col items-center justify-center text-center p-8"
    >
      <div
        class="w-16 h-16 bg-white rounded-2xl shadow-sm flex items-center justify-center mb-4"
      >
        <PhUser :size="32" class="text-[#D1D5DB]" />
      </div>
      <h2 class="text-lg font-semibold text-[#111827]">Belum Ada Pengumpulan</h2>
      <p class="text-sm text-[#6B7280] mt-2 max-w-xs">
        Belum ada siswa yang mengumpulkan tugas ini.
      </p>
    </div>

    <div v-else class="flex-1 flex overflow-hidden">
      <!-- Main Content (Submission Viewer) -->
      <section class="flex-1 overflow-y-auto p-6 space-y-6">
        <!-- Student Info Header -->
        <div
          class="bg-white rounded-2xl p-5 border border-[#EBEBEB] shadow-sm flex items-center justify-between"
        >
          <div class="flex items-center gap-4">
            <div
              class="w-12 h-12 rounded-full bg-[#4F46E5] flex items-center justify-center text-white font-bold"
            >
              {{ currentSubmission?.studentName?.charAt(0) }}
            </div>
            <div>
              <h2 class="text-base font-bold text-[#111827]">
                {{ currentSubmission?.studentName }}
              </h2>
              <p
                class="text-xs text-[#6B7280] flex items-center gap-1.5 mt-0.5"
              >
                <PhClock :size="14" />
                Dikumpulkan pada
                {{ formatDateTime(currentSubmission?.submittedAt) }}
                <span
                  v-if="currentSubmission?.isLate"
                  class="ml-2 px-2 py-0.5 bg-[#FEF2F2] text-[#DC2626] rounded-full text-[10px] font-bold"
                  >TERLAMBAT</span
                >
                <span
                  v-else
                  class="ml-2 px-2 py-0.5 bg-[#ECFDF5] text-[#059669] rounded-full text-[10px] font-bold"
                  >TEPAT WAKTU</span
                >
              </p>
            </div>
          </div>
          <div v-if="currentSubmission?.assessment" class="text-right">
            <p
              class="text-[10px] font-bold text-[#6B7280] uppercase tracking-wider mb-1"
            >
              Status
            </p>
            <span
              class="px-3 py-1 bg-[#ECFDF5] text-[#059669] rounded-full text-xs font-bold flex items-center gap-1.5 justify-end"
            >
              <PhCheckCircle :size="14" weight="bold" />
              Sudah Dinilai
            </span>
          </div>
        </div>

        <!-- Attachments -->
        <div class="space-y-4">
          <h3
            class="text-xs font-bold text-[#374151] uppercase tracking-wider flex items-center gap-2"
          >
            <PhFile :size="16" weight="bold" />
            File Jawaban ({{ currentSubmission?.attachments?.length || 0 }})
          </h3>

          <div class="grid gap-3">
            <div
              v-for="file in currentSubmission?.attachments"
              :key="file.mediaId"
              class="bg-white rounded-xl border border-[#EBEBEB] overflow-hidden group hover:border-[#4F46E5]/30 transition"
            >
              <div class="p-4 flex items-center justify-between gap-3">
                <div class="flex min-w-0 flex-1 items-center gap-4 overflow-hidden">
                  <div
                    class="w-10 h-10 shrink-0 bg-[#FEF2F2] rounded-lg flex items-center justify-center text-[#DC2626]"
                  >
                    <PhFile :size="20" weight="bold" />
                  </div>
                  <div class="min-w-0 flex-1 overflow-hidden">
                    <p class="text-sm font-semibold text-[#111827] truncate">
                      {{ file.mediaName }}
                    </p>
                    <p class="text-[11px] text-[#9CA3AF]">
                      {{ formatFileSize(file.fileSize) }}
                    </p>
                  </div>
                </div>
                <a
                  :href="file.fileUrl"
                  target="_blank"
                  class="flex shrink-0 items-center gap-1.5 text-xs font-bold text-[#4F46E5] hover:text-[#4338CA] transition"
                >
                  <PhDownloadSimple :size="16" weight="bold" />
                  Unduh
                </a>
              </div>
              <!-- Preview Area (Simulated) -->
              <div
                class="h-48 bg-[#F3F4F6] border-t border-[#EBEBEB] flex flex-col items-center justify-center text-[#9CA3AF]"
              >
                <PhFile :size="32" class="mb-2" />
                <p class="text-xs">Klik unduh untuk melihat dokumen lengkap</p>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- Right Sidebar (Grading Panel) -->
      <aside
        class="w-80 bg-white border-l border-[#EBEBEB] overflow-y-auto p-5 flex flex-col gap-6"
      >
        <section>
          <h3
            class="text-xs font-bold text-[#374151] uppercase tracking-wider mb-4"
          >
            Panel Penilaian
          </h3>

          <div class="space-y-4">
            <div>
              <label
                class="block text-xs font-bold text-[#6B7280] mb-2 uppercase"
                >Nilai (0-100)</label
              >
              <div class="flex items-end gap-3">
                <input
                  v-model="score"
                  type="number"
                  min="0"
                  max="100"
                  class="w-24 px-4 py-3 bg-[#F9FAFB] border border-[#EBEBEB] rounded-2xl text-2xl font-bold text-[#111827] text-center outline-none focus:border-[#4F46E5] transition"
                  placeholder="0"
                />
                <span class="text-sm text-[#9CA3AF] mb-3">/ 100</span>
              </div>
            </div>

            <div>
              <label
                class="block text-xs font-bold text-[#6B7280] mb-2 uppercase"
                >Feedback Guru</label
              >
              <textarea
                v-model="feedback"
                rows="6"
                class="w-full px-4 py-3 bg-[#F9FAFB] border border-[#EBEBEB] rounded-2xl text-xs text-[#374151] outline-none focus:border-[#4F46E5] transition resize-none leading-relaxed"
                placeholder="Berikan masukan untuk siswa..."
              ></textarea>
            </div>

            <button
              @click="handleGrade"
              :disabled="submitting"
              class="w-full py-3 bg-[#171322] text-white rounded-xl text-sm font-bold flex items-center justify-center gap-2 hover:bg-[#2f2b3a] transition disabled:opacity-50"
            >
              <PhPaperPlaneTilt v-if="!submitting" :size="18" weight="bold" />
              {{
                submitting
                  ? "Menyimpan..."
                  : currentSubmission?.assessment
                    ? "Edit Nilai"
                    : "Beri Nilai"
              }}
            </button>
          </div>
        </section>

        <section class="border-t border-[#F3F4F6] pt-6">
          <h3
            class="text-xs font-bold text-[#374151] uppercase tracking-wider mb-4"
          >
            Daftar Siswa
          </h3>
          <div class="space-y-1.5 max-h-100 overflow-y-auto">
            <button
              v-for="(sub, index) in submissions"
              :key="sub.submissionId"
              @click="activeIndex = index"
              :class="[
                'w-full flex items-center justify-between p-2.5 rounded-xl transition text-left',
                activeIndex === index
                  ? 'bg-[#EEF2FF] border border-[#4F46E5]/10'
                  : 'hover:bg-[#F9FAFB]',
              ]"
            >
              <div class="flex items-center gap-3 overflow-hidden">
                <div
                  class="w-7 h-7 rounded-full shrink-0 flex items-center justify-center text-[10px] font-bold text-white"
                  :class="
                    activeIndex === index ? 'bg-[#4F46E5]' : 'bg-[#9CA3AF]'
                  "
                >
                  {{ sub.studentName?.charAt(0) }}
                </div>
                <div class="min-w-0">
                  <p
                    :class="[
                      'text-xs truncate',
                      activeIndex === index
                        ? 'font-bold text-[#4F46E5]'
                        : 'text-[#374151]',
                    ]"
                  >
                    {{ sub.studentName }}
                  </p>
                </div>
              </div>
              <div
                v-if="sub.assessment"
                class="shrink-0 w-1.5 h-1.5 rounded-full bg-[#059669]"
              ></div>
              <div
                v-else
                class="shrink-0 w-1.5 h-1.5 rounded-full bg-[#EA580C]"
              ></div>
            </button>
          </div>
        </section>
      </aside>
    </div>
  </main>
</template>

<style scoped>
/* Chrome, Safari, Edge, Opera */
input::-webkit-outer-spin-button,
input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

/* Firefox */
input[type="number"] {
  -moz-appearance: textfield;
}
</style>
