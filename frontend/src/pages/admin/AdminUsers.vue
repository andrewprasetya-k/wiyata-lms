<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import {
  PhCopy,
  PhDownloadSimple,
  PhEnvelopeSimple,
  PhEye,
  PhEyeSlash,
  PhFileCsv,
  PhMagnifyingGlass,
  PhPlusCircle,
  PhShieldCheck,
  PhTrash,
  PhUploadSimple,
  PhUsers,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import { useToastStore } from "../../stores/toast";
import { useConfirmStore } from "../../stores/confirm";
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
import {
  convertXlsxToCsvFile,
  downloadExcelTemplate,
  downloadTemplate,
  isExcelFile,
} from "../../utils/schoolMemberImportFile";
import { usePasswordVisibility } from "../../composables/usePasswordVisibility";
import PaginationBar from "../../components/common/PaginationBar.vue";
import InlineFormError from "../../components/common/InlineFormError.vue";

const allowedRoleNames = ["student", "teacher", "admin"];
const auth = useAuthStore();
const toast = useToastStore();
const confirm = useConfirmStore();

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
const memberRoleDrafts = ref<Record<string, string[]>>({});
const memberRoleErrors = ref<Record<string, string>>({});

const membersPage = ref(1);
const membersTotalPages = ref(1);
const membersTotalItems = ref(0);
const MEMBERS_LIMIT = 20;

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
const memberFormError = ref("");

const memberSearch = ref("");
const memberEntryMode = ref<"invite" | "direct">("invite");
const importFile = ref<File | null>(null);
const importDefaultPassword = ref("");
const {
  visible: importPasswordVisible,
  inputType: importPasswordInputType,
  toggle: toggleImportPasswordVisibility,
} = usePasswordVisibility();
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
const {
  visible: manualPasswordVisible,
  inputType: manualPasswordInputType,
  toggle: toggleManualPasswordVisibility,
} = usePasswordVisibility();

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
  const nextDrafts: Record<string, string[]> = {};

  for (const member of members.value) {
    const assignedRoleIds = (member.roles ?? [])
      .filter((roleName) => roleByName.has(normalizeRoleName(roleName)))
      .map((roleName) => roleByName.get(normalizeRoleName(roleName)) as string);
    // De-dupe in case a member's role list repeats the same role name.
    nextDrafts[member.schoolUserId] = [...new Set(assignedRoleIds)];
  }

  memberRoleDrafts.value = nextDrafts;
  memberRoleErrors.value = {};
}

function assignedRoleNames(member: AdminSchoolMemberItem) {
  return (member.roles ?? [])
    .filter((roleName) => allowedRoleNames.includes(normalizeRoleName(roleName)))
    .sort((a, b) => rolePriority(a) - rolePriority(b));
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

async function loadMembers(targetPage = membersPage.value) {
  if (!currentSchool.value.hasContext) return;

  membersLoading.value = true;
  membersError.value = "";
  try {
    const data = await getAdminSchoolMembers({
      page: targetPage,
      limit: MEMBERS_LIMIT,
      search: memberSearch.value.trim(),
    });
    members.value = data.data ?? [];
    membersPage.value = data.page ?? targetPage;
    membersTotalPages.value = data.totalPages ?? 1;
    membersTotalItems.value = Number(data.totalItems ?? 0);
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

    membersPage.value = 1;
    membersLoading.value = true;
    membersError.value = "";
    try {
      const data = await getAdminSchoolMembers({
        page: 1,
        limit: MEMBERS_LIMIT,
        search: memberSearch.value.trim(),
      });
      if (version !== searchVersion) return;
      members.value = data.data ?? [];
      membersTotalPages.value = data.totalPages ?? 1;
      membersTotalItems.value = Number(data.totalItems ?? 0);
      initializeRoleDrafts();
    } catch {
      if (version !== searchVersion) return;
      membersError.value = "Warga sekolah belum bisa dimuat.";
    } finally {
      if (version === searchVersion) membersLoading.value = false;
    }
  }, 300);
});

