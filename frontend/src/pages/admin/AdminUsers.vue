<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import {
  PhMagnifyingGlass,
  PhShieldCheck,
  PhUsers,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import {
  enrollUserToSchool,
  getAdminUsers,
  getRoles,
  getSchoolMembers,
  syncUserRoles,
} from "../../services/adminUser";
import type {
  AdminUserItem,
  RoleItem,
  SchoolMemberItem,
} from "../../types/adminUser";
import { formatDateTime } from "../../utils/date";

const allowedRoleNames = ["student", "teacher", "admin"];
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

const members = ref<SchoolMemberItem[]>([]);
const roles = ref<RoleItem[]>([]);
const userResults = ref<AdminUserItem[]>([]);
const memberRoleDrafts = ref<Record<string, string>>({});

const membersLoading = ref(false);
const rolesLoading = ref(false);
const userSearchLoading = ref(false);
const addExistingLoadingUserId = ref("");
const savingRolesSchoolUserId = ref("");

const membersError = ref("");
const rolesError = ref("");
const userSearchError = ref("");
const actionError = ref("");
const actionMessage = ref("");

const memberSearch = ref("");
const userSearch = ref("");
const existingRoleId = ref("");

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
  if (normalized === "student") return "Student";
  if (normalized === "teacher") return "Teacher";
  if (normalized === "admin") return "Admin";
  return roleName;
}

function rolePriority(roleName: string) {
  const normalized = normalizeRoleName(roleName);
  if (normalized === "admin") return 0;
  if (normalized === "teacher") return 1;
  if (normalized === "student") return 2;
  return 99;
}

function isMember(userId: string) {
  return members.value.some((member) => member.userId === userId);
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

function primaryRoleName(member: SchoolMemberItem) {
  return (
    member.roles
      ?.filter((roleName) =>
        allowedRoleNames.includes(normalizeRoleName(roleName)),
      )
      .sort((a, b) => rolePriority(a) - rolePriority(b))[0] ?? ""
  );
}

function hasMultipleAllowedRoles(member: SchoolMemberItem) {
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
    if (!existingRoleId.value) {
      existingRoleId.value =
        allowedRoles.value.find(
          (role) => normalizeRoleName(role.roleName) === "student",
        )?.roleId ?? "";
    }
  } catch {
    rolesError.value = "Role belum bisa dimuat.";
  } finally {
    rolesLoading.value = false;
  }
}

async function loadMembers() {
  if (!currentSchool.value.hasContext) return;

  membersLoading.value = true;
  membersError.value = "";
  try {
    const data = await getSchoolMembers(currentSchool.value.schoolCode, {
      page: 1,
      limit: 50,
      search: memberSearch.value.trim(),
    });
    members.value = data.members?.data ?? [];
    initializeRoleDrafts();
  } catch {
    membersError.value = "Member sekolah belum bisa dimuat.";
  } finally {
    membersLoading.value = false;
  }
}

async function searchUsers() {
  userSearchError.value = "";
  userResults.value = [];

  if (!userSearch.value.trim()) {
    userSearchError.value = "Masukkan nama atau email user global.";
    return;
  }

  userSearchLoading.value = true;
  try {
    const data = await getAdminUsers({
      page: 1,
      limit: 20,
      search: userSearch.value.trim(),
    });
    userResults.value = data.data ?? [];
  } catch {
    userSearchError.value = "Akun global belum bisa dicari.";
  } finally {
    userSearchLoading.value = false;
  }
}

async function reloadMembersAndFind(userId: string) {
  memberSearch.value = "";
  await loadMembers();
  return members.value.find((member) => member.userId === userId) ?? null;
}

async function syncRoleForMember(schoolUserId: string, roleId: string) {
  if (!roleId) {
    actionError.value = "Pilih satu role.";
    return;
  }

  savingRolesSchoolUserId.value = schoolUserId;
  actionError.value = "";
  actionMessage.value = "";
  try {
    await syncUserRoles(schoolUserId, { roleIds: [roleId] });
    actionMessage.value = "Role member berhasil diperbarui.";
    await loadMembers();
  } catch {
    actionError.value = "Role member belum bisa diperbarui.";
  } finally {
    savingRolesSchoolUserId.value = "";
  }
}

