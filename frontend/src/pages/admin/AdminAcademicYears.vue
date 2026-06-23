<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useAuthStore } from "../../stores/auth";
import { useToastStore } from "../../stores/toast";
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
  getAssessmentWeightsBySubject,
  getSubjectsBySchool,
  getTermsByAcademicYear,
  saveAssessmentWeights,
} from "../../services/adminAcademic";
import type {
  AcademicYearItem,
  AssessmentWeightItem,
  AssignmentCategoryItem,
  SubjectItem,
  TermItem,
} from "../../types/adminAcademic";
import { formatDateTime } from "../../utils/date";
import {
  PhBookOpen,
  PhCalendarBlank,
  PhChartBar,
  PhChecks,
  PhPlusCircle,
  PhTag,
  PhWarningCircle,
} from "@phosphor-icons/vue";

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
const subjects = ref<SubjectItem[]>([]);
const categories = ref<AssignmentCategoryItem[]>([]);
const selectedAcademicYearId = ref("");
const selectedWeightSubjectId = ref("");
const weightInputs = ref<Record<string, string>>({});

const academicYearsLoading = ref(false);
const academicYearsError = ref("");
const termsLoading = ref(false);
const termsError = ref("");
const subjectsLoading = ref(false);
const subjectsError = ref("");
const categoriesLoading = ref(false);
const categoriesError = ref("");
const weightsLoading = ref(false);
const weightsError = ref("");
const weightsInfoMessage = ref("");
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

const selectedWeightSubject = computed(
  () =>
    subjects.value.find(
      (subject) => subject.subjectId === selectedWeightSubjectId.value,
    ) ?? null,
);

const totalWeight = computed(() =>
  categories.value.reduce(
    (total, category) => total + parseWeightValue(weightInputs.value[category.categoryId]),
    0,
  ),
);

const hasInvalidWeight = computed(() =>
  categories.value.some((category) => {
    const rawValue = weightInputs.value[category.categoryId];
    if (rawValue === "" || rawValue === undefined) return false;
    const value = Number(rawValue);
    return Number.isNaN(value) || value < 0 || value > 100;
  }),
);

const isWeightTotalValid = computed(
  () => Math.abs(totalWeight.value - 100) <= 0.01,
);

const canSubmitWeights = computed(
  () =>
    currentSchool.value.hasContext &&
    Boolean(selectedWeightSubjectId.value) &&
    categories.value.length > 0 &&
    !hasInvalidWeight.value &&
    isWeightTotalValid.value &&
    activeAction.value !== "weights-save",
);

function parseWeightValue(value?: string) {
  if (value === undefined || value.trim() === "") return 0;
  const numericValue = Number(value);
  if (Number.isNaN(numericValue)) return 0;
  return numericValue;
}

function formatWeight(value: number) {
  return new Intl.NumberFormat("id-ID", {
    maximumFractionDigits: 2,
  }).format(value);
}

function getApiErrorMessage(error: unknown, fallback: string) {
  if (typeof error === "object" && error !== null && "response" in error) {
    const response = (error as {
      response?: { data?: { error?: unknown; message?: unknown } };
    }).response;
    if (typeof response?.data?.error === "string") return response.data.error;
    if (typeof response?.data?.message === "string") return response.data.message;
  }

  return fallback;
}

function getApiStatus(error: unknown) {
  if (typeof error === "object" && error !== null && "response" in error) {
    return (error as { response?: { status?: number } }).response?.status;
  }

  return undefined;
}

function resetWeightInputs(weights: AssessmentWeightItem[] = []) {
  const nextInputs: Record<string, string> = {};
  const weightMap = new Map(
    weights.map((weight) => [weight.categoryId, String(weight.weight)]),
  );

  for (const category of categories.value) {
    nextInputs[category.categoryId] = weightMap.get(category.categoryId) ?? "0";
  }

  weightInputs.value = nextInputs;
}

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
    if (
      !selectedWeightSubjectId.value ||
      !subjects.value.some(
        (subject) => subject.subjectId === selectedWeightSubjectId.value,
      )
    ) {
      selectedWeightSubjectId.value = subjects.value[0]?.subjectId ?? "";
    }
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
    resetWeightInputs();
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
  await loadAssessmentWeights();
}

async function submitAcademicYear() {
  if (!currentSchool.value.schoolId) {
    toast.error("Context sekolah aktif belum tersedia.");
    return;
  }
  if (!academicYearForm.value.academicYearName.trim()) {
    toast.error("Nama tahun ajaran wajib diisi.");
    return;
  }

  activeAction.value = "academic-year-create";

  try {
    await createAcademicYear({
      schoolId: currentSchool.value.schoolId,
      academicYearName: academicYearForm.value.academicYearName.trim(),
    });
    academicYearForm.value.academicYearName = "";
    toast.success("Tahun ajaran berhasil dibuat.");
    await loadAcademicYears();
    await loadTerms();
  } catch {
    toast.error("Tahun ajaran belum bisa dibuat.");
  } finally {
    activeAction.value = "";
  }
}

