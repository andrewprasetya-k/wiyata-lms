<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  PhArrowLeft,
  PhBookOpen,
  PhCalendarBlank,
  PhChalkboardTeacher,
  PhMagnifyingGlass,
  PhStudent,
  PhToggleLeft,
  PhToggleRight,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { useToastStore } from "../../stores/toast";
import { getAdminClassById, updateAdminClass } from "../../services/adminClass";
import { getClassEnrollments } from "../../services/adminEnrollment";
import { getSubjectClassesByClass } from "../../services/adminSubjectClass";
import type { AdminClassItem } from "../../types/adminClass";
import type { EnrollmentMemberItem } from "../../types/adminEnrollment";
import type { SubjectClassItem } from "../../types/adminSubjectClass";
import { formatDateTime } from "../../utils/date";
import { resolveSubjectColor } from "../../utils/color";

const route = useRoute();
const router = useRouter();
const toast = useToastStore();

const classId = computed(() => String(route.params.classId ?? ""));

const classInfo = ref<AdminClassItem | null>(null);
const enrollments = ref<EnrollmentMemberItem[]>([]);
const subjectClasses = ref<SubjectClassItem[]>([]);

const classLoading = ref(false);
const enrollmentsLoading = ref(false);
const subjectClassesLoading = ref(false);
const classError = ref("");
const enrollmentsError = ref("");
const subjectClassesError = ref("");

const togglingActive = ref(false);

const activeTab = ref<"members" | "subjects">("members");

const memberSearch = ref("");

const students = computed(() =>
  enrollments.value.filter((e) => e.role === "student"),
);
const teachers = computed(() =>
  enrollments.value.filter((e) => e.role === "teacher"),
);

const filteredStudents = computed(() => {
  const q = memberSearch.value.trim().toLowerCase();
  if (!q) return students.value;
  return students.value.filter(
    (e) =>
      e.userFullName?.toLowerCase().includes(q) ||
      e.userEmail?.toLowerCase().includes(q),
  );
});

const filteredTeachers = computed(() => {
  const q = memberSearch.value.trim().toLowerCase();
  if (!q) return teachers.value;
  return teachers.value.filter(
    (e) =>
      e.userFullName?.toLowerCase().includes(q) ||
      e.userEmail?.toLowerCase().includes(q),
  );
});

async function loadClassInfo() {
  if (!classId.value) return;
  classLoading.value = true;
  classError.value = "";
  try {
    classInfo.value = await getAdminClassById(classId.value);
  } catch {
    classError.value = "Informasi kelas belum bisa dimuat.";
  } finally {
    classLoading.value = false;
  }
}

async function loadEnrollments() {
  if (!classId.value) return;
  enrollmentsLoading.value = true;
  enrollmentsError.value = "";
  try {
    const data = await getClassEnrollments(classId.value, { limit: 200 });
    enrollments.value = data.members?.data ?? [];
  } catch {
    enrollmentsError.value = "Daftar anggota kelas belum bisa dimuat.";
  } finally {
    enrollmentsLoading.value = false;
  }
}

async function loadSubjectClasses() {
  if (!classId.value) return;
  subjectClassesLoading.value = true;
  subjectClassesError.value = "";
  try {
    const data = await getSubjectClassesByClass(classId.value);
    subjectClasses.value = data.subjects ?? [];
  } catch {
    subjectClassesError.value = "Penugasan mengajar belum bisa dimuat.";
  } finally {
    subjectClassesLoading.value = false;
  }
}

async function toggleActive() {
  if (!classInfo.value) return;
  togglingActive.value = true;
  try {
    const updated = await updateAdminClass(classId.value, {
      isActive: !classInfo.value.isActive,
    });
    const merged = { ...classInfo.value!, ...updated } as AdminClassItem;
    classInfo.value = merged;
    toast.success(
      merged.isActive ? "Kelas diaktifkan." : "Kelas dinonaktifkan.",
    );
  } catch {
    toast.error("Status kelas belum bisa diubah.");
  } finally {
    togglingActive.value = false;
  }
}

