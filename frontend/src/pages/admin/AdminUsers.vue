<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import {
  PhMagnifyingGlass,
  PhShieldCheck,
  PhUsers,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import { useToastStore } from "../../stores/toast";
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
  if (normalized === "student") return "Siswa";
  if (normalized === "teacher") return "Guru";
  if (normalized === "admin") return "Admin sekolah";
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
    const data = await getSchoolMembers(currentSchool.value.schoolCode, {
      page: 1,
      limit: 50,
      search: memberSearch.value.trim(),
    });
    members.value = data.members?.data ?? [];
    initializeRoleDrafts();
  } catch {
    membersError.value = "Warga sekolah belum bisa dimuat.";
  } finally {
    membersLoading.value = false;
  }
}

async function searchUsers() {
  userSearchError.value = "";
  userResults.value = [];

  if (!userSearch.value.trim()) {
    userSearchError.value = "Masukkan nama atau email pengguna.";
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

async function addExistingUser(user: AdminUserItem) {
  if (!currentSchool.value.schoolId || !currentSchool.value.schoolCode) {
    toast.error("Konteks sekolah aktif belum tersedia.");
    return;
  }
  if (!existingRoleId.value) {
    toast.error("Pilih peran untuk pengguna yang akan ditambahkan.");
    return;
  }
  if (isMember(user.userId)) {
    toast.info("Pengguna ini sudah menjadi warga sekolah aktif.");
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
      toast.error(
        "Akses sekolah belum ditemukan setelah pengguna ditambahkan.",
      );
      return;
    }

    await syncUserRoles(member.schoolUserId, {
      roleIds: [existingRoleId.value],
    });
    toast.success("Pengguna berhasil ditambahkan sebagai warga sekolah.");
    await loadMembers();
  } catch {
    toast.error("Pengguna belum bisa ditambahkan ke sekolah.");
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
            Hubungkan akun yang sudah ada ke sekolah dan atur perannya sebelum
            melakukan penempatan kelas.
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

    <section class="px-5 py-5 sm:px-6 lg:px-8">
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
          class="order-2 min-w-0 rounded-xl border border-[#ebe7df] bg-white lg:order-1"
        >
          <div
            class="flex flex-col gap-4 border-b border-[#ebe7df] px-4 py-4 sm:px-5"
          >
            <div
              class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between"
            >
              <div>
                <p
                  class="text-[10px] font-medium uppercase tracking-[0.08em] text-[#9ca3af]"
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

            <form
              class="flex min-w-0 flex-col gap-2 sm:flex-row"
              @submit.prevent="loadMembers"
            >
              <input
                v-model="memberSearch"
                type="search"
                placeholder="Cari nama atau email warga sekolah"
                class="min-w-0 flex-1 rounded-lg border border-[#ebe7df] bg-[#fbfaf8] px-3.5 py-2.5 text-sm text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#4f46e5] focus:bg-white"
              />
              <button
                type="submit"
                class="inline-flex shrink-0 items-center justify-center gap-2 rounded-lg bg-[#171322] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#374151] disabled:opacity-60"
                :disabled="membersLoading || !currentSchool.hasContext"
              >
                <PhMagnifyingGlass :size="17" weight="duotone" />
                Cari
              </button>
            </form>
          </div>

          <div class="p-4 sm:p-5">
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
                class="mt-4 rounded-lg bg-[#171322] px-4 py-2 text-sm font-medium text-white transition hover:bg-[#374151]"
                @click="rolesError ? loadRoles() : loadMembers()"
              >
                Coba lagi
              </button>
            </div>

            <div
              v-else-if="members.length === 0"
              class="rounded-lg bg-[#fbfaf8] px-5 py-10 text-center"
            >
              <PhUsers
                :size="28"
                class="mx-auto text-[#9ca3af]"
                weight="duotone"
              />
              <h3 class="mt-3 text-sm font-semibold text-[#171322]">
                Belum ada warga sekolah
              </h3>
              <p class="mt-2 text-sm leading-6 text-[#6b7280]">
                Cari pengguna melalui panel tambah pengguna untuk memulai.
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
                      class="mt-2.5 inline-flex w-full items-center justify-center gap-2 rounded-lg bg-[#171322] px-3 py-2.5 text-sm font-medium text-white transition hover:bg-[#374151] disabled:opacity-60"
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
                  </div>
                </div>
              </article>
            </div>
          </div>
        </section>

        <aside class="order-1 min-w-0 lg:order-2">
          <section
            class="rounded-xl border border-[#ebe7df] bg-white p-5 lg:sticky lg:top-6"
          >
            <div class="flex items-start justify-between gap-3">
              <div>
                <p
                  class="text-[10px] font-medium uppercase tracking-[0.08em] text-[#9ca3af]"
                >
                  Tambah pengguna
                </p>
                <h2 class="mt-1 text-base font-semibold text-[#171322]">
                  Cari akun yang sudah ada
                </h2>
                <p class="mt-1 text-xs leading-5 text-[#6b7280]">
                  Cari akun EduVerse, pilih peran, lalu hubungkan ke sekolah.
                </p>
              </div>
              <PhMagnifyingGlass
                :size="21"
                class="text-[#ea580c]"
                weight="duotone"
              />
            </div>

            <form class="mt-5 space-y-3" @submit.prevent="searchUsers">
              <label class="block text-xs font-medium text-[#6b7280]">
                Nama atau email
                <input
                  v-model="userSearch"
                  type="search"
                  placeholder="Cari pengguna"
                  class="mt-2 w-full rounded-lg border border-[#ebe7df] bg-[#fbfaf8] px-3.5 py-2.5 text-sm text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#4f46e5] focus:bg-white"
                />
              </label>
              <label class="block text-xs font-medium text-[#6b7280]">
                Peran saat ditambahkan
                <select
                  v-model="existingRoleId"
                  class="mt-2 w-full rounded-lg border border-[#ebe7df] bg-[#fbfaf8] px-3.5 py-2.5 text-sm text-[#171322] outline-none transition focus:border-[#4f46e5] focus:bg-white"
                  :disabled="rolesLoading || allowedRoles.length === 0"
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
                type="submit"
                class="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-[#ea580c] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#c2410c] disabled:opacity-60"
                :disabled="userSearchLoading"
              >
                <PhMagnifyingGlass :size="17" weight="duotone" />
                {{ userSearchLoading ? "Mencari..." : "Cari pengguna" }}
              </button>
            </form>

            <div class="mt-4 space-y-2">
              <p
                v-if="userSearchError"
                class="rounded-lg bg-[#fef2f2] px-3 py-2 text-xs leading-5 text-[#dc2626]"
              >
                {{ userSearchError }}
              </p>
              <p
                v-else-if="!userSearchLoading && userResults.length === 0"
                class="rounded-lg bg-[#fbfaf8] px-3 py-3 text-xs leading-5 text-[#6b7280]"
              >
                Hasil pencarian pengguna akan tampil di sini.
              </p>

              <article
                v-for="user in userResults"
                :key="user.userId"
                class="rounded-lg border border-[#ebe7df] bg-[#fbfaf8] p-3"
              >
                <h3
                  class="wrap-break-word text-sm font-semibold text-[#171322]"
                >
                  {{ user.fullName }}
                </h3>
                <p class="mt-1 break-all text-xs text-[#6b7280]">
                  {{ user.email }}
                </p>
                <button
                  type="button"
                  class="mt-3 inline-flex w-full items-center justify-center rounded-lg border border-[#ebe7df] bg-white px-3 py-2 text-xs font-medium text-[#171322] transition hover:border-[#ea580c] hover:text-[#ea580c] disabled:opacity-60"
                  :disabled="
                    isMember(user.userId) ||
                    !existingRoleId ||
                    addExistingLoadingUserId === user.userId
                  "
                  @click="addExistingUser(user)"
                >
                  {{
                    addExistingLoadingUserId === user.userId
                      ? "Menambahkan..."
                      : isMember(user.userId)
                        ? "Sudah terhubung"
                        : "Tambahkan ke sekolah"
                  }}
                </button>
              </article>
            </div>
          </section>
        </aside>
      </div>
    </section>
  </main>
</template>
