<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import {
  PhArrowLeft,
  PhBookOpen,
  PhCalendarBlank,
  PhCheckCircle,
  PhClipboardText,
  PhFileText,
  PhPaperclip,
  PhUsersThree,
  PhWarningCircle,
} from '@phosphor-icons/vue'
import { getSubjectAssignments } from '../../services/assignment'
import { getAssignmentDetailWithSubmissions } from '../../services/teacherAssignment'
import { getSubjectMaterials } from '../../services/teacherMaterial'
import { getMyTeachingSubjectClassById } from '../../services/teacherSubjects'
import type { AssignmentItem } from '../../types/assignment'
import type { TeacherSubmission } from '../../types/teacherAssignment'
import type { MaterialItem } from '../../types/teacherMaterial'
import type { TeacherSubjectClass } from '../../types/teacherSubjects'
import { getSubjectColor } from '../../utils/color'
import { formatDate, formatDateTime } from '../../utils/date'

type WorkspaceTab = 'materials' | 'assignments' | 'submissions'

interface SubmissionRow extends TeacherSubmission {
  assignmentId: string
  assignmentTitle: string
  categoryName?: string
}

const route = useRoute()
const subjectClassId = computed(() => String(route.params.subjectClassId ?? ''))

const activeTab = ref<WorkspaceTab>('materials')
const subject = ref<TeacherSubjectClass | null>(null)
const materials = ref<MaterialItem[]>([])
const assignments = ref<AssignmentItem[]>([])
const submissions = ref<SubmissionRow[]>([])
const loading = ref(false)
const submissionsLoading = ref(false)
const errorMessage = ref('')
const submissionsError = ref('')

const tabs = computed(() => [
  { id: 'materials' as const, label: 'Materials', count: materials.value.length },
  { id: 'assignments' as const, label: 'Assignments', count: assignments.value.length },
  { id: 'submissions' as const, label: 'Submissions', count: submissions.value.length },
])

async function loadWorkspace() {
  loading.value = true
  errorMessage.value = ''
  submissionsError.value = ''
  subject.value = null
  materials.value = []
  assignments.value = []
  submissions.value = []

  try {
    const subjectData = await getMyTeachingSubjectClassById(subjectClassId.value)
    subject.value = subjectData

    if (!subjectData) return

    const [materialData, assignmentData] = await Promise.all([
      getSubjectMaterials(subjectClassId.value),
      getSubjectAssignments(subjectClassId.value, 1, 50),
    ])

    materials.value = materialData.data?.data ?? []
    assignments.value = assignmentData.data?.data ?? []

    await loadSubmissions(assignments.value)
  } catch {
    errorMessage.value = 'Workspace subject belum bisa dimuat. Coba lagi beberapa saat.'
  } finally {
    loading.value = false
  }
}

async function loadSubmissions(subjectAssignments: AssignmentItem[]) {
  if (subjectAssignments.length === 0) return

  submissionsLoading.value = true
  submissionsError.value = ''
  try {
    const details = await Promise.all(
      subjectAssignments.map((assignment) =>
        getAssignmentDetailWithSubmissions(assignment.assignmentId),
      ),
    )

    submissions.value = details.flatMap((detail) =>
      detail.submissions.map((submission) => ({
        ...submission,
        assignmentId: detail.assignment.assignmentId,
        assignmentTitle: detail.assignment.assignmentTitle,
        categoryName: detail.assignment.categoryName,
      })),
    )
  } catch {
    submissionsError.value =
      'Submission belum bisa dimuat dari detail assignment. Endpoint agregat subject-class submissions belum tersedia.'
  } finally {
    submissionsLoading.value = false
  }
}

function materialAttachmentLabel(material: MaterialItem) {
  const count = material.attachments?.length ?? 0
  if (count === 0) return 'Tidak ada attachment'
  return `${count} attachment`
}

function materialTypeLabel(type?: string) {
  if (!type) return 'Materi'
  return type.toUpperCase()
}

onMounted(loadWorkspace)
</script>

