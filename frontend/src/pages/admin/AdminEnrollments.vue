<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import {
  PhArrowRight,
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
import { getApiError } from "../../utils/error";

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

const eligibleAvailableMembers = computed(() =>
  availableMembers.value.filter((member) => inferPlacementRole(member)),
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
  if (role === "teacher") return "Guru";
  if (role === "student") return "Siswa";
  return role;
}

function normalizeRoleName(roleName: string) {
  return roleName.trim().toLowerCase();
}

function inferPlacementRole(
  member: SchoolMemberItem,
): ClassEnrollmentRole | null {
  const roles = (member.roles ?? []).map(normalizeRoleName);
  const hasStudentRole = roles.includes("student");
  const hasTeacherRole = roles.includes("teacher");

  if (hasStudentRole && !hasTeacherRole) return "student";
  if (hasTeacherRole && !hasStudentRole) return "teacher";
  return null;
}

function placementRoleLabel(member: SchoolMemberItem) {
  const role = inferPlacementRole(member);
  if (role) return classRoleLabel(role);
  return "Tidak dapat ditempatkan";
}

function unenrollConfirmationCopy(enrollment: EnrollmentMemberItem) {
  if (enrollment.role === "teacher") {
    return "Guru akan dikeluarkan dari kelas. Jika masih memiliki penugasan mengajar di kelas ini, lepaskan penugasan tersebut terlebih dahulu.";
  }

  return "Warga sekolah ini akan dikeluarkan dari kelas. Akses ke materi, tugas, dan nilai kelas akan berhenti, tetapi riwayat pengumpulan dan nilai tidak dihapus.";
}

function toggleMember(schoolUserId: string) {
  const member = availableMembers.value.find(
    (item) => item.schoolUserId === schoolUserId,
  );
  if (!member || !inferPlacementRole(member)) return;

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
    eligibleAvailableMembers.value.map((member) => member.schoolUserId),
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
    membersError.value = "Warga sekolah belum bisa dimuat.";
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
    enrollmentsError.value = "Penempatan kelas belum bisa dimuat.";
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
    toast.error("Konteks sekolah aktif belum tersedia.");
    return;
  }
  if (!selectedClassId.value) {
    toast.error("Pilih kelas terlebih dahulu.");
    return;
  }
  if (selectedSchoolUserIds.value.length === 0) {
    toast.error("Pilih minimal satu warga sekolah.");
    return;
  }
  const grouped = selectedMembers.value.reduce(
    (acc, member) => {
      const role = inferPlacementRole(member);
      if (role) acc[role].push(member.schoolUserId);
      return acc;
    },
    { student: [], teacher: [] } as Record<ClassEnrollmentRole, string[]>,
  );
  const requests = (["student", "teacher"] as ClassEnrollmentRole[])
    .filter((role) => grouped[role].length > 0)
    .map((role) => ({
      role,
      schoolUserIds: grouped[role],
    }));

  if (requests.length === 0) {
    toast.error("Pilih warga sekolah dengan peran Siswa atau Guru.");
    return;
  }

  submitting.value = true;
  try {
    const results = await Promise.allSettled(
      requests.map((request) =>
        createClassEnrollments({
          schoolId: currentSchool.value.schoolId,
          schoolUserIds: request.schoolUserIds,
          classId: selectedClassId.value,
          role: request.role,
        }),
      ),
    );

    const successCount = results.filter(
      (result) => result.status === "fulfilled",
    ).length;
    const failureCount = results.length - successCount;

    if (successCount > 0 && failureCount > 0) {
      toast.error(
        "Sebagian penempatan berhasil, tetapi ada kelompok peran yang belum bisa diproses.",
      );
    } else if (failureCount > 0) {
      toast.error("Warga sekolah belum bisa ditambahkan ke kelas.");
    } else {
      toast.success(
        "Penempatan berhasil diproses. Warga yang sudah terdaftar akan dilewati.",
      );
    }

    if (successCount > 0) {
      selectedSchoolUserIds.value = [];
    }
    await Promise.all([loadEnrollments(), loadMembers()]);
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


async function confirmUnenroll(enrollment: EnrollmentMemberItem) {
  if (!enrollment.enrollmentId || unenrollingId.value) return;

  unenrollingId.value = enrollment.enrollmentId;
  try {
    await deleteEnrollment(enrollment.enrollmentId);
    toast.success("Warga sekolah berhasil dikeluarkan dari kelas.");
    pendingUnenroll.value = null;
    await loadEnrollments();
  } catch (error) {
    toast.error(getApiError(error));
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
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-[#f8f7f4]">
    <header class="border-b border-[#ebe7df] bg-white">
      <div
        class="flex min-w-0 flex-col gap-3 px-5 py-5 sm:px-6 lg:flex-row lg:items-end lg:justify-between lg:px-8"
      >
        <div class="min-w-0">
          <h1 class="mt-1 text-2xl font-semibold text-[#171322] sm:text-3xl">
            Penempatan Kelas
          </h1>
          <p class="mt-2 max-w-3xl text-sm leading-6 text-[#6b7280]">
            Tempatkan siswa atau guru ke kelas aktif sebelum kegiatan belajar
            dan penugasan mengajar dimulai.
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

    <section class="px-5 py-5 sm:px-6 lg:px-8 lg:py-6">
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
          class="order-2 min-w-0 rounded-2xl border border-[#ebe7df] bg-white lg:order-1"
        >
          <div
            class="flex flex-col gap-3 border-b border-[#ebe7df] p-5 sm:flex-row sm:items-start sm:justify-between"
          >
            <div class="min-w-0">
              <p
                class="text-[10px] font-medium uppercase tracking-[0.08em] text-[#9ca3af]"
              >
                Penempatan aktif
              </p>
              <h2 class="mt-1 text-base font-semibold text-[#171322]">
                {{ selectedClass?.classTitle || "Pilih kelas" }}
              </h2>
              <p class="mt-1 text-sm leading-6 text-[#6b7280]">
                {{
                  selectedClass
                    ? `${selectedTerm?.termName || "Semester"} · ${selectedAcademicYear?.academicYearName || "Tahun ajaran"}`
                    : "Pilih konteks kelas pada panel untuk melihat warga yang ditempatkan."
                }}
              </p>
            </div>
            <div class="flex shrink-0 flex-wrap gap-2 text-xs font-medium">
              <span
                class="inline-flex items-center gap-2 rounded-lg bg-[#ecfdf5] px-3 py-2 text-[#059669]"
              >
                <PhStudent :size="16" weight="duotone" />
                {{ studentEnrollmentCount }} siswa
              </span>
              <span
                class="inline-flex items-center gap-2 rounded-lg bg-[#eef2ff] px-3 py-2 text-[#4f46e5]"
              >
                <PhChalkboardTeacher :size="16" weight="duotone" />
                {{ teacherEnrollmentCount }} guru
              </span>
            </div>
          </div>

          <div class="p-5">
            <div v-if="enrollmentsLoading" class="space-y-3">
              <div
                v-for="item in 3"
                :key="item"
                class="h-24 animate-pulse rounded-lg bg-[#fbfaf8]"
              />
            </div>

            <div
              v-else-if="enrollmentsError"
              class="rounded-lg border border-[#fecaca] bg-[#fef2f2] p-5 text-center"
            >
              <PhWarningCircle
                :size="26"
                class="mx-auto text-[#dc2626]"
                weight="duotone"
              />
              <h3 class="mt-3 text-sm font-semibold text-[#171322]">
                Penempatan belum bisa dimuat
              </h3>
              <p class="mt-2 text-sm leading-6 text-[#6b7280]">
                {{ enrollmentsError }}
              </p>
              <button
                type="button"
                class="mt-4 inline-flex items-center justify-center gap-2 rounded-lg bg-[#171322] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#374151] disabled:cursor-not-allowed disabled:opacity-60"
                @click="loadEnrollments"
              >
                Coba lagi
              </button>
            </div>

            <div
              v-else-if="!selectedClassId"
              class="rounded-lg bg-[#fbfaf8] px-5 py-8 text-center"
            >
              <PhCalendarBlank
                class="mx-auto h-7 w-7 text-[#9ca3af]"
                weight="duotone"
              />
              <h3 class="mt-3 text-sm font-semibold text-[#171322]">
                Belum ada kelas dipilih
              </h3>
              <p class="mt-2 text-sm leading-6 text-[#6b7280]">
                Pilih tahun ajaran, semester, dan kelas untuk mulai mengelola
                penempatan.
              </p>
            </div>

            <div
              v-else-if="enrollments.length === 0"
              class="rounded-lg bg-[#fbfaf8] px-5 py-8 text-center"
            >
              <PhUsers
                class="mx-auto h-7 w-7 text-[#9ca3af]"
                weight="duotone"
              />
              <h3 class="mt-3 text-sm font-semibold text-[#171322]">
                Belum ada warga di kelas ini
              </h3>
              <p class="mt-2 text-sm leading-6 text-[#6b7280]">
                Pilih warga sekolah melalui panel penempatan untuk
                menambahkannya.
              </p>
            </div>

            <div v-else class="divide-y divide-[#ebe7df]">
              <article
                v-for="enrollment in enrollments"
                :key="enrollment.enrollmentId"
                class="min-w-0 py-4 first:pt-0 last:pb-0"
              >
                <div
                  class="flex min-w-0 flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
                >
                  <div class="flex min-w-0 items-center gap-3">
                    <div
                      class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg text-xs font-semibold"
                      :class="
                        enrollment.role === 'teacher'
                          ? 'bg-[#eef2ff] text-[#4f46e5]'
                          : 'bg-[#ecfdf5] text-[#059669]'
                      "
                    >
                      {{
                        (enrollment.userFullName || "W").charAt(0).toUpperCase()
                      }}
                    </div>
                    <div class="min-w-0">
                      <div class="flex min-w-0 flex-wrap items-center gap-2">
                        <h3
                          class="wrap-break-word text-sm font-semibold text-[#171322]"
                        >
                          {{ enrollment.userFullName || "Nama tidak tersedia" }}
                        </h3>
                        <span
                          class="rounded-lg px-2 py-1 text-[11px] font-medium"
                          :class="
                            enrollment.role === 'teacher'
                              ? 'bg-[#eef2ff] text-[#4f46e5]'
                              : 'bg-[#ecfdf5] text-[#059669]'
                          "
                        >
                          {{ classRoleLabel(enrollment.role) }}
                        </span>
                      </div>
                      <p class="mt-1 break-all text-xs text-[#6b7280]">
                        {{ enrollment.userEmail || "Email tidak tersedia" }}
                      </p>
                      <p class="mt-1 text-[11px] text-[#9ca3af]">
                        Ditempatkan {{ formatDateTime(enrollment.joinedAt) }}
                      </p>
                    </div>
                  </div>
                  <button
                    type="button"
                    class="inline-flex shrink-0 items-center justify-center gap-2 rounded-lg border border-[#fecaca] bg-white px-3 py-2 text-sm font-medium text-[#dc2626] transition hover:bg-[#fef2f2] disabled:cursor-not-allowed disabled:opacity-60"
                    :disabled="Boolean(unenrollingId)"
                    @click="requestUnenroll(enrollment)"
                  >
                    <PhTrash :size="14" weight="duotone" />
                    Keluarkan
                  </button>
                </div>

                <div
                  v-if="
                    pendingUnenroll?.enrollmentId === enrollment.enrollmentId
                  "
                  class="mt-3 rounded-lg border border-[#fecaca] bg-[#fef2f2] p-3"
                >
                  <p class="text-xs leading-5 text-[#991b1b]">
                    {{ unenrollConfirmationCopy(enrollment) }}
                  </p>
                  <div class="mt-3 flex flex-wrap gap-2">
                    <button
                      type="button"
                      class="inline-flex items-center justify-center gap-2 rounded-lg bg-[#dc2626] px-3 py-2 text-sm font-medium text-white transition hover:bg-[#b91c1c] disabled:cursor-not-allowed disabled:opacity-60"
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
                      class="inline-flex items-center justify-center gap-2 rounded-lg border border-[#fecaca] bg-white px-3 py-2 text-sm font-medium text-[#dc2626] transition hover:bg-[#fef2f2] disabled:cursor-not-allowed disabled:opacity-60"
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

        <aside class="order-1 min-w-0 lg:order-2">
          <div class="space-y-5 lg:sticky lg:top-6">
            <section class="rounded-2xl border border-[#ebe7df] bg-white p-5">
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

            <section class="rounded-2xl border border-[#ebe7df] bg-white p-5">
              <div class="flex items-start justify-between gap-3">
                <div>
                  <p
                    class="text-[10px] font-medium uppercase tracking-[0.08em] text-[#9ca3af]"
                  >
                    Tambah penempatan
                  </p>
                  <h2 class="mt-1 text-base font-semibold text-[#171322]">
                    Pilih warga sekolah
                  </h2>
                  <p class="mt-1 text-xs leading-5 text-[#6b7280]">
                    Peran penempatan mengikuti peran warga sekolah.
                  </p>
                </div>
                <span
                  class="shrink-0 rounded-lg bg-[#eef2ff] px-2.5 py-1.5 text-xs font-medium text-[#4f46e5]"
                >
                  {{ selectedMembers.length }} dipilih
                </span>
              </div>

              <div class="mt-5 space-y-3">
                <label class="block text-xs font-medium text-[#6b7280]">
                  Cari warga sekolah
                  <div class="mt-2 flex gap-2">
                    <input
                      v-model="memberSearch"
                      type="search"
                      placeholder="Nama atau email"
                      class="min-w-0 flex-1 rounded-lg border border-[#ebe7df] bg-[#fbfaf8] px-3.5 py-2.5 text-sm text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#4f46e5] focus:bg-white"
                    />
                    <button
                      type="button"
                      class="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border border-[#ebe7df] bg-white text-[#6b7280] transition hover:border-[#4f46e5] hover:text-[#4f46e5] disabled:cursor-not-allowed disabled:opacity-60"
                      :disabled="membersLoading"
                      @click="loadMembers"
                    >
                      <PhMagnifyingGlass :size="17" weight="duotone" />
                    </button>
                  </div>
                </label>
              </div>

              <div class="mt-4">
                <div v-if="membersLoading" class="space-y-2">
                  <div
                    v-for="item in 2"
                    :key="item"
                    class="h-20 animate-pulse rounded-lg bg-[#fbfaf8]"
                  />
                </div>
                <div
                  v-else-if="membersError"
                  class="rounded-lg bg-[#fef2f2] p-3 text-xs leading-5 text-[#dc2626]"
                >
                  <p>{{ membersError }}</p>
                  <button
                    type="button"
                    class="mt-2 font-medium underline underline-offset-2"
                    @click="loadMembers"
                  >
                    Coba lagi
                  </button>
                </div>
                <div
                  v-else-if="!selectedClassId"
                  class="rounded-lg bg-[#fbfaf8] p-3 text-xs leading-5 text-[#6b7280]"
                >
                  Pilih kelas sebelum menambahkan warga sekolah.
                </div>
                <div
                  v-else-if="availableMembers.length === 0"
                  class="rounded-lg bg-[#fbfaf8] p-3 text-xs leading-5 text-[#6b7280]"
                >
                  Tidak ada warga sekolah yang dapat ditambahkan ke kelas ini.
                </div>
                <div
                  v-else-if="eligibleAvailableMembers.length === 0"
                  class="rounded-lg bg-[#fbfaf8] p-3 text-xs leading-5 text-[#6b7280]"
                >
                  Belum ada siswa atau guru yang dapat ditempatkan. Tambahkan
                  peran Siswa atau Guru di Warga Sekolah terlebih dahulu.
                </div>
                <div v-else class="max-h-72 space-y-2 overflow-y-auto pr-1">
                  <label
                    v-for="member in availableMembers"
                    :key="member.schoolUserId"
                    class="flex items-start gap-3 rounded-lg border border-[#ebe7df] bg-[#fbfaf8] p-3 transition"
                    :class="
                      inferPlacementRole(member)
                        ? 'cursor-pointer hover:border-[#d1d5db]'
                        : 'cursor-not-allowed opacity-65'
                    "
                  >
                    <input
                      type="checkbox"
                      class="mt-1 h-4 w-4 shrink-0 rounded border-[#d1d5db] text-[#4f46e5] focus:ring-[#4f46e5]"
                      :checked="
                        selectedSchoolUserIds.includes(member.schoolUserId)
                      "
                      :disabled="!inferPlacementRole(member)"
                      @change="toggleMember(member.schoolUserId)"
                    />
                    <span class="min-w-0 flex-1">
                      <span
                        class="block wrap-break-word text-sm font-medium text-[#171322]"
                      >
                        {{ member.fullName || "Nama tidak tersedia" }}
                      </span>
                      <span class="mt-1 block break-all text-xs text-[#6b7280]">
                        {{ member.email || "Email tidak tersedia" }}
                      </span>
                      <span
                        class="ml-2 mt-2 inline-flex rounded-lg px-2 py-1 text-[11px] font-medium"
                        :class="
                          inferPlacementRole(member) === 'teacher'
                            ? 'bg-[#eef2ff] text-[#4f46e5]'
                            : inferPlacementRole(member) === 'student'
                              ? 'bg-[#ecfdf5] text-[#059669]'
                              : 'bg-[#f3f4f6] text-[#6b7280]'
                        "
                      >
                        {{ placementRoleLabel(member) }}
                      </span>
                    </span>
                  </label>
                </div>
              </div>

              <button
                type="button"
                class="mt-4 inline-flex w-full items-center justify-center gap-2 rounded-lg bg-[#ea580c] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#c2410c] disabled:cursor-not-allowed disabled:opacity-60"
                :disabled="
                  submitting ||
                  !currentSchool.hasContext ||
                  !selectedClassId ||
                  selectedSchoolUserIds.length === 0
                "
                @click="submitEnrollment"
              >
                <PhStudent :size="17" weight="duotone" />
                {{ submitting ? "Menempatkan..." : "Tambahkan ke kelas" }}
              </button>
            </section>
          </div>
        </aside>
      </div>

      <RouterLink
        to="/admin/subject-classes"
        class="mt-5 flex items-center justify-between gap-4 rounded-2xl border border-[#ebe7df] bg-white p-5 transition hover:border-[#4f46e5] hover:shadow-sm"
      >
        <div>
          <p class="text-xs font-semibold uppercase tracking-[0.16em] text-[#ea580c]">
            Langkah berikutnya
          </p>
          <p class="mt-1 text-base font-semibold text-[#171322]">Buka Ruang Mengajar</p>
          <p class="mt-1 text-sm text-[#6b7280]">
            Setelah penempatan selesai, hubungkan guru ke kelas dan mata pelajaran.
          </p>
        </div>
        <PhArrowRight :size="20" class="shrink-0 text-[#4f46e5]" weight="bold" />
      </RouterLink>
    </section>
  </main>
</template>
