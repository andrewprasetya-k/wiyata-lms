<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from "vue";
import {
  PhArrowClockwise,
  PhChatCircleText,
  PhPaperPlaneTilt,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import {
  createChatGroup,
  getChatRooms,
  getMessages,
  markRoomRead,
  openSchoolChatRoom,
  searchChatMembers,
  sendMessage,
} from "../../services/chat";
import type { ChatMember, ChatMessage, ChatRoom } from "../../types/chat";

defineProps<{
  audience: "student" | "teacher" | "admin";
}>();

const rooms = ref<ChatRoom[]>([]);
const selectedRoom = ref<ChatRoom | null>(null);
const messages = ref<ChatMessage[]>([]);
const nextBefore = ref<string | null>(null);
const hasMore = ref(false);
const draft = ref("");
const isBooting = ref(true);
const isLoadingMessages = ref(false);
const isLoadingOlder = ref(false);
const isSending = ref(false);
const isRefreshing = ref(false);
const accessError = ref("");
const threadError = ref("");
const composerError = ref("");
const messagesEl = ref<HTMLElement | null>(null);
const isCreateGroupOpen = ref(false);
const groupRoomName = ref("");
const memberSearch = ref("");
const memberResults = ref<ChatMember[]>([]);
const selectedMemberIds = ref<string[]>([]);
const isLoadingMembers = ref(false);
const isCreatingGroup = ref(false);
const createGroupError = ref("");
let poller: number | undefined;

const selectedRoomName = computed(
  () => selectedRoom.value?.roomName || "Ruang sekolah",
);
const selectedSchoolName = computed(
  () => selectedRoom.value?.schoolName || "Sekolah aktif",
);
const canSend = computed(
  () =>
    Boolean(selectedRoom.value?.canSend) &&
    draft.value.trim().length > 0 &&
    !isSending.value,
);
const roomInitial = computed(() => {
  const source = selectedRoomName.value || selectedSchoolName.value;
  return getInitials(source);
});
const schoolRooms = computed(() =>
  rooms.value.filter((room) => isSchoolRoom(room)),
);
const groupRooms = computed(() =>
  rooms.value.filter((room) => !isSchoolRoom(room)),
);

onMounted(async () => {
  await bootstrapChat();
  poller = window.setInterval(() => {
    if (selectedRoom.value && !isRefreshing.value && !isLoadingMessages.value) {
      refreshMessages({ silent: true });
    }
  }, 8000);
});

onUnmounted(() => {
  if (poller) {
    window.clearInterval(poller);
  }
});

async function bootstrapChat() {
  isBooting.value = true;
  accessError.value = "";
  threadError.value = "";
  try {
    const room = await openSchoolChatRoom();
    rooms.value = await getChatRooms();
    selectedRoom.value =
      rooms.value.find((item) => item.roomId === room.roomId) ?? room;
    await loadLatestMessages();
  } catch (error) {
    accessError.value = resolveChatError(error);
  } finally {
    isBooting.value = false;
  }
}

async function loadLatestMessages() {
  if (!selectedRoom.value) return;
  isLoadingMessages.value = true;
  threadError.value = "";
  try {
    const response = await getMessages(selectedRoom.value.roomId, {
      limit: 50,
    });
    messages.value = dedupeMessages(response.messages);
    nextBefore.value = response.nextBefore ?? null;
    hasMore.value = response.hasMore;
    await markSelectedRoomRead();
    await nextTick();
    scrollToBottom();
  } catch (error) {
    threadError.value = resolveChatError(error);
  } finally {
    isLoadingMessages.value = false;
  }
}

async function refreshMessages(options: { silent?: boolean } = {}) {
  if (!selectedRoom.value) return;
  if (!options.silent) {
    isRefreshing.value = true;
  }
  try {
    const response = await getMessages(selectedRoom.value.roomId, {
      limit: 50,
    });
    const previousLastId = lastMessage(messages.value)?.messageId;
    messages.value = dedupeMessages(response.messages);
    nextBefore.value = response.nextBefore ?? null;
    hasMore.value = response.hasMore;
    await refreshRooms();
    await markSelectedRoomRead();
    await nextTick();
    if (
      !previousLastId ||
      lastMessage(messages.value)?.messageId !== previousLastId
    ) {
      scrollToBottom();
    }
  } catch (error) {
    if (!options.silent) {
      threadError.value = resolveChatError(error);
    }
  } finally {
    isRefreshing.value = false;
  }
}

async function refreshRooms() {
  try {
    const latestRooms = await getChatRooms();
    rooms.value = latestRooms;
    if (selectedRoom.value) {
      selectedRoom.value =
        latestRooms.find(
          (room) => room.roomId === selectedRoom.value?.roomId,
        ) ?? selectedRoom.value;
    }
  } catch {
    // Room summary refresh is non-critical for the thread.
  }
}

async function openCreateGroupModal() {
  isCreateGroupOpen.value = true;
  createGroupError.value = "";
  if (memberResults.value.length === 0) {
    await loadChatMembers();
  }
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
    await refreshRooms();
    selectedRoom.value = room;
    groupRoomName.value = "";
    memberSearch.value = "";
    selectedMemberIds.value = [];
    isCreateGroupOpen.value = false;
    await loadLatestMessages();
  } catch (error) {
    createGroupError.value = resolveChatError(error);
  } finally {
    isCreatingGroup.value = false;
  }
}

