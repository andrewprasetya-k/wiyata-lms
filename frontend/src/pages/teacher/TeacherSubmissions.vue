<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import {
  PhArrowRight,
  PhCheckCircle,
  PhClipboardText,
  PhClock,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { getTeacherSubmissionInbox } from "../../services/teacherAssignment";
import type {
  TeacherSubmissionInboxItem,
  TeacherSubmissionInboxSummary,
} from "../../types/teacherAssignment";
import { formatDate } from "../../utils/date";

type InboxFilter = "all" | "pending" | "graded";

const loading = ref(false);
const errorMessage = ref("");
const inboxItems = ref<TeacherSubmissionInboxItem[]>([]);
const inboxSummary = ref<TeacherSubmissionInboxSummary>({
  totalSubmissions: 0,
  pendingCount: 0,
  gradedCount: 0,
  lateCount: 0,
});
const activeFilter = ref<InboxFilter>("all");

const summary = computed(() => ({
  submissions: inboxSummary.value.totalSubmissions,
  pending: inboxSummary.value.pendingCount,
  graded: inboxSummary.value.gradedCount,
  late: inboxSummary.value.lateCount,
}));

const filterTabs = computed(() => [
  { id: "all" as const, label: "Semua", count: inboxItems.value.length },
  {
    id: "pending" as const,
    label: "Perlu review",
    count: inboxItems.value.filter((item) => item.pendingCount > 0).length,
  },
  {
    id: "graded" as const,
    label: "Sudah dinilai",
    count: inboxItems.value.filter(
      (item) => item.submissionCount > 0 && item.pendingCount === 0,
    ).length,
  },
]);

const filteredItems = computed(() => {
  const items = inboxItems.value.filter((item) => {
    if (activeFilter.value === "pending") return item.pendingCount > 0;
    if (activeFilter.value === "graded") {
      return item.submissionCount > 0 && item.pendingCount === 0;
    }
    return true;
  });

  return [...items].sort(compareInboxItems);
});

function compareInboxItems(
  a: TeacherSubmissionInboxItem,
  b: TeacherSubmissionInboxItem,
) {
  const pendingDiff = Number(b.pendingCount > 0) - Number(a.pendingCount > 0);
  if (pendingDiff !== 0) return pendingDiff;

  const aDeadline = getDeadlineTime(a.deadline);
  const bDeadline = getDeadlineTime(b.deadline);
  if (aDeadline !== bDeadline) return aDeadline - bDeadline;

  return (a.assignmentTitle || "").localeCompare(b.assignmentTitle || "");
}

function getDeadlineTime(deadline?: string | null) {
  if (!deadline) return Number.MAX_SAFE_INTEGER;
  const value = new Date(deadline).getTime();
  return Number.isNaN(value) ? Number.MAX_SAFE_INTEGER : value;
}

async function loadInbox() {
  loading.value = true;
  errorMessage.value = "";
  inboxItems.value = [];
  inboxSummary.value = {
    totalSubmissions: 0,
    pendingCount: 0,
    gradedCount: 0,
    lateCount: 0,
  };

  try {
    const response = await getTeacherSubmissionInbox();
    inboxItems.value = response.items ?? [];
    inboxSummary.value = response.summary ?? inboxSummary.value;
  } catch {
    errorMessage.value =
      "Inbox pengumpulan belum bisa dimuat. Coba lagi beberapa saat.";
  } finally {
    loading.value = false;
  }
}

onMounted(loadInbox);
</script>

<template>
  <main class="min-h-screen flex-1 px-5 py-5 sm:px-6 lg:px-8">
    <section class="flex w-full max-w-none flex-col gap-5">
      <header
        class="rounded-[22px] bg-[#f0e9dd] px-5 py-5 shadow-sm ring-1 ring-black/5 md:px-6"
      >
        <p class="text-sm font-medium text-[#8a6d3b]">Pengumpulan siswa</p>
        <h1 class="mt-3 text-3xl font-medium text-[#171322] md:text-4xl">
          Inbox pengumpulan
        </h1>
        <p class="mt-3 max-w-2xl text-sm leading-6 text-[#6b6475]">
          Pantau pengumpulan dari semua subject yang diajar. Proses nilai dan
          feedback tetap dilakukan dari halaman review tugas.
        </p>
      </header>

      <section v-if="loading" class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5">
        <p class="text-sm text-[#6b6475]">Memuat inbox pengumpulan...</p>
      </section>

      <section
        v-else-if="errorMessage"
        class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5"
      >
        <div
          class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"
        >
          <div class="flex items-start gap-3">
            <PhWarningCircle
              :size="24"
              class="mt-0.5 text-[#e58f86]"
              weight="duotone"
            />
            <div>
              <h2 class="text-lg font-medium text-[#171322]">
                Gagal memuat inbox
              </h2>
              <p class="mt-2 text-sm leading-6 text-[#6b6475]">
                {{ errorMessage }}
              </p>
            </div>
          </div>
          <button
            type="button"
            class="rounded-2xl bg-[#171322] px-4 py-3 text-sm font-medium text-white"
            @click="loadInbox"
          >
            Coba lagi
          </button>
        </div>
      </section>

      <template v-else>
        <section class="grid gap-4 md:grid-cols-4">
          <article class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5">
            <PhClipboardText :size="24" class="text-[#7aa7d9]" weight="duotone" />
            <p class="mt-4 text-sm text-[#8a8494]">Total submission</p>
            <p class="mt-1 text-2xl font-medium text-[#171322]">
              {{ summary.submissions }}
            </p>
          </article>
          <article class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5">
            <PhWarningCircle :size="24" class="text-[#e58f86]" weight="duotone" />
            <p class="mt-4 text-sm text-[#8a8494]">Perlu review</p>
            <p class="mt-1 text-2xl font-medium text-[#171322]">
              {{ summary.pending }}
            </p>
          </article>
          <article class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5">
            <PhCheckCircle :size="24" class="text-[#74bfa5]" weight="duotone" />
            <p class="mt-4 text-sm text-[#8a8494]">Sudah dinilai</p>
            <p class="mt-1 text-2xl font-medium text-[#171322]">
              {{ summary.graded }}
            </p>
          </article>
          <article class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5">
            <PhClock :size="24" class="text-[#b889c9]" weight="duotone" />
            <p class="mt-4 text-sm text-[#8a8494]">Terlambat</p>
            <p class="mt-1 text-2xl font-medium text-[#171322]">
              {{ summary.late }}
            </p>
          </article>
        </section>

        <section class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5">
          <div class="flex flex-col gap-4 border-b border-[#ece8df] pb-4 lg:flex-row lg:items-end lg:justify-between">
            <div>
              <p class="text-sm font-medium text-[#171322]">
                Daftar assignment dengan pengumpulan
              </p>
              <p class="mt-1 text-sm text-[#8a8494]">
                {{ inboxItems.length }} assignment memiliki pengumpulan dalam school aktif.
              </p>
            </div>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="tab in filterTabs"
                :key="tab.id"
                type="button"
                class="rounded-2xl px-4 py-2.5 text-sm font-medium transition"
                :class="
                  activeFilter === tab.id
                    ? 'bg-[#171322] text-white'
                    : 'bg-[#faf8f4] text-[#6b6475] hover:bg-[#f0e9dd] hover:text-[#171322]'
                "
                @click="activeFilter = tab.id"
              >
                {{ tab.label }}
                <span class="ml-2 opacity-70">{{ tab.count }}</span>
              </button>
            </div>
          </div>

          <div v-if="inboxItems.length === 0" class="py-10 text-center">
            <PhClipboardText
              :size="34"
              class="mx-auto text-[#b5afbf]"
              weight="duotone"
            />
            <h2 class="mt-3 text-lg font-medium text-[#171322]">
              Belum ada pengumpulan
            </h2>
            <p class="mx-auto mt-2 max-w-xl text-sm leading-6 text-[#6b6475]">
              Pengumpulan akan tampil setelah siswa mengumpulkan tugas pada
              subject yang kamu ajar.
            </p>
          </div>

          <div v-else-if="filteredItems.length === 0" class="py-10 text-center">
            <PhCheckCircle
              :size="34"
              class="mx-auto text-[#b5afbf]"
              weight="duotone"
            />
            <h2 class="mt-3 text-lg font-medium text-[#171322]">
              Belum ada pengumpulan
            </h2>
            <p class="mx-auto mt-2 max-w-xl text-sm leading-6 text-[#6b6475]">
              Tidak ada assignment yang sesuai dengan filter saat ini.
            </p>
          </div>

          <div v-else class="space-y-3 pt-5">
            <article
              v-for="item in filteredItems"
              :key="`${item.subjectClassId}-${item.assignmentId}`"
              class="rounded-[18px] bg-[#faf8f4] p-5 ring-1 ring-black/5"
            >
              <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                <div class="min-w-0">
                  <div class="flex flex-wrap gap-2 text-xs font-medium">
                    <span class="rounded-2xl bg-white px-3 py-1.5 text-[#4f46e5]">
                      {{ item.subjectName }}
                    </span>
                    <span
                      v-if="item.subjectCode"
                      class="rounded-2xl bg-white px-3 py-1.5 text-[#6b6475]"
                    >
                      {{ item.subjectCode }}
                    </span>
                    <span class="rounded-2xl bg-white px-3 py-1.5 text-[#6b6475]">
                      {{ item.className || item.classCode || "Kelas" }}
                    </span>
                  </div>

                  <h2 class="mt-4 text-lg font-medium text-[#171322]">
                    {{ item.assignmentTitle }}
                  </h2>
                  <p
                    v-if="item.deadline"
                    class="mt-2 text-sm text-[#6b6475]"
                  >
                    <span>
                      Deadline {{ formatDate(item.deadline) }}
                    </span>
                  </p>
                </div>

                <RouterLink
                  :to="{
                    name: 'teacher-assignment-review',
                    params: { assignmentId: item.assignmentId },
                  }"
                  class="inline-flex items-center justify-center gap-2 rounded-2xl bg-[#171322] px-4 py-3 text-sm font-medium text-white transition hover:bg-[#2f2b3a]"
                >
                  Review pengumpulan
                  <PhArrowRight :size="16" />
                </RouterLink>
              </div>

              <div class="mt-5 grid gap-3 sm:grid-cols-4">
                <div class="rounded-2xl bg-white p-4">
                  <p class="text-xs text-[#8a8494]">Submission</p>
                  <p class="mt-1 text-xl font-medium text-[#171322]">
                    {{ item.submissionCount }}
                  </p>
                </div>
                <div class="rounded-2xl bg-white p-4">
                  <p class="text-xs text-[#8a8494]">Perlu review</p>
                  <p class="mt-1 text-xl font-medium text-[#171322]">
                    {{ item.pendingCount }}
                  </p>
                </div>
                <div class="rounded-2xl bg-white p-4">
                  <p class="text-xs text-[#8a8494]">Sudah dinilai</p>
                  <p class="mt-1 text-xl font-medium text-[#171322]">
                    {{ item.gradedCount }}
                  </p>
                </div>
                <div class="rounded-2xl bg-white p-4">
                  <p class="text-xs text-[#8a8494]">Terlambat</p>
                  <p class="mt-1 text-xl font-medium text-[#171322]">
                    {{ item.lateCount }}
                  </p>
                </div>
              </div>
            </article>
          </div>
        </section>
      </template>
    </section>
  </main>
</template>