async function toggleAcademicYear(year: AcademicYearItem) {
  activeAction.value = `academic-year-toggle-${year.academicYearId}`;

  try {
    if (year.isActive) {
      await deactivateAcademicYear(year.academicYearId);
      toast.success("Tahun ajaran dinonaktifkan.");
    } else {
      await activateAcademicYear(year.academicYearId);
      toast.success("Tahun ajaran diaktifkan.");
    }
    await loadAcademicYears();
    await loadTerms();
  } catch {
    toast.error("Perubahan status tahun ajaran belum bisa disimpan.");
  } finally {
    activeAction.value = "";
  }
}

async function submitTerm() {
  if (!selectedAcademicYearId.value) {
    toast.error("Pilih tahun ajaran terlebih dahulu.");
    return;
  }
  if (!termForm.value.termName.trim()) {
    toast.error("Nama semester wajib diisi.");
    return;
  }

  activeAction.value = "term-create";

  try {
    await createTerm({
      academicYearId: selectedAcademicYearId.value,
      termName: termForm.value.termName.trim(),
    });
    termForm.value.termName = "";
    toast.success("Semester berhasil dibuat.");
    await loadTerms();
  } catch {
    toast.error("Semester belum bisa dibuat.");
  } finally {
    activeAction.value = "";
  }
}

async function toggleTerm(term: TermItem) {
  activeAction.value = `term-toggle-${term.termId}`;

  try {
    if (term.isActive) {
      await deactivateTerm(term.termId);
      toast.success("Semester dinonaktifkan.");
    } else {
      await activateTerm(term.termId);
      toast.success("Semester diaktifkan.");
    }
    await loadTerms();
  } catch {
    toast.error("Perubahan status semester belum bisa disimpan.");
  } finally {
    activeAction.value = "";
  }
}

async function submitSubject() {
  if (!currentSchool.value.schoolId) {
    toast.error("Context sekolah aktif belum tersedia.");
    return;
  }
  if (
    !subjectForm.value.subjectName.trim() ||
    !subjectForm.value.subjectCode.trim()
  ) {
    toast.error("Nama dan kode mata pelajaran wajib diisi.");
    return;
  }

  activeAction.value = "subject-create";

  try {
    await createSubject({
      schoolId: currentSchool.value.schoolId,
      subjectName: subjectForm.value.subjectName.trim(),
      subjectCode: subjectForm.value.subjectCode.trim(),
    });
    subjectForm.value.subjectName = "";
    subjectForm.value.subjectCode = "";
    toast.success("Mata pelajaran berhasil dibuat.");
    await loadSubjects();
  } catch {
    toast.error("Mata pelajaran belum bisa dibuat.");
  } finally {
    activeAction.value = "";
  }
}

async function submitCategory() {
  if (!currentSchool.value.schoolId) {
    toast.error("Context sekolah aktif belum tersedia.");
    return;
  }
  if (!categoryForm.value.categoryName.trim()) {
    toast.error("Nama kategori wajib diisi.");
    return;
  }

  activeAction.value = "category-create";

  try {
    await createAssignmentCategory({
      schoolId: currentSchool.value.schoolId,
      categoryName: categoryForm.value.categoryName.trim(),
    });
    categoryForm.value.categoryName = "";
    toast.success("Kategori tugas berhasil dibuat.");
    await loadCategories();
    await loadAssessmentWeights();
  } catch {
    toast.error("Kategori tugas belum bisa dibuat.");
  } finally {
    activeAction.value = "";
  }
}

async function loadAssessmentWeights() {
  weightsError.value = "";
  weightsInfoMessage.value = "";

  if (!selectedWeightSubjectId.value || categories.value.length === 0) {
    resetWeightInputs();
    return;
  }

  weightsLoading.value = true;
  resetWeightInputs();

  try {
    const response = await getAssessmentWeightsBySubject(selectedWeightSubjectId.value);
    resetWeightInputs(response.weights ?? []);
    weightsInfoMessage.value = response.weights?.length
      ? "Bobot tersimpan sudah dimuat."
      : "";
  } catch (error) {
    resetWeightInputs();
    if (getApiStatus(error) === 404) {
      weightsInfoMessage.value =
        "Bobot belum dikonfigurasi untuk mata pelajaran ini.";
    } else {
      weightsError.value = getApiErrorMessage(
        error,
        "Bobot nilai belum bisa dimuat.",
      );
    }
  } finally {
    weightsLoading.value = false;
  }
}

