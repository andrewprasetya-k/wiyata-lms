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
  updateSubject,
} from "../../services/adminAcademic";
import type {
  AcademicYearItem,
  AssessmentWeightItem,
  AssignmentCategoryItem,
  SubjectItem,
  TermItem,
} from "../../types/adminAcademic";
import { formatDateTime } from "../../utils/date";
import { getSubjectColor } from "../../utils/color";
import { getApiError } from "../../utils/error";
import {
  PhArrowRight,
  PhBookOpen,
  PhCalendarBlank,
  PhChartBar,
  PhChecks,
  PhPencilSimple,
  PhPlusCircle,
  PhTag,
  PhX,
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
type WeightInputValue = string | number | null | undefined;
const weightInputs = ref<Record<string, string | number>>({});

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
const activeTab = ref<'periode' | 'mapel'>('periode');
const mapelLoaded = ref(false);

const academicYearForm = ref({ academicYearName: "" });
const termForm = ref({ termName: "" });
const subjectForm = ref({ subjectName: "", subjectCode: "", color: "" });
const editingSubjectId = ref("");
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
    (total, category) =>
      total + parseWeightValue(weightInputs.value[category.categoryId]),
    0,
  ),
);

const hasInvalidWeight = computed(() =>
  categories.value.some((category) => {
    return isWeightInputInvalid(weightInputs.value[category.categoryId]);
  }),
);

const isWeightTotalValid = computed(
  () => Math.abs(totalWeight.value - 100) <= 0.01,
);

const subjectColorPreview = computed(() => {
  const color = normalizeSubjectColor(subjectForm.value.color);
  if (color && isValidSubjectColor(color)) return color;
  const seed =
    subjectForm.value.subjectName ||
    subjectForm.value.subjectCode ||
    editingSubjectId.value;
  return getSubjectColor(seed);
});

const subjectColorPickerValue = computed({
  get: () => toColorPickerValue(subjectColorPreview.value),
  set: (value: string) => {
    subjectForm.value.color = value;
  },
});

const canSubmitWeights = computed(
  () =>
    currentSchool.value.hasContext &&
    Boolean(selectedWeightSubjectId.value) &&
    categories.value.length > 0 &&
    activeAction.value !== "weights-save",
);

function normalizeSubjectColor(color?: string | null) {
  return color?.trim() ?? "";
}

function isValidSubjectColor(color: string) {
  return /^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$/.test(color);
}

