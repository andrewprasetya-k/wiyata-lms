<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import {
  PhBookOpen,
  PhCalendarBlank,
  PhChalkboardTeacher,
  PhPlusCircle,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import { useToastStore } from "../../stores/toast";
import {
  getAcademicYearsBySchool,
  getTermsByAcademicYear,
} from "../../services/adminAcademic";
import { createAdminClass, getAdminClasses } from "../../services/adminClass";
import type { AcademicYearItem, TermItem } from "../../types/adminAcademic";
import type { AdminClassItem } from "../../types/adminClass";
import { formatDateTime } from "../../utils/date";

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
const selectedAcademicYearId = ref("");
const selectedTermId = ref("");

const yearsLoading = ref(false);
const termsLoading = ref(false);
const classesLoading = ref(false);
const yearsError = ref("");
const termsError = ref("");
const classesError = ref("");
const isCreating = ref(false);

const classForm = ref({
  classCode: "",
  classTitle: "",
  classDesc: "",
});

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
  selectedTermId.value = selectDefault ? "" : selectedTermId.value;

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

async function loadClasses() {
  classes.value = [];
  classesError.value = "";

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
  } catch {
    classesError.value = "Daftar kelas belum bisa dimuat.";
  } finally {
    classesLoading.value = false;
  }
}

async function handleAcademicYearChange() {
  await loadTerms(true);
  await loadClasses();
}

async function handleTermChange() {
  await loadClasses();
}

async function submitClass() {
  if (!currentSchool.value.schoolId || !currentSchool.value.schoolCode) {
    toast.error("Konteks sekolah aktif belum tersedia.");
    return;
  }
  if (!selectedTermId.value) {
    toast.error("Pilih semester terlebih dahulu.");
    return;
  }
  if (!classForm.value.classCode.trim() || !classForm.value.classTitle.trim()) {
    toast.error("Kode dan nama kelas wajib diisi.");
    return;
  }

  isCreating.value = true;
  try {
    await createAdminClass({
      schoolId: currentSchool.value.schoolId,
      termId: selectedTermId.value,
      classCode: classForm.value.classCode.trim(),
      classTitle: classForm.value.classTitle.trim(),
      classDesc: classForm.value.classDesc.trim(),
    });
    classForm.value = { classCode: "", classTitle: "", classDesc: "" };
    toast.success("Kelas berhasil dibuat.");
    await loadClasses();
  } catch {
    toast.error("Kelas belum bisa dibuat.");
  } finally {
    isCreating.value = false;
  }
}

