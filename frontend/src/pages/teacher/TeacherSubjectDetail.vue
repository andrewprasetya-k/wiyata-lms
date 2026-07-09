<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink, useRoute } from "vue-router";
import {
  PhArrowLeft,
  PhBookOpen,
  PhCalendarBlank,
  PhCheckCircle,
  PhClipboardText,
  PhFileText,
  PhPaperclip,
  PhPlusCircle,
  PhUsersThree,
  PhWarningCircle,
  PhPencilSimple,
  PhTrash,
} from "@phosphor-icons/vue";
import AttachmentPreviewList from "../../components/common/AttachmentPreviewList.vue";
import { getSubjectAssignments } from "../../services/assignment";
import {
  getSubjectClassSubmissions,
  deleteAssignment,
} from "../../services/teacherAssignment";
import { getSubjectMaterials } from "../../services/teacherMaterial";
import { getMyTeachingSubjectClassById } from "../../services/teacherSubjects";
import type { AssignmentItem } from "../../types/assignment";
import type {
  TeacherSubmissionGroup,
  TeacherSubmissionSummary,
} from "../../types/teacherAssignment";
import type { MaterialItem } from "../../types/teacherMaterial";
import type { TeacherSubjectClass } from "../../types/teacherSubjects";
import { resolveSubjectColor } from "../../utils/color";
import { formatDate, formatDateTime } from "../../utils/date";
import { useToastStore } from "../../stores/toast";

type WorkspaceTab = "materials" | "assignments" | "submissions";

const route = useRoute();
const toast = useToastStore();
const subjectClassId = computed(() =>
  String(route.params.subjectClassId ?? ""),
);

const activeTab = ref<WorkspaceTab>("materials");
const subject = ref<TeacherSubjectClass | null>(null);
const materials = ref<MaterialItem[]>([]);
const assignments = ref<AssignmentItem[]>([]);
const submissionGroups = ref<TeacherSubmissionGroup[]>([]);
const submissionSummary = ref<TeacherSubmissionSummary | null>(null);
const loading = ref(false);
const submissionsLoading = ref(false);
const errorMessage = ref("");
const submissionsError = ref("");
const deletingAssignmentId = ref<string | null>(null);
const subjectAccentColor = computed(() =>
  subject.value
    ? resolveSubjectColor(subject.value)
    : resolveSubjectColor({ subjectClassId: subjectClassId.value }),
);

const submissionCount = computed(
  () =>
    submissionSummary.value?.submissionCount ??
    submissionGroups.value.reduce(
      (total, group) => total + group.submissionCount,
      0,
    ),
);

const visibleSubmissionGroups = computed(() =>
  submissionGroups.value.filter((group) => group.submissions.length > 0),
);

const tabs = computed(() => [
  { id: "materials" as const, label: "Materi" },
  {
    id: "assignments" as const,
    label: "Tugas",
  },
  {
    id: "submissions" as const,
    label: "Pengumpulan",
  },
]);

async function loadWorkspace() {
  loading.value = true;
  errorMessage.value = "";
  submissionsError.value = "";
  subject.value = null;
  materials.value = [];
  assignments.value = [];
  submissionGroups.value = [];
  submissionSummary.value = null;

  try {
    const subjectData = await getMyTeachingSubjectClassById(
      subjectClassId.value,
    );
    subject.value = subjectData;

    if (!subjectData) return;

    const [materialData, assignmentData] = await Promise.all([
      getSubjectMaterials(subjectClassId.value),
      getSubjectAssignments(subjectClassId.value, 1, 50),
    ]);

    materials.value = materialData.data?.data ?? [];
    assignments.value = assignmentData.data?.data ?? [];

    await loadSubmissions();
  } catch {
    errorMessage.value =
      "Ruang kerja mata pelajaran belum bisa dimuat. Coba lagi beberapa saat.";
  } finally {
    loading.value = false;
  }
}

async function loadSubmissions() {
  submissionsLoading.value = true;
  submissionsError.value = "";
  try {
    const data = await getSubjectClassSubmissions(subjectClassId.value);
    submissionGroups.value = data.assignments ?? [];
    submissionSummary.value = data.summary;
  } catch {
    submissionsError.value =
      "Pengumpulan siswa belum bisa dimuat. Coba lagi beberapa saat.";
  } finally {
    submissionsLoading.value = false;
  }
}

