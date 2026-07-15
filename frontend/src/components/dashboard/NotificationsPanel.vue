<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterLink, useRouter } from "vue-router";
import { PhArrowClockwise } from "@phosphor-icons/vue";
import {
  getNotifications,
  markAllNotificationsAsRead,
  markNotificationAsRead,
} from "../../services/notifications";
import { useNotificationUnreadCount } from "../../composables/useNotificationUnreadCount";
import type { NotificationItem } from "../../types/dashboard";
import { formatDateTime } from "../../utils/date";
import { getSubjectColor } from "../../utils/color";
import { useToastStore } from "../../stores/toast";
import { getApiError } from "../../utils/error";
import {
  isInternalNotificationLink,
  notificationAriaLabel,
  notificationBadge,
  notificationMessage,
  notificationTitle,
} from "../../utils/studentNotificationPreview";

withDefaults(
  defineProps<{
    to?: string;
  }>(),
  {
    to: "/notifications",
  },
);

const router = useRouter();
const toast = useToastStore();
const notificationUnread = useNotificationUnreadCount();

const notifications = ref<NotificationItem[]>([]);
const notificationsLoading = ref(false);
const notificationsError = ref("");
const markingNotificationIds = ref<Set<string>>(new Set());
const markingAllNotifications = ref(false);

async function loadNotifications() {
  notificationsLoading.value = true;
  notificationsError.value = "";

  try {
    const notificationData = await getNotifications({ page: 1, limit: 5 });
    notifications.value = notificationData.data ?? [];
    notificationUnread.set(notificationData.unreadCount ?? 0);
  } catch {
    notifications.value = [];
    notificationsError.value = "Notifikasi belum bisa dimuat.";
  } finally {
    notificationsLoading.value = false;
  }
}

function markNotificationRead(item: NotificationItem) {
  if (markingNotificationIds.value.has(item.notificationId)) {
    return;
  }

  if (item.isRead) {
    return;
  }

  const previousUnreadCount = notificationUnread.unreadCount.value;
  const previousNotification = notifications.value.find(
    (notification) => notification.notificationId === item.notificationId,
  );

  notifications.value = notifications.value.map((notification) =>
    notification.notificationId === item.notificationId
      ? { ...notification, isRead: true }
      : notification,
  );
  notificationUnread.decrement();
  markingNotificationIds.value = new Set([
    ...markingNotificationIds.value,
    item.notificationId,
  ]);

  void markNotificationAsRead(item.notificationId)
    .catch((error) => {
      if (previousNotification && !previousNotification.isRead) {
        notifications.value = notifications.value.map((notification) =>
          notification.notificationId === item.notificationId
            ? { ...notification, isRead: false }
            : notification,
        );
        notificationUnread.set(previousUnreadCount);
      }
      toast.error(getApiError(error));
    })
    .finally(() => {
      const next = new Set(markingNotificationIds.value);
      next.delete(item.notificationId);
      markingNotificationIds.value = next;
    });
}

async function handleNotificationClick(item: NotificationItem) {
  markNotificationRead(item);
  if (isInternalNotificationLink(item.link)) {
    await router.push(item.link as string);
  }
}

function markAllNotificationsRead() {
  if (
    notificationUnread.unreadCount.value <= 0 ||
    markingAllNotifications.value
  ) {
    return;
  }

  const previousNotifications = notifications.value;
  const previousUnreadCount = notificationUnread.unreadCount.value;
  markingAllNotifications.value = true;
  notifications.value = notifications.value.map((notification) => ({
    ...notification,
    isRead: true,
  }));
  notificationUnread.clear();

  void markAllNotificationsAsRead()
    .then(() => {
      toast.success("Semua notifikasi ditandai sudah dibaca.");
    })
    .catch((error) => {
      notifications.value = previousNotifications;
      notificationUnread.set(previousUnreadCount);
      toast.error(getApiError(error));
    })
    .finally(() => {
      markingAllNotifications.value = false;
    });
}

