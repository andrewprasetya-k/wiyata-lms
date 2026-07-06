<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
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

const schools = ref<SuperAdminSchoolItem[]>([]);
const summary = ref<SuperAdminSchoolSummary | null>(null);
const isLoading = ref(false);
const isBootstrapping = ref(false);
const errorMessage = ref("");
const summaryError = ref("");
const searchQuery = ref("");
const bootstrapResult = ref<SuperAdminSchoolBootstrapResponse | null>(null);

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

const filteredSchools = computed(() => {
  const query = searchQuery.value.trim().toLowerCase();
  if (!query) return schools.value;

  return schools.value.filter((school) => {
    return [
      school.schoolName,
      school.schoolCode,
      school.schoolEmail,
      school.schoolPhone,
    ]
      .filter(Boolean)
      .some((value) => value.toLowerCase().includes(query));
  });
});

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

function getApiErrorMessage(error: unknown, fallback: string) {
  if (typeof error === "object" && error !== null && "response" in error) {
    const response = (
      error as {
        response?: { data?: { error?: unknown; message?: unknown } | string };
      }
    ).response;
    if (typeof response?.data === "string") return response.data;
    if (typeof response?.data?.error === "string") return response.data.error;
    if (typeof response?.data?.message === "string")
      return response.data.message;
  }

  return fallback;
}

async function loadSummary() {
  summaryError.value = "";

  try {
    summary.value = await getSuperAdminSchoolSummary();
  } catch (error) {
    summary.value = null;
    summaryError.value = getApiErrorMessage(
      error,
      "Ringkasan sekolah belum bisa dimuat.",
    );
  }
}

async function loadSchools() {
  isLoading.value = true;
  errorMessage.value = "";

  try {
    const response = await getSuperAdminSchools({
      page: 1,
      limit: 100,
      status: "all",
    });
    schools.value = response.data ?? [];
  } catch (error) {
    schools.value = [];
    errorMessage.value = getApiErrorMessage(
      error,
      "Daftar sekolah belum bisa dimuat.",
    );
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
    toast.error(
      getApiErrorMessage(
        error,
        "Setup sekolah belum bisa disimpan. Pastikan data sekolah dan admin awal valid.",
      ),
    );
  } finally {
    isBootstrapping.value = false;
  }
}