async function loadOlderMessages() {
  if (!selectedRoom.value || !nextBefore.value || isLoadingOlder.value) return;
  const previousScrollHeight = messagesEl.value?.scrollHeight ?? 0;
  isLoadingOlder.value = true;
  threadError.value = "";
  try {
    const response = await getMessages(selectedRoom.value.roomId, {
      limit: 50,
      before: nextBefore.value,
    });
    messages.value = dedupeMessages([...response.messages, ...messages.value]);
    nextBefore.value = response.nextBefore ?? null;
    hasMore.value = response.hasMore;
    await nextTick();
    if (messagesEl.value) {
      messagesEl.value.scrollTop =
        messagesEl.value.scrollHeight - previousScrollHeight;
    }
  } catch (error) {
    threadError.value = resolveChatError(error);
  } finally {
    isLoadingOlder.value = false;
  }
}

async function submitMessage() {
  if (!selectedRoom.value) return;
  const content = draft.value.trim();
  if (!content) return;
  isSending.value = true;
  composerError.value = "";
  try {
    const created = await sendMessage(selectedRoom.value.roomId, content);
    messages.value = dedupeMessages([...messages.value, created]);
    draft.value = "";
    await markSelectedRoomRead(created.messageId);
    await refreshRooms();
    await nextTick();
    scrollToBottom();
  } catch (error) {
    composerError.value = resolveChatError(error);
  } finally {
    isSending.value = false;
  }
}

async function markSelectedRoomRead(lastReadMessageId?: string) {
  if (!selectedRoom.value) return;
  const latestMessageId =
    lastReadMessageId ?? lastMessage(messages.value)?.messageId;
  try {
    await markRoomRead(
      selectedRoom.value.roomId,
      latestMessageId ? { lastReadMessageId: latestMessageId } : {},
    );
  } catch {
    // Read receipt failure should not block chat usage.
  }
}

function handleComposerKeydown(event: KeyboardEvent) {
  if (event.key === "Enter" && !event.shiftKey) {
    event.preventDefault();
    submitMessage();
  }
}

function dedupeMessages(items: ChatMessage[]) {
  const seen = new Set<string>();
  const result: ChatMessage[] = [];
  for (const item of items) {
    if (!item.messageId || seen.has(item.messageId)) continue;
    seen.add(item.messageId);
    result.push(item);
  }
  return result;
}

function scrollToBottom() {
  if (messagesEl.value) {
    messagesEl.value.scrollTop = messagesEl.value.scrollHeight;
  }
}

