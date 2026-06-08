<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useAuthStore } from "../../stores/auth";
import {
  activateAcademicYear,
  activateTerm,
  createAcademicYear,
  createAssignmentCategory,
  createSubject,
  createTerm,
  deactivateAcademicYear,
  deactivateTerm,
  getAcademicYearsBySchool,
  getAssignmentCategoriesBySchool,
  getSubjectsBySchool,
  getTermsByAcademicYear,
} from "../../services/adminAcademic";
import type {
  AcademicYearItem,
  AssignmentCategoryItem,
  SubjectItem,
  TermItem,
} from "../../types/adminAcademic";
import { formatDateTime } from "../../utils/date";
import {
  PhBookOpen,
  PhCalendarBlank,
  PhChecks,
  PhPlusCircle,
  PhTag,
  PhWarningCircle,
} from "@phosphor-icons/vue";

const auth = useAuthStore();

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
const subjects = ref<SubjectItem[]>([]);
const categories = ref<AssignmentCategoryItem[]>([]);
const selectedAcademicYearId = ref("");

const academicYearsLoading = ref(false);
const academicYearsError = ref("");
const termsLoading = ref(false);
const termsError = ref("");
const subjectsLoading = ref(false);
const subjectsError = ref("");
const categoriesLoading = ref(false);
const categoriesError = ref("");
const actionMessage = ref("");
const actionError = ref("");
const activeAction = ref("");

const academicYearForm = ref({ academicYearName: "" });
const termForm = ref({ termName: "" });
const subjectForm = ref({ subjectName: "", subjectCode: "" });
const categoryForm = ref({ categoryName: "" });

const selectedAcademicYear = computed(
  () =>
    academicYears.value.find(
      (year) => year.academicYearId === selectedAcademicYearId.value,
    ) ?? null,
);

async function loadAcademicYears() {
  if (!currentSchool.value.hasContext) return;
  academicYearsLoading.value = true;
  academicYearsError.value = "";

  try {
    const data = await getAcademicYearsBySchool(currentSchool.value.schoolCode);
    academicYears.value = data.data ?? [];

    const activeYear =
      academicYears.value.find((year) => year.isActive) ??
      academicYears.value[0] ??
      null;

    if (!selectedAcademicYearId.value && activeYear) {
      selectedAcademicYearId.value = activeYear.academicYearId;
    }

    if (
      selectedAcademicYearId.value &&
      !academicYears.value.some(
        (year) => year.academicYearId === selectedAcademicYearId.value,
      )
    ) {
      selectedAcademicYearId.value = activeYear?.academicYearId ?? "";
    }
  } catch {
    academicYearsError.value = "Tahun ajaran belum bisa dimuat.";
  } finally {
    academicYearsLoading.value = false;
  }
}

async function loadTerms() {
  terms.value = [];
  termsError.value = "";

  if (!selectedAcademicYearId.value) return;

  termsLoading.value = true;
  try {
    const data = await getTermsByAcademicYear(selectedAcademicYearId.value);
    terms.value = data ?? [];
  } catch {
    termsError.value = "Semester belum bisa dimuat.";
  } finally {
    termsLoading.value = false;
  }
}

async function loadSubjects() {
  if (!currentSchool.value.hasContext) return;
  subjectsLoading.value = true;
  subjectsError.value = "";

  try {
    const data = await getSubjectsBySchool(currentSchool.value.schoolCode);
    subjects.value = data.subjects ?? [];
  } catch {
    subjectsError.value = "Mata pelajaran belum bisa dimuat.";
  } finally {
    subjectsLoading.value = false;
  }
}

async function loadCategories() {
  if (!currentSchool.value.hasContext) return;
  categoriesLoading.value = true;
  categoriesError.value = "";

  try {
    const data = await getAssignmentCategoriesBySchool(
      currentSchool.value.schoolCode,
    );
    categories.value = data.categories ?? [];
  } catch {
    categoriesError.value = "Kategori tugas belum bisa dimuat.";
  } finally {
    categoriesLoading.value = false;
  }
}

async function refreshAll() {
  await loadAcademicYears();
  await Promise.all([loadSubjects(), loadCategories()]);
  await loadTerms();
}