onMounted(() => {
  Promise.all([loadClassInfo(), loadEnrollments(), loadSubjectClasses()]);
});
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-[#f8f7f4]">
    <!-- Header breadcrumb -->
    <header class="border-b border-[#ebe7df] bg-white">
      <div class="px-5 py-4 sm:px-6 lg:px-8">
        <nav
          class="flex items-center gap-2 text-sm text-[#6b7280]"
          aria-label="Breadcrumb"
        >
          <button
            type="button"
            class="inline-flex items-center gap-1.5 font-medium transition hover:text-[#4f46e5]"
            @click="router.back()"
          >
            <PhArrowLeft :size="15" />
            Kelas
          </button>
          <span class="text-[#d1ccd5]">/</span>
          <span class="max-w-50 truncate font-medium text-[#171322]">
            {{ classInfo?.classTitle || classId }}
          </span>
        </nav>
      </div>

      <div
        class="flex min-w-0 flex-col gap-4 px-5 pb-5 sm:px-6 lg:flex-row lg:items-end lg:justify-between lg:px-8"
      >
        <div class="min-w-0">
          <div
            v-if="classLoading"
            class="h-8 w-48 animate-pulse rounded-lg bg-[#f3f1ec]"
          />
          <template v-else-if="classInfo">
            <div class="flex min-w-0 flex-wrap items-center gap-2">
              <h1 class="text-2xl font-semibold text-[#171322] sm:text-3xl">
                {{ classInfo.classTitle }}
              </h1>
              <span
                class="rounded-lg px-2.5 py-1 text-xs font-medium"
                :class="
                  classInfo.isActive
                    ? 'bg-[#f0fdf4] text-[#059669]'
                    : 'bg-[#f3f1ec] text-[#6b7280]'
                "
              >
                {{ classInfo.isActive ? "Aktif" : "Nonaktif" }}
              </span>
            </div>
            <div
              class="mt-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-[#6b7280]"
            >
              <span class="flex items-center gap-1.5">
                <PhBookOpen :size="15" weight="duotone" />
                {{ classInfo.classCode }}
              </span>
              <span v-if="classInfo.termName" class="flex items-center gap-1.5">
                <PhCalendarBlank :size="15" weight="duotone" />
                {{ classInfo.termName }}
                <span v-if="classInfo.academicYearName">
                  · {{ classInfo.academicYearName }}
                </span>
              </span>
            </div>
          </template>
          <p v-else-if="classError" class="text-sm text-[#dc2626]">
            {{ classError }}
          </p>
        </div>

        <button
          v-if="classInfo"
          type="button"
          class="inline-flex items-center gap-2 rounded-lg border px-4 py-2.5 text-sm font-medium transition disabled:cursor-not-allowed disabled:opacity-60"
          :class="
            classInfo.isActive
              ? 'border-[#fecaca] bg-white text-[#dc2626] hover:bg-[#fef2f2]'
              : 'border-[#bbf7d0] bg-white text-[#059669] hover:bg-[#f0fdf4]'
          "
          :disabled="togglingActive"
          @click="toggleActive"
        >
          <component
            :is="classInfo.isActive ? PhToggleRight : PhToggleLeft"
            :size="16"
            weight="fill"
          />
          {{
            togglingActive
              ? "Mengubah..."
              : classInfo.isActive
                ? "Nonaktifkan kelas"
                : "Aktifkan kelas"
          }}
        </button>
      </div>
    </header>

    <section class="px-5 py-6 sm:px-6 lg:px-8">
      <!-- Stat cards -->
      <div v-if="classInfo" class="grid grid-cols-2 gap-3 sm:grid-cols-4">
        <article
          class="rounded-xl border border-[#ebe7df] bg-white p-4 shadow-sm"
        >
          <p class="text-xs font-medium text-[#6b7280]">Siswa</p>
          <p class="mt-2 text-2xl font-semibold text-[#171322]">
            {{ enrollmentsLoading ? "–" : students.length }}
          </p>
        </article>
        <article
          class="rounded-xl border border-[#ebe7df] bg-white p-4 shadow-sm"
        >
          <p class="text-xs font-medium text-[#6b7280]">Guru</p>
          <p class="mt-2 text-2xl font-semibold text-[#171322]">
            {{ enrollmentsLoading ? "–" : teachers.length }}
          </p>
        </article>
        <article
          class="rounded-xl border border-[#ebe7df] bg-white p-4 shadow-sm"
        >
          <p class="text-xs font-medium text-[#6b7280]">Ruang mengajar</p>
          <p class="mt-2 text-2xl font-semibold text-[#171322]">
            {{ subjectClassesLoading ? "–" : subjectClasses.length }}
          </p>
        </article>
        <article
          class="rounded-xl border border-[#ebe7df] bg-white p-4 shadow-sm"
        >
          <p class="text-xs font-medium text-[#6b7280]">Dibuat oleh</p>
          <p class="mt-2 truncate text-sm font-semibold text-[#171322]">
            {{ classInfo.creatorName || "–" }}
          </p>
          <p class="mt-0.5 text-[11px] text-[#9ca3af]">
            {{ formatDateTime(classInfo.createdAt) }}
          </p>
        </article>
      </div>

      <!-- Tab switcher -->
      <div
        class="mt-6 flex gap-1 rounded-xl border border-[#ebe7df] bg-white p-1 shadow-sm w-fit"
      >
        <button
          type="button"
          class="rounded-lg px-4 py-2 text-sm font-medium transition"
          :class="
            activeTab === 'members'
              ? 'bg-[#eef2ff] text-[#4f46e5]'
              : 'text-[#6b7280] hover:text-[#374151]'
          "
          @click="activeTab = 'members'"
        >
          <span class="flex items-center gap-2">
            <PhStudent :size="16" weight="duotone" />
            Anggota Kelas
            <span
              v-if="!enrollmentsLoading"
              class="rounded-full bg-[#f3f1ec] px-2 py-0.5 text-[11px] font-semibold text-[#6b7280]"
            >
              {{ enrollments.length }}
            </span>
          </span>
        </button>
        <button
          type="button"
          class="rounded-lg px-4 py-2 text-sm font-medium transition"
          :class="
            activeTab === 'subjects'
              ? 'bg-[#eef2ff] text-[#4f46e5]'
              : 'text-[#6b7280] hover:text-[#374151]'
          "
          @click="activeTab = 'subjects'"
        >
          <span class="flex items-center gap-2">
            <PhChalkboardTeacher :size="16" weight="duotone" />
            Penugasan Mengajar
            <span
              v-if="!subjectClassesLoading"
              class="rounded-full bg-[#f3f1ec] px-2 py-0.5 text-[11px] font-semibold text-[#6b7280]"
            >
              {{ subjectClasses.length }}
            </span>
          </span>
        </button>
      </div>

      <!-- Tab: Anggota Kelas -->
      <template v-if="activeTab === 'members'">
        <div class="mt-4 rounded-xl border border-[#ebe7df] bg-white shadow-sm">
          <div
            class="flex flex-col gap-4 border-b border-[#ebe7df] p-5 sm:flex-row sm:items-center sm:justify-between"
          >
            <div>
              <p class="eyebrow-muted">Anggota kelas</p>
              <h2 class="mt-1 text-base font-semibold text-[#171322]">
                Siswa dan guru di kelas ini
              </h2>
            </div>
            <label class="relative block w-full sm:max-w-xs">
              <PhMagnifyingGlass
                :size="16"
                class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-[#9ca3af]"
              />
              <input
                v-model="memberSearch"
                type="search"
                placeholder="Cari nama atau email..."
                class="w-full rounded-lg border border-[#ebe7df] bg-[#fbfaf8] py-2 pl-9 pr-3 text-sm text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#4f46e5] focus:bg-white"
              />
            </label>
          </div>

          <div class="p-5">
            <div v-if="enrollmentsLoading" class="space-y-3">
              <div
                v-for="i in 4"
                :key="i"
                class="h-14 animate-pulse rounded-lg bg-[#fbfaf8]"
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
                Gagal memuat anggota
              </h3>
              <p class="mt-2 text-sm text-[#6b7280]">{{ enrollmentsError }}</p>
              <button
                type="button"
                class="mt-4 inline-flex items-center gap-2 rounded-lg border border-[#ebe7df] bg-white px-4 py-2 text-sm font-medium text-[#374151] transition hover:border-[#4f46e5] hover:text-[#4f46e5]"
                @click="loadEnrollments"
              >
                Coba lagi
              </button>
            </div>

            <div
              v-else-if="enrollments.length === 0"
              class="rounded-lg bg-[#fbfaf8] px-5 py-8 text-center"
            >
              <PhStudent
                class="mx-auto h-7 w-7 text-[#9ca3af]"
                weight="duotone"
              />
              <h3 class="mt-3 text-sm font-semibold text-[#171322]">
                Belum ada anggota
              </h3>
              <p class="mt-2 text-sm text-[#6b7280]">
                Tambahkan siswa dan guru melalui halaman Penempatan Kelas.
              </p>
            </div>

            <template v-else>
              <!-- Siswa -->
              <section v-if="filteredStudents.length > 0" class="mb-6">
                <h3 class="eyebrow-muted mb-3">
                  Siswa ({{ filteredStudents.length }})
                </h3>
                <div class="divide-y divide-[#f3f4f6]">
                  <div
                    v-for="member in filteredStudents"
                    :key="member.enrollmentId"
                    class="flex items-center gap-3 py-3 first:pt-0"
                  >
                    <div
                      class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-[#eef2ff] text-xs font-semibold text-[#4f46e5]"
                    >
                      {{ (member.userFullName || "S").charAt(0).toUpperCase() }}
                    </div>
                    <div class="min-w-0 flex-1">
                      <p class="truncate text-sm font-medium text-[#171322]">
                        {{ member.userFullName || "Nama tidak tersedia" }}
                      </p>
                      <p class="truncate text-xs text-[#9ca3af]">
                        {{ member.userEmail || "–" }}
                      </p>
                    </div>
                    <span
                      class="rounded-lg bg-[#eef2ff] px-2 py-1 text-[11px] font-medium text-[#4f46e5]"
                    >
                      Siswa
                    </span>
                  </div>
                </div>
              </section>

              <!-- Guru -->
              <section v-if="filteredTeachers.length > 0">
                <h3 class="eyebrow-muted mb-3">
                  Guru ({{ filteredTeachers.length }})
                </h3>
                <div class="divide-y divide-[#f3f4f6]">
                  <div
                    v-for="member in filteredTeachers"
                    :key="member.enrollmentId"
                    class="flex items-center gap-3 py-3 first:pt-0"
                  >
                    <div
                      class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-[#fff4ee] text-xs font-semibold text-[#ea580c]"
                    >
                      {{ (member.userFullName || "G").charAt(0).toUpperCase() }}
                    </div>
                    <div class="min-w-0 flex-1">
                      <p class="truncate text-sm font-medium text-[#171322]">
                        {{ member.userFullName || "Nama tidak tersedia" }}
                      </p>
                      <p class="truncate text-xs text-[#9ca3af]">
                        {{ member.userEmail || "–" }}
                      </p>
                    </div>
                    <span
                      class="rounded-lg bg-[#fff4ee] px-2 py-1 text-[11px] font-medium text-[#ea580c]"
                    >
                      Guru
                    </span>
                  </div>
                </div>
              </section>

              <div
                v-if="
                  memberSearch &&
                  filteredStudents.length === 0 &&
                  filteredTeachers.length === 0
                "
                class="rounded-lg bg-[#fbfaf8] px-5 py-8 text-center"
              >
                <PhMagnifyingGlass
                  class="mx-auto h-7 w-7 text-[#9ca3af]"
                  weight="duotone"
                />
                <h3 class="mt-3 text-sm font-semibold text-[#171322]">
                  Tidak ada anggota yang cocok
                </h3>
                <p class="mt-2 text-sm text-[#6b7280]">
                  Ubah kata kunci pencarian.
                </p>
              </div>
            </template>
          </div>
        </div>
      </template>

      <!-- Tab: Penugasan Mengajar -->
      <template v-if="activeTab === 'subjects'">
        <div class="mt-4 rounded-xl border border-[#ebe7df] bg-white shadow-sm">
          <div class="border-b border-[#ebe7df] p-5">
            <p class="eyebrow-muted">Penugasan mengajar</p>
            <h2 class="mt-1 text-base font-semibold text-[#171322]">
              Ruang mengajar yang terdaftar di kelas ini
            </h2>
          </div>

          <div class="p-5">
            <div v-if="subjectClassesLoading" class="space-y-3">
              <div
                v-for="i in 3"
                :key="i"
                class="h-16 animate-pulse rounded-lg bg-[#fbfaf8]"
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
                Gagal memuat penugasan
              </h3>
              <p class="mt-2 text-sm text-[#6b7280]">
                {{ subjectClassesError }}
              </p>
              <button
                type="button"
                class="mt-4 inline-flex items-center gap-2 rounded-lg border border-[#ebe7df] bg-white px-4 py-2 text-sm font-medium text-[#374151] transition hover:border-[#4f46e5] hover:text-[#4f46e5]"
                @click="loadSubjectClasses"
              >
                Coba lagi
              </button>
            </div>

            <div
              v-else-if="subjectClasses.length === 0"
              class="rounded-lg bg-[#fbfaf8] px-5 py-8 text-center"
            >
              <PhChalkboardTeacher
                class="mx-auto h-7 w-7 text-[#9ca3af]"
                weight="duotone"
              />
              <h3 class="mt-3 text-sm font-semibold text-[#171322]">
                Belum ada penugasan mengajar
              </h3>
              <p class="mt-2 text-sm text-[#6b7280]">
                Hubungkan guru ke mata pelajaran melalui halaman Penugasan
                Mengajar.
              </p>
            </div>

            <div v-else class="divide-y divide-[#f3f4f6]">
              <div
                v-for="sc in subjectClasses"
                :key="sc.subjectClassId"
                class="flex items-center gap-4 py-3 first:pt-0"
              >
                <div
                  class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl text-white text-xs font-bold"
                  :style="{ backgroundColor: resolveSubjectColor(sc) }"
                >
                  {{
                    (sc.subjectCode || sc.subjectName || "?")
                      .slice(0, 2)
                      .toUpperCase()
                  }}
                </div>
                <div class="min-w-0 flex-1">
                  <p class="truncate text-sm font-semibold text-[#171322]">
                    {{ sc.subjectName || sc.subjectCode || "Mata pelajaran" }}
                  </p>
                  <p class="truncate text-xs text-[#6b7280]">
                    <span class="flex items-center gap-1">
                      <PhChalkboardTeacher :size="12" />
                      {{ sc.teacherName || "Guru belum ditentukan" }}
                    </span>
                  </p>
                </div>
                <span
                  v-if="sc.subjectCode"
                  class="rounded-lg bg-[#f3f1ec] px-2 py-1 text-[11px] font-medium text-[#6b7280]"
                >
                  {{ sc.subjectCode }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </template>
    </section>
  </main>
</template>