function lastMessage(items: ChatMessage[]) {
  return items.length > 0 ? items[items.length - 1] : undefined;
}

function getInitials(value?: string | null) {
  return (
    value
      ?.split(/\s+/)
      .filter(Boolean)
      .slice(0, 2)
      .map((word) => word.charAt(0).toUpperCase())
      .join("") || "RS"
  );
}

function isSchoolRoom(room: ChatRoom) {
  return room.roomRefType === "school";
}

function roomDisplayName(room?: ChatRoom | null) {
  if (!room) return "Ruang Sekolah";
  return isSchoolRoom(room) ? "Ruang Sekolah" : room.roomName || "Ruang Grup";
}

function resolveChatError(error: unknown) {
  const maybeError = error as {
    response?: { status?: number; data?: { error?: string } };
  };
  if (
    maybeError.response?.status === 401 ||
    maybeError.response?.status === 403
  ) {
    return "Kamu tidak memiliki akses ke chat sekolah ini.";
  }
  return maybeError.response?.data?.error || "Gagal memuat chat. Coba lagi.";
}

function formatTime(value?: string | null) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return new Intl.DateTimeFormat("id-ID", {
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

function formatDateTime(value?: string | null) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return new Intl.DateTimeFormat("id-ID", {
    day: "2-digit",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-[#f8f7f4]">
    <section class="px-0 py-0 sm:px-0 lg:px-0">
      <div
        class="mx-auto flex max-h-screen min-h-screen max-w-screen flex-col gap-5"
      >
        <div
          v-if="isBooting"
          class="grid min-h-screen gap-4 overflow-hidden rounded-xl bg-white p-4 lg:grid-cols-[300px_minmax(0,1fr)]"
        >
          <div class="space-y-3 border-[#ebe7df] lg:border-r lg:pr-4">
            <div class="h-16 animate-pulse rounded-xl bg-[#f1eee8]" />
            <div class="h-20 animate-pulse rounded-xl bg-[#f8f7f4]" />
            <div class="h-20 animate-pulse rounded-xl bg-[#f8f7f4]" />
          </div>
          <div class="flex flex-col gap-3">
            <div class="h-16 animate-pulse rounded-xl bg-[#f1eee8]" />
            <div class="flex-1 animate-pulse rounded-xl bg-[#f8f7f4]" />
            <div class="h-16 animate-pulse rounded-xl bg-[#f1eee8]" />
          </div>
        </div>

        <div
          v-else-if="accessError"
          class="flex min-h-105 flex-col items-center justify-center rounded-xl bg-white px-6 py-12 text-center"
        >
          <div
            class="flex h-12 w-12 items-center justify-center rounded-lg bg-[#fef2f2] text-[#dc2626]"
          >
            <PhWarningCircle :size="24" weight="duotone" />
          </div>
          <h2 class="mt-4 text-base font-semibold text-[#171322]">
            Chat belum bisa dibuka
          </h2>
          <p class="mt-2 max-w-md text-sm leading-6 text-[#6b7280]">
            {{ accessError }}
          </p>
          <button
            type="button"
            class="mt-5 rounded-lg bg-[#171322] px-4 py-2 text-sm font-medium text-white transition hover:bg-[#374151]"
            @click="bootstrapChat"
          >
            Coba lagi
          </button>
        </div>

        <div
          v-else
          class="grid min-h-155 flex-1 overflow-hidden rounded-xl bg-white lg:grid-cols-[300px_minmax(0,1fr)]"
        >
          <aside class="min-w-0 border-[#ebe7df] bg-[#fbfaf8] lg:border-r">
            <div class="border-b border-[#ebe7df] bg-white px-4 py-4 sm:px-5">
              <div class="flex items-center justify-between gap-3">
                <p class="text-sm font-semibold text-[#171322]">
                  Ruang Diskusi
                </p>
                <button
                  type="button"
                  class="rounded-lg bg-[#4f46e5] px-3 py-1.5 text-xs font-semibold text-white transition hover:bg-[#4338ca]"
                  @click="openCreateGroupModal"
                >
                  Buat ruang
                </button>
              </div>
              <div
                class="mt-3 rounded-lg bg-[#f3f4f6] px-3 py-2 text-xs text-[#9ca3af]"
              >
                Cari pesan...
              </div>
            </div>

            <div class="space-y-2 p-4">
              <p
                class="text-[10px] font-semibold uppercase tracking-[0.06em] text-[#9ca3af]"
              >
                Sekolah
              </p>
              <button
                v-for="room in schoolRooms"
                :key="room.roomId"
                type="button"
                class="flex w-full min-w-0 items-center gap-3 rounded-lg border px-3 py-3 text-left transition hover:bg-white"
                :class="
                  selectedRoom?.roomId === room.roomId
                    ? 'border-[#d7d1ff] bg-white shadow-sm'
                    : 'border-[#ebe7df] bg-[#fbfaf8]'
                "
                @click="
                  selectedRoom = room;
                  loadLatestMessages();
                "
              >
                <span
                  class="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg bg-[#4f46e5] text-sm font-semibold text-white"
                >
                  {{ getInitials(room.schoolName || room.roomName) }}
                </span>
                <span class="min-w-0 flex-1">
                  <span
                    class="block truncate text-sm font-semibold text-[#171322]"
                  >
                    {{ roomDisplayName(room) }}
                  </span>
                  <span class="mt-0.5 block truncate text-xs text-[#6b7280]">
                    {{
                      room.lastMessage?.content ||
                      room.schoolName ||
                      "Belum ada pesan."
                    }}
                  </span>
                </span>
                <span class="flex shrink-0 flex-col items-end gap-1">
                  <span class="text-[11px] text-[#9ca3af]">{{
                    formatTime(room.lastMessageAt)
                  }}</span>
                  <span
                    v-if="room.unreadCount > 0"
                    class="rounded-full bg-[#4f46e5] px-2 py-0.5 text-[11px] font-semibold text-white"
                  >
                    {{ room.unreadCount }}
                  </span>
                </span>
              </button>

              <p
                class="pt-3 text-[10px] font-semibold uppercase tracking-[0.06em] text-[#9ca3af]"
              >
                Grup
              </p>
              <button
                v-for="room in groupRooms"
                :key="room.roomId"
                type="button"
                class="flex w-full min-w-0 items-center gap-3 rounded-lg border px-3 py-3 text-left transition hover:bg-white"
                :class="
                  selectedRoom?.roomId === room.roomId
                    ? 'border-[#d7d1ff] bg-white shadow-sm'
                    : 'border-[#ebe7df] bg-[#fbfaf8]'
                "
                @click="
                  selectedRoom = room;
                  loadLatestMessages();
                "
              >
                <span
                  class="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg bg-[#4f46e5] text-sm font-semibold text-white"
                >
                  {{ getInitials(room.roomName) }}
                </span>
                <span class="min-w-0 flex-1">
                  <span
                    class="block truncate text-sm font-semibold text-[#171322]"
                  >
                    {{ roomDisplayName(room) }}
                  </span>
                  <span class="mt-0.5 block truncate text-xs text-[#6b7280]">
                    {{ room.lastMessage?.content || "Belum ada pesan." }}
                  </span>
                </span>
                <span class="flex shrink-0 flex-col items-end gap-1">
                  <span class="text-[11px] text-[#9ca3af]">{{
                    formatTime(room.lastMessageAt)
                  }}</span>
                  <span
                    v-if="room.unreadCount > 0"
                    class="rounded-full bg-[#4f46e5] px-2 py-0.5 text-[11px] font-semibold text-white"
                  >
                    {{ room.unreadCount }}
                  </span>
                </span>
              </button>

              <div
                v-if="groupRooms.length === 0"
                class="rounded-lg border border-dashed border-[#d8d2c8] bg-white px-3 py-4 text-center text-xs text-[#6b7280]"
              >
                Belum ada grup khusus.
              </div>

              <div
                v-if="rooms.length === 0"
                class="rounded-lg border border-dashed border-[#d8d2c8] bg-white px-4 py-8 text-center"
              >
                <PhChatCircleText class="mx-auto h-8 w-8 text-[#b5aa9c]" />
                <p class="mt-3 text-sm font-semibold text-[#171322]">
                  Ruang belum tersedia
                </p>
                <p class="mt-1 text-xs text-[#6b7280]">
                  Buka ulang halaman untuk membuat ruang sekolah.
                </p>
              </div>
            </div>
          </aside>

          <section class="flex min-w-0 flex-col bg-[#f8f7f4]">
            <div
              class="flex items-center gap-3 border-b border-[#ebe7df] bg-white px-4 py-3 sm:px-5"
            >
              <div
                class="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg bg-[#4f46e5] text-sm font-semibold text-white"
              >
                {{ roomInitial }}
              </div>
              <div class="min-w-0 flex-1">
                <h2 class="truncate text-sm font-semibold text-[#171322]">
                  {{ roomDisplayName(selectedRoom) }}
                </h2>
                <p class="truncate text-xs text-[#6b7280]">
                  {{ selectedSchoolName }} ·
                  {{
                    selectedRoom && isSchoolRoom(selectedRoom)
                      ? "Ruang sekolah"
                      : "Grup"
                  }}
                </p>
              </div>
              <button
                type="button"
                class="inline-flex shrink-0 items-center justify-center gap-2 rounded-lg border border-[#ebe7df] bg-white px-3 py-2 text-xs font-medium text-[#4f46e5] transition hover:border-[#4f46e5] disabled:cursor-not-allowed disabled:opacity-60"
                :disabled="isRefreshing || !selectedRoom"
                @click="refreshMessages()"
              >
                <PhArrowClockwise
                  :class="['h-4 w-4', isRefreshing ? 'animate-spin' : '']"
                />
                Segarkan
              </button>
            </div>

            <div
              ref="messagesEl"
              class="min-h-0 flex-1 overflow-y-auto px-4 py-4 sm:px-5"
            >
              <div class="ml-auto mr-auto flex max-w-screen flex-col gap-3">
                <button
                  v-if="hasMore"
                  type="button"
                  class="mx-auto rounded-full border border-[#d8d2c8] bg-white px-4 py-2 text-xs font-semibold text-[#4f46e5] transition hover:border-[#4f46e5] disabled:opacity-60"
                  :disabled="isLoadingOlder"
                  @click="loadOlderMessages"
                >
                  {{ isLoadingOlder ? "Memuat..." : "Muat pesan sebelumnya" }}
                </button>

                <div v-if="isLoadingMessages" class="space-y-3">
                  <div class="h-12 w-2/3 animate-pulse rounded-2xl bg-white" />
                  <div
                    class="ml-auto h-12 w-1/2 animate-pulse rounded-2xl bg-[#dfe3ff]"
                  />
                  <div class="h-16 w-3/4 animate-pulse rounded-2xl bg-white" />
                </div>

                <div
                  v-else-if="threadError"
                  class="rounded-2xl border border-red-100 bg-white px-4 py-6 text-center"
                >
                  <p class="text-sm font-semibold text-red-600">
                    {{ threadError }}
                  </p>
                  <button
                    type="button"
                    class="mt-3 rounded-xl bg-[#4f46e5] px-4 py-2 text-sm font-semibold text-white"
                    @click="loadLatestMessages"
                  >
                    Coba lagi
                  </button>
                </div>

                <div
                  v-else-if="messages.length === 0"
                  class="flex min-h-80 flex-col items-center justify-center rounded-2xl px-6 text-center"
                >
                  <PhChatCircleText class="h-10 w-10 text-[#b5aa9c]" />
                  <h3 class="mt-4 text-base font-semibold text-[#171322]">
                    Belum ada pesan.
                  </h3>
                  <p class="mt-2 max-w-sm text-sm text-[#6b7280]">
                    Mulai percakapan pertama di ruang sekolah.
                  </p>
                </div>

                <template v-else>
                  <article
                    v-for="message in messages"
                    :key="message.messageId"
                    class="flex gap-2"
                    :class="message.isMine ? 'justify-end' : 'justify-start'"
                  >
                    <div
                      class="flex max-w-[88%] flex-col gap-1 sm:max-w-[72%]"
                      :class="message.isMine ? 'items-end' : 'items-start'"
                    >
                      <p
                        v-if="!message.isMine"
                        class="px-2 text-xs font-medium text-[#6b7280]"
                      >
                        {{ message.senderName }}
                      </p>
                      <div
                        class="rounded-2xl px-4 py-2 text-sm leading-relaxed shadow-sm"
                        :class="
                          message.isMine
                            ? 'rounded-br-md bg-[#4f46e5] text-white'
                            : 'rounded-bl-md border border-[#ebe7df] bg-white text-[#171322]'
                        "
                      >
                        <p class="whitespace-pre-wrap wrap-break-word">
                          {{ message.content }}
                        </p>
                      </div>
                      <p class="px-2 text-[11px] text-[#9ca3af]">
                        {{ formatDateTime(message.createdAt) }}
                      </p>
                    </div>
                  </article>
                </template>
              </div>
            </div>

            <form
              class="border-t border-[#ebe7df] bg-white px-4 py-3 sm:px-5"
              @submit.prevent="submitMessage"
            >
              <p v-if="composerError" class="mb-2 text-sm text-red-600">
                {{ composerError }}
              </p>
              <div class="flex items-end gap-2">
                <textarea
                  v-model="draft"
                  rows="1"
                  class="max-h-32 min-h-11 flex-1 resize-none rounded-xl border border-transparent bg-[#f3f4f6] px-4 py-3 text-sm text-[#171322] outline-none transition placeholder:text-[#aaa29a] focus:border-[#c7d2fe] focus:bg-white focus:ring-2 focus:ring-[#4f46e5]/15"
                  placeholder="Tulis pesan..."
                  :disabled="!selectedRoom?.canSend || isSending"
                  @keydown="handleComposerKeydown"
                />
                <button
                  type="submit"
                  class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-[#4f46e5] text-white transition hover:bg-[#4338ca] focus:outline-none focus:ring-2 focus:ring-[#4f46e5]/30 disabled:cursor-not-allowed disabled:bg-[#c7c3d7]"
                  :disabled="!canSend"
                  aria-label="Kirim pesan"
                >
                  <PhPaperPlaneTilt class="h-5 w-5" weight="fill" />
                </button>
              </div>
              <p class="mt-2 text-xs text-[#9ca3af]">
                Enter untuk kirim, Shift+Enter untuk baris baru.
              </p>
            </form>
          </section>
        </div>
      </div>
    </section>

    <div
      v-if="isCreateGroupOpen"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/30 px-4 py-6"
    >
      <div
        class="max-h-[90vh] w-full max-w-xl overflow-hidden rounded-xl bg-white shadow-xl"
      >
        <div class="border-b border-[#ebe7df] px-5 py-4">
          <h2 class="text-base font-semibold text-[#171322]">
            Buat ruang grup
          </h2>
          <p class="mt-1 text-sm text-[#6b7280]">
            Pilih warga aktif dari sekolah ini untuk membuat ruang diskusi
            khusus.
          </p>
        </div>

        <form
          class="flex max-h-[calc(90vh-5rem)] flex-col"
          @submit.prevent="submitCreateGroup"
        >
          <div class="space-y-4 overflow-y-auto px-5 py-4">
            <div>
              <label
                class="text-sm font-medium text-[#171322]"
                for="chat-group-name"
              >
                Nama ruang
              </label>
              <input
                id="chat-group-name"
                v-model="groupRoomName"
                type="text"
                class="mt-1 w-full rounded-lg border border-[#d8d2c8] px-3 py-2 text-sm text-[#171322] outline-none transition focus:border-[#4f46e5] focus:ring-2 focus:ring-[#4f46e5]/15"
                placeholder="Contoh: Grup Belajar Fisika"
              />
            </div>

            <div>
              <label
                class="text-sm font-medium text-[#171322]"
                for="chat-member-search"
              >
                Cari warga sekolah
              </label>
              <div class="mt-1 flex gap-2">
                <input
                  id="chat-member-search"
                  v-model="memberSearch"
                  type="text"
                  class="min-w-0 flex-1 rounded-lg border border-[#d8d2c8] px-3 py-2 text-sm text-[#171322] outline-none transition focus:border-[#4f46e5] focus:ring-2 focus:ring-[#4f46e5]/15"
                  placeholder="Cari nama atau email..."
                  @keydown.enter.prevent="loadChatMembers"
                />
                <button
                  type="button"
                  class="rounded-lg border border-[#d8d2c8] px-3 py-2 text-sm font-medium text-[#4f46e5] transition hover:border-[#4f46e5] disabled:opacity-60"
                  :disabled="isLoadingMembers"
                  @click="loadChatMembers"
                >
                  Cari
                </button>
              </div>
            </div>

            <p
              v-if="createGroupError"
              class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600"
            >
              {{ createGroupError }}
            </p>

            <div class="rounded-lg border border-[#ebe7df]">
              <div
                class="border-b border-[#ebe7df] bg-[#fbfaf8] px-3 py-2 text-xs font-semibold uppercase tracking-[0.06em] text-[#9ca3af]"
              >
                Anggota
              </div>
              <div v-if="isLoadingMembers" class="space-y-2 p-3">
                <div class="h-10 animate-pulse rounded-lg bg-[#f3f4f6]" />
                <div class="h-10 animate-pulse rounded-lg bg-[#f3f4f6]" />
              </div>
              <div
                v-else-if="memberResults.length === 0"
                class="px-3 py-8 text-center text-sm text-[#6b7280]"
              >
                Tidak ada warga yang cocok.
              </div>
              <div v-else class="max-h-64 overflow-y-auto p-2">
                <label
                  v-for="member in memberResults"
                  :key="member.userId"
                  class="flex cursor-pointer items-center gap-3 rounded-lg px-2 py-2 hover:bg-[#fbfaf8]"
                >
                  <input
                    type="checkbox"
                    class="h-4 w-4 rounded border-[#d8d2c8] text-[#4f46e5]"
                    :checked="selectedMemberIds.includes(member.userId)"
                    @change="toggleMember(member.userId)"
                  />
                  <span
                    class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-[#4f46e5] text-xs font-semibold text-white"
                  >
                    {{ getInitials(member.fullName || member.email) }}
                  </span>
                  <span class="min-w-0 flex-1">
                    <span
                      class="block truncate text-sm font-medium text-[#171322]"
                    >
                      {{ member.fullName || member.email }}
                    </span>
                    <span class="block truncate text-xs text-[#6b7280]">
                      {{ member.email }}
                    </span>
                  </span>
                </label>
              </div>
            </div>
          </div>

          <div
            class="flex flex-col gap-2 border-t border-[#ebe7df] px-5 py-4 sm:flex-row sm:justify-end"
          >
            <button
              type="button"
              class="rounded-lg border border-[#d8d2c8] px-4 py-2 text-sm font-medium text-[#6b7280] transition hover:bg-[#fbfaf8]"
              :disabled="isCreatingGroup"
              @click="isCreateGroupOpen = false"
            >
              Batal
            </button>
            <button
              type="submit"
              class="rounded-lg bg-[#4f46e5] px-4 py-2 text-sm font-semibold text-white transition hover:bg-[#4338ca] disabled:cursor-not-allowed disabled:bg-[#c7c3d7]"
              :disabled="isCreatingGroup"
            >
              {{ isCreatingGroup ? "Membuat..." : "Buat ruang" }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </main>
</template>
