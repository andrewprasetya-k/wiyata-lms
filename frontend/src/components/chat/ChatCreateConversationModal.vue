<script setup lang="ts">
import { ref, watch } from "vue";
import { PhX } from "@phosphor-icons/vue";
import {
  createChatGroup,
  openDirectMessage,
  searchChatMembers,
} from "../../services/chat";
import { getInitials, resolveChatError } from "../../utils/chatDisplay";
import type { ChatMember, ChatRoom } from "../../types/chat";

const props = defineProps<{
  open: boolean;
  initialTab: "dm" | "group";
  currentUserId: string;
}>();

const emit = defineEmits<{
  (event: "update:open", value: boolean): void;
  (event: "dm-opened", room: ChatRoom): void;
  (event: "group-created", room: ChatRoom): void;
}>();

const activeCreateTab = ref<"dm" | "group">(props.initialTab);

const groupRoomName = ref("");
const memberSearch = ref("");
const memberResults = ref<ChatMember[]>([]);
const selectedMemberIds = ref<string[]>([]);
const isLoadingMembers = ref(false);
const isCreatingGroup = ref(false);
const createGroupError = ref("");

const dmSearch = ref("");
const dmResults = ref<ChatMember[]>([]);
const selectedDMTargetId = ref("");
const isLoadingDMTargets = ref(false);
const isOpeningDM = ref(false);
const directMessageError = ref("");

function close() {
  emit("update:open", false);
}

async function loadChatMembers() {
  isLoadingMembers.value = true;
  createGroupError.value = "";
  try {
    memberResults.value = await searchChatMembers(memberSearch.value.trim());
  } catch (error) {
    createGroupError.value = resolveChatError(error);
  } finally {
    isLoadingMembers.value = false;
  }
}

async function loadDMTargets() {
  isLoadingDMTargets.value = true;
  directMessageError.value = "";
  try {
    const members = await searchChatMembers(dmSearch.value.trim());
    dmResults.value = members.filter(
      (member) => member.userId !== props.currentUserId,
    );
  } catch (error) {
    directMessageError.value = resolveChatError(error);
  } finally {
    isLoadingDMTargets.value = false;
  }
}

function toggleMember(userId: string) {
  if (selectedMemberIds.value.includes(userId)) {
    selectedMemberIds.value = selectedMemberIds.value.filter(
      (id) => id !== userId,
    );
    return;
  }
  selectedMemberIds.value = [...selectedMemberIds.value, userId];
}

async function submitCreateGroup() {
  const roomName = groupRoomName.value.trim();
  if (!roomName) {
    createGroupError.value = "Nama ruang wajib diisi.";
    return;
  }
  if (selectedMemberIds.value.length === 0) {
    createGroupError.value = "Pilih minimal satu anggota ruang.";
    return;
  }

  isCreatingGroup.value = true;
  createGroupError.value = "";
  try {
    const room = await createChatGroup({
      roomName,
      memberUserIds: selectedMemberIds.value,
    });
    groupRoomName.value = "";
    memberSearch.value = "";
    selectedMemberIds.value = [];
    emit("group-created", room);
  } catch (error) {
    createGroupError.value = resolveChatError(error);
  } finally {
    isCreatingGroup.value = false;
  }
}

async function submitDirectMessage() {
  if (!selectedDMTargetId.value) {
    directMessageError.value =
      "Pilih satu warga sekolah untuk memulai percakapan.";
    return;
  }

  isOpeningDM.value = true;
  directMessageError.value = "";
  try {
    const room = await openDirectMessage({
      targetUserId: selectedDMTargetId.value,
    });
    dmSearch.value = "";
    selectedDMTargetId.value = "";
    emit("dm-opened", room);
  } catch (error) {
    directMessageError.value = resolveChatError(error);
  } finally {
    isOpeningDM.value = false;
  }
}

async function loadForTab(tab: "dm" | "group") {
  if (tab === "group" && memberResults.value.length === 0) {
    await loadChatMembers();
  } else if (tab === "dm" && dmResults.value.length === 0) {
    directMessageError.value = "";
    selectedDMTargetId.value = "";
    await loadDMTargets();
  }
}

watch(
  () => props.open,
  (isOpen) => {
    if (!isOpen) return;
    activeCreateTab.value = props.initialTab;
    void loadForTab(activeCreateTab.value);
  },
);

