<script setup lang="ts">
import { computed, ref } from "vue";
import {
  PhArrowClockwise,
  PhChatCircleText,
  PhPaperPlaneTilt,
  PhTrash,
} from "@phosphor-icons/vue";
import {
  createFeedComment,
  deleteFeedComment,
  getFeedComments,
} from "../../services/feed";
import type { FeedComment, FeedPost } from "../../types/feed";
import { formatDateTime } from "../../utils/date";
import { useConfirmStore } from "../../stores/confirm";

const props = defineProps<{
  post: FeedPost;
}>();

const emit = defineEmits<{
  (event: "comment-count-change", feedId: string, count: number): void;
}>();

const isExpanded = ref(false);
const hasLoaded = ref(false);
const isLoading = ref(false);
const pendingSubmitCount = ref(0);
const errorMessage = ref("");
const submitErrorMessage = ref("");
const commentText = ref("");
type LocalFeedComment = FeedComment & {
  optimisticStatus?: "pending";
  localOnly?: boolean;
};

const comments = ref<LocalFeedComment[]>([]);
const deletingCommentIds = ref<Set<string>>(new Set());
const confirm = useConfirmStore();

const visibleCommentCount = computed(
  () => props.post.commentCount ?? comments.value.length,
);

function getCommentErrorMessage(error: unknown, fallback: string) {
  if (typeof error === "object" && error !== null && "response" in error) {
    const response = (
      error as {
        response?: {
          status?: number;
          data?: { error?: unknown; message?: unknown };
        };
      }
    ).response;

    if (response?.status === 403) {
      return "Anda tidak memiliki akses untuk komentar ini.";
    }

    if (response?.status === 404) {
      return "Komentar atau feed tidak ditemukan.";
    }

    if (typeof response?.data?.error === "string") {
      return response.data.error;
    }

    if (typeof response?.data?.message === "string") {
      return response.data.message;
    }
  }

  return fallback;
}

function emitCurrentCommentCount() {
  emit("comment-count-change", props.post.feedId, comments.value.length);
}

function createTempComment(content: string): LocalFeedComment {
  return {
    commentId: `temp-${Date.now()}-${Math.random().toString(36).slice(2)}`,
    sourceType: "feed",
    sourceId: props.post.feedId,
    content,
    creatorName: "Anda",
    createdAt: new Date().toISOString(),
    isMine: true,
    optimisticStatus: "pending",
  };
}

async function loadComments(
  options: {
    preserveLocal?: boolean;
    excludeTempId?: string;
    silent?: boolean;
  } = {},
) {
  if (!options.silent) {
    isLoading.value = true;
  }
  errorMessage.value = "";

  try {
    const serverComments = await getFeedComments(props.post.feedId);
    const localComments = options.preserveLocal
      ? comments.value.filter(
          (comment) =>
            comment.optimisticStatus === "pending" &&
            comment.commentId !== options.excludeTempId,
        )
      : [];

    comments.value = [...serverComments, ...localComments];
    hasLoaded.value = true;
    emitCurrentCommentCount();
    return true;
  } catch (error) {
    if (!options.silent) {
      errorMessage.value = getCommentErrorMessage(
        error,
        "Komentar belum bisa dimuat.",
      );
    }
    return false;
  } finally {
    if (!options.silent) {
      isLoading.value = false;
    }
  }
}

async function toggleComments() {
  isExpanded.value = !isExpanded.value;

  if (isExpanded.value && !hasLoaded.value) {
    await loadComments();
  }
}

async function submitComment() {
  const trimmed = commentText.value.trim();
  if (!trimmed) {
    return;
  }

  const tempComment = createTempComment(trimmed);
  comments.value = [...comments.value, tempComment];
  commentText.value = "";
  pendingSubmitCount.value += 1;
  submitErrorMessage.value = "";
  emitCurrentCommentCount();

  try {
    await createFeedComment(props.post.feedId, trimmed);
    const refreshed = await loadComments({
      preserveLocal: true,
      excludeTempId: tempComment.commentId,
      silent: true,
    });

    if (!refreshed) {
      comments.value = comments.value.map((comment) =>
        comment.commentId === tempComment.commentId
          ? { ...comment, optimisticStatus: undefined, localOnly: true }
          : comment,
      );
    }
  } catch (error) {
    comments.value = comments.value.filter(
      (comment) => comment.commentId !== tempComment.commentId,
    );
    if (!commentText.value.trim()) {
      commentText.value = trimmed;
    }
    submitErrorMessage.value = getCommentErrorMessage(
      error,
      "Komentar belum bisa dikirim.",
    );
    emitCurrentCommentCount();
  } finally {
    pendingSubmitCount.value = Math.max(0, pendingSubmitCount.value - 1);
  }
}

async function removeComment(comment: FeedComment) {
  if (!comment.isMine || deletingCommentIds.value.has(comment.commentId)) {
    return;
  }

  const ok = await confirm.confirm({
    title: "Hapus komentar?",
    description: "Komentar ini akan dihapus permanen.",
    confirmLabel: "Hapus",
    variant: "danger",
  });
  if (!ok) return;

  deletingCommentIds.value = new Set([
    ...deletingCommentIds.value,
    comment.commentId,
  ]);
  errorMessage.value = "";

  try {
    await deleteFeedComment(comment.commentId);
    comments.value = comments.value.filter(
      (item) => item.commentId !== comment.commentId,
    );
    emitCurrentCommentCount();
  } catch (error) {
    errorMessage.value = getCommentErrorMessage(
      error,
      "Komentar belum bisa dihapus.",
    );
  } finally {
    const nextDeletingIds = new Set(deletingCommentIds.value);
    nextDeletingIds.delete(comment.commentId);
    deletingCommentIds.value = nextDeletingIds;
  }
}
</script>

