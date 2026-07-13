<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from "vue";
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
  getChatRooms,
  getMessages,
  getRoomReadSummary,
  markRoomRead,
  sendMessage,
} from "../../services/chat";
import { connectChatSocket } from "../../services/chatSocket";
import { deleteMedia, uploadMediaFile } from "../../services/media";
import type { ChatSocketStatus } from "../../services/chatSocket";
import { useAuthStore } from "../../stores/auth";
import { useToastStore } from "../../stores/toast";
import type {
  ChatAttachment,
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
import {
  getInitials,
  isCustomGroupRoom,
  isDirectMessageRoom,
  resolveChatError,
  roomDisplayName,
} from "../../utils/chatDisplay";
import ChatCreateConversationModal from "./ChatCreateConversationModal.vue";
import ChatGroupInfoModal from "./ChatGroupInfoModal.vue";

defineProps<{
  audience: "student" | "teacher" | "admin";
}>();

const toast = useToastStore();

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
const createConversationInitialTab = ref<"dm" | "group">("dm");
const roomSearch = ref("");
const isGroupInfoOpen = ref(false);
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

function openCreateConversation(tab: "dm" | "group" = "dm") {
  createConversationInitialTab.value = tab;
  isCreateConversationOpen.value = true;
}

async function handleDmOpened(room: ChatRoom) {
  await refreshRooms();
  selectedRoom.value =
    rooms.value.find((item) => item.roomId === room.roomId) ?? room;
  isCreateConversationOpen.value = false;
  await loadLatestMessages();
}

async function handleGroupCreated(room: ChatRoom) {
  await refreshRooms();
  selectedRoom.value = room;
  isCreateConversationOpen.value = false;
  await loadLatestMessages();
  toast.success("Ruang chat berhasil dibuat.");
}

function openGroupInfo() {
  if (!selectedRoom.value || !isCustomGroupRoom(selectedRoom.value)) return;
  isGroupInfoOpen.value = true;
}

async function handleGroupRenamed(room: ChatRoom) {
  selectedRoom.value = room;
  await refreshRooms();
}

async function handleGroupMembersChanged() {
  await refreshRooms();
}

async function handleGroupLeft() {
  isGroupInfoOpen.value = false;
  await refreshRooms();
  selectedRoom.value = rooms.value[0] ?? null;
  messages.value = [];
  if (selectedRoom.value) {
    await loadLatestMessages();
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

// function isSchoolRoom(room: ChatRoom) {
//   return room.roomRefType === "school";
// }

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
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-background">
    <section class="px-0 py-0 sm:px-0 lg:px-0">
      <div
        class="mx-auto flex max-h-screen min-h-screen max-w-screen flex-col gap-5"
      >
        <div
          v-if="isBooting"
          class="grid min-h-screen gap-4 overflow-hidden rounded-xl bg-surface p-4 lg:grid-cols-[300px_minmax(0,1fr)]"
        >
          <div class="space-y-3 border-border lg:border-r lg:pr-4">
            <div class="h-16 animate-pulse rounded-xl bg-[#f1eee8]" />
            <div class="h-20 animate-pulse rounded-xl bg-background" />
            <div class="h-20 animate-pulse rounded-xl bg-background" />
          </div>
          <div class="flex flex-col gap-3">
            <div class="h-16 animate-pulse rounded-xl bg-[#f1eee8]" />
            <div class="flex-1 animate-pulse rounded-xl bg-background" />
            <div class="h-16 animate-pulse rounded-xl bg-[#f1eee8]" />
          </div>
        </div>

        <div
          v-else-if="accessError"
          class="flex min-h-105 flex-col items-center justify-center rounded-xl bg-surface px-6 py-12 text-center"
        >
          <div
            class="flex h-12 w-12 items-center justify-center rounded-lg bg-danger-soft text-danger"
          >
            <PhWarningCircle :size="24" weight="duotone" />
          </div>
          <h2 class="mt-4 text-base font-semibold text-foreground">
            Chat belum bisa dibuka
          </h2>
          <p class="mt-2 max-w-md text-sm leading-6 text-muted">
            {{ accessError }}
          </p>
          <button
            type="button"
            class="mt-5 rounded-lg bg-foreground px-4 py-2 text-sm font-medium text-white transition hover:bg-foreground-secondary"
            @click="bootstrapChat"
          >
            Coba lagi
          </button>
        </div>

        <div
          v-else
          class="grid h-[calc(100vh-1.5rem)] min-h-155 flex-1 overflow-hidden rounded-xl bg-surface lg:grid-cols-[300px_minmax(0,1fr)]"
        >
          <aside
            ref="roomListEl"
            class="min-w-0 overflow-y-auto border-border bg-surface-subtle lg:border-r"
          >
            <div class="px-4 py-4 sm:px-5">
              <div class="flex items-center justify-between gap-3">
                <p class="text-sm font-semibold text-foreground">Percakapan</p>
                <button
                  type="button"
                  class="flex items-center gap-1.5 rounded-lg bg-brand px-3 py-1.5 text-xs font-semibold text-white transition hover:bg-brand-hover"
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
                class="min-w-0 flex-1 rounded-lg border border-transparent bg-surface-strong px-3 py-2 text-xs text-foreground outline-none transition placeholder:text-muted focus:border-brand-line focus:bg-surface focus:ring-2 focus:ring-brand/15"
                placeholder="Cari ruang..."
                @keydown.enter.prevent="searchRooms"
              />
              <button
                type="button"
                class="rounded-lg border border-border px-3 py-2 text-xs font-semibold text-brand transition hover:border-brand"
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
                class="flex w-full min-w-0 items-center gap-3 rounded-lg border px-3 py-3 text-left transition hover:bg-surface"
                :class="
                  selectedRoom?.roomId === room.roomId
                    ? 'border-[#d7d1ff] bg-surface'
                    : room.unreadCount > 0
                      ? 'border-brand-line bg-surface'
                      : 'border-border bg-surface-subtle'
                "
                @click="
                  selectedRoom = room;
                  loadLatestMessages();
                "
              >
                <span
                  class="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg text-sm font-semibold text-white"
                  :class="
                    isDirectMessageRoom(room) ? 'bg-[#059669]' : 'bg-brand'
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
                    class="block truncate text-sm text-foreground"
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
                        ? 'font-semibold text-foreground'
                        : 'text-muted'
                    "
                  >
                    {{ roomPreview(room) }}
                  </span>
                </span>
                <span class="flex shrink-0 flex-col items-end gap-1">
                  <span class="text-[11px] text-muted">{{
                    formatTime(room.lastMessageAt)
                  }}</span>
                  <span
                    v-if="room.unreadCount > 0"
                    class="rounded-full bg-brand px-2 py-0.5 text-[11px] font-semibold text-white"
                    :aria-label="`${room.unreadCount} pesan belum dibaca`"
                  >
                    {{ room.unreadCount }}
                  </span>
                </span>
              </button>

              <div
                v-if="conversationList.length === 0"
                class="rounded-lg bg-surface-subtle px-4 py-8 text-center"
              >
                <PhChatCircleText
                  class="mx-auto h-7 w-7 text-muted"
                  weight="duotone"
                />
                <p class="mt-3 text-sm font-semibold text-foreground">
                  Belum ada percakapan
                </p>
              </div>
            </div>
          </aside>

          <section
            class="relative flex min-h-0 min-w-0 flex-col bg-surface-subtle"
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
                class="flex h-12 w-12 items-center justify-center rounded-xl bg-brand-soft text-brand"
              >
                <PhChatCircleText class="h-6 w-6" weight="duotone" />
              </div>
              <h2 class="mt-4 text-base font-semibold text-foreground">
                Belum ada percakapan
              </h2>
              <p class="mt-2 max-w-xs text-sm leading-6 text-muted">
                Mulailah dengan membuat ruang chat atau mengirim pesan langsung.
              </p>
              <button
                type="button"
                class="mt-5 flex items-center gap-2 rounded-lg bg-brand px-4 py-2 text-sm font-semibold text-white transition hover:bg-brand-hover"
                @click="openCreateConversation('dm')"
              >
                <PhPlus :size="15" weight="bold" />
                Buat chat baru
              </button>
            </div>

            <template v-else>
              <div
                v-if="isDragActive"
                class="pointer-events-none absolute inset-4 z-20 flex items-center justify-center rounded-2xl border-2 border-dashed border-brand bg-brand-soft/80"
              >
                <div
                  class="rounded-2xl bg-surface px-5 py-4 text-center shadow-sm"
                >
                  <p class="text-sm font-semibold text-foreground">
                    Lepas file di sini
                  </p>
                  <p class="mt-1 text-xs text-muted">
                    Maksimal {{ maxChatAttachments }} file, masing-masing hingga
                    {{ maxChatAttachmentSizeMb }}MB
                  </p>
                </div>
              </div>
              <div
                class="flex items-center gap-3 border-b border-border bg-surface px-4 py-3 sm:px-5"
              >
                <div
                  class="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg bg-brand text-sm font-semibold text-white"
                >
                  {{ roomInitial }}
                </div>
                <div class="min-w-0 flex-1">
                  <button
                    v-if="selectedRoomIsGroup"
                    type="button"
                    class="block max-w-full truncate text-left text-sm font-semibold text-foreground transition hover:text-brand"
                    @click="openGroupInfo"
                  >
                    {{ roomDisplayName(selectedRoom) }}
                  </button>
                  <h2
                    v-else
                    class="truncate text-sm font-semibold text-foreground"
                  >
                    {{ roomDisplayName(selectedRoom) }}
                  </h2>
                  <p class="truncate text-xs text-muted">
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
                    class="mx-auto rounded-full border border-border bg-surface px-4 py-2 text-xs font-semibold text-brand transition hover:border-brand disabled:opacity-60"
                    :disabled="isLoadingOlder"
                    @click="loadOlderMessages"
                  >
                    {{ isLoadingOlder ? "Memuat..." : "Muat pesan sebelumnya" }}
                  </button>

                  <div v-if="isLoadingMessages" class="space-y-3">
                    <div
                      class="h-12 w-2/3 animate-pulse rounded-2xl bg-surface"
                    />
                    <div
                      class="ml-auto h-12 w-1/2 animate-pulse rounded-2xl bg-[#dfe3ff]"
                    />
                    <div
                      class="h-16 w-3/4 animate-pulse rounded-2xl bg-surface"
                    />
                  </div>

                  <div
                    v-else-if="threadError"
                    class="rounded-2xl border border-danger-line bg-surface px-4 py-6 text-center"
                  >
                    <p class="text-sm font-semibold text-danger">
                      {{ threadError }}
                    </p>
                    <button
                      type="button"
                      class="mt-3 rounded-xl bg-brand px-4 py-2 text-sm font-semibold text-white"
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
                      class="mb-3 flex h-9 w-9 items-center justify-center rounded-lg bg-brand-soft text-brand"
                    >
                      <PhChatCircleText class="h-5 w-5" weight="duotone" />
                    </div>
                    <h3 class="text-sm font-semibold text-foreground">
                      Belum ada pesan.
                    </h3>
                    <p class="mt-1 max-w-sm text-sm leading-6 text-muted">
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
                          class="shrink-0 rounded-full bg-surface px-3 py-1 text-[11px] font-medium text-muted"
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
                            class="px-2 text-xs font-medium text-muted"
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
                                    'bg-brand text-white',
                                  ]
                                : [
                                    isGroupedWithPrevious(message, index)
                                      ? 'rounded-tl-lg'
                                      : 'rounded-tl-2xl',
                                    isGroupedWithNext(message, index)
                                      ? 'rounded-bl-lg'
                                      : 'rounded-bl-md',
                                    'border border-border bg-surface text-foreground',
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
                                    ? 'bg-surface/10 text-white ring-1 ring-white/20'
                                    : 'border border-border bg-surface-subtle text-foreground'
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
                                        ? 'bg-surface/15'
                                        : 'bg-surface text-brand'
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
                                          : 'text-muted'
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
                            class="flex items-center gap-2 px-2 text-[11px] text-muted"
                          >
                            <span>{{ formatDateTime(message.createdAt) }}</span>
                            <span
                              v-if="message.deliveryStatus === 'uploading'"
                              class="inline-flex items-center gap-1.5 font-medium text-muted"
                            >
                              <PhSpinnerGap class="h-3.5 w-3.5 animate-spin" />
                              Mengunggah...
                            </span>
                            <span
                              v-else-if="message.deliveryStatus === 'sending'"
                              class="inline-flex items-center gap-1.5 font-medium text-muted"
                            >
                              <PhCheck
                                :size="13"
                                weight="bold"
                                class="text-muted"
                              />
                            </span>
                            <button
                              v-else-if="message.deliveryStatus === 'failed'"
                              type="button"
                              class="inline-flex items-center gap-1 rounded-full bg-danger-soft px-2 py-0.5 font-medium text-danger"
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
                              class="inline-flex items-center gap-1 text-muted"
                              :title="readIndicatorLabel(message)"
                              :aria-label="readIndicatorLabel(message)"
                            >
                              <PhChecks
                                v-if="isReadByOthers(message)"
                                :size="13"
                                weight="bold"
                                class="text-brand"
                              />
                              <PhCheck
                                v-else
                                :size="13"
                                weight="bold"
                                class="text-muted"
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
                  class="pointer-events-auto inline-flex items-center gap-2 rounded-full border border-[#d7d1ff] bg-surface px-3 py-2 text-xs font-semibold text-brand shadow-sm transition hover:border-brand hover:bg-brand-soft"
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
                <p v-if="composerError" class="mb-2 text-sm text-danger">
                  {{ composerError }}
                </p>
                <div
                  v-if="selectedFiles.length"
                  class="mb-3 grid gap-2 sm:grid-cols-2 xl:grid-cols-3"
                >
                  <div
                    v-for="(attachment, index) in selectedFiles"
                    :key="attachment.id"
                    class="min-w-0 overflow-hidden rounded-xl border border-border bg-surface-subtle"
                  >
                    <img
                      v-if="attachment.previewUrl"
                      :src="attachment.previewUrl"
                      :alt="attachment.file.name"
                      class="h-28 w-full object-cover"
                    />
                    <div class="flex min-w-0 items-center gap-3 p-3">
                      <span
                        class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-surface text-brand"
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
                          class="block truncate text-xs font-semibold text-foreground"
                        >
                          {{ attachment.file.name }}
                        </span>
                        <span
                          class="mt-0.5 block truncate text-[11px] text-muted"
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
                        class="rounded-lg p-1.5 text-muted transition hover:bg-surface hover:text-danger"
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
                    class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl border border-border text-muted transition hover:border-brand-line hover:text-brand focus:outline-none focus:ring-2 focus:ring-brand/15 disabled:cursor-not-allowed disabled:opacity-60"
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
                    class="max-h-32 min-h-11 flex-1 resize-none rounded-xl border border-transparent bg-surface-strong px-4 py-3 text-sm text-foreground outline-none transition placeholder:text-[#aaa29a] focus:border-brand-line focus:bg-surface focus:ring-2 focus:ring-brand/15"
                    placeholder="Tulis pesan..."
                    :disabled="!selectedRoom?.canSend"
                    @keydown="handleComposerKeydown"
                  />
                  <button
                    type="submit"
                    class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-brand text-white transition hover:bg-brand-hover focus:outline-none focus:ring-2 focus:ring-brand/30 disabled:cursor-not-allowed disabled:bg-[#c7c3d7]"
                    :disabled="!canSend"
                    aria-label="Kirim pesan"
                  >
                    <PhPaperPlaneTilt class="h-5 w-5" weight="fill" />
                  </button>
                </div>
                <p
                  class="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted"
                >
                  <span>Enter untuk kirim, Shift+Enter untuk baris baru.</span>
                  <span
                    v-if="composerStatusLabel"
                    class="inline-flex items-center gap-1.5 font-medium text-muted"
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

    <ChatCreateConversationModal
      :open="isCreateConversationOpen"
      :initial-tab="createConversationInitialTab"
      :current-user-id="currentUserId"
      @update:open="isCreateConversationOpen = $event"
      @dm-opened="handleDmOpened"
      @group-created="handleGroupCreated"
    />

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

    <ChatGroupInfoModal
      :open="isGroupInfoOpen"
      :room="selectedRoom"
      :current-user-id="currentUserId"
      @update:open="isGroupInfoOpen = $event"
      @renamed="handleGroupRenamed"
      @members-changed="handleGroupMembersChanged"
      @left="handleGroupLeft"
    />
  </main>
</template>