<template>
  <main class="min-h-screen flex-1 px-5 py-8 md:px-8 lg:px-10">
    <section class="mx-auto flex max-w-6xl flex-col gap-6">
      <RouterLink
        to="/teacher/subjects"
        class="inline-flex items-center gap-2 self-start text-sm font-medium text-[#6b6475] transition hover:text-[#171322]"
      >
        <PhArrowLeft :size="18" />
        Kembali ke subjects
      </RouterLink>

      <header class="rounded-[32px] bg-white p-6 shadow-sm ring-1 ring-black/5 md:p-8">
        <div
          class="mb-5 flex h-14 w-14 items-center justify-center rounded-2xl text-white"
          :style="{ backgroundColor: getSubjectColor(subjectClassId) }"
        >
          <PhBookOpen :size="28" weight="duotone" />
        </div>
        <p class="text-sm font-medium text-[#7b61a8]">Subject workspace</p>
        <h1 class="mt-3 text-3xl font-medium text-[#171322]">
          {{ subject?.subjectName ?? (loading ? 'Memuat subject...' : 'Workspace subject') }}
        </h1>
        <p class="mt-3 max-w-2xl text-sm leading-6 text-[#6b6475]">
          <span v-if="subject">
            {{ subject.className }} menjadi konteks class untuk subject ini. Material dan tugas
            berada di level subject class.
          </span>
          <span v-else>
            Detail subject class mengambil data dari endpoint current teacher agar guru hanya melihat
            subject yang dia ampu.
          </span>
        </p>
      </header>

      <section v-if="loading" class="rounded-[28px] bg-white p-6 shadow-sm ring-1 ring-black/5">
        <p class="text-sm text-[#6b6475]">Memuat workspace subject...</p>
      </section>

      <section v-else-if="errorMessage" class="rounded-[28px] bg-white p-6 shadow-sm ring-1 ring-black/5">
        <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div class="flex items-start gap-3">
            <PhWarningCircle :size="24" class="mt-0.5 text-[#e58f86]" weight="duotone" />
            <div>
              <h2 class="text-lg font-medium text-[#171322]">Gagal memuat workspace</h2>
              <p class="mt-2 text-sm leading-6 text-[#6b6475]">{{ errorMessage }}</p>
            </div>
          </div>
          <button
            type="button"
            class="rounded-2xl bg-[#171322] px-4 py-3 text-sm font-medium text-white"
            @click="loadWorkspace"
          >
            Coba lagi
          </button>
        </div>
      </section>

      <section
        v-else-if="!subject"
        class="rounded-[28px] bg-white p-6 shadow-sm ring-1 ring-black/5"
      >
        <h2 class="text-lg font-medium text-[#171322]">Subject tidak ditemukan</h2>
        <p class="mt-2 text-sm leading-6 text-[#6b6475]">
          Subject class ini tidak tersedia untuk akun guru pada school aktif.
        </p>
      </section>

      <template v-else>
        <section class="grid gap-4 md:grid-cols-4">
          <article class="rounded-[24px] bg-white p-5 shadow-sm ring-1 ring-black/5">
            <PhUsersThree :size="24" class="text-[#74bfa5]" weight="duotone" />
            <p class="mt-4 text-sm text-[#8a8494]">Siswa</p>
            <p class="mt-1 text-2xl font-medium text-[#171322]">{{ subject.studentCount }}</p>
          </article>
          <article class="rounded-[24px] bg-white p-5 shadow-sm ring-1 ring-black/5">
            <PhFileText :size="24" class="text-[#7aa7d9]" weight="duotone" />
            <p class="mt-4 text-sm text-[#8a8494]">Materi</p>
            <p class="mt-1 text-2xl font-medium text-[#171322]">{{ subject.materialCount }}</p>
          </article>
          <article class="rounded-[24px] bg-white p-5 shadow-sm ring-1 ring-black/5">
            <PhClipboardText :size="24" class="text-[#e58f86]" weight="duotone" />
            <p class="mt-4 text-sm text-[#8a8494]">Tugas</p>
            <p class="mt-1 text-2xl font-medium text-[#171322]">{{ subject.assignmentCount }}</p>
          </article>
          <article class="rounded-[24px] bg-white p-5 shadow-sm ring-1 ring-black/5">
            <PhWarningCircle :size="24" class="text-[#b889c9]" weight="duotone" />
            <p class="mt-4 text-sm text-[#8a8494]">Perlu review</p>
            <p class="mt-1 text-2xl font-medium text-[#171322]">{{ subject.pendingSubmissions }}</p>
          </article>
        </section>

        <section class="rounded-[32px] bg-white p-5 shadow-sm ring-1 ring-black/5 md:p-6">
          <div class="flex flex-wrap gap-2 border-b border-[#ece8df] pb-4">
            <button
              v-for="tab in tabs"
              :key="tab.id"
              type="button"
              class="rounded-2xl px-4 py-2.5 text-sm font-medium transition"
              :class="
                activeTab === tab.id
                  ? 'bg-[#171322] text-white'
                  : 'bg-[#faf8f4] text-[#6b6475] hover:bg-[#f0e9dd] hover:text-[#171322]'
              "
              @click="activeTab = tab.id"
            >
              {{ tab.label }}
              <span class="ml-2 opacity-70">{{ tab.count }}</span>
            </button>
          </div>

          <div class="pt-5">
            <div v-if="activeTab === 'materials'" class="space-y-3">
              <div
                v-if="materials.length === 0"
                class="rounded-[24px] bg-[#faf8f4] p-6 text-center"
              >
                <PhFileText :size="30" class="mx-auto text-[#b5afbf]" weight="duotone" />
                <h2 class="mt-3 text-lg font-medium text-[#171322]">Belum ada materi</h2>
                <p class="mt-2 text-sm leading-6 text-[#6b6475]">
                  Materi yang dibuat untuk subject class ini akan tampil di sini.
                </p>
              </div>

              <article
                v-for="material in materials"
                v-else
                :key="material.materialId"
                class="rounded-[24px] bg-[#faf8f4] p-5 ring-1 ring-black/5"
              >
                <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                  <div>
                    <p class="text-xs font-medium uppercase tracking-wide text-[#7aa7d9]">
                      {{ materialTypeLabel(material.materialType) }}
                    </p>
                    <h2 class="mt-2 text-lg font-medium text-[#171322]">
                      {{ material.materialTitle }}
                    </h2>
                    <p
                      v-if="material.materialDesc"
                      class="mt-2 line-clamp-2 text-sm leading-6 text-[#6b6475]"
                    >
                      {{ material.materialDesc }}
                    </p>
                  </div>
                  <p class="shrink-0 text-sm text-[#8a8494]">
                    {{ formatDateTime(material.createdAt) }}
                  </p>
                </div>
                <div class="mt-4 flex flex-wrap items-center gap-3 text-sm text-[#6b6475]">
                  <span class="inline-flex items-center gap-2 rounded-2xl bg-white px-3 py-2">
                    <PhPaperclip :size="16" weight="duotone" />
                    {{ materialAttachmentLabel(material) }}
                  </span>
                  <span v-if="material.creatorName" class="rounded-2xl bg-white px-3 py-2">
                    Oleh {{ material.creatorName }}
                  </span>
                </div>
              </article>
            </div>

            <div v-else-if="activeTab === 'assignments'" class="space-y-3">
              <div
                v-if="assignments.length === 0"
                class="rounded-[24px] bg-[#faf8f4] p-6 text-center"
              >
                <PhClipboardText :size="30" class="mx-auto text-[#b5afbf]" weight="duotone" />
                <h2 class="mt-3 text-lg font-medium text-[#171322]">Belum ada tugas</h2>
                <p class="mt-2 text-sm leading-6 text-[#6b6475]">
                  Tugas yang dibuat untuk subject class ini akan tampil di sini.
                </p>
              </div>

              <article
                v-for="assignment in assignments"
                v-else
                :key="assignment.assignmentId"
                class="rounded-[24px] bg-[#faf8f4] p-5 ring-1 ring-black/5"
              >
                <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                  <div>
                    <p class="text-xs font-medium uppercase tracking-wide text-[#e58f86]">
                      {{ assignment.categoryName || 'Tanpa kategori' }}
                    </p>
                    <h2 class="mt-2 text-lg font-medium text-[#171322]">
                      {{ assignment.assignmentTitle }}
                    </h2>
                    <p
                      v-if="assignment.assignmentDescription"
                      class="mt-2 line-clamp-2 text-sm leading-6 text-[#6b6475]"
                    >
                      {{ assignment.assignmentDescription }}
                    </p>
                  </div>
                  <span
                    class="shrink-0 rounded-2xl px-3 py-2 text-xs font-medium"
                    :class="
                      assignment.allowLateSubmission
                        ? 'bg-[#eef7f2] text-[#2f7d5c]'
                        : 'bg-[#fff1ed] text-[#b86845]'
                    "
                  >
                    {{ assignment.allowLateSubmission ? 'Late allowed' : 'No late submit' }}
                  </span>
                </div>
                <div class="mt-4 flex flex-wrap items-center gap-3 text-sm text-[#6b6475]">
                  <span class="inline-flex items-center gap-2 rounded-2xl bg-white px-3 py-2">
                    <PhCalendarBlank :size="16" weight="duotone" />
                    Deadline {{ formatDate(assignment.deadline) }}
                  </span>
                  <span class="inline-flex items-center gap-2 rounded-2xl bg-white px-3 py-2">
                    <PhPaperclip :size="16" weight="duotone" />
                    {{ assignment.attachments?.length ?? 0 }} attachment
                  </span>
                </div>
              </article>
            </div>

            <div v-else class="space-y-3">
              <div class="rounded-[24px] bg-[#faf8f4] p-5 text-sm leading-6 text-[#6b6475]">
                Submission dibaca dari endpoint detail assignment per tugas. Endpoint agregat
                `/assignments/subject-class/:subjectClassId/submissions` belum tersedia, jadi halaman
                ini tetap read-only dan tidak menampilkan data buatan.
              </div>

              <div
                v-if="submissionsLoading"
                class="rounded-[24px] bg-white p-6 text-sm text-[#6b6475] ring-1 ring-black/5"
              >
                Memuat submissions...
              </div>

              <div
                v-else-if="submissionsError"
                class="rounded-[24px] bg-white p-6 ring-1 ring-black/5"
              >
                <h2 class="text-lg font-medium text-[#171322]">Submission gagal dimuat</h2>
                <p class="mt-2 text-sm leading-6 text-[#6b6475]">{{ submissionsError }}</p>
              </div>

              <div
                v-else-if="submissions.length === 0"
                class="rounded-[24px] bg-white p-6 text-center ring-1 ring-black/5"
              >
                <PhCheckCircle :size="30" class="mx-auto text-[#b5afbf]" weight="duotone" />
                <h2 class="mt-3 text-lg font-medium text-[#171322]">Belum ada submission</h2>
                <p class="mt-2 text-sm leading-6 text-[#6b6475]">
                  Submission siswa akan tampil setelah ada pengumpulan tugas pada subject ini.
                </p>
              </div>

              <article
                v-for="submission in submissions"
                v-else
                :key="submission.submissionId"
                class="rounded-[24px] bg-white p-5 ring-1 ring-black/5"
              >
                <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                  <div>
                    <p class="text-xs font-medium uppercase tracking-wide text-[#7b61a8]">
                      {{ submission.assignmentTitle }}
                    </p>
                    <h2 class="mt-2 text-lg font-medium text-[#171322]">
                      {{ submission.studentName }}
                    </h2>
                    <p class="mt-2 text-sm text-[#6b6475]">
                      Dikumpulkan {{ formatDateTime(submission.submittedAt) }}
                    </p>
                  </div>
                  <span
                    class="shrink-0 rounded-2xl px-3 py-2 text-xs font-medium"
                    :class="
                      submission.assessment
                        ? 'bg-[#eef7f2] text-[#2f7d5c]'
                        : 'bg-[#fff7e8] text-[#9f6b1d]'
                    "
                  >
                    {{ submission.assessment ? 'Sudah dinilai' : 'Menunggu penilaian' }}
                  </span>
                </div>
                <div class="mt-4 flex flex-wrap items-center gap-3 text-sm text-[#6b6475]">
                  <span class="rounded-2xl bg-[#faf8f4] px-3 py-2">
                    {{ submission.attachments?.length ?? 0 }} file
                  </span>
                  <span v-if="submission.isLate" class="rounded-2xl bg-[#fff1ed] px-3 py-2 text-[#b86845]">
                    Terlambat
                  </span>
                  <span v-if="submission.assessment" class="rounded-2xl bg-[#faf8f4] px-3 py-2">
                    Nilai {{ submission.assessment.score }}
                  </span>
                </div>
              </article>
            </div>
          </div>
        </section>
      </template>
    </section>
  </main>
</template>