async function syncRoleForMember(schoolUserId: string, roleIds: string[]) {
  if (roleIds.length === 0) {
    memberRoleErrors.value = {
      ...memberRoleErrors.value,
      [schoolUserId]: "Pilih minimal satu peran.",
    };
    return;
  }
  memberRoleErrors.value = { ...memberRoleErrors.value, [schoolUserId]: "" };

  savingRolesSchoolUserId.value = schoolUserId;
  try {
    await syncUserRoles(schoolUserId, { roleIds });
    toast.success("Peran warga sekolah berhasil diperbarui.");
    await loadMembers();
  } catch {
    toast.error("Peran warga sekolah belum bisa diperbarui.");
  } finally {
    savingRolesSchoolUserId.value = "";
  }
}

function toggleRoleDraft(schoolUserId: string, roleId: string) {
  const current = memberRoleDrafts.value[schoolUserId] ?? [];
  const next = current.includes(roleId)
    ? current.filter((id) => id !== roleId)
    : [...current, roleId];

  memberRoleDrafts.value = {
    ...memberRoleDrafts.value,
    [schoolUserId]: next,
  };
  if (next.length > 0 && memberRoleErrors.value[schoolUserId]) {
    memberRoleErrors.value = { ...memberRoleErrors.value, [schoolUserId]: "" };
  }
}


