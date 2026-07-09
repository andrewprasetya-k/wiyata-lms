<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from "vue";
import { RouterLink, useRoute } from "vue-router";
import {
  PhArrowLeft,
  PhChatCircleText,
  PhMegaphone,
  PhPaperclip,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import AttachmentPreviewList from "../../components/common/AttachmentPreviewList.vue";
import FeedComments from "../../components/feed/FeedComments.vue";
import { getClassFeed, markFeedNotificationsRead } from "../../services/feed";
import {
  clearFeedUnreadOptimistically,
  restoreFeedUnreadCount,
} from "../../composables/useFeedUnreadCount";
import { useActiveClassStore } from "../../stores/activeClass";
import { useAuthStore } from "../../stores/auth";
import type { FeedClassHeader, FeedPost } from "../../types/feed";
import { formatDateTime } from "../../utils/date";

const route = useRoute();
const auth = useAuthStore();
const activeClassStore = useActiveClassStore();
const classHeader = ref<FeedClassHeader | null>(null);
const posts = ref<FeedPost[]>([]);
const isLoading = ref(true);
const errorMessage = ref("");

const schoolUserId = computed(() => auth.activeSchoolUserId);
const activeClass = computed(() => activeClassStore.activeClass);
const classTitle = computed(
  () =>
    classHeader.value?.classTitle ||
    activeClass.value?.classTitle ||
    "Kelas aktif",
);
const classCode = computed(() => classHeader.value?.classCode || "");

async function loadContext() {
  if (!schoolUserId.value) {
    isLoading.value = false;
    errorMessage.value =
      "Konteks sekolah belum tersedia. Silakan login ulang atau pilih sekolah aktif terlebih dahulu.";
    return;
  }

  isLoading.value = true;
  errorMessage.value = "";

  try {
    await activeClassStore.loadClasses(schoolUserId.value);
    if (activeClassStore.errorMessage) {
      errorMessage.value = activeClassStore.errorMessage;
      posts.value = [];
      return;
    }

    if (!activeClassStore.activeClassId) {
      posts.value = [];
      return;
    }

    const feed = await getClassFeed(activeClassStore.activeClassId);
    classHeader.value = feed.class;
    posts.value = feed.data.data || [];
    void markCurrentFeedRead();
    void scrollToLinkedPost();
  } catch {
    errorMessage.value =
      "Feed kelas belum bisa dimuat. Periksa koneksi atau coba lagi nanti.";
  } finally {
    isLoading.value = false;
  }
}

onMounted(loadContext);

async function markCurrentFeedRead() {
  const previousUnreadCount = clearFeedUnreadOptimistically();

  try {
    await markFeedNotificationsRead();
  } catch {
    restoreFeedUnreadCount(previousUnreadCount);
    // Feed read marker should not block the feed page.
  }
}

async function scrollToLinkedPost() {
  const postId = route.query.post as string | undefined;
  if (!postId) return;
  await nextTick();
  const el = document.getElementById(`post-${postId}`);
  if (el) {
    el.scrollIntoView({ behavior: "smooth", block: "center" });
    el.classList.add("ring-2", "ring-[#4f46e5]", "ring-offset-2");
    setTimeout(() => el.classList.remove("ring-2", "ring-[#4f46e5]", "ring-offset-2"), 3000);
  }
}

function updatePostCommentCount(feedId: string, count: number) {
  posts.value = posts.value.map((post) =>
    post.feedId === feedId ? { ...post, commentCount: count } : post,
  );
}

function getInitials(name?: string) {
  const normalized = name?.trim();
  if (!normalized) return "EV";

  return normalized
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part.charAt(0))
    .join("")
    .toUpperCase();
}
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 bg-[#f8f7f4]">
    <header class="border-b border-[#ebe7df] bg-white">
      <div class="px-5 py-5 sm:px-6 lg:px-8">
        <div class="flex min-w-0 items-center gap-2 text-xs text-[#6b7280]">
          <RouterLink
            class="inline-flex shrink-0 items-center gap-1.5 transition hover:text-[#4f46e5]"
            to="/student/subjects"
          >
            <PhArrowLeft :size="15" />
            Mata pelajaran
          </RouterLink>
          <span class="text-[#d1d5db]">/</span>
          <span class="min-w-0 truncate font-medium text-[#171322]">
            Pengumuman kelas
          </span>
        </div>

        <div class="mt-4 flex min-w-0 items-start gap-3">
          <div
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-[#eef2ff] text-[#4f46e5]"
          >
            <PhMegaphone :size="21" weight="duotone" />
          </div>
          <div class="min-w-0">
            <h1 class="truncate text-2xl font-semibold text-[#171322] sm:text-3xl">
              Feed Kelas
            </h1>
            <p class="mt-2 max-w-3xl text-sm leading-6 text-[#6b7280]">
              Pengumuman dari teacher dan admin untuk kelas aktifmu.
            </p>
          </div>
        </div>
      </div>
    </header>

    <section
      v-if="isLoading || activeClassStore.isLoading"
      class="mx-auto grid max-w-6xl gap-4 px-5 py-5 sm:px-6 lg:grid-cols-[minmax(0,1fr)_280px] lg:px-8 lg:py-6"
    >
      <div class="space-y-3">
        <div
          v-for="item in 3"
          :key="item"
          class="h-48 animate-pulse rounded-xl border border-[#ebe7df] bg-white"
        />
      </div>
      <div
        class="h-52 animate-pulse rounded-xl border border-[#ebe7df] bg-white"
      />
    </section>

    <section
      v-else-if="errorMessage"
      class="flex min-h-[calc(100vh-109px)] items-center justify-center px-5 py-10"
    >
      <article
        class="w-full max-w-xl rounded-xl border border-[#fecaca] bg-[#fef2f2] p-6"
      >
        <div class="flex items-start gap-3">
          <div
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-[#fef2f2] text-[#dc2626]"
          >
            <PhWarningCircle :size="22" weight="duotone" />
          </div>
          <div>
            <h2 class="text-base font-medium text-[#171322]">
              Feed kelas tidak dapat dimuat
            </h2>
            <p class="mt-1 text-sm leading-6 text-[#7a7385]">
              {{ errorMessage }}
            </p>
            <button
              class="mt-4 rounded-lg bg-[#4f46e5] px-4 py-2 text-sm font-medium text-white transition hover:bg-[#4338ca]"
              type="button"
              @click="loadContext"
            >
              Coba lagi
            </button>
          </div>
        </div>
      </article>
    </section>

    <section
      v-else-if="activeClass"
      class="mx-auto grid max-w-screen min-w-0 gap-4 px-5 py-5 sm:px-6 lg:grid-cols-[minmax(0,1fr)_280px] lg:items-start lg:px-8 lg:py-6"
    >
      <div class="min-w-0">
        <div
          class="mb-4 flex min-w-0 items-center justify-between gap-3 rounded-xl border border-[#ebe7df] bg-white px-4 py-3"
        >
          <div class="flex min-w-0 items-center gap-3">
            <span
              class="h-2.5 w-2.5 shrink-0 rounded-full bg-[#4f8ef7]"
              aria-hidden="true"
            />
            <div class="min-w-0">
              <p class="truncate text-sm font-medium text-[#171322]">
                {{ classTitle }}
              </p>
              <p class="mt-0.5 text-[11px] text-[#9ca3af]">
                Pengumuman terbaru ditampilkan lebih dulu.
              </p>
            </div>
          </div>
          <span
            class="shrink-0 rounded-full bg-[#eef2ff] px-2.5 py-1 text-[10px] font-medium text-[#4f46e5]"
          >
            {{ posts.length }} post
          </span>
        </div>

        <article
          v-if="posts.length === 0"
          class="rounded-xl border border-[#ebe7df] bg-white p-8 text-center"
        >
          <div
            class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-[#eef2ff] text-[#4f46e5]"
          >
            <PhMegaphone class="h-6 w-6" weight="duotone" />
          </div>
          <h2 class="mt-3 text-base font-semibold text-[#171322]">
            Belum ada pengumuman
          </h2>
          <p class="mx-auto mt-2 max-w-md text-sm leading-6 text-[#6b7280]">
            Pengumuman akan tampil setelah teacher atau admin membagikan
            informasi untuk kelas ini.
          </p>
        </article>

        <div v-else class="space-y-3">
          <article
            v-for="post in posts"
            :key="post.feedId"
            :id="`post-${post.feedId}`"
            class="min-w-0 rounded-xl border border-[#ebe7df] bg-white p-4 sm:p-5 transition-shadow"
          >
            <div class="flex min-w-0 items-start gap-3">
              <div
                class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-[#059669] text-xs font-medium text-white"
                aria-hidden="true"
              >
                {{ getInitials(post.creatorName) }}
              </div>
              <div class="min-w-0 flex-1">
                <div
                  class="flex flex-col gap-1 sm:flex-row sm:items-start sm:justify-between"
                >
                  <div class="min-w-0">
                    <p class="truncate text-sm font-medium text-[#171322]">
                      {{ post.creatorName || "Pengirim tidak tersedia" }}
                    </p>
                    <p class="mt-0.5 text-[11px] text-[#9ca3af]">
                      {{ classTitle }}
                      <span v-if="classCode"> · {{ classCode }}</span>
                    </p>
                  </div>
                  <span class="shrink-0 text-[11px] text-[#9ca3af]">
                    {{ formatDateTime(post.createdAt) }}
                  </span>
                </div>

                <p
                  class="mt-3 whitespace-pre-line wrap-break-word text-sm leading-7 text-[#374151]"
                >
                  {{ post.content }}
                </p>

                <div
                  v-if="post.attachments?.length"
                  class="mt-4 rounded-xl border border-[#ebe7df] bg-[#fbfaf8] p-3"
                >
                  <div class="flex items-center justify-between gap-3">
                    <p
                      class="inline-flex items-center gap-1.5 text-xs font-medium text-[#374151]"
                    >
                      <PhPaperclip :size="15" class="text-[#4f46e5]" />
                      Lampiran
                    </p>
                    <span class="text-[11px] text-[#9ca3af]">
                      {{ post.attachments.length }} file
                    </span>
                  </div>
                  <AttachmentPreviewList
                    class="mt-3"
                    :attachments="post.attachments"
                  />
                </div>

                <FeedComments
                  :post="post"
                  @comment-count-change="updatePostCommentCount"
                />
              </div>
            </div>
          </article>
        </div>
      </div>

      <aside class="min-w-0 space-y-3 lg:sticky lg:top-6">
        <article class="rounded-xl border border-[#ebe7df] bg-white p-4">
          <div class="flex items-center gap-2">
            <PhMegaphone :size="17" class="text-[#4f46e5]" weight="duotone" />
            <h2 class="text-sm font-medium text-[#171322]">Info kelas</h2>
          </div>
          <dl class="mt-3 divide-y divide-[#f0ede8]">
            <div class="flex items-start justify-between gap-4 py-3">
              <dt class="text-xs text-[#7a7385]">Kelas aktif</dt>
              <dd
                class="max-w-[58%] text-right text-xs font-medium text-[#171322]"
              >
                {{ classTitle }}
              </dd>
            </div>
            <div class="flex items-start justify-between gap-4 py-3">
              <dt class="text-xs text-[#7a7385]">Pengumuman</dt>
              <dd class="text-right text-xs font-medium text-[#171322]">
                {{ posts.length }} post
              </dd>
            </div>
          </dl>
        </article>

        <article class="rounded-xl border border-[#dfe3ff] bg-[#eef2ff] p-4">
          <div class="flex items-start gap-3">
            <PhChatCircleText
              :size="19"
              class="mt-0.5 shrink-0 text-[#4f46e5]"
              weight="duotone"
            />
            <div>
              <h2 class="text-sm font-medium text-[#3730a3]">
                Diskusi pengumuman
              </h2>
              <p class="mt-1 text-xs leading-5 text-[#6366a8]">
                Buka komentar pada setiap post untuk membaca atau menambahkan
                tanggapan.
              </p>
            </div>
          </div>
        </article>
      </aside>
    </section>

    <section
      v-else
      class="flex min-h-[calc(100vh-109px)] items-center justify-center px-5 py-10"
    >
      <article
        class="w-full max-w-xl rounded-xl border border-[#ebe7df] bg-white p-8 text-center"
      >
        <div
          class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-[#eef2ff] text-[#4f46e5]"
        >
          <PhMegaphone class="h-6 w-6" weight="duotone" />
        </div>
        <h2 class="mt-3 text-base font-semibold text-[#171322]">
          Belum ada kelas aktif
        </h2>
        <p class="mx-auto mt-2 max-w-md text-sm leading-6 text-[#6b7280]">
          Feed kelas akan tersedia setelah kamu ditempatkan pada kelas aktif.
        </p>
        <RouterLink
          class="mt-5 inline-flex items-center gap-2 rounded-lg border border-[#ddd8e4] px-4 py-2 text-sm font-medium text-[#4f46e5] transition hover:bg-[#eef2ff]"
          to="/student/subjects"
        >
          Lihat mata pelajaran
        </RouterLink>
      </article>
    </section>
  </main>
</template>
