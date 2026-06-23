<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import {
  PhArrowClockwise,
  PhChalkboardTeacher,
  PhMegaphone,
  PhPaperPlaneTilt,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { getMyTeachingSubjectClasses } from "../../services/teacherSubjects";
import { createClassFeed, getClassFeed } from "../../services/feed";
import type { TeacherSubjectClass } from "../../types/teacherSubjects";
import type { FeedClassHeader, FeedPost } from "../../types/feed";
import { useAuthStore } from "../../stores/auth";
import { useToastStore } from "../../stores/toast";
import { formatDateTime } from "../../utils/date";

interface TeacherFeedClass {
  classId: string;
  className: string;
  classCode?: string;
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

const activeSchoolId = computed(
  () => auth.activeSchoolId ?? auth.defaultContext?.schoolId ?? "",
);
const selectedClass = computed(
  () => classes.value.find((item) => item.classId === selectedClassId.value) ?? null,
);
const canSubmit = computed(
  () => Boolean(activeSchoolId.value && selectedClassId.value && content.value.trim()) && !submitting.value,
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
      classCode: subject.classCode,
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

function classOptionLabel(item: TeacherFeedClass) {
  return item.classCode ? `${item.className} · ${item.classCode}` : item.className;
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
    return;
  }

  feedLoading.value = true;
  feedError.value = "";

  try {
    const response = await getClassFeed(selectedClassId.value);
    classHeader.value = response.class;
    posts.value = response.data.data ?? [];
  } catch (error) {
    posts.value = [];
    classHeader.value = null;
    feedError.value = getErrorMessage(
      error,
      "Feed kelas belum bisa dimuat.",
    );
  } finally {
    feedLoading.value = false;
  }
}

async function submitFeed() {
  if (!activeSchoolId.value) {
    toast.error("Context sekolah aktif belum tersedia.");
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

  try {
    await createClassFeed({
      schoolId: activeSchoolId.value,
      classId: selectedClassId.value,
      content: content.value.trim(),
    });
    toast.success("Pengumuman kelas berhasil dikirim.");
    content.value = "";
    await loadFeed();
  } catch (error) {
    toast.error(getErrorMessage(error, "Pengumuman belum bisa dikirim."));
  } finally {
    submitting.value = false;
  }
}

watch(selectedClassId, () => {
  loadFeed();
});

onMounted(async () => {
  await loadClasses();
  await loadFeed();
});
</script>

<template>
  <main class="min-h-screen flex-1 px-5 py-5 sm:px-6 lg:px-8">
    <section class="flex w-full max-w-none flex-col gap-5">
      <header
        class="rounded-[22px] bg-[#f0e9dd] px-5 py-5 shadow-sm ring-1 ring-black/5 md:px-6"
      >
        <p class="text-sm font-medium text-[#8a6d3b]">Feed kelas</p>
        <h1 class="mt-3 text-3xl font-medium text-[#171322] md:text-4xl">
          Pengumuman kelas
        </h1>
        <p class="mt-3 max-w-2xl text-sm leading-6 text-[#6b6475]">
          Buat pengumuman untuk kelas yang Anda ajar. Komentar, attachment, dan
          realtime feed belum termasuk dalam MVP ini.
        </p>
      </header>

      <section
        v-if="classesLoading"
        class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5"
      >
        <p class="text-sm text-[#6b6475]">Memuat kelas yang diajar...</p>
      </section>

      <section
        v-else-if="classesError"
        class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5"
      >
        <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div class="flex items-start gap-3">
            <PhWarningCircle
              :size="24"
              class="mt-0.5 text-[#e58f86]"
              weight="duotone"
            />
            <div>
              <h2 class="text-lg font-medium text-[#171322]">
                Gagal memuat kelas
              </h2>
              <p class="mt-2 text-sm leading-6 text-[#6b6475]">
                {{ classesError }}
              </p>
            </div>
          </div>
          <button
            class="inline-flex items-center gap-2 rounded-2xl border border-[#d8d1c5] px-4 py-2 text-sm font-medium text-[#4f46e5] transition hover:border-[#4f46e5] hover:bg-[#eef2ff]"
            type="button"
            @click="loadClasses"
          >
            <PhArrowClockwise :size="16" />
            Coba lagi
          </button>
        </div>
      </section>

      <section
        v-else-if="classes.length === 0"
        class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5"
      >
        <div
          class="mb-4 flex h-12 w-12 items-center justify-center rounded-2xl bg-[#eef2ff] text-[#4f46e5]"
        >
          <PhChalkboardTeacher :size="24" weight="duotone" />
        </div>
        <p class="text-sm font-medium text-[#171322]">
          Belum ada kelas yang diajar
        </p>
        <p class="mt-2 text-sm leading-6 text-[#6b6475]">
          Feed dapat dibuat setelah admin menugaskan Anda ke subject class dan
          enrollment teacher di kelas tersebut masih aktif.
        </p>
      </section>

      <template v-else>
        <section
          class="grid gap-5 lg:grid-cols-[minmax(0,0.92fr)_minmax(0,1.08fr)]"
        >
          <article class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5">
            <div class="mb-5 flex items-start gap-3">
              <div
                class="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-[#eef2ff] text-[#4f46e5]"
              >
                <PhMegaphone :size="22" weight="duotone" />
              </div>
              <div>
                <p class="text-sm font-medium text-[#171322]">
                  Buat pengumuman
                </p>
                <p class="mt-1 text-xs leading-5 text-[#7a7385]">
                  Pilih kelas, lalu tulis pengumuman singkat untuk semua member
                  aktif di kelas tersebut.
                </p>
              </div>
            </div>

            <label class="text-xs font-medium text-[#6b6475]" for="feed-class">
              Kelas
            </label>
            <select
              id="feed-class"
              v-model="selectedClassId"
              class="mt-2 w-full rounded-2xl border border-[#ebe7df] bg-white px-4 py-3 text-sm text-[#171322] outline-none transition focus:border-[#4f46e5]"
            >
              <option
                v-for="item in classes"
                :key="item.classId"
                :value="item.classId"
              >
                {{ classOptionLabel(item) }}
              </option>
            </select>

            <label
              class="mt-5 block text-xs font-medium text-[#6b6475]"
              for="feed-content"
            >
              Isi pengumuman
            </label>
            <textarea
              id="feed-content"
              v-model="content"
              class="mt-2 min-h-36 w-full resize-y rounded-2xl border border-[#ebe7df] bg-white px-4 py-3 text-sm leading-6 text-[#171322] outline-none transition placeholder:text-[#a09aa8] focus:border-[#4f46e5]"
              placeholder="Tulis pengumuman untuk kelas ini..."
              maxlength="1200"
            />
            <div class="mt-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <p class="text-xs text-[#8b8592]">
                Attachment dan komentar belum diaktifkan untuk feed MVP.
              </p>
              <button
                class="inline-flex items-center justify-center gap-2 rounded-2xl bg-[#4f46e5] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#4338ca] disabled:cursor-not-allowed disabled:opacity-60"
                type="button"
                :disabled="!canSubmit"
                @click="submitFeed"
              >
                <PhPaperPlaneTilt :size="16" weight="duotone" />
                {{ submitting ? "Mengirim..." : "Kirim pengumuman" }}
              </button>
            </div>
          </article>

          <article class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5">
            <div class="mb-5 flex items-start justify-between gap-4">
              <div>
                <p class="text-sm font-medium text-[#171322]">
                  {{ classHeader?.classTitle || selectedClass?.className || "Feed kelas" }}
                </p>
                <p class="mt-1 text-xs text-[#7a7385]">
                  <span v-if="classHeader?.classCode || selectedClass?.classCode">
                    {{ classHeader?.classCode || selectedClass?.classCode }} ·
                  </span>
                  {{ selectedClass?.subjectCount || 0 }} subject yang Anda ajar
                </p>
              </div>
              <button
                class="inline-flex items-center gap-2 rounded-2xl border border-[#ebe7df] px-3 py-2 text-xs font-medium text-[#4f46e5] transition hover:border-[#4f46e5] hover:bg-[#eef2ff]"
                type="button"
                :disabled="feedLoading"
                @click="loadFeed"
              >
                <PhArrowClockwise :size="14" />
                Refresh
              </button>
            </div>

            <div v-if="feedLoading" class="space-y-3">
              <div
                v-for="item in 3"
                :key="item"
                class="h-24 animate-pulse rounded-2xl bg-[#fbfaf8]"
              />
            </div>

            <div
              v-else-if="feedError"
              class="rounded-2xl bg-[#fff7ed] p-4 text-sm leading-6 text-[#9a3412]"
            >
              {{ feedError }}
            </div>

            <div
              v-else-if="posts.length === 0"
              class="rounded-2xl bg-[#fbfaf8] p-4"
            >
              <p class="text-sm font-medium text-[#171322]">Belum ada feed</p>
              <p class="mt-2 text-sm leading-6 text-[#7a7385]">
                Pengumuman yang Anda kirim untuk kelas ini akan tampil di sini
                dan bisa dibaca siswa yang masih aktif di kelas.
              </p>
            </div>

            <div v-else class="space-y-3">
              <article
                v-for="post in posts"
                :key="post.feedId"
                class="rounded-2xl border border-[#ebe7df] bg-[#fbfaf8] p-4"
              >
                <div class="flex items-start justify-between gap-3">
                  <p class="text-sm font-medium text-[#171322]">
                    {{ post.creatorName || "Pengirim tidak tersedia" }}
                  </p>
                  <span class="shrink-0 text-xs text-[#a09aa8]">
                    {{ formatDateTime(post.createdAt) }}
                  </span>
                </div>
                <p class="mt-3 whitespace-pre-line text-sm leading-6 text-[#4a4356]">
                  {{ post.content }}
                </p>
              </article>
            </div>
          </article>
        </section>
      </template>
    </section>
  </main>
</template>