const inviteLink = computed(() => {
  const acceptUrl = inviteResult.value?.acceptUrl;
  if (!acceptUrl) return "";
  if (/^https?:\/\//i.test(acceptUrl)) return acceptUrl;
  return `${window.location.origin}${acceptUrl.startsWith("/") ? "" : "/"}${acceptUrl}`;
});

function setMemberEntryMode(mode: "invite" | "direct") {
  memberEntryMode.value = mode;
  memberFormError.value = "";
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

  memberFormError.value = "";
  if (!payload.fullName || !payload.email || !payload.role) {
    memberFormError.value = "Nama, email, dan peran wajib diisi.";
    return;
  }
  if (payload.role === "student" && !payload.classCode) {
    memberFormError.value = "Kode kelas wajib diisi untuk undangan siswa.";
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
  importError.value = "";
  if (!importPreview.value || importPreview.value.rows.length === 0) {
    importError.value = "Preview import belum tersedia.";
    return;
  }
  if (importPreview.value.invalidCount > 0) {
    importError.value = "Perbaiki baris yang tidak valid sebelum import.";
    return;
  }
  if (!importDefaultPassword.value.trim()) {
    importError.value = "Password awal wajib diisi.";
    return;
  }

  importCommitLoading.value = true;
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
  memberFormError.value = "";
  if (
    !payload.fullName ||
    !payload.email ||
    !payload.password ||
    !payload.role
  ) {
    memberFormError.value = "Nama, email, password awal, dan peran wajib diisi.";
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
  const ok = await confirm.confirm({
    title: "Keluarkan warga sekolah?",
    description:
      "Akun global tidak akan dihapus. Warga ini hanya dikeluarkan dari sekolah aktif.",
    confirmLabel: "Keluarkan",
    variant: "danger",
  });
  if (!ok) return;

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
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-background">
    <header class="border-b border-border bg-surface">
      <div
        class="flex min-w-0 flex-col gap-3 px-5 py-5 sm:px-6 lg:flex-row lg:items-end lg:justify-between lg:px-8"
      >
        <div class="min-w-0">
          <h1 class="mt-1 text-2xl font-semibold text-foreground sm:text-3xl">
            Warga Sekolah
          </h1>
          <p class="mt-2 max-w-3xl text-sm leading-6 text-muted">
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
            class="rounded-lg bg-surface-strong px-3 py-2 font-medium text-muted"
          >
            {{ currentSchool.schoolCode || "Kode belum tersedia" }}
          </span>
        </div>
      </div>
    </header>

    <section class="px-5 py-5 sm:px-6 lg:px-8 lg:py-6">
      <div
        v-if="!currentSchool.hasContext"
        class="mb-5 flex items-start gap-3 rounded-xl border border-danger-line bg-danger-soft p-4 text-sm leading-6 text-danger"
      >
        <PhWarningCircle :size="20" class="mt-0.5 shrink-0" weight="duotone" />
        <p>
          Konteks sekolah aktif belum tersedia. Pastikan akun admin terhubung
          dengan sekolah.
        </p>
      </div>

      <div class="grid min-w-0 gap-5 lg:grid-cols-[minmax(0,1fr)_360px]">
        <section
          class="order-2 min-w-0 rounded-xl border border-border bg-surface shadow-sm lg:order-1"
        >
          <div
            class="flex flex-col gap-4 border-b border-border p-5"
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
                <h2 class="mt-1 text-base font-semibold text-foreground">
                  Daftar pengguna sekolah
                </h2>
                <p class="mt-1 text-sm text-muted">
                  Kelola peran utama setiap pengguna pada sekolah aktif.
                </p>
              </div>
              <span
                class="inline-flex shrink-0 items-center gap-2 self-start rounded-lg bg-brand-soft px-3 py-2 text-xs font-medium text-brand"
              >
                <PhUsers :size="16" weight="duotone" />
                {{ membersTotalItems || members.length }} warga
              </span>
            </div>

            <div class="relative min-w-0">
              <PhMagnifyingGlass
                :size="17"
                class="pointer-events-none absolute left-3.5 top-1/2 -translate-y-1/2 text-muted"
              />
              <input
                v-model="memberSearch"
                type="search"
                placeholder="Cari nama atau email warga sekolah"
                class="w-full rounded-lg border border-border bg-surface-subtle py-2.5 pl-10 pr-3.5 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-brand focus:bg-surface"
              />
            </div>
          </div>

          <div class="p-5">
            <div v-if="rolesLoading || membersLoading" class="space-y-3">
              <div
                v-for="item in 3"
                :key="item"
                class="h-28 animate-pulse rounded-lg bg-surface-subtle"
              />
            </div>

            <div
              v-else-if="rolesError || membersError"
              class="rounded-lg border border-danger-line bg-danger-soft p-5 text-center"
            >
              <PhWarningCircle
                :size="26"
                class="mx-auto text-danger"
                weight="duotone"
              />
              <h3 class="mt-3 text-sm font-semibold text-foreground">
                Warga sekolah belum bisa dimuat
              </h3>
              <p class="mt-2 text-sm leading-6 text-muted">
                {{ rolesError || membersError }}
              </p>
              <button
                type="button"
                class="mt-4 inline-flex items-center justify-center gap-2 rounded-lg border border-border bg-surface px-4 py-2.5 text-sm font-medium text-foreground-secondary transition hover:border-brand hover:text-brand disabled:cursor-not-allowed disabled:opacity-60"
                @click="rolesError ? loadRoles() : loadMembers()"
              >
                Coba lagi
              </button>
            </div>

            <div
              v-else-if="members.length === 0"
              class="rounded-lg bg-surface-subtle px-5 py-8 text-center"
            >
              <PhUsers
                class="mx-auto h-7 w-7 text-muted"
                weight="duotone"
              />
              <h3 class="mt-3 text-sm font-semibold text-foreground">
                Belum ada warga sekolah
              </h3>
              <p class="mt-2 text-sm leading-6 text-muted">
                Import warga sekolah dari panel kanan untuk memulai.
              </p>
            </div>

            <div v-else class="flex flex-col gap-4">
              <div class="divide-y divide-border">
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
                          class="min-w-0 wrap-break-word text-sm font-semibold text-foreground"
                        >
                          {{ member.fullName || "Nama tidak tersedia" }}
                        </h3>
                        <span
                          v-for="roleName in assignedRoleNames(member)"
                          :key="roleName"
                          class="rounded-lg bg-brand-soft px-2 py-1 text-[11px] font-medium text-brand"
                        >
                          {{ roleLabel(roleName) }}
                        </span>
                      </div>
                      <p class="mt-1 break-all text-xs text-muted">
                        {{ member.email || "Email tidak tersedia" }}
                      </p>
                      <p class="mt-2 text-[11px] text-muted">
                        Bergabung {{ formatDateTime(member.createdAt) }}
                      </p>
                      <p
                        v-if="member.classCodes?.length"
                        class="mt-2 text-[11px] font-medium text-muted"
                      >
                        Kelas: {{ member.classCodes.join(", ") }}
                      </p>
                    </div>
                  </div>

                  <div class="min-w-0 rounded-lg bg-surface-subtle p-3">
                    <fieldset class="m-0 min-w-0 border-0 p-0">
                      <legend class="block text-xs font-medium text-muted">
                        Peran sekolah
                      </legend>
                      <div class="mt-2 space-y-2">
                        <label
                          v-for="role in allowedRoles"
                          :key="role.roleId"
                          class="flex items-center gap-2 rounded-lg border border-border bg-surface px-3 py-2 text-sm text-foreground transition"
                          :class="
                            rolesLoading
                              ? 'cursor-not-allowed opacity-60'
                              : 'cursor-pointer hover:border-brand'
                          "
                        >
                          <input
                            type="checkbox"
                            class="h-4 w-4 shrink-0 rounded border-border-strong text-brand focus:ring-brand"
                            :checked="
                              memberRoleDrafts[member.schoolUserId]?.includes(
                                role.roleId,
                              ) ?? false
                            "
                            :disabled="rolesLoading || allowedRoles.length === 0"
                            @change="toggleRoleDraft(member.schoolUserId, role.roleId)"
                          />
                          {{ roleLabel(role.roleName) }}
                        </label>
                      </div>
                    </fieldset>
                    <InlineFormError
                      class="mt-2"
                      :message="memberRoleErrors[member.schoolUserId] ?? ''"
                    />
                    <button
                      type="button"
                      class="mt-2.5 inline-flex w-full items-center justify-center gap-2 rounded-lg bg-[#ea580c] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#c2410c] disabled:cursor-not-allowed disabled:opacity-60"
                      :disabled="savingRolesSchoolUserId === member.schoolUserId"
                      @click="
                        syncRoleForMember(
                          member.schoolUserId,
                          memberRoleDrafts[member.schoolUserId] ?? [],
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
                      class="mt-2 inline-flex w-full items-center justify-center gap-2 rounded-lg border border-danger-line bg-surface px-3 py-2 text-sm font-medium text-danger transition hover:bg-danger-soft disabled:cursor-not-allowed disabled:opacity-60"
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
              <PaginationBar
                :page="membersPage"
                :total-pages="membersTotalPages"
                :total-items="membersTotalItems"
                :limit="MEMBERS_LIMIT"
                @change="(p) => loadMembers(p)"
              />
            </div>
          </div>
        </section>

        <aside class="order-1 min-w-0 lg:order-2">
          <section
            class="rounded-xl border border-border bg-surface shadow-sm p-5 lg:sticky lg:top-6"
          >
            <div class="flex items-start justify-between gap-3">
              <div>
                <p
                  class="eyebrow-muted"
                >
                  Tambah warga sekolah
                </p>
                <h2 class="mt-1 text-base font-semibold text-foreground">
                  Undang atau buat akun
                </h2>
                <p class="mt-1 text-xs leading-5 text-muted">
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
              class="mt-5 grid rounded-lg bg-surface-subtle p-1 text-xs font-medium text-muted sm:grid-cols-2"
              role="tablist"
              aria-label="Mode tambah warga sekolah"
            >
              <button
                type="button"
                class="rounded-md px-3 py-2 transition"
                :class="
                  memberEntryMode === 'invite'
                    ? 'bg-surface text-foreground shadow-sm'
                    : 'hover:text-foreground'
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
                    ? 'bg-surface text-foreground shadow-sm'
                    : 'hover:text-foreground'
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
              <label class="block text-xs font-medium text-muted">
                Nama lengkap
                <input
                  v-model="inviteForm.fullName"
                  type="text"
                  placeholder="Nama guru atau siswa"
                  class="mt-2 w-full rounded-lg border border-border bg-surface-subtle px-3.5 py-2.5 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-brand focus:bg-surface"
                />
              </label>
              <label class="block text-xs font-medium text-muted">
                Email
                <input
                  v-model="inviteForm.email"
                  type="email"
                  placeholder="email@sekolah.sch.id"
                  class="mt-2 w-full rounded-lg border border-border bg-surface-subtle px-3.5 py-2.5 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-brand focus:bg-surface"
                />
              </label>
              <label class="block text-xs font-medium text-muted">
                Peran
                <select
                  v-model="inviteForm.role"
                  class="mt-2 w-full rounded-lg border border-border bg-surface-subtle px-3.5 py-2.5 text-sm text-foreground outline-none transition focus:border-brand focus:bg-surface"
                >
                  <option value="student">Siswa</option>
                  <option value="teacher">Guru</option>
                </select>
              </label>
              <label class="block text-xs font-medium text-muted">
                Kode kelas
                <input
                  v-model="inviteForm.classCode"
                  type="text"
                  placeholder="Wajib untuk siswa"
                  class="mt-2 w-full rounded-lg border border-border bg-surface-subtle px-3.5 py-2.5 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-brand focus:bg-surface disabled:cursor-not-allowed disabled:opacity-60"
                  :disabled="inviteForm.role !== 'student'"
                />
              </label>
              <p
                class="rounded-lg border border-[#dbeafe] bg-info-soft px-3 py-2 text-xs leading-5 text-info-hover"
              >
                Pengguna akan menerima email undangan dan membuat password
                sendiri. Jika email tidak terkirim, gunakan link manual setelah
                undangan dibuat.
              </p>
              <InlineFormError :message="memberFormError" />
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
                class="rounded-lg border border-success-line bg-success-soft p-3 text-xs leading-5 text-success"
              >
                <p class="font-semibold">Undangan berhasil dibuat.</p>
                <p class="mt-1">
                  Paste kode undangan di browser untuk membuat akun baru, atau
                  salin link undangan di bawah.
                </p>
                <div
                  v-if="inviteLink"
                  class="mt-3 rounded-lg border border-[#dcfce7] bg-surface p-2 text-foreground-secondary"
                >
                  <p class="break-all text-[11px]">{{ inviteLink }}</p>
                  <button
                    type="button"
                    class="mt-2 inline-flex w-full items-center justify-center gap-2 rounded-lg border border-border bg-surface px-4 py-2.5 text-sm font-medium text-foreground-secondary transition hover:border-brand hover:text-brand disabled:cursor-not-allowed disabled:opacity-60"
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
              <label class="block text-xs font-medium text-muted">
                Nama lengkap
                <input
                  v-model="manualForm.fullName"
                  type="text"
                  placeholder="Nama warga sekolah"
                  class="mt-2 w-full rounded-lg border border-border bg-surface-subtle px-3.5 py-2.5 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-brand focus:bg-surface"
                />
              </label>
              <label class="block text-xs font-medium text-muted">
                Email
                <input
                  v-model="manualForm.email"
                  type="email"
                  placeholder="email@sekolah.sch.id"
                  class="mt-2 w-full rounded-lg border border-border bg-surface-subtle px-3.5 py-2.5 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-brand focus:bg-surface"
                />
              </label>
              <label class="block text-xs font-medium text-muted">
                Password awal
                <div class="relative mt-2">
                  <input
                    v-model="manualForm.password"
                    :type="manualPasswordInputType"
                    placeholder="Minimal 6 karakter"
                    class="w-full rounded-lg border border-border bg-surface-subtle px-3.5 py-2.5 pr-11 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-brand focus:bg-surface"
                  />
                  <button
                    type="button"
                    class="absolute right-2.5 top-1/2 -translate-y-1/2 rounded-lg p-1.5 text-muted transition hover:text-foreground"
                    :aria-label="
                      manualPasswordVisible
                        ? 'Sembunyikan password'
                        : 'Tampilkan password'
                    "
                    :aria-pressed="manualPasswordVisible"
                    @click="toggleManualPasswordVisibility"
                  >
                    <PhEyeSlash v-if="manualPasswordVisible" :size="17" />
                    <PhEye v-else :size="17" />
                  </button>
                </div>
              </label>
              <label class="block text-xs font-medium text-muted">
                Peran
                <select
                  v-model="manualForm.role"
                  class="mt-2 w-full rounded-lg border border-border bg-surface-subtle px-3.5 py-2.5 text-sm text-foreground outline-none transition focus:border-brand focus:bg-surface"
                >
                  <option value="student">Siswa</option>
                  <option value="teacher">Guru</option>
                  <option value="admin">Admin sekolah</option>
                </select>
              </label>
              <label class="block text-xs font-medium text-muted">
                Kode kelas
                <input
                  v-model="manualForm.classCode"
                  type="text"
                  placeholder="Opsional untuk siswa"
                  class="mt-2 w-full rounded-lg border border-border bg-surface-subtle px-3.5 py-2.5 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-brand focus:bg-surface disabled:cursor-not-allowed disabled:opacity-60"
                  :disabled="manualForm.role !== 'student'"
                />
              </label>
              <p
                class="rounded-lg border border-warning-line bg-warning-soft px-3 py-2 text-xs leading-5 text-warning-hover"
              >
                Password awal hanya dipakai untuk akun baru. Jika email sudah
                terdaftar, password tidak akan diubah dan pengguna login
                memakai password yang sudah ada. Password tidak dikirim melalui
                email.
              </p>
              <InlineFormError :message="memberFormError" />
              <button
                type="submit"
                class="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-[#ea580c] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#c2410c] disabled:cursor-not-allowed disabled:opacity-60"
                :disabled="isCreatingMember"
              >
                <PhPlusCircle :size="17" weight="duotone" />
                {{ isCreatingMember ? "Menambahkan..." : "Tambah warga" }}
              </button>
            </form>

            <div class="mt-6 border-t border-border pt-5">
              <div class="flex items-start gap-3">
                <PhFileCsv
                  :size="21"
                  class="mt-0.5 text-[#ea580c]"
                  weight="duotone"
                />
                <div>
                  <h3 class="text-sm font-semibold text-foreground">
                    Import warga sekolah
                  </h3>
                  <p class="mt-1 text-xs leading-5 text-muted">
                    Upload template CSV atau Excel untuk menambahkan banyak
                    warga sekaligus.
                  </p>
                </div>
              </div>

              <div class="mt-4 space-y-4">
                <div class="grid gap-2 sm:grid-cols-2">
                  <button
                    type="button"
                    class="inline-flex w-full items-center justify-center gap-2 rounded-lg border border-border bg-surface px-4 py-2.5 text-sm font-medium text-foreground-secondary transition hover:border-brand hover:text-brand disabled:cursor-not-allowed disabled:opacity-60"
                    @click="downloadTemplate"
                  >
                    <PhDownloadSimple :size="17" weight="duotone" />
                    Template CSV
                  </button>

                  <button
                    type="button"
                    class="inline-flex w-full items-center justify-center gap-2 rounded-lg border border-border bg-surface px-4 py-2.5 text-sm font-medium text-foreground-secondary transition hover:border-brand hover:text-brand disabled:cursor-not-allowed disabled:opacity-60"
                    @click="downloadExcelTemplate"
                  >
                    <PhDownloadSimple :size="17" weight="duotone" />
                    Template Excel
                  </button>
                </div>

                <label class="block text-xs font-medium text-muted">
                  File CSV atau Excel
                  <input
                    type="file"
                    accept=".csv,text/csv,.xlsx,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
                    class="mt-2 w-full rounded-lg border border-border bg-surface-subtle px-3.5 py-2.5 text-sm text-foreground outline-none transition file:mr-3 file:rounded-md file:border-0 file:bg-[#fff4ee] file:px-3 file:py-1.5 file:text-xs file:font-semibold file:text-[#ea580c] focus:border-brand focus:bg-surface"
                    @change="handleImportFileChange"
                  />
                </label>

                <label class="block text-xs font-medium text-muted">
                  Password awal untuk akun baru
                  <div class="relative mt-2">
                    <input
                      v-model="importDefaultPassword"
                      :type="importPasswordInputType"
                      placeholder="Minimal 6 karakter"
                      class="w-full rounded-lg border border-border bg-surface-subtle px-3.5 py-2.5 pr-11 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-brand focus:bg-surface"
                    />
                    <button
                      type="button"
                      class="absolute right-2.5 top-1/2 -translate-y-1/2 rounded-lg p-1.5 text-muted transition hover:text-foreground"
                      :aria-label="
                        importPasswordVisible
                          ? 'Sembunyikan password'
                          : 'Tampilkan password'
                      "
                      :aria-pressed="importPasswordVisible"
                      @click="toggleImportPasswordVisibility"
                    >
                      <PhEyeSlash v-if="importPasswordVisible" :size="17" />
                      <PhEye v-else :size="17" />
                    </button>
                  </div>
                </label>

                <p
                  class="rounded-lg border border-warning-line bg-warning-soft px-3 py-2 text-xs leading-5 text-warning-hover"
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
                  class="inline-flex w-full items-center justify-center gap-2 rounded-lg border border-border bg-surface px-4 py-2.5 text-sm font-medium text-foreground-secondary transition hover:border-brand hover:text-brand disabled:cursor-not-allowed disabled:opacity-60"
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
                class="rounded-lg bg-danger-soft px-3 py-2 text-xs leading-5 text-danger"
              >
                {{ importError }}
              </p>
              <p
                v-else-if="importPreviewLoading"
                class="rounded-lg bg-surface-subtle px-3 py-3 text-xs leading-5 text-muted"
              >
                Memvalidasi file import...
              </p>
              <p
                v-else-if="!importPreview"
                class="rounded-lg bg-surface-subtle px-3 py-3 text-xs leading-5 text-muted"
              >
                Pilih file CSV untuk melihat preview validasi.
              </p>

              <div
                v-if="importPreview"
                class="rounded-lg bg-surface-subtle p-3"
              >
                <div class="flex flex-wrap gap-2 text-xs">
                  <span
                    class="rounded-lg bg-success-soft px-2.5 py-1 font-semibold text-success"
                  >
                    {{ importPreview.validCount }} valid
                  </span>
                  <span
                    class="rounded-lg bg-danger-soft px-2.5 py-1 font-semibold text-danger"
                  >
                    {{ importPreview.invalidCount }} invalid
                  </span>
                </div>

                <div class="mt-3 max-h-72 space-y-2 overflow-y-auto pr-1">
                  <article
                    v-for="row in importPreview.rows"
                    :key="row.rowNumber"
                    class="rounded-lg border bg-surface p-3"
                    :class="
                      row.status === 'valid'
                        ? 'border-success-line'
                        : 'border-danger-line'
                    "
                  >
                    <div class="flex items-start justify-between gap-2">
                      <div class="min-w-0">
                        <p class="text-xs font-semibold text-foreground">
                          Baris {{ row.rowNumber }} ·
                          {{ row.fullName || "Nama kosong" }}
                        </p>
                        <p class="mt-1 break-all text-xs text-muted">
                          {{ row.email || "Email kosong" }}
                        </p>
                        <p class="mt-1 text-xs text-muted">
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
                            ? 'bg-success-soft text-success'
                            : 'bg-danger-soft text-danger'
                        "
                      >
                        {{ row.status === "valid" ? "Valid" : "Invalid" }}
                      </span>
                    </div>
                    <ul
                      v-if="row.errors.length > 0"
                      class="mt-2 list-disc space-y-1 pl-4 text-xs leading-5 text-danger"
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
                class="rounded-lg border border-success-line bg-success-soft p-3 text-xs leading-5 text-success"
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
                    class="rounded-md border border-success-line bg-surface/70 px-3 py-2"
                  >
                    <div class="flex items-start justify-between gap-3">
                      <div class="min-w-0">
                        <p class="truncate font-semibold text-[#14532d]">
                          {{ result.fullName || result.email }}
                        </p>
                        <p class="mt-0.5 break-all text-success">
                          {{ result.email }}
                        </p>
                      </div>
                      <span
                        class="shrink-0 rounded-full px-2 py-1 text-[10px] font-semibold"
                        :class="
                          result.status === 'imported'
                            ? 'bg-[#dcfce7] text-success'
                            : result.status === 'skipped'
                              ? 'bg-[#fef3c7] text-warning-hover'
                              : 'bg-[#fee2e2] text-danger-hover'
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
                    <p class="mt-2 text-success">
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
