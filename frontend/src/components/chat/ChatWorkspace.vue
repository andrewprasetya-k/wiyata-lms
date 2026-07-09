<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from "vue";
import {
  PhChatCircleText,
  PhCheck,
  PhChecks,
  PhDownloadSimple,
  PhFile,
  PhFileArchive,
  PhFileDoc,
  PhFilePdf,
  PhFilePpt,
  PhFileText,
  PhFileXls,
  PhImage,
  PhPaperclip,
  PhPaperPlaneTilt,
  PhSpinnerGap,
  PhWarningCircle,
  PhX,
  PhPlus,
} from "@phosphor-icons/vue";
import {
  addChatGroupMembers,
  createChatGroup,
  getChatRooms,
  getChatGroupInfo,
  getMessages,
  getRoomReadSummary,
  leaveChatGroup,
  markRoomRead,
  openDirectMessage,
  removeChatGroupMember,
  renameChatGroup,
  searchChatMembers,
  sendMessage,
} from "../../services/chat";
import { connectChatSocket } from "../../services/chatSocket";
import { deleteMedia, uploadMediaFile } from "../../services/media";
import type { ChatSocketStatus } from "../../services/chatSocket";
import { useAuthStore } from "../../stores/auth";
import { useToastStore } from "../../stores/toast";
import { useConfirmStore } from "../../stores/confirm";
import type {
  ChatAttachment,
  ChatGroupInfo,
  ChatGroupMember,
  ChatMember,
  ChatMessage,
  ChatSocketEvent,
  MessageReadEvent,
  NewMessageEvent,
  RoomUpdatedEvent,
  ChatReadSummary,
  ChatRoom,
} from "../../types/chat";
import {
  APP_TIME_ZONE,
  formatTime as formatBackendTime,
  parseBackendTimestamp,
} from "../../utils/date";

defineProps<{
  audience: "student" | "teacher" | "admin";
}>();

const toast = useToastStore();
const confirm = useConfirmStore();

interface PendingAttachmentItem {
  id: string;
  file: File;
  previewUrl: string;
}

const rooms = ref<ChatRoom[]>([]);
const selectedRoom = ref<ChatRoom | null>(null);
const messages = ref<ChatMessage[]>([]);
const readSummary = ref<ChatReadSummary | null>(null);
const nextBefore = ref<string | null>(null);
const hasMore = ref(false);
const draft = ref("");
const selectedFiles = ref<PendingAttachmentItem[]>([]);
const isBooting = ref(true);
const isLoadingMessages = ref(false);
const isLoadingOlder = ref(false);
const pendingSendCount = ref(0);
const isRefreshing = ref(false);
const isDragActive = ref(false);
const accessError = ref("");
const threadError = ref("");
const composerError = ref("");
const messagesEl = ref<HTMLElement | null>(null);
const roomListEl = ref<HTMLElement | null>(null);
const fileInputEl = ref<HTMLInputElement | null>(null);
const isLightboxOpen = ref(false);
const lightboxImage = ref<{ url: string; name: string } | null>(null);
const isCreateConversationOpen = ref(false);
const activeCreateTab = ref<"dm" | "group">("dm");
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
const roomSearch = ref("");
const isGroupInfoOpen = ref(false);
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
let pollTimeout: number | undefined;
let roomRefreshTimer: number | undefined;
let readSummaryRefreshTimer: number | undefined;
let socketConnection: { close: () => void } | null = null;
let isDestroyed = false;
const maxChatAttachments = 5;
const maxChatAttachmentSizeMb = 10;
const authStore = useAuthStore();
const socketStatus = ref<ChatSocketStatus>("disconnected");
const showJumpToLatest = ref(false);

const selectedRoomName = computed(() => roomDisplayName(selectedRoom.value));
const selectedSchoolName = computed(
  () => selectedRoom.value?.schoolName || "Sekolah aktif",
);
const canSend = computed(
  () =>
    Boolean(selectedRoom.value?.canSend) &&
    (draft.value.trim().length > 0 || selectedFiles.value.length > 0),
);
const composerStatusLabel = computed(() => {
  const message = lastMessage(messages.value);
  if (!message?.isMine) return "";
  if (message.deliveryStatus === "uploading") return "Mengunggah lampiran...";
  if (message.deliveryStatus === "sending") return "Mengirim pesan...";
  return "";
});
const roomInitial = computed(() => {
  const source = selectedRoomName.value || selectedSchoolName.value;
  return getInitials(source);
});
const currentUserId = computed(() => authStore.user?.id || "");
const conversationList = computed(() =>
  rooms.value
    .filter(
      (room) =>
        (isDirectMessageRoom(room) || isCustomGroupRoom(room)) &&
        roomMatchesSearch(room),
    )
    .slice()
    .sort((a, b) => {
      const aTime = a.lastMessageAt ? new Date(a.lastMessageAt).getTime() : 0;
      const bTime = b.lastMessageAt ? new Date(b.lastMessageAt).getTime() : 0;
      return bTime - aTime;
    }),
);
const selectedRoomIsGroup = computed(() =>
  selectedRoom.value ? isCustomGroupRoom(selectedRoom.value) : false,
);
const selectedRoomIsDM = computed(() =>
  selectedRoom.value ? isDirectMessageRoom(selectedRoom.value) : false,
);
const currentUserIsGroupAdmin = computed(() =>
  Boolean(
    groupInfo.value?.admins.some(
      (member) => member.userId === currentUserId.value,
    ),
  ),
);
const selectedAddMembers = computed(() =>
  addMemberResults.value.filter((member) =>
    selectedAddMemberIds.value.includes(member.userId),
  ),
);
const messageGroupGapMs = 5 * 60 * 1000;

onMounted(async () => {
  isDestroyed = false;
  await bootstrapChat();
  connectRealtimeChat();
  scheduleNextPoll();
  document.addEventListener("visibilitychange", handleVisibilityChange);
  document.addEventListener("keydown", handleGlobalKeydown);
});

onUnmounted(() => {
  isDestroyed = true;
  if (pollTimeout) {
    window.clearTimeout(pollTimeout);
  }
  if (roomRefreshTimer) {
    window.clearTimeout(roomRefreshTimer);
  }
  if (readSummaryRefreshTimer) {
    window.clearTimeout(readSummaryRefreshTimer);
  }
  socketConnection?.close();
  revokePendingAttachmentUrls();
  revokeMessageAttachmentUrls();
  document.removeEventListener("visibilitychange", handleVisibilityChange);
  document.removeEventListener("keydown", handleGlobalKeydown);
});

async function bootstrapChat() {
  isBooting.value = true;
  accessError.value = "";
  threadError.value = "";
  try {
    rooms.value = await getChatRooms(roomSearch.value.trim());
    selectedRoom.value = rooms.value[0] ?? null;
    if (selectedRoom.value) {
      await loadLatestMessages();
    }
  } catch (error) {
    accessError.value = resolveChatError(error);
  } finally {
    isBooting.value = false;
  }
}

async function loadLatestMessages() {
  if (!selectedRoom.value) return;
  const roomId = selectedRoom.value.roomId;
  isLoadingMessages.value = true;
  threadError.value = "";
  showJumpToLatest.value = false;
  try {
    const response = await getMessages(selectedRoom.value.roomId, {
      limit: 50,
    });
    if (selectedRoom.value?.roomId !== roomId) return;
    messages.value = dedupeMessages(response.messages);
    nextBefore.value = response.nextBefore ?? null;
    hasMore.value = response.hasMore;
    await markSelectedRoomRead();
    await refreshReadSummary();
    await nextTick();
    scrollToBottom();
  } catch (error) {
    if (selectedRoom.value?.roomId !== roomId) return;
    threadError.value = resolveChatError(error);
  } finally {
    if (selectedRoom.value?.roomId === roomId) {
      isLoadingMessages.value = false;
    }
  }
}

async function refreshMessages(options: { silent?: boolean } = {}) {
  if (!selectedRoom.value) return;
  const wasNearBottom = isNearBottom();
  if (!options.silent) {
    isRefreshing.value = true;
  }
  try {
    const response = await getMessages(selectedRoom.value.roomId, {
      limit: 50,
    });
    const previousLastId = lastMessage(messages.value)?.messageId;
    const latestMessageId = lastMessage(response.messages)?.messageId;
    messages.value = dedupeMessages(response.messages);
    nextBefore.value = response.nextBefore ?? null;
    hasMore.value = response.hasMore;
    await refreshRooms();
    if (
      wasNearBottom &&
      (lastMessage(messages.value)?.messageId !== previousLastId ||
        (selectedRoom.value?.unreadCount ?? 0) > 0)
    ) {
      await markSelectedRoomRead(latestMessageId);
    }
    await refreshReadSummary();
    await nextTick();
    if (
      !previousLastId ||
      lastMessage(messages.value)?.messageId !== previousLastId
    ) {
      if (wasNearBottom) {
        scrollToBottom();
      } else {
        showJumpToLatest.value = true;
      }
    }
  } catch (error) {
    if (!options.silent) {
      threadError.value = resolveChatError(error);
    }
  } finally {
    isRefreshing.value = false;
  }
}

function scheduleNextPoll() {
  if (isDestroyed) return;
  if (pollTimeout) {
    window.clearTimeout(pollTimeout);
  }
  const delay = socketStatus.value === "connected" ? 90000 : 18000;
  pollTimeout = window.setTimeout(async () => {
    if (selectedRoom.value && !isRefreshing.value && !isLoadingMessages.value) {
      await refreshMessages({ silent: true });
    }
    scheduleNextPoll();
  }, delay);
}

