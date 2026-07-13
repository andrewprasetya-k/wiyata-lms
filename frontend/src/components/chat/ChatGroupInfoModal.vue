<script setup lang="ts">
import { computed, ref, watch } from "vue";
import {
  addChatGroupMembers,
  getChatGroupInfo,
  leaveChatGroup,
  removeChatGroupMember,
  renameChatGroup,
  searchChatMembers,
} from "../../services/chat";
import { APP_TIME_ZONE, parseBackendTimestamp } from "../../utils/date";
import {
  getInitials,
  resolveChatError,
  roomDisplayName,
} from "../../utils/chatDisplay";
import { useConfirmStore } from "../../stores/confirm";
import { useToastStore } from "../../stores/toast";
import type {
  ChatGroupInfo,
  ChatGroupMember,
  ChatMember,
  ChatRoom,
} from "../../types/chat";

const props = defineProps<{
  open: boolean;
  room: ChatRoom | null;
  currentUserId: string;
}>();

const emit = defineEmits<{
  (event: "update:open", value: boolean): void;
  (event: "renamed", room: ChatRoom): void;
  (event: "members-changed"): void;
  (event: "left"): void;
}>();

const toast = useToastStore();
const confirm = useConfirmStore();

const groupInfo = ref<ChatGroupInfo | null>(null);
const isLoadingGroupInfo = ref(false);
const groupInfoError = ref("");
const groupActionError = ref("");
const renameRoomName = ref("");
const isRenamingGroup = ref(false);
const addMemberSearch = ref("");
const addMemberResults = ref<ChatMember[]>([]);
const selectedAddMemberIds = ref<string[]>([]);
const isLoadingAddMembers = ref(false);
const isAddingMembers = ref(false);
const isLeavingGroup = ref(false);
const removingMemberId = ref<string | null>(null);

const selectedSchoolName = computed(
  () => props.room?.schoolName || "Sekolah aktif",
);
const currentUserIsGroupAdmin = computed(() =>
  Boolean(
    groupInfo.value?.admins.some(
      (member) => member.userId === props.currentUserId,
    ),
  ),
);
const selectedAddMembers = computed(() =>
  addMemberResults.value.filter((member) =>
    selectedAddMemberIds.value.includes(member.userId),
  ),
);

function formatDateTime(value?: string | null) {
  if (!value) return "";
  const date = parseBackendTimestamp(value);
  if (!date || Number.isNaN(date.getTime())) return "";
  return new Intl.DateTimeFormat("id-ID", {
    day: "2-digit",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
    timeZone: APP_TIME_ZONE,
  }).format(date);
}

async function loadGroupInfo() {
  if (!props.room) return;
  isLoadingGroupInfo.value = true;
  groupInfoError.value = "";
  groupActionError.value = "";
  try {
    groupInfo.value = await getChatGroupInfo(props.room.roomId);
    renameRoomName.value = groupInfo.value.roomName;
    if (
      groupInfo.value.admins.some(
        (member) => member.userId === props.currentUserId,
      )
    ) {
      await loadEligibleMembers();
    }
  } catch (error) {
    groupInfoError.value = resolveChatError(error);
  } finally {
    isLoadingGroupInfo.value = false;
  }
}

async function submitRenameGroup() {
  if (!props.room || !groupInfo.value) return;
  const roomName = renameRoomName.value.trim();
  if (roomName.length < 3) {
    groupActionError.value = "Nama ruang minimal 3 karakter.";
    return;
  }
  isRenamingGroup.value = true;
  groupActionError.value = "";
  try {
    const room = await renameChatGroup(props.room.roomId, { roomName });
    groupInfo.value = { ...groupInfo.value, roomName: room.roomName };
    emit("renamed", room);
    toast.success("Nama ruang chat berhasil diperbarui.");
  } catch (error) {
    groupActionError.value = resolveChatError(error);
  } finally {
    isRenamingGroup.value = false;
  }
}