onMounted(loadNotifications);
</script>

<template>
  <section
    class="shrink-0 rounded-xl border border-border bg-surface shadow-sm p-4 sm:p-5"
  >
    <div class="mb-3 flex items-center justify-between gap-3">
      <div>
        <p class="text-sm font-medium text-foreground">Notifikasi</p>
        <p class="mt-0.5 text-xs text-muted">
          {{ notificationUnread.unreadCount.value }} belum dibaca
        </p>
      </div>
      <div class="flex items-center gap-2">
        <RouterLink
          class="rounded-lg border border-border bg-surface px-2 py-1 text-xs font-medium text-brand transition hover:border-brand hover:bg-brand-soft"
          :to="to"
        >
          Lihat semua
        </RouterLink>
        <button
          v-if="notificationUnread.unreadCount.value > 0"
          class="rounded-lg bg-brand-soft px-2 py-1 text-xs font-medium text-brand transition hover:bg-[#e0e7ff] disabled:cursor-not-allowed disabled:opacity-60"
          type="button"
          :disabled="markingAllNotifications"
          @click="markAllNotificationsRead"
        >
          {{ markingAllNotifications ? "Menyimpan..." : "Tandai semua dibaca" }}
        </button>
      </div>
    </div>

    <div v-if="notificationsLoading" class="space-y-2">
      <div
        v-for="item in 3"
        :key="item"
        class="h-16 animate-pulse rounded-lg bg-surface-strong"
      />
    </div>
    <div
      v-else-if="notificationsError"
      class="rounded-lg bg-surface-subtle p-4 text-sm leading-6 text-muted"
    >
      <p>{{ notificationsError }}</p>
      <button
        type="button"
        class="mt-3 inline-flex items-center gap-1.5 rounded-lg border border-border bg-surface px-3 py-1.5 text-xs font-medium text-brand transition hover:border-brand hover:bg-brand-soft disabled:cursor-not-allowed disabled:opacity-60"
        :disabled="notificationsLoading"
        @click="loadNotifications"
      >
        <PhArrowClockwise :size="14" />
        Coba lagi
      </button>
    </div>
    <div v-else-if="notifications.length > 0" class="space-y-1">
      <button
        v-for="item in notifications"
        :key="item.notificationId"
        class="flex min-w-0 w-full gap-3 p-3 text-left transition hover:bg-background disabled:cursor-wait disabled:opacity-75 border-b border-border"
        :class="!item.isRead ? 'bg-[#f5f7ff]' : ''"
        type="button"
        :disabled="markingNotificationIds.has(item.notificationId)"
        :aria-label="notificationAriaLabel(item)"
        @click="handleNotificationClick(item)"
      >
        <div
          class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full text-[11px] font-medium text-white"
          :style="{
            backgroundColor: getSubjectColor(item.type || item.notificationId),
          }"
        >
          {{ notificationBadge(item) }}
        </div>
        <div class="min-w-0 flex-1">
          <div class="flex items-baseline justify-between gap-2">
            <p class="line-clamp-1 text-sm font-medium text-foreground">
              {{ notificationTitle(item) }}
            </p>
            <span class="shrink-0 text-[10px] text-muted">{{
              formatDateTime(item.createdAt)
            }}</span>
          </div>
          <p class="line-clamp-2 text-xs leading-5 text-muted">
            {{ notificationMessage(item) }}
          </p>
          <span
            v-if="!item.isRead"
            class="mt-1 inline-flex rounded-full bg-brand px-2 py-0.5 text-[10px] font-medium text-white"
          >
            baru
          </span>
        </div>
      </button>
    </div>
    <div v-else class="rounded-lg bg-surface-subtle p-4">
      <p class="text-sm font-semibold text-foreground">
        Belum ada notifikasi terbaru
      </p>
      <p class="mt-1 text-sm leading-6 text-muted">
        Notifikasi baru akan tampil di sini.
      </p>
    </div>
  </section>
</template>
