<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import {
  PhArrowRight,
  PhChatCircleText,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import { getChatRooms } from "../../services/chat";
import type { ChatRoom } from "../../types/chat";

const props = withDefaults(
  defineProps<{
    to: string;
    limit?: number;
  }>(),
  {
    limit: 4,
  },
);

const rooms = ref<ChatRoom[]>([]);
const isLoading = ref(false);
const hasError = ref(false);
const authStore = useAuthStore();
const currentUserId = computed(() => authStore.user?.id || "");

const unreadRooms = computed(() =>
  [...rooms.value]
    .filter((room) => room.unreadCount > 0)
    .sort((left, right) => {
      const leftTime = new Date(left.lastMessageAt || 0).getTime();
      const rightTime = new Date(right.lastMessageAt || 0).getTime();
      return rightTime - leftTime;
    }),
);

const visibleRooms = computed(() => unreadRooms.value.slice(0, props.limit));

onMounted(loadLatestChats);

async function loadLatestChats() {
  isLoading.value = true;
  hasError.value = false;
  try {
    rooms.value = await getChatRooms();
  } catch {
    rooms.value = [];
    hasError.value = true;
  } finally {
    isLoading.value = false;
  }
}

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
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return new Intl.DateTimeFormat("id-ID", {
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}
</script>

<template>
  <article
    class="min-w-0 max-w-full overflow-hidden rounded-xl border border-[#ebe7df] bg-white p-4 sm:p-5"
  >
    <div class="mb-4 flex items-center justify-between gap-3">
      <div class="min-w-0">
        <p class="text-sm font-semibold text-[#171322]">Chat terbaru</p>
        <p class="mt-1 text-xs leading-5 text-[#8b8592]">
          Percakapan yang belum kamu baca.
        </p>
      </div>
      <RouterLink
        :to="to"
        class="inline-flex shrink-0 items-center gap-1 text-xs font-semibold text-[#4f46e5] transition hover:text-[#4338ca]"
      >
        Buka chat
        <PhArrowRight :size="14" />
      </RouterLink>
    </div>

    <div v-if="isLoading" class="space-y-2">
      <div
        v-for="item in 3"
        :key="item"
        class="h-14 animate-pulse rounded-lg bg-[#f3f4f6]"
      />
    </div>

    <div
      v-else-if="hasError"
      class="flex gap-3 rounded-lg border border-[#f1d6d3] bg-[#fffafa] p-3"
    >
      <PhWarningCircle :size="18" class="mt-0.5 shrink-0 text-[#dc2626]" />
      <p class="text-xs leading-5 text-[#7a7385]">
        Ringkasan chat belum bisa dimuat. Halaman lain tetap dapat digunakan.
      </p>
    </div>

    <div v-else-if="visibleRooms.length > 0" class="min-w-0 space-y-2">
      <RouterLink
        v-for="room in visibleRooms"
        :key="room.roomId"
        :to="to"
        class="flex min-w-0 max-w-full overflow-hidden items-center gap-3 rounded-lg border border-[#ebe7df] bg-[#fbfaf8] p-3 transition hover:border-[#c7d2fe] hover:bg-white"
      >
        <span
          class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-[#eef2ff] text-[#4f46e5]"
        >
          <PhChatCircleText :size="18" weight="duotone" />
        </span>
        <span class="min-w-0 flex-1 overflow-hidden">
          <span
            class="block max-w-full truncate text-sm text-[#171322]"
            :class="room.unreadCount > 0 ? 'font-bold' : 'font-semibold'"
          >
            {{ roomDisplayName(room) }}
          </span>
          <span
            class="mt-0.5 block max-w-full overflow-hidden text-xs break-all line-clamp-2"
            :class="
              room.unreadCount > 0
                ? 'font-semibold text-[#3f3a4a]'
                : 'text-[#7a7385]'
            "
          >
            {{ roomPreview(room) }}
          </span>
        </span>
        <span class="flex w-10 shrink-0 flex-col items-end gap-1">
          <span class="text-[11px] text-[#9ca3af]">
            {{ formatTime(room.lastMessageAt) }}
          </span>
          <span
            v-if="room.unreadCount > 0"
            class="rounded-full bg-[#4f46e5] px-2 py-0.5 text-[10px] font-semibold text-white"
          >
            {{ room.unreadCount }}
          </span>
        </span>
      </RouterLink>
    </div>

    <p
      v-else
      class="rounded-lg border border-[#ebe7df] bg-[#fbfaf8] p-4 text-sm leading-6 text-[#7a7385]"
    >
      Semua percakapan telah dibaca.
    </p>
  </article>
</template>