async function handleVisibilityChange() {
  if (document.visibilityState !== "visible") return;
  await refreshRooms();
  if (selectedRoom.value) {
    await refreshMessages({ silent: true });
  }
}

async function refreshRooms() {
  const previousScrollTop = roomListEl.value?.scrollTop ?? 0;
  try {
    const latestRooms = await getChatRooms(roomSearch.value.trim());
    rooms.value = latestRooms;
    if (selectedRoom.value) {
      selectedRoom.value =
        latestRooms.find(
          (room) => room.roomId === selectedRoom.value?.roomId,
        ) ?? selectedRoom.value;
    }
    await nextTick();
    if (roomListEl.value) {
      roomListEl.value.scrollTop = previousScrollTop;
    }
  } catch {
    // Room summary refresh is non-critical for the thread.
  }
}

async function searchRooms() {
  await refreshRooms();
}

async function openCreateConversation(tab: "dm" | "group" = "dm") {
  activeCreateTab.value = tab;
  isCreateConversationOpen.value = true;
  if (tab === "group" && memberResults.value.length === 0) {
    await loadChatMembers();
  } else if (tab === "dm" && dmResults.value.length === 0) {
    directMessageError.value = "";
    selectedDMTargetId.value = "";
    await loadDMTargets();
  }
}

watch(activeCreateTab, async (tab) => {
  if (!isCreateConversationOpen.value) return;
  if (tab === "group" && memberResults.value.length === 0) {
    await loadChatMembers();
  } else if (tab === "dm" && dmResults.value.length === 0) {
    directMessageError.value = "";
    selectedDMTargetId.value = "";
    await loadDMTargets();
  }
});

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
      (member) => member.userId !== currentUserId.value,
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
    await refreshRooms();
    selectedRoom.value = room;
    groupRoomName.value = "";
    memberSearch.value = "";
    selectedMemberIds.value = [];
    isCreateConversationOpen.value = false;
    await loadLatestMessages();
    toast.success("Ruang chat berhasil dibuat.");
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
    await refreshRooms();
    selectedRoom.value =
      rooms.value.find((item) => item.roomId === room.roomId) ?? room;
    isCreateConversationOpen.value = false;
    dmSearch.value = "";
    selectedDMTargetId.value = "";
    await loadLatestMessages();
  } catch (error) {
    directMessageError.value = resolveChatError(error);
  } finally {
    isOpeningDM.value = false;
  }
}

async function openGroupInfo() {
  if (!selectedRoom.value || !isCustomGroupRoom(selectedRoom.value)) return;
  isGroupInfoOpen.value = true;
  await loadGroupInfo();
}

