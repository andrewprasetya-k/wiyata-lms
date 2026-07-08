<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import {
  PhCopy,
  PhDownloadSimple,
  PhEnvelopeSimple,
  PhFileCsv,
  PhMagnifyingGlass,
  PhPlusCircle,
  PhShieldCheck,
  PhTrash,
  PhUploadSimple,
  PhUsers,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import * as XLSX from "xlsx";
import { useAuthStore } from "../../stores/auth";
import { useToastStore } from "../../stores/toast";
import { getRoles, syncUserRoles } from "../../services/adminUser";
import {
  createAdminSchoolMember,
  getAdminSchoolMembers,
  removeAdminSchoolMember,
} from "../../services/adminSchoolMember";
import { createSchoolMemberInvitation } from "../../services/adminSchoolMemberInvitation";
import {
  commitSchoolMemberImport,
  previewSchoolMemberImport,
} from "../../services/adminSchoolMemberImport";
import type { RoleItem } from "../../types/adminUser";
import type {
  AdminSchoolMemberCreatePayload,
  AdminSchoolMemberItem,
} from "../../types/adminSchoolMember";
import type {
  AdminSchoolMemberImportCommitResponse,
  AdminSchoolMemberImportPreviewResponse,
} from "../../types/adminSchoolMemberImport";
import type {
  CreateSchoolMemberInvitationPayload,
  CreateSchoolMemberInvitationResponse,
} from "../../types/adminSchoolMemberInvitation";
import { formatDateTime } from "../../utils/date";
import { getApiError } from "../../utils/error";

const allowedRoleNames = ["student", "teacher", "admin"];
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

const members = ref<AdminSchoolMemberItem[]>([]);
const roles = ref<RoleItem[]>([]);
const memberRoleDrafts = ref<Record<string, string>>({});

const membersLoading = ref(false);
const rolesLoading = ref(false);
const savingRolesSchoolUserId = ref("");
const importPreviewLoading = ref(false);
const importCommitLoading = ref(false);
const isInvitingMember = ref(false);
const isCreatingMember = ref(false);
const removingSchoolUserId = ref("");

const membersError = ref("");
const rolesError = ref("");
const importError = ref("");

const memberSearch = ref("");
const memberEntryMode = ref<"invite" | "direct">("invite");
const importFile = ref<File | null>(null);
const importDefaultPassword = ref("");
const importPreview = ref<AdminSchoolMemberImportPreviewResponse | null>(null);
const importResult = ref<AdminSchoolMemberImportCommitResponse | null>(null);
const inviteResult = ref<CreateSchoolMemberInvitationResponse | null>(null);
const inviteForm = ref<CreateSchoolMemberInvitationPayload>({
  fullName: "",
  email: "",
  role: "student",
  classCode: "",
});
const manualForm = ref<AdminSchoolMemberCreatePayload>({
  fullName: "",
  email: "",
  password: "",
  role: "student",
  classCode: "",
});

const allowedRoles = computed(() =>
  roles.value.filter((role) =>
    allowedRoleNames.includes(normalizeRoleName(role.roleName)),
  ),
);

function normalizeRoleName(roleName: string) {
  return roleName.trim().toLowerCase();
}

function roleLabel(roleName: string) {
  const normalized = normalizeRoleName(roleName);
  if (normalized === "student") return "Siswa";
  if (normalized === "teacher") return "Guru";
  if (normalized === "admin") return "Admin sekolah";
  return roleName;
}

function memberCreateSuccessMessage(member: AdminSchoolMemberItem) {
  if (member.userCreated === true) {
    return "Akun baru berhasil dibuat. Password awal diberikan oleh admin/sekolah.";
  }
  if (member.userCreated === false) {
    return "Akun global sudah ada. Password tidak diubah; pengguna login memakai password yang sudah ada.";
  }
  return "Warga sekolah berhasil ditambahkan.";
}

function importResultNote(result: {
  status: string;
  userCreated?: boolean;
  reason?: string;
}) {
  if (result.status === "failed") {
    return result.reason || "Baris belum bisa diproses.";
  }
  if (result.status === "skipped") {
    return result.reason || "Data sudah sesuai, tidak ada perubahan.";
  }
  if (result.userCreated === true) {
    return "Akun baru dibuat. Password awal diberikan oleh admin/sekolah.";
  }
  if (result.userCreated === false) {
    return "Akun global sudah ada. Password tidak diubah.";
  }
  return result.reason || "Berhasil diproses.";
}

function importEmailNote(emailNotification?: string) {
  if (emailNotification === "account_created") {
    return "Email info akun dikirim tanpa password.";
  }
  if (emailNotification === "added_to_school") {
    return "Email penambahan sekolah dikirim tanpa password.";
  }
  return "";
}

function rolePriority(roleName: string) {
  const normalized = normalizeRoleName(roleName);
  if (normalized === "admin") return 0;
  if (normalized === "teacher") return 1;
  if (normalized === "student") return 2;
  return 99;
}

function initializeRoleDrafts() {
  const roleByName = new Map(
    allowedRoles.value.map((role) => [
      normalizeRoleName(role.roleName),
      role.roleId,
    ]),
  );
  const nextDrafts: Record<string, string> = {};

  for (const member of members.value) {
    const selectedRole =
      member.roles
        ?.filter((roleName) => roleByName.has(normalizeRoleName(roleName)))
        .sort((a, b) => rolePriority(a) - rolePriority(b))[0] ?? "";
    nextDrafts[member.schoolUserId] = selectedRole
      ? (roleByName.get(normalizeRoleName(selectedRole)) ?? "")
      : "";
  }

  memberRoleDrafts.value = nextDrafts;
}

function primaryRoleName(member: AdminSchoolMemberItem) {
  return (
    member.roles
      ?.filter((roleName) =>
        allowedRoleNames.includes(normalizeRoleName(roleName)),
      )
      .sort((a, b) => rolePriority(a) - rolePriority(b))[0] ?? ""
  );
}

function hasMultipleAllowedRoles(member: AdminSchoolMemberItem) {
  const uniqueRoles = new Set(
    member.roles
      ?.map((roleName) => normalizeRoleName(roleName))
      .filter((roleName) => allowedRoleNames.includes(roleName)) ?? [],
  );
  return uniqueRoles.size > 1;
}

async function loadRoles() {
  rolesLoading.value = true;
  rolesError.value = "";
  try {
    const data = await getRoles();
    roles.value = data ?? [];
  } catch {
    rolesError.value = "Daftar peran belum bisa dimuat.";
  } finally {
    rolesLoading.value = false;
  }
}

async function loadMembers() {
  if (!currentSchool.value.hasContext) return;

  membersLoading.value = true;
  membersError.value = "";
  try {
    const data = await getAdminSchoolMembers({
      page: 1,
      limit: 50,
      search: memberSearch.value.trim(),
    });
    members.value = data.data ?? [];
    initializeRoleDrafts();
  } catch {
    membersError.value = "Warga sekolah belum bisa dimuat.";
  } finally {
    membersLoading.value = false;
  }
}

let searchVersion = 0;
let searchTimer: ReturnType<typeof setTimeout> | null = null;

watch(memberSearch, () => {
  if (searchTimer) clearTimeout(searchTimer);
  searchTimer = setTimeout(async () => {
    const version = ++searchVersion;
    if (!currentSchool.value.hasContext) return;

    membersLoading.value = true;
    membersError.value = "";
    try {
      const data = await getAdminSchoolMembers({
        page: 1,
        limit: 50,
        search: memberSearch.value.trim(),
      });
      if (version !== searchVersion) return;
      members.value = data.data ?? [];
      initializeRoleDrafts();
    } catch {
      if (version !== searchVersion) return;
      membersError.value = "Warga sekolah belum bisa dimuat.";
    } finally {
      if (version === searchVersion) membersLoading.value = false;
    }
  }, 300);
});

async function syncRoleForMember(schoolUserId: string, roleId: string) {
  if (!roleId) {
    toast.error("Pilih satu peran.");
    return;
  }

  savingRolesSchoolUserId.value = schoolUserId;
  try {
    await syncUserRoles(schoolUserId, { roleIds: [roleId] });
    toast.success("Peran warga sekolah berhasil diperbarui.");
    await loadMembers();
  } catch {
    toast.error("Peran warga sekolah belum bisa diperbarui.");
  } finally {
    savingRolesSchoolUserId.value = "";
  }
}

function setRoleDraft(schoolUserId: string, roleId: string) {
  memberRoleDrafts.value = {
    ...memberRoleDrafts.value,
    [schoolUserId]: roleId,
  };
}


const inviteLink = computed(() => {
  const acceptUrl = inviteResult.value?.acceptUrl;
  if (!acceptUrl) return "";
  if (/^https?:\/\//i.test(acceptUrl)) return acceptUrl;
  return `${window.location.origin}${acceptUrl.startsWith("/") ? "" : "/"}${acceptUrl}`;
});

function setMemberEntryMode(mode: "invite" | "direct") {
  memberEntryMode.value = mode;
}

function resetInviteForm() {
  inviteForm.value = {
    fullName: "",
    email: "",
    role: "student",
    classCode: "",
  };
}

async function submitInviteMember() {
  const payload: CreateSchoolMemberInvitationPayload = {
    fullName: inviteForm.value.fullName.trim(),
    email: inviteForm.value.email.trim(),
    role: inviteForm.value.role,
    classCode:
      inviteForm.value.role === "student"
        ? inviteForm.value.classCode?.trim() || undefined
        : undefined,
  };

  if (!payload.fullName || !payload.email || !payload.role) {
    toast.error("Nama, email, dan peran wajib diisi.");
    return;
  }
  if (payload.role === "student" && !payload.classCode) {
    toast.error("Kode kelas wajib diisi untuk undangan siswa.");
    return;
  }

  isInvitingMember.value = true;
  inviteResult.value = null;
  try {
    inviteResult.value = await createSchoolMemberInvitation(payload);
    toast.success("Undangan email berhasil dibuat.");
    resetInviteForm();
  } catch (error) {
    toast.error(getApiError(error));
  } finally {
    isInvitingMember.value = false;
  }
}

async function copyInviteLink() {
  if (!inviteLink.value) return;
  try {
    await navigator.clipboard.writeText(inviteLink.value);
    toast.success("Link undangan disalin.");
  } catch {
    toast.error("Link belum bisa disalin otomatis.");
  }
}

const importTemplateRows = [
  ["fullName", "email", "role", "classCode"],
  ["Budi Santoso", "budi@siswa.sch.id", "student", "X-IPA-1"],
  ["Siti Rahma", "siti@guru.sch.id", "teacher", ""],
  ["Admin Sekolah", "admin@sekolah.sch.id", "admin", ""],
];

type ImportHeader = "fullName" | "email" | "role" | "classCode";

function downloadTemplate() {
  const csv = toCsv(importTemplateRows);
  const blob = new Blob([csv], { type: "text/csv;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = "template-import-warga-sekolah.csv";
  link.click();
  URL.revokeObjectURL(url);
}

function downloadExcelTemplate() {
  const worksheet = XLSX.utils.aoa_to_sheet(importTemplateRows);
  worksheet["!cols"] = [{ wch: 24 }, { wch: 28 }, { wch: 14 }, { wch: 16 }];
  worksheet["!autofilter"] = { ref: "A1:D4" };

  const workbook = XLSX.utils.book_new();
  XLSX.utils.book_append_sheet(workbook, worksheet, "Import Warga");
  XLSX.writeFile(workbook, "template-import-warga-sekolah.xlsx", {
    compression: true,
  });
}

function csvEscape(value: unknown) {
  const text = String(value ?? "");
  if (/[",\n\r]/.test(text)) {
    return `"${text.replace(/"/g, '""')}"`;
  }
  return text;
}

function toCsv(rows: unknown[][]) {
  return rows.map((row) => row.map(csvEscape).join(",")).join("\n");
}

function normalizeImportHeader(value: unknown): ImportHeader | "" {
  const normalized = String(value ?? "")
    .trim()
    .toLowerCase()
    .replace(/[\s_-]+/g, "");

  if (
    normalized === "fullname" ||
    normalized === "nama" ||
    normalized === "namalengkap"
  ) {
    return "fullName";
  }
  if (normalized === "email" || normalized === "alamatemail") {
    return "email";
  }
  if (normalized === "role" || normalized === "peran") {
    return "role";
  }
  if (normalized === "classcode" || normalized === "kodekelas") {
    return "classCode";
  }
  return "";
}

function isExcelFile(file: File) {
  const name = file.name.toLowerCase();
  return (
    name.endsWith(".xlsx") ||
    file.type ===
      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
  );
}

async function convertXlsxToCsvFile(file: File) {
  const workbook = XLSX.read(await file.arrayBuffer(), {
    type: "array",
    cellDates: false,
  });
  const sheetName = workbook.SheetNames[0];
  if (!sheetName) {
    throw new Error("Workbook Excel tidak memiliki sheet.");
  }

  const worksheet = workbook.Sheets[sheetName];
  const rawRows = XLSX.utils.sheet_to_json<unknown[]>(worksheet, {
    header: 1,
    blankrows: false,
    defval: "",
  });

  const [rawHeader, ...dataRows] = rawRows;
  if (!rawHeader) {
    throw new Error("Sheet Excel kosong.");
  }

  const headers = rawHeader.map(normalizeImportHeader);
  const requiredHeaders: ImportHeader[] = ["fullName", "email", "role"];
  const missingHeaders = requiredHeaders.filter(
    (header) => !headers.includes(header),
  );
  if (missingHeaders.length > 0) {
    throw new Error("Header Excel wajib memuat fullName, email, dan role.");
  }

  const rows = dataRows
    .map((row) => {
      const mapped = new Map<string, unknown>();
      headers.forEach((header, index) => {
        if (header) mapped.set(header, row[index]);
      });
      return [
        mapped.get("fullName") ?? "",
        mapped.get("email") ?? "",
        mapped.get("role") ?? "",
        mapped.get("classCode") ?? "",
      ];
    })
    .filter((row) => row.some((value) => String(value ?? "").trim() !== ""));

  if (rows.length === 0) {
    throw new Error("Sheet Excel belum memiliki baris data.");
  }

  const csv = toCsv([["fullName", "email", "role", "classCode"], ...rows]);
  return new File([csv], file.name.replace(/\.xlsx$/i, ".csv"), {
    type: "text/csv;charset=utf-8",
  });
}

function resetImportState() {
  importFile.value = null;
  importPreview.value = null;
  importResult.value = null;
  importError.value = "";
}

async function handleImportFileChange(event: Event) {
  const target = event.target as HTMLInputElement;
  const file = target.files?.[0] ?? null;
  importFile.value = file;
  importPreview.value = null;
  importResult.value = null;
  importError.value = "";
  if (!file) return;

  importPreviewLoading.value = true;
  try {
    const previewFile = isExcelFile(file)
      ? await convertXlsxToCsvFile(file)
      : file;
    importPreview.value = await previewSchoolMemberImport(previewFile);
  } catch (error) {
    importError.value = getApiError(error);
  } finally {
    importPreviewLoading.value = false;
  }
}

async function submitImportCommit() {
  if (!importPreview.value || importPreview.value.rows.length === 0) {
    toast.error("Preview import belum tersedia.");
    return;
  }
  if (importPreview.value.invalidCount > 0) {
    toast.error("Perbaiki baris yang tidak valid sebelum import.");
    return;
  }
  if (!importDefaultPassword.value.trim()) {
    toast.error("Password awal wajib diisi.");
    return;
  }

  importCommitLoading.value = true;
  importError.value = "";
  importResult.value = null;
  try {
    importResult.value = await commitSchoolMemberImport({
      defaultPassword: importDefaultPassword.value,
      rows: importPreview.value.rows,
    });
    toast.success("Import warga sekolah selesai.");
    memberSearch.value = "";
    await loadMembers();
  } catch (error) {
    if (
      typeof error === "object" &&
      error !== null &&
      "response" in error &&
      (error as { response?: { data?: AdminSchoolMemberImportCommitResponse } })
        .response?.data?.results
    ) {
      importResult.value = (
        error as { response: { data: AdminSchoolMemberImportCommitResponse } }
      ).response.data;
    }
    importError.value = getApiError(error);
  } finally {
    importCommitLoading.value = false;
  }
}

function resetManualForm() {
  manualForm.value = {
    fullName: "",
    email: "",
    password: "",
    role: "student",
    classCode: "",
  };
}

async function submitManualMember() {
  const payload: AdminSchoolMemberCreatePayload = {
    fullName: manualForm.value.fullName.trim(),
    email: manualForm.value.email.trim(),
    password: manualForm.value.password,
    role: manualForm.value.role,
    classCode:
      manualForm.value.role === "student"
        ? manualForm.value.classCode?.trim() || undefined
        : undefined,
  };
  if (
    !payload.fullName ||
    !payload.email ||
    !payload.password ||
    !payload.role
  ) {
    toast.error("Nama, email, password awal, dan peran wajib diisi.");
    return;
  }

  isCreatingMember.value = true;
  try {
    const createdMember = await createAdminSchoolMember(payload);
    toast.success(memberCreateSuccessMessage(createdMember));
    resetManualForm();
    memberSearch.value = "";
    await loadMembers();
  } catch (error) {
    toast.error(getApiError(error));
  } finally {
    isCreatingMember.value = false;
  }
}

async function removeMember(member: AdminSchoolMemberItem) {
  const confirmed = window.confirm(
    "Akun global tidak akan dihapus. Warga ini hanya dikeluarkan dari sekolah aktif. Lanjutkan?",
  );
  if (!confirmed) return;

  removingSchoolUserId.value = member.schoolUserId;
  try {
    await removeAdminSchoolMember(member.schoolUserId);
    toast.success("Warga sekolah berhasil dihapus dari sekolah aktif.");
    await loadMembers();
  } catch (error) {
    toast.error(getApiError(error));
  } finally {
    removingSchoolUserId.value = "";
  }
}

onMounted(async () => {
  if (!currentSchool.value.hasContext) return;
  await loadRoles();
  await loadMembers();
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
            Warga Sekolah
          </h1>
          <p class="mt-2 max-w-3xl text-sm leading-6 text-[#6b7280]">
            Kelola warga pada sekolah aktif dan import data siswa, guru, atau
            admin sekolah dari template CSV.
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

      <div class="grid min-w-0 gap-5 lg:grid-cols-[minmax(0,1fr)_360px]">
        <section
          class="order-2 min-w-0 rounded-2xl border border-[#ebe7df] bg-white lg:order-1"
        >
          <div
            class="flex flex-col gap-4 border-b border-[#ebe7df] p-5"
          >
            <div
              class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between"
            >
              <div>
                <p
                  class="eyebrow-muted"
                >
                  Warga sekolah aktif
                </p>
                <h2 class="mt-1 text-base font-semibold text-[#171322]">
                  Daftar pengguna sekolah
                </h2>
                <p class="mt-1 text-sm text-[#6b7280]">
                  Kelola peran utama setiap pengguna pada sekolah aktif.
                </p>
              </div>
              <span
                class="inline-flex shrink-0 items-center gap-2 self-start rounded-lg bg-[#eef2ff] px-3 py-2 text-xs font-medium text-[#4f46e5]"
              >
                <PhUsers :size="16" weight="duotone" />
                {{ members.length }} warga
              </span>
            </div>

            <div class="relative min-w-0">
              <PhMagnifyingGlass
                :size="17"
                class="pointer-events-none absolute left-3.5 top-1/2 -translate-y-1/2 text-[#9ca3af]"
              />
              <input
                v-model="memberSearch"
                type="search"
                placeholder="Cari nama atau email warga sekolah"
                class="w-full rounded-lg border border-[#ebe7df] bg-[#fbfaf8] py-2.5 pl-10 pr-3.5 text-sm text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#4f46e5] focus:bg-white"
              />
            </div>
          </div>

          <div class="p-5">
            <div v-if="rolesLoading || membersLoading" class="space-y-3">
              <div
                v-for="item in 3"
                :key="item"
                class="h-28 animate-pulse rounded-lg bg-[#fbfaf8]"
              />
            </div>

            <div
              v-else-if="rolesError || membersError"
              class="rounded-lg border border-[#fecaca] bg-[#fef2f2] p-5 text-center"
            >
              <PhWarningCircle
                :size="26"
                class="mx-auto text-[#dc2626]"
                weight="duotone"
              />
              <h3 class="mt-3 text-sm font-semibold text-[#171322]">
                Warga sekolah belum bisa dimuat
              </h3>
              <p class="mt-2 text-sm leading-6 text-[#6b7280]">
                {{ rolesError || membersError }}
              </p>
              <button
                type="button"
                class="mt-4 inline-flex items-center justify-center gap-2 rounded-lg border border-[#ebe7df] bg-white px-4 py-2.5 text-sm font-medium text-[#374151] transition hover:border-[#4f46e5] hover:text-[#4f46e5] disabled:cursor-not-allowed disabled:opacity-60"
                @click="rolesError ? loadRoles() : loadMembers()"
              >
                Coba lagi
              </button>
            </div>

            <div
              v-else-if="members.length === 0"
              class="rounded-lg bg-[#fbfaf8] px-5 py-8 text-center"
            >
              <PhUsers
                class="mx-auto h-7 w-7 text-[#9ca3af]"
                weight="duotone"
              />
              <h3 class="mt-3 text-sm font-semibold text-[#171322]">
                Belum ada warga sekolah
              </h3>
              <p class="mt-2 text-sm leading-6 text-[#6b7280]">
                Import warga sekolah dari panel kanan untuk memulai.
              </p>
            </div>

            <div v-else class="divide-y divide-[#ebe7df]">
              <article
                v-for="member in members"
                :key="member.schoolUserId"
                class="min-w-0 py-4 first:pt-0 last:pb-0"
              >
                <div
                  class="grid min-w-0 gap-4 xl:grid-cols-[minmax(0,1fr)_280px]"
                >
                  <div class="flex min-w-0 items-start gap-3">
                    <div
                      class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-[#ea580c] text-xs font-semibold text-white"
                    >
                      {{ (member.fullName || "W").charAt(0).toUpperCase() }}
                    </div>
                    <div class="min-w-0">
                      <div class="flex min-w-0 flex-wrap items-center gap-2">
                        <h3
                          class="min-w-0 wrap-break-word text-sm font-semibold text-[#171322]"
                        >
                          {{ member.fullName || "Nama tidak tersedia" }}
                        </h3>
                        <span
                          v-if="primaryRoleName(member)"
                          class="rounded-lg bg-[#eef2ff] px-2 py-1 text-[11px] font-medium text-[#4f46e5]"
                        >
                          {{ roleLabel(primaryRoleName(member)) }}
                        </span>
                      </div>
                      <p class="mt-1 break-all text-xs text-[#6b7280]">
                        {{ member.email || "Email tidak tersedia" }}
                      </p>
                      <p class="mt-2 text-[11px] text-[#9ca3af]">
                        Bergabung {{ formatDateTime(member.createdAt) }}
                      </p>
                      <p
                        v-if="member.classCodes?.length"
                        class="mt-2 text-[11px] font-medium text-[#6b7280]"
                      >
                        Kelas: {{ member.classCodes.join(", ") }}
                      </p>
                      <p
                        v-if="hasMultipleAllowedRoles(member)"
                        class="mt-2 rounded-lg border border-[#fde68a] bg-[#fff7ed] px-3 py-2 text-xs leading-5 text-[#92400e]"
                      >
                        Data lama memiliki lebih dari satu peran. Saat
                        diperbarui, satu peran utama akan disimpan.
                      </p>
                    </div>
                  </div>

                  <div class="min-w-0 rounded-lg bg-[#fbfaf8] p-3">
                    <label class="block text-xs font-medium text-[#6b7280]">
                      Peran sekolah
                      <select
                        class="mt-2 w-full rounded-lg border border-[#ebe7df] bg-white px-3 py-2.5 text-sm text-[#171322] outline-none transition focus:border-[#4f46e5]"
                        :value="memberRoleDrafts[member.schoolUserId] ?? ''"
                        :disabled="rolesLoading || allowedRoles.length === 0"
                        @change="
                          setRoleDraft(
                            member.schoolUserId,
                            ($event.target as HTMLSelectElement).value,
                          )
                        "
                      >
                        <option value="" disabled>Pilih peran</option>
                        <option
                          v-for="role in allowedRoles"
                          :key="role.roleId"
                          :value="role.roleId"
                        >
                          {{ roleLabel(role.roleName) }}
                        </option>
                      </select>
                    </label>
                    <button
                      type="button"
                      class="mt-2.5 inline-flex w-full items-center justify-center gap-2 rounded-lg bg-[#ea580c] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#c2410c] disabled:cursor-not-allowed disabled:opacity-60"
                      :disabled="
                        savingRolesSchoolUserId === member.schoolUserId ||
                        !memberRoleDrafts[member.schoolUserId]
                      "
                      @click="
                        syncRoleForMember(
                          member.schoolUserId,
                          memberRoleDrafts[member.schoolUserId] ?? '',
                        )
                      "
                    >
                      <PhShieldCheck :size="17" weight="duotone" />
                      {{
                        savingRolesSchoolUserId === member.schoolUserId
                          ? "Menyimpan..."
                          : "Simpan peran"
                      }}
                    </button>
                    <button
                      type="button"
                      class="mt-2 inline-flex w-full items-center justify-center gap-2 rounded-lg border border-[#fecaca] bg-white px-3 py-2 text-sm font-medium text-[#dc2626] transition hover:bg-[#fef2f2] disabled:cursor-not-allowed disabled:opacity-60"
                      :disabled="removingSchoolUserId === member.schoolUserId"
                      @click="removeMember(member)"
                    >
                      <PhTrash :size="17" weight="duotone" />
                      {{
                        removingSchoolUserId === member.schoolUserId
                          ? "Menghapus..."
                          : "Hapus dari sekolah"
                      }}
                    </button>
                  </div>
                </div>
              </article>
            </div>
          </div>
        </section>

        <aside class="order-1 min-w-0 lg:order-2">
          <section
            class="rounded-2xl border border-[#ebe7df] bg-white p-5 lg:sticky lg:top-6"
          >
            <div class="flex items-start justify-between gap-3">
              <div>
                <p
                  class="eyebrow-muted"
                >
                  Tambah warga sekolah
                </p>
                <h2 class="mt-1 text-base font-semibold text-[#171322]">
                  Undang atau buat akun
                </h2>
                <p class="mt-1 text-xs leading-5 text-[#6b7280]">
                  Undangan email menjadi alur utama agar guru dan siswa membuat
                  password sendiri. Pembuatan akun langsung tetap tersedia
                  sebagai fallback.
                </p>
              </div>
              <PhPlusCircle
                :size="21"
                class="text-[#ea580c]"
                weight="duotone"
              />
            </div>

            <div
              class="mt-5 grid rounded-lg border border-[#ebe7df] bg-[#fbfaf8] p-1 text-xs font-medium text-[#6b7280] sm:grid-cols-2"
              role="tablist"
              aria-label="Mode tambah warga sekolah"
            >
              <button
                type="button"
                class="rounded-md px-3 py-2 transition"
                :class="
                  memberEntryMode === 'invite'
                    ? 'bg-white text-[#171322] shadow-sm'
                    : 'hover:text-[#171322]'
                "
                :aria-selected="memberEntryMode === 'invite'"
                role="tab"
                @click="setMemberEntryMode('invite')"
              >
                Undang via Email
              </button>
              <button
                type="button"
                class="rounded-md px-3 py-2 transition"
                :class="
                  memberEntryMode === 'direct'
                    ? 'bg-white text-[#171322] shadow-sm'
                    : 'hover:text-[#171322]'
                "
                :aria-selected="memberEntryMode === 'direct'"
                role="tab"
                @click="setMemberEntryMode('direct')"
              >
                Buat Akun Langsung
              </button>
            </div>

            <form
              v-if="memberEntryMode === 'invite'"
              class="mt-5 space-y-3"
              @submit.prevent="submitInviteMember"
            >
              <label class="block text-xs font-medium text-[#6b7280]">
                Nama lengkap
                <input
                  v-model="inviteForm.fullName"
                  type="text"
                  placeholder="Nama guru atau siswa"
                  class="mt-2 w-full rounded-lg border border-[#ebe7df] bg-[#fbfaf8] px-3.5 py-2.5 text-sm text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#4f46e5] focus:bg-white"
                />
              </label>
              <label class="block text-xs font-medium text-[#6b7280]">
                Email
                <input
                  v-model="inviteForm.email"
                  type="email"
                  placeholder="email@sekolah.sch.id"
                  class="mt-2 w-full rounded-lg border border-[#ebe7df] bg-[#fbfaf8] px-3.5 py-2.5 text-sm text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#4f46e5] focus:bg-white"
                />
              </label>
              <label class="block text-xs font-medium text-[#6b7280]">
                Peran
                <select
                  v-model="inviteForm.role"
                  class="mt-2 w-full rounded-lg border border-[#ebe7df] bg-[#fbfaf8] px-3.5 py-2.5 text-sm text-[#171322] outline-none transition focus:border-[#4f46e5] focus:bg-white"
                >
                  <option value="student">Siswa</option>
                  <option value="teacher">Guru</option>
                </select>
              </label>
              <label class="block text-xs font-medium text-[#6b7280]">
                Kode kelas
                <input
                  v-model="inviteForm.classCode"
                  type="text"
                  placeholder="Wajib untuk siswa"
                  class="mt-2 w-full rounded-lg border border-[#ebe7df] bg-[#fbfaf8] px-3.5 py-2.5 text-sm text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#4f46e5] focus:bg-white disabled:cursor-not-allowed disabled:opacity-60"
                  :disabled="inviteForm.role !== 'student'"
                />
              </label>
              <p
                class="rounded-lg border border-[#dbeafe] bg-[#eff6ff] px-3 py-2 text-xs leading-5 text-[#1d4ed8]"
              >
                Pengguna akan menerima email undangan dan membuat password
                sendiri. Jika email tidak terkirim, gunakan link manual setelah
                undangan dibuat.
              </p>
              <button
                type="submit"
                class="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-[#ea580c] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#c2410c] disabled:cursor-not-allowed disabled:opacity-60"
                :disabled="isInvitingMember"
              >
                <PhEnvelopeSimple :size="17" weight="duotone" />
                {{
                  isInvitingMember ? "Mengirim undangan..." : "Kirim undangan"
                }}
              </button>

              <div
                v-if="inviteResult"
                class="rounded-lg border border-[#bbf7d0] bg-[#f0fdf4] p-3 text-xs leading-5 text-[#166534]"
              >
                <p class="font-semibold">Undangan berhasil dibuat.</p>
                <p class="mt-1">
                  Paste kode undangan di browser untuk membuat akun baru, atau
                  salin link undangan di bawah.
                </p>
                <div
                  v-if="inviteLink"
                  class="mt-3 rounded-lg border border-[#dcfce7] bg-white p-2 text-[#374151]"
                >
                  <p class="break-all text-[11px]">{{ inviteLink }}</p>
                  <button
                    type="button"
                    class="mt-2 inline-flex w-full items-center justify-center gap-2 rounded-lg border border-[#ebe7df] bg-white px-4 py-2.5 text-sm font-medium text-[#374151] transition hover:border-[#4f46e5] hover:text-[#4f46e5] disabled:cursor-not-allowed disabled:opacity-60"
                    @click="copyInviteLink"
                  >
                    <PhCopy :size="15" weight="duotone" />
                    Salin link
                  </button>
                </div>
              </div>
            </form>

            <form
              v-else
              class="mt-5 space-y-3"
              @submit.prevent="submitManualMember"
            >
              <label class="block text-xs font-medium text-[#6b7280]">
                Nama lengkap
                <input
                  v-model="manualForm.fullName"
                  type="text"
                  placeholder="Nama warga sekolah"
                  class="mt-2 w-full rounded-lg border border-[#ebe7df] bg-[#fbfaf8] px-3.5 py-2.5 text-sm text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#4f46e5] focus:bg-white"
                />
              </label>
              <label class="block text-xs font-medium text-[#6b7280]">
                Email
                <input
                  v-model="manualForm.email"
                  type="email"
                  placeholder="email@sekolah.sch.id"
                  class="mt-2 w-full rounded-lg border border-[#ebe7df] bg-[#fbfaf8] px-3.5 py-2.5 text-sm text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#4f46e5] focus:bg-white"
                />
              </label>
              <label class="block text-xs font-medium text-[#6b7280]">
                Password awal
                <input
                  v-model="manualForm.password"
                  type="password"
                  placeholder="Minimal 6 karakter"
                  class="mt-2 w-full rounded-lg border border-[#ebe7df] bg-[#fbfaf8] px-3.5 py-2.5 text-sm text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#4f46e5] focus:bg-white"
                />
              </label>
              <label class="block text-xs font-medium text-[#6b7280]">
                Peran
                <select
                  v-model="manualForm.role"
                  class="mt-2 w-full rounded-lg border border-[#ebe7df] bg-[#fbfaf8] px-3.5 py-2.5 text-sm text-[#171322] outline-none transition focus:border-[#4f46e5] focus:bg-white"
                >
                  <option value="student">Siswa</option>
                  <option value="teacher">Guru</option>
                  <option value="admin">Admin sekolah</option>
                </select>
              </label>
              <label class="block text-xs font-medium text-[#6b7280]">
                Kode kelas
                <input
                  v-model="manualForm.classCode"
                  type="text"
                  placeholder="Opsional untuk siswa"
                  class="mt-2 w-full rounded-lg border border-[#ebe7df] bg-[#fbfaf8] px-3.5 py-2.5 text-sm text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#4f46e5] focus:bg-white disabled:cursor-not-allowed disabled:opacity-60"
                  :disabled="manualForm.role !== 'student'"
                />
              </label>
              <p
                class="rounded-lg border border-[#fed7aa] bg-[#fff7ed] px-3 py-2 text-xs leading-5 text-[#92400e]"
              >
                Password awal hanya dipakai untuk akun baru. Jika email sudah
                terdaftar, password tidak akan diubah dan pengguna login
                memakai password yang sudah ada. Password tidak dikirim melalui
                email.
              </p>
              <button
                type="submit"
                class="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-[#ea580c] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#c2410c] disabled:cursor-not-allowed disabled:opacity-60"
                :disabled="isCreatingMember"
              >
                <PhPlusCircle :size="17" weight="duotone" />
                {{ isCreatingMember ? "Menambahkan..." : "Tambah warga" }}
              </button>
            </form>

            <div class="mt-6 border-t border-[#ebe7df] pt-5">
              <div class="flex items-start gap-3">
                <PhFileCsv
                  :size="21"
                  class="mt-0.5 text-[#ea580c]"
                  weight="duotone"
                />
                <div>
                  <h3 class="text-sm font-semibold text-[#171322]">
                    Import warga sekolah
                  </h3>
                  <p class="mt-1 text-xs leading-5 text-[#6b7280]">
                    Upload template CSV atau Excel untuk menambahkan banyak
                    warga sekaligus.
                  </p>
                </div>
              </div>

              <div class="mt-4 space-y-4">
                <div class="grid gap-2 sm:grid-cols-2">
                  <button
                    type="button"
                    class="inline-flex w-full items-center justify-center gap-2 rounded-lg border border-[#ebe7df] bg-white px-4 py-2.5 text-sm font-medium text-[#374151] transition hover:border-[#4f46e5] hover:text-[#4f46e5] disabled:cursor-not-allowed disabled:opacity-60"
                    @click="downloadTemplate"
                  >
                    <PhDownloadSimple :size="17" weight="duotone" />
                    Template CSV
                  </button>

                  <button
                    type="button"
                    class="inline-flex w-full items-center justify-center gap-2 rounded-lg border border-[#ebe7df] bg-white px-4 py-2.5 text-sm font-medium text-[#374151] transition hover:border-[#4f46e5] hover:text-[#4f46e5] disabled:cursor-not-allowed disabled:opacity-60"
                    @click="downloadExcelTemplate"
                  >
                    <PhDownloadSimple :size="17" weight="duotone" />
                    Template Excel
                  </button>
                </div>

                <label class="block text-xs font-medium text-[#6b7280]">
                  File CSV atau Excel
                  <input
                    type="file"
                    accept=".csv,text/csv,.xlsx,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
                    class="mt-2 w-full rounded-lg border border-[#ebe7df] bg-[#fbfaf8] px-3.5 py-2.5 text-sm text-[#171322] outline-none transition file:mr-3 file:rounded-md file:border-0 file:bg-[#fff4ee] file:px-3 file:py-1.5 file:text-xs file:font-semibold file:text-[#ea580c] focus:border-[#4f46e5] focus:bg-white"
                    @change="handleImportFileChange"
                  />
                </label>

                <label class="block text-xs font-medium text-[#6b7280]">
                  Password awal untuk akun baru
                  <input
                    v-model="importDefaultPassword"
                    type="password"
                    placeholder="Minimal 6 karakter"
                    class="mt-2 w-full rounded-lg border border-[#ebe7df] bg-[#fbfaf8] px-3.5 py-2.5 text-sm text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#4f46e5] focus:bg-white"
                  />
                </label>

                <p
                  class="rounded-lg border border-[#fed7aa] bg-[#fff7ed] px-3 py-2 text-xs leading-5 text-[#92400e]"
                >
                  Password awal hanya dipakai untuk akun baru. Jika email sudah
                  terdaftar, password tidak akan diubah dan pengguna login
                  memakai password yang sudah ada. Password tidak dikirim
                  melalui email.
                </p>

                <button
                  type="button"
                  class="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-[#ea580c] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#c2410c] disabled:cursor-not-allowed disabled:opacity-60"
                  :disabled="
                    importCommitLoading ||
                    importPreviewLoading ||
                    !importPreview ||
                    importPreview.invalidCount > 0
                  "
                  @click="submitImportCommit"
                >
                  <PhUploadSimple :size="17" weight="duotone" />
                  {{ importCommitLoading ? "Mengimport..." : "Import warga" }}
                </button>

                <button
                  v-if="importPreview || importResult || importFile"
                  type="button"
                  class="inline-flex w-full items-center justify-center gap-2 rounded-lg border border-[#ebe7df] bg-white px-4 py-2.5 text-sm font-medium text-[#374151] transition hover:border-[#4f46e5] hover:text-[#4f46e5] disabled:cursor-not-allowed disabled:opacity-60"
                  :disabled="importCommitLoading || importPreviewLoading"
                  @click="resetImportState"
                >
                  Reset import
                </button>
              </div>
            </div>

            <div class="mt-4 space-y-3">
              <p
                v-if="importError"
                class="rounded-lg bg-[#fef2f2] px-3 py-2 text-xs leading-5 text-[#dc2626]"
              >
                {{ importError }}
              </p>
              <p
                v-else-if="importPreviewLoading"
                class="rounded-lg bg-[#fbfaf8] px-3 py-3 text-xs leading-5 text-[#6b7280]"
              >
                Memvalidasi file import...
              </p>
              <p
                v-else-if="!importPreview"
                class="rounded-lg bg-[#fbfaf8] px-3 py-3 text-xs leading-5 text-[#6b7280]"
              >
                Pilih file CSV untuk melihat preview validasi.
              </p>

              <div
                v-if="importPreview"
                class="rounded-lg border border-[#ebe7df] bg-[#fbfaf8] p-3"
              >
                <div class="flex flex-wrap gap-2 text-xs">
                  <span
                    class="rounded-lg bg-[#ecfdf3] px-2.5 py-1 font-semibold text-[#027a48]"
                  >
                    {{ importPreview.validCount }} valid
                  </span>
                  <span
                    class="rounded-lg bg-[#fef2f2] px-2.5 py-1 font-semibold text-[#dc2626]"
                  >
                    {{ importPreview.invalidCount }} invalid
                  </span>
                </div>

                <div class="mt-3 max-h-72 space-y-2 overflow-y-auto pr-1">
                  <article
                    v-for="row in importPreview.rows"
                    :key="row.rowNumber"
                    class="rounded-lg border bg-white p-3"
                    :class="
                      row.status === 'valid'
                        ? 'border-[#bbf7d0]'
                        : 'border-[#fecaca]'
                    "
                  >
                    <div class="flex items-start justify-between gap-2">
                      <div class="min-w-0">
                        <p class="text-xs font-semibold text-[#171322]">
                          Baris {{ row.rowNumber }} ·
                          {{ row.fullName || "Nama kosong" }}
                        </p>
                        <p class="mt-1 break-all text-xs text-[#6b7280]">
                          {{ row.email || "Email kosong" }}
                        </p>
                        <p class="mt-1 text-xs text-[#6b7280]">
                          {{ roleLabel(row.role) }}
                          <span v-if="row.classCode">
                            · {{ row.classCode }}</span
                          >
                        </p>
                      </div>
                      <span
                        class="shrink-0 rounded-full px-2 py-1 text-[10px] font-semibold"
                        :class="
                          row.status === 'valid'
                            ? 'bg-[#ecfdf3] text-[#027a48]'
                            : 'bg-[#fef2f2] text-[#dc2626]'
                        "
                      >
                        {{ row.status === "valid" ? "Valid" : "Invalid" }}
                      </span>
                    </div>
                    <ul
                      v-if="row.errors.length > 0"
                      class="mt-2 list-disc space-y-1 pl-4 text-xs leading-5 text-[#dc2626]"
                    >
                      <li v-for="error in row.errors" :key="error">
                        {{ error }}
                      </li>
                    </ul>
                  </article>
                </div>
              </div>

              <div
                v-if="importResult"
                class="rounded-lg border border-[#bbf7d0] bg-[#f0fdf4] p-3 text-xs leading-5 text-[#166534]"
              >
                <p class="font-semibold">Import selesai</p>
                <p class="mt-1">
                  {{ importResult.importedCount }} diproses,
                  {{ importResult.skippedCount }} dilewati,
                  {{ importResult.failedCount }} gagal.
                </p>
                <div
                  v-if="importResult.results.length > 0"
                  class="mt-3 max-h-64 space-y-2 overflow-y-auto pr-1"
                >
                  <article
                    v-for="result in importResult.results"
                    :key="`${result.rowNumber}-${result.email}`"
                    class="rounded-md border border-[#bbf7d0] bg-white/70 px-3 py-2"
                  >
                    <div class="flex items-start justify-between gap-3">
                      <div class="min-w-0">
                        <p class="truncate font-semibold text-[#14532d]">
                          {{ result.fullName || result.email }}
                        </p>
                        <p class="mt-0.5 break-all text-[#166534]">
                          {{ result.email }}
                        </p>
                      </div>
                      <span
                        class="shrink-0 rounded-full px-2 py-1 text-[10px] font-semibold"
                        :class="
                          result.status === 'imported'
                            ? 'bg-[#dcfce7] text-[#166534]'
                            : result.status === 'skipped'
                              ? 'bg-[#fef3c7] text-[#92400e]'
                              : 'bg-[#fee2e2] text-[#b91c1c]'
                        "
                      >
                        {{
                          result.status === "imported"
                            ? "Diproses"
                            : result.status === "skipped"
                              ? "Dilewati"
                              : "Gagal"
                        }}
                      </span>
                    </div>
                    <p class="mt-2 text-[#166534]">
                      {{ importResultNote(result) }}
                    </p>
                    <p
                      v-if="importEmailNote(result.emailNotification)"
                      class="mt-1 text-[#15803d]"
                    >
                      {{ importEmailNote(result.emailNotification) }}
                    </p>
                  </article>
                </div>
              </div>
            </div>
          </section>
        </aside>
      </div>
    </section>
  </main>
</template>