function materialAttachmentLabel(material: MaterialItem) {
  const count = material.attachments?.length ?? 0;
  if (count === 0) return "Tidak ada lampiran";
  return `${count} lampiran`;
}

function materialTypeLabel(type?: string) {
  if (!type) return "Materi";
  return type.toUpperCase();
}

async function handleDeleteAssignment(id: string) {
  if (
    !window.confirm(
      "Apakah anda yakin ingin menghapus tugas ini? Tugas dengan pengumpulan siswa tidak bisa dihapus.",
    )
  )
    return;
  deletingAssignmentId.value = id;
  try {
    await deleteAssignment(id);
    toast.success("Tugas berhasil dihapus.");
    const assignmentData = await getSubjectAssignments(
      subjectClassId.value,
      1,
      50,
    );
    assignments.value = assignmentData.data?.data ?? [];
  } catch (err: any) {
    const msg =
      err.response?.data?.error ||
      err.response?.data?.message ||
      "Gagal menghapus tugas.";
    toast.error(msg);
  } finally {
    deletingAssignmentId.value = null;
  }
}

onMounted(loadWorkspace);
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-[#f8f7f4]">
    <header class="border-b border-[#ebe7df] bg-white">
      <div class="px-5 py-5 sm:px-6 lg:px-8">
        <div class="flex min-w-0 items-center gap-2 text-xs text-[#6b7280]">
          <RouterLink
            to="/teacher/subjects"
            class="inline-flex shrink-0 items-center gap-1.5 transition hover:text-[#4f46e5]"
          >
            <PhArrowLeft :size="15" />
            Mata pelajaran
          </RouterLink>
          <span class="text-[#d1d5db]">/</span>
          <span class="min-w-0 truncate font-medium text-[#171322]">
            {{
              subject?.subjectName ??
              (loading ? "Memuat..." : "Ruang kerja mata pelajaran")
            }}
          </span>
        </div>

        <div
          class="mt-4 flex min-w-0 flex-col gap-4 lg:flex-row lg:items-center lg:justify-between"
        >
          <div class="flex min-w-0 items-start gap-3">
            <div
              class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl text-white"
              :style="{ backgroundColor: subjectAccentColor }"
            >
              <PhBookOpen :size="21" weight="duotone" />
            </div>
            <div class="min-w-0">
              <h1
                class="truncate text-xl font-medium text-[#171322] sm:text-2xl"
              >
                {{
                  subject?.subjectName ??
                  (loading ? "Memuat mata pelajaran..." : "Mata pelajaran")
                }}
              </h1>
              <p class="mt-1 truncate text-xs text-[#6b7280] sm:text-sm">
                {{
                  subject
                    ? [subject.className, subject.subjectCode]
                        .filter(Boolean)
                        .join(" · ")
                    : "Kelola materi, tugas, dan pengumpulan siswa."
                }}
              </p>
            </div>
          </div>

          <RouterLink
            v-if="subject"
            :to="`/teacher/subjects/${subjectClassId}/create`"
            class="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-[#4f46e5] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#4338ca] sm:w-auto"
          >
            <PhPlusCircle :size="17" weight="duotone" />
            Post di kelas ini
          </RouterLink>
        </div>
      </div>
    </header>

    <section
      class="mx-auto max-w-screen min-w-0 px-5 py-5 sm:px-6 lg:px-8 lg:py-6"
    >
      <section v-if="loading" class="space-y-4">
        <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          <div
            v-for="item in 4"
            :key="item"
            class="h-24 animate-pulse rounded-xl border border-[#ebe7df] bg-white"
          />
        </div>
        <div
          class="h-72 animate-pulse rounded-xl border border-[#ebe7df] bg-white"
        />
      </section>

      <section
        v-else-if="errorMessage"
        class="flex min-h-[55vh] items-center justify-center"
      >
        <article
          class="w-full max-w-xl rounded-xl border border-[#fecaca] bg-white p-6"
        >
          <div class="flex items-start gap-3">
            <div
              class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-[#fef2f2] text-[#dc2626]"
            >
              <PhWarningCircle :size="22" weight="duotone" />
            </div>
            <div class="min-w-0">
              <h2 class="text-base font-medium text-[#171322]">
                Ruang kerja tidak dapat dimuat
              </h2>
              <p class="mt-1 text-sm leading-6 text-[#7a7385]">
                {{ errorMessage }}
              </p>
              <button
                type="button"
                class="mt-4 rounded-lg bg-[#4f46e5] px-4 py-2 text-sm font-medium text-white transition hover:bg-[#4338ca]"
                @click="loadWorkspace"
              >
                Coba lagi
              </button>
            </div>
          </div>
        </article>
      </section>

      <section
        v-else-if="!subject"
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
            Mata pelajaran tidak ditemukan
          </h2>
          <p class="mx-auto mt-2 max-w-md text-sm leading-6 text-[#6b7280]">
            Mata pelajaran ini tidak tersedia untuk akun guru pada sekolah
            aktif.
          </p>
        </article>
      </section>

      <template v-else>
        <section class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          <article class="rounded-xl border border-[#ebe7df] bg-white p-4">
            <PhUsersThree :size="21" class="text-[#059669]" weight="duotone" />
            <p class="mt-3 text-xs text-[#7a7385]">Siswa</p>
            <p class="mt-1 text-2xl font-medium text-[#171322]">
              {{ subject.studentCount }}
            </p>
          </article>
          <article class="rounded-xl border border-[#ebe7df] bg-white p-4">
            <PhFileText :size="21" class="text-[#4f8ef7]" weight="duotone" />
            <p class="mt-3 text-xs text-[#7a7385]">Materi</p>
            <p class="mt-1 text-2xl font-medium text-[#171322]">
              {{ subject.materialCount }}
            </p>
          </article>
          <article class="rounded-xl border border-[#ebe7df] bg-white p-4">
            <PhClipboardText
              :size="21"
              class="text-[#4f46e5]"
              weight="duotone"
            />
            <p class="mt-3 text-xs text-[#7a7385]">Tugas</p>
            <p class="mt-1 text-2xl font-medium text-[#171322]">
              {{ subject.assignmentCount }}
            </p>
          </article>
          <article class="rounded-xl border border-[#ebe7df] bg-white p-4">
            <PhWarningCircle
              :size="21"
              class="text-[#ea580c]"
              weight="duotone"
            />
            <p class="mt-3 text-xs text-[#7a7385]">Perlu dinilai</p>
            <p class="mt-1 text-2xl font-medium text-[#171322]">
              {{ subject.pendingSubmissions }}
            </p>
          </article>
        </section>

        <section
          class="mt-5 min-w-0 rounded-xl border border-[#ebe7df] bg-white p-4 sm:p-5"
        >
          <div
            class="flex min-w-0 gap-1 overflow-x-auto border-b border-[#ebe7df]"
          >
            <button
              v-for="tab in tabs"
              :key="tab.id"
              type="button"
              class="inline-flex h-11 shrink-0 items-center gap-2 border-b-2 px-3 text-sm transition sm:px-4"
              :class="
                activeTab === tab.id
                  ? 'border-[#4f46e5] font-medium text-[#4f46e5]'
                  : 'border-transparent text-[#6b7280] hover:text-[#374151]'
              "
              @click="activeTab = tab.id"
            >
              {{ tab.label }}
            </button>
          </div>

          <div class="pt-4">
            <div v-if="activeTab === 'materials'" class="space-y-3">
              <div
                v-if="materials.length === 0"
                class="rounded-lg border border-[#ebe7df] bg-[#fbfaf8] p-8 text-center"
              >
                <div
                  class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-[#eef2ff] text-[#4f46e5]"
                >
                  <PhFileText class="h-6 w-6" weight="duotone" />
                </div>
                <h2 class="mt-3 text-base font-semibold text-[#171322]">
                  Belum ada materi
                </h2>
                <p class="mt-2 text-sm leading-6 text-[#6b7280]">
                  Materi yang dibuat untuk mata pelajaran ini akan tampil di
                  sini.
                </p>
              </div>

              <RouterLink
                v-for="material in materials"
                v-else
                :key="material.materialId"
                :to="{
                  name: 'teacher-material-detail',
                  params: {
                    subjectClassId: subjectClassId,
                    matId: material.materialId,
                  },
                }"
                class="group block min-w-0 rounded-lg border border-[#ebe7df] bg-[#fbfaf8] p-4 transition hover:bg-white hover:shadow-[0_8px_24px_rgba(66,55,40,0.08)]"
              >
                <div
                  class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between"
                >
                  <div class="min-w-0">
                    <p
                      class="text-[11px] font-medium uppercase tracking-wide text-[#4f46e5]"
                    >
                      {{ materialTypeLabel(material.materialType) }}
                    </p>
                    <h2
                      class="mt-1 line-clamp-2 wrap-break-word text-base font-medium text-[#171322] group-hover:text-[#4f46e5]"
                    >
                      {{ material.materialTitle }}
                    </h2>
                    <p
                      v-if="material.materialDesc"
                      class="mt-2 line-clamp-2 text-sm leading-6 text-[#7a7385]"
                    >
                      {{ material.materialDesc }}
                    </p>
                  </div>
                  <p class="shrink-0 text-xs text-[#8a8494]">
                    {{ formatDateTime(material.createdAt) }}
                  </p>
                </div>
                <div
                  class="mt-3 flex flex-wrap items-center gap-2 text-xs text-[#6b7280]"
                >
                  <span
                    class="inline-flex items-center gap-1.5 rounded-lg bg-white px-3 py-2"
                  >
                    <PhPaperclip :size="14" weight="duotone" />
                    {{ materialAttachmentLabel(material) }}
                  </span>
                  <span
                    v-if="material.creatorName"
                    class="rounded-lg bg-white px-3 py-2"
                  >
                    Oleh {{ material.creatorName }}
                  </span>
                </div>
              </RouterLink>
            </div>

            <div v-else-if="activeTab === 'assignments'" class="space-y-3">
              <div
                v-if="assignments.length === 0"
                class="rounded-lg border border-[#ebe7df] bg-[#fbfaf8] p-8 text-center"
              >
                <div
                  class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-[#eef2ff] text-[#4f46e5]"
                >
                  <PhClipboardText class="h-6 w-6" weight="duotone" />
                </div>
                <h2 class="mt-3 text-base font-semibold text-[#171322]">
                  Belum ada tugas
                </h2>
                <p class="mt-2 text-sm leading-6 text-[#6b7280]">
                  Tugas yang dibuat untuk mata pelajaran ini akan tampil di
                  sini.
                </p>
              </div>

              <article
                v-for="assignment in assignments"
                v-else
                :key="assignment.assignmentId"
                class="min-w-0 rounded-lg border border-[#ebe7df] bg-[#fbfaf8] p-4"
              >
                <div
                  class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between"
                >
                  <div class="min-w-0">
                    <p
                      class="text-[11px] font-medium uppercase tracking-wide text-[#ea580c]"
                    >
                      {{ assignment.categoryName || "Tanpa kategori" }}
                    </p>
                    <RouterLink
                      :to="{
                        name: 'teacher-assignment-detail',
                        params: {
                          subjectClassId: subjectClassId,
                          assignmentId: assignment.assignmentId,
                        },
                      }"
                      class="mt-1 block line-clamp-2 wrap-break-word text-base font-medium text-[#171322] transition hover:text-[#4f46e5]"
                    >
                      {{ assignment.assignmentTitle }}
                    </RouterLink>
                    <p
                      v-if="assignment.assignmentDescription"
                      class="mt-2 line-clamp-2 text-sm leading-6 text-[#7a7385]"
                    >
                      {{ assignment.assignmentDescription }}
                    </p>
                  </div>
                  <span
                    class="shrink-0 rounded-full px-3 py-1.5 text-xs font-medium"
                    :class="
                      assignment.allowLateSubmission
                        ? 'bg-[#eef7f2] text-[#2f7d5c]'
                        : 'bg-[#fff1ed] text-[#b86845]'
                    "
                  >
                    {{
                      assignment.allowLateSubmission
                        ? "Terlambat diizinkan"
                        : "Tidak menerima terlambat"
                    }}
                  </span>
                </div>
                <div
                  class="mt-4 flex flex-col gap-3 border-t border-[#ebe7df] pt-4 sm:flex-row sm:flex-wrap sm:items-center"
                >
                  <span
                    class="inline-flex items-center gap-1.5 text-xs text-[#6b7280]"
                  >
                    <PhCalendarBlank :size="14" weight="duotone" />
                    Tenggat {{ formatDate(assignment.deadline) }}
                  </span>
                  <span
                    class="inline-flex items-center gap-1.5 text-xs text-[#6b7280]"
                  >
                    <PhPaperclip :size="14" weight="duotone" />
                    {{ assignment.attachments?.length ?? 0 }} lampiran
                  </span>
                  <div class="flex flex-wrap items-center gap-2 sm:ml-auto">
                    <RouterLink
                      :to="{
                        name: 'teacher-assignment-edit',
                        params: {
                          subjectClassId: subjectClassId,
                          asgId: assignment.assignmentId,
                        },
                      }"
                      class="inline-flex items-center justify-center rounded-lg border border-[#ebe7df] bg-white p-2 text-[#6b7280] transition hover:border-[#4f46e5] hover:text-[#4f46e5]"
                      title="Edit Tugas"
                    >
                      <PhPencilSimple :size="16" weight="bold" />
                    </RouterLink>
                    <button
                      @click="handleDeleteAssignment(assignment.assignmentId)"
                      :disabled="
                        deletingAssignmentId === assignment.assignmentId
                      "
                      class="inline-flex items-center justify-center rounded-lg border border-[#ebe7df] bg-white p-2 text-[#6b7280] transition hover:border-[#dc2626] hover:text-[#dc2626] disabled:opacity-50"
                      title="Hapus Tugas"
                    >
                      <PhTrash :size="16" weight="bold" />
                    </button>
                    <RouterLink
                      :to="{
                        name: 'teacher-assignment-review',
                        params: { assignmentId: assignment.assignmentId },
                      }"
                      class="inline-flex items-center gap-1.5 rounded-lg bg-[#4f46e5] px-4 py-2 text-xs font-medium text-white transition hover:bg-[#4338ca]"
                    >
                      Nilai pengumpulan
                    </RouterLink>
                  </div>
                </div>
                <AttachmentPreviewList
                  v-if="assignment.attachments?.length"
                  class="mt-4"
                  :attachments="assignment.attachments"
                  :initially-expanded="false"
                />
              </article>
            </div>

            <div v-else class="space-y-3">
              <div
                v-if="submissionSummary"
                class="flex flex-wrap gap-2 text-sm"
              >
                <span class="rounded-lg bg-[#faf8f4] px-3 py-2 text-[#6b6475]">
                  {{ submissionSummary.submissionCount }} pengumpulan
                </span>
                <span class="rounded-lg bg-[#eef7f2] px-3 py-2 text-[#2f7d5c]">
                  {{ submissionSummary.gradedCount }} sudah dinilai
                </span>
                <span class="rounded-lg bg-[#fff7ed] px-3 py-2 text-[#9f6b1d]">
                  {{ submissionSummary.pendingCount }} perlu dinilai
                </span>
                <span class="rounded-lg bg-[#fff1ed] px-3 py-2 text-[#b86845]">
                  {{ submissionSummary.lateCount }} terlambat
                </span>
              </div>

              <div
                v-if="submissionsLoading"
                class="rounded-lg border border-[#ebe7df] bg-white p-5 text-sm text-[#6b6475]"
              >
                Memuat pengumpulan...
              </div>

              <div
                v-else-if="submissionsError"
                class="rounded-lg border border-[#ebe7df] bg-white p-5"
              >
                <h2 class="text-lg font-medium text-[#171322]">
                  Pengumpulan gagal dimuat
                </h2>
                <p class="mt-2 text-sm leading-6 text-[#6b6475]">
                  {{ submissionsError }}
                </p>
              </div>

              <div
                v-else-if="submissionCount === 0"
                class="rounded-lg border border-[#ebe7df] bg-white p-8 text-center"
              >
                <div
                  class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-[#eef2ff] text-[#4f46e5]"
                >
                  <PhCheckCircle class="h-6 w-6" weight="duotone" />
                </div>
                <h2 class="mt-3 text-base font-semibold text-[#171322]">
                  Belum ada pengumpulan
                </h2>
                <p class="mt-2 text-sm leading-6 text-[#6b7280]">
                  Pengumpulan siswa akan tampil setelah ada tugas yang
                  dikumpulkan pada mata pelajaran ini.
                </p>
              </div>

              <article
                v-for="group in visibleSubmissionGroups"
                v-else
                :key="group.assignment.assignmentId"
                class="rounded-lg border border-[#ebe7df] bg-white p-5"
              >
                <div
                  class="flex flex-col gap-3 border-b border-[#ece8df] pb-4 sm:flex-row sm:items-start sm:justify-between"
                >
                  <div>
                    <p
                      class="text-xs font-medium uppercase tracking-wide text-[#7b61a8]"
                    >
                      {{ group.assignment.categoryName || "Tugas" }}
                    </p>
                    <h2 class="mt-2 text-lg font-medium text-[#171322]">
                      {{ group.assignment.assignmentTitle }}
                    </h2>
                    <p
                      v-if="group.assignment.deadline"
                      class="mt-2 text-sm text-[#6b6475]"
                    >
                      Tenggat {{ formatDate(group.assignment.deadline) }}
                    </p>
                  </div>
                  <div class="flex flex-col items-start gap-2 sm:items-end">
                    <div class="flex flex-wrap gap-2 text-xs font-medium">
                      <span
                        class="rounded-lg bg-[#faf8f4] px-3 py-2 text-[#6b6475]"
                      >
                        {{ group.submissionCount }} pengumpulan
                      </span>
                      <span
                        class="rounded-lg bg-[#eef7f2] px-3 py-2 text-[#2f7d5c]"
                      >
                        {{ group.gradedCount }} dinilai
                      </span>
                      <span
                        class="rounded-lg bg-[#fff7ed] px-3 py-2 text-[#9f6b1d]"
                      >
                        {{ group.pendingCount }} perlu dinilai
                      </span>
                    </div>
                    <RouterLink
                      v-if="
                        group.submissionCount > 0 ||
                        group.submissions.length > 0
                      "
                      :to="{
                        name: 'teacher-assignment-review',
                        params: { assignmentId: group.assignment.assignmentId },
                      }"
                      class="inline-flex items-center gap-2 rounded-lg bg-[#171322] px-4 py-2 text-sm font-medium text-white transition hover:bg-[#2f2b3a]"
                    >
                      Nilai pengumpulan
                    </RouterLink>
                  </div>
                </div>

                <div class="mt-4 space-y-3">
                  <div
                    v-for="submission in group.submissions"
                    :key="submission.submissionId"
                    class="flex flex-col gap-3 rounded-lg border border-[#ebe7df] bg-[#faf8f4] px-4 py-3 sm:flex-row sm:items-center sm:justify-between"
                  >
                    <div class="min-w-0">
                      <h3 class="truncate text-sm font-medium text-[#171322]">
                        {{ submission.studentName }}
                      </h3>
                      <p class="mt-1 text-xs text-[#6b6475]">
                        Dikumpulkan {{ formatDateTime(submission.submittedAt) }}
                      </p>
                    </div>
                    <div
                      class="flex flex-wrap items-center gap-2 text-xs font-medium text-[#6b6475]"
                    >
                      <span class="rounded-lg bg-white px-3 py-2">
                        {{ submission.attachments?.length ?? 0 }} lampiran
                      </span>
                      <span
                        v-if="submission.isLate"
                        class="rounded-lg bg-[#fff1ed] px-3 py-2 text-[#b86845]"
                      >
                        Terlambat
                      </span>
                      <span
                        v-if="submission.assessment"
                        class="rounded-lg bg-white px-3 py-2"
                      >
                        Nilai {{ submission.assessment.score }}
                      </span>
                      <span
                        class="shrink-0 rounded-lg px-3 py-2 text-xs font-medium"
                        :class="
                          submission.assessment
                            ? 'bg-[#eef7f2] text-[#2f7d5c]'
                            : 'bg-[#fff7ed] text-[#9f6b1d]'
                        "
                      >
                        {{
                          submission.assessment
                            ? "Sudah dinilai"
                            : "Menunggu penilaian"
                        }}
                      </span>
                    </div>
                  </div>
                </div>
              </article>
            </div>
          </div>
        </section>
      </template>
    </section>
  </main>
</template>
