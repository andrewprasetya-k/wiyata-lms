<script setup lang="ts">
import { onMounted, ref } from "vue";
import {
  PhArrowClockwise,
  PhBuildings,
  PhCheckCircle,
  PhEnvelopeSimple,
  PhGlobe,
  PhIdentificationBadge,
  PhMagnifyingGlass,
  PhMapPin,
  PhPhone,
  PhPlusCircle,
} from "@phosphor-icons/vue";
import { useToastStore } from "../../stores/toast";
import { getApiError } from "../../utils/error";
import PaginationBar from "../../components/common/PaginationBar.vue";
import {
  bootstrapSuperAdminSchool,
  getSuperAdminSchools,
  getSuperAdminSchoolSummary,
} from "../../services/superAdminSchool";
import type {
  SuperAdminSchoolBootstrapPayload,
  SuperAdminSchoolBootstrapResponse,
  SuperAdminSchoolItem,
  SuperAdminSchoolSummary,
} from "../../types/superAdminSchool";

const toast = useToastStore();

const SCHOOLS_LIMIT = 20;

const schools = ref<SuperAdminSchoolItem[]>([]);
const summary = ref<SuperAdminSchoolSummary | null>(null);
const isLoading = ref(false);
const isBootstrapping = ref(false);
const errorMessage = ref("");
const summaryError = ref("");
const searchQuery = ref("");
const bootstrapResult = ref<SuperAdminSchoolBootstrapResponse | null>(null);
const schoolsPage = ref(1);
const schoolsTotalPages = ref(1);
const schoolsTotalItems = ref(0);

let searchVersion = 0;
let searchTimer: ReturnType<typeof setTimeout> | null = null;

const schoolForm = ref({
  schoolName: "",
  schoolCode: "",
  schoolAddress: "",
  schoolEmail: "",
  schoolPhone: "",
  schoolWebsite: "",
});

const adminMode = ref<"new" | "existing">("new");
const newAdminForm = ref({
  fullName: "",
  email: "",
  password: "",
});
const existingAdminForm = ref({
  userId: "",
});

function onSearchInput() {
  if (searchTimer) clearTimeout(searchTimer);
  searchTimer = setTimeout(async () => {
    const version = ++searchVersion;
    schoolsPage.value = 1;
    isLoading.value = true;
    errorMessage.value = "";
    try {
      const response = await getSuperAdminSchools({
        page: 1,
        limit: SCHOOLS_LIMIT,
        search: searchQuery.value.trim() || undefined,
        status: "all",
      });
      if (version !== searchVersion) return;
      schools.value = response.data ?? [];
      schoolsTotalPages.value = response.totalPages ?? 1;
      schoolsTotalItems.value = Number(response.totalItems ?? 0);
    } catch (error) {
      if (version !== searchVersion) return;
      errorMessage.value = getApiError(error);
    } finally {
      if (version === searchVersion) isLoading.value = false;
    }
  }, 300);
}

function resetForm() {
  schoolForm.value = {
    schoolName: "",
    schoolCode: "",
    schoolAddress: "",
    schoolEmail: "",
    schoolPhone: "",
    schoolWebsite: "",
  };
  adminMode.value = "new";
  newAdminForm.value = {
    fullName: "",
    email: "",
    password: "",
  };
  existingAdminForm.value = {
    userId: "",
  };
}


async function loadSummary() {
  summaryError.value = "";

  try {
    summary.value = await getSuperAdminSchoolSummary();
  } catch (error) {
    summary.value = null;
    summaryError.value = getApiError(error);
  }
}

async function loadSchools(targetPage = schoolsPage.value) {
  isLoading.value = true;
  errorMessage.value = "";

  try {
    const response = await getSuperAdminSchools({
      page: targetPage,
      limit: SCHOOLS_LIMIT,
      search: searchQuery.value.trim() || undefined,
      status: "all",
    });
    schools.value = response.data ?? [];
    schoolsPage.value = response.page ?? targetPage;
    schoolsTotalPages.value = response.totalPages ?? 1;
    schoolsTotalItems.value = Number(response.totalItems ?? 0);
  } catch (error) {
    schools.value = [];
    errorMessage.value = getApiError(error);
  } finally {
    isLoading.value = false;
  }
}