async function loadGroupInfo() {
  if (!selectedRoom.value || !isCustomGroupRoom(selectedRoom.value)) return;
  isLoadingGroupInfo.value = true;
  groupInfoError.value = "";
  groupActionError.value = "";
  try {
    groupInfo.value = await getChatGroupInfo(selectedRoom.value.roomId);
    renameRoomName.value = groupInfo.value.roomName;
    if (
      groupInfo.value.admins.some(
        (member) => member.userId === currentUserId.value,
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
  if (!selectedRoom.value || !groupInfo.value) return;
  const roomName = renameRoomName.value.trim();
  if (roomName.length < 3) {
    groupActionError.value = "Nama ruang minimal 3 karakter.";
    return;
  }
  isRenamingGroup.value = true;
  groupActionError.value = "";
  try {
    const room = await renameChatGroup(selectedRoom.value.roomId, { roomName });
    selectedRoom.value = room;
    groupInfo.value = { ...groupInfo.value, roomName: room.roomName };
    await refreshRooms();
    toast.success("Nama ruang chat berhasil diperbarui.");
  } catch (error) {
    groupActionError.value = resolveChatError(error);
  } finally {
    isRenamingGroup.value = false;
  }
}

async function loadEligibleMembers() {
  if (!selectedRoom.value || !isCustomGroupRoom(selectedRoom.value)) return;
  isLoadingAddMembers.value = true;
  groupActionError.value = "";
  try {
    addMemberResults.value = await searchChatMembers(
      addMemberSearch.value.trim(),
      selectedRoom.value.roomId,
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
  if (!selectedRoom.value || selectedAddMemberIds.value.length === 0) {
    groupActionError.value = "Pilih minimal satu anggota.";
    return;
  }
  isAddingMembers.value = true;
  groupActionError.value = "";
  try {
    await addChatGroupMembers(selectedRoom.value.roomId, {
      memberUserIds: selectedAddMemberIds.value,
    });
    selectedAddMemberIds.value = [];
    addMemberSearch.value = "";
    await loadGroupInfo();
    await refreshRooms();
    toast.success("Anggota berhasil ditambahkan.");
  } catch (error) {
    groupActionError.value = resolveChatError(error);
  } finally {
    isAddingMembers.value = false;
  }
}

async function leaveSelectedGroup() {
  if (!selectedRoom.value || !isCustomGroupRoom(selectedRoom.value)) return;
  const ok = await confirm.confirm({
    title: `Keluar dari ${roomDisplayName(selectedRoom.value)}?`,
    description: "Kamu tidak akan bisa mengakses pesan grup ini lagi.",
    confirmLabel: "Keluar",
    variant: "warning",
  });
  if (!ok) return;

  const previousRoomID = selectedRoom.value.roomId;
  isLeavingGroup.value = true;
  groupActionError.value = "";
  try {
    await leaveChatGroup(previousRoomID);
    isGroupInfoOpen.value = false;
    groupInfo.value = null;
    await refreshRooms();
    selectedRoom.value = rooms.value[0] ?? null;
    messages.value = [];
    if (selectedRoom.value) {
      await loadLatestMessages();
    }
    toast.success("Kamu berhasil keluar dari ruang chat.");
  } catch (error) {
    groupActionError.value = resolveChatError(error);
  } finally {
    isLeavingGroup.value = false;
  }
}

async function removeMember(member: ChatGroupMember) {
  if (!selectedRoom.value) return;
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
    await removeChatGroupMember(selectedRoom.value.roomId, member.userId);
    await loadGroupInfo();
    await refreshRooms();
    toast.success("Anggota berhasil dikeluarkan.");
  } catch (error) {
    groupActionError.value = resolveChatError(error);
  } finally {
    removingMemberId.value = null;
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

function openFilePicker() {
  fileInputEl.value?.click();
}

function handleFileSelection(event: Event) {
  const input = event.target as HTMLInputElement;
  const files = Array.from(input.files ?? []);
  appendSelectedFiles(files);
  input.value = "";
}

function appendSelectedFiles(files: File[]) {
  if (files.length === 0) return;
  const availableSlots = maxChatAttachments - selectedFiles.value.length;
  if (availableSlots <= 0) {
    composerError.value = `Maksimal ${maxChatAttachments} file per pesan.`;
    return;
  }
  if (files.length > availableSlots) {
    composerError.value = `Maksimal ${maxChatAttachments} file per pesan.`;
  } else {
    composerError.value = "";
  }
  const validFiles = files.filter((file) => {
    if (file.size > maxChatAttachmentSizeMb * 1024 * 1024) {
      composerError.value = `File ${file.name} melebihi batas ${maxChatAttachmentSizeMb}MB.`;
      return false;
    }
    return true;
  });
  selectedFiles.value = [
    ...selectedFiles.value,
    ...validFiles.slice(0, availableSlots).map((file) => ({
      id: `pending-${crypto.randomUUID?.() || `${Date.now()}-${Math.random()}`}`,
      file,
      previewUrl: isSafeImageType(file.type) ? URL.createObjectURL(file) : "",
    })),
  ];
}

function handleDragOver(event: DragEvent) {
  if (!selectedRoom.value?.canSend) return;
  if (!event.dataTransfer?.types.includes("Files")) return;
  event.preventDefault();
  event.dataTransfer.dropEffect = "copy";
  isDragActive.value = true;
}

function handleDragEnter(event: DragEvent) {
  if (!selectedRoom.value?.canSend) return;
  if (!event.dataTransfer?.types.includes("Files")) return;
  event.preventDefault();
  isDragActive.value = true;
}

function handleDragLeave(event: DragEvent) {
  const nextTarget = event.relatedTarget as Node | null;
  const currentTarget = event.currentTarget as HTMLElement | null;
  if (nextTarget && currentTarget?.contains(nextTarget)) return;
  isDragActive.value = false;
}

function handleDropFiles(event: DragEvent) {
  if (!selectedRoom.value?.canSend) return;
  if (!event.dataTransfer?.files?.length) return;
  event.preventDefault();
  isDragActive.value = false;
  appendSelectedFiles(Array.from(event.dataTransfer.files));
}

function removeSelectedFile(index: number) {
  const attachment = selectedFiles.value[index];
  if (attachment?.previewUrl) {
    URL.revokeObjectURL(attachment.previewUrl);
  }
  selectedFiles.value = selectedFiles.value.filter(
    (_, itemIndex) => itemIndex !== index,
  );
}

async function submitMessage() {
  if (!selectedRoom.value) return;
  const content = draft.value.trim();
  const attachmentsToSend = [...selectedFiles.value];
  const filesToSend = attachmentsToSend.map((attachment) => attachment.file);
  if (!content && filesToSend.length === 0) return;
  const roomID = selectedRoom.value.roomId;
  const optimistic = createOptimisticMessage(
    roomID,
    content,
    attachmentsToSend,
  );
  messages.value = dedupeMessages([...messages.value, optimistic]);
  draft.value = "";
  revokePendingAttachmentUrls();
  selectedFiles.value = [];
  pendingSendCount.value += 1;
  composerError.value = "";
  await nextTick();
  scrollToBottom();
  let uploadedMediaIds: string[] = [];
  try {
    markMessageAsUploading(optimistic.messageId, filesToSend.length > 0);
    const mediaIds = await uploadChatFiles(filesToSend);
    uploadedMediaIds = mediaIds;
    markMessageAsSending(optimistic.messageId);
    const created = await sendMessage(roomID, {
      content,
      mediaIds,
    });
    replaceOptimisticMessage(optimistic.messageId, {
      ...created,
      attachments: created.attachments ?? [],
      deliveryStatus: "sent",
    });
    await markSelectedRoomRead(created.messageId);
    await refreshReadSummary();
    await refreshRooms();
    await nextTick();
    scrollToBottom();
  } catch (error) {
    await cleanupUploadedMedia(uploadedMediaIds);
    markOptimisticFailed(optimistic.messageId);
    composerError.value = resolveChatError(error);
  } finally {
    pendingSendCount.value = Math.max(0, pendingSendCount.value - 1);
  }
}

async function uploadChatFiles(files: File[]) {
  if (!selectedRoom.value || files.length === 0) return [];
  if (files.length > maxChatAttachments) {
    throw new Error(`Maksimal ${maxChatAttachments} file per pesan.`);
  }
  const mediaIds: string[] = [];
  for (const file of files) {
    const uploaded = await uploadMediaFile(
      file,
      selectedRoom.value.schoolId,
      "user",
    );
    mediaIds.push(uploaded.mediaId);
  }
  return mediaIds;
}

async function cleanupUploadedMedia(mediaIds: string[]) {
  if (mediaIds.length === 0) return;
  await Promise.allSettled(mediaIds.map((mediaId) => deleteMedia(mediaId)));
}

function markMessageAsUploading(messageId: string, hasAttachments: boolean) {
  if (!hasAttachments) {
    markMessageAsSending(messageId);
    return;
  }
  messages.value = messages.value.map((message) =>
    message.messageId === messageId
      ? { ...message, deliveryStatus: "uploading" }
      : message,
  );
}

function markMessageAsSending(messageId: string) {
  messages.value = messages.value.map((message) =>
    message.messageId === messageId
      ? { ...message, deliveryStatus: "sending" }
      : message,
  );
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

async function refreshReadSummary() {
  if (!selectedRoom.value) {
    readSummary.value = null;
    return;
  }
  try {
    readSummary.value = await getRoomReadSummary(selectedRoom.value.roomId);
  } catch {
    // Read summary should not block chat usage.
  }
}

function connectRealtimeChat() {
  socketConnection?.close();
  socketConnection = connectChatSocket({
    onEvent: handleRealtimeEvent,
    onStatusChange(status) {
      if (isDestroyed) return;
      socketStatus.value = status;
      scheduleNextPoll();
    },
  });
}

function handleRealtimeEvent(event: ChatSocketEvent) {
  if (event.type === "new_message") {
    void handleNewMessageEvent(event as NewMessageEvent);
    return;
  }
  if (event.type === "message_read") {
    void handleMessageReadEvent(event as MessageReadEvent);
    return;
  }
  if (event.type === "room_updated") {
    handleRoomUpdatedEvent(event as RoomUpdatedEvent);
  }
}

async function handleNewMessageEvent(event: NewMessageEvent) {
  const message = event.payload;
  if (!message?.messageId) return;

  if (selectedRoom.value?.roomId === event.roomId) {
    const wasNearBottom = isNearBottom();
    messages.value = dedupeMessages([...messages.value, message]);
    if (
      !message.isMine &&
      document.visibilityState === "visible" &&
      wasNearBottom
    ) {
      await markSelectedRoomRead(message.messageId);
    }
    await refreshReadSummary();
    await nextTick();
    if (message.isMine || wasNearBottom) {
      scrollToBottom();
    } else {
      showJumpToLatest.value = true;
    }
  }
  scheduleRoomRefresh();
}

async function handleMessageReadEvent(event: MessageReadEvent) {
  if (selectedRoom.value?.roomId === event.roomId) {
    scheduleReadSummaryRefresh();
  }
  scheduleRoomRefresh();
}

function handleRoomUpdatedEvent(event: RoomUpdatedEvent) {
  scheduleRoomRefresh();
  if (selectedRoom.value?.roomId === event.roomId) {
    scheduleReadSummaryRefresh();
  }
}

function createOptimisticMessage(
  roomId: string,
  content: string,
  attachments: PendingAttachmentItem[] = [],
): ChatMessage {
  return {
    messageId: `temp-${Date.now()}-${Math.random().toString(36).slice(2)}`,
    roomId,
    senderId: currentUserId.value,
    senderName: authStore.user?.fullName || "Anda",
    senderRole: authStore.activeRoles[0] || "member",
    content,
    messageType: attachments.length > 0 ? "file" : "text",
    attachments: attachments.map((attachment, index) => ({
      attachmentId: `temp-${index}-${attachment.file.name}`,
      mediaId: "",
      fileName: attachment.file.name,
      mimeType: attachment.file.type || "application/octet-stream",
      sizeBytes: attachment.file.size,
      url: attachment.previewUrl,
    })),
    createdAt: new Date().toISOString(),
    isMine: true,
    deliveryStatus: attachments.length > 0 ? "uploading" : "sending",
    pendingFiles: attachments.map((attachment) => attachment.file),
  };
}

function replaceOptimisticMessage(tempId: string, canonical: ChatMessage) {
  const existing = messages.value.find(
    (message) => message.messageId === tempId,
  );
  if (existing) {
    revokeAttachmentObjectUrls(existing);
  }
  messages.value = dedupeMessages(
    messages.value.map((message) =>
      message.messageId === tempId ? canonical : message,
    ),
  );
}

function markOptimisticFailed(tempId: string) {
  messages.value = messages.value.map((message) =>
    message.messageId === tempId
      ? {
          ...message,
          deliveryStatus: "failed",
        }
      : message,
  );
}

function retryFailedMessage(message: ChatMessage) {
  messages.value = messages.value.filter(
    (item) => item.messageId !== message.messageId,
  );
  void sendRetriedMessage(message);
}

async function sendRetriedMessage(message: ChatMessage) {
  const files = message.pendingFiles ?? [];
  const optimistic = createOptimisticMessage(
    message.roomId,
    message.content,
    files.map((file) => ({
      id: `retry-${crypto.randomUUID?.() || `${Date.now()}-${Math.random()}`}`,
      file,
      previewUrl: isSafeImageType(file.type) ? URL.createObjectURL(file) : "",
    })),
  );
  messages.value = dedupeMessages([...messages.value, optimistic]);
  pendingSendCount.value += 1;
  composerError.value = "";
  await nextTick();
  scrollToBottom();
  let uploadedMediaIds: string[] = [];
  try {
    markMessageAsUploading(optimistic.messageId, files.length > 0);
    const mediaIds = await uploadChatFiles(files);
    uploadedMediaIds = mediaIds;
    markMessageAsSending(optimistic.messageId);
    const created = await sendMessage(message.roomId, {
      content: message.content,
      mediaIds,
    });
    replaceOptimisticMessage(optimistic.messageId, {
      ...created,
      attachments: created.attachments ?? [],
      deliveryStatus: "sent",
    });
    await markSelectedRoomRead(created.messageId);
    await refreshReadSummary();
    await refreshRooms();
  } catch (error) {
    await cleanupUploadedMedia(uploadedMediaIds);
    markOptimisticFailed(optimistic.messageId);
    composerError.value = resolveChatError(error);
  } finally {
    pendingSendCount.value = Math.max(0, pendingSendCount.value - 1);
  }
}

function scheduleRoomRefresh() {
  if (roomRefreshTimer) {
    window.clearTimeout(roomRefreshTimer);
  }
  roomRefreshTimer = window.setTimeout(() => {
    void refreshRooms();
  }, 150);
}

function scheduleReadSummaryRefresh() {
  if (readSummaryRefreshTimer) {
    window.clearTimeout(readSummaryRefreshTimer);
  }
  readSummaryRefreshTimer = window.setTimeout(() => {
    void refreshReadSummary();
  }, 150);
}

function handleComposerKeydown(event: KeyboardEvent) {
  if (event.key === "Enter" && !event.shiftKey) {
    event.preventDefault();
    submitMessage();
  }
}

function dedupeMessages(items: ChatMessage[]) {
  const seen = new Set<string>();
  const canonicalKeys = new Set(
    items
      .filter((item) => !item.messageId.startsWith("temp-"))
      .map((item) => messageContentKey(item)),
  );
  const result: ChatMessage[] = [];
  for (const item of items) {
    if (!item.messageId || seen.has(item.messageId)) continue;
    if (
      item.messageId.startsWith("temp-") &&
      item.deliveryStatus !== "failed" &&
      canonicalKeys.has(messageContentKey(item))
    ) {
      continue;
    }
    seen.add(item.messageId);
    result.push(item);
  }
  return result;
}

function messageContentKey(message: ChatMessage) {
  const attachmentKey = (message.attachments ?? [])
    .map(
      (attachment) =>
        attachment.mediaId || `${attachment.fileName}:${attachment.sizeBytes}`,
    )
    .join(",");
  return `${message.roomId}:${message.senderId}:${message.content}:${attachmentKey}`;
}

function scrollToBottom() {
  if (messagesEl.value) {
    messagesEl.value.scrollTop = messagesEl.value.scrollHeight;
  }
  showJumpToLatest.value = false;
}

function isNearBottom(threshold = 96) {
  const element = messagesEl.value;
  if (!element) return true;
  const remaining =
    element.scrollHeight - element.scrollTop - element.clientHeight;
  return remaining <= threshold;
}

function handleMessageScroll() {
  if (isNearBottom()) {
    showJumpToLatest.value = false;
    if (
      document.visibilityState === "visible" &&
      selectedRoom.value &&
      (selectedRoom.value.unreadCount ?? 0) > 0
    ) {
      void markSelectedRoomRead(lastMessage(messages.value)?.messageId);
      void refreshRooms();
      void refreshReadSummary();
    }
  }
}

function lastMessage(items: ChatMessage[]) {
  return items.length > 0 ? items[items.length - 1] : undefined;
}

function previousMessage(index: number) {
  return index > 0 ? messages.value[index - 1] : null;
}

function nextMessage(index: number) {
  return index < messages.value.length - 1 ? messages.value[index + 1] : null;
}

function messageTimeGapExceeded(
  currentMessage: ChatMessage,
  compareMessage: ChatMessage | null,
) {
  if (!compareMessage) return true;
  const currentTime =
    parseBackendTimestamp(currentMessage.createdAt)?.getTime() ?? Number.NaN;
  const previousTime =
    parseBackendTimestamp(compareMessage.createdAt)?.getTime() ?? Number.NaN;
  if (Number.isNaN(currentTime) || Number.isNaN(previousTime)) return true;
  return Math.abs(currentTime - previousTime) > messageGroupGapMs;
}

function isSameDay(left?: string | null, right?: string | null) {
  if (!left || !right) return false;
  const leftDate = parseBackendTimestamp(left);
  const rightDate = parseBackendTimestamp(right);
  if (
    !leftDate ||
    !rightDate ||
    Number.isNaN(leftDate.getTime()) ||
    Number.isNaN(rightDate.getTime())
  ) {
    return false;
  }
  return (
    leftDate.getFullYear() === rightDate.getFullYear() &&
    leftDate.getMonth() === rightDate.getMonth() &&
    leftDate.getDate() === rightDate.getDate()
  );
}

function shouldShowDateDivider(message: ChatMessage, index: number) {
  const previous = previousMessage(index);
  if (!previous) return true;
  return !isSameDay(message.createdAt, previous.createdAt);
}

function formatDateDivider(value?: string | null) {
  if (!value) return "";
  const date = parseBackendTimestamp(value);
  if (!date || Number.isNaN(date.getTime())) return "";
  const today = new Date();
  const yesterday = new Date();
  yesterday.setDate(today.getDate() - 1);
  if (isSameDay(value, today.toISOString())) return "Hari Ini";
  if (isSameDay(value, yesterday.toISOString())) return "Kemarin";
  return new Intl.DateTimeFormat("id-ID", {
    weekday: "long",
    day: "numeric",
    month: "long",
    timeZone: APP_TIME_ZONE,
  }).format(date);
}

function shouldShowSender(message: ChatMessage, index: number) {
  if (selectedRoomIsDM.value || message.isMine) return false;
  const previous = previousMessage(index);
  if (!previous) return true;
  if (previous.senderId !== message.senderId) return true;
  if (!isSameDay(message.createdAt, previous.createdAt)) return true;
  return messageTimeGapExceeded(message, previous);
}

function isGroupedWithPrevious(message: ChatMessage, index: number) {
  const previous = previousMessage(index);
  if (!previous) return false;
  if (previous.senderId !== message.senderId) return false;
  if (previous.isMine !== message.isMine) return false;
  if (!isSameDay(message.createdAt, previous.createdAt)) return false;
  return !messageTimeGapExceeded(message, previous);
}

function isGroupedWithNext(message: ChatMessage, index: number) {
  const next = nextMessage(index);
  if (!next) return false;
  if (next.senderId !== message.senderId) return false;
  if (next.isMine !== message.isMine) return false;
  if (!isSameDay(message.createdAt, next.createdAt)) return false;
  return !messageTimeGapExceeded(message, next);
}

function readIndicatorFor(message: ChatMessage) {
  if (
    !message.isMine ||
    message.deliveryStatus === "uploading" ||
    message.deliveryStatus === "sending" ||
    message.deliveryStatus === "failed"
  ) {
    return "";
  }
  const count = readByOtherCount(message);
  if (selectedRoomIsDM.value) {
    return count > 0 ? "Dibaca" : "Terkirim";
  }
  return count > 0 ? `Dibaca ${count} orang` : "Terkirim";
}

function readIndicatorLabel(message: ChatMessage) {
  const label = readIndicatorFor(message);
  return label || undefined;
}

function isReadByOthers(message: ChatMessage) {
  return readByOtherCount(message) > 0;
}

function readIndicatorCount(message: ChatMessage) {
  if (selectedRoomIsDM.value) return "";
  const count = readByOtherCount(message);
  return count > 0 ? String(count) : "";
}

function readByOtherCount(message: ChatMessage) {
  if (!readSummary.value?.members.length) return 0;
  const messageCreatedAt =
    parseBackendTimestamp(message.createdAt)?.getTime() ?? Number.NaN;
  return readSummary.value.members.filter((member) => {
    if (member.userId === currentUserId.value) return false;
    if (member.lastReadMessageId === message.messageId) return true;
    if (!member.lastReadAt || Number.isNaN(messageCreatedAt)) return false;
    const lastReadAt =
      parseBackendTimestamp(member.lastReadAt)?.getTime() ?? Number.NaN;
    return !Number.isNaN(lastReadAt) && lastReadAt >= messageCreatedAt;
  }).length;
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

// function isSchoolRoom(room: ChatRoom) {
//   return room.roomRefType === "school";
// }

function isDirectMessageRoom(room: ChatRoom) {
  return room.roomType === "dm";
}

function isCustomGroupRoom(room: ChatRoom) {
  return room.roomType === "group" && room.roomRefType == null;
}

function roomDisplayName(room?: ChatRoom | null) {
  if (!room) return "Ruang Sekolah";
  if (isDirectMessageRoom(room)) {
    return room.dmTargetName || room.dmTargetEmail || "Direct Message";
  }
  return room.roomName || "Ruang Grup";
}

function roomSubtitle(room?: ChatRoom | null) {
  if (!room) return "Ruang sekolah";
  if (isDirectMessageRoom(room)) {
    return room.dmTargetEmail || "Direct Message";
  }
  return "Grup";
}

function roomMatchesSearch(room: ChatRoom) {
  const query = roomSearch.value.trim().toLowerCase();
  if (!query) return true;
  const haystack = [
    roomDisplayName(room),
    room.roomName,
    room.schoolName,
    room.dmTargetName,
    room.dmTargetEmail,
    room.lastMessage?.content,
  ]
    .filter(Boolean)
    .join(" ")
    .toLowerCase();
  return haystack.includes(query);
}

function roomPreview(room: ChatRoom) {
  if (!room.lastMessage) {
    return isDirectMessageRoom(room)
      ? room.dmTargetEmail || "Belum ada pesan."
      : room.schoolName || "Belum ada pesan.";
  }
  const content = roomPreviewContent(room);
  if (isDirectMessageRoom(room)) {
    return content;
  }
  if (room.lastMessage.senderId === currentUserId.value) {
    return roomPreviewReadPrefix(room) + content;
  }
  const senderName = room.lastMessage.senderName || "Pengguna";
  return `${senderName}: ${content}`;
}

function attachmentPreviewTextForRoom(
  count: number,
  mimeType?: string,
  fileName?: string,
) {
  if (count <= 0) return "Mengirim file";
  if (count === 1) {
    if (isSafeImageType(mimeType)) return "📷 Foto";
    if (fileName) return `📄 ${shortAttachmentName(fileName)}`;
    return "📄 File";
  }
  if (isSafeImageType(mimeType)) {
    return `📷 ${count} foto`;
  }
  return `📎 ${count} file`;
}

function roomPreviewContent(room: ChatRoom) {
  return (
    room.lastMessage?.content ||
    attachmentPreviewTextForRoom(
      room.lastMessage?.attachmentCount ?? 0,
      room.lastMessage?.attachmentMimeType,
      room.lastMessage?.attachmentFileName,
    )
  );
}

function roomPreviewReadPrefix(room: ChatRoom) {
  if (!room.lastMessage?.messageId) return "";
  const memberCount = roomPreviewReadCount(room);
  return memberCount > 0 ? "✓✓ " : "✓ ";
}

function roomPreviewReadCount(room: ChatRoom) {
  if (selectedRoom.value?.roomId !== room.roomId) return 0;
  if (!room.lastMessage?.messageId || !readSummary.value?.members?.length)
    return 0;
  const createdAt =
    parseBackendTimestamp(room.lastMessage.createdAt)?.getTime() ?? Number.NaN;
  return readSummary.value.members.filter((member) => {
    if (member.userId === currentUserId.value) return false;
    if (member.lastReadMessageId === room.lastMessage?.messageId) return true;
    if (!member.lastReadAt || Number.isNaN(createdAt)) return false;
    const lastReadAt =
      parseBackendTimestamp(member.lastReadAt)?.getTime() ?? Number.NaN;
    return !Number.isNaN(lastReadAt) && lastReadAt >= createdAt;
  }).length;
}

function shortAttachmentName(fileName?: string) {
  if (!fileName) return "File";
  return fileName.length > 18 ? `${fileName.slice(0, 15)}...` : fileName;
}

function isSafeImageType(mimeType?: string) {
  return ["image/png", "image/jpeg", "image/webp", "image/gif"].includes(
    (mimeType || "").toLowerCase(),
  );
}

function fileExtension(fileName?: string) {
  const value = fileName?.trim().toLowerCase() || "";
  const lastDot = value.lastIndexOf(".");
  if (lastDot < 0) return "";
  return value.slice(lastDot + 1);
}

function fileTypeLabel(mimeType?: string, fileName?: string) {
  const value = (mimeType || "").toLowerCase();
  const ext = fileExtension(fileName);
  if (isSafeImageType(value)) return "Gambar";
  if (value === "application/pdf" || ext === "pdf") return "PDF";
  if (
    value.includes("spreadsheet") ||
    value.includes("excel") ||
    ["xls", "xlsx", "csv"].includes(ext)
  ) {
    return "Spreadsheet";
  }
  if (
    value.includes("presentation") ||
    value.includes("powerpoint") ||
    ["ppt", "pptx"].includes(ext)
  ) {
    return "Presentasi";
  }
  if (
    value.includes("word") ||
    value.includes("document") ||
    ["doc", "docx"].includes(ext)
  ) {
    return "Dokumen";
  }
  if (value.startsWith("text/") || ["txt", "md", "rtf"].includes(ext))
    return "Teks";
  if (
    value.includes("zip") ||
    value.includes("compressed") ||
    ["zip", "rar", "7z"].includes(ext)
  ) {
    return "Arsip";
  }
  return value || "File";
}

function filePreviewIcon(mimeType?: string, fileName?: string) {
  const label = fileTypeLabel(mimeType, fileName);
  switch (label) {
    case "PDF":
      return PhFilePdf;
    case "Dokumen":
      return PhFileDoc;
    case "Spreadsheet":
      return PhFileXls;
    case "Presentasi":
      return PhFilePpt;
    case "Arsip":
      return PhFileArchive;
    case "Teks":
      return PhFileText;
    case "Gambar":
      return PhImage;
    default:
      return PhFile;
  }
}

function formatFileSize(size?: number) {
  if (!size || size < 0) return "";
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${Math.round(size / 1024)} KB`;
  return `${(size / (1024 * 1024)).toFixed(1)} MB`;
}

function openImageLightbox(attachment: ChatAttachment) {
  if (!attachment.url || !isSafeImageType(attachment.mimeType)) return;
  lightboxImage.value = {
    url: attachment.url,
    name: attachment.fileName || "Gambar",
  };
  isLightboxOpen.value = true;
}

function openAttachment(attachment: ChatAttachment) {
  if (isSafeImageType(attachment.mimeType)) {
    openImageLightbox(attachment);
    return;
  }
  if (!attachment.url) return;
  window.open(attachment.url, "_blank", "noopener,noreferrer");
}

function closeImageLightbox() {
  isLightboxOpen.value = false;
  lightboxImage.value = null;
}

function handleGlobalKeydown(event: KeyboardEvent) {
  if (event.key === "Escape" && isLightboxOpen.value) {
    closeImageLightbox();
  }
}

function revokeAttachmentObjectUrls(message: ChatMessage) {
  for (const attachment of message.attachments ?? []) {
    if (attachment.url?.startsWith("blob:")) {
      URL.revokeObjectURL(attachment.url);
    }
  }
}

function revokeMessageAttachmentUrls() {
  for (const message of messages.value) {
    revokeAttachmentObjectUrls(message);
  }
}

function revokePendingAttachmentUrls() {
  for (const attachment of selectedFiles.value) {
    if (attachment.previewUrl) {
      URL.revokeObjectURL(attachment.previewUrl);
    }
  }
}

function resolveChatError(error: unknown) {
  if (error instanceof Error && error.message) {
    return error.message;
  }
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
  const formatted = formatBackendTime(value);
  return formatted === "Waktu tidak tersedia" ? "" : formatted;
}

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
          class="grid h-[calc(100vh-1.5rem)] min-h-155 flex-1 overflow-hidden rounded-xl bg-white lg:grid-cols-[300px_minmax(0,1fr)]"
        >
          <aside
            ref="roomListEl"
            class="min-w-0 overflow-y-auto border-[#ebe7df] bg-[#fbfaf8] lg:border-r"
          >
            <div class="px-4 py-4 sm:px-5">
              <div class="flex items-center justify-between gap-3">
                <p class="text-sm font-semibold text-[#171322]">Percakapan</p>
                <button
                  type="button"
                  class="flex items-center gap-1.5 rounded-lg bg-[#4f46e5] px-3 py-1.5 text-xs font-semibold text-white transition hover:bg-[#4338ca]"
                  @click="openCreateConversation('dm')"
                >
                  <PhPlus :size="13" weight="bold" />
                  Buat
                </button>
              </div>
            </div>
            <div class="flex gap-2 p-4">
              <input
                v-model="roomSearch"
                type="search"
                class="min-w-0 flex-1 rounded-lg border border-transparent bg-[#f3f1ec] px-3 py-2 text-xs text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#c7d2fe] focus:bg-white focus:ring-2 focus:ring-[#4f46e5]/15"
                placeholder="Cari ruang..."
                @keydown.enter.prevent="searchRooms"
              />
              <button
                type="button"
                class="rounded-lg border border-[#d8d2c8] px-3 py-2 text-xs font-semibold text-[#4f46e5] transition hover:border-[#4f46e5]"
                @click="searchRooms"
              >
                Cari
              </button>
            </div>

            <div class="space-y-1 px-4 pt-2 pb-4">
              <button
                v-for="room in conversationList"
                :key="room.roomId"
                type="button"
                class="flex w-full min-w-0 items-center gap-3 rounded-lg border px-3 py-3 text-left transition hover:bg-white"
                :class="
                  selectedRoom?.roomId === room.roomId
                    ? 'border-[#d7d1ff] bg-white'
                    : room.unreadCount > 0
                      ? 'border-[#c7d2fe] bg-white'
                      : 'border-[#ebe7df] bg-[#fbfaf8]'
                "
                @click="
                  selectedRoom = room;
                  loadLatestMessages();
                "
              >
                <span
                  class="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg text-sm font-semibold text-white"
                  :class="
                    isDirectMessageRoom(room) ? 'bg-[#059669]' : 'bg-[#4f46e5]'
                  "
                >
                  {{
                    isDirectMessageRoom(room)
                      ? getInitials(room.dmTargetName || room.dmTargetEmail)
                      : getInitials(room.roomName)
                  }}
                </span>
                <span class="min-w-0 flex-1">
                  <span
                    class="block truncate text-sm text-[#171322]"
                    :class="
                      room.unreadCount > 0 ? 'font-bold' : 'font-semibold'
                    "
                  >
                    {{ roomDisplayName(room) }}
                  </span>
                  <span
                    class="mt-0.5 block truncate text-xs"
                    :class="
                      room.unreadCount > 0
                        ? 'font-semibold text-[#3f3a4a]'
                        : 'text-[#6b7280]'
                    "
                  >
                    {{ roomPreview(room) }}
                  </span>
                </span>
                <span class="flex shrink-0 flex-col items-end gap-1">
                  <span class="text-[11px] text-[#9ca3af]">{{
                    formatTime(room.lastMessageAt)
                  }}</span>
                  <span
                    v-if="room.unreadCount > 0"
                    class="rounded-full bg-[#4f46e5] px-2 py-0.5 text-[11px] font-semibold text-white"
                    :aria-label="`${room.unreadCount} pesan belum dibaca`"
                  >
                    {{ room.unreadCount }}
                  </span>
                </span>
              </button>

              <div
                v-if="conversationList.length === 0"
                class="rounded-lg bg-[#fbfaf8] px-4 py-8 text-center"
              >
                <PhChatCircleText
                  class="mx-auto h-7 w-7 text-[#9ca3af]"
                  weight="duotone"
                />
                <p class="mt-3 text-sm font-semibold text-[#171322]">
                  Belum ada percakapan
                </p>
              </div>
            </div>
          </aside>

          <section
            class="relative flex min-h-0 min-w-0 flex-col bg-[#fbfaf8]"
            @dragenter="handleDragEnter"
            @dragover="handleDragOver"
            @dragleave="handleDragLeave"
            @drop="handleDropFiles"
          >
            <div
              v-if="!selectedRoom"
              class="flex flex-1 flex-col items-center justify-center px-6 py-16 text-center"
            >
              <div
                class="flex h-12 w-12 items-center justify-center rounded-xl bg-[#eef2ff] text-[#4f46e5]"
              >
                <PhChatCircleText class="h-6 w-6" weight="duotone" />
              </div>
              <h2 class="mt-4 text-base font-semibold text-[#171322]">
                Belum ada percakapan
              </h2>
              <p class="mt-2 max-w-xs text-sm leading-6 text-[#6b7280]">
                Mulailah dengan membuat ruang chat atau mengirim pesan langsung.
              </p>
              <button
                type="button"
                class="mt-5 flex items-center gap-2 rounded-lg bg-[#4f46e5] px-4 py-2 text-sm font-semibold text-white transition hover:bg-[#4338ca]"
                @click="openCreateConversation('dm')"
              >
                <PhPlus :size="15" weight="bold" />
                Buat chat baru
              </button>
            </div>

            <template v-else>
              <div
                v-if="isDragActive"
                class="pointer-events-none absolute inset-4 z-20 flex items-center justify-center rounded-2xl border-2 border-dashed border-[#4f46e5] bg-[#eef2ff]/80"
              >
                <div
                  class="rounded-2xl bg-white px-5 py-4 text-center shadow-sm"
                >
                  <p class="text-sm font-semibold text-[#171322]">
                    Lepas file di sini
                  </p>
                  <p class="mt-1 text-xs text-[#6b7280]">
                    Maksimal {{ maxChatAttachments }} file, masing-masing hingga
                    {{ maxChatAttachmentSizeMb }}MB
                  </p>
                </div>
              </div>
              <div
                class="flex items-center gap-3 border-b border-[#ebe7df] bg-white px-4 py-3 sm:px-5"
              >
                <div
                  class="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg bg-[#4f46e5] text-sm font-semibold text-white"
                >
                  {{ roomInitial }}
                </div>
                <div class="min-w-0 flex-1">
                  <button
                    v-if="selectedRoomIsGroup"
                    type="button"
                    class="block max-w-full truncate text-left text-sm font-semibold text-[#171322] transition hover:text-[#4f46e5]"
                    @click="openGroupInfo"
                  >
                    {{ roomDisplayName(selectedRoom) }}
                  </button>
                  <h2
                    v-else
                    class="truncate text-sm font-semibold text-[#171322]"
                  >
                    {{ roomDisplayName(selectedRoom) }}
                  </h2>
                  <p class="truncate text-xs text-[#6b7280]">
                    {{ selectedSchoolName }} · {{ roomSubtitle(selectedRoom) }}
                  </p>
                </div>
              </div>

              <div
                ref="messagesEl"
                class="min-h-0 flex-1 overflow-y-auto px-4 py-4 sm:px-5"
                @scroll="handleMessageScroll"
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
                    <div
                      class="h-12 w-2/3 animate-pulse rounded-2xl bg-white"
                    />
                    <div
                      class="ml-auto h-12 w-1/2 animate-pulse rounded-2xl bg-[#dfe3ff]"
                    />
                    <div
                      class="h-16 w-3/4 animate-pulse rounded-2xl bg-white"
                    />
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
                    <div
                      class="mb-3 flex h-9 w-9 items-center justify-center rounded-lg bg-[#eef2ff] text-[#4f46e5]"
                    >
                      <PhChatCircleText class="h-5 w-5" weight="duotone" />
                    </div>
                    <h3 class="text-sm font-semibold text-[#171322]">
                      Belum ada pesan.
                    </h3>
                    <p class="mt-1 max-w-sm text-sm leading-6 text-[#6b7280]">
                      {{
                        selectedRoomIsDM
                          ? "Mulai percakapan pertama di pesan langsung ini."
                          : "Mulai percakapan pertama di ruang ini."
                      }}
                    </p>
                  </div>

                  <template v-else>
                    <template
                      v-for="(message, index) in messages"
                      :key="message.messageId"
                    >
                      <div
                        v-if="shouldShowDateDivider(message, index)"
                        class="flex items-center gap-3 py-2"
                      >
                        <div class="h-px flex-1 bg-[#e7e1d7]" />
                        <span
                          class="shrink-0 rounded-full bg-white px-3 py-1 text-[11px] font-medium text-[#8b8592]"
                        >
                          {{ formatDateDivider(message.createdAt) }}
                        </span>
                        <div class="h-px flex-1 bg-[#e7e1d7]" />
                      </div>

                      <article
                        class="flex gap-2"
                        :class="[
                          message.isMine ? 'justify-end' : 'justify-start',
                          isGroupedWithPrevious(message, index)
                            ? 'mt-1'
                            : 'mt-3',
                        ]"
                      >
                        <div
                          class="flex max-w-[88%] flex-col gap-1 sm:max-w-[72%]"
                          :class="message.isMine ? 'items-end' : 'items-start'"
                        >
                          <p
                            v-if="shouldShowSender(message, index)"
                            class="px-2 text-xs font-medium text-[#6b7280]"
                          >
                            {{ message.senderName }}
                          </p>
                          <div
                            class="rounded-2xl px-4 py-2 text-sm leading-relaxed"
                            :class="
                              message.isMine
                                ? [
                                    isGroupedWithPrevious(message, index)
                                      ? 'rounded-tr-lg'
                                      : 'rounded-tr-2xl',
                                    isGroupedWithNext(message, index)
                                      ? 'rounded-br-lg'
                                      : 'rounded-br-md',
                                    'bg-[#4f46e5] text-white',
                                  ]
                                : [
                                    isGroupedWithPrevious(message, index)
                                      ? 'rounded-tl-lg'
                                      : 'rounded-tl-2xl',
                                    isGroupedWithNext(message, index)
                                      ? 'rounded-bl-lg'
                                      : 'rounded-bl-md',
                                    'border border-[#ebe7df] bg-white text-[#171322]',
                                  ]
                            "
                          >
                            <p
                              v-if="message.content"
                              class="whitespace-pre-wrap wrap-break-word"
                            >
                              {{ message.content }}
                            </p>
                            <div
                              v-if="message.attachments?.length"
                              class="mt-2 grid gap-2"
                              :class="
                                message.attachments.length > 1
                                  ? 'sm:grid-cols-2'
                                  : ''
                              "
                            >
                              <button
                                v-for="attachment in message.attachments"
                                :key="
                                  attachment.attachmentId ||
                                  attachment.mediaId ||
                                  attachment.fileName
                                "
                                type="button"
                                class="group overflow-hidden rounded-xl text-left"
                                :class="
                                  message.isMine
                                    ? 'bg-white/10 text-white ring-1 ring-white/20'
                                    : 'border border-[#ebe7df] bg-[#fbfaf8] text-[#171322]'
                                "
                                @click="openAttachment(attachment)"
                                :aria-label="
                                  isSafeImageType(attachment.mimeType)
                                    ? `Buka pratinjau ${attachment.fileName || 'gambar'}`
                                    : `Buka ${attachment.fileName || 'file'}`
                                "
                              >
                                <img
                                  v-if="
                                    isSafeImageType(attachment.mimeType) &&
                                    attachment.url
                                  "
                                  :src="attachment.url"
                                  :alt="
                                    attachment.fileName || 'Lampiran gambar'
                                  "
                                  class="max-h-64 w-full object-cover transition duration-200 group-hover:scale-[1.01]"
                                />
                                <div
                                  class="flex min-w-0 items-center gap-3 p-3"
                                >
                                  <span
                                    class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg"
                                    :class="
                                      message.isMine
                                        ? 'bg-white/15'
                                        : 'bg-white text-[#4f46e5]'
                                    "
                                  >
                                    <component
                                      :is="
                                        filePreviewIcon(
                                          attachment.mimeType,
                                          attachment.fileName,
                                        )
                                      "
                                      class="h-5 w-5"
                                      weight="duotone"
                                    />
                                  </span>
                                  <span class="min-w-0 flex-1">
                                    <span
                                      class="block truncate text-xs font-semibold"
                                    >
                                      {{ attachment.fileName || "Lampiran" }}
                                    </span>
                                    <span
                                      class="mt-0.5 block truncate text-[11px]"
                                      :class="
                                        message.isMine
                                          ? 'text-white/70'
                                          : 'text-[#8b8592]'
                                      "
                                    >
                                      {{
                                        fileTypeLabel(
                                          attachment.mimeType,
                                          attachment.fileName,
                                        )
                                      }}
                                      <template
                                        v-if="
                                          formatFileSize(attachment.sizeBytes)
                                        "
                                      >
                                        ·
                                        {{
                                          formatFileSize(attachment.sizeBytes)
                                        }}
                                      </template>
                                    </span>
                                  </span>
                                  <span
                                    class="inline-flex items-center gap-1 text-[11px] font-semibold"
                                  >
                                    <PhDownloadSimple class="h-3.5 w-3.5" />
                                    {{
                                      isSafeImageType(attachment.mimeType)
                                        ? "Lihat"
                                        : "Buka"
                                    }}
                                  </span>
                                </div>
                              </button>
                            </div>
                          </div>
                          <p
                            class="flex items-center gap-2 px-2 text-[11px] text-[#9ca3af]"
                          >
                            <span>{{ formatDateTime(message.createdAt) }}</span>
                            <span
                              v-if="message.deliveryStatus === 'uploading'"
                              class="inline-flex items-center gap-1.5 font-medium text-[#9ca3af]"
                            >
                              <PhSpinnerGap class="h-3.5 w-3.5 animate-spin" />
                              Mengunggah...
                            </span>
                            <span
                              v-else-if="message.deliveryStatus === 'sending'"
                              class="inline-flex items-center gap-1.5 font-medium text-[#9ca3af]"
                            >
                              <PhCheck
                                :size="13"
                                weight="bold"
                                class="text-[#9ca3af]"
                              />
                            </span>
                            <button
                              v-else-if="message.deliveryStatus === 'failed'"
                              type="button"
                              class="inline-flex items-center gap-1 rounded-full bg-[#fef2f2] px-2 py-0.5 font-medium text-[#dc2626]"
                              @click="retryFailedMessage(message)"
                            >
                              Gagal
                              <span
                                class="underline decoration-dotted underline-offset-2"
                                >Coba lagi</span
                              >
                            </button>
                            <span
                              v-else-if="readIndicatorLabel(message)"
                              class="inline-flex items-center gap-1 text-[#6b7280]"
                              :title="readIndicatorLabel(message)"
                              :aria-label="readIndicatorLabel(message)"
                            >
                              <PhChecks
                                v-if="isReadByOthers(message)"
                                :size="13"
                                weight="bold"
                                class="text-[#4f46e5]"
                              />
                              <PhCheck
                                v-else
                                :size="13"
                                weight="bold"
                                class="text-[#9ca3af]"
                              />
                              <span
                                v-if="readIndicatorCount(message)"
                                class="text-[10px] font-medium"
                              >
                                {{ readIndicatorCount(message) }}
                              </span>
                            </span>
                          </p>
                        </div>
                      </article>
                    </template>
                  </template>
                </div>
              </div>

              <div
                v-if="showJumpToLatest"
                class="pointer-events-none absolute bottom-24 right-5 z-10 sm:right-6"
              >
                <button
                  type="button"
                  class="pointer-events-auto inline-flex items-center gap-2 rounded-full border border-[#d7d1ff] bg-white px-3 py-2 text-xs font-semibold text-[#4f46e5] shadow-sm transition hover:border-[#4f46e5] hover:bg-[#eef2ff]"
                  aria-label="Lompat ke pesan terbaru"
                  @click="scrollToBottom"
                >
                  <span class="text-sm leading-none">↓</span>
                  Pesan baru
                </button>
              </div>

              <form
                class="shrink-0 px-4 py-3 sm:px-5"
                @submit.prevent="submitMessage"
              >
                <p v-if="composerError" class="mb-2 text-sm text-red-600">
                  {{ composerError }}
                </p>
                <div
                  v-if="selectedFiles.length"
                  class="mb-3 grid gap-2 sm:grid-cols-2 xl:grid-cols-3"
                >
                  <div
                    v-for="(attachment, index) in selectedFiles"
                    :key="attachment.id"
                    class="min-w-0 overflow-hidden rounded-xl border border-[#ebe7df] bg-[#fbfaf8]"
                  >
                    <img
                      v-if="attachment.previewUrl"
                      :src="attachment.previewUrl"
                      :alt="attachment.file.name"
                      class="h-28 w-full object-cover"
                    />
                    <div class="flex min-w-0 items-center gap-3 p-3">
                      <span
                        class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-white text-[#4f46e5]"
                      >
                        <component
                          :is="
                            filePreviewIcon(
                              attachment.file.type,
                              attachment.file.name,
                            )
                          "
                          class="h-5 w-5"
                          weight="duotone"
                        />
                      </span>
                      <span class="min-w-0 flex-1">
                        <span
                          class="block truncate text-xs font-semibold text-[#171322]"
                        >
                          {{ attachment.file.name }}
                        </span>
                        <span
                          class="mt-0.5 block truncate text-[11px] text-[#8b8592]"
                        >
                          {{
                            fileTypeLabel(
                              attachment.file.type,
                              attachment.file.name,
                            )
                          }}
                          · {{ formatFileSize(attachment.file.size) }}
                        </span>
                      </span>
                      <button
                        type="button"
                        class="rounded-lg p-1.5 text-[#9ca3af] transition hover:bg-white hover:text-[#dc2626]"
                        title="Hapus lampiran"
                        aria-label="Hapus lampiran"
                        @click="removeSelectedFile(index)"
                      >
                        <PhX class="h-4 w-4" />
                      </button>
                    </div>
                  </div>
                </div>
                <div class="flex items-end gap-2">
                  <input
                    ref="fileInputEl"
                    type="file"
                    class="hidden"
                    multiple
                    @change="handleFileSelection"
                  />
                  <button
                    type="button"
                    class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl border border-[#ebe7df] text-[#6b7280] transition hover:border-[#c7d2fe] hover:text-[#4f46e5] focus:outline-none focus:ring-2 focus:ring-[#4f46e5]/15 disabled:cursor-not-allowed disabled:opacity-60"
                    :disabled="
                      !selectedRoom?.canSend ||
                      selectedFiles.length >= maxChatAttachments
                    "
                    title="Tambah lampiran"
                    aria-label="Tambah lampiran"
                    @click="openFilePicker"
                  >
                    <PhPaperclip class="h-5 w-5" />
                  </button>
                  <textarea
                    v-model="draft"
                    rows="1"
                    class="max-h-32 min-h-11 flex-1 resize-none rounded-xl border border-transparent bg-[#f3f1ec] px-4 py-3 text-sm text-[#171322] outline-none transition placeholder:text-[#aaa29a] focus:border-[#c7d2fe] focus:bg-white focus:ring-2 focus:ring-[#4f46e5]/15"
                    placeholder="Tulis pesan..."
                    :disabled="!selectedRoom?.canSend"
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
                <p
                  class="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-[#9ca3af]"
                >
                  <span>Enter untuk kirim, Shift+Enter untuk baris baru.</span>
                  <span
                    v-if="composerStatusLabel"
                    class="inline-flex items-center gap-1.5 font-medium text-[#6b7280]"
                  >
                    <PhSpinnerGap class="h-3.5 w-3.5 animate-spin" />
                    {{ composerStatusLabel }}
                  </span>
                </p>
              </form>
            </template>
          </section>
        </div>
      </div>
    </section>

    <!-- Unified Create Conversation Modal (DM + Grup tabs) -->
    <div
      v-if="isCreateConversationOpen"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/30 px-4 py-6"
      @click.self="isCreateConversationOpen = false"
    >
      <div
        class="max-h-[90vh] w-full max-w-xl overflow-hidden rounded-xl bg-white"
      >
        <!-- Header -->
        <div class="px-5 py-4">
          <div class="flex items-center justify-between">
            <h2 class="text-base font-semibold text-[#171322]">
              Buat Percakapan
            </h2>
            <button
              type="button"
              class="rounded-lg p-1.5 text-[#9ca3af] transition hover:bg-[#f3f1ec] hover:text-[#171322]"
              aria-label="Tutup"
              @click="isCreateConversationOpen = false"
            >
              <PhX class="h-5 w-5" />
            </button>
          </div>
          <!-- Tab bar -->
          <div class="mt-3 flex gap-1 rounded-lg bg-[#f3f1ec] p-1">
            <button
              type="button"
              class="flex-1 rounded-md py-1.5 text-sm font-medium transition"
              :class="
                activeCreateTab === 'dm'
                  ? 'bg-white text-[#171322] shadow-sm'
                  : 'text-[#6b7280] hover:text-[#171322]'
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
                  ? 'bg-white text-[#171322] shadow-sm'
                  : 'text-[#6b7280] hover:text-[#171322]'
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
                class="text-sm font-medium text-[#171322]"
                for="chat-dm-search"
              >
                Cari warga sekolah
              </label>
              <div class="mt-1 flex gap-2">
                <input
                  id="chat-dm-search"
                  v-model="dmSearch"
                  type="text"
                  class="min-w-0 flex-1 rounded-lg border border-[#d8d2c8] px-3 py-2 text-sm text-[#171322] outline-none transition focus:border-[#4f46e5] focus:ring-2 focus:ring-[#4f46e5]/15"
                  placeholder="Cari warga sekolah..."
                  @keydown.enter.prevent="loadDMTargets"
                />
                <button
                  type="button"
                  class="rounded-lg border border-[#d8d2c8] px-3 py-2 text-sm font-medium text-[#4f46e5] transition hover:border-[#4f46e5] disabled:opacity-60"
                  :disabled="isLoadingDMTargets"
                  @click="loadDMTargets"
                >
                  Cari
                </button>
              </div>
            </div>

            <p
              v-if="directMessageError"
              class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600"
            >
              {{ directMessageError }}
            </p>

            <div class="rounded-lg border border-[#ebe7df]">
              <div
                class="border-b border-[#ebe7df] bg-[#fbfaf8] px-3 py-2 text-xs font-semibold uppercase tracking-[0.06em] text-[#9ca3af]"
              >
                Warga sekolah
              </div>
              <div v-if="isLoadingDMTargets" class="space-y-2 p-3">
                <div class="h-10 animate-pulse rounded-lg bg-[#f3f1ec]" />
                <div class="h-10 animate-pulse rounded-lg bg-[#f3f1ec]" />
              </div>
              <div
                v-else-if="dmResults.length === 0"
                class="rounded-lg bg-[#fbfaf8] p-3 text-sm leading-6 text-[#6b7280]"
              >
                Tidak ada warga sekolah yang cocok.
              </div>
              <div v-else class="max-h-64 overflow-y-auto p-2">
                <label
                  v-for="member in dmResults"
                  :key="member.userId"
                  class="flex cursor-pointer items-center gap-3 rounded-lg px-2 py-2 hover:bg-[#fbfaf8]"
                >
                  <input
                    v-model="selectedDMTargetId"
                    type="radio"
                    name="dm-target"
                    class="h-4 w-4 border-[#d8d2c8] text-[#4f46e5]"
                    :value="member.userId"
                  />
                  <span
                    class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-[#059669] text-xs font-semibold text-white"
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

          <div class="flex flex-col gap-2 px-5 py-4 sm:flex-row sm:justify-end">
            <button
              type="button"
              class="rounded-lg border border-[#d8d2c8] px-4 py-2 text-sm font-medium text-[#6b7280] transition hover:bg-[#fbfaf8]"
              :disabled="isOpeningDM"
              @click="isCreateConversationOpen = false"
            >
              Batal
            </button>
            <button
              type="submit"
              class="rounded-lg bg-[#4f46e5] px-4 py-2 text-sm font-semibold text-white transition hover:bg-[#4338ca] disabled:cursor-not-allowed disabled:bg-[#c7c3d7]"
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
                class="text-sm font-medium text-[#171322]"
                for="chat-group-name"
              >
                Nama ruang
              </label>
              <input
                id="chat-group-name"
                v-model="groupRoomName"
                type="text"
                class="mt-1 w-full rounded-lg border border-[#d8d2c8] px-3 py-2 text-sm text-[#171322] outline-none transition focus:border-blue focus:ring-1 focus:ring-[#4f46e5]"
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
                <div class="h-10 animate-pulse rounded-lg bg-[#f3f1ec]" />
                <div class="h-10 animate-pulse rounded-lg bg-[#f3f1ec]" />
              </div>
              <div
                v-else-if="memberResults.length === 0"
                class="rounded-lg bg-[#fbfaf8] p-3 text-sm leading-6 text-[#6b7280]"
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
              @click="isCreateConversationOpen = false"
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

    <div
      v-if="isLightboxOpen && lightboxImage"
      class="fixed inset-0 z-70 flex items-center justify-center bg-black/85 px-4 py-6"
      @click="closeImageLightbox"
    >
      <div
        class="relative flex max-h-full max-w-6xl flex-col items-center gap-3"
        @click.stop
      >
        <button
          type="button"
          class="absolute right-0 top-0 inline-flex h-10 w-10 items-center justify-center rounded-full bg-black/45 text-white transition hover:bg-black/65"
          aria-label="Tutup pratinjau gambar"
          @click="closeImageLightbox"
        >
          <PhX class="h-5 w-5" />
        </button>
        <img
          :src="lightboxImage.url"
          :alt="lightboxImage.name"
          class="max-h-[78vh] max-w-full rounded-2xl object-contain shadow-2xl"
        />
        <p class="max-w-full truncate px-12 text-sm text-white/85">
          {{ lightboxImage.name }}
        </p>
      </div>
    </div>

    <div
      v-if="isGroupInfoOpen"
      class="fixed inset-0 z-50 flex justify-end bg-black/30"
    >
      <div
        class="flex h-full w-full max-w-lg flex-col overflow-hidden bg-white"
      >
        <div class="border-b border-[#ebe7df] px-5 py-4">
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <p
                class="text-xs font-semibold uppercase tracking-[0.06em] text-[#9ca3af]"
              >
                Info grup
              </p>
              <h2 class="mt-1 truncate text-lg font-semibold text-[#171322]">
                {{ groupInfo?.roomName || roomDisplayName(selectedRoom) }}
              </h2>
              <p class="mt-1 text-sm text-[#6b7280]">
                {{ groupInfo?.memberCount || 0 }} anggota ·
                {{ groupInfo?.schoolName || selectedSchoolName }}
              </p>
            </div>
            <button
              type="button"
              class="rounded-lg border border-[#d8d2c8] px-3 py-2 text-sm font-medium text-[#6b7280] transition hover:bg-[#fbfaf8]"
              @click="isGroupInfoOpen = false"
            >
              Tutup
            </button>
          </div>
        </div>

        <div class="min-h-0 flex-1 overflow-y-auto px-5 py-4">
          <div v-if="isLoadingGroupInfo" class="space-y-3">
            <div class="h-16 animate-pulse rounded-xl bg-[#f3f1ec]" />
            <div class="h-32 animate-pulse rounded-xl bg-[#f3f1ec]" />
            <div class="h-40 animate-pulse rounded-xl bg-[#f3f1ec]" />
          </div>

          <div
            v-else-if="groupInfoError"
            class="rounded-xl border border-red-100 bg-red-50 px-4 py-5 text-sm text-red-600"
          >
            <p>{{ groupInfoError }}</p>
            <button
              type="button"
              class="mt-3 rounded-lg bg-[#4f46e5] px-3 py-2 text-xs font-semibold text-white"
              @click="loadGroupInfo"
            >
              Coba lagi
            </button>
          </div>

          <template v-else>
            <p
              v-if="groupActionError"
              class="mb-4 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600"
            >
              {{ groupActionError }}
            </p>

            <section
              class="rounded-xl border border-[#ebe7df] bg-[#fbfaf8] p-4"
            >
              <p class="text-sm font-semibold text-[#171322]">Dibuat oleh</p>
              <div class="mt-3 flex items-center gap-3">
                <span
                  class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-[#4f46e5] text-xs font-semibold text-white"
                >
                  {{
                    getInitials(
                      groupInfo?.creator?.fullName || groupInfo?.creator?.email,
                    )
                  }}
                </span>
                <span class="min-w-0">
                  <span
                    class="block truncate text-sm font-medium text-[#171322]"
                  >
                    {{
                      groupInfo?.creator?.fullName ||
                      groupInfo?.creator?.email ||
                      "Tidak tersedia"
                    }}
                  </span>
                  <span class="block truncate text-xs text-[#6b7280]">
                    {{ groupInfo?.creator?.email }}
                  </span>
                </span>
              </div>
              <p class="mt-3 text-xs text-[#9ca3af]">
                Dibuat {{ formatDateTime(groupInfo?.createdAt) }}
              </p>
            </section>

            <section
              v-if="currentUserIsGroupAdmin"
              class="mt-4 rounded-xl border border-[#ebe7df] bg-white p-4"
            >
              <p class="text-sm font-semibold text-[#171322]">Ubah nama grup</p>
              <form class="mt-3 flex gap-2" @submit.prevent="submitRenameGroup">
                <input
                  v-model="renameRoomName"
                  type="text"
                  class="min-w-0 flex-1 rounded-lg border border-[#d8d2c8] px-3 py-2 text-sm text-[#171322] outline-none transition focus:border-[#4f46e5] focus:ring-2 focus:ring-[#4f46e5]/15"
                  placeholder="Nama ruang grup"
                />
                <button
                  type="submit"
                  class="rounded-lg bg-[#4f46e5] px-3 py-2 text-sm font-semibold text-white transition hover:bg-[#4338ca] disabled:cursor-not-allowed disabled:bg-[#c7c3d7]"
                  :disabled="isRenamingGroup"
                >
                  {{ isRenamingGroup ? "Menyimpan..." : "Simpan" }}
                </button>
              </form>
            </section>

            <section
              v-if="currentUserIsGroupAdmin"
              class="mt-4 rounded-xl border border-[#ebe7df] bg-white p-4"
            >
              <p class="text-sm font-semibold text-[#171322]">Tambah anggota</p>
              <p class="mt-1 text-xs text-[#6b7280]">
                Hanya warga aktif sekolah yang belum ada di grup ini.
              </p>
              <div class="mt-3 flex gap-2">
                <input
                  v-model="addMemberSearch"
                  type="search"
                  class="min-w-0 flex-1 rounded-lg border border-[#d8d2c8] px-3 py-2 text-sm text-[#171322] outline-none transition focus:border-[#4f46e5] focus:ring-2 focus:ring-[#4f46e5]/15"
                  placeholder="Cari nama atau email..."
                  @keydown.enter.prevent="loadEligibleMembers"
                />
                <button
                  type="button"
                  class="rounded-lg border border-[#d8d2c8] px-3 py-2 text-sm font-medium text-[#4f46e5] transition hover:border-[#4f46e5] disabled:opacity-60"
                  :disabled="isLoadingAddMembers"
                  @click="loadEligibleMembers"
                >
                  Cari
                </button>
              </div>

              <div class="mt-3 rounded-lg border border-[#ebe7df]">
                <div v-if="isLoadingAddMembers" class="space-y-2 p-3">
                  <div class="h-10 animate-pulse rounded-lg bg-[#f3f1ec]" />
                  <div class="h-10 animate-pulse rounded-lg bg-[#f3f1ec]" />
                </div>
                <div
                  v-else-if="addMemberResults.length === 0"
                  class="rounded-lg bg-[#fbfaf8] p-3 text-sm leading-6 text-[#6b7280]"
                >
                  Tidak ada warga yang bisa ditambahkan.
                </div>
                <div v-else class="max-h-52 overflow-y-auto p-2">
                  <label
                    v-for="member in addMemberResults"
                    :key="member.userId"
                    class="flex cursor-pointer items-center gap-3 rounded-lg px-2 py-2 hover:bg-[#fbfaf8]"
                  >
                    <input
                      type="checkbox"
                      class="h-4 w-4 rounded border-[#d8d2c8] text-[#4f46e5]"
                      :checked="selectedAddMemberIds.includes(member.userId)"
                      @change="toggleAddMember(member.userId)"
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

              <div class="mt-3 flex items-center justify-between gap-3">
                <p class="text-xs text-[#6b7280]">
                  {{ selectedAddMembers.length }} anggota dipilih
                </p>
                <button
                  type="button"
                  class="rounded-lg bg-[#4f46e5] px-3 py-2 text-sm font-semibold text-white transition hover:bg-[#4338ca] disabled:cursor-not-allowed disabled:bg-[#c7c3d7]"
                  :disabled="
                    isAddingMembers || selectedAddMemberIds.length === 0
                  "
                  @click="submitAddMembers"
                >
                  {{ isAddingMembers ? "Menambahkan..." : "Tambah anggota" }}
                </button>
              </div>
            </section>

            <section class="mt-4 rounded-xl border border-[#ebe7df] bg-white">
              <div class="border-b border-[#ebe7df] px-4 py-3">
                <p class="text-sm font-semibold text-[#171322]">Anggota</p>
              </div>
              <div class="divide-y divide-[#ebe7df]">
                <div
                  v-for="member in groupInfo?.members || []"
                  :key="member.userId"
                  class="flex items-center gap-3 px-4 py-3"
                >
                  <span
                    class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-[#4f46e5] text-xs font-semibold text-white"
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
                  <span
                    class="rounded-full px-2 py-1 text-[11px] font-semibold"
                    :class="
                      member.role === 'admin'
                        ? 'bg-[#eef2ff] text-[#4f46e5]'
                        : 'bg-[#f3f1ec] text-[#6b7280]'
                    "
                  >
                    {{ member.role === "admin" ? "Admin" : "Anggota" }}
                  </span>
                  <button
                    v-if="
                      currentUserIsGroupAdmin && member.userId !== currentUserId
                    "
                    type="button"
                    class="rounded-lg border border-red-100 px-3 py-1.5 text-xs font-semibold text-red-600 transition hover:bg-red-50 disabled:opacity-60"
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

        <div class="border-t border-[#ebe7df] px-5 py-4">
          <button
            type="button"
            class="w-full rounded-lg border border-red-100 px-4 py-2 text-sm font-semibold text-red-600 transition hover:bg-red-50 disabled:opacity-60"
            :disabled="isLeavingGroup"
            @click="leaveSelectedGroup"
          >
            {{ isLeavingGroup ? "Keluar..." : "Keluar grup" }}
          </button>
        </div>
      </div>
    </div>
  </main>
</template>