async function loadEligibleMembers() {
  if (!props.room) return;
  isLoadingAddMembers.value = true;
  groupActionError.value = "";
  try {
    addMemberResults.value = await searchChatMembers(
      addMemberSearch.value.trim(),
      props.room.roomId,
    );
  } catch (error) {
    groupActionError.value = resolveChatError(error);
  } finally {
    isLoadingAddMembers.value = false;
  }
}

function toggleAddMember(userId: string) {
  if (selectedAddMemberIds.value.includes(userId)) {
    selectedAddMemberIds.value = selectedAddMemberIds.value.filter(
      (id) => id !== userId,
    );
    return;
  }
  selectedAddMemberIds.value = [...selectedAddMemberIds.value, userId];
}

async function submitAddMembers() {
  if (!props.room || selectedAddMemberIds.value.length === 0) {
    groupActionError.value = "Pilih minimal satu anggota.";
    return;
  }
  isAddingMembers.value = true;
  groupActionError.value = "";
  try {
    await addChatGroupMembers(props.room.roomId, {
      memberUserIds: selectedAddMemberIds.value,
    });
    selectedAddMemberIds.value = [];
    addMemberSearch.value = "";
    await loadGroupInfo();
    emit("members-changed");
    toast.success("Anggota berhasil ditambahkan.");
  } catch (error) {
    groupActionError.value = resolveChatError(error);
  } finally {
    isAddingMembers.value = false;
  }
}

async function leaveSelectedGroup() {
  if (!props.room) return;
  const ok = await confirm.confirm({
    title: `Keluar dari ${roomDisplayName(props.room)}?`,
    description: "Kamu tidak akan bisa mengakses pesan grup ini lagi.",
    confirmLabel: "Keluar",
    variant: "warning",
  });
  if (!ok) return;

  isLeavingGroup.value = true;
  groupActionError.value = "";
  try {
    await leaveChatGroup(props.room.roomId);
    groupInfo.value = null;
    emit("left");
    toast.success("Kamu berhasil keluar dari ruang chat.");
  } catch (error) {
    groupActionError.value = resolveChatError(error);
  } finally {
    isLeavingGroup.value = false;
  }
}

async function removeMember(member: ChatGroupMember) {
  if (!props.room) return;
  const ok = await confirm.confirm({
    title: `Keluarkan ${member.fullName || member.email}?`,
    description: "Anggota ini akan dihapus dari grup chat.",
    confirmLabel: "Keluarkan",
    variant: "danger",
  });
  if (!ok) return;

  removingMemberId.value = member.userId;
  groupActionError.value = "";
  try {
    await removeChatGroupMember(props.room.roomId, member.userId);
    await loadGroupInfo();
    emit("members-changed");
    toast.success("Anggota berhasil dikeluarkan.");
  } catch (error) {
    groupActionError.value = resolveChatError(error);
  } finally {
    removingMemberId.value = null;
  }
}

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) void loadGroupInfo();
  },
);
</script>