async function refreshPage() {
  await Promise.all([loadSchools(), loadSummary()]);
}

function buildSchoolPayload(): SuperAdminSchoolBootstrapPayload["school"] {
  const payload: SuperAdminSchoolBootstrapPayload["school"] = {
    schoolName: schoolForm.value.schoolName.trim(),
    schoolAddress: schoolForm.value.schoolAddress.trim(),
    schoolEmail: schoolForm.value.schoolEmail.trim(),
    schoolPhone: schoolForm.value.schoolPhone.trim(),
  };

  const schoolCode = schoolForm.value.schoolCode.trim();
  const schoolWebsite = schoolForm.value.schoolWebsite.trim();

  if (schoolCode) payload.schoolCode = schoolCode;
  if (schoolWebsite) payload.schoolWebsite = schoolWebsite;

  return payload;
}

function buildBootstrapPayload(): SuperAdminSchoolBootstrapPayload {
  const school = buildSchoolPayload();

  if (adminMode.value === "existing") {
    return {
      school,
      adminUser: {
        mode: "existing",
        userId: existingAdminForm.value.userId.trim(),
      },
    };
  }

  return {
    school,
    adminUser: {
      mode: "new",
      fullName: newAdminForm.value.fullName.trim(),
      email: newAdminForm.value.email.trim(),
      password: newAdminForm.value.password,
    },
  };
}

function validateBootstrapPayload(payload: SuperAdminSchoolBootstrapPayload) {
  if (
    !payload.school.schoolName ||
    !payload.school.schoolAddress ||
    !payload.school.schoolEmail ||
    !payload.school.schoolPhone
  ) {
    return "Nama, alamat, email, dan telepon sekolah wajib diisi.";
  }

  if (payload.adminUser.mode === "new") {
    if (
      !payload.adminUser.fullName ||
      !payload.adminUser.email ||
      !payload.adminUser.password
    ) {
      return "Nama, email, dan password admin awal wajib diisi.";
    }
    return "";
  }

  if (!payload.adminUser.userId) {
    return "User ID akun global wajib diisi.";
  }

  return "";
}

async function submitBootstrap() {
  const payload = buildBootstrapPayload();
  const validationMessage = validateBootstrapPayload(payload);

  if (validationMessage) {
    toast.error(validationMessage);
    return;
  }

  isBootstrapping.value = true;
  bootstrapResult.value = null;

  try {
    const response = await bootstrapSuperAdminSchool(payload);
    bootstrapResult.value = response;
    toast.success("Sekolah dan Admin Sekolah awal berhasil disiapkan.");
    resetForm();
    await refreshPage();
  } catch (error) {
    toast.error(getApiError(error));
  } finally {
    isBootstrapping.value = false;
  }
}