async function submitAcademicYear() {
  if (!currentSchool.value.schoolId) {
    actionError.value = "Context sekolah aktif belum tersedia.";
    return;
  }
  if (!academicYearForm.value.academicYearName.trim()) {
    actionError.value = "Nama tahun ajaran wajib diisi.";
    return;
  }

  activeAction.value = "academic-year-create";
  actionError.value = "";
  actionMessage.value = "";

  try {
    await createAcademicYear({
      schoolId: currentSchool.value.schoolId,
      academicYearName: academicYearForm.value.academicYearName.trim(),
    });
    academicYearForm.value.academicYearName = "";
    actionMessage.value = "Tahun ajaran berhasil dibuat.";
    await loadAcademicYears();
    await loadTerms();
  } catch {
    actionError.value = "Tahun ajaran belum bisa dibuat.";
  } finally {
    activeAction.value = "";
  }
}

async function toggleAcademicYear(year: AcademicYearItem) {
  activeAction.value = `academic-year-toggle-${year.academicYearId}`;
  actionError.value = "";
  actionMessage.value = "";

  try {
    if (year.isActive) {
      await deactivateAcademicYear(year.academicYearId);
      actionMessage.value = "Tahun ajaran dinonaktifkan.";
    } else {
      await activateAcademicYear(year.academicYearId);
      actionMessage.value = "Tahun ajaran diaktifkan.";
    }
    await loadAcademicYears();
    await loadTerms();
  } catch {
    actionError.value = "Perubahan status tahun ajaran belum bisa disimpan.";
  } finally {
    activeAction.value = "";
  }
}

async function submitTerm() {
  if (!selectedAcademicYearId.value) {
    actionError.value = "Pilih tahun ajaran terlebih dahulu.";
    return;
  }
  if (!termForm.value.termName.trim()) {
    actionError.value = "Nama semester wajib diisi.";
    return;
  }

  activeAction.value = "term-create";
  actionError.value = "";
  actionMessage.value = "";

  try {
    await createTerm({
      academicYearId: selectedAcademicYearId.value,
      termName: termForm.value.termName.trim(),
    });
    termForm.value.termName = "";
    actionMessage.value = "Semester berhasil dibuat.";
    await loadTerms();
  } catch {
    actionError.value = "Semester belum bisa dibuat.";
  } finally {
    activeAction.value = "";
  }
}

async function toggleTerm(term: TermItem) {
  activeAction.value = `term-toggle-${term.termId}`;
  actionError.value = "";
  actionMessage.value = "";

  try {
    if (term.isActive) {
      await deactivateTerm(term.termId);
      actionMessage.value = "Semester dinonaktifkan.";
    } else {
      await activateTerm(term.termId);
      actionMessage.value = "Semester diaktifkan.";
    }
    await loadTerms();
  } catch {
    actionError.value = "Perubahan status semester belum bisa disimpan.";
  } finally {
    activeAction.value = "";
  }
}

async function submitSubject() {
  if (!currentSchool.value.schoolId) {
    actionError.value = "Context sekolah aktif belum tersedia.";
    return;
  }
  if (
    !subjectForm.value.subjectName.trim() ||
    !subjectForm.value.subjectCode.trim()
  ) {
    actionError.value = "Nama dan kode mata pelajaran wajib diisi.";
    return;
  }

  activeAction.value = "subject-create";
  actionError.value = "";
  actionMessage.value = "";

  try {
    await createSubject({
      schoolId: currentSchool.value.schoolId,
      subjectName: subjectForm.value.subjectName.trim(),
      subjectCode: subjectForm.value.subjectCode.trim(),
    });
    subjectForm.value.subjectName = "";
    subjectForm.value.subjectCode = "";
    actionMessage.value = "Mata pelajaran berhasil dibuat.";
    await loadSubjects();
  } catch {
    actionError.value = "Mata pelajaran belum bisa dibuat.";
  } finally {
    activeAction.value = "";
  }
}