<template>
  <div class="mt-4 border-t border-[#ebe7df] pt-3">
    <button
      class="inline-flex items-center gap-2 rounded-2xl px-3 py-2 text-xs font-medium text-[#4f46e5] transition hover:bg-[#eef2ff] focus:outline-none focus:ring-2 focus:ring-[#4f46e5]/25"
      type="button"
      @click="toggleComments"
    >
      <PhChatCircleText :size="16" weight="duotone" />
      {{
        isExpanded
          ? "Sembunyikan komentar"
          : `Lihat komentar${visibleCommentCount ? ` (${visibleCommentCount})` : ""}`
      }}
    </button>

    <div v-if="isExpanded" class="mt-3 space-y-3 rounded-2xl bg-white/70 p-3">
      <div
        v-if="isLoading"
        class="space-y-3 rounded-xl border border-[#ebe7df] bg-[#fbfaf8] p-3"
        aria-label="Memuat komentar"
      >
        <div v-for="item in 2" :key="item" class="flex animate-pulse gap-3">
          <div class="h-7 w-7 shrink-0 rounded-full bg-[#e9e5dd]" />
          <div class="min-w-0 flex-1 space-y-2">
            <div class="flex items-center gap-2">
              <div class="h-3 w-24 rounded bg-[#e9e5dd]" />
              <div class="h-2.5 w-12 rounded bg-[#eeeae3]" />
            </div>
            <div class="h-3 w-3/4 rounded bg-[#eeeae3]" />
          </div>
        </div>
      </div>

      <div v-else-if="errorMessage" class="rounded-2xl bg-[#fff7ed] p-3">
        <p class="text-xs leading-5 text-[#9a3412]">{{ errorMessage }}</p>
        <button
          class="mt-3 inline-flex items-center gap-2 rounded-2xl border border-[#fed7aa] px-3 py-2 text-xs font-medium text-[#9a3412] transition hover:bg-[#ffedd5]"
          type="button"
          @click="() => loadComments()"
        >
          <PhArrowClockwise :size="14" />
          Coba lagi
        </button>
      </div>

      <div v-else class="space-y-3">
        <div v-if="comments.length === 0" class="rounded-2xl bg-[#fbfaf8] p-3">
          <p class="text-xs text-[#7a7385]">Belum ada komentar.</p>
        </div>

        <div
          v-for="comment in comments"
          :key="comment.commentId"
          class="rounded-2xl bg-[#fbfaf8] p-3"
        >
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <p class="truncate text-xs font-medium text-[#171322]">
                {{ comment.creatorName || "Pengirim tidak tersedia" }}
              </p>
              <p class="mt-0.5 text-[11px] text-[#a09aa8]">
                {{
                  comment.optimisticStatus === "pending"
                    ? "Mengirim..."
                    : formatDateTime(comment.createdAt)
                }}
              </p>
            </div>
            <button
              v-if="
                comment.isMine &&
                !comment.optimisticStatus &&
                !comment.localOnly
              "
              class="inline-flex shrink-0 items-center gap-1 rounded-xl px-2 py-1 text-[11px] font-medium text-[#b42318] transition hover:bg-[#fef2f2] disabled:cursor-not-allowed disabled:opacity-60"
              type="button"
              :disabled="deletingCommentIds.has(comment.commentId)"
              @click="removeComment(comment)"
            >
              <PhTrash :size="13" />
              Hapus
            </button>
          </div>
          <p
            class="mt-2 whitespace-pre-line wrap-break-word text-xs leading-5 text-[#4a4356]"
          >
            {{ comment.content }}
          </p>
        </div>
      </div>

      <form class="space-y-2" @submit.prevent="submitComment">
        <label class="sr-only" :for="`feed-comment-${post.feedId}`"
          >Tulis komentar</label
        >
        <textarea
          :id="`feed-comment-${post.feedId}`"
          v-model="commentText"
          class="min-h-20 w-full resize-y rounded-2xl border border-[#ebe7df] bg-white px-3 py-2 text-xs leading-5 text-[#171322] outline-none transition placeholder:text-[#a09aa8] focus:border-[#4f46e5]"
          maxlength="800"
          placeholder="Tulis komentar singkat..."
        />
        <p
          v-if="submitErrorMessage"
          class="text-[11px] font-medium text-[#b42318]"
        >
          {{ submitErrorMessage }}
        </p>
        <div class="flex items-center justify-between gap-3">
          <p class="text-[11px] text-[#8b8592]">
            Komentar hanya untuk feed kelas.
          </p>
          <button
            class="inline-flex items-center gap-2 rounded-2xl bg-[#4f46e5] px-3 py-2 text-xs font-medium text-white transition hover:bg-[#4338ca] disabled:cursor-not-allowed disabled:opacity-60"
            type="submit"
            :disabled="!commentText.trim()"
          >
            <PhPaperPlaneTilt :size="14" weight="duotone" />
            {{ pendingSubmitCount > 0 ? "Kirim lagi" : "Kirim" }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