onMounted(() => {
  refreshPage();
});
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-[#f8f7f4]">
    <header class="border-b border-[#ebe7df] bg-white">
      <div
        class="flex min-w-0 flex-col gap-3 px-5 py-5 sm:px-6 lg:flex-row lg:items-end lg:justify-between lg:px-8"
      >
        <div class="min-w-0">
          <p
            class="text-xs font-semibold uppercase tracking-[0.18em] text-[#ea580c]"
          >
            Super Admin
          </p>
          <h1 class="mt-2 text-2xl font-semibold text-[#171322] sm:text-3xl">
            Sekolah
          </h1>
          <p class="mt-2 max-w-3xl text-sm leading-6 text-[#6b7280]">
            Kelola tenant sekolah Wiyata dari tingkat platform. Operasional
            akademik tetap berada di area Admin Sekolah.
          </p>
        </div>
        <button
          type="button"
          class="inline-flex w-full items-center justify-center gap-2 rounded-lg border border-[#e5e7eb] bg-white px-4 py-2.5 text-sm font-semibold text-[#171322] transition hover:bg-[#fafafa] disabled:cursor-not-allowed disabled:opacity-60 sm:w-auto"
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
            class="rounded-xl border border-[#ebe7df] bg-white p-4 shadow-sm"
          >
            <p class="text-xs font-medium text-[#6b7280]">Total sekolah</p>
            <p class="mt-2 text-2xl font-semibold text-[#171322]">
              {{ summary ? summary.totalSchools : "-" }}
            </p>
          </article>
          <article
            class="rounded-xl border border-[#ebe7df] bg-white p-4 shadow-sm"
          >
            <p class="text-xs font-medium text-[#6b7280]">Aktif</p>
            <p class="mt-2 text-2xl font-semibold text-[#027a48]">
              {{ summary ? summary.totalActive : "-" }}
            </p>
          </article>
          <article
            class="rounded-xl border border-[#ebe7df] bg-white p-4 shadow-sm"
          >
            <p class="text-xs font-medium text-[#6b7280]">Diarsipkan</p>
            <p class="mt-2 text-2xl font-semibold text-[#b45309]">
              {{ summary ? summary.totalDeleted : "-" }}
            </p>
          </article>
        </section>

        <p
          v-if="summaryError"
          class="rounded-lg border border-[#fed7aa] bg-[#fff7ed] px-4 py-3 text-sm leading-6 text-[#9a3412]"
        >
          {{ summaryError }}
        </p>

        <section
          class="rounded-xl border border-[#ebe7df] bg-white p-5 shadow-sm"
        >
          <div
            class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between"
          >
            <div class="min-w-0">
              <p
                class="text-xs font-semibold uppercase tracking-[0.16em] text-[#ea580c]"
              >
                Daftar tenant sekolah
              </p>
              <h2 class="mt-2 text-xl font-semibold text-[#171322]">
                Sekolah platform
              </h2>
              <p class="mt-2 text-sm leading-6 text-[#6b7280]">
                Daftar ini hanya mengelola identitas tenant sekolah. Operasional
                akademik tetap dikelola oleh Admin Sekolah.
              </p>
            </div>
            <label class="relative block w-full lg:max-w-xs">
              <PhMagnifyingGlass
                :size="17"
                class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-[#9ca3af]"
              />
              <input
                v-model="searchQuery"
                type="search"
                placeholder="Cari nama, kode, email..."
                class="w-full rounded-lg border border-[#e5e7eb] bg-white py-2.5 pl-10 pr-3 text-sm text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
              />
            </label>
          </div>

          <div class="mt-5 space-y-3">
            <div
              v-if="isLoading"
              class="rounded-lg border border-[#e5e7eb] bg-[#fafafa] px-4 py-5 text-sm text-[#6b7280]"
            >
              Memuat daftar sekolah...
            </div>

            <div
              v-else-if="errorMessage"
              class="rounded-lg border border-[#fecaca] bg-[#fff8f6] px-4 py-4"
            >
              <p class="text-sm leading-6 text-[#a8665d]">{{ errorMessage }}</p>
              <button
                type="button"
                class="mt-3 inline-flex items-center gap-2 rounded-lg border border-[#fecaca] bg-white px-3 py-2 text-sm font-semibold text-[#a8665d] transition hover:bg-[#fff8f6]"
                @click="loadSchools"
              >
                Coba lagi
              </button>
            </div>

            <div
              v-else-if="schools.length === 0"
              class="rounded-lg bg-[#fbfaf8] px-5 py-8 text-center"
            >
              <PhBuildings
                class="mx-auto h-7 w-7 text-[#9ca3af]"
                weight="duotone"
              />
              <h3 class="mt-3 text-sm font-semibold text-[#171322]">
                Belum ada sekolah
              </h3>
              <p class="mt-2 text-sm leading-6 text-[#6b7280]">
                Tambahkan sekolah pertama dari panel di kanan.
              </p>
            </div>

            <div
              v-else-if="filteredSchools.length === 0"
              class="rounded-lg bg-[#fbfaf8] px-5 py-8 text-center"
            >
              <PhMagnifyingGlass
                class="mx-auto h-7 w-7 text-[#9ca3af]"
                weight="duotone"
              />
              <h3 class="mt-3 text-sm font-semibold text-[#171322]">
                Tidak ada sekolah yang cocok
              </h3>
              <p class="mt-2 text-sm leading-6 text-[#6b7280]">
                Ubah kata kunci pencarian untuk melihat sekolah lain.
              </p>
            </div>

            <template v-else>
              <article
                v-for="school in filteredSchools"
                :key="school.schoolId"
                class="rounded-xl border border-[#ebe7df] bg-[#fcfbf8] p-4"
              >
                <div
                  class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between"
                >
                  <div class="min-w-0">
                    <div class="flex min-w-0 flex-wrap items-center gap-2">
                      <h3
                        class="truncate text-base font-semibold text-[#171322]"
                      >
                        {{ school.schoolName }}
                      </h3>
                      <span
                        class="rounded-full px-2.5 py-1 text-xs font-semibold"
                        :class="
                          school.isDeleted
                            ? 'bg-[#fff7ed] text-[#b45309]'
                            : 'bg-[#ecfdf3] text-[#027a48]'
                        "
                      >
                        {{ school.isDeleted ? "Diarsipkan" : "Aktif" }}
                      </span>
                      <span
                        class="rounded-full bg-[#f3f4f6] px-2.5 py-1 text-xs font-semibold text-[#6b7280]"
                      >
                        {{ school.schoolCode || "Kode otomatis" }}
                      </span>
                    </div>
                    <div
                      class="mt-3 grid gap-2 text-sm leading-6 text-[#6b7280] md:grid-cols-2"
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

                  <p class="shrink-0 text-xs leading-5 text-[#9ca3af]">
                    Dibuat {{ school.createdAt }}
                  </p>
                </div>
              </article>
            </template>
          </div>
        </section>
      </div>

      <aside class="min-w-0">
        <section
          class="rounded-xl border border-[#ebe7df] bg-white p-5 shadow-sm xl:sticky xl:top-6"
        >
          <div class="flex items-start justify-between gap-4">
            <div>
              <p
                class="text-xs font-semibold uppercase tracking-[0.16em] text-[#ea580c]"
              >
                Setup sekolah
              </p>
              <h2 class="mt-2 text-xl font-semibold text-[#171322]">
                Sekolah + admin awal
              </h2>
              <p class="mt-2 text-sm leading-6 text-[#6b7280]">
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
            class="mt-5 rounded-xl border border-[#bbf7d0] bg-[#f0fdf4] p-4"
          >
            <div class="flex items-start gap-3">
              <PhCheckCircle
                :size="22"
                class="mt-0.5 shrink-0 text-[#027a48]"
                weight="duotone"
              />
              <div class="min-w-0">
                <p class="text-sm font-semibold text-[#166534]">
                  Setup berhasil
                </p>
                <p class="mt-1 text-xs leading-5 text-[#166534]">
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
                <p class="text-sm font-semibold text-[#171322]">
                  Identitas sekolah
                </p>
                <p class="mt-1 text-xs leading-5 text-[#6b7280]">
                  Data ini membuat tenant sekolah platform, bukan pengaturan
                  akademik.
                </p>
              </div>

              <label class="block text-sm font-medium text-[#374151]">
                Nama sekolah
                <input
                  v-model="schoolForm.schoolName"
                  type="text"
                  class="mt-2 w-full rounded-lg border border-[#e5e7eb] bg-white px-4 py-3 text-sm text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
                  placeholder="Contoh: SMA Wiyata"
                />
              </label>

              <label class="block text-sm font-medium text-[#374151]">
                Kode sekolah
                <input
                  v-model="schoolForm.schoolCode"
                  type="text"
                  class="mt-2 w-full rounded-lg border border-[#e5e7eb] bg-white px-4 py-3 text-sm text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
                  placeholder="Kosongkan untuk kode otomatis"
                />
              </label>

              <label class="block text-sm font-medium text-[#374151]">
                Email sekolah
                <input
                  v-model="schoolForm.schoolEmail"
                  type="email"
                  class="mt-2 w-full rounded-lg border border-[#e5e7eb] bg-white px-4 py-3 text-sm text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
                  placeholder="admin@sekolah.sch.id"
                />
              </label>

              <label class="block text-sm font-medium text-[#374151]">
                Telepon sekolah
                <input
                  v-model="schoolForm.schoolPhone"
                  type="tel"
                  class="mt-2 w-full rounded-lg border border-[#e5e7eb] bg-white px-4 py-3 text-sm text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
                  placeholder="081234567890"
                />
              </label>

              <label class="block text-sm font-medium text-[#374151]">
                Website
                <input
                  v-model="schoolForm.schoolWebsite"
                  type="url"
                  class="mt-2 w-full rounded-lg border border-[#e5e7eb] bg-white px-4 py-3 text-sm text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
                  placeholder="https://sekolah.sch.id"
                />
              </label>

              <label class="block text-sm font-medium text-[#374151]">
                Alamat sekolah
                <textarea
                  v-model="schoolForm.schoolAddress"
                  rows="4"
                  class="mt-2 w-full resize-none rounded-lg border border-[#e5e7eb] bg-white px-4 py-3 text-sm text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
                  placeholder="Alamat lengkap sekolah"
                />
              </label>
            </section>

            <section class="space-y-4 border-t border-[#ebe7df] pt-5">
              <div>
                <p class="text-sm font-semibold text-[#171322]">Admin awal</p>
                <p class="mt-1 text-xs leading-5 text-[#6b7280]">
                  Akun ini akan menjadi Admin Sekolah untuk tenant baru.
                </p>
              </div>

              <div class="grid gap-2 sm:grid-cols-2">
                <label
                  class="flex cursor-pointer items-start gap-3 rounded-lg border p-3 text-sm transition"
                  :class="
                    adminMode === 'new'
                      ? 'border-[#ea580c] bg-[#fff7ed]'
                      : 'border-[#e5e7eb] bg-white hover:bg-[#fafafa]'
                  "
                >
                  <input
                    v-model="adminMode"
                    type="radio"
                    value="new"
                    class="mt-1"
                  />
                  <span>
                    <span class="block font-semibold text-[#171322]">
                      Akun admin baru
                    </span>
                    <span class="mt-1 block text-xs leading-5 text-[#6b7280]">
                      Buat user global baru.
                    </span>
                  </span>
                </label>

                <label
                  class="flex cursor-pointer items-start gap-3 rounded-lg border p-3 text-sm transition"
                  :class="
                    adminMode === 'existing'
                      ? 'border-[#ea580c] bg-[#fff7ed]'
                      : 'border-[#e5e7eb] bg-white hover:bg-[#fafafa]'
                  "
                >
                  <input
                    v-model="adminMode"
                    type="radio"
                    value="existing"
                    class="mt-1"
                  />
                  <span>
                    <span class="block font-semibold text-[#171322]">
                      Gunakan akun global
                    </span>
                    <span class="mt-1 block text-xs leading-5 text-[#6b7280]">
                      Pakai user ID yang sudah ada.
                    </span>
                  </span>
                </label>
              </div>

              <div v-if="adminMode === 'new'" class="space-y-4">
                <label class="block text-sm font-medium text-[#374151]">
                  Nama admin sekolah
                  <input
                    v-model="newAdminForm.fullName"
                    type="text"
                    class="mt-2 w-full rounded-lg border border-[#e5e7eb] bg-white px-4 py-3 text-sm text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
                    placeholder="Admin Sekolah"
                  />
                </label>

                <label class="block text-sm font-medium text-[#374151]">
                  Email admin sekolah
                  <input
                    v-model="newAdminForm.email"
                    type="email"
                    class="mt-2 w-full rounded-lg border border-[#e5e7eb] bg-white px-4 py-3 text-sm text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
                    placeholder="admin@sma.sch.id"
                  />
                </label>

                <label class="block text-sm font-medium text-[#374151]">
                  Password awal
                  <input
                    v-model="newAdminForm.password"
                    type="password"
                    class="mt-2 w-full rounded-lg border border-[#e5e7eb] bg-white px-4 py-3 text-sm text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
                    placeholder="Minimal 6 karakter"
                  />
                </label>
              </div>

              <div v-else class="space-y-3">
                <label class="block text-sm font-medium text-[#374151]">
                  User ID akun global
                  <input
                    v-model="existingAdminForm.userId"
                    type="text"
                    class="mt-2 w-full rounded-lg border border-[#e5e7eb] bg-white px-4 py-3 text-sm text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
                    placeholder="UUID user global"
                  />
                </label>
                <p
                  class="rounded-lg border border-[#e5e7eb] bg-[#fafafa] px-3 py-2 text-xs leading-5 text-[#6b7280]"
                >
                  Ambil User ID dari data akun global yang sudah ada. Setelah
                  submit, akun tersebut mendapat role admin sekolah untuk
                  sekolah baru.
                </p>
              </div>
            </section>

            <button
              type="submit"
              class="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-[#171322] px-4 py-3 text-sm font-semibold text-white transition hover:bg-[#2f2b3a] disabled:cursor-not-allowed disabled:opacity-60"
              :disabled="isBootstrapping || isLoading"
            >
              <PhPlusCircle :size="18" weight="duotone" />
              {{ isBootstrapping ? "Menyiapkan..." : "Setup sekolah" }}
            </button>

            <div
              class="flex gap-3 rounded-lg border border-[#ebe7df] bg-[#fcfbf8] p-3"
            >
              <PhIdentificationBadge
                :size="20"
                class="mt-0.5 shrink-0 text-[#ea580c]"
                weight="duotone"
              />
              <p class="text-xs leading-5 text-[#6b7280]">
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