async function addExistingUser(user: AdminUserItem) {
  actionError.value = "";
  actionMessage.value = "";

  if (!currentSchool.value.schoolId || !currentSchool.value.schoolCode) {
    actionError.value = "Context sekolah aktif belum tersedia.";
    return;
  }
  if (!existingRoleId.value) {
    actionError.value = "Pilih role untuk user yang akan ditambahkan.";
    return;
  }
  if (isMember(user.userId)) {
    actionError.value = "User ini sudah menjadi member sekolah aktif.";
    return;
  }

  addExistingLoadingUserId.value = user.userId;
  try {
    await enrollUserToSchool({
      userId: user.userId,
      schoolId: currentSchool.value.schoolId,
    });

    const member = await reloadMembersAndFind(user.userId);
    if (!member) {
      actionError.value =
        "Membership belum ditemukan setelah user ditambahkan.";
      return;
    }

    await syncUserRoles(member.schoolUserId, {
      roleIds: [existingRoleId.value],
    });
    actionMessage.value = "User berhasil ditambahkan sebagai member sekolah.";
    await loadMembers();
  } catch {
    actionError.value = "User belum bisa ditambahkan ke sekolah.";
  } finally {
    addExistingLoadingUserId.value = "";
  }
}

function setRoleDraft(schoolUserId: string, roleId: string) {
  memberRoleDrafts.value = {
    ...memberRoleDrafts.value,
    [schoolUserId]: roleId,
  };
}

onMounted(async () => {
  if (!currentSchool.value.hasContext) return;
  await loadRoles();
  await loadMembers();
});
</script>