onMounted(() => {
  refreshPage();
});
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-background">
    <header class="border-b border-border bg-surface">
      <div
        class="flex min-w-0 flex-col gap-3 px-5 py-5 sm:px-6 lg:flex-row lg:items-end lg:justify-between lg:px-8"
      >
        <div class="min-w-0">
          <p
            class="eyebrow"
          >
            Super Admin
          </p>
          <h1 class="mt-2 text-2xl font-semibold text-foreground sm:text-3xl">
            Sekolah
          </h1>
          <p class="mt-2 max-w-3xl text-sm leading-6 text-muted">
            Kelola tenant sekolah Wiyata dari tingkat platform. Operasional
            akademik tetap berada di area Admin Sekolah.
          </p>
        </div>
        <button
          type="button"
          class="inline-flex w-full items-center justify-center gap-2 rounded-lg border border-border bg-surface px-4 py-2.5 text-sm font-medium text-foreground-secondary transition hover:border-brand hover:text-brand disabled:cursor-not-allowed disabled:opacity-60 sm:w-auto"
          :disabled="isLoading || isBootstrapping"
          @click="refreshPage"
        >
          <PhArrowClockwise :size="16" weight="bold" />
          Muat ulang
        </button>
      </div>
    </header>

    <section
      class="grid w-full max-w-none gap-6 px-5 py-6 sm:px-6 lg:px-8 xl:grid-cols-[minmax(0,1fr)_380px]"
    >
      <div class="flex min-w-0 flex-col gap-6">
        <section class="grid gap-3 sm:grid-cols-3">
          <article
            class="rounded-xl border border-border bg-surface p-4 shadow-sm"
          >
            <p class="text-xs font-medium text-muted">Total sekolah</p>
            <p class="mt-2 text-2xl font-semibold text-foreground">
              {{ summary ? summary.totalSchools : "-" }}
            </p>
          </article>
          <article
            class="rounded-xl border border-border bg-surface p-4 shadow-sm"
          >
            <p class="text-xs font-medium text-muted">Aktif</p>
            <p class="mt-2 text-2xl font-semibold text-success">
              {{ summary ? summary.totalActive : "-" }}
            </p>
          </article>
          <article
            class="rounded-xl border border-border bg-surface p-4 shadow-sm"
          >
            <p class="text-xs font-medium text-muted">Diarsipkan</p>
            <p class="mt-2 text-2xl font-semibold text-warning">
              {{ summary ? summary.totalDeleted : "-" }}
            </p>
          </article>
        </section>

        <p
          v-if="summaryError"
          class="rounded-lg border border-warning-line bg-warning-soft px-4 py-3 text-sm leading-6 text-warning"
        >
          {{ summaryError }}
        </p>

        <section
          class="rounded-xl border border-border bg-surface p-5 shadow-sm"
        >
          <div
            class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between"
          >
            <div class="min-w-0">
              <p
                class="eyebrow"
              >
                Daftar tenant sekolah
              </p>
              <h2 class="mt-2 text-xl font-semibold text-foreground">
                Sekolah platform
              </h2>
              <p class="mt-2 text-sm leading-6 text-muted">
                Daftar ini hanya mengelola identitas tenant sekolah. Operasional
                akademik tetap dikelola oleh Admin Sekolah.
              </p>
            </div>
            <label class="relative block w-full lg:max-w-xs">
              <PhMagnifyingGlass
                :size="17"
                class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted"
              />
              <input
                v-model="searchQuery"
                type="search"
                placeholder="Cari nama, kode, email..."
                class="w-full rounded-lg border border-[#e5e7eb] bg-surface py-2.5 pl-10 pr-3 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
                @input="onSearchInput"
              />
            </label>
          </div>

          <div class="mt-5 flex flex-col gap-4">
            <div v-if="isLoading" class="space-y-3">
              <div v-for="item in 3" :key="item" class="h-24 animate-pulse rounded-xl bg-surface-subtle" />
            </div>

            <div
              v-else-if="errorMessage"
              class="rounded-lg border border-danger-line bg-danger-soft px-4 py-4"
            >
              <p class="text-sm leading-6 text-danger">{{ errorMessage }}</p>
              <button
                type="button"
                class="mt-3 inline-flex items-center justify-center gap-2 rounded-lg border border-border bg-surface px-3 py-2 text-sm font-medium text-foreground-secondary transition hover:border-brand hover:text-brand disabled:cursor-not-allowed disabled:opacity-60"
                @click="loadSchools(1)"
              >
                Coba lagi
              </button>
            </div>

            <div
              v-else-if="schools.length === 0 && !searchQuery"
              class="rounded-lg bg-surface-subtle px-5 py-8 text-center"
            >
              <PhBuildings
                class="mx-auto h-7 w-7 text-muted"
                weight="duotone"
              />
              <h3 class="mt-3 text-sm font-semibold text-foreground">
                Belum ada sekolah
              </h3>
              <p class="mt-2 text-sm leading-6 text-muted">
                Tambahkan sekolah pertama dari panel di kanan.
              </p>
            </div>

            <div
              v-else-if="schools.length === 0 && searchQuery"
              class="rounded-lg bg-surface-subtle px-5 py-8 text-center"
            >
              <PhMagnifyingGlass
                class="mx-auto h-7 w-7 text-muted"
                weight="duotone"
              />
              <h3 class="mt-3 text-sm font-semibold text-foreground">
                Tidak ada sekolah yang cocok
              </h3>
              <p class="mt-2 text-sm leading-6 text-muted">
                Ubah kata kunci pencarian untuk melihat sekolah lain.
              </p>
            </div>

            <template v-else>
              <div class="space-y-3">
              <article
                v-for="school in schools"
                :key="school.schoolId"
                class="rounded-xl border border-border bg-[#fcfbf8] p-4"
              >
                <div
                  class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between"
                >
                  <div class="min-w-0">
                    <div class="flex min-w-0 flex-wrap items-center gap-2">
                      <h3
                        class="truncate text-base font-semibold text-foreground"
                      >
                        {{ school.schoolName }}
                      </h3>
                      <span
                        class="rounded-full px-2.5 py-1 text-xs font-semibold"
                        :class="
                          school.isDeleted
                            ? 'bg-warning-soft text-warning'
                            : 'bg-success-soft text-success'
                        "
                      >
                        {{ school.isDeleted ? "Diarsipkan" : "Aktif" }}
                      </span>
                      <span
                        class="rounded-full bg-surface-strong px-2.5 py-1 text-xs font-semibold text-muted"
                      >
                        {{ school.schoolCode || "Kode otomatis" }}
                      </span>
                    </div>
                    <div
                      class="mt-3 grid gap-2 text-sm leading-6 text-muted md:grid-cols-2"
                    >
                      <p class="flex min-w-0 items-center gap-2">
                        <PhEnvelopeSimple :size="16" class="shrink-0" />
                        <span class="truncate">{{ school.schoolEmail }}</span>
                      </p>
                      <p class="flex min-w-0 items-center gap-2">
                        <PhPhone :size="16" class="shrink-0" />
                        <span class="truncate">{{ school.schoolPhone }}</span>
                      </p>
                      <p class="flex min-w-0 items-center gap-2 md:col-span-2">
                        <PhMapPin :size="16" class="shrink-0" />
                        <span class="truncate">{{ school.schoolAddress }}</span>
                      </p>
                      <p
                        v-if="school.schoolWebsite"
                        class="flex min-w-0 items-center gap-2 md:col-span-2"
                      >
                        <PhGlobe :size="16" class="shrink-0" />
                        <a
                          :href="school.schoolWebsite"
                          target="_blank"
                          rel="noopener noreferrer"
                          class="truncate text-[#ea580c] hover:underline"
                        >
                          {{ school.schoolWebsite }}
                        </a>
                      </p>
                    </div>
                  </div>

                  <p class="shrink-0 text-xs leading-5 text-muted">
                    Dibuat {{ school.createdAt }}
                  </p>
                </div>
              </article>
              </div>
              <PaginationBar
                :page="schoolsPage"
                :total-pages="schoolsTotalPages"
                :total-items="schoolsTotalItems"
                :limit="SCHOOLS_LIMIT"
                @change="(p) => loadSchools(p)"
              />
            </template>
          </div>
        </section>
      </div>

      <aside class="min-w-0">
        <section
          class="rounded-xl border border-border bg-surface p-5 shadow-sm xl:sticky xl:top-6"
        >
          <div class="flex items-start justify-between gap-4">
            <div>
              <p
                class="eyebrow"
              >
                Setup sekolah
              </p>
              <h2 class="mt-2 text-xl font-semibold text-foreground">
                Sekolah + admin awal
              </h2>
              <p class="mt-2 text-sm leading-6 text-muted">
                Buat tenant sekolah dan beri akses Admin Sekolah dalam satu
                proses atomik.
              </p>
            </div>
            <span
              class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-[#fff4ee] text-[#ea580c]"
            >
              <PhBuildings :size="22" weight="duotone" />
            </span>
          </div>

          <div
            v-if="bootstrapResult"
            class="mt-5 rounded-xl border border-success-line bg-success-soft p-4"
          >
            <div class="flex items-start gap-3">
              <PhCheckCircle
                :size="22"
                class="mt-0.5 shrink-0 text-success"
                weight="duotone"
              />
              <div class="min-w-0">
                <p class="text-sm font-semibold text-success">
                  Setup berhasil
                </p>
                <p class="mt-1 text-xs leading-5 text-success">
                  {{ bootstrapResult.school.schoolName }} sudah dibuat dan
                  {{ bootstrapResult.adminUser.fullName }} mendapat role admin
                  sekolah.
                </p>
              </div>
            </div>
          </div>

          <form class="mt-5 space-y-5" @submit.prevent="submitBootstrap">
            <section class="space-y-4">
              <div>
                <p class="text-sm font-semibold text-foreground">
                  Identitas sekolah
                </p>
                <p class="mt-1 text-xs leading-5 text-muted">
                  Data ini membuat tenant sekolah platform, bukan pengaturan
                  akademik.
                </p>
              </div>

              <label class="block text-sm font-medium text-foreground-secondary">
                Nama sekolah
                <input
                  v-model="schoolForm.schoolName"
                  type="text"
                  class="mt-2 w-full rounded-lg border border-[#e5e7eb] bg-surface px-4 py-3 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
                  placeholder="Contoh: SMA Wiyata"
                />
              </label>

              <label class="block text-sm font-medium text-foreground-secondary">
                Kode sekolah
                <input
                  v-model="schoolForm.schoolCode"
                  type="text"
                  class="mt-2 w-full rounded-lg border border-[#e5e7eb] bg-surface px-4 py-3 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
                  placeholder="Kosongkan untuk kode otomatis"
                />
              </label>

              <label class="block text-sm font-medium text-foreground-secondary">
                Email sekolah
                <input
                  v-model="schoolForm.schoolEmail"
                  type="email"
                  class="mt-2 w-full rounded-lg border border-[#e5e7eb] bg-surface px-4 py-3 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
                  placeholder="admin@sekolah.sch.id"
                />
              </label>

              <label class="block text-sm font-medium text-foreground-secondary">
                Telepon sekolah
                <input
                  v-model="schoolForm.schoolPhone"
                  type="tel"
                  class="mt-2 w-full rounded-lg border border-[#e5e7eb] bg-surface px-4 py-3 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
                  placeholder="081234567890"
                />
              </label>

              <label class="block text-sm font-medium text-foreground-secondary">
                Website
                <input
                  v-model="schoolForm.schoolWebsite"
                  type="url"
                  class="mt-2 w-full rounded-lg border border-[#e5e7eb] bg-surface px-4 py-3 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
                  placeholder="https://sekolah.sch.id"
                />
              </label>

              <label class="block text-sm font-medium text-foreground-secondary">
                Alamat sekolah
                <textarea
                  v-model="schoolForm.schoolAddress"
                  rows="4"
                  class="mt-2 w-full resize-none rounded-lg border border-[#e5e7eb] bg-surface px-4 py-3 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
                  placeholder="Alamat lengkap sekolah"
                />
              </label>
            </section>

            <section class="space-y-4 border-t border-border pt-5">
              <div>
                <p class="text-sm font-semibold text-foreground">Admin awal</p>
                <p class="mt-1 text-xs leading-5 text-muted">
                  Akun ini akan menjadi Admin Sekolah untuk tenant baru.
                </p>
              </div>

              <div class="grid gap-2 sm:grid-cols-2">
                <label
                  class="flex cursor-pointer items-start gap-3 rounded-lg border p-3 text-sm transition"
                  :class="
                    adminMode === 'new'
                      ? 'border-[#ea580c] bg-warning-soft'
                      : 'border-[#e5e7eb] bg-surface hover:bg-[#fafafa]'
                  "
                >
                  <input
                    v-model="adminMode"
                    type="radio"
                    value="new"
                    class="mt-1"
                  />
                  <span>
                    <span class="block font-semibold text-foreground">
                      Akun admin baru
                    </span>
                    <span class="mt-1 block text-xs leading-5 text-muted">
                      Buat user global baru.
                    </span>
                  </span>
                </label>

                <label
                  class="flex cursor-pointer items-start gap-3 rounded-lg border p-3 text-sm transition"
                  :class="
                    adminMode === 'existing'
                      ? 'border-[#ea580c] bg-warning-soft'
                      : 'border-[#e5e7eb] bg-surface hover:bg-[#fafafa]'
                  "
                >
                  <input
                    v-model="adminMode"
                    type="radio"
                    value="existing"
                    class="mt-1"
                  />
                  <span>
                    <span class="block font-semibold text-foreground">
                      Gunakan akun global
                    </span>
                    <span class="mt-1 block text-xs leading-5 text-muted">
                      Pakai user ID yang sudah ada.
                    </span>
                  </span>
                </label>
              </div>

              <div v-if="adminMode === 'new'" class="space-y-4">
                <label class="block text-sm font-medium text-foreground-secondary">
                  Nama admin sekolah
                  <input
                    v-model="newAdminForm.fullName"
                    type="text"
                    class="mt-2 w-full rounded-lg border border-[#e5e7eb] bg-surface px-4 py-3 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
                    placeholder="Admin Sekolah"
                  />
                </label>

                <label class="block text-sm font-medium text-foreground-secondary">
                  Email admin sekolah
                  <input
                    v-model="newAdminForm.email"
                    type="email"
                    class="mt-2 w-full rounded-lg border border-[#e5e7eb] bg-surface px-4 py-3 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
                    placeholder="admin@sma.sch.id"
                  />
                </label>

                <label class="block text-sm font-medium text-foreground-secondary">
                  Password awal
                  <input
                    v-model="newAdminForm.password"
                    type="password"
                    class="mt-2 w-full rounded-lg border border-[#e5e7eb] bg-surface px-4 py-3 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
                    placeholder="Minimal 6 karakter"
                  />
                </label>
              </div>

              <div v-else class="space-y-3">
                <label class="block text-sm font-medium text-foreground-secondary">
                  User ID akun global
                  <input
                    v-model="existingAdminForm.userId"
                    type="text"
                    class="mt-2 w-full rounded-lg border border-[#e5e7eb] bg-surface px-4 py-3 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
                    placeholder="UUID user global"
                  />
                </label>
                <p
                  class="rounded-lg border border-[#e5e7eb] bg-[#fafafa] px-3 py-2 text-xs leading-5 text-muted"
                >
                  Ambil User ID dari data akun global yang sudah ada. Setelah
                  submit, akun tersebut mendapat role admin sekolah untuk
                  sekolah baru.
                </p>
              </div>
            </section>

            <button
              type="submit"
              class="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-[#ea580c] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#c2410c] disabled:cursor-not-allowed disabled:opacity-60"
              :disabled="isBootstrapping || isLoading"
            >
              <PhPlusCircle :size="18" weight="duotone" />
              {{ isBootstrapping ? "Menyiapkan..." : "Setup sekolah" }}
            </button>

            <div
              class="flex gap-3 rounded-lg border border-border bg-[#fcfbf8] p-3"
            >
              <PhIdentificationBadge
                :size="20"
                class="mt-0.5 shrink-0 text-[#ea580c]"
                weight="duotone"
              />
              <p class="text-xs leading-5 text-muted">
                Flow ini hanya membuat tenant sekolah dan akses Admin Sekolah
                awal. Tahun ajaran, kelas, dan penempatan tetap dikelola dari
                area Admin Sekolah.
              </p>
            </div>
          </form>
        </section>
      </aside>
    </section>
  </main>
</template>