function toColorPickerValue(color: string) {
  const normalized = normalizeSubjectColor(color);
  if (/^#[0-9a-fA-F]{6}$/.test(normalized)) return normalized;
  if (/^#[0-9a-fA-F]{3}$/.test(normalized)) {
    const [, r, g, b] = normalized;
    return `#${r}${r}${g}${g}${b}${b}`;
  }
  if (/^#[0-9a-fA-F]{8}$/.test(normalized)) {
    return normalized.slice(0, 7);
  }
  return "#4f46e5";
}

function subjectDisplayColor(subject: SubjectItem) {
  return subject.color || getSubjectColor(subject.subjectId || subject.subjectName);
}

function normalizeWeightInput(value: WeightInputValue) {
  if (value === undefined || value === null) {
    return { value: 0, valid: true };
  }

  if (typeof value === "number") {
    return {
      value: Number.isFinite(value) ? value : 0,
      valid: Number.isFinite(value),
    };
  }

  const normalized = value.trim().replace(",", ".");
  if (normalized === "") {
    return { value: 0, valid: true };
  }

  const numericValue = Number(normalized);
  return {
    value: Number.isFinite(numericValue) ? numericValue : 0,
    valid: Number.isFinite(numericValue),
  };
}

function parseWeightValue(value: WeightInputValue) {
  return normalizeWeightInput(value).value;
}

function isWeightInputInvalid(value: WeightInputValue) {
  const parsed = normalizeWeightInput(value);
  return !parsed.valid || parsed.value < 0 || parsed.value > 100;
}

function formatWeight(value: number) {
  return new Intl.NumberFormat("id-ID", {
    maximumFractionDigits: 2,
  }).format(value);
}


function getApiStatus(error: unknown) {
  if (typeof error === "object" && error !== null && "response" in error) {
    return (error as { response?: { status?: number } }).response?.status;
  }

  return undefined;
}

function resetWeightInputs(weights: AssessmentWeightItem[] = []) {
  const nextInputs: Record<string, string | number> = {};
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

async function switchTab(tab: 'periode' | 'mapel') {
  activeTab.value = tab;
  if (tab === 'mapel' && !mapelLoaded.value) {
    mapelLoaded.value = true;
    await Promise.all([loadSubjects(), loadCategories()]);
    await loadAssessmentWeights();
  }
}

async function submitAcademicYear() {
  if (!currentSchool.value.schoolId) {
    toast.error("Konteks sekolah aktif belum tersedia.");
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
    toast.error("Konteks sekolah aktif belum tersedia.");
    return;
  }
  if (
    !subjectForm.value.subjectName.trim() ||
    !subjectForm.value.subjectCode.trim()
  ) {
    toast.error("Nama dan kode mata pelajaran wajib diisi.");
    return;
  }
  const color = normalizeSubjectColor(subjectForm.value.color);
  if (color && !isValidSubjectColor(color)) {
    toast.error("Warna mata pelajaran harus berupa hex, contoh #4f46e5.");
    return;
  }

  activeAction.value = editingSubjectId.value
    ? `subject-update-${editingSubjectId.value}`
    : "subject-create";

  try {
    const payload = {
      subjectName: subjectForm.value.subjectName.trim(),
      subjectCode: subjectForm.value.subjectCode.trim(),
      color,
    };

    if (editingSubjectId.value) {
      await updateSubject(editingSubjectId.value, payload);
      toast.success("Mata pelajaran berhasil diperbarui.");
    } else {
      await createSubject({
        schoolId: currentSchool.value.schoolId,
        ...payload,
      });
      toast.success("Mata pelajaran berhasil dibuat.");
    }
    resetSubjectForm();
    await loadSubjects();
  } catch (error) {
    toast.error(getApiError(error));
  } finally {
    activeAction.value = "";
  }
}

function editSubject(subject: SubjectItem) {
  editingSubjectId.value = subject.subjectId;
  subjectForm.value = {
    subjectName: subject.subjectName,
    subjectCode: subject.subjectCode,
    color: subject.color ?? "",
  };
}

function resetSubjectForm() {
  editingSubjectId.value = "";
  subjectForm.value = { subjectName: "", subjectCode: "", color: "" };
}

async function submitCategory() {
  if (!currentSchool.value.schoolId) {
    toast.error("Konteks sekolah aktif belum tersedia.");
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
    const response = await getAssessmentWeightsBySubject(
      selectedWeightSubjectId.value,
    );
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
      weightsError.value = getApiError(error);
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
    toast.error(getApiError(error));
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
  await loadAcademicYears();
  await loadTerms();
});

watch(selectedWeightSubjectId, () => {
  loadAssessmentWeights();
});
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-[#f8f7f4]">
    <header class="border-b border-border bg-white">
      <div
        class="flex min-w-0 flex-col gap-3 px-5 py-5 sm:px-6 lg:flex-row lg:items-end lg:justify-between lg:px-8"
      >
        <div class="min-w-0">
          <h1 class="mt-2 text-2xl font-semibold text-foreground sm:text-3xl">
            Struktur Akademik
          </h1>
          <p class="mt-2 max-w-3xl text-sm leading-6 text-muted">
            Kelola tahun ajaran, semester, mata pelajaran, kategori tugas, dan
            bobot penilaian sebagai dasar operasional akademik sekolah.
          </p>
        </div>

        <div class="flex min-w-0 flex-wrap gap-2 text-xs">
          <span
            class="max-w-full truncate rounded-lg bg-[#fff4ee] px-3 py-2 font-medium text-[#ea580c]"
          >
            {{ currentSchool.schoolName || "Sekolah belum tersedia" }}
          </span>
          <span
            class="rounded-lg bg-[#f3f1ec] px-3 py-2 font-medium text-muted"
          >
            {{ currentSchool.schoolCode || "Kode belum tersedia" }}
          </span>
        </div>
      </div>
    </header>

    <section
      class="flex w-full max-w-none flex-col gap-5 px-5 py-5 sm:px-6 lg:px-8 lg:py-6"
    >
      <div
        v-if="!currentSchool.hasContext"
        class="rounded-xl border border-[#fecaca] bg-[#fef2f2] px-4 py-3 text-sm text-[#a8665d]"
      >
        Konteks sekolah aktif belum ditemukan. Pastikan akun admin sudah punya
        akses sekolah yang valid.
      </div>

      <div class="flex gap-1 self-start rounded-xl border border-border bg-[#f3f1ec] p-1">
        <button
          type="button"
          class="rounded-lg px-4 py-2 text-sm font-medium transition"
          :class="activeTab === 'periode' ? 'bg-white text-foreground shadow-sm' : 'text-muted hover:text-[#374151]'"
          @click="switchTab('periode')"
        >
          Periode Akademik
        </button>
        <button
          type="button"
          class="rounded-lg px-4 py-2 text-sm font-medium transition"
          :class="activeTab === 'mapel' ? 'bg-white text-foreground shadow-sm' : 'text-muted hover:text-[#374151]'"
          @click="switchTab('mapel')"
        >
          Mata Pelajaran
        </button>
      </div>

      <template v-if="activeTab === 'periode'">
      <section class="grid gap-5 lg:grid-cols-2">
        <article
          class="rounded-xl border border-border bg-white shadow-sm p-5"
        >
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <p
                class="eyebrow"
              >
                Tahun ajaran
              </p>
              <h2 class="mt-2 text-xl font-semibold text-foreground">
                Tahun ajaran
              </h2>
              <p class="mt-2 text-sm leading-6 text-muted">
                Buka atau tutup periode akademik yang digunakan oleh semester
                dan kelas.
              </p>
            </div>
            <span
              class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-[#fff4ee] text-[#ea580c]"
            >
              <PhCalendarBlank :size="22" weight="duotone" />
            </span>
          </div>

          <form
            class="mt-5 flex flex-col gap-3 sm:flex-row"
            @submit.prevent="submitAcademicYear"
          >
            <input
              v-model="academicYearForm.academicYearName"
              type="text"
              placeholder="Contoh: 2026/2027"
              class="min-w-0 flex-1 rounded-lg border border-[#e5e7eb] bg-white px-4 py-3 text-sm text-foreground outline-none transition placeholder:text-[#9ca3af] focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
            />
            <button
              type="submit"
              class="inline-flex items-center justify-center gap-2 rounded-lg bg-[#ea580c] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#c2410c] disabled:cursor-not-allowed disabled:opacity-60"
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
            <div v-if="academicYearsLoading" class="space-y-3">
              <div v-for="item in 2" :key="item" class="h-14 animate-pulse rounded-lg bg-[#fbfaf8]" />
            </div>
            <p
              v-else-if="academicYearsError"
              class="rounded-lg border border-[#fecaca] bg-[#fef2f2] px-4 py-3 text-sm text-[#a8665d]"
            >
              {{ academicYearsError }}
            </p>
            <div
              v-else-if="academicYears.length === 0"
              class="rounded-lg border border-dashed border-[#d1d5db] bg-[#fafafa] px-4 py-8 text-center"
            >
              <PhCalendarBlank class="mx-auto h-7 w-7 text-[#9ca3af]" weight="duotone" />
              <p class="mt-3 text-sm font-semibold text-foreground">Belum ada tahun ajaran</p>
              <p class="mt-1 text-sm text-muted">Buat tahun ajaran pertama menggunakan form di atas.</p>
            </div>

            <article
              v-for="year in academicYears"
              :key="year.academicYearId"
              class="rounded-lg bg-[#fbfaf8] p-4"
            >
              <div
                class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between"
              >
                <div class="min-w-0">
                  <div class="flex flex-wrap items-center gap-2">
                    <h3 class="truncate text-base font-semibold text-foreground">
                      {{ year.academicYearName }}
                    </h3>
                    <span
                      class="rounded-full px-2.5 py-1 text-xs font-semibold"
                      :class="
                        year.isActive
                          ? 'bg-[#ecfdf3] text-[#027a48]'
                          : 'bg-[#f3f1ec] text-muted'
                      "
                    >
                      {{ year.isActive ? "Aktif" : "Nonaktif" }}
                    </span>
                  </div>
                  <p class="mt-2 text-sm text-muted">
                    {{ year.schoolCode || currentSchool.schoolCode }} • dibuat
                    {{ formatDateTime(year.createdAt) }}
                  </p>
                </div>

                <button
                  type="button"
                  class="inline-flex items-center justify-center gap-2 rounded-lg border border-border bg-white px-4 py-2.5 text-sm font-medium text-[#374151] transition hover:border-brand hover:text-brand disabled:cursor-not-allowed disabled:opacity-60"
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

        <article
          class="rounded-xl border border-border bg-white shadow-sm p-5"
        >
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <p
                class="eyebrow"
              >
                Semester
              </p>
              <h2 class="mt-2 text-xl font-semibold text-foreground">
                Semester
              </h2>
              <p class="mt-2 text-sm leading-6 text-muted">
                Pilih tahun ajaran, lalu buat atau aktifkan semester.
              </p>
            </div>
            <span
              class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-[#ecfdf3] text-[#027a48]"
            >
              <PhCalendarBlank :size="22" weight="duotone" />
            </span>
          </div>

          <label class="mt-5 block text-sm font-medium text-[#3f3a4a]">
            Tahun ajaran
            <select
              v-model="selectedAcademicYearId"
              class="mt-2 w-full rounded-lg border border-[#e5e7eb] bg-white px-4 py-3 text-sm text-foreground outline-none transition focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
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
              class="min-w-0 flex-1 rounded-lg border border-[#e5e7eb] bg-white px-4 py-3 text-sm text-foreground outline-none transition placeholder:text-[#9ca3af] focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
            />
            <button
              type="submit"
              class="inline-flex items-center justify-center gap-2 rounded-lg bg-[#ea580c] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#c2410c] disabled:cursor-not-allowed disabled:opacity-60"
              :disabled="
                activeAction === 'term-create' || !selectedAcademicYearId
              "
            >
              <PhPlusCircle :size="18" weight="duotone" />
              Tambah
            </button>
          </form>

          <div class="mt-5 space-y-3">
            <div v-if="termsLoading" class="space-y-3">
              <div v-for="item in 2" :key="item" class="h-14 animate-pulse rounded-lg bg-[#fbfaf8]" />
            </div>
            <p
              v-else-if="termsError"
              class="rounded-lg border border-[#fecaca] bg-[#fef2f2] px-4 py-3 text-sm text-[#a8665d]"
            >
              {{ termsError }}
            </p>
            <div
              v-else-if="!selectedAcademicYearId"
              class="rounded-lg border border-dashed border-[#d1d5db] bg-[#fafafa] px-4 py-8 text-center"
            >
              <PhCalendarBlank class="mx-auto h-7 w-7 text-[#9ca3af]" weight="duotone" />
              <p class="mt-3 text-sm font-semibold text-foreground">Pilih tahun ajaran</p>
              <p class="mt-1 text-sm text-muted">Pilih tahun ajaran di atas untuk melihat dan mengelola semester.</p>
            </div>
            <div
              v-else-if="terms.length === 0"
              class="rounded-lg border border-dashed border-[#d1d5db] bg-[#fafafa] px-4 py-8 text-center"
            >
              <PhCalendarBlank class="mx-auto h-7 w-7 text-[#9ca3af]" weight="duotone" />
              <p class="mt-3 text-sm font-semibold text-foreground">Belum ada semester</p>
              <p class="mt-1 text-sm text-muted">Buat semester pertama untuk tahun ajaran ini menggunakan form di atas.</p>
            </div>

            <article
              v-for="term in terms"
              :key="term.termId"
              class="rounded-lg bg-[#fbfaf8] p-4"
            >
              <div
                class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between"
              >
                <div class="min-w-0">
                  <div class="flex flex-wrap items-center gap-2">
                    <h3 class="truncate text-base font-semibold text-foreground">
                      {{ term.termName }}
                    </h3>
                    <span
                      class="rounded-full px-2.5 py-1 text-xs font-semibold"
                      :class="
                        term.isActive
                          ? 'bg-[#ecfdf3] text-[#027a48]'
                          : 'bg-[#f3f1ec] text-muted'
                      "
                    >
                      {{ term.isActive ? "Aktif" : "Nonaktif" }}
                    </span>
                  </div>
                  <p class="mt-2 text-sm text-muted">
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
                  class="inline-flex items-center justify-center gap-2 rounded-lg border border-border bg-white px-4 py-2.5 text-sm font-medium text-[#374151] transition hover:border-brand hover:text-brand disabled:cursor-not-allowed disabled:opacity-60"
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

      <RouterLink
        v-if="terms.length > 0"
        to="/admin/classes"
        class="flex items-center justify-between gap-4 rounded-xl border border-border bg-white shadow-sm p-5 transition hover:border-brand hover:shadow-sm"
      >
        <div>
          <p class="eyebrow">
            Langkah berikutnya
          </p>
          <p class="mt-1 text-base font-semibold text-foreground">Buat Kelas</p>
          <p class="mt-1 text-sm text-muted">
            Semester sudah siap — buat kelas untuk tiap tingkat dan semester.
          </p>
        </div>
        <PhArrowRight :size="20" class="shrink-0 text-brand" weight="bold" />
      </RouterLink>
      </template>

      <template v-if="activeTab === 'mapel'">
      <section class="grid gap-5 lg:grid-cols-2">
        <article
          class="rounded-xl border border-border bg-white shadow-sm p-5"
        >
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <p
                class="eyebrow"
              >
                Mata pelajaran
              </p>
              <h2 class="mt-2 text-xl font-semibold text-foreground">
                Mata pelajaran
              </h2>
              <p class="mt-2 text-sm leading-6 text-muted">
                Daftarkan mata pelajaran yang akan dipakai pada penugasan
                mengajar dan konten belajar.
              </p>
            </div>
            <span
              class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-[#eff6ff] text-[#2563eb]"
            >
              <PhBookOpen :size="22" weight="duotone" />
            </span>
          </div>

          <form
            class="mt-5 grid gap-3 sm:grid-cols-2"
            @submit.prevent="submitSubject"
          >
            <input
              v-model="subjectForm.subjectName"
              type="text"
              placeholder="Nama mata pelajaran"
              class="min-w-0 rounded-lg border border-[#e5e7eb] bg-white px-4 py-3 text-sm text-foreground outline-none transition placeholder:text-[#9ca3af] focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
            />
            <input
              v-model="subjectForm.subjectCode"
              type="text"
              placeholder="Kode"
              class="min-w-0 rounded-lg border border-[#e5e7eb] bg-white px-4 py-3 text-sm text-foreground outline-none transition placeholder:text-[#9ca3af] focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
            />
            <div class="sm:col-span-2">
              <label class="text-xs font-semibold text-muted">
                Warna mata pelajaran
              </label>
              <div class="mt-2 grid gap-3 sm:grid-cols-[auto_minmax(0,1fr)]">
                <input
                  v-model="subjectColorPickerValue"
                  type="color"
                  class="h-11 w-14 rounded-lg border border-[#e5e7eb] bg-white p-1"
                  aria-label="Pilih warna mata pelajaran"
                />
                <div class="flex min-w-0 items-center gap-3">
                  <span
                    class="h-8 w-8 shrink-0 rounded-full border border-border"
                    :style="{ backgroundColor: subjectColorPreview }"
                    aria-hidden="true"
                  />
                  <input
                    v-model="subjectForm.color"
                    type="text"
                    placeholder="#4f46e5"
                    class="min-w-0 flex-1 rounded-lg border border-[#e5e7eb] bg-white px-4 py-3 text-sm text-foreground outline-none transition placeholder:text-[#9ca3af] focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
                  />
                </div>
              </div>
              <p class="mt-2 text-xs leading-5 text-[#8a8494]">
                Opsional. Kosongkan untuk memakai warna fallback otomatis.
              </p>
            </div>
            <button
              type="submit"
              class="inline-flex items-center justify-center gap-2 rounded-lg bg-[#ea580c] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#c2410c] disabled:cursor-not-allowed disabled:opacity-60 sm:col-span-2"
              :disabled="
                activeAction === 'subject-create' ||
                activeAction === `subject-update-${editingSubjectId}` ||
                !currentSchool.hasContext
              "
            >
              <PhPlusCircle :size="18" weight="duotone" />
              {{
                editingSubjectId
                  ? "Simpan mata pelajaran"
                  : "Tambah mata pelajaran"
              }}
            </button>
            <button
              v-if="editingSubjectId"
              type="button"
              class="inline-flex items-center justify-center gap-2 rounded-lg border border-border bg-white px-4 py-2.5 text-sm font-medium text-[#374151] transition hover:border-brand hover:text-brand disabled:cursor-not-allowed disabled:opacity-60 sm:col-span-2"
              @click="resetSubjectForm"
            >
              <PhX :size="18" weight="duotone" />
              Batalkan edit
            </button>
          </form>

          <div class="mt-5 space-y-3">
            <div v-if="subjectsLoading" class="space-y-3">
              <div v-for="item in 2" :key="item" class="h-14 animate-pulse rounded-lg bg-[#fbfaf8]" />
            </div>
            <p
              v-else-if="subjectsError"
              class="rounded-lg border border-[#fecaca] bg-[#fef2f2] px-4 py-3 text-sm text-[#a8665d]"
            >
              {{ subjectsError }}
            </p>
            <div
              v-else-if="subjects.length === 0"
              class="rounded-lg border border-dashed border-[#d1d5db] bg-[#fafafa] px-4 py-8 text-center"
            >
              <PhBookOpen class="mx-auto h-7 w-7 text-[#9ca3af]" weight="duotone" />
              <p class="mt-3 text-sm font-semibold text-foreground">Belum ada mata pelajaran</p>
              <p class="mt-1 text-sm text-muted">Tambah mata pelajaran pertama menggunakan form di atas.</p>
            </div>

            <article
              v-for="subject in subjects"
              :key="subject.subjectId"
              class="rounded-lg bg-[#fbfaf8] p-4"
            >
              <div class="flex min-w-0 items-start justify-between gap-3">
                <div class="flex min-w-0 items-start gap-3">
                  <span
                    class="mt-1 h-3 w-3 shrink-0 rounded-full"
                    :style="{ backgroundColor: subjectDisplayColor(subject) }"
                    aria-hidden="true"
                  />
                  <div class="min-w-0">
                    <h3 class="truncate text-base font-semibold text-foreground">
                      {{ subject.subjectName }}
                    </h3>
                    <p class="mt-2 text-sm text-muted">
                      {{ subject.subjectCode }} •
                      {{ subject.schoolCode || currentSchool.schoolCode }} •
                      dibuat {{ formatDateTime(subject.createdAt) }}
                    </p>
                    <p class="mt-1 text-xs text-[#8a8494]">
                      Warna:
                      <span class="font-medium text-[#4a4356]">
                        {{ subject.color || "fallback otomatis" }}
                      </span>
                    </p>
                  </div>
                </div>
                <button
                  type="button"
                  class="inline-flex shrink-0 items-center justify-center gap-2 rounded-lg border border-border bg-white px-3 py-2 text-xs font-medium text-[#374151] transition hover:border-brand hover:text-brand disabled:cursor-not-allowed disabled:opacity-60"
                  @click="editSubject(subject)"
                >
                  <PhPencilSimple :size="16" weight="duotone" />
                  Edit
                </button>
              </div>
            </article>
          </div>
        </article>

        <article
          class="rounded-xl border border-border bg-white shadow-sm p-5"
        >
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <p
                class="eyebrow"
              >
                Kategori tugas
              </p>
              <h2 class="mt-2 text-xl font-semibold text-foreground">
                Kategori tugas
              </h2>
              <p class="mt-2 text-sm leading-6 text-muted">
                Siapkan kategori yang digunakan saat guru membuat tugas dan
                admin mengatur bobot penilaian.
              </p>
            </div>
            <span
              class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-[#fef2f2] text-[#dc2626]"
            >
              <PhTag :size="22" weight="duotone" />
            </span>
          </div>

          <form
            class="mt-5 flex flex-col gap-3 sm:flex-row"
            @submit.prevent="submitCategory"
          >
            <input
              v-model="categoryForm.categoryName"
              type="text"
              placeholder="Contoh: Kuis"
              class="min-w-0 flex-1 rounded-lg border border-[#e5e7eb] bg-white px-4 py-3 text-sm text-foreground outline-none transition placeholder:text-[#9ca3af] focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
            />
            <button
              type="submit"
              class="inline-flex items-center justify-center gap-2 rounded-lg bg-[#ea580c] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#c2410c] disabled:cursor-not-allowed disabled:opacity-60"
              :disabled="
                activeAction === 'category-create' || !currentSchool.hasContext
              "
            >
              <PhPlusCircle :size="18" weight="duotone" />
              Tambah
            </button>
          </form>

          <div class="mt-5 space-y-3">
            <div v-if="categoriesLoading" class="space-y-3">
              <div v-for="item in 2" :key="item" class="h-14 animate-pulse rounded-lg bg-[#fbfaf8]" />
            </div>
            <p
              v-else-if="categoriesError"
              class="rounded-lg border border-[#fecaca] bg-[#fef2f2] px-4 py-3 text-sm text-[#a8665d]"
            >
              {{ categoriesError }}
            </p>
            <div
              v-else-if="categories.length === 0"
              class="rounded-lg border border-dashed border-[#d1d5db] bg-[#fafafa] px-4 py-8 text-center"
            >
              <PhTag class="mx-auto h-7 w-7 text-[#9ca3af]" weight="duotone" />
              <p class="mt-3 text-sm font-semibold text-foreground">Belum ada kategori tugas</p>
              <p class="mt-1 text-sm text-muted">Tambah kategori seperti "Kuis" atau "UTS" untuk dipakai di bobot penilaian.</p>
            </div>

            <article
              v-for="category in categories"
              :key="category.categoryId"
              class="rounded-lg bg-[#fbfaf8] p-4"
            >
              <h3 class="truncate text-base font-semibold text-foreground">
                {{ category.categoryName }}
              </h3>
              <p class="mt-2 text-sm text-muted">
                Dibuat {{ formatDateTime(category.createdAt) }}
              </p>
            </article>
          </div>
        </article>
      </section>

      <section
        class="rounded-xl border border-border bg-white shadow-sm p-5"
      >
        <div
          class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between"
        >
          <div class="max-w-3xl">
            <p
              class="eyebrow"
            >
              Bobot penilaian
            </p>
            <h2 class="mt-2 text-xl font-semibold text-foreground">
              Bobot penilaian
            </h2>
            <p class="mt-2 text-sm leading-6 text-muted">
              Bobot berlaku per mata pelajaran dan digunakan untuk menghitung
              rata-rata berbobot sementara. Ini bukan nilai rapor final resmi.
            </p>
          </div>
          <div
            class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-[#fff4ee] text-[#ea580c]"
          >
            <PhChartBar :size="22" weight="duotone" />
          </div>
        </div>

        <div
          v-if="subjects.length === 0"
          class="mt-5 rounded-lg border border-dashed border-[#d1d5db] bg-[#fafafa] p-4 text-sm leading-6 text-muted"
        >
          Tambahkan mata pelajaran terlebih dahulu sebelum mengatur bobot nilai.
        </div>

        <div
          v-else-if="categories.length === 0"
          class="mt-5 rounded-lg border border-dashed border-[#d1d5db] bg-[#fafafa] p-4 text-sm leading-6 text-muted"
        >
          Tambahkan kategori tugas terlebih dahulu sebelum mengatur bobot nilai.
        </div>

        <div
          v-else
          class="mt-5 grid gap-5 xl:grid-cols-[minmax(0,0.85fr)_minmax(0,1.15fr)]"
        >
          <div class="rounded-lg bg-[#fbfaf8] p-4">
            <label class="block text-sm font-medium text-[#3f3a4a]">
              Mata pelajaran
              <select
                v-model="selectedWeightSubjectId"
                class="mt-2 w-full rounded-lg border border-[#e5e7eb] bg-white px-4 py-3 text-sm text-foreground outline-none transition focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
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

            <div class="mt-4 rounded-lg border border-border bg-white p-4">
              <p class="text-xs font-medium text-muted">Total bobot</p>
              <div class="mt-2 flex flex-wrap items-end justify-between gap-3">
                <p
                  class="text-3xl font-semibold"
                  :class="
                    isWeightTotalValid ? 'text-[#027a48]' : 'text-[#b45309]'
                  "
                >
                  {{ formatWeight(totalWeight) }}%
                </p>
                <span
                  class="rounded-full px-3 py-1 text-xs font-semibold"
                  :class="
                    isWeightTotalValid
                      ? 'bg-[#ecfdf3] text-[#027a48]'
                      : 'bg-[#fff7ed] text-[#b45309]'
                  "
                >
                  {{ isWeightTotalValid ? "Valid" : "Harus 100%" }}
                </span>
              </div>
              <p class="mt-3 text-xs leading-5 text-muted">
                Total bobot harus 100% sebelum disimpan.
              </p>
            </div>

            <p
              v-if="selectedWeightSubject"
              class="mt-4 rounded-lg border border-[#fed7aa] bg-[#fff7ed] px-4 py-3 text-xs leading-5 text-[#9a3412]"
            >
              Bobot yang disimpan akan berlaku untuk semua kelas pada mata
              pelajaran {{ selectedWeightSubject.subjectName }}.
            </p>
          </div>

          <form
            class="rounded-lg bg-[#fbfaf8] p-4"
            @submit.prevent="submitAssessmentWeights"
          >
            <div
              class="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between"
            >
              <div>
                <p class="text-sm font-semibold text-foreground">
                  Kategori dan bobot
                </p>
                <p class="mt-1 text-xs leading-5 text-muted">
                  Kosong dianggap 0. Setiap kategori hanya muncul satu kali.
                </p>
              </div>
              <button
                class="inline-flex items-center justify-center gap-2 rounded-lg bg-[#ea580c] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#c2410c] disabled:cursor-not-allowed disabled:opacity-60"
                type="submit"
                :disabled="!canSubmitWeights"
              >
                <PhChecks :size="16" weight="duotone" />
                {{
                  activeAction === "weights-save"
                    ? "Menyimpan..."
                    : "Simpan bobot"
                }}
              </button>
            </div>

            <div v-if="weightsLoading" class="mt-4 space-y-3">
              <div
                v-for="item in 3"
                :key="item"
                class="h-14 animate-pulse rounded-lg bg-white"
              />
            </div>

            <div v-else class="mt-4 space-y-3">
              <p
                v-if="weightsError"
                class="rounded-lg border border-[#fecaca] bg-[#fef2f2] px-4 py-3 text-sm leading-6 text-[#a8665d]"
              >
                {{ weightsError }}
              </p>
              <p
                v-else-if="weightsInfoMessage"
                class="rounded-lg border border-[#fed7aa] bg-[#fff7ed] px-4 py-3 text-sm leading-6 text-[#9a3412]"
              >
                {{ weightsInfoMessage }}
              </p>

              <div
                v-for="category in categories"
                :key="category.categoryId"
                class="grid gap-3 rounded-lg border border-border bg-white p-4 sm:grid-cols-[minmax(0,1fr)_140px]"
              >
                <div class="min-w-0">
                  <p class="truncate text-sm font-semibold text-foreground">
                    {{ category.categoryName }}
                  </p>
                  <p class="mt-1 text-xs text-muted">
                    Kategori tugas sekolah aktif
                  </p>
                </div>
                <label class="text-xs font-medium text-muted">
                  Bobot (%)
                  <input
                    v-model="weightInputs[category.categoryId]"
                    type="number"
                    min="0"
                    max="100"
                    step="0.01"
                    class="mt-1 w-full rounded-lg border border-[#e5e7eb] bg-[#fafafa] px-3 py-2 text-right text-sm text-foreground outline-none transition focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
                  />
                </label>
              </div>

              <p
                v-if="hasInvalidWeight"
                class="rounded-lg border border-[#fecaca] bg-[#fef2f2] px-4 py-3 text-sm leading-6 text-[#a8665d]"
              >
                Bobot harus berada di antara 0 sampai 100.
              </p>
              <p
                v-else-if="!isWeightTotalValid"
                class="rounded-lg border border-[#fed7aa] bg-[#fff7ed] px-4 py-3 text-sm leading-6 text-[#b45309]"
              >
                Total bobot harus 100%.
              </p>
            </div>
          </form>
        </div>
      </section>
      </template>
    </section>
  </main>
</template>
