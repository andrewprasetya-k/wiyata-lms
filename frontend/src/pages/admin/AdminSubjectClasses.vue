<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import {
  PhBookOpen,
  PhCalendarBlank,
  PhChalkboardTeacher,
  PhLinkSimple,
  PhTrash,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import {
  getAcademicYearsBySchool,
  getSubjectsBySchool,
  getTermsByAcademicYear,
} from "../../services/adminAcademic";
import { getAdminClasses } from "../../services/adminClass";
import { getClassEnrollments } from "../../services/adminEnrollment";
import { getSchoolMembers } from "../../services/adminUser";
import {
  assignSubjectClass,
  deleteSubjectClass,
  getSubjectClassesByClass,
} from "../../services/adminSubjectClass";
import type {
  AcademicYearItem,
  SubjectItem,
  TermItem,
} from "../../types/adminAcademic";
import type { AdminClassItem } from "../../types/adminClass";
import type { EnrollmentMemberItem } from "../../types/adminEnrollment";
import type { SchoolMemberItem } from "../../types/adminUser";
import type { SubjectClassItem } from "../../types/adminSubjectClass";
import { useToastStore } from "../../stores/toast";

const auth = useAuthStore();
const toast = useToastStore();

const currentSchool = computed(() => {
  const activeId = auth.activeSchoolId ?? auth.defaultContext?.schoolId ?? null;
  const current =
    auth.memberships.find((membership) => membership.school.id === activeId) ??
    auth.memberships.find((membership) => membership.isDefault) ??
    auth.memberships[0] ??
    null;

  return {
    schoolId: activeId ?? "",
    schoolCode: current?.school.code ?? "",
    schoolName: current?.school.name ?? "",
    hasContext: Boolean(activeId && current?.school.code),
  };
});

const academicYears = ref<AcademicYearItem[]>([]);
const terms = ref<TermItem[]>([]);
const classes = ref<AdminClassItem[]>([]);
const subjects = ref<SubjectItem[]>([]);
const members = ref<SchoolMemberItem[]>([]);
const enrollments = ref<EnrollmentMemberItem[]>([]);
const subjectClasses = ref<SubjectClassItem[]>([]);

const selectedAcademicYearId = ref("");
const selectedTermId = ref("");
const selectedClassId = ref("");
const selectedSubjectId = ref("");
const selectedTeacherSchoolUserId = ref("");
const pendingUnassign = ref<SubjectClassItem | null>(null);

const yearsLoading = ref(false);
const termsLoading = ref(false);
const classesLoading = ref(false);
const subjectsLoading = ref(false);
const membersLoading = ref(false);
const enrollmentsLoading = ref(false);
const subjectClassesLoading = ref(false);
const submitting = ref(false);
const unassigningId = ref("");

const yearsError = ref("");
const termsError = ref("");
const classesError = ref("");
const subjectsError = ref("");
const membersError = ref("");
const enrollmentsError = ref("");
const subjectClassesError = ref("");

const selectedAcademicYear = computed(
  () =>
    academicYears.value.find(
      (year) => year.academicYearId === selectedAcademicYearId.value,
    ) ?? null,
);

const selectedTerm = computed(
  () =>
    terms.value.find((term) => term.termId === selectedTermId.value) ?? null,
);

const selectedClass = computed(
  () =>
    classes.value.find(
      (classItem) => classItem.classId === selectedClassId.value,
    ) ?? null,
);

const assignedSubjectIds = computed(
  () =>
    new Set(subjectClasses.value.map((subjectClass) => subjectClass.subjectId)),
);

const availableSubjects = computed(() =>
  subjects.value.filter(
    (subject) => !assignedSubjectIds.value.has(subject.subjectId),
  ),
);

const teacherEnrollmentIds = computed(
  () =>
    new Set(
      enrollments.value
        .filter((enrollment) => enrollment.role === "teacher")
        .map((enrollment) => enrollment.schoolUserId),
    ),
);

const teacherCandidates = computed(() =>
  members.value.filter(
    (member) =>
      teacherEnrollmentIds.value.has(member.schoolUserId) &&
      (member.roles ?? []).some(
        (role) => normalizeRoleName(role) === "teacher",
      ),
  ),
);

const selectedSubject = computed(
  () =>
    subjects.value.find(
      (subject) => subject.subjectId === selectedSubjectId.value,
    ) ?? null,
);

const selectedTeacher = computed(
  () =>
    teacherCandidates.value.find(
      (member) => member.schoolUserId === selectedTeacherSchoolUserId.value,
    ) ?? null,
);

function normalizeRoleName(roleName: string) {
  return roleName.trim().toLowerCase();
}

function getErrorMessage(error: unknown) {
  if (
    typeof error === "object" &&
    error !== null &&
    "response" in error &&
    typeof (error as { response?: { data?: { error?: unknown } } }).response
      ?.data?.error === "string"
  ) {
    return (error as { response: { data: { error: string } } }).response.data
      .error;
  }

  return "Penugasan mengajar belum bisa dilepas.";
}

function resetAssignmentForm() {
  if (
    selectedSubjectId.value &&
    assignedSubjectIds.value.has(selectedSubjectId.value)
  ) {
    selectedSubjectId.value = "";
  }
  if (
    selectedTeacherSchoolUserId.value &&
    !teacherCandidates.value.some(
      (teacher) => teacher.schoolUserId === selectedTeacherSchoolUserId.value,
    )
  ) {
    selectedTeacherSchoolUserId.value = "";
  }
}

async function loadAcademicYears() {
  if (!currentSchool.value.hasContext) return;
  yearsLoading.value = true;
  yearsError.value = "";

  try {
    const data = await getAcademicYearsBySchool(currentSchool.value.schoolCode);
    academicYears.value = data.data ?? [];
    const defaultYear =
      academicYears.value.find((year) => year.isActive) ??
      academicYears.value[0] ??
      null;
    selectedAcademicYearId.value = defaultYear?.academicYearId ?? "";
  } catch {
    yearsError.value = "Tahun ajaran belum bisa dimuat.";
  } finally {
    yearsLoading.value = false;
  }
}

async function loadTerms(selectDefault = false) {
  terms.value = [];
  termsError.value = "";
  if (selectDefault) selectedTermId.value = "";

  if (!selectedAcademicYearId.value) return;

  termsLoading.value = true;
  try {
    const data = await getTermsByAcademicYear(selectedAcademicYearId.value);
    terms.value = data ?? [];

    const selectedStillValid = terms.value.some(
      (term) => term.termId === selectedTermId.value,
    );
    if (selectDefault || !selectedStillValid) {
      const defaultTerm =
        terms.value.find((term) => term.isActive) ?? terms.value[0] ?? null;
      selectedTermId.value = defaultTerm?.termId ?? "";
    }
  } catch {
    termsError.value = "Semester belum bisa dimuat.";
  } finally {
    termsLoading.value = false;
  }
}

async function loadClasses(selectDefault = false) {
  classes.value = [];
  classesError.value = "";
  if (selectDefault) selectedClassId.value = "";

  if (!currentSchool.value.hasContext || !selectedTermId.value) return;

  classesLoading.value = true;
  try {
    const data = await getAdminClasses({
      schoolCode: currentSchool.value.schoolCode,
      termId: selectedTermId.value,
      page: 1,
      limit: 50,
    });
    classes.value = data.data?.data ?? [];

    const selectedStillValid = classes.value.some(
      (classItem) => classItem.classId === selectedClassId.value,
    );
    if (selectDefault || !selectedStillValid) {
      selectedClassId.value = classes.value[0]?.classId ?? "";
    }
  } catch {
    classesError.value = "Daftar kelas belum bisa dimuat.";
  } finally {
    classesLoading.value = false;
  }
}

async function loadSubjects() {
  subjects.value = [];
  subjectsError.value = "";

  if (!currentSchool.value.hasContext) return;

  subjectsLoading.value = true;
  try {
    const data = await getSubjectsBySchool(currentSchool.value.schoolCode);
    subjects.value = data.subjects ?? [];
    resetAssignmentForm();
  } catch {
    subjectsError.value = "Mata pelajaran belum bisa dimuat.";
  } finally {
    subjectsLoading.value = false;
  }
}

async function loadMembers() {
  members.value = [];
  membersError.value = "";

  if (!currentSchool.value.hasContext) return;

  membersLoading.value = true;
  try {
    const data = await getSchoolMembers(currentSchool.value.schoolCode, {
      page: 1,
      limit: 50,
    });
    members.value = data.members?.data ?? [];
    resetAssignmentForm();
  } catch {
    membersError.value = "Warga sekolah belum bisa dimuat.";
  } finally {
    membersLoading.value = false;
  }
}

async function loadClassContext() {
  enrollments.value = [];
  subjectClasses.value = [];
  enrollmentsError.value = "";
  subjectClassesError.value = "";
  selectedSubjectId.value = "";
  selectedTeacherSchoolUserId.value = "";
  pendingUnassign.value = null;

  if (!selectedClassId.value) return;

  enrollmentsLoading.value = true;
  subjectClassesLoading.value = true;

  try {
    const [enrollmentData, subjectClassData] = await Promise.all([
      getClassEnrollments(selectedClassId.value, { page: 1, limit: 50 }),
      getSubjectClassesByClass(selectedClassId.value),
    ]);

    enrollments.value = enrollmentData.members?.data ?? [];
    subjectClasses.value = subjectClassData.subjects ?? [];
    resetAssignmentForm();
  } catch {
    enrollmentsError.value = "Penempatan kelas belum bisa dimuat.";
    subjectClassesError.value = "Penugasan mengajar belum bisa dimuat.";
  } finally {
    enrollmentsLoading.value = false;
    subjectClassesLoading.value = false;
  }
}

async function handleAcademicYearChange() {
  await loadTerms(true);
  await loadClasses(true);
  await loadClassContext();
}

async function handleTermChange() {
  await loadClasses(true);
  await loadClassContext();
}

async function handleClassChange() {
  await loadClassContext();
}

async function submitSubjectClass() {
  if (!selectedClassId.value) {
    toast.error("Pilih kelas terlebih dahulu.");
    return;
  }
  if (!selectedSubjectId.value) {
    toast.error("Pilih mata pelajaran yang belum ditugaskan.");
    return;
  }
  if (teacherCandidates.value.length === 0) {
    toast.error(
      "Belum ada guru yang memenuhi syarat. Atur peran guru di Warga Sekolah dan tempatkan guru di kelas terlebih dahulu.",
    );
    return;
  }
  if (!selectedTeacherSchoolUserId.value) {
    toast.error("Pilih guru yang akan mengampu mata pelajaran.");
    return;
  }
  if (assignedSubjectIds.value.has(selectedSubjectId.value)) {
    toast.info(
      "Mata pelajaran ini sudah memiliki penugasan di kelas terpilih.",
    );
    return;
  }

  submitting.value = true;
  try {
    await assignSubjectClass({
      classId: selectedClassId.value,
      subjectId: selectedSubjectId.value,
      teacherId: selectedTeacherSchoolUserId.value,
    });
    toast.success("Penugasan mengajar berhasil dibuat.");
    await loadClassContext();
  } catch {
    toast.error(
      "Penugasan mengajar belum bisa dibuat. Periksa penempatan guru dan mata pelajaran yang dipilih.",
    );
  } finally {
    submitting.value = false;
  }
}

function requestUnassign(subjectClass: SubjectClassItem) {
  pendingUnassign.value = subjectClass;
}

function cancelUnassign() {
  pendingUnassign.value = null;
}

async function confirmUnassign(subjectClass: SubjectClassItem) {
  if (!subjectClass.subjectClassId || unassigningId.value) return;

  unassigningId.value = subjectClass.subjectClassId;
  try {
    await deleteSubjectClass(subjectClass.subjectClassId);
    toast.success("Penugasan mengajar berhasil dilepas.");
    pendingUnassign.value = null;
    await loadClassContext();
  } catch (error) {
    toast.error(getErrorMessage(error));
  } finally {
    unassigningId.value = "";
  }
}

onMounted(async () => {
  if (!currentSchool.value.hasContext) return;
  await Promise.all([loadAcademicYears(), loadSubjects(), loadMembers()]);
  await loadTerms(true);
  await loadClasses(true);
  await loadClassContext();
});
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-[#f8f7f4]">
    <header class="border-b border-[#ebe7df] bg-white">
      <div
        class="flex min-w-0 flex-col gap-3 px-5 py-5 sm:px-6 lg:flex-row lg:items-end lg:justify-between lg:px-8"
      >
        <div class="min-w-0">
          <h1 class="mt-1 text-2xl font-semibold text-[#171322] sm:text-3xl">
            Penugasan Mengajar
          </h1>
          <p class="mt-2 max-w-3xl text-sm leading-6 text-[#6b7280]">
            Hubungkan guru, kelas, dan mata pelajaran untuk menyiapkan ruang
            mengajar yang dapat digunakan.
          </p>
        </div>
        <div class="flex min-w-0 flex-wrap gap-2 text-xs">
          <span
            class="max-w-full truncate rounded-lg bg-[#fff4ee] px-3 py-2 font-medium text-[#ea580c]"
          >
            {{ currentSchool.schoolName || "Sekolah belum tersedia" }}
          </span>
          <span
            class="rounded-lg bg-[#f3f4f6] px-3 py-2 font-medium text-[#6b7280]"
          >
            {{ currentSchool.schoolCode || "Kode belum tersedia" }}
          </span>
        </div>
      </div>
    </header>

    <section class="px-5 py-5 sm:px-6 lg:px-8">
      <div
        v-if="!currentSchool.hasContext"
        class="mb-5 flex items-start gap-3 rounded-xl border border-[#fecaca] bg-[#fef2f2] p-4 text-sm leading-6 text-[#dc2626]"
      >
        <PhWarningCircle :size="20" class="mt-0.5 shrink-0" weight="duotone" />
        <p>
          Konteks sekolah aktif belum tersedia. Pastikan akun admin terhubung
          dengan sekolah.
        </p>
      </div>

      <div class="grid min-w-0 gap-5 lg:grid-cols-[minmax(0,1fr)_380px]">
        <section
          class="order-2 min-w-0 rounded-xl border border-[#ebe7df] bg-white lg:order-1"
        >
          <div
            class="flex flex-col gap-3 border-b border-[#ebe7df] px-4 py-4 sm:flex-row sm:items-start sm:justify-between sm:px-5"
          >
            <div class="min-w-0">
              <p
                class="text-[10px] font-medium uppercase tracking-[0.08em] text-[#9ca3af]"
              >
                Penugasan aktif
              </p>
              <h2 class="mt-1 text-base font-semibold text-[#171322]">
                {{ selectedClass?.classTitle || "Pilih kelas" }}
              </h2>
              <p class="mt-1 text-sm leading-6 text-[#6b7280]">
                {{
                  selectedClass
                    ? `${selectedTerm?.termName || "Semester"} · ${selectedAcademicYear?.academicYearName || "Tahun ajaran"}`
                    : "Pilih konteks kelas pada panel untuk melihat penugasan mengajar."
                }}
              </p>
            </div>
            <div class="flex shrink-0 flex-wrap gap-2 text-xs font-medium">
              <span
                class="inline-flex items-center gap-2 rounded-lg bg-[#eef2ff] px-3 py-2 text-[#4f46e5]"
              >
                <PhBookOpen :size="16" weight="duotone" />
                {{ subjectClasses.length }} mata pelajaran
              </span>
              <span
                class="inline-flex items-center gap-2 rounded-lg bg-[#ecfdf5] px-3 py-2 text-[#059669]"
              >
                <PhChalkboardTeacher :size="16" weight="duotone" />
                {{ teacherCandidates.length }} guru tersedia
              </span>
            </div>
          </div>

          <div class="p-4 sm:p-5">
            <div v-if="subjectClassesLoading" class="space-y-3">
              <div
                v-for="item in 3"
                :key="item"
                class="h-24 animate-pulse rounded-lg bg-[#fbfaf8]"
              />
            </div>

            <div
              v-else-if="subjectClassesError"
              class="rounded-lg border border-[#fecaca] bg-[#fef2f2] p-5 text-center"
            >
              <PhWarningCircle
                :size="26"
                class="mx-auto text-[#dc2626]"
                weight="duotone"
              />
              <h3 class="mt-3 text-sm font-semibold text-[#171322]">
                Penugasan belum bisa dimuat
              </h3>
              <p class="mt-2 text-sm leading-6 text-[#6b7280]">
                {{ subjectClassesError }}
              </p>
              <button
                type="button"
                class="mt-4 rounded-lg bg-[#171322] px-4 py-2 text-sm font-medium text-white transition hover:bg-[#374151]"
                @click="loadClassContext"
              >
                Coba lagi
              </button>
            </div>

            <div
              v-else-if="!selectedClassId"
              class="rounded-lg bg-[#fbfaf8] px-5 py-10 text-center"
            >
              <PhCalendarBlank
                :size="28"
                class="mx-auto text-[#9ca3af]"
                weight="duotone"
              />
              <h3 class="mt-3 text-sm font-semibold text-[#171322]">
                Belum ada kelas dipilih
              </h3>
              <p class="mt-2 text-sm leading-6 text-[#6b7280]">
                Pilih tahun ajaran, semester, dan kelas untuk mengatur guru
                pengampu.
              </p>
            </div>

            <div
              v-else-if="subjectClasses.length === 0"
              class="rounded-lg bg-[#fbfaf8] px-5 py-10 text-center"
            >
              <PhBookOpen
                :size="28"
                class="mx-auto text-[#9ca3af]"
                weight="duotone"
              />
              <h3 class="mt-3 text-sm font-semibold text-[#171322]">
                Belum ada penugasan mengajar
              </h3>
              <p class="mt-2 text-sm leading-6 text-[#6b7280]">
                Pilih mata pelajaran dan guru melalui panel penugasan.
              </p>
            </div>

            <div v-else class="divide-y divide-[#ebe7df]">
              <article
                v-for="subjectClass in subjectClasses"
                :key="subjectClass.subjectClassId"
                class="min-w-0 py-4 first:pt-0 last:pb-0"
              >
                <div
                  class="flex min-w-0 flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
                >
                  <div class="flex min-w-0 items-center gap-3">
                    <div
                      class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-[#eef2ff] text-[#4f46e5]"
                    >
                      <PhBookOpen :size="20" weight="duotone" />
                    </div>
                    <div class="min-w-0">
                      <div class="flex min-w-0 flex-wrap items-center gap-2">
                        <h3
                          class="wrap-break-word text-sm font-semibold text-[#171322]"
                        >
                          {{
                            subjectClass.subjectName ||
                            "Mata pelajaran tidak tersedia"
                          }}
                        </h3>
                        <span
                          class="rounded-lg bg-[#f3f4f6] px-2 py-1 text-[11px] font-medium text-[#6b7280]"
                        >
                          {{
                            subjectClass.subjectCode || "Kode tidak tersedia"
                          }}
                        </span>
                      </div>
                      <p class="mt-1 text-xs text-[#6b7280]">
                        Guru:
                        <span class="font-medium text-[#374151]">
                          {{
                            subjectClass.teacherName || "Guru tidak tersedia"
                          }}
                        </span>
                      </p>
                    </div>
                  </div>
                  <button
                    type="button"
                    class="inline-flex shrink-0 items-center justify-center gap-2 rounded-lg border border-[#fecaca] bg-white px-3 py-2 text-xs font-medium text-[#dc2626] transition hover:bg-[#fef2f2] disabled:opacity-60"
                    :disabled="Boolean(unassigningId)"
                    @click="requestUnassign(subjectClass)"
                  >
                    <PhTrash :size="14" weight="duotone" />
                    Lepaskan
                  </button>
                </div>

                <div
                  v-if="
                    pendingUnassign?.subjectClassId ===
                    subjectClass.subjectClassId
                  "
                  class="mt-3 rounded-lg border border-[#fecaca] bg-[#fef2f2] p-3"
                >
                  <p class="text-xs leading-5 text-[#991b1b]">
                    Penugasan ini akan dilepas. Guru tidak lagi melihat ruang
                    mata pelajaran ini. Materi, tugas, pengumpulan, dan nilai
                    tidak akan dihapus.
                  </p>
                  <div class="mt-3 flex flex-wrap gap-2">
                    <button
                      type="button"
                      class="rounded-lg bg-[#dc2626] px-3 py-2 text-xs font-medium text-white transition hover:bg-[#b91c1c] disabled:opacity-60"
                      :disabled="unassigningId === subjectClass.subjectClassId"
                      @click="confirmUnassign(subjectClass)"
                    >
                      {{
                        unassigningId === subjectClass.subjectClassId
                          ? "Melepaskan..."
                          : "Ya, lepaskan"
                      }}
                    </button>
                    <button
                      type="button"
                      class="rounded-lg border border-[#fecaca] bg-white px-3 py-2 text-xs font-medium text-[#991b1b] transition hover:bg-[#fee2e2]"
                      :disabled="unassigningId === subjectClass.subjectClassId"
                      @click="cancelUnassign"
                    >
                      Batal
                    </button>
                  </div>
                </div>
              </article>
            </div>
          </div>
        </section>

        <aside class="order-1 min-w-0 lg:order-2">
          <div class="space-y-5 lg:sticky lg:top-6">
            <section class="rounded-xl border border-[#ebe7df] bg-white p-5">
              <div class="flex items-start justify-between gap-3">
                <div>
                  <p
                    class="text-[10px] font-medium uppercase tracking-[0.08em] text-[#9ca3af]"
                  >
                    Konteks kelas
                  </p>
                  <h2 class="mt-1 text-base font-semibold text-[#171322]">
                    Pilih periode dan kelas
                  </h2>
                </div>
                <PhCalendarBlank
                  :size="21"
                  class="text-[#ea580c]"
                  weight="duotone"
                />
              </div>

              <div class="mt-5 space-y-3">
                <label class="block text-xs font-medium text-[#6b7280]">
                  Tahun ajaran
                  <select
                    v-model="selectedAcademicYearId"
                    class="mt-2 w-full rounded-lg border border-[#ebe7df] bg-[#fbfaf8] px-3.5 py-2.5 text-sm text-[#171322] outline-none transition focus:border-[#4f46e5] focus:bg-white"
                    :disabled="yearsLoading || academicYears.length === 0"
                    @change="handleAcademicYearChange"
                  >
                    <option value="" disabled>Pilih tahun ajaran</option>
                    <option
                      v-for="year in academicYears"
                      :key="year.academicYearId"
                      :value="year.academicYearId"
                    >
                      {{ year.academicYearName
                      }}{{ year.isActive ? " - Aktif" : "" }}
                    </option>
                  </select>
                </label>
                <label class="block text-xs font-medium text-[#6b7280]">
                  Semester
                  <select
                    v-model="selectedTermId"
                    class="mt-2 w-full rounded-lg border border-[#ebe7df] bg-[#fbfaf8] px-3.5 py-2.5 text-sm text-[#171322] outline-none transition focus:border-[#4f46e5] focus:bg-white"
                    :disabled="termsLoading || terms.length === 0"
                    @change="handleTermChange"
                  >
                    <option value="" disabled>Pilih semester</option>
                    <option
                      v-for="term in terms"
                      :key="term.termId"
                      :value="term.termId"
                    >
                      {{ term.termName }}{{ term.isActive ? " - Aktif" : "" }}
                    </option>
                  </select>
                </label>
                <label class="block text-xs font-medium text-[#6b7280]">
                  Kelas
                  <select
                    v-model="selectedClassId"
                    class="mt-2 w-full rounded-lg border border-[#ebe7df] bg-[#fbfaf8] px-3.5 py-2.5 text-sm text-[#171322] outline-none transition focus:border-[#4f46e5] focus:bg-white"
                    :disabled="classesLoading || classes.length === 0"
                    @change="handleClassChange"
                  >
                    <option value="" disabled>Pilih kelas</option>
                    <option
                      v-for="classItem in classes"
                      :key="classItem.classId"
                      :value="classItem.classId"
                    >
                      {{ classItem.classTitle }} - {{ classItem.classCode }}
                    </option>
                  </select>
                </label>
              </div>

              <div class="mt-4 space-y-2 text-xs leading-5">
                <p
                  v-if="yearsLoading || termsLoading || classesLoading"
                  class="text-[#6b7280]"
                >
                  Memuat konteks kelas...
                </p>
                <div
                  v-else-if="yearsError || termsError || classesError"
                  class="rounded-lg bg-[#fef2f2] px-3 py-2 text-[#dc2626]"
                >
                  <p>{{ yearsError || termsError || classesError }}</p>
                  <button
                    type="button"
                    class="mt-2 font-medium underline underline-offset-2"
                    @click="
                      yearsError
                        ? loadAcademicYears()
                        : termsError
                          ? loadTerms(true)
                          : loadClasses(true)
                    "
                  >
                    Coba lagi
                  </button>
                </div>
                <p
                  v-else-if="academicYears.length === 0"
                  class="rounded-lg bg-[#fbfaf8] px-3 py-2 text-[#6b7280]"
                >
                  Belum ada tahun ajaran.
                </p>
                <p
                  v-else-if="selectedAcademicYearId && terms.length === 0"
                  class="rounded-lg bg-[#fbfaf8] px-3 py-2 text-[#6b7280]"
                >
                  Belum ada semester untuk tahun ajaran ini.
                </p>
                <p
                  v-else-if="selectedTermId && classes.length === 0"
                  class="rounded-lg bg-[#fbfaf8] px-3 py-2 text-[#6b7280]"
                >
                  Belum ada kelas untuk semester ini.
                </p>
              </div>
            </section>

            <section class="rounded-xl border border-[#ebe7df] bg-white p-5">
              <div class="flex items-start justify-between gap-3">
                <div>
                  <p
                    class="text-[10px] font-medium uppercase tracking-[0.08em] text-[#9ca3af]"
                  >
                    Tambah penugasan
                  </p>
                  <h2 class="mt-1 text-base font-semibold text-[#171322]">
                    Hubungkan guru dan mata pelajaran
                  </h2>
                  <p class="mt-1 text-xs leading-5 text-[#6b7280]">
                    Guru harus sudah ditempatkan aktif pada kelas terpilih.
                  </p>
                </div>
                <span
                  class="shrink-0 rounded-lg bg-[#eef2ff] px-2.5 py-1.5 text-xs font-medium text-[#4f46e5]"
                >
                  {{ teacherCandidates.length }} guru
                </span>
              </div>

              <form class="mt-5 space-y-3" @submit.prevent="submitSubjectClass">
                <label class="block text-xs font-medium text-[#6b7280]">
                  Mata pelajaran
                  <select
                    v-model="selectedSubjectId"
                    class="mt-2 w-full rounded-lg border border-[#ebe7df] bg-[#fbfaf8] px-3.5 py-2.5 text-sm text-[#171322] outline-none transition focus:border-[#4f46e5] focus:bg-white"
                    :disabled="
                      subjectsLoading || availableSubjects.length === 0
                    "
                  >
                    <option value="" disabled>Pilih mata pelajaran</option>
                    <option
                      v-for="subject in availableSubjects"
                      :key="subject.subjectId"
                      :value="subject.subjectId"
                    >
                      {{ subject.subjectName }} - {{ subject.subjectCode }}
                    </option>
                  </select>
                </label>

                <label class="block text-xs font-medium text-[#6b7280]">
                  Guru pengampu
                  <select
                    v-model="selectedTeacherSchoolUserId"
                    class="mt-2 w-full rounded-lg border border-[#ebe7df] bg-[#fbfaf8] px-3.5 py-2.5 text-sm text-[#171322] outline-none transition focus:border-[#4f46e5] focus:bg-white"
                    :disabled="
                      enrollmentsLoading ||
                      membersLoading ||
                      teacherCandidates.length === 0
                    "
                  >
                    <option value="" disabled>Pilih guru</option>
                    <option
                      v-for="teacher in teacherCandidates"
                      :key="teacher.schoolUserId"
                      :value="teacher.schoolUserId"
                    >
                      {{ teacher.fullName || "Nama guru tidak tersedia" }} -
                      {{ teacher.email || "Email tidak tersedia" }}
                    </option>
                  </select>
                </label>

                <div
                  v-if="selectedSubject || selectedTeacher"
                  class="rounded-lg bg-[#fbfaf8] p-3"
                >
                  <p
                    class="text-[10px] font-medium uppercase tracking-[0.08em] text-[#9ca3af]"
                  >
                    Ringkasan
                  </p>
                  <dl class="mt-2 space-y-2 text-xs">
                    <div class="flex items-start justify-between gap-3">
                      <dt class="text-[#6b7280]">Mata pelajaran</dt>
                      <dd
                        class="max-w-[65%] text-right font-medium text-[#171322]"
                      >
                        {{ selectedSubject?.subjectName || "Belum dipilih" }}
                      </dd>
                    </div>
                    <div class="flex items-start justify-between gap-3">
                      <dt class="text-[#6b7280]">Guru</dt>
                      <dd
                        class="max-w-[65%] text-right font-medium text-[#171322]"
                      >
                        {{ selectedTeacher?.fullName || "Belum dipilih" }}
                      </dd>
                    </div>
                  </dl>
                </div>

                <div class="space-y-2 text-xs leading-5">
                  <p v-if="subjectsLoading" class="text-[#6b7280]">
                    Memuat mata pelajaran...
                  </p>
                  <div
                    v-else-if="subjectsError"
                    class="rounded-lg bg-[#fef2f2] px-3 py-2 text-[#dc2626]"
                  >
                    <p>{{ subjectsError }}</p>
                    <button
                      type="button"
                      class="mt-2 font-medium underline underline-offset-2"
                      @click="loadSubjects"
                    >
                      Coba lagi
                    </button>
                  </div>
                  <p
                    v-else-if="subjects.length === 0"
                    class="rounded-lg bg-[#fbfaf8] px-3 py-2 text-[#6b7280]"
                  >
                    Belum ada mata pelajaran pada Struktur Akademik.
                  </p>
                  <p
                    v-else-if="
                      selectedClassId && availableSubjects.length === 0
                    "
                    class="rounded-lg bg-[#fbfaf8] px-3 py-2 text-[#6b7280]"
                  >
                    Semua mata pelajaran sudah ditugaskan untuk kelas ini.
                  </p>
                  <p
                    v-if="membersLoading || enrollmentsLoading"
                    class="text-[#6b7280]"
                  >
                    Memuat guru yang tersedia...
                  </p>
                  <div
                    v-else-if="membersError || enrollmentsError"
                    class="rounded-lg bg-[#fef2f2] px-3 py-2 text-[#dc2626]"
                  >
                    <p>{{ membersError || enrollmentsError }}</p>
                    <button
                      type="button"
                      class="mt-2 font-medium underline underline-offset-2"
                      @click="membersError ? loadMembers() : loadClassContext()"
                    >
                      Coba lagi
                    </button>
                  </div>
                  <p
                    v-else-if="
                      selectedClassId && teacherCandidates.length === 0
                    "
                    class="rounded-lg bg-[#fff7ed] px-3 py-2 text-[#92400e]"
                  >
                    Belum ada guru aktif yang dapat ditugaskan pada kelas ini.
                  </p>
                </div>

                <button
                  type="submit"
                  class="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-[#ea580c] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#c2410c] disabled:cursor-not-allowed disabled:opacity-60"
                  :disabled="
                    submitting ||
                    !selectedClassId ||
                    !selectedSubjectId ||
                    !selectedTeacherSchoolUserId ||
                    teacherCandidates.length === 0
                  "
                >
                  <PhLinkSimple :size="17" weight="duotone" />
                  {{ submitting ? "Menyimpan..." : "Buat penugasan" }}
                </button>
              </form>
            </section>
          </div>
        </aside>
      </div>
    </section>
  </main>
</template>