watch(activeCreateTab, (tab) => {
  if (!props.open) return;
  void loadForTab(tab);
});
</script>

<template>
  <div
    v-if="open"
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/30 px-4 py-6"
    @click.self="close"
  >
    <div class="max-h-[90vh] w-full max-w-xl overflow-hidden rounded-xl bg-surface">
      <!-- Header -->
      <div class="px-5 py-4">
        <div class="flex items-center justify-between">
          <h2 class="text-base font-semibold text-foreground">
            Buat Percakapan
          </h2>
          <button
            type="button"
            class="rounded-lg p-1.5 text-muted transition hover:bg-surface-strong hover:text-foreground"
            aria-label="Tutup"
            @click="close"
          >
            <PhX class="h-5 w-5" />
          </button>
        </div>
        <!-- Tab bar -->
        <div class="mt-3 flex gap-1 rounded-lg bg-surface-strong p-1">
          <button
            type="button"
            class="flex-1 rounded-md py-1.5 text-sm font-medium transition"
            :class="
              activeCreateTab === 'dm'
                ? 'bg-surface text-foreground shadow-sm'
                : 'text-muted hover:text-foreground'
            "
            @click="activeCreateTab = 'dm'"
          >
            Pesan Langsung
          </button>
          <button
            type="button"
            class="flex-1 rounded-md py-1.5 text-sm font-medium transition"
            :class="
              activeCreateTab === 'group'
                ? 'bg-surface text-foreground shadow-sm'
                : 'text-muted hover:text-foreground'
            "
            @click="activeCreateTab = 'group'"
          >
            Grup
          </button>
        </div>
      </div>

      <!-- DM Panel -->
      <form
        v-if="activeCreateTab === 'dm'"
        class="flex max-h-[calc(90vh-9rem)] flex-col"
        @submit.prevent="submitDirectMessage"
      >
        <div class="space-y-4 overflow-y-auto px-5 py-4">
          <div>
            <label
              class="text-sm font-medium text-foreground"
              for="chat-dm-search"
            >
              Cari warga sekolah
            </label>
            <div class="mt-1 flex gap-2">
              <input
                id="chat-dm-search"
                v-model="dmSearch"
                type="text"
                class="min-w-0 flex-1 rounded-lg border border-border px-3 py-2 text-sm text-foreground outline-none transition focus:border-brand focus:ring-2 focus:ring-brand/15"
                placeholder="Cari warga sekolah..."
                @keydown.enter.prevent="loadDMTargets"
              />
              <button
                type="button"
                class="rounded-lg border border-border px-3 py-2 text-sm font-medium text-brand transition hover:border-brand disabled:opacity-60"
                :disabled="isLoadingDMTargets"
                @click="loadDMTargets"
              >
                Cari
              </button>
            </div>
          </div>

          <p
            v-if="directMessageError"
            class="rounded-lg bg-red-50 px-3 py-2 text-sm text-danger"
          >
            {{ directMessageError }}
          </p>

          <div class="rounded-lg border border-border">
            <div
              class="border-b border-border bg-surface-subtle px-3 py-2 text-xs font-semibold uppercase tracking-[0.06em] text-muted"
            >
              Warga sekolah
            </div>
            <div v-if="isLoadingDMTargets" class="space-y-2 p-3">
              <div class="h-10 animate-pulse rounded-lg bg-surface-strong" />
              <div class="h-10 animate-pulse rounded-lg bg-surface-strong" />
            </div>
            <div
              v-else-if="dmResults.length === 0"
              class="rounded-lg bg-surface-subtle p-3 text-sm leading-6 text-muted"
            >
              Tidak ada warga sekolah yang cocok.
            </div>
            <div v-else class="max-h-64 overflow-y-auto p-2">
              <label
                v-for="member in dmResults"
                :key="member.userId"
                class="flex cursor-pointer items-center gap-3 rounded-lg px-2 py-2 hover:bg-surface-subtle"
              >
                <input
                  v-model="selectedDMTargetId"
                  type="radio"
                  name="dm-target"
                  class="h-4 w-4 border-border text-brand"
                  :value="member.userId"
                />
                <span
                  class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-[#059669] text-xs font-semibold text-white"
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
        </div>

        <div class="flex flex-col gap-2 px-5 py-4 sm:flex-row sm:justify-end">
          <button
            type="button"
            class="rounded-lg border border-border px-4 py-2 text-sm font-medium text-muted transition hover:bg-surface-subtle"
            :disabled="isOpeningDM"
            @click="close"
          >
            Batal
          </button>
          <button
            type="submit"
            class="rounded-lg bg-brand px-4 py-2 text-sm font-semibold text-white transition hover:bg-brand-hover disabled:cursor-not-allowed disabled:bg-[#c7c3d7]"
            :disabled="isOpeningDM"
          >
            {{ isOpeningDM ? "Membuka..." : "Buka percakapan" }}
          </button>
        </div>
      </form>

      <!-- Group Panel -->
      <form
        v-else-if="activeCreateTab === 'group'"
        class="flex max-h-[calc(90vh-9rem)] flex-col"
        @submit.prevent="submitCreateGroup"
      >
        <div class="space-y-4 overflow-y-auto px-5 py-4">
          <div>
            <label
              class="text-sm font-medium text-foreground"
              for="chat-group-name"
            >
              Nama ruang
            </label>
            <input
              id="chat-group-name"
              v-model="groupRoomName"
              type="text"
              class="mt-1 w-full rounded-lg border border-border px-3 py-2 text-sm text-foreground outline-none transition focus:border-blue focus:ring-1 focus:ring-brand"
              placeholder="Contoh: Grup Belajar Fisika"
            />
          </div>

          <div>
            <label
              class="text-sm font-medium text-foreground"
              for="chat-member-search"
            >
              Cari warga sekolah
            </label>
            <div class="mt-1 flex gap-2">
              <input
                id="chat-member-search"
                v-model="memberSearch"
                type="text"
                class="min-w-0 flex-1 rounded-lg border border-border px-3 py-2 text-sm text-foreground outline-none transition focus:border-brand focus:ring-2 focus:ring-brand/15"
                placeholder="Cari nama atau email..."
                @keydown.enter.prevent="loadChatMembers"
              />
              <button
                type="button"
                class="rounded-lg border border-border px-3 py-2 text-sm font-medium text-brand transition hover:border-brand disabled:opacity-60"
                :disabled="isLoadingMembers"
                @click="loadChatMembers"
              >
                Cari
              </button>
            </div>
          </div>

          <p
            v-if="createGroupError"
            class="rounded-lg bg-red-50 px-3 py-2 text-sm text-danger"
          >
            {{ createGroupError }}
          </p>

          <div class="rounded-lg border border-border">
            <div
              class="border-b border-border bg-surface-subtle px-3 py-2 text-xs font-semibold uppercase tracking-[0.06em] text-muted"
            >
              Anggota
            </div>
            <div v-if="isLoadingMembers" class="space-y-2 p-3">
              <div class="h-10 animate-pulse rounded-lg bg-surface-strong" />
              <div class="h-10 animate-pulse rounded-lg bg-surface-strong" />
            </div>
            <div
              v-else-if="memberResults.length === 0"
              class="rounded-lg bg-surface-subtle p-3 text-sm leading-6 text-muted"
            >
              Tidak ada warga yang cocok.
            </div>
            <div v-else class="max-h-64 overflow-y-auto p-2">
              <label
                v-for="member in memberResults"
                :key="member.userId"
                class="flex cursor-pointer items-center gap-3 rounded-lg px-2 py-2 hover:bg-surface-subtle"
              >
                <input
                  type="checkbox"
                  class="h-4 w-4 rounded border-border text-brand"
                  :checked="selectedMemberIds.includes(member.userId)"
                  @change="toggleMember(member.userId)"
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
        </div>

        <div
          class="flex flex-col gap-2 border-t border-border px-5 py-4 sm:flex-row sm:justify-end"
        >
          <button
            type="button"
            class="rounded-lg border border-border px-4 py-2 text-sm font-medium text-muted transition hover:bg-surface-subtle"
            :disabled="isCreatingGroup"
            @click="close"
          >
            Batal
          </button>
          <button
            type="submit"
            class="rounded-lg bg-brand px-4 py-2 text-sm font-semibold text-white transition hover:bg-brand-hover disabled:cursor-not-allowed disabled:bg-[#c7c3d7]"
            :disabled="isCreatingGroup"
          >
            {{ isCreatingGroup ? "Membuat..." : "Buat ruang" }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
