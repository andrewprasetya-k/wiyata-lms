<script setup lang="ts">
import { computed, watch } from "vue";
import { RouterLink } from "vue-router";
import { PhArrowRight, PhWarningCircle } from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import type { ChatRoom } from "../../types/chat";
import { useChatRoomSummary } from "../../composables/useChatRoomSummary";
import {
  formatTime as formatBackendTime,
  parseBackendTimestamp,
} from "../../utils/date";

const props = withDefaults(
  defineProps<{
    to: string;
    limit?: number;
    embedded?: boolean;
  }>(),
  {
    limit: 4,
    embedded: false,
  },
);

const emit = defineEmits<{
  unreadChange: [count: number];
}>();

const { rooms, totalUnreadCount, loading, error } = useChatRoomSummary();
const authStore = useAuthStore();
const currentUserId = computed(() => authStore.user?.id || "");
const isLoading = computed(() => loading.value && rooms.value.length === 0);
const hasError = computed(
  () => Boolean(error.value) && rooms.value.length === 0,
);

const unreadRooms = computed(() =>
  [...rooms.value]
    .filter((room) => room.unreadCount > 0)
    .sort((left, right) => {
      const leftTime =
        parseBackendTimestamp(left.lastMessageAt)?.getTime() ?? 0;
      const rightTime =
        parseBackendTimestamp(right.lastMessageAt)?.getTime() ?? 0;
      return rightTime - leftTime;
    }),
);

const visibleRooms = computed(() => unreadRooms.value.slice(0, props.limit));

watch(
  totalUnreadCount,
  (count) => {
    emit("unreadChange", count);
  },
  { immediate: true },
);

function roomDisplayName(room: ChatRoom) {
  if (room.roomRefType === "school") return "Ruang Sekolah";
  if (room.roomType === "dm") {
    return room.dmTargetName || room.dmTargetEmail || "Pesan Langsung";
  }
  return room.roomName || "Ruang Grup";
}

function roomPreview(room: ChatRoom) {
  if (!room.lastMessage) return "Belum ada pesan.";
  const content =
    room.lastMessage.content ||
    attachmentPreview(
      room.lastMessage.attachmentCount,
      room.lastMessage.attachmentMimeType,
      room.lastMessage.attachmentFileName,
    );
  if (room.roomType === "dm") return content;
  if (room.lastMessage.senderId === currentUserId.value) {
    return `✓ ${content}`;
  }
  const sender = room.lastMessage.senderName.split(" ")[0] || "Pengguna";
  return `${sender}: ${content}`;
}

function attachmentPreview(
  count?: number,
  mimeType?: string,
  fileName?: string,
) {
  if ((count || 0) <= 0) return "Belum ada pesan.";
  if (count === 1) {
    if (isImageMimeType(mimeType)) return "📷 Foto";
    if (fileName) return `📄 ${shortAttachmentName(fileName)}`;
    return "📄 File";
  }
  if (isImageMimeType(mimeType)) return `📷 ${count} foto`;
  return `📎 ${count} file`;
}

function isImageMimeType(mimeType?: string | null) {
  return ["image/png", "image/jpeg", "image/webp", "image/gif"].includes(
    (mimeType || "").toLowerCase(),
  );
}

function shortAttachmentName(fileName?: string | null) {
  if (!fileName) return "File";
  return fileName.length > 18 ? `${fileName.slice(0, 15)}...` : fileName;
}

function formatTime(value?: string | null) {
  if (!value) return "";
  const formatted = formatBackendTime(value);
  return formatted === "Waktu tidak tersedia" ? "" : formatted;
}
</script>

<template>
  <article
    class="min-w-0 max-w-full overflow-hidden rounded-xl"
    :class="
      embedded
        ? 'bg-transparent'
        : 'border border-border bg-white p-4 sm:p-5'
    "
  >
    <div class="mb-4 flex items-center justify-between gap-3">
      <RouterLink
        :to="to"
        class="inline-flex shrink-0 items-center gap-1 text-xs font-semibold text-brand transition hover:text-brand-hover pt-1"
      >
        Buka chat
        <PhArrowRight :size="14" />
      </RouterLink>
    </div>

    <div v-if="isLoading" class="space-y-2">
      <div
        v-for="item in 3"
        :key="item"
        class="h-14 animate-pulse bg-[#f3f1ec]"
      />
    </div>

    <div
      v-else-if="hasError"
      class="flex gap-3 rounded-lg border border-danger-line bg-[#fffafa] p-3"
    >
      <PhWarningCircle :size="18" class="mt-0.5 shrink-0 text-danger" />
      <p class="text-xs leading-5 text-[#7a7385]">
        Ringkasan chat belum bisa dimuat. Halaman lain tetap dapat digunakan.
      </p>
    </div>

    <div v-else-if="visibleRooms.length > 0" class="min-w-0 space-y-2">
      <RouterLink
        v-for="room in visibleRooms"
        :key="room.roomId"
        :to="to"
        class="flex h-16 min-w-0 max-w-full overflow-hidden items-center gap-1 border-b border-border bg-[#fbfaf8] transition hover:border-[#c7d2fe] hover:bg-white focus:outline-none focus-visible:ring-2 focus-visible:ring-brand focus-visible:ring-offset-2"
      >
        <span class="min-w-0 flex-1 overflow-hidden">
          <span
            class="block w-full overflow-hidden text-ellipsis whitespace-nowrap text-sm text-foreground"
            :class="room.unreadCount > 0 ? 'font-bold' : 'font-semibold'"
          >
            {{ roomDisplayName(room) }}
          </span>
          <span
            class="mt-0.5 block w-full overflow-hidden text-ellipsis whitespace-nowrap text-xs"
            :class="
              room.unreadCount > 0
                ? 'font-semibold text-[#3f3a4a]'
                : 'text-[#7a7385]'
            "
          >
            {{ roomPreview(room) }}
          </span>
        </span>
        <span
          class="flex w-12 shrink-0 flex-col items-end gap-1 overflow-hidden"
        >
          <span class="w-full truncate text-right text-[11px] text-[#9ca3af]">
            {{ formatTime(room.lastMessageAt) }}
          </span>
          <span
            v-if="room.unreadCount > 0"
            class="max-w-full rounded-full bg-brand px-2 py-0.5 text-[10px] font-semibold text-white"
          >
            {{ room.unreadCount }}
          </span>
        </span>
      </RouterLink>
    </div>

    <div v-else class="rounded-lg p-4">
      <p class="text-sm font-semibold text-foreground">
        Tidak ada percakapan baru
      </p>
      <p class="mt-1 text-sm leading-6 text-muted">
        Percakapan yang belum dibaca akan tampil di sini.
      </p>
    </div>
  </article>
</template>