async function submitCategory() {
  if (!currentSchool.value.schoolId) {
    actionError.value = "Context sekolah aktif belum tersedia.";
    return;
  }
  if (!categoryForm.value.categoryName.trim()) {
    actionError.value = "Nama kategori wajib diisi.";
    return;
  }

  activeAction.value = "category-create";
  actionError.value = "";
  actionMessage.value = "";

  try {
    await createAssignmentCategory({
      schoolId: currentSchool.value.schoolId,
      categoryName: categoryForm.value.categoryName.trim(),
    });
    categoryForm.value.categoryName = "";
    actionMessage.value = "Kategori tugas berhasil dibuat.";
    await loadCategories();
  } catch {
    actionError.value = "Kategori tugas belum bisa dibuat.";
  } finally {
    activeAction.value = "";
  }
}

function isAcademicYearActionPending(yearId: string) {
  return activeAction.value === `academic-year-toggle-${yearId}`;
}

function isTermActionPending(termId: string) {
  return activeAction.value === `term-toggle-${termId}`;
}

onMounted(async () => {
  if (!currentSchool.value.hasContext) return;
  await refreshAll();
});
</script>

<template>
  <main class="min-h-screen flex-1 px-5 py-8 md:px-8 lg:px-10">
    <section class="mx-auto flex max-w-6xl flex-col gap-6">
      <header class="soft-card rounded-4xl p-6 md:p-8">
        <p class="text-sm font-medium text-[#4f46e5]">School admin workspace</p>
        <h1 class="mt-3 text-3xl font-medium text-[#171322]">Academic setup</h1>
        <p class="mt-3 max-w-3xl text-sm leading-6 text-[#6b6475]">
          Kelola tahun ajaran, semester, mata pelajaran, dan kategori tugas agar
          alur teacher dan student bisa berjalan tanpa seed manual.
        </p>

        <div class="mt-5 flex flex-wrap gap-2 text-sm">
          <span
            class="rounded-full bg-[#eef2ff] px-3 py-1 font-medium text-[#4f46e5]"
          >
            {{ currentSchool.schoolName || "Sekolah aktif belum tersedia" }}
          </span>
          <span
            class="rounded-full bg-[#f3efe8] px-3 py-1 font-medium text-[#6b6475]"
          >
            {{ currentSchool.schoolCode || "Kode sekolah tidak tersedia" }}
          </span>
          <span
            class="rounded-full bg-[#ecf8f1] px-3 py-1 font-medium text-[#4e8a73]"
          >
            {{ currentSchool.schoolId || "School ID belum tersedia" }}
          </span>
        </div>

        <div
          v-if="!currentSchool.hasContext"
          class="mt-5 rounded-2xl border border-[#f0c5bf] bg-[#fff8f6] px-4 py-3 text-sm text-[#a8665d]"
        >
          Context sekolah aktif belum ditemukan. Pastikan akun admin sudah punya
          membership yang valid.
        </div>

        <div
          v-if="actionMessage"
          class="mt-5 rounded-2xl border border-[#d8ecdf] bg-[#f5fbf7] px-4 py-3 text-sm text-[#4e8a73]"
        >
          {{ actionMessage }}
        </div>
        <div
          v-if="actionError"
          class="mt-5 rounded-2xl border border-[#f0c5bf] bg-[#fff8f6] px-4 py-3 text-sm text-[#a8665d]"
        >
          {{ actionError }}
        </div>
      </header>

      <section class="grid gap-6 lg:grid-cols-2">
        <article class="soft-card rounded-[28px] p-6">
          <div class="flex items-start justify-between gap-4">
            <div>
              <p class="text-sm font-medium text-[#4f46e5]">Academic Years</p>
              <h2 class="mt-2 text-xl font-medium text-[#171322]">
                Tahun ajaran
              </h2>
            </div>
            <PhCalendarBlank
              :size="24"
              class="text-[#7b61a8]"
              weight="duotone"
            />
          </div>

          <form
            class="mt-5 flex flex-col gap-3 sm:flex-row"
            @submit.prevent="submitAcademicYear"
          >
            <input
              v-model="academicYearForm.academicYearName"
              type="text"
              placeholder="Contoh: 2026/2027"
              class="w-full rounded-2xl border border-[#e7e1ec] bg-white px-4 py-3 text-sm text-[#171322] outline-none transition placeholder:text-[#b3acbb] focus:border-[#7b61a8]"
            />
            <button
              type="submit"
              class="inline-flex items-center justify-center gap-2 rounded-2xl bg-[#171322] px-4 py-3 text-sm font-medium text-white transition hover:bg-[#2f2b3a] disabled:cursor-not-allowed disabled:opacity-60"
              :disabled="
                activeAction === 'academic-year-create' ||
                !currentSchool.hasContext
              "
            >
              <PhPlusCircle :size="18" weight="duotone" />
              Tambah
            </button>
          </form>

          <div class="mt-5 space-y-3">
            <p v-if="academicYearsLoading" class="text-sm text-[#6b6475]">
              Memuat tahun ajaran...
            </p>
            <p v-else-if="academicYearsError" class="text-sm text-[#a8665d]">
              {{ academicYearsError }}
            </p>
            <p
              v-else-if="academicYears.length === 0"
              class="text-sm text-[#6b6475]"
            >
              Belum ada tahun ajaran untuk sekolah ini.
            </p>

            <article
              v-for="year in academicYears"
              :key="year.academicYearId"
              class="rounded-3xl border border-[#ece7f2] bg-[#fcfbfd] p-4"
            >
              <div
                class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between"
              >
                <div>
                  <div class="flex flex-wrap items-center gap-2">
                    <h3 class="text-base font-medium text-[#171322]">
                      {{ year.academicYearName }}
                    </h3>
                    <span
                      class="rounded-full px-2.5 py-1 text-xs font-medium"
                      :class="
                        year.isActive
                          ? 'bg-[#ecf8f1] text-[#4e8a73]'
                          : 'bg-[#f3efe8] text-[#6b6475]'
                      "
                    >
                      {{ year.isActive ? "Aktif" : "Nonaktif" }}
                    </span>
                  </div>
                  <p class="mt-2 text-sm text-[#6b6475]">
                    {{ year.schoolCode || currentSchool.schoolCode }} •
                    {{ formatDateTime(year.createdAt) }}
                  </p>
                </div>

                <button
                  type="button"
                  class="inline-flex items-center gap-2 rounded-2xl border border-[#e7e1ec] px-3 py-2 text-sm font-medium text-[#171322] transition hover:bg-white disabled:cursor-not-allowed disabled:opacity-60"
                  :disabled="isAcademicYearActionPending(year.academicYearId)"
                  @click="toggleAcademicYear(year)"
                >
                  <PhChecks v-if="!year.isActive" :size="16" weight="duotone" />
                  <PhWarningCircle v-else :size="16" weight="duotone" />
                  {{ year.isActive ? "Nonaktifkan" : "Aktifkan" }}
                </button>
              </div>
            </article>
          </div>
        </article>

        <article class="soft-card rounded-[28px] p-6">
          <div class="flex items-start justify-between gap-4">
            <div>
              <p class="text-sm font-medium text-[#4f46e5]">Terms</p>
              <h2 class="mt-2 text-xl font-medium text-[#171322]">Semester</h2>
            </div>
            <PhCalendarBlank
              :size="24"
              class="text-[#74bfa5]"
              weight="duotone"
            />
          </div>

          <label class="mt-5 block text-sm font-medium text-[#3f3a4a]">
            Tahun ajaran
            <select
              v-model="selectedAcademicYearId"
              class="mt-2 w-full rounded-2xl border border-[#e7e1ec] bg-white px-4 py-3 text-sm text-[#171322] outline-none transition focus:border-[#7b61a8]"
              @change="loadTerms"
            >
              <option value="" disabled>Pilih tahun ajaran</option>
              <option
                v-for="year in academicYears"
                :key="year.academicYearId"
                :value="year.academicYearId"
              >
                {{ year.academicYearName }}
              </option>
            </select>
          </label>

          <form
            class="mt-4 flex flex-col gap-3 sm:flex-row"
            @submit.prevent="submitTerm"
          >
            <input
              v-model="termForm.termName"
              type="text"
              placeholder="Contoh: Semester Ganjil"
              class="w-full rounded-2xl border border-[#e7e1ec] bg-white px-4 py-3 text-sm text-[#171322] outline-none transition placeholder:text-[#b3acbb] focus:border-[#7b61a8]"
            />
            <button
              type="submit"
              class="inline-flex items-center justify-center gap-2 rounded-2xl bg-[#171322] px-4 py-3 text-sm font-medium text-white transition hover:bg-[#2f2b3a] disabled:cursor-not-allowed disabled:opacity-60"
              :disabled="
                activeAction === 'term-create' || !selectedAcademicYearId
              "
            >
              <PhPlusCircle :size="18" weight="duotone" />
              Tambah
            </button>
          </form>

          <div class="mt-5 space-y-3">
            <p v-if="termsLoading" class="text-sm text-[#6b6475]">
              Memuat semester...
            </p>
            <p v-else-if="termsError" class="text-sm text-[#a8665d]">
              {{ termsError }}
            </p>
            <p
              v-else-if="!selectedAcademicYearId"
              class="text-sm text-[#6b6475]"
            >
              Pilih tahun ajaran untuk melihat semester.
            </p>
            <p v-else-if="terms.length === 0" class="text-sm text-[#6b6475]">
              Belum ada semester untuk tahun ajaran ini.
            </p>

            <article
              v-for="term in terms"
              :key="term.termId"
              class="rounded-3xl border border-[#ece7f2] bg-[#fcfbfd] p-4"
            >
              <div
                class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between"
              >
                <div>
                  <div class="flex flex-wrap items-center gap-2">
                    <h3 class="text-base font-medium text-[#171322]">
                      {{ term.termName }}
                    </h3>
                    <span
                      class="rounded-full px-2.5 py-1 text-xs font-medium"
                      :class="
                        term.isActive
                          ? 'bg-[#ecf8f1] text-[#4e8a73]'
                          : 'bg-[#f3efe8] text-[#6b6475]'
                      "
                    >
                      {{ term.isActive ? "Aktif" : "Nonaktif" }}
                    </span>
                  </div>
                  <p class="mt-2 text-sm text-[#6b6475]">
                    {{
                      term.academicYearName ||
                      selectedAcademicYear?.academicYearName ||
                      "Tahun ajaran"
                    }}
                    •
                    {{ formatDateTime(term.createdAt) }}
                  </p>
                </div>

                <button
                  type="button"
                  class="inline-flex items-center gap-2 rounded-2xl border border-[#e7e1ec] px-3 py-2 text-sm font-medium text-[#171322] transition hover:bg-white disabled:cursor-not-allowed disabled:opacity-60"
                  :disabled="isTermActionPending(term.termId)"
                  @click="toggleTerm(term)"
                >
                  <PhChecks v-if="!term.isActive" :size="16" weight="duotone" />
                  <PhWarningCircle v-else :size="16" weight="duotone" />
                  {{ term.isActive ? "Nonaktifkan" : "Aktifkan" }}
                </button>
              </div>
            </article>
          </div>
        </article>
      </section>

      <section class="grid gap-6 lg:grid-cols-2">
        <article class="soft-card rounded-[28px] p-6">
          <div class="flex items-start justify-between gap-4">
            <div>
              <p class="text-sm font-medium text-[#4f46e5]">Subjects</p>
              <h2 class="mt-2 text-xl font-medium text-[#171322]">
                Mata pelajaran
              </h2>
            </div>
            <PhBookOpen :size="24" class="text-[#74bfa5]" weight="duotone" />
          </div>

          <form
            class="mt-5 grid gap-3 sm:grid-cols-2"
            @submit.prevent="submitSubject"
          >
            <input
              v-model="subjectForm.subjectName"
              type="text"
              placeholder="Nama mata pelajaran"
              class="rounded-2xl border border-[#e7e1ec] bg-white px-4 py-3 text-sm text-[#171322] outline-none transition placeholder:text-[#b3acbb] focus:border-[#7b61a8]"
            />
            <input
              v-model="subjectForm.subjectCode"
              type="text"
              placeholder="Kode"
              class="rounded-2xl border border-[#e7e1ec] bg-white px-4 py-3 text-sm text-[#171322] outline-none transition placeholder:text-[#b3acbb] focus:border-[#7b61a8]"
            />
            <button
              type="submit"
              class="sm:col-span-2 inline-flex items-center justify-center gap-2 rounded-2xl bg-[#171322] px-4 py-3 text-sm font-medium text-white transition hover:bg-[#2f2b3a] disabled:cursor-not-allowed disabled:opacity-60"
              :disabled="
                activeAction === 'subject-create' || !currentSchool.hasContext
              "
            >
              <PhPlusCircle :size="18" weight="duotone" />
              Tambah mata pelajaran
            </button>
          </form>

          <div class="mt-5 space-y-3">
            <p v-if="subjectsLoading" class="text-sm text-[#6b6475]">
              Memuat mata pelajaran...
            </p>
            <p v-else-if="subjectsError" class="text-sm text-[#a8665d]">
              {{ subjectsError }}
            </p>
            <p v-else-if="subjects.length === 0" class="text-sm text-[#6b6475]">
              Belum ada mata pelajaran untuk sekolah ini.
            </p>

            <article
              v-for="subject in subjects"
              :key="subject.subjectId"
              class="rounded-3xl border border-[#ece7f2] bg-[#fcfbfd] p-4"
            >
              <h3 class="text-base font-medium text-[#171322]">
                {{ subject.subjectName }}
              </h3>
              <p class="mt-2 text-sm text-[#6b6475]">
                {{ subject.subjectCode }} •
                {{ subject.schoolCode || currentSchool.schoolCode }} •
                {{ formatDateTime(subject.createdAt) }}
              </p>
            </article>
          </div>
        </article>

        <article class="soft-card rounded-[28px] p-6">
          <div class="flex items-start justify-between gap-4">
            <div>
              <p class="text-sm font-medium text-[#4f46e5]">
                Assignment categories
              </p>
              <h2 class="mt-2 text-xl font-medium text-[#171322]">
                Kategori tugas
              </h2>
            </div>
            <PhTag :size="24" class="text-[#e58f86]" weight="duotone" />
          </div>

          <form
            class="mt-5 flex flex-col gap-3 sm:flex-row"
            @submit.prevent="submitCategory"
          >
            <input
              v-model="categoryForm.categoryName"
              type="text"
              placeholder="Contoh: Quiz"
              class="w-full rounded-2xl border border-[#e7e1ec] bg-white px-4 py-3 text-sm text-[#171322] outline-none transition placeholder:text-[#b3acbb] focus:border-[#7b61a8]"
            />
            <button
              type="submit"
              class="inline-flex items-center justify-center gap-2 rounded-2xl bg-[#171322] px-4 py-3 text-sm font-medium text-white transition hover:bg-[#2f2b3a] disabled:cursor-not-allowed disabled:opacity-60"
              :disabled="
                activeAction === 'category-create' || !currentSchool.hasContext
              "
            >
              <PhPlusCircle :size="18" weight="duotone" />
              Tambah
            </button>
          </form>

          <div class="mt-5 space-y-3">
            <p v-if="categoriesLoading" class="text-sm text-[#6b6475]">
              Memuat kategori tugas...
            </p>
            <p v-else-if="categoriesError" class="text-sm text-[#a8665d]">
              {{ categoriesError }}
            </p>
            <p
              v-else-if="categories.length === 0"
              class="text-sm text-[#6b6475]"
            >
              Belum ada kategori tugas untuk sekolah ini.
            </p>

            <article
              v-for="category in categories"
              :key="category.categoryId"
              class="rounded-3xl border border-[#ece7f2] bg-[#fcfbfd] p-4"
            >
              <h3 class="text-base font-medium text-[#171322]">
                {{ category.categoryName }}
              </h3>
              <p class="mt-2 text-sm text-[#6b6475]">
                {{ category.schoolId || currentSchool.schoolId }} •
                {{ formatDateTime(category.createdAt) }}
              </p>
            </article>
          </div>
        </article>
      </section>
    </section>
  </main>
</template>
