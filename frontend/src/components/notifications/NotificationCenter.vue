<script setup lang="ts">
import {
  PhArrowClockwise,
  PhArrowRight,
  PhBell,
  PhCheckCircle,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { useNotificationUnreadCount } from "../../composables/useNotificationUnreadCount";
import {
  getNotifications,
  markAllNotificationsAsRead,
  markNotificationAsRead,
} from "../../services/notifications";
import { useToastStore } from "../../stores/toast";
import type { NotificationItem } from "../../types/dashboard";
import { formatDateTime } from "../../utils/date";
import { getApiError } from "../../utils/error";

type NotificationFilter = "all" | "unread";

const PAGE_LIMIT = 12;

const router = useRouter();
const toast = useToastStore();
const notificationUnread = useNotificationUnreadCount();

const notifications = ref<NotificationItem[]>([]);
const activeFilter = ref<NotificationFilter>("all");
const loading = ref(false);
const loadingMore = ref(false);
const error = ref("");
const page = ref(1);
const totalPages = ref(1);
const totalItems = ref(0);
const markingAll = ref(false);
const markingNotificationIds = ref(new Set<string>());
let requestId = 0;

const unreadOnly = computed(() => activeFilter.value === "unread");
const hasMore = computed(() => page.value < totalPages.value);
const currentRole = computed(() =>
  router.currentRoute.value.path.startsWith("/teacher") ? "teacher" : "student",
);

const subtitle = computed(() =>
  currentRole.value === "teacher"
    ? "Pantau pengumpulan, komentar, dan aktivitas kelas yang perlu ditindaklanjuti."
    : "Pantau tugas, materi, nilai, dan pengumuman kelas terbaru.",
);

const emptyTitle = computed(() =>
  unreadOnly.value ? "Semua notifikasi sudah dibaca" : "Belum ada notifikasi",
);

const emptyDescription = computed(() =>
  unreadOnly.value
    ? "Notifikasi baru akan muncul di sini saat ada aktivitas akademik yang belum dibaca."
    : "Saat ada aktivitas akademik baru, notifikasi akan tampil di halaman ini.",
);

function normalizeTotalPages(value: number) {
  return Math.max(1, Number.isFinite(value) ? value : 1);
}


function notificationTitle(item: NotificationItem) {
  if (item.type === "assignment_created") return "Tugas baru";
  if (item.type === "feed_posted") return "Pengumuman kelas baru";
  if (item.type === "comment_added") return "Komentar baru";
  if (item.type === "assignment_graded") return "Tugas sudah dinilai";
  if (item.type === "material_added") return "Materi baru";
  return item.title || "Notifikasi";
}

function notificationMessage(item: NotificationItem) {
  return item.message || "Buka notifikasi untuk melihat informasi terbaru.";
}

function notificationTypeLabel(item: NotificationItem) {
  if (item.type === "assignment_created") return "Tugas";
  if (item.type === "feed_posted") return "Feed";
  if (item.type === "comment_added") return "Komentar";
  if (item.type === "assignment_graded") return "Nilai";
  if (item.type === "material_added") return "Materi";
  return "Info";
}

function isInternalNotificationLink(link?: string) {
  return Boolean(link && link.startsWith("/") && !link.startsWith("//"));
}

function dedupeNotifications(
  currentItems: NotificationItem[],
  nextItems: NotificationItem[],
) {
  const seen = new Set(currentItems.map((item) => item.notificationId));
  return [
    ...currentItems,
    ...nextItems.filter((item) => !seen.has(item.notificationId)),
  ];
}

async function loadNotifications(reset = true) {
  const currentRequestId = ++requestId;
  if (reset) {
    loading.value = true;
    page.value = 1;
  } else {
    loadingMore.value = true;
  }
  error.value = "";

  try {
    const nextPage = reset ? 1 : page.value + 1;
    const response = await getNotifications({
      page: nextPage,
      limit: PAGE_LIMIT,
      unreadOnly: unreadOnly.value,
    });

    if (currentRequestId !== requestId) return;

    const nextItems = response.data ?? [];
    notifications.value = reset
      ? nextItems
      : dedupeNotifications(notifications.value, nextItems);
    page.value = response.page ?? nextPage;
    totalPages.value = normalizeTotalPages(response.totalPages);
    totalItems.value = response.totalItems ?? notifications.value.length;
    notificationUnread.set(response.unreadCount ?? 0);
  } catch (loadError) {
    if (currentRequestId === requestId) {
      error.value = getApiError(loadError);
      if (reset) notifications.value = [];
    }
  } finally {
    if (currentRequestId === requestId) {
      loading.value = false;
      loadingMore.value = false;
    }
  }
}

function selectFilter(filter: NotificationFilter) {
  if (activeFilter.value === filter) return;
  activeFilter.value = filter;
  void loadNotifications(true);
}

function markNotificationRead(item: NotificationItem) {
  if (item.isRead || markingNotificationIds.value.has(item.notificationId)) {
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
    .catch((markError) => {
      if (previousNotification && !previousNotification.isRead) {
        notifications.value = notifications.value.map((notification) =>
          notification.notificationId === item.notificationId
            ? { ...notification, isRead: false }
            : notification,
        );
        notificationUnread.set(previousUnreadCount);
      }
      toast.error(getApiError(markError));
    })
    .finally(() => {
      const next = new Set(markingNotificationIds.value);
      next.delete(item.notificationId);
      markingNotificationIds.value = next;
    });
}

async function openNotification(item: NotificationItem) {
  markNotificationRead(item);

  if (isInternalNotificationLink(item.link)) {
    await router.push(item.link as string);
  }
}

function markAllRead() {
  if (markingAll.value || notificationUnread.unreadCount.value <= 0) return;

  const previousNotifications = notifications.value;
  const previousUnreadCount = notificationUnread.unreadCount.value;
  markingAll.value = true;
  notifications.value = notifications.value.map((notification) => ({
    ...notification,
    isRead: true,
  }));
  notificationUnread.clear();

  void markAllNotificationsAsRead()
    .then(() => {
      toast.success("Semua notifikasi ditandai sudah dibaca.");
    })
    .catch((markError) => {
      notifications.value = previousNotifications;
      notificationUnread.set(previousUnreadCount);
      toast.error(getApiError(markError));
    })
    .finally(() => {
      markingAll.value = false;
    });
}

onMounted(() => {
  void loadNotifications(true);
});
</script>

<template>
  <main class="min-h-full bg-background px-4 py-6 sm:px-6 lg:px-8">
    <section class="mx-auto flex w-full max-w-5xl flex-col gap-5">
      <header
        class="rounded-xl border border-border bg-surface px-5 py-5 sm:px-6"
      >
        <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div class="min-w-0">
            <p class="text-xs font-medium uppercase tracking-[0.16em] text-[#a09aa8]">
              Pusat notifikasi
            </p>
            <h1 class="mt-2 text-2xl font-semibold text-foreground">
              Notifikasi
            </h1>
            <p class="mt-2 max-w-2xl text-sm leading-6 text-muted">
              {{ subtitle }}
            </p>
          </div>

          <button
            type="button"
            class="inline-flex shrink-0 items-center justify-center gap-2 rounded-lg border border-border bg-surface-subtle px-3 py-2 text-sm font-medium text-brand transition hover:border-brand hover:bg-brand-soft disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="markingAll || notificationUnread.unreadCount.value <= 0"
            @click="markAllRead"
          >
            <PhCheckCircle :size="17" weight="bold" />
            {{ markingAll ? "Menyimpan..." : "Tandai semua dibaca" }}
          </button>
        </div>
      </header>

      <section class="rounded-xl border border-border bg-surface">
        <div
          class="flex flex-col gap-3 border-b border-border px-4 py-4 sm:flex-row sm:items-center sm:justify-between sm:px-5"
        >
          <div
            class="inline-flex w-full rounded-lg bg-background p-1 sm:w-auto"
            role="tablist"
            aria-label="Filter notifikasi"
          >
            <button
              type="button"
              class="flex-1 rounded-md px-3 py-1.5 text-sm font-medium transition sm:flex-none"
              :class="
                activeFilter === 'all'
                  ? 'bg-surface text-foreground shadow-sm'
                  : 'text-muted hover:text-foreground'
              "
              :aria-pressed="activeFilter === 'all'"
              @click="selectFilter('all')"
            >
              Semua
            </button>
            <button
              type="button"
              class="flex-1 rounded-md px-3 py-1.5 text-sm font-medium transition sm:flex-none"
              :class="
                activeFilter === 'unread'
                  ? 'bg-surface text-foreground shadow-sm'
                  : 'text-muted hover:text-foreground'
              "
              :aria-pressed="activeFilter === 'unread'"
              @click="selectFilter('unread')"
            >
              Belum dibaca
            </button>
          </div>

          <p class="text-xs text-[#a09aa8]">
            {{ notificationUnread.unreadCount.value }} belum dibaca
          </p>
        </div>

        <div v-if="loading" class="space-y-0 divide-y divide-border">
          <div v-for="item in 5" :key="item" class="flex gap-3 px-4 py-4 sm:px-5">
            <div class="h-10 w-10 shrink-0 animate-pulse rounded-full bg-surface-strong" />
            <div class="min-w-0 flex-1 space-y-2">
              <div class="h-3 w-28 animate-pulse rounded-full bg-surface-strong" />
              <div class="h-4 w-3/5 animate-pulse rounded-full bg-surface-strong" />
              <div class="h-3 w-4/5 animate-pulse rounded-full bg-surface-strong" />
            </div>
          </div>
        </div>

        <div
          v-else-if="error"
          class="flex flex-col items-center gap-3 px-5 py-14 text-center"
        >
          <div
            class="flex h-11 w-11 items-center justify-center rounded-full bg-warning-soft text-warning"
          >
            <PhWarningCircle :size="23" weight="duotone" />
          </div>
          <div>
            <h2 class="text-sm font-semibold text-foreground">
              Notifikasi belum bisa dimuat
            </h2>
            <p class="mt-1 text-sm leading-6 text-muted">
              {{ error }}
            </p>
          </div>
          <button
            type="button"
            class="inline-flex items-center gap-2 rounded-lg border border-border bg-surface px-3 py-2 text-sm font-medium text-brand transition hover:border-brand hover:bg-brand-soft"
            @click="loadNotifications(true)"
          >
            <PhArrowClockwise :size="16" />
            Coba lagi
          </button>
        </div>

        <div
          v-else-if="notifications.length === 0"
          class="flex flex-col items-center px-5 py-14 text-center"
        >
          <div
            class="mb-3 flex h-9 w-9 items-center justify-center rounded-lg bg-brand-soft text-brand"
          >
            <PhBell class="h-5 w-5" weight="duotone" />
          </div>
          <div>
            <h2 class="text-sm font-semibold text-foreground">
              {{ emptyTitle }}
            </h2>
            <p class="mt-1 max-w-md text-sm leading-6 text-muted">
              {{ emptyDescription }}
            </p>
          </div>
        </div>

        <div v-else class="divide-y divide-border">
          <button
            v-for="item in notifications"
            :key="item.notificationId"
            type="button"
            class="flex w-full min-w-0 gap-3 px-4 py-4 text-left transition hover:bg-background focus:outline-none focus-visible:ring-2 focus-visible:ring-brand focus-visible:ring-offset-2 disabled:cursor-wait disabled:opacity-75 sm:px-5"
            :class="!item.isRead ? 'bg-[#f5f7ff]' : 'bg-surface'"
            :disabled="markingNotificationIds.has(item.notificationId)"
            :aria-label="`${item.isRead ? 'Buka' : 'Buka dan tandai dibaca'} ${notificationTitle(item)}`"
            @click="openNotification(item)"
          >
            <div
              class="mt-0.5 flex h-10 w-10 shrink-0 items-center justify-center rounded-full border border-border bg-surface-subtle text-brand"
              aria-hidden="true"
            >
              <PhBell :size="18" weight="duotone" />
            </div>

            <div class="min-w-0 flex-1">
              <div class="flex min-w-0 flex-wrap items-center gap-2">
                <span
                  class="rounded-full bg-background px-2 py-0.5 text-[11px] font-medium text-muted"
                >
                  {{ notificationTypeLabel(item) }}
                </span>
                <span
                  v-if="!item.isRead"
                  class="rounded-full bg-brand px-2 py-0.5 text-[11px] font-medium text-white"
                >
                  baru
                </span>
                <span class="text-xs text-[#a09aa8]">
                  {{ formatDateTime(item.createdAt) }}
                </span>
              </div>

              <div class="mt-1 flex min-w-0 items-start gap-3">
                <div class="min-w-0 flex-1">
                  <h2
                    class="line-clamp-1 text-sm font-semibold text-foreground"
                  >
                    {{ notificationTitle(item) }}
                  </h2>
                  <p class="mt-1 line-clamp-2 text-sm leading-6 text-muted">
                    {{ notificationMessage(item) }}
                  </p>
                </div>
                <PhArrowRight
                  v-if="isInternalNotificationLink(item.link)"
                  :size="16"
                  class="mt-1 shrink-0 text-[#a09aa8]"
                  aria-hidden="true"
                />
              </div>
            </div>
          </button>
        </div>

        <div
          v-if="!loading && !error && notifications.length > 0"
          class="flex flex-col gap-2 border-t border-border px-4 py-4 sm:flex-row sm:items-center sm:justify-between sm:px-5"
        >
          <p class="text-xs text-[#a09aa8]">
            Menampilkan {{ notifications.length }} dari {{ totalItems }}
            notifikasi.
          </p>
          <button
            v-if="hasMore"
            type="button"
            class="inline-flex items-center justify-center gap-2 rounded-lg border border-border bg-surface px-3 py-2 text-sm font-medium text-brand transition hover:border-brand hover:bg-brand-soft disabled:cursor-wait disabled:opacity-60"
            :disabled="loadingMore"
            @click="loadNotifications(false)"
          >
            <PhArrowClockwise :size="16" :class="loadingMore ? 'animate-spin' : ''" />
            {{ loadingMore ? "Memuat..." : "Muat lagi" }}
          </button>
        </div>
      </section>
    </section>
  </main>
</template>