async function submitAssessmentWeights() {
  if (!selectedWeightSubjectId.value) {
    toast.error("Pilih mata pelajaran terlebih dahulu.");
    return;
  }
  if (categories.value.length === 0) {
    toast.error("Tambahkan kategori tugas terlebih dahulu.");
    return;
  }
  if (hasInvalidWeight.value) {
    toast.error("Bobot harus berada di antara 0 sampai 100.");
    return;
  }
  if (!isWeightTotalValid.value) {
    toast.error("Total bobot harus 100%.");
    return;
  }

  activeAction.value = "weights-save";
  weightsError.value = "";

  try {
    await saveAssessmentWeights({
      subjectId: selectedWeightSubjectId.value,
      weights: categories.value.map((category) => ({
        categoryId: category.categoryId,
        weight: parseWeightValue(weightInputs.value[category.categoryId]),
      })),
    });
    toast.success("Bobot nilai berhasil disimpan.");
    await loadAssessmentWeights();
  } catch (error) {
    toast.error(getApiErrorMessage(error, "Bobot nilai belum bisa disimpan."));
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

watch(selectedWeightSubjectId, () => {
  loadAssessmentWeights();
});
</script>

<template>
  <main class="min-h-screen flex-1 px-5 py-5 sm:px-6 lg:px-8">
    <section class="flex w-full max-w-none flex-col gap-5">
      <header class="soft-card rounded-[22px] p-5 md:p-6">
        <p class="text-sm font-medium text-[#4f46e5]">Admin sekolah</p>
        <h1 class="mt-3 text-3xl font-medium text-[#171322]">
          Struktur Akademik
        </h1>
        <p class="mt-3 max-w-3xl text-sm leading-6 text-[#6b6475]">
          Kelola tahun ajaran, semester, mata pelajaran, dan kategori tugas
          sebagai fondasi kelas, konten belajar, dan penilaian.
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

      </header>

      <section class="grid gap-5 lg:grid-cols-2">
        <article class="soft-card rounded-[18px] p-5">
          <div class="flex items-start justify-between gap-4">
            <div>
              <p class="text-sm font-medium text-[#4f46e5]">Struktur akademik</p>
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

        <article class="soft-card rounded-[18px] p-5">
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

      <section class="grid gap-5 lg:grid-cols-2">
        <article class="soft-card rounded-[18px] p-5">
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

        <article class="soft-card rounded-[18px] p-5">
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

      <section class="soft-card rounded-[18px] p-5">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div class="max-w-3xl">
            <p class="text-sm font-medium text-[#4f46e5]">Grading policy</p>
            <h2 class="mt-2 text-xl font-medium text-[#171322]">Bobot Nilai</h2>
            <p class="mt-2 text-sm leading-6 text-[#6b6475]">
              Bobot berlaku per mata pelajaran dan digunakan untuk menghitung
              Rata-rata berbobot sementara. Ini bukan nilai rapor final resmi.
            </p>
          </div>
          <div
            class="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-[#eef2ff] text-[#4f46e5]"
          >
            <PhChartBar :size="24" weight="duotone" />
          </div>
        </div>

        <div
          v-if="subjects.length === 0"
          class="mt-5 rounded-2xl bg-[#fbfaf8] p-4 text-sm leading-6 text-[#6b6475]"
        >
          Tambahkan mata pelajaran terlebih dahulu sebelum mengatur bobot nilai.
        </div>

        <div
          v-else-if="categories.length === 0"
          class="mt-5 rounded-2xl bg-[#fbfaf8] p-4 text-sm leading-6 text-[#6b6475]"
        >
          Tambahkan kategori tugas terlebih dahulu sebelum mengatur bobot nilai.
        </div>

        <div
          v-else
          class="mt-5 grid gap-5 xl:grid-cols-[minmax(0,0.85fr)_minmax(0,1.15fr)]"
        >
          <div class="rounded-2xl border border-[#ece7f2] bg-[#fcfbfd] p-4">
            <label class="block text-sm font-medium text-[#3f3a4a]">
              Mata pelajaran
              <select
                v-model="selectedWeightSubjectId"
                class="mt-2 w-full rounded-2xl border border-[#e7e1ec] bg-white px-4 py-3 text-sm text-[#171322] outline-none transition focus:border-[#7b61a8]"
              >
                <option value="" disabled>Pilih mata pelajaran</option>
                <option
                  v-for="subject in subjects"
                  :key="subject.subjectId"
                  :value="subject.subjectId"
                >
                  {{ subject.subjectName }} · {{ subject.subjectCode }}
                </option>
              </select>
            </label>

            <div class="mt-4 rounded-2xl bg-white p-4">
              <p class="text-xs text-[#8b8592]">Total bobot</p>
              <div class="mt-2 flex flex-wrap items-end justify-between gap-3">
                <p
                  class="text-3xl font-medium"
                  :class="isWeightTotalValid ? 'text-[#027a48]' : 'text-[#b45309]'"
                >
                  {{ formatWeight(totalWeight) }}%
                </p>
                <span
                  class="rounded-full px-3 py-1 text-xs font-medium"
                  :class="
                    isWeightTotalValid
                      ? 'bg-[#ecf8f1] text-[#4e8a73]'
                      : 'bg-[#fff7ed] text-[#b45309]'
                  "
                >
                  {{ isWeightTotalValid ? "Valid" : "Harus 100%" }}
                </span>
              </div>
              <p class="mt-3 text-xs leading-5 text-[#7a7385]">
                Total bobot harus 100% sebelum disimpan.
              </p>
            </div>

            <p
              v-if="selectedWeightSubject"
              class="mt-4 rounded-2xl bg-[#eef2ff] px-4 py-3 text-xs leading-5 text-[#4f46e5]"
            >
              Bobot yang disimpan akan berlaku untuk semua kelas pada mata
              pelajaran {{ selectedWeightSubject.subjectName }}.
            </p>
          </div>

          <form
            class="rounded-2xl border border-[#ece7f2] bg-[#fcfbfd] p-4"
            @submit.prevent="submitAssessmentWeights"
          >
            <div
              class="flex flex-col gap-3 border-b border-[#ece7f2] pb-4 sm:flex-row sm:items-start sm:justify-between"
            >
              <div>
                <p class="text-sm font-medium text-[#171322]">Kategori dan bobot</p>
                <p class="mt-1 text-xs leading-5 text-[#7a7385]">
                  Kosong dianggap 0. Setiap kategori hanya muncul satu kali.
                </p>
              </div>
              <button
                class="inline-flex items-center justify-center gap-2 rounded-2xl bg-[#171322] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#2f2b3a] disabled:cursor-not-allowed disabled:opacity-60"
                type="submit"
                :disabled="!canSubmitWeights"
              >
                <PhChecks :size="16" weight="duotone" />
                {{ activeAction === "weights-save" ? "Menyimpan..." : "Simpan bobot" }}
              </button>
            </div>

            <div v-if="weightsLoading" class="mt-4 space-y-3">
              <div
                v-for="item in 3"
                :key="item"
                class="h-14 animate-pulse rounded-2xl bg-white"
              />
            </div>

            <div v-else class="mt-4 space-y-3">
              <p
                v-if="weightsError"
                class="rounded-2xl bg-[#fff8f6] px-4 py-3 text-sm leading-6 text-[#a8665d]"
              >
                {{ weightsError }}
              </p>
              <p
                v-else-if="weightsInfoMessage"
                class="rounded-2xl bg-[#eef2ff] px-4 py-3 text-sm leading-6 text-[#4f46e5]"
              >
                {{ weightsInfoMessage }}
              </p>

              <div
                v-for="category in categories"
                :key="category.categoryId"
                class="grid gap-3 rounded-2xl bg-white p-4 sm:grid-cols-[minmax(0,1fr)_140px]"
              >
                <div class="min-w-0">
                  <p class="truncate text-sm font-medium text-[#171322]">
                    {{ category.categoryName }}
                  </p>
                  <p class="mt-1 text-xs text-[#8b8592]">
                    Kategori tugas sekolah aktif
                  </p>
                </div>
                <label class="text-xs font-medium text-[#6b6475]">
                  Bobot (%)
                  <input
                    v-model="weightInputs[category.categoryId]"
                    type="number"
                    min="0"
                    max="100"
                    step="0.01"
                    class="mt-1 w-full rounded-2xl border border-[#e7e1ec] bg-[#fbfaf8] px-3 py-2 text-right text-sm text-[#171322] outline-none transition focus:border-[#7b61a8]"
                  />
                </label>
              </div>

              <p
                v-if="hasInvalidWeight"
                class="rounded-2xl bg-[#fff8f6] px-4 py-3 text-sm leading-6 text-[#a8665d]"
              >
                Bobot harus berada di antara 0 sampai 100.
              </p>
              <p
                v-else-if="!isWeightTotalValid"
                class="rounded-2xl bg-[#fff7ed] px-4 py-3 text-sm leading-6 text-[#b45309]"
              >
                Total bobot harus 100%.
              </p>
            </div>
          </form>
        </div>
      </section>
    </section>
  </main>
</template>