<template>
  <main class="min-h-screen flex-1 px-5 py-6 sm:px-8 lg:px-10">
    <section class="mx-auto flex max-w-6xl flex-col gap-6">
      <header class="soft-card rounded-[22px] p-5">
        <div
          class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between"
        >
          <div>
            <p class="text-[11px] font-medium uppercase text-[#9CA3AF]">
              School admin
            </p>
            <h1 class="mt-2 text-2xl font-medium text-[#111827]">
              User, membership, dan role
            </h1>
            <p class="mt-2 max-w-3xl text-sm leading-6 text-[#6B7280]">
              Kelola membership sekolah dari akun global yang sudah ada, lalu
              beri satu role dalam konteks sekolah aktif sebelum enrollment
              kelas.
            </p>
          </div>
          <div class="flex flex-wrap gap-2 text-xs">
            <span
              class="rounded-lg bg-[#EEF2FF] px-3 py-1.5 font-medium text-[#4F46E5]"
            >
              {{ currentSchool.schoolName || "Sekolah belum tersedia" }}
            </span>
            <span
              class="rounded-lg bg-[#F9FAFB] px-3 py-1.5 font-medium text-[#6B7280]"
            >
              {{ currentSchool.schoolCode || "Kode sekolah belum tersedia" }}
            </span>
          </div>
        </div>

        <div
          v-if="!currentSchool.hasContext"
          class="mt-4 rounded-[10px] border border-[#FECACA] bg-[#FEF2F2] px-4 py-3 text-sm text-[#DC2626]"
        >
          Context sekolah aktif belum tersedia. Pastikan akun admin memiliki
          membership sekolah.
        </div>

        <div
          v-if="actionMessage"
          class="mt-4 rounded-[10px] border border-[#BBF7D0] bg-[#ECFDF5] px-4 py-3 text-sm text-[#059669]"
        >
          {{ actionMessage }}
        </div>
        <div
          v-if="actionError"
          class="mt-4 rounded-[10px] border border-[#FECACA] bg-[#FEF2F2] px-4 py-3 text-sm text-[#DC2626]"
        >
          {{ actionError }}
        </div>
      </header>

      <section class="grid gap-5">
        <article class="bg-white border border-[#EBEBEB] rounded-[18px] p-5">
          <div class="flex items-start justify-between gap-4">
            <div>
              <p class="text-[11px] font-medium uppercase text-[#9CA3AF]">
                Existing global account
              </p>
              <h2 class="mt-2 text-base font-medium text-[#111827]">
                Tambah user yang sudah ada
              </h2>
              <p class="mt-1 max-w-2xl text-sm leading-6 text-[#6B7280]">
                Akun user bersifat global. School admin hanya menghubungkan akun
                yang sudah ada sebagai membership sekolah aktif, lalu memberi
                satu role sekolah.
              </p>
            </div>
            <PhMagnifyingGlass
              :size="22"
              class="text-[#4F46E5]"
              weight="duotone"
            />
          </div>

          <form
            class="mt-5 flex flex-col gap-3 sm:flex-row"
            @submit.prevent="searchUsers"
          >
            <input
              v-model="userSearch"
              type="search"
              placeholder="Cari nama atau email akun global"
              class="w-full rounded-2xl border border-[#EBEBEB] bg-white px-4 py-3 text-sm text-[#111827] outline-none transition placeholder:text-[#9CA3AF] focus:border-[#4F46E5]"
            />
            <button
              type="submit"
              class="inline-flex items-center justify-center gap-2 rounded-2xl bg-[#111827] px-4 py-3 text-sm font-medium text-white transition hover:bg-[#374151] disabled:cursor-not-allowed disabled:opacity-60"
              :disabled="userSearchLoading"
            >
              <PhMagnifyingGlass :size="18" weight="duotone" />
              Cari
            </button>
          </form>

          <label class="mt-4 block text-sm font-medium text-[#374151]">
            Role saat ditambahkan
            <select
              v-model="existingRoleId"
              class="mt-2 w-full rounded-2xl border border-[#EBEBEB] bg-white px-4 py-3 text-sm text-[#111827] outline-none transition focus:border-[#4F46E5]"
              :disabled="rolesLoading || allowedRoles.length === 0"
            >
              <option value="" disabled>Pilih role</option>
              <option
                v-for="role in allowedRoles"
                :key="role.roleId"
                :value="role.roleId"
              >
                {{ roleLabel(role.roleName) }}
              </option>
            </select>
          </label>

          <div class="mt-5 space-y-3">
            <p v-if="userSearchLoading" class="text-sm text-[#6B7280]">
              Mencari akun global...
            </p>
            <p v-else-if="userSearchError" class="text-sm text-[#DC2626]">
              {{ userSearchError }}
            </p>
            <p
              v-else-if="userResults.length === 0"
              class="text-sm text-[#6B7280]"
            >
              Cari akun global yang sudah ada untuk ditambahkan ke sekolah
              aktif.
            </p>

            <article
              v-for="user in userResults"
              :key="user.userId"
              class="rounded-[18px] border border-[#EBEBEB] bg-[#FBFAF8] p-4"
            >
              <div
                class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
              >
                <div>
                  <h3 class="text-sm font-medium text-[#111827]">
                    {{ user.fullName }}
                  </h3>
                  <p class="mt-1 text-xs text-[#6B7280]">{{ user.email }}</p>
                </div>
                <button
                  type="button"
                  class="inline-flex items-center justify-center gap-2 rounded-2xl border border-[#EBEBEB] bg-white px-3 py-2 text-sm font-medium text-[#111827] transition hover:bg-[#F9FAFB] disabled:cursor-not-allowed disabled:opacity-60"
                  :disabled="
                    isMember(user.userId) ||
                    !existingRoleId ||
                    addExistingLoadingUserId === user.userId
                  "
                  @click="addExistingUser(user)"
                >
                  {{ isMember(user.userId) ? "Sudah member" : "Tambahkan" }}
                </button>
              </div>
            </article>
          </div>
        </article>
      </section>

      <section class="bg-white border border-[#EBEBEB] rounded-[18px] p-5">
        <div
          class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between"
        >
          <div>
            <p class="text-[11px] font-medium uppercase text-[#9CA3AF]">
              School members
            </p>
            <h2 class="mt-2 text-base font-medium text-[#111827]">
              Member sekolah aktif
            </h2>
            <p class="mt-1 text-sm text-[#6B7280]">
              Role disimpan pada school_user. MVP saat ini mendukung satu role
              per member sekolah.
            </p>
          </div>
          <div
            class="inline-flex items-center gap-2 rounded-lg bg-[#EEF2FF] px-3 py-2 text-xs font-medium text-[#4F46E5]"
          >
            <PhUsers :size="16" weight="duotone" />
            {{ members.length }} member
          </div>
        </div>

        <form
          class="mt-5 flex flex-col gap-3 sm:flex-row"
          @submit.prevent="loadMembers"
        >
          <input
            v-model="memberSearch"
            type="search"
            placeholder="Cari member sekolah"
            class="w-full rounded-2xl border border-[#EBEBEB] bg-white px-4 py-3 text-sm text-[#111827] outline-none transition placeholder:text-[#9CA3AF] focus:border-[#4F46E5]"
          />
          <button
            type="submit"
            class="inline-flex items-center justify-center gap-2 rounded-2xl bg-[#111827] px-4 py-3 text-sm font-medium text-white transition hover:bg-[#374151] disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="membersLoading || !currentSchool.hasContext"
          >
            <PhMagnifyingGlass :size="18" weight="duotone" />
            Cari
          </button>
        </form>

        <div class="mt-5">
          <div
            v-if="rolesLoading"
            class="rounded-[18px] bg-[#FBFAF8] p-5 text-sm text-[#6B7280]"
          >
            Memuat role...
          </div>
          <div
            v-else-if="rolesError"
            class="rounded-[18px] border border-[#FECACA] bg-[#FEF2F2] p-5 text-sm text-[#DC2626]"
          >
            {{ rolesError }}
          </div>

          <div
            v-if="membersLoading"
            class="rounded-[18px] bg-[#FBFAF8] p-5 text-sm text-[#6B7280]"
          >
            Memuat member sekolah...
          </div>
          <div
            v-else-if="membersError"
            class="flex items-start gap-3 rounded-[18px] border border-[#FECACA] bg-[#FEF2F2] p-5 text-sm text-[#DC2626]"
          >
            <PhWarningCircle :size="20" weight="duotone" />
            <p>{{ membersError }}</p>
          </div>
          <div
            v-else-if="members.length === 0"
            class="rounded-[18px] bg-[#FBFAF8] p-5 text-sm text-[#6B7280]"
          >
            Belum ada member pada sekolah aktif.
          </div>

          <div v-else class="mt-4 grid gap-3">
            <article
              v-for="member in members"
              :key="member.schoolUserId"
              class="rounded-[18px] border border-[#EBEBEB] bg-[#FBFAF8] p-4"
            >
              <div
                class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between"
              >
                <div>
                  <div class="flex flex-wrap items-center gap-2">
                    <h3 class="text-sm font-medium text-[#111827]">
                      {{ member.fullName || "Nama tidak tersedia" }}
                    </h3>
                    <span
                      class="rounded-lg bg-white px-2 py-1 text-[11px] font-medium text-[#6B7280]"
                    >
                      school_user
                    </span>
                    <span
                      v-if="primaryRoleName(member)"
                      class="rounded-lg bg-[#EEF2FF] px-2 py-1 text-[11px] font-medium text-[#4F46E5]"
                    >
                      {{ roleLabel(primaryRoleName(member)) }}
                    </span>
                  </div>
                  <p class="mt-1 text-xs text-[#6B7280]">
                    {{ member.email || "Email tidak tersedia" }}
                  </p>
                  <p
                    v-if="hasMultipleAllowedRoles(member)"
                    class="mt-2 max-w-xl rounded-[10px] border border-[#FDE68A] bg-[#FFF7ED] px-3 py-2 text-xs leading-5 text-[#92400E]"
                  >
                    Data lama memiliki lebih dari satu role. Untuk MVP ini,
                    hanya satu role utama yang akan disimpan saat diperbarui.
                  </p>
                  <p class="mt-2 text-xs text-[#9CA3AF]">
                    Bergabung {{ formatDateTime(member.createdAt) }}
                  </p>
                </div>

                <div class="w-full max-w-xl">
                  <label class="block text-xs font-medium text-[#6B7280]">
                    Role sekolah
                    <select
                      class="mt-2 w-full rounded-2xl border border-[#EBEBEB] bg-white px-4 py-3 text-sm text-[#111827] outline-none transition focus:border-[#4F46E5]"
                      :value="memberRoleDrafts[member.schoolUserId] ?? ''"
                      :disabled="rolesLoading || allowedRoles.length === 0"
                      @change="
                        setRoleDraft(
                          member.schoolUserId,
                          ($event.target as HTMLSelectElement).value,
                        )
                      "
                    >
                      <option value="" disabled>Pilih role</option>
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
                    class="mt-3 inline-flex items-center gap-2 rounded-2xl bg-[#111827] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#374151] disabled:cursor-not-allowed disabled:opacity-60"
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
                    <PhShieldCheck :size="18" weight="duotone" />
                    Simpan role
                  </button>
                </div>
              </div>
            </article>
          </div>
        </div>
      </section>
    </section>
  </main>
</template>
