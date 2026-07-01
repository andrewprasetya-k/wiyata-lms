<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import {
  PhArrowClockwise,
  PhChalkboardTeacher,
  PhMegaphone,
  PhPaperPlaneTilt,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import FeedComments from "../../components/feed/FeedComments.vue";
import { getMyTeachingSubjectClasses } from "../../services/teacherSubjects";
import {
  createClassFeed,
  getClassFeed,
  markFeedNotificationsRead,
} from "../../services/feed";
import {
  clearFeedUnreadOptimistically,
  restoreFeedUnreadCount,
} from "../../composables/useFeedUnreadCount";
import type { TeacherSubjectClass } from "../../types/teacherSubjects";
import type { FeedClassHeader, FeedPost } from "../../types/feed";
import { useAuthStore } from "../../stores/auth";
import { useToastStore } from "../../stores/toast";
import { formatDateTime } from "../../utils/date";

interface TeacherFeedClass {
  classId: string;
  className: string;
  subjectCount: number;
}

const auth = useAuthStore();
const toast = useToastStore();

const classes = ref<TeacherFeedClass[]>([]);
const selectedClassId = ref("");
const classHeader = ref<FeedClassHeader | null>(null);
const posts = ref<FeedPost[]>([]);
const content = ref("");

const classesLoading = ref(false);
const feedLoading = ref(false);
const submitting = ref(false);
const classesError = ref("");
const feedError = ref("");
const feedAccessMessage = ref("");

const activeSchoolId = computed(
  () => auth.activeSchoolId ?? auth.defaultContext?.schoolId ?? "",
);
const selectedClass = computed(
  () =>
    classes.value.find((item) => item.classId === selectedClassId.value) ??
    null,
);
const canSubmit = computed(
  () =>
    Boolean(
      activeSchoolId.value && selectedClassId.value && content.value.trim(),
    ) &&
    !submitting.value &&
    !feedAccessMessage.value,
);

function mapTeachingClasses(subjects: TeacherSubjectClass[]) {
  const classMap = new Map<string, TeacherFeedClass>();

  for (const subject of subjects) {
    const current = classMap.get(subject.classId);
    if (current) {
      current.subjectCount += 1;
      continue;
    }

    classMap.set(subject.classId, {
      classId: subject.classId,
      className: subject.className || "Kelas",
      subjectCount: 1,
    });
  }

  return [...classMap.values()].sort((a, b) =>
    (a.className || "").localeCompare(b.className || ""),
  );
}

function getErrorMessage(error: unknown, fallback: string) {
  if (
    typeof error === "object" &&
    error !== null &&
    "response" in error &&
    typeof (error as { response?: { data?: { error?: unknown } } }).response
      ?.data?.error === "string"
  ) {
    return (error as { response: { data: { error: string } } }).response.data
      .error;
  }

  return fallback;
}

function isForbiddenError(error: unknown) {
  return (
    typeof error === "object" &&
    error !== null &&
    "response" in error &&
    (error as { response?: { status?: number } }).response?.status === 403
  );
}

async function loadClasses() {
  classesLoading.value = true;
  classesError.value = "";

  try {
    const subjects = await getMyTeachingSubjectClasses();
    classes.value = mapTeachingClasses(subjects);
    selectedClassId.value = classes.value[0]?.classId ?? "";
  } catch (error) {
    classes.value = [];
    selectedClassId.value = "";
    classesError.value = getErrorMessage(
      error,
      "Kelas yang diajar belum bisa dimuat.",
    );
  } finally {
    classesLoading.value = false;
  }
}

async function loadFeed() {
  if (!selectedClassId.value) {
    posts.value = [];
    classHeader.value = null;
    feedError.value = "";
    feedAccessMessage.value = "";
    return;
  }

  feedLoading.value = true;
  feedError.value = "";
  feedAccessMessage.value = "";

  try {
    const response = await getClassFeed(selectedClassId.value);
    classHeader.value = response.class;
    posts.value = response.data.data ?? [];
    void markCurrentFeedRead();
  } catch (error) {
    posts.value = [];
    classHeader.value = null;
    if (isForbiddenError(error)) {
      feedAccessMessage.value =
        "Pengumuman kelas ini belum bisa dimuat. Pastikan guru masih aktif di Penempatan Kelas.";
    } else {
      feedError.value = getErrorMessage(
        error,
        "Pengumuman kelas belum bisa dimuat.",
      );
    }
  } finally {
    feedLoading.value = false;
  }
}

async function submitFeed() {
  if (!activeSchoolId.value) {
    toast.error("Konteks sekolah aktif belum tersedia.");
    return;
  }
  if (!selectedClassId.value) {
    toast.error("Pilih kelas terlebih dahulu.");
    return;
  }
  if (!content.value.trim()) {
    toast.error("Isi pengumuman wajib diisi.");
    return;
  }

  submitting.value = true;
  const submittedClassId = selectedClassId.value;

  try {
    const response = await createClassFeed({
      schoolId: activeSchoolId.value,
      classId: submittedClassId,
      content: content.value.trim(),
    });
    toast.success("Pengumuman kelas berhasil dikirim.");
    content.value = "";
    if (response.feed && selectedClassId.value === submittedClassId) {
      prependCreatedFeed(response.feed);
    }
    void refreshFeedAfterCreate();
  } catch (error) {
    if (isForbiddenError(error)) {
      feedAccessMessage.value =
        "Pengumuman kelas ini belum bisa dimuat. Pastikan guru masih aktif di Penempatan Kelas.";
    } else {
      toast.error(getErrorMessage(error, "Pengumuman belum bisa dikirim."));
    }
  } finally {
    submitting.value = false;
  }
}

function prependCreatedFeed(feed: FeedPost) {
  if (posts.value.some((post) => post.feedId === feed.feedId)) {
    return;
  }

  posts.value = [feed, ...posts.value];
}

async function refreshFeedAfterCreate() {
  try {
    await loadFeed();
  } catch {
    // loadFeed owns visible feed error state; create already succeeded.
  }
}

watch(selectedClassId, () => {
  loadFeed();
});

onMounted(async () => {
  await loadClasses();
  await loadFeed();
});

async function markCurrentFeedRead() {
  const previousUnreadCount = clearFeedUnreadOptimistically();

  try {
    await markFeedNotificationsRead();
  } catch {
    restoreFeedUnreadCount(previousUnreadCount);
    // Feed read marker should not block the feed page.
  }
}

function updatePostCommentCount(feedId: string, count: number) {
  posts.value = posts.value.map((post) =>
    post.feedId === feedId ? { ...post, commentCount: count } : post,
  );
}
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-[#f8f7f4]">
    <header class="border-b border-[#ebe7df] bg-white">
      <div class="px-5 py-5 sm:px-6 lg:px-8">
        <div
          class="mt-2 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between"
        >
          <div class="min-w-0">
            <h1 class="text-2xl font-semibold text-[#171322] sm:text-3xl">
              Pengumuman Kelas
            </h1>
            <p class="mt-2 max-w-2xl text-sm leading-6 text-[#6b6475]">
              Sampaikan informasi kepada siswa dan lanjutkan diskusi melalui
              komentar pada setiap pengumuman.
            </p>
          </div>
          <span
            v-if="selectedClass"
            class="inline-flex max-w-full items-center gap-2 self-start rounded-lg bg-[#eef7f2] px-3 py-2 text-xs font-medium text-[#2f7d5c] sm:self-auto"
          >
            <PhChalkboardTeacher :size="15" weight="duotone" />
            <span class="truncate">{{ selectedClass.className }}</span>
          </span>
        </div>
      </div>
    </header>

    <section class="px-5 py-5 sm:px-6 lg:px-8">
      <template v-if="classesLoading">
        <div class="grid gap-5 lg:grid-cols-[minmax(0,1fr)_360px]">
          <div
            class="h-72 animate-pulse rounded-xl border border-[#ebe7df] bg-white"
          />
          <div
            class="h-72 animate-pulse rounded-xl border border-[#ebe7df] bg-white"
          />
        </div>
      </template>

      <section
        v-else-if="classesError"
        class="mx-auto max-w-xl rounded-xl border border-[#f0d8d2] bg-white px-5 py-8 text-center"
      >
        <PhWarningCircle
          :size="30"
          class="mx-auto text-[#d97757]"
          weight="duotone"
        />
        <h2 class="mt-3 text-lg font-semibold text-[#171322]">
          Kelas belum bisa dimuat
        </h2>
        <p class="mt-2 text-sm leading-6 text-[#6b6475]">
          {{ classesError }}
        </p>
        <button
          class="mt-5 inline-flex items-center gap-2 rounded-lg bg-[#171322] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#2f2b3a]"
          type="button"
          @click="loadClasses"
        >
          <PhArrowClockwise :size="16" />
          Coba lagi
        </button>
      </section>

      <section
        v-else-if="classes.length === 0"
        class="mx-auto max-w-xl rounded-xl border border-[#ebe7df] bg-white px-5 py-10 text-center"
      >
        <div
          class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-[#eef2ff] text-[#4f46e5]"
        >
          <PhChalkboardTeacher :size="24" weight="duotone" />
        </div>
        <h2 class="mt-3 text-lg font-semibold text-[#171322]">
          Belum ada kelas aktif
        </h2>
        <p class="mt-2 text-sm leading-6 text-[#6b6475]">
          Belum ada kelas aktif yang bisa digunakan untuk mengirim pengumuman.
        </p>
      </section>

      <section
        v-else
        class="grid min-w-0 gap-5 lg:grid-cols-[minmax(0,1fr)_360px]"
      >
        <aside class="order-1 min-w-0 lg:order-2">
          <article
            class="rounded-xl border border-[#ebe7df] bg-white p-5 lg:sticky lg:top-6"
          >
            <div class="flex items-start gap-3">
              <div
                class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-[#eef2ff] text-[#4f46e5]"
              >
                <PhMegaphone :size="20" weight="duotone" />
              </div>
              <div class="min-w-0">
                <h2 class="text-base font-semibold text-[#171322]">
                  Buat pengumuman
                </h2>
                <p class="mt-1 text-xs leading-5 text-[#7a7385]">
                  Pilih kelas dan tulis informasi yang perlu diketahui siswa.
                </p>
              </div>
            </div>

            <label
              class="mt-5 block text-xs font-medium text-[#6b6475]"
              for="feed-class"
            >
              Kelas tujuan
            </label>
            <select
              id="feed-class"
              v-model="selectedClassId"
              class="mt-2 w-full rounded-lg border border-[#ebe7df] bg-white px-3.5 py-2.5 text-sm text-[#171322] outline-none transition focus:border-[#4f46e5]"
            >
              <option
                v-for="item in classes"
                :key="item.classId"
                :value="item.classId"
              >
                {{ item.className }}
              </option>
            </select>

            <label
              class="mt-4 block text-xs font-medium text-[#6b6475]"
              for="feed-content"
            >
              Isi pengumuman
            </label>
            <textarea
              id="feed-content"
              v-model="content"
              class="mt-2 min-h-40 w-full resize-y rounded-lg border border-[#ebe7df] bg-[#fbfaf8] px-3.5 py-3 text-sm leading-6 text-[#171322] outline-none transition placeholder:text-[#a09aa8] focus:border-[#4f46e5] focus:bg-white"
              placeholder="Tulis pengumuman untuk kelas ini..."
              maxlength="1200"
            />
            <p class="mt-2 text-xs leading-5 text-[#8b8592]">
              <span v-if="feedAccessMessage">
                Pengumuman belum bisa dikirim untuk kelas ini.
              </span>
              <span v-else>
                Pengumuman ini akan terlihat oleh siswa aktif di kelas.
              </span>
            </p>
            <button
              class="mt-4 inline-flex w-full items-center justify-center gap-2 rounded-lg bg-[#4f46e5] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#4338ca] disabled:cursor-not-allowed disabled:opacity-60"
              type="button"
              :disabled="!canSubmit"
              @click="submitFeed"
            >
              <PhPaperPlaneTilt :size="16" weight="duotone" />
              {{ submitting ? "Mengirim..." : "Kirim pengumuman" }}
            </button>
          </article>
        </aside>

        <section class="order-2 min-w-0 lg:order-1">
          <div
            class="flex flex-col gap-3 rounded-xl border border-[#ebe7df] bg-white px-4 py-4 sm:flex-row sm:items-center sm:justify-between sm:px-5"
          >
            <div class="min-w-0">
              <h2 class="truncate text-base font-semibold text-[#171322]">
                {{
                  classHeader?.classTitle ||
                  selectedClass?.className ||
                  "Pengumuman kelas"
                }}
              </h2>
              <p class="mt-1 text-xs text-[#7a7385]">
                {{ selectedClass?.subjectCount || 0 }} mata pelajaran yang Anda
                ajar
              </p>
            </div>
            <button
              class="inline-flex shrink-0 items-center justify-center gap-2 rounded-lg border border-[#ebe7df] px-3 py-2 text-xs font-medium text-[#4f46e5] transition hover:border-[#4f46e5] hover:bg-[#eef2ff] disabled:opacity-50"
              type="button"
              :disabled="feedLoading"
              @click="loadFeed"
            >
              <PhArrowClockwise :size="14" />
              Muat ulang
            </button>
          </div>

          <div v-if="feedLoading" class="mt-3 space-y-3">
            <div
              v-for="item in 3"
              :key="item"
              class="h-32 animate-pulse rounded-xl border border-[#ebe7df] bg-white"
            />
          </div>

          <div
            v-else-if="feedAccessMessage"
            class="mt-3 rounded-xl border border-[#ebe7df] bg-white p-5"
          >
            <h3 class="text-sm font-semibold text-[#171322]">
              Pengumuman belum bisa dimuat
            </h3>
            <p class="mt-2 text-sm leading-6 text-[#7a7385]">
              {{ feedAccessMessage }}
            </p>
          </div>

          <div
            v-else-if="feedError"
            class="mt-3 rounded-xl border border-[#fed7aa] bg-[#fff7ed] p-5"
          >
            <h3 class="text-sm font-semibold text-[#9a3412]">
              Terjadi kendala
            </h3>
            <p class="mt-2 text-sm leading-6 text-[#9a3412]">
              {{ feedError }}
            </p>
            <button
              type="button"
              class="mt-4 inline-flex items-center gap-2 rounded-lg border border-[#fdba74] bg-white px-3 py-2 text-xs font-medium text-[#9a3412]"
              @click="loadFeed"
            >
              <PhArrowClockwise :size="14" />
              Coba lagi
            </button>
          </div>

          <div
            v-else-if="posts.length === 0"
            class="mt-3 rounded-xl border border-[#ebe7df] bg-white px-5 py-10 text-center"
          >
            <div
              class="mx-auto flex h-11 w-11 items-center justify-center rounded-xl bg-[#eef2ff] text-[#4f46e5]"
            >
              <PhMegaphone :size="22" weight="duotone" />
            </div>
            <h3 class="mt-3 text-base font-semibold text-[#171322]">
              Belum ada pengumuman
            </h3>
            <p class="mx-auto mt-2 max-w-md text-sm leading-6 text-[#7a7385]">
              Pengumuman yang Anda kirim untuk kelas ini akan tampil di sini dan
              dapat dibaca oleh siswa aktif.
            </p>
          </div>

          <div v-else class="mt-3 space-y-3">
            <article
              v-for="post in posts"
              :key="post.feedId"
              class="min-w-0 rounded-xl border border-[#ebe7df] bg-white"
            >
              <div class="flex min-w-0 items-start gap-3 px-4 pt-4 sm:px-5">
                <div
                  class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-[#059669] text-xs font-semibold text-white"
                >
                  {{ (post.creatorName || "G").charAt(0).toUpperCase() }}
                </div>
                <div class="min-w-0 flex-1">
                  <div
                    class="flex min-w-0 flex-col gap-1 sm:flex-row sm:items-center sm:justify-between"
                  >
                    <h3 class="truncate text-sm font-semibold text-[#171322]">
                      {{ post.creatorName || "Pengirim tidak tersedia" }}
                    </h3>
                    <span class="shrink-0 text-xs text-[#a09aa8]">
                      {{ formatDateTime(post.createdAt) }}
                    </span>
                  </div>
                  <span
                    class="mt-2 inline-flex rounded-lg bg-[#fff7ed] px-2 py-1 text-[10px] font-medium uppercase tracking-wide text-[#ea580c]"
                  >
                    Pengumuman
                  </span>
                </div>
              </div>
              <p
                class="whitespace-pre-line wrap-break-word px-4 pb-4 pt-3 text-sm leading-6 text-[#4a4356] sm:px-5"
              >
                {{ post.content }}
              </p>
              <div class="border-t border-[#f3f1ec] px-4 pb-4 sm:px-5">
                <FeedComments
                  :post="post"
                  @comment-count-change="updatePostCommentCount"
                />
              </div>
            </article>
          </div>
        </section>
      </section>
    </section>
  </main>
</template>
