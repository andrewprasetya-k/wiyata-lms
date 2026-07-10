<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import {
  PhArrowClockwise,
  PhChatCircleText,
  PhChatsCircle,
  PhPaperPlaneTilt,
  PhTrash,
} from "@phosphor-icons/vue";
import {
  createComment,
  deleteComment,
  getComments,
} from "../../services/comments";
import type { CommentItem, CommentSourceType } from "../../types/comment";
import { formatDateTime } from "../../utils/date";
import { useConfirmStore } from "../../stores/confirm";

type LocalComment = CommentItem & {
  optimisticStatus?: "pending";
  localOnly?: boolean;
};

const props = withDefaults(
  defineProps<{
    sourceType: CommentSourceType;
    sourceId: string;
    title?: string;
    placeholder?: string;
    emptyText?: string;
    // Feed posts render this thread collapsed inline under the post, with a
    // toggle to expand. Material/assignment discussions render it as an
    // always-visible standalone card. The two source types map 1:1 to these
    // UI modes in every current usage, so we derive the mode from sourceType
    // instead of adding a second prop that would just duplicate it.
    initialCount?: number;
  }>(),
  {
    title: "Diskusi",
    placeholder: "Tulis komentar atau pertanyaan singkat...",
    emptyText: "Belum ada diskusi.",
  },
);

const emit = defineEmits<{
  (event: "count-change", count: number): void;
}>();

const isFeedVariant = computed(() => props.sourceType === "feed");

const isExpanded = ref(!isFeedVariant.value);
const hasLoaded = ref(false);
const isLoading = ref(false);
const errorMessage = ref("");
const submitErrorMessage = ref("");
const commentText = ref("");
const pendingSubmitCount = ref(0);
const comments = ref<LocalComment[]>([]);
const deletingCommentIds = ref<Set<string>>(new Set());
const confirm = useConfirmStore();
let loadRequestId = 0;