<template>
  <div v-if="open" class="fixed inset-0 z-50 flex justify-end bg-black/30">
    <div class="flex h-full w-full max-w-lg flex-col overflow-hidden bg-surface">
      <div class="border-b border-border px-5 py-4">
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0">
            <p
              class="text-xs font-semibold uppercase tracking-[0.06em] text-muted"
            >
              Info grup
            </p>
            <h2 class="mt-1 truncate text-lg font-semibold text-foreground">
              {{ groupInfo?.roomName || roomDisplayName(room) }}
            </h2>
            <p class="mt-1 text-sm text-muted">
              {{ groupInfo?.memberCount || 0 }} anggota ·
              {{ groupInfo?.schoolName || selectedSchoolName }}
            </p>
          </div>
          <button
            type="button"
            class="rounded-lg border border-border px-3 py-2 text-sm font-medium text-muted transition hover:bg-surface-subtle"
            @click="emit('update:open', false)"
          >
            Tutup
          </button>
        </div>
      </div>

      <div class="min-h-0 flex-1 overflow-y-auto px-5 py-4">
        <div v-if="isLoadingGroupInfo" class="space-y-3">
          <div class="h-16 animate-pulse rounded-xl bg-surface-strong" />
          <div class="h-32 animate-pulse rounded-xl bg-surface-strong" />
          <div class="h-40 animate-pulse rounded-xl bg-surface-strong" />
        </div>

        <div
          v-else-if="groupInfoError"
          class="rounded-xl border border-red-100 bg-red-50 px-4 py-5 text-sm text-danger"
        >
          <p>{{ groupInfoError }}</p>
          <button
            type="button"
            class="mt-3 rounded-lg bg-brand px-3 py-2 text-xs font-semibold text-white"
            @click="loadGroupInfo"
          >
            Coba lagi
          </button>
        </div>

        <template v-else>
          <p
            v-if="groupActionError"
            class="mb-4 rounded-lg bg-red-50 px-3 py-2 text-sm text-danger"
          >
            {{ groupActionError }}
          </p>

          <section class="rounded-xl border border-border bg-surface-subtle p-4">
            <p class="text-sm font-semibold text-foreground">Dibuat oleh</p>
            <div class="mt-3 flex items-center gap-3">
              <span
                class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-brand text-xs font-semibold text-white"
              >
                {{
                  getInitials(
                    groupInfo?.creator?.fullName || groupInfo?.creator?.email,
                  )
                }}
              </span>
              <span class="min-w-0">
                <span
                  class="block truncate text-sm font-medium text-foreground"
                >
                  {{
                    groupInfo?.creator?.fullName ||
                    groupInfo?.creator?.email ||
                    "Tidak tersedia"
                  }}
                </span>
                <span class="block truncate text-xs text-muted">
                  {{ groupInfo?.creator?.email }}
                </span>
              </span>
            </div>
            <p class="mt-3 text-xs text-muted">
              Dibuat {{ formatDateTime(groupInfo?.createdAt) }}
            </p>
          </section>

          <section
            v-if="currentUserIsGroupAdmin"
            class="mt-4 rounded-xl border border-border bg-surface p-4"
          >
            <p class="text-sm font-semibold text-foreground">Ubah nama grup</p>
            <form class="mt-3 flex gap-2" @submit.prevent="submitRenameGroup">
              <input
                v-model="renameRoomName"
                type="text"
                class="min-w-0 flex-1 rounded-lg border border-border px-3 py-2 text-sm text-foreground outline-none transition focus:border-brand focus:ring-2 focus:ring-brand/15"
                placeholder="Nama ruang grup"
              />
              <button
                type="submit"
                class="rounded-lg bg-brand px-3 py-2 text-sm font-semibold text-white transition hover:bg-brand-hover disabled:cursor-not-allowed disabled:bg-[#c7c3d7]"
                :disabled="isRenamingGroup"
              >
                {{ isRenamingGroup ? "Menyimpan..." : "Simpan" }}
              </button>
            </form>
          </section>

          <section
            v-if="currentUserIsGroupAdmin"
            class="mt-4 rounded-xl border border-border bg-surface p-4"
          >
            <p class="text-sm font-semibold text-foreground">Tambah anggota</p>
            <p class="mt-1 text-xs text-muted">
              Hanya warga aktif sekolah yang belum ada di grup ini.
            </p>
            <div class="mt-3 flex gap-2">
              <input
                v-model="addMemberSearch"
                type="search"
                class="min-w-0 flex-1 rounded-lg border border-border px-3 py-2 text-sm text-foreground outline-none transition focus:border-brand focus:ring-2 focus:ring-brand/15"
                placeholder="Cari nama atau email..."
                @keydown.enter.prevent="loadEligibleMembers"
              />
              <button
                type="button"
                class="rounded-lg border border-border px-3 py-2 text-sm font-medium text-brand transition hover:border-brand disabled:opacity-60"
                :disabled="isLoadingAddMembers"
                @click="loadEligibleMembers"
              >
                Cari
              </button>
            </div>

            <div class="mt-3 rounded-lg border border-border">
              <div v-if="isLoadingAddMembers" class="space-y-2 p-3">
                <div class="h-10 animate-pulse rounded-lg bg-surface-strong" />
                <div class="h-10 animate-pulse rounded-lg bg-surface-strong" />
              </div>
              <div
                v-else-if="addMemberResults.length === 0"
                class="rounded-lg bg-surface-subtle p-3 text-sm leading-6 text-muted"
              >
                Tidak ada warga yang bisa ditambahkan.
              </div>
              <div v-else class="max-h-52 overflow-y-auto p-2">
                <label
                  v-for="member in addMemberResults"
                  :key="member.userId"
                  class="flex cursor-pointer items-center gap-3 rounded-lg px-2 py-2 hover:bg-surface-subtle"
                >
                  <input
                    type="checkbox"
                    class="h-4 w-4 rounded border-border text-brand"
                    :checked="selectedAddMemberIds.includes(member.userId)"
                    @change="toggleAddMember(member.userId)"
                  />
                  <span
                    class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-brand text-xs font-semibold text-white"
                  >
                    {{ getInitials(member.fullName || member.email) }}
                  </span>
                  <span class="min-w-0 flex-1">
                    <span
                      class="block truncate text-sm font-medium text-foreground"
                    >
                      {{ member.fullName || member.email }}
                    </span>
                    <span class="block truncate text-xs text-muted">
                      {{ member.email }}
                    </span>
                  </span>
                </label>
              </div>
            </div>

            <div class="mt-3 flex items-center justify-between gap-3">
              <p class="text-xs text-muted">
                {{ selectedAddMembers.length }} anggota dipilih
              </p>
              <button
                type="button"
                class="rounded-lg bg-brand px-3 py-2 text-sm font-semibold text-white transition hover:bg-brand-hover disabled:cursor-not-allowed disabled:bg-[#c7c3d7]"
                :disabled="isAddingMembers || selectedAddMemberIds.length === 0"
                @click="submitAddMembers"
              >
                {{ isAddingMembers ? "Menambahkan..." : "Tambah anggota" }}
              </button>
            </div>
          </section>

          <section class="mt-4 rounded-xl border border-border bg-surface">
            <div class="border-b border-border px-4 py-3">
              <p class="text-sm font-semibold text-foreground">Anggota</p>
            </div>
            <div class="divide-y divide-border">
              <div
                v-for="member in groupInfo?.members || []"
                :key="member.userId"
                class="flex items-center gap-3 px-4 py-3"
              >
                <span
                  class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-brand text-xs font-semibold text-white"
                >
                  {{ getInitials(member.fullName || member.email) }}
                </span>
                <span class="min-w-0 flex-1">
                  <span
                    class="block truncate text-sm font-medium text-foreground"
                  >
                    {{ member.fullName || member.email }}
                  </span>
                  <span class="block truncate text-xs text-muted">
                    {{ member.email }}
                  </span>
                </span>
                <span
                  class="rounded-full px-2 py-1 text-[11px] font-semibold"
                  :class="
                    member.role === 'admin'
                      ? 'bg-brand-soft text-brand'
                      : 'bg-surface-strong text-muted'
                  "
                >
                  {{ member.role === "admin" ? "Admin" : "Anggota" }}
                </span>
                <button
                  v-if="
                    currentUserIsGroupAdmin && member.userId !== currentUserId
                  "
                  type="button"
                  class="rounded-lg border border-red-100 px-3 py-1.5 text-xs font-semibold text-danger transition hover:bg-red-50 disabled:opacity-60"
                  :disabled="removingMemberId === member.userId"
                  @click="removeMember(member)"
                >
                  {{
                    removingMemberId === member.userId
                      ? "Menghapus..."
                      : "Keluarkan"
                  }}
                </button>
              </div>
            </div>
          </section>
        </template>
      </div>

      <div class="border-t border-border px-5 py-4">
        <button
          type="button"
          class="w-full rounded-lg border border-red-100 px-4 py-2 text-sm font-semibold text-danger transition hover:bg-red-50 disabled:opacity-60"
          :disabled="isLeavingGroup"
          @click="leaveSelectedGroup"
        >
          {{ isLeavingGroup ? "Keluar..." : "Keluar grup" }}
        </button>
      </div>
    </div>
  </div>
</template>
