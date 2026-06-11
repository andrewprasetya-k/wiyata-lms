<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import {
  PhBookOpen,
  PhCalendarBlank,
  PhChalkboardTeacher,
  PhLinkSimple,
  PhTrash,
  PhUsers,
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
import type { AcademicYearItem, SubjectItem, TermItem } from "../../types/adminAcademic";
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
    academicYears.value.find((year) => year.academicYearId === selectedAcademicYearId.value) ??
    null,
);

const selectedTerm = computed(
  () => terms.value.find((term) => term.termId === selectedTermId.value) ?? null,
);

const selectedClass = computed(
  () =>
    classes.value.find((classItem) => classItem.classId === selectedClassId.value) ?? null,
);

const assignedSubjectIds = computed(
  () => new Set(subjectClasses.value.map((subjectClass) => subjectClass.subjectId)),
);

const availableSubjects = computed(() =>
  subjects.value.filter((subject) => !assignedSubjectIds.value.has(subject.subjectId)),
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
      (member.roles ?? []).some((role) => normalizeRoleName(role) === "teacher"),
  ),
);

const selectedSubject = computed(
  () => subjects.value.find((subject) => subject.subjectId === selectedSubjectId.value) ?? null,
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
  if (selectedSubjectId.value && assignedSubjectIds.value.has(selectedSubjectId.value)) {
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
      academicYears.value.find((year) => year.isActive) ?? academicYears.value[0] ?? null;
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

    const selectedStillValid = terms.value.some((term) => term.termId === selectedTermId.value);
    if (selectDefault || !selectedStillValid) {
      const defaultTerm = terms.value.find((term) => term.isActive) ?? terms.value[0] ?? null;
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
    membersError.value = "Member sekolah belum bisa dimuat.";
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
    enrollmentsError.value = "Enrollment kelas belum bisa dimuat.";
    subjectClassesError.value = "Subject class belum bisa dimuat.";
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
    toast.error("Pilih subject yang belum ditugaskan.");
    return;
  }
  if (teacherCandidates.value.length === 0) {
    toast.error(
      "Belum ada guru yang eligible. Atur role teacher di Warga Sekolah dan penempatan teacher di Penempatan Kelas.",
    );
    return;
  }
  if (!selectedTeacherSchoolUserId.value) {
    toast.error("Pilih teacher yang akan mengampu subject.");
    return;
  }
  if (assignedSubjectIds.value.has(selectedSubjectId.value)) {
    toast.info("Subject ini sudah punya assignment di kelas terpilih.");
    return;
  }

  submitting.value = true;
  try {
    await assignSubjectClass({
      classId: selectedClassId.value,
      subjectId: selectedSubjectId.value,
      teacherId: selectedTeacherSchoolUserId.value,
    });
    toast.success("Subject class berhasil dibuat untuk teacher workspace.");
    await loadClassContext();
  } catch {
    toast.error(
      "Subject class belum bisa dibuat. Periksa eligibility teacher dan duplikasi subject.",
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
  <main class="min-h-screen flex-1 px-5 py-5 sm:px-6 lg:px-8">
    <section class="flex w-full max-w-none flex-col gap-5">
      <header class="soft-card rounded-[22px] p-5">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <p class="text-[11px] font-medium uppercase text-[#9CA3AF]">Admin sekolah</p>
            <h1 class="mt-2 text-2xl font-medium text-[#111827]">Penugasan Mengajar</h1>
            <p class="mt-2 max-w-3xl text-sm leading-6 text-[#6B7280]">
              Hubungkan guru, kelas, dan mata pelajaran untuk membuat teacher workspace. Akses student mengikuti penempatan kelas.
            </p>
          </div>
          <div class="flex flex-wrap gap-2 text-xs">
            <span class="rounded-lg bg-[#EEF2FF] px-3 py-1.5 font-medium text-[#4F46E5]">
              {{ currentSchool.schoolName || "Sekolah belum tersedia" }}
            </span>
            <span class="rounded-lg bg-[#F9FAFB] px-3 py-1.5 font-medium text-[#6B7280]">
              {{ currentSchool.schoolCode || "Kode sekolah belum tersedia" }}
            </span>
          </div>
        </div>

        <div
          v-if="!currentSchool.hasContext"
          class="mt-4 rounded-[10px] border border-[#FECACA] bg-[#FEF2F2] px-4 py-3 text-sm text-[#DC2626]"
        >
          Context sekolah aktif belum tersedia. Pastikan akun admin memiliki membership sekolah.
        </div>

      </header>

      <section class="grid gap-5 xl:grid-cols-[minmax(0,0.82fr)_minmax(0,1.18fr)]">
        <article class="rounded-[18px] border border-[#EBEBEB] bg-white p-5">
          <div class="flex items-start justify-between gap-4">
            <div>
              <p class="text-[11px] font-medium uppercase text-[#9CA3AF]">Context kelas</p>
              <h2 class="mt-2 text-base font-medium text-[#111827]">Pilih periode dan kelas</h2>
            </div>
            <PhCalendarBlank :size="22" class="text-[#4F46E5]" weight="duotone" />
          </div>

          <div class="mt-5 space-y-4">
            <label class="block text-sm font-medium text-[#374151]">
              Tahun ajaran
              <select
                v-model="selectedAcademicYearId"
                class="mt-2 w-full rounded-2xl border border-[#EBEBEB] bg-white px-4 py-3 text-sm text-[#111827] outline-none transition focus:border-[#4F46E5]"
                :disabled="yearsLoading || academicYears.length === 0"
                @change="handleAcademicYearChange"
              >
                <option value="" disabled>Pilih tahun ajaran</option>
                <option
                  v-for="year in academicYears"
                  :key="year.academicYearId"
                  :value="year.academicYearId"
                >
                  {{ year.academicYearName }}{{ year.isActive ? " - Aktif" : "" }}
                </option>
              </select>
            </label>

            <label class="block text-sm font-medium text-[#374151]">
              Semester
              <select
                v-model="selectedTermId"
                class="mt-2 w-full rounded-2xl border border-[#EBEBEB] bg-white px-4 py-3 text-sm text-[#111827] outline-none transition focus:border-[#4F46E5]"
                :disabled="termsLoading || terms.length === 0"
                @change="handleTermChange"
              >
                <option value="" disabled>Pilih semester</option>
                <option v-for="term in terms" :key="term.termId" :value="term.termId">
                  {{ term.termName }}{{ term.isActive ? " - Aktif" : "" }}
                </option>
              </select>
            </label>

            <label class="block text-sm font-medium text-[#374151]">
              Kelas
              <select
                v-model="selectedClassId"
                class="mt-2 w-full rounded-2xl border border-[#EBEBEB] bg-white px-4 py-3 text-sm text-[#111827] outline-none transition focus:border-[#4F46E5]"
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

          <div class="mt-4 space-y-2 text-sm">
            <p v-if="yearsLoading" class="text-[#6B7280]">Memuat tahun ajaran...</p>
            <p v-else-if="yearsError" class="text-[#DC2626]">{{ yearsError }}</p>
            <p v-else-if="academicYears.length === 0" class="text-[#6B7280]">
              Belum ada tahun ajaran. Buat data akademik terlebih dahulu.
            </p>

            <p v-if="termsLoading" class="text-[#6B7280]">Memuat semester...</p>
            <p v-else-if="termsError" class="text-[#DC2626]">{{ termsError }}</p>
            <p v-else-if="selectedAcademicYearId && terms.length === 0" class="text-[#6B7280]">
              Belum ada semester untuk tahun ajaran ini.
            </p>

            <p v-if="classesLoading" class="text-[#6B7280]">Memuat kelas...</p>
            <p v-else-if="classesError" class="text-[#DC2626]">{{ classesError }}</p>
            <p v-else-if="selectedTermId && classes.length === 0" class="text-[#6B7280]">
              Belum ada kelas untuk semester ini.
            </p>
          </div>

          <div class="mt-5 rounded-[18px] bg-[#FBFAF8] p-4">
            <p class="text-[11px] font-medium uppercase text-[#9CA3AF]">Context aktif</p>
            <div class="mt-3 space-y-2 text-sm text-[#374151]">
              <p>
                Tahun ajaran:
                <span class="font-medium text-[#111827]">
                  {{ selectedAcademicYear?.academicYearName || "Belum dipilih" }}
                </span>
              </p>
              <p>
                Semester:
                <span class="font-medium text-[#111827]">
                  {{ selectedTerm?.termName || "Belum dipilih" }}
                </span>
              </p>
              <p>
                Kelas:
                <span class="font-medium text-[#111827]">
                  {{ selectedClass?.classTitle || "Belum dipilih" }}
                </span>
              </p>
            </div>
          </div>
        </article>

        <article class="rounded-[18px] border border-[#EBEBEB] bg-white p-5">
          <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div>
              <p class="text-[11px] font-medium uppercase text-[#9CA3AF]">Assign teacher</p>
              <h2 class="mt-2 text-base font-medium text-[#111827]">Buat penugasan mengajar</h2>
              <p class="mt-1 text-sm leading-6 text-[#6B7280]">
                Guru harus punya role teacher di Warga Sekolah dan sudah ditempatkan sebagai teacher kelas di Penempatan Kelas.
              </p>
            </div>
            <div class="inline-flex items-center gap-2 rounded-lg bg-[#EEF2FF] px-3 py-2 text-xs font-medium text-[#4F46E5]">
              <PhChalkboardTeacher :size="16" weight="duotone" />
              {{ teacherCandidates.length }} teacher eligible
            </div>
          </div>

          <form class="mt-5 grid gap-4" @submit.prevent="submitSubjectClass">
            <label class="block text-sm font-medium text-[#374151]">
              Subject
              <select
                v-model="selectedSubjectId"
                class="mt-2 w-full rounded-2xl border border-[#EBEBEB] bg-white px-4 py-3 text-sm text-[#111827] outline-none transition focus:border-[#4F46E5]"
                :disabled="subjectsLoading || availableSubjects.length === 0"
              >
                <option value="" disabled>Pilih subject</option>
                <option
                  v-for="subject in availableSubjects"
                  :key="subject.subjectId"
                  :value="subject.subjectId"
                >
                  {{ subject.subjectName }} - {{ subject.subjectCode }}
                </option>
              </select>
            </label>

            <label class="block text-sm font-medium text-[#374151]">
              Teacher
              <select
                v-model="selectedTeacherSchoolUserId"
                class="mt-2 w-full rounded-2xl border border-[#EBEBEB] bg-white px-4 py-3 text-sm text-[#111827] outline-none transition focus:border-[#4F46E5]"
                :disabled="enrollmentsLoading || membersLoading || teacherCandidates.length === 0"
              >
                <option value="" disabled>Pilih teacher</option>
                <option
                  v-for="teacher in teacherCandidates"
                  :key="teacher.schoolUserId"
                  :value="teacher.schoolUserId"
                >
                  {{ teacher.fullName || "Nama teacher tidak tersedia" }} - {{ teacher.email || "Email tidak tersedia" }}
                </option>
              </select>
            </label>

            <div class="rounded-[18px] bg-[#FBFAF8] p-4 text-sm leading-6 text-[#6B7280]">
              <p>
                Student tidak dipilih satu per satu di halaman ini. Semua student yang enrolled di kelas otomatis mendapat akses ke subject_class kelas tersebut.
              </p>
              <p class="mt-2">
                Payload memakai <span class="font-medium text-[#374151]">teacherId = schoolUserId</span>, bukan userId global.
              </p>
            </div>

            <div
              v-if="selectedSubject || selectedTeacher"
              class="rounded-[18px] border border-[#EBEBEB] bg-white p-4"
            >
              <p class="text-[11px] font-medium uppercase text-[#9CA3AF]">Preview</p>
              <div class="mt-3 grid gap-2 text-sm text-[#374151] sm:grid-cols-2">
                <p>
                  Subject:
                  <span class="font-medium text-[#111827]">
                    {{ selectedSubject?.subjectName || "Belum dipilih" }}
                  </span>
                </p>
                <p>
                  Teacher:
                  <span class="font-medium text-[#111827]">
                    {{ selectedTeacher?.fullName || "Belum dipilih" }}
                  </span>
                </p>
              </div>
            </div>

            <div class="space-y-2 text-sm">
              <p v-if="subjectsLoading" class="text-[#6B7280]">Memuat subject...</p>
              <p v-else-if="subjectsError" class="text-[#DC2626]">{{ subjectsError }}</p>
              <p v-else-if="subjects.length === 0" class="text-[#6B7280]">
                Belum ada subject. Buat subject di Struktur Akademik terlebih dahulu.
              </p>
              <p v-else-if="selectedClassId && availableSubjects.length === 0" class="text-[#6B7280]">
                Semua subject sudah ditugaskan untuk kelas ini.
              </p>

              <p v-if="membersLoading || enrollmentsLoading" class="text-[#6B7280]">
                Memuat teacher eligible...
              </p>
              <p v-else-if="membersError" class="text-[#DC2626]">{{ membersError }}</p>
              <p v-else-if="enrollmentsError" class="text-[#DC2626]">{{ enrollmentsError }}</p>
              <p v-else-if="selectedClassId && teacherCandidates.length === 0" class="text-[#6B7280]">
                Belum ada teacher eligible. Pastikan teacher punya school role teacher dan class_role teacher di kelas ini.
              </p>
            </div>

            <button
              type="submit"
              class="inline-flex items-center justify-center gap-2 rounded-2xl bg-[#111827] px-4 py-3 text-sm font-medium text-white transition hover:bg-[#374151] disabled:cursor-not-allowed disabled:opacity-60"
              :disabled="
                submitting ||
                !selectedClassId ||
                !selectedSubjectId ||
                !selectedTeacherSchoolUserId ||
                teacherCandidates.length === 0
              "
            >
              <PhLinkSimple :size="18" weight="duotone" />
              Buat subject class
            </button>
          </form>
        </article>
      </section>

      <section class="rounded-[18px] border border-[#EBEBEB] bg-white p-5">
        <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <p class="text-[11px] font-medium uppercase text-[#9CA3AF]">Current assignments</p>
            <h2 class="mt-2 text-base font-medium text-[#111827]">Subject class kelas ini</h2>
            <p class="mt-1 text-sm text-[#6B7280]">
              Setiap subject_class menjadi workspace teacher untuk materi, tugas, dan penilaian.
            </p>
          </div>
          <div class="flex flex-wrap gap-2 text-xs font-medium">
            <span class="inline-flex items-center gap-2 rounded-lg bg-[#EEF2FF] px-3 py-2 text-[#4F46E5]">
              <PhBookOpen :size="16" weight="duotone" />
              {{ subjectClasses.length }} subject class
            </span>
            <span class="inline-flex items-center gap-2 rounded-lg bg-[#ECFDF5] px-3 py-2 text-[#059669]">
              <PhUsers :size="16" weight="duotone" />
              Student via class enrollment
            </span>
          </div>
        </div>

        <div class="mt-5">
          <div
            v-if="subjectClassesLoading"
            class="rounded-[18px] bg-[#FBFAF8] p-5 text-sm text-[#6B7280]"
          >
            Memuat subject class...
          </div>

          <div
            v-else-if="subjectClassesError"
            class="flex items-start gap-3 rounded-[18px] border border-[#FECACA] bg-[#FEF2F2] p-5 text-sm text-[#DC2626]"
          >
            <PhWarningCircle :size="20" weight="duotone" />
            <p>{{ subjectClassesError }}</p>
          </div>

          <div
            v-else-if="!selectedClassId"
            class="rounded-[18px] bg-[#FBFAF8] p-5 text-sm text-[#6B7280]"
          >
            Pilih kelas untuk melihat subject class.
          </div>

          <div
            v-else-if="subjectClasses.length === 0"
            class="rounded-[18px] bg-[#FBFAF8] p-5 text-sm text-[#6B7280]"
          >
            Belum ada subject class untuk kelas ini.
          </div>

          <div v-else class="grid gap-3 md:grid-cols-2">
            <article
              v-for="subjectClass in subjectClasses"
              :key="subjectClass.subjectClassId"
              class="rounded-[18px] border border-[#EBEBEB] bg-[#FBFAF8] p-4"
            >
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0">
                  <div class="flex flex-wrap items-center gap-2">
                    <h3 class="text-sm font-medium text-[#111827]">
                      {{ subjectClass.subjectName || "Subject tidak tersedia" }}
                    </h3>
                    <span class="rounded-lg bg-white px-2 py-1 text-[11px] font-medium text-[#6B7280]">
                      {{ subjectClass.subjectCode || "Kode tidak tersedia" }}
                    </span>
                  </div>
                  <p class="mt-2 text-sm text-[#6B7280]">
                    Teacher:
                    <span class="font-medium text-[#374151]">
                      {{ subjectClass.teacherName || "Teacher tidak tersedia" }}
                    </span>
                  </p>
                </div>
                <div class="flex shrink-0 flex-col items-end gap-2">
                  <PhChalkboardTeacher
                    :size="22"
                    class="text-[#4F46E5]"
                    weight="duotone"
                  />
                  <button
                    type="button"
                    class="inline-flex items-center justify-center gap-2 rounded-xl border border-[#FECACA] bg-white px-3 py-2 text-xs font-medium text-[#DC2626] transition hover:bg-[#FEF2F2] disabled:cursor-not-allowed disabled:opacity-60"
                    :disabled="Boolean(unassigningId)"
                    @click="requestUnassign(subjectClass)"
                  >
                    <PhTrash :size="14" weight="duotone" />
                    Lepaskan
                  </button>
                </div>
              </div>

              <div
                v-if="pendingUnassign?.subjectClassId === subjectClass.subjectClassId"
                class="mt-3 rounded-2xl border border-[#FECACA] bg-[#FEF2F2] p-3"
              >
                <p class="text-xs leading-5 text-[#991B1B]">
                  Penugasan mengajar ini akan dilepas. Teacher tidak lagi melihat workspace subject ini. Materi, tugas, submission, dan nilai tidak akan dihapus.
                </p>
                <div class="mt-3 flex flex-wrap gap-2">
                  <button
                    type="button"
                    class="rounded-xl bg-[#DC2626] px-3 py-2 text-xs font-medium text-white transition hover:bg-[#B91C1C] disabled:cursor-not-allowed disabled:opacity-60"
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
                    class="rounded-xl border border-[#FECACA] bg-white px-3 py-2 text-xs font-medium text-[#991B1B] transition hover:bg-[#FEE2E2]"
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
    </section>
  </main>
</template>