const canSubmit = computed(() => commentText.value.trim().length > 0);
const discussionId = computed(
  () => `discussion-${props.sourceType}-${props.sourceId}`,
);
const visibleCommentCount = computed(
  () => props.initialCount ?? comments.value.length,
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
      return isFeedVariant.value
        ? "Anda tidak memiliki akses untuk komentar ini."
        : "Anda tidak memiliki akses ke diskusi ini.";
    }

    if (response?.status === 404) {
      return isFeedVariant.value
        ? "Komentar atau feed tidak ditemukan."
        : "Diskusi atau konten tidak ditemukan.";
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

function emitCurrentCount() {
  emit("count-change", comments.value.length);
}

function createTempComment(content: string): LocalComment {
  return {
    commentId: `temp-${Date.now()}-${Math.random().toString(36).slice(2)}`,
    sourceType: props.sourceType,
    sourceId: props.sourceId,
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
  if (!props.sourceId) return false;

  const currentRequestId = ++loadRequestId;
  if (!options.silent) {
    isLoading.value = true;
  }
  errorMessage.value = "";

  try {
    const serverComments = await getComments({
      sourceType: props.sourceType,
      sourceId: props.sourceId,
    });
    if (currentRequestId !== loadRequestId) return false;

    const localComments = options.preserveLocal
      ? comments.value.filter(
          (comment) =>
            comment.optimisticStatus === "pending" &&
            comment.commentId !== options.excludeTempId,
        )
      : [];

    comments.value = [...serverComments, ...localComments];
    hasLoaded.value = true;
    emitCurrentCount();
    return true;
  } catch (error) {
    if (!options.silent && currentRequestId === loadRequestId) {
      errorMessage.value = getCommentErrorMessage(
        error,
        isFeedVariant.value
          ? "Komentar belum bisa dimuat."
          : "Diskusi belum bisa dimuat.",
      );
    }
    return false;
  } finally {
    if (!options.silent && currentRequestId === loadRequestId) {
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
  if (!trimmed || !props.sourceId) return;

  const tempComment = createTempComment(trimmed);
  comments.value = [...comments.value, tempComment];
  commentText.value = "";
  pendingSubmitCount.value += 1;
  submitErrorMessage.value = "";
  emitCurrentCount();

  try {
    await createComment({
      sourceType: props.sourceType,
      sourceId: props.sourceId,
      content: trimmed,
    });
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
    emitCurrentCount();
  } finally {
    pendingSubmitCount.value = Math.max(0, pendingSubmitCount.value - 1);
  }
}

async function removeComment(comment: CommentItem) {
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
    await deleteComment(comment.commentId);
    comments.value = comments.value.filter(
      (item) => item.commentId !== comment.commentId,
    );
    emitCurrentCount();
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

watch(
  () => [props.sourceType, props.sourceId],
  () => {
    // The feed variant never re-targets an already-mounted thread to a
    // different post, so this mirrors the discussion-only reset behavior
    // the two original components had before merging.
    if (isFeedVariant.value) return;
    comments.value = [];
    hasLoaded.value = false;
    errorMessage.value = "";
    submitErrorMessage.value = "";
    void loadComments();
  },
);

onMounted(() => {
  if (!isFeedVariant.value) {
    void loadComments();
  }
});
</script>

<template>
  <div v-if="isFeedVariant" class="mt-4 border-t border-border pt-3">
    <button
      class="inline-flex items-center gap-2 rounded-2xl px-3 py-2 text-xs font-medium text-brand transition hover:bg-[#eef2ff] focus:outline-none focus:ring-2 focus:ring-brand/25"
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
        class="space-y-3 rounded-xl border border-border bg-[#fbfaf8] p-3"
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
              <p class="truncate text-xs font-medium text-foreground">
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
        <label class="sr-only" :for="discussionId">Tulis komentar</label>
        <textarea
          :id="discussionId"
          v-model="commentText"
          class="min-h-20 w-full resize-y rounded-2xl border border-border bg-white px-3 py-2 text-xs leading-5 text-foreground outline-none transition placeholder:text-[#a09aa8] focus:border-brand"
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
            class="inline-flex items-center gap-2 rounded-2xl bg-brand px-3 py-2 text-xs font-medium text-white transition hover:bg-[#4338ca] disabled:cursor-not-allowed disabled:opacity-60"
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

  <article v-else class="rounded-xl border border-border bg-white p-5 sm:p-6">
    <div class="flex items-start justify-between gap-3">
      <div class="min-w-0">
        <div class="flex items-center gap-2">
          <PhChatsCircle :size="18" class="text-brand" weight="duotone" />
          <h2 class="text-sm font-medium text-foreground">{{ title }}</h2>
        </div>
        <p class="mt-1 text-xs leading-5 text-[#7a7385]">
          Ajukan pertanyaan atau lanjutkan pembahasan terkait konten ini.
        </p>
      </div>
      <span
        v-if="hasLoaded"
        class="shrink-0 rounded-full bg-[#f8f7f4] px-2.5 py-1 text-[11px] text-muted"
      >
        {{ comments.length }} komentar
      </span>
    </div>

    <div
      v-if="isLoading"
      class="mt-4 space-y-3 rounded-xl border border-border bg-[#fbfaf8] p-3"
      aria-label="Memuat diskusi"
    >
      <div v-for="item in 2" :key="item" class="flex animate-pulse gap-3">
        <div class="h-8 w-8 shrink-0 rounded-full bg-[#e9e5dd]" />
        <div class="min-w-0 flex-1 space-y-2">
          <div class="flex items-center gap-2">
            <div class="h-3 w-24 rounded bg-[#e9e5dd]" />
            <div class="h-2.5 w-12 rounded bg-[#eeeae3]" />
          </div>
          <div class="h-3 w-4/5 rounded bg-[#eeeae3]" />
        </div>
      </div>
    </div>

    <div v-else-if="errorMessage" class="mt-4 rounded-xl bg-[#fff7ed] p-4">
      <p class="text-sm leading-6 text-[#9a3412]">{{ errorMessage }}</p>
      <button
        class="mt-3 inline-flex items-center gap-2 rounded-lg border border-[#fed7aa] px-3 py-2 text-xs font-medium text-[#9a3412] transition hover:bg-[#ffedd5]"
        type="button"
        @click="() => loadComments()"
      >
        <PhArrowClockwise :size="14" />
        Coba lagi
      </button>
    </div>

    <div v-else class="mt-4 space-y-3">
      <div v-if="comments.length === 0" class="rounded-lg bg-[#fbfaf8] p-3">
        <p class="text-sm leading-6 text-muted">{{ emptyText }}</p>
      </div>

      <div
        v-for="comment in comments"
        :key="comment.commentId"
        class="rounded-xl bg-[#fbfaf8] p-4"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <p class="truncate text-sm font-medium text-foreground">
              {{ comment.creatorName || "Pengirim tidak tersedia" }}
            </p>
            <p class="mt-0.5 text-xs text-[#a09aa8]">
              {{
                comment.optimisticStatus === "pending"
                  ? formatDateTime(new Date().toLocaleString())
                  : formatDateTime(comment.createdAt)
              }}
            </p>
          </div>
          <button
            v-if="
              comment.isMine && !comment.optimisticStatus && !comment.localOnly
            "
            class="inline-flex shrink-0 items-center gap-1 rounded-lg px-2 py-1 text-xs font-medium text-[#b42318] transition hover:bg-[#fef2f2] disabled:cursor-not-allowed disabled:opacity-60"
            type="button"
            :disabled="deletingCommentIds.has(comment.commentId)"
            @click="removeComment(comment)"
          >
            <PhTrash :size="14" />
            Hapus
          </button>
        </div>
        <p
          class="mt-3 whitespace-pre-line wrap-break-word text-sm leading-6 text-[#4a4356]"
        >
          {{ comment.content }}
        </p>
      </div>
    </div>

    <form class="mt-4 space-y-2" @submit.prevent="submitComment">
      <label class="sr-only" :for="discussionId">Tulis komentar</label>
      <textarea
        :id="discussionId"
        v-model="commentText"
        class="min-h-24 w-full resize-y rounded-xl border border-border bg-white px-3 py-2 text-sm leading-6 text-foreground outline-none transition placeholder:text-[#a09aa8] focus:border-brand"
        maxlength="800"
        :placeholder="placeholder"
      />
      <p v-if="submitErrorMessage" class="text-xs font-medium text-[#b42318]">
        {{ submitErrorMessage }}
      </p>
      <div class="flex items-center justify-between gap-3">
        <p class="text-xs text-[#8b8592]">
          Diskusi terlihat oleh peserta yang memiliki akses ke konten ini.
        </p>
        <button
          class="inline-flex shrink-0 items-center gap-2 rounded-lg bg-brand px-3 py-2 text-sm font-medium text-white transition hover:bg-[#4338ca] disabled:cursor-not-allowed disabled:opacity-60"
          type="submit"
          :disabled="!canSubmit"
        >
          <PhPaperPlaneTilt :size="15" weight="duotone" />
          {{ pendingSubmitCount > 0 ? "Kirim lagi" : "Kirim" }}
        </button>
      </div>
    </form>
  </article>
</template>