onMounted(async () => {
  if (!currentSchool.value.hasContext) return;
  await loadAcademicYears();
  await loadTerms(true);
  await loadClasses();
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
            Kelas
          </h1>
          <p class="mt-2 max-w-3xl text-sm leading-6 text-[#6b7280]">
            Buat dan kelola kelas berdasarkan tahun ajaran serta semester
            sebelum melakukan penempatan kelas dan penugasan mengajar.
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

      <div class="grid min-w-0 gap-5 lg:grid-cols-[minmax(0,1fr)_340px]">
        <section
          class="order-2 min-w-0 rounded-2xl border border-[#ebe7df] bg-white lg:order-1"
        >
          <div
            class="flex flex-col gap-3 border-b border-[#ebe7df] p-5 sm:flex-row sm:items-start sm:justify-between"
          >
            <div>
              <p
                class="text-[10px] font-medium uppercase tracking-[0.08em] text-[#9ca3af]"
              >
                Daftar kelas
              </p>
              <h2 class="mt-1 text-base font-semibold text-[#171322]">
                Kelas pada semester terpilih
              </h2>
              <p class="mt-1 text-sm text-[#6b7280]">
                {{
                  selectedTerm
                    ? `${selectedTerm.termName} · ${selectedAcademicYear?.academicYearName || "Tahun ajaran"}`
                    : "Pilih periode akademik untuk menampilkan kelas."
                }}
              </p>
            </div>
            <span
              class="inline-flex shrink-0 items-center gap-2 self-start rounded-lg bg-[#eef2ff] px-3 py-2 text-xs font-medium text-[#4f46e5]"
            >
              <PhBookOpen :size="16" weight="duotone" />
              {{ classes.length }} kelas
            </span>
          </div>

          <div class="p-5">
            <div v-if="classesLoading" class="space-y-3">
              <div
                v-for="item in 3"
                :key="item"
                class="h-28 animate-pulse rounded-lg bg-[#fbfaf8]"
              />
            </div>

            <div
              v-else-if="classesError"
              class="rounded-lg border border-[#fecaca] bg-[#fef2f2] p-5 text-center"
            >
              <PhWarningCircle
                :size="26"
                class="mx-auto text-[#dc2626]"
                weight="duotone"
              />
              <h3 class="mt-3 text-sm font-semibold text-[#171322]">
                Daftar kelas belum bisa dimuat
              </h3>
              <p class="mt-2 text-sm leading-6 text-[#6b7280]">
                {{ classesError }}
              </p>
              <button
                type="button"
                class="mt-4 rounded-lg bg-[#171322] px-4 py-2 text-sm font-medium text-white transition hover:bg-[#374151]"
                @click="loadClasses"
              >
                Coba lagi
              </button>
            </div>

            <div
              v-else-if="!selectedTermId"
              class="rounded-lg bg-[#fbfaf8] px-5 py-8 text-center"
            >
              <PhCalendarBlank
                class="mx-auto h-7 w-7 text-[#9ca3af]"
                weight="duotone"
              />
              <h3 class="mt-3 text-sm font-semibold text-[#171322]">
                Semester belum dipilih
              </h3>
              <p class="mt-2 text-sm leading-6 text-[#6b7280]">
                Pilih tahun ajaran dan semester untuk melihat daftar kelas.
              </p>
            </div>

            <div
              v-else-if="classes.length === 0"
              class="rounded-lg bg-[#fbfaf8] px-5 py-8 text-center"
            >
              <PhBookOpen
                class="mx-auto h-7 w-7 text-[#9ca3af]"
                weight="duotone"
              />
              <h3 class="mt-3 text-sm font-semibold text-[#171322]">
                Belum ada kelas
              </h3>
              <p class="mt-2 text-sm leading-6 text-[#6b7280]">
                Tambahkan kelas baru untuk semester yang sedang dipilih.
              </p>
            </div>

            <div v-else class="divide-y divide-[#ebe7df]">
              <article
                v-for="classItem in classes"
                :key="classItem.classId"
                class="min-w-0 py-4 first:pt-0 last:pb-0"
              >
                <div
                  class="flex min-w-0 flex-col gap-3 sm:flex-row sm:items-start sm:justify-between"
                >
                  <div class="min-w-0">
                    <div class="flex min-w-0 flex-wrap items-center gap-2">
                      <h3
                        class="min-w-0 wrap-break-word text-sm font-semibold text-[#171322]"
                      >
                        {{ classItem.classTitle }}
                      </h3>
                      <span
                        class="rounded-lg bg-[#f3f4f6] px-2 py-1 text-[11px] font-medium text-[#6b7280]"
                      >
                        {{ classItem.classCode }}
                      </span>
                    </div>
                    <p
                      class="mt-1 wrap-break-word text-sm leading-6 text-[#6b7280]"
                    >
                      {{
                        classItem.classDesc || "Deskripsi belum ditambahkan."
                      }}
                    </p>
                  </div>
                  <span
                    class="shrink-0 self-start rounded-lg px-2.5 py-1 text-[11px] font-medium"
                    :class="
                      classItem.isActive
                        ? 'bg-[#ecfdf5] text-[#059669]'
                        : 'bg-[#f3f4f6] text-[#6b7280]'
                    "
                  >
                    {{ classItem.isActive ? "Aktif" : "Nonaktif" }}
                  </span>
                </div>

                <dl
                  class="mt-3 grid min-w-0 gap-x-5 gap-y-2 rounded-lg bg-[#fbfaf8] p-3 text-xs sm:grid-cols-2"
                >
                  <div class="min-w-0">
                    <dt class="text-[#9ca3af]">Semester</dt>
                    <dd
                      class="mt-0.5 wrap-break-word font-medium text-[#374151]"
                    >
                      {{ classItem.termName || selectedTerm?.termName || "-" }}
                    </dd>
                  </div>
                  <div class="min-w-0">
                    <dt class="text-[#9ca3af]">Tahun ajaran</dt>
                    <dd
                      class="mt-0.5 wrap-break-word font-medium text-[#374151]"
                    >
                      {{
                        classItem.academicYearName ||
                        selectedAcademicYear?.academicYearName ||
                        "-"
                      }}
                    </dd>
                  </div>
                  <div class="min-w-0">
                    <dt class="text-[#9ca3af]">Dibuat</dt>
                    <dd
                      class="mt-0.5 wrap-break-word font-medium text-[#374151]"
                    >
                      {{ formatDateTime(classItem.createdAt) }}
                    </dd>
                  </div>
                  <div class="min-w-0">
                    <dt class="text-[#9ca3af]">Dibuat oleh</dt>
                    <dd
                      class="mt-0.5 wrap-break-word font-medium text-[#374151]"
                    >
                      {{ classItem.creatorName || "Tidak tersedia" }}
                    </dd>
                  </div>
                </dl>
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
                    Periode akademik
                  </p>
                  <h2 class="mt-1 text-base font-semibold text-[#171322]">
                    Pilih periode
                  </h2>
                </div>
                <PhCalendarBlank
                  :size="21"
                  class="text-[#4f46e5]"
                  weight="duotone"
                />
              </div>

              <div class="mt-5 space-y-4">
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
              </div>

              <div class="mt-4 space-y-2 text-xs leading-5">
                <p v-if="yearsLoading" class="text-[#6b7280]">
                  Memuat tahun ajaran...
                </p>
                <p v-else-if="yearsError" class="text-[#dc2626]">
                  {{ yearsError }}
                </p>
                <p
                  v-else-if="academicYears.length === 0"
                  class="text-[#6b7280]"
                >
                  Belum ada tahun ajaran. Buat melalui Struktur Akademik.
                </p>
                <p v-if="termsLoading" class="text-[#6b7280]">
                  Memuat semester...
                </p>
                <p v-else-if="termsError" class="text-[#dc2626]">
                  {{ termsError }}
                </p>
                <p
                  v-else-if="selectedAcademicYearId && terms.length === 0"
                  class="text-[#6b7280]"
                >
                  Belum ada semester untuk tahun ajaran ini.
                </p>
              </div>

              <div class="mt-4 rounded-lg bg-[#fbfaf8] p-3">
                <p
                  class="text-[10px] font-medium uppercase tracking-[0.08em] text-[#9ca3af]"
                >
                  Konteks aktif
                </p>
                <p class="mt-2 text-xs leading-5 text-[#6b7280]">
                  {{
                    selectedAcademicYear?.academicYearName ||
                    "Tahun belum dipilih"
                  }}
                  ·
                  {{ selectedTerm?.termName || "Semester belum dipilih" }}
                </p>
              </div>
            </section>

            <section class="rounded-2xl border border-[#ebe7df] bg-white p-5">
              <div class="flex items-start justify-between gap-3">
                <div>
                  <p
                    class="text-[10px] font-medium uppercase tracking-[0.08em] text-[#9ca3af]"
                  >
                    Tambah kelas
                  </p>
                  <h2 class="mt-1 text-base font-semibold text-[#171322]">
                    Kelas baru
                  </h2>
                </div>
                <PhPlusCircle
                  :size="21"
                  class="text-[#059669]"
                  weight="duotone"
                />
              </div>

              <form class="mt-5 space-y-3" @submit.prevent="submitClass">
                <label class="block text-xs font-medium text-[#6b7280]">
                  Kode kelas
                  <input
                    v-model="classForm.classCode"
                    type="text"
                    placeholder="Contoh: X-IPA-1"
                    class="mt-2 w-full rounded-lg border border-[#ebe7df] bg-[#fbfaf8] px-3.5 py-2.5 text-sm text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#4f46e5] focus:bg-white"
                  />
                </label>
                <label class="block text-xs font-medium text-[#6b7280]">
                  Nama kelas
                  <input
                    v-model="classForm.classTitle"
                    type="text"
                    placeholder="Nama kelas"
                    class="mt-2 w-full rounded-lg border border-[#ebe7df] bg-[#fbfaf8] px-3.5 py-2.5 text-sm text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#4f46e5] focus:bg-white"
                  />
                </label>
                <label class="block text-xs font-medium text-[#6b7280]">
                  Deskripsi
                  <textarea
                    v-model="classForm.classDesc"
                    rows="3"
                    placeholder="Deskripsi singkat, opsional"
                    class="mt-2 w-full resize-none rounded-lg border border-[#ebe7df] bg-[#fbfaf8] px-3.5 py-2.5 text-sm leading-6 text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#4f46e5] focus:bg-white"
                  />
                </label>
                <button
                  type="submit"
                  class="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-[#ea580c] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#c2410c] disabled:cursor-not-allowed disabled:opacity-60"
                  :disabled="
                    isCreating || !currentSchool.hasContext || !selectedTermId
                  "
                >
                  <PhChalkboardTeacher :size="17" weight="duotone" />
                  {{ isCreating ? "Membuat..." : "Buat kelas" }}
                </button>
              </form>
            </section>
          </div>
        </aside>
      </div>
    </section>
  </main>
</template>
