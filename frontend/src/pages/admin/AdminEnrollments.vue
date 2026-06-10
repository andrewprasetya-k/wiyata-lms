<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import {
  PhCalendarBlank,
  PhChalkboardTeacher,
  PhMagnifyingGlass,
  PhStudent,
  PhTrash,
  PhUsers,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import {
  getAcademicYearsBySchool,
  getTermsByAcademicYear,
} from "../../services/adminAcademic";
import { getAdminClasses } from "../../services/adminClass";
import { getSchoolMembers } from "../../services/adminUser";
import {
  createClassEnrollments,
  deleteEnrollment,
  getClassEnrollments,
} from "../../services/adminEnrollment";
import type { AcademicYearItem, TermItem } from "../../types/adminAcademic";
import type { AdminClassItem } from "../../types/adminClass";
import type { SchoolMemberItem } from "../../types/adminUser";
import type {
  ClassEnrollmentRole,
  EnrollmentMemberItem,
} from "../../types/adminEnrollment";
import { formatDateTime } from "../../utils/date";
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
const members = ref<SchoolMemberItem[]>([]);
const enrollments = ref<EnrollmentMemberItem[]>([]);

const selectedAcademicYearId = ref("");
const selectedTermId = ref("");
const selectedClassId = ref("");
const selectedSchoolUserIds = ref<string[]>([]);
const classRole = ref<ClassEnrollmentRole | "">("student");
const memberSearch = ref("");
const pendingUnenroll = ref<EnrollmentMemberItem | null>(null);

const yearsLoading = ref(false);
const termsLoading = ref(false);
const classesLoading = ref(false);
const membersLoading = ref(false);
const enrollmentsLoading = ref(false);
const submitting = ref(false);
const unenrollingId = ref("");

const yearsError = ref("");
const termsError = ref("");
const classesError = ref("");
const membersError = ref("");
const enrollmentsError = ref("");

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

const enrolledSchoolUserIds = computed(
  () => new Set(enrollments.value.map((enrollment) => enrollment.schoolUserId)),
);

const availableMembers = computed(() =>
  members.value.filter(
    (member) => !enrolledSchoolUserIds.value.has(member.schoolUserId),
  ),
);

const selectedMembers = computed(() =>
  availableMembers.value.filter((member) =>
    selectedSchoolUserIds.value.includes(member.schoolUserId),
  ),
);

const studentEnrollmentCount = computed(
  () =>
    enrollments.value.filter((enrollment) => enrollment.role === "student")
      .length,
);

const teacherEnrollmentCount = computed(
  () =>
    enrollments.value.filter((enrollment) => enrollment.role === "teacher")
      .length,
);

function classRoleLabel(role: string) {
  if (role === "teacher") return "Teacher kelas";
  if (role === "student") return "Student kelas";
  return role;
}

function schoolRolesLabel(member: SchoolMemberItem) {
  return member.roles?.length
    ? member.roles.join(", ")
    : "Role sekolah belum tersedia";
}

function unenrollConfirmationCopy(enrollment: EnrollmentMemberItem) {
  if (enrollment.role === "teacher") {
    return "Teacher akan dikeluarkan dari kelas. Jika teacher masih ditugaskan mengajar subject di kelas ini, lepaskan penugasan mengajar terlebih dahulu.";
  }

  return "Member ini akan dikeluarkan dari kelas. Akses ke materi, tugas, dan nilai kelas ini akan berhenti, tetapi histori submission/nilai tidak dihapus.";
}

function toggleMember(schoolUserId: string) {
  if (selectedSchoolUserIds.value.includes(schoolUserId)) {
    selectedSchoolUserIds.value = selectedSchoolUserIds.value.filter(
      (id) => id !== schoolUserId,
    );
    return;
  }

  selectedSchoolUserIds.value = [...selectedSchoolUserIds.value, schoolUserId];
}

function resetSelectedMembers() {
  const availableIds = new Set(
    availableMembers.value.map((member) => member.schoolUserId),
  );
  selectedSchoolUserIds.value = selectedSchoolUserIds.value.filter((id) =>
    availableIds.has(id),
  );
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

async function loadMembers() {
  members.value = [];
  membersError.value = "";

  if (!currentSchool.value.hasContext) return;

  membersLoading.value = true;
  try {
    const data = await getSchoolMembers(currentSchool.value.schoolCode, {
      page: 1,
      limit: 50,
      search: memberSearch.value.trim(),
    });
    members.value = data.members?.data ?? [];
    resetSelectedMembers();
  } catch {
    membersError.value = "Member sekolah belum bisa dimuat.";
  } finally {
    membersLoading.value = false;
  }
}

async function loadEnrollments() {
  enrollments.value = [];
  enrollmentsError.value = "";
  selectedSchoolUserIds.value = [];
  pendingUnenroll.value = null;

  if (!selectedClassId.value) return;

  enrollmentsLoading.value = true;
  try {
    const data = await getClassEnrollments(selectedClassId.value, {
      page: 1,
      limit: 50,
    });
    enrollments.value = data.members?.data ?? [];
    resetSelectedMembers();
  } catch {
    enrollmentsError.value = "Enrollment kelas belum bisa dimuat.";
  } finally {
    enrollmentsLoading.value = false;
  }
}

async function handleAcademicYearChange() {
  await loadTerms(true);
  await loadClasses(true);
  await loadEnrollments();
}

async function handleTermChange() {
  await loadClasses(true);
  await loadEnrollments();
}

async function handleClassChange() {
  await loadEnrollments();
}

async function submitEnrollment() {
  if (!currentSchool.value.schoolId || !currentSchool.value.schoolCode) {
    toast.error("Context sekolah aktif belum tersedia.");
    return;
  }
  if (!selectedClassId.value) {
    toast.error("Pilih kelas terlebih dahulu.");
    return;
  }
  if (selectedSchoolUserIds.value.length === 0) {
    toast.error("Pilih minimal satu member sekolah.");
    return;
  }
  if (!classRole.value) {
    toast.error("Pilih peran kelas.");
    return;
  }

  submitting.value = true;
  try {
    await createClassEnrollments({
      schoolId: currentSchool.value.schoolId,
      schoolUserIds: selectedSchoolUserIds.value,
      classId: selectedClassId.value,
      role: classRole.value,
    });
    toast.success(
      "Enrollment berhasil diproses. Member yang sudah terdaftar akan dilewati.",
    );
    selectedSchoolUserIds.value = [];
    await loadEnrollments();
  } catch {
    toast.error("Member belum bisa ditambahkan ke kelas.");
  } finally {
    submitting.value = false;
  }
}

function requestUnenroll(enrollment: EnrollmentMemberItem) {
  pendingUnenroll.value = enrollment;
}

function cancelUnenroll() {
  pendingUnenroll.value = null;
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

  return "Member belum bisa dikeluarkan dari kelas.";
}

async function confirmUnenroll(enrollment: EnrollmentMemberItem) {
  if (!enrollment.enrollmentId || unenrollingId.value) return;

  unenrollingId.value = enrollment.enrollmentId;
  try {
    await deleteEnrollment(enrollment.enrollmentId);
    toast.success("Member berhasil dikeluarkan dari kelas.");
    pendingUnenroll.value = null;
    await loadEnrollments();
  } catch (error) {
    toast.error(getErrorMessage(error));
  } finally {
    unenrollingId.value = "";
  }
}

onMounted(async () => {
  if (!currentSchool.value.hasContext) return;
  await Promise.all([loadAcademicYears(), loadMembers()]);
  await loadTerms(true);
  await loadClasses(true);
  await loadEnrollments();
});
</script>

<template>
  <main class="min-h-screen flex-1 px-5 py-5 sm:px-6 lg:px-8">
    <section class="flex w-full max-w-none flex-col gap-5">
      <header class="soft-card rounded-[22px] p-5">
        <div
          class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between"
        >
          <div>
            <p class="text-[11px] font-medium uppercase text-[#9CA3AF]">
              Admin sekolah
            </p>
            <h1 class="mt-2 text-2xl font-medium text-[#111827]">
              Penempatan Kelas
            </h1>
            <p class="mt-2 max-w-3xl text-sm leading-6 text-[#6B7280]">
              Tempatkan member sekolah ke kelas sebagai student atau teacher.
              Role sekolah tetap dikelola di halaman Warga Sekolah.
            </p>
          </div>
          <div class="flex flex-wrap gap-2 text-xs">
            <span
              class="rounded-lg bg-[#EEF2FF] px-3 py-1.5 font-medium text-[#4F46E5]"
            >
              {{ currentSchool.schoolName || "Sekolah belum tersedia" }}
            </span>
            <span
              class="rounded-lg bg-[#F9FAFB] px-3 py-1.5 font-medium text-[#6B7280]"
            >
              {{ currentSchool.schoolCode || "Kode sekolah belum tersedia" }}
            </span>
          </div>
        </div>

        <div
          v-if="!currentSchool.hasContext"
          class="mt-4 rounded-[10px] border border-[#FECACA] bg-[#FEF2F2] px-4 py-3 text-sm text-[#DC2626]"
        >
          Context sekolah aktif belum tersedia. Pastikan akun admin memiliki
          membership sekolah.
        </div>

      </header>

      <section
        class="grid gap-5 xl:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)]"
      >
        <article class="rounded-[18px] border border-[#EBEBEB] bg-white p-5">
          <div class="flex items-start justify-between gap-4">
            <div>
              <p class="text-[11px] font-medium uppercase text-[#9CA3AF]">
                Context kelas
              </p>
              <h2 class="mt-2 text-base font-medium text-[#111827]">
                Pilih periode dan kelas
              </h2>
            </div>
            <PhCalendarBlank
              :size="22"
              class="text-[#4F46E5]"
              weight="duotone"
            />
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
                  {{ year.academicYearName
                  }}{{ year.isActive ? " - Aktif" : "" }}
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
                <option
                  v-for="term in terms"
                  :key="term.termId"
                  :value="term.termId"
                >
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
            <p v-if="yearsLoading" class="text-[#6B7280]">
              Memuat tahun ajaran...
            </p>
            <p v-else-if="yearsError" class="text-[#DC2626]">
              {{ yearsError }}
            </p>
            <p v-else-if="academicYears.length === 0" class="text-[#6B7280]">
              Belum ada tahun ajaran. Buat data akademik terlebih dahulu.
            </p>

            <p v-if="termsLoading" class="text-[#6B7280]">Memuat semester...</p>
            <p v-else-if="termsError" class="text-[#DC2626]">
              {{ termsError }}
            </p>
            <p
              v-else-if="selectedAcademicYearId && terms.length === 0"
              class="text-[#6B7280]"
            >
              Belum ada semester untuk tahun ajaran ini.
            </p>

            <p v-if="classesLoading" class="text-[#6B7280]">Memuat kelas...</p>
            <p v-else-if="classesError" class="text-[#DC2626]">
              {{ classesError }}
            </p>
            <p
              v-else-if="selectedTermId && classes.length === 0"
              class="text-[#6B7280]"
            >
              Belum ada kelas untuk semester ini.
            </p>
          </div>

          <div class="mt-5 rounded-[18px] bg-[#FBFAF8] p-4">
            <p class="text-[11px] font-medium uppercase text-[#9CA3AF]">
              Context aktif
            </p>
            <div class="mt-3 space-y-2 text-sm text-[#374151]">
              <p>
                Tahun ajaran:
                <span class="font-medium text-[#111827]">
                  {{
                    selectedAcademicYear?.academicYearName || "Belum dipilih"
                  }}
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
          <div
            class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between"
          >
            <div>
              <p class="text-[11px] font-medium uppercase text-[#9CA3AF]">
                Tambah member kelas
              </p>
              <h2 class="mt-2 text-base font-medium text-[#111827]">
                Pilih member sekolah
              </h2>
              <p class="mt-1 text-sm leading-6 text-[#6B7280]">
                Peran di sini adalah class_role. Enroll teacher ke kelas belum
                berarti guru tersebut mengampu subject.
              </p>
            </div>
            <div
              class="inline-flex items-center gap-2 rounded-lg bg-[#EEF2FF] px-3 py-2 text-xs font-medium text-[#4F46E5]"
            >
              <PhUsers :size="16" weight="duotone" />
              {{ selectedMembers.length }} dipilih
            </div>
          </div>

          <div class="mt-5 grid gap-3 md:grid-cols-[1fr_auto]">
            <label class="block text-sm font-medium text-[#374151]">
              Cari member sekolah
              <div class="mt-2 flex gap-2">
                <input
                  v-model="memberSearch"
                  type="search"
                  placeholder="Nama atau email"
                  class="min-w-0 flex-1 rounded-2xl border border-[#EBEBEB] bg-white px-4 py-3 text-sm text-[#111827] outline-none transition placeholder:text-[#9CA3AF] focus:border-[#4F46E5]"
                />
                <button
                  type="button"
                  class="inline-flex items-center justify-center rounded-2xl border border-[#EBEBEB] bg-white px-4 py-3 text-sm font-medium text-[#374151] transition hover:bg-[#F9FAFB]"
                  :disabled="membersLoading"
                  @click="loadMembers"
                >
                  <PhMagnifyingGlass :size="18" weight="duotone" />
                </button>
              </div>
            </label>

            <label class="block text-sm font-medium text-[#374151]">
              Peran kelas
              <select
                v-model="classRole"
                class="mt-2 w-full rounded-2xl border border-[#EBEBEB] bg-white px-4 py-3 text-sm text-[#111827] outline-none transition focus:border-[#4F46E5] md:w-48"
              >
                <option value="" disabled>Pilih role</option>
                <option value="student">Student kelas</option>
                <option value="teacher">Teacher kelas</option>
              </select>
            </label>
          </div>

          <div class="mt-5">
            <div
              v-if="membersLoading"
              class="rounded-[18px] bg-[#FBFAF8] p-5 text-sm text-[#6B7280]"
            >
              Memuat member sekolah...
            </div>

            <div
              v-else-if="membersError"
              class="flex items-start gap-3 rounded-[18px] border border-[#FECACA] bg-[#FEF2F2] p-5 text-sm text-[#DC2626]"
            >
              <PhWarningCircle :size="20" weight="duotone" />
              <p>{{ membersError }}</p>
            </div>

            <div
              v-else-if="!selectedClassId"
              class="rounded-[18px] bg-[#FBFAF8] p-5 text-sm text-[#6B7280]"
            >
              Pilih kelas sebelum menambahkan member.
            </div>

            <div
              v-else-if="availableMembers.length === 0"
              class="rounded-[18px] bg-[#FBFAF8] p-5 text-sm text-[#6B7280]"
            >
              Tidak ada member sekolah yang bisa ditambahkan. Semua member yang
              tampil sudah terdaftar di kelas ini atau belum ada member sekolah.
            </div>

            <div v-else class="max-h-105 space-y-2 overflow-y-auto pr-1">
              <label
                v-for="member in availableMembers"
                :key="member.schoolUserId"
                class="flex cursor-pointer items-start gap-3 rounded-[18px] border border-[#EBEBEB] bg-[#FBFAF8] p-4 transition hover:border-[#D1D5DB]"
              >
                <input
                  type="checkbox"
                  class="mt-1 h-4 w-4 rounded border-[#D1D5DB] text-[#4F46E5] focus:ring-[#4F46E5]"
                  :checked="selectedSchoolUserIds.includes(member.schoolUserId)"
                  @change="toggleMember(member.schoolUserId)"
                />
                <span class="min-w-0 flex-1">
                  <span class="block text-sm font-medium text-[#111827]">
                    {{ member.fullName || "Nama member tidak tersedia" }}
                  </span>
                  <span class="mt-1 block text-xs text-[#6B7280]">
                    {{ member.email || "Email tidak tersedia" }}
                  </span>
                  <span
                    class="mt-2 inline-flex rounded-lg bg-white px-2 py-1 text-[11px] font-medium text-[#6B7280]"
                  >
                    Role sekolah: {{ schoolRolesLabel(member) }}
                  </span>
                </span>
              </label>
            </div>
          </div>

          <button
            type="button"
            class="mt-5 inline-flex w-full items-center justify-center gap-2 rounded-2xl bg-[#111827] px-4 py-3 text-sm font-medium text-white transition hover:bg-[#374151] disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="
              submitting ||
              !currentSchool.hasContext ||
              !selectedClassId ||
              selectedSchoolUserIds.length === 0 ||
              !classRole
            "
            @click="submitEnrollment"
          >
            <PhStudent :size="18" weight="duotone" />
            Tambahkan ke kelas
          </button>
        </article>
      </section>

      <section class="rounded-[18px] border border-[#EBEBEB] bg-white p-5">
        <div
          class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between"
        >
          <div>
            <p class="text-[11px] font-medium uppercase text-[#9CA3AF]">
              Class enrollment
            </p>
            <h2 class="mt-2 text-base font-medium text-[#111827]">
              Member kelas saat ini
            </h2>
            <p class="mt-1 text-sm text-[#6B7280]">
              Daftar ini hanya menampilkan membership kelas. Pengampu subject
              akan diatur pada fase subject_class assignment.
            </p>
          </div>
          <div class="flex flex-wrap gap-2 text-xs font-medium">
            <span
              class="inline-flex items-center gap-2 rounded-lg bg-[#ECFDF5] px-3 py-2 text-[#059669]"
            >
              <PhStudent :size="16" weight="duotone" />
              {{ studentEnrollmentCount }} student
            </span>
            <span
              class="inline-flex items-center gap-2 rounded-lg bg-[#EEF2FF] px-3 py-2 text-[#4F46E5]"
            >
              <PhChalkboardTeacher :size="16" weight="duotone" />
              {{ teacherEnrollmentCount }} teacher
            </span>
          </div>
        </div>

        <div class="mt-5">
          <div
            v-if="enrollmentsLoading"
            class="rounded-[18px] bg-[#FBFAF8] p-5 text-sm text-[#6B7280]"
          >
            Memuat enrollment kelas...
          </div>

          <div
            v-else-if="enrollmentsError"
            class="flex items-start gap-3 rounded-[18px] border border-[#FECACA] bg-[#FEF2F2] p-5 text-sm text-[#DC2626]"
          >
            <PhWarningCircle :size="20" weight="duotone" />
            <p>{{ enrollmentsError }}</p>
          </div>

          <div
            v-else-if="!selectedClassId"
            class="rounded-[18px] bg-[#FBFAF8] p-5 text-sm text-[#6B7280]"
          >
            Pilih kelas untuk melihat enrollment.
          </div>

          <div
            v-else-if="enrollments.length === 0"
            class="rounded-[18px] bg-[#FBFAF8] p-5 text-sm text-[#6B7280]"
          >
            Belum ada member yang terdaftar di kelas ini.
          </div>

          <div v-else class="grid gap-3 md:grid-cols-2">
            <article
              v-for="enrollment in enrollments"
              :key="enrollment.enrollmentId"
              class="rounded-[18px] border border-[#EBEBEB] bg-[#FBFAF8] p-4"
            >
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0">
                  <h3 class="truncate text-sm font-medium text-[#111827]">
                    {{
                      enrollment.userFullName || "Nama member tidak tersedia"
                    }}
                  </h3>
                  <p class="mt-1 truncate text-xs text-[#6B7280]">
                    {{ enrollment.userEmail || "Email tidak tersedia" }}
                  </p>
                </div>
                <span
                  class="shrink-0 rounded-lg px-2 py-1 text-[11px] font-medium"
                  :class="
                    enrollment.role === 'teacher'
                      ? 'bg-[#EEF2FF] text-[#4F46E5]'
                      : 'bg-[#ECFDF5] text-[#059669]'
                  "
                >
                  {{ classRoleLabel(enrollment.role) }}
                </span>
              </div>
              <div class="mt-3 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <p class="text-xs text-[#6B7280]">
                  Bergabung:
                  <span class="font-medium text-[#374151]">
                    {{ formatDateTime(enrollment.joinedAt) }}
                  </span>
                </p>
                <button
                  type="button"
                  class="inline-flex items-center justify-center gap-2 rounded-xl border border-[#FECACA] bg-white px-3 py-2 text-xs font-medium text-[#DC2626] transition hover:bg-[#FEF2F2] disabled:cursor-not-allowed disabled:opacity-60"
                  :disabled="Boolean(unenrollingId)"
                  @click="requestUnenroll(enrollment)"
                >
                  <PhTrash :size="14" weight="duotone" />
                  Keluarkan
                </button>
              </div>

              <div
                v-if="pendingUnenroll?.enrollmentId === enrollment.enrollmentId"
                class="mt-3 rounded-2xl border border-[#FECACA] bg-[#FEF2F2] p-3"
              >
                <p class="text-xs leading-5 text-[#991B1B]">
                  {{ unenrollConfirmationCopy(enrollment) }}
                </p>
                <div class="mt-3 flex flex-wrap gap-2">
                  <button
                    type="button"
                    class="rounded-xl bg-[#DC2626] px-3 py-2 text-xs font-medium text-white transition hover:bg-[#B91C1C] disabled:cursor-not-allowed disabled:opacity-60"
                    :disabled="unenrollingId === enrollment.enrollmentId"
                    @click="confirmUnenroll(enrollment)"
                  >
                    {{
                      unenrollingId === enrollment.enrollmentId
                        ? "Mengeluarkan..."
                        : "Ya, keluarkan"
                    }}
                  </button>
                  <button
                    type="button"
                    class="rounded-xl border border-[#FECACA] bg-white px-3 py-2 text-xs font-medium text-[#991B1B] transition hover:bg-[#FEE2E2]"
                    :disabled="unenrollingId === enrollment.enrollmentId"
                    @click="cancelUnenroll"
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
