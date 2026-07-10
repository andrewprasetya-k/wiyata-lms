<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink, useRoute } from "vue-router";
import {
  PhArrowLeft,
  PhCalendarBlank,
  PhClipboardText,
  PhPaperclip,
  PhPencilSimple,
  PhUsersThree,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import AttachmentPreviewList from "../../components/common/AttachmentPreviewList.vue";
import CommentThread from "../../components/comments/CommentThread.vue";
import { getSubjectAssignmentDetail } from "../../services/assignment";
import type {
  AssignmentItem,
  SubjectClassHeader,
} from "../../types/assignment";
import { formatDateTime } from "../../utils/date";

const route = useRoute();
const subjectClassId = computed(() =>
  String(route.params.subjectClassId ?? ""),
);
const assignmentId = computed(() => String(route.params.assignmentId ?? ""));
const assignment = ref<AssignmentItem | null>(null);
const subjectClass = ref<SubjectClassHeader | null>(null);
const isLoading = ref(true);
const errorMessage = ref("");
const didLoad = ref(false);

async function loadAssignment() {
  if (!subjectClassId.value || !assignmentId.value) {
    isLoading.value = false;
    errorMessage.value = "Konteks tugas tidak lengkap.";
    return;
  }

  isLoading.value = true;
  errorMessage.value = "";
  didLoad.value = false;

  try {
    const response = await getSubjectAssignmentDetail(
      subjectClassId.value,
      assignmentId.value,
    );
    subjectClass.value = response.subjectClass;
    assignment.value = response.assignment;
    didLoad.value = true;
  } catch {
    errorMessage.value =
      "Detail tugas belum bisa dimuat. Periksa koneksi atau coba lagi nanti.";
  } finally {
    isLoading.value = false;
  }
}

onMounted(loadAssignment);
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-[#f8f7f4]">
    <header class="border-b border-border bg-white">
      <div
        class="flex min-w-0 items-center gap-2 px-5 py-3 text-xs text-muted sm:px-6 lg:px-8"
      >
        <RouterLink
          :to="`/teacher/subjects/${subjectClassId}`"
          class="inline-flex shrink-0 items-center gap-1.5 transition hover:text-brand"
        >
          <PhArrowLeft :size="15" />
          Ruang mengajar
        </RouterLink>
        <span class="text-[#d1d5db]">/</span>
        <span class="shrink-0">Tugas</span>
        <span class="text-[#d1d5db]">/</span>
        <span class="min-w-0 truncate font-medium text-foreground">
          {{
            assignment?.assignmentTitle ??
            (isLoading ? "Memuat..." : "Detail tugas")
          }}
        </span>
      </div>

      <div
        v-if="assignment"
        class="flex min-w-0 flex-col gap-4 border-t border-[#f3f1ec] px-5 py-4 sm:px-6 lg:flex-row lg:items-center lg:justify-between lg:px-8"
      >
        <div class="flex min-w-0 items-start gap-3">
          <div
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-[#eef2ff] text-brand"
          >
            <PhClipboardText :size="21" weight="duotone" />
          </div>
          <div class="min-w-0">
            <h1
              class="wrap-break-word text-xl font-semibold text-foreground sm:text-2xl"
            >
              {{ assignment.assignmentTitle }}
            </h1>
            <p class="mt-1 text-sm text-muted">
              {{
                assignment.subjectName ||
                subjectClass?.subjectName ||
                "Tugas mata pelajaran"
              }}
            </p>
          </div>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <RouterLink
            :to="{
              name: 'teacher-assignment-edit',
              params: {
                subjectClassId,
                asgId: assignment.assignmentId,
              },
            }"
            class="inline-flex items-center gap-2 rounded-lg border border-border bg-white px-3 py-2 text-xs font-medium text-brand transition hover:border-brand hover:bg-[#eef2ff]"
          >
            <PhPencilSimple :size="15" weight="bold" />
            Edit tugas
          </RouterLink>
          <RouterLink
            :to="{
              name: 'teacher-assignment-review',
              params: { assignmentId: assignment.assignmentId },
            }"
            class="inline-flex items-center gap-2 rounded-lg bg-brand px-3 py-2 text-xs font-medium text-white transition hover:bg-[#4338ca]"
          >
            <PhUsersThree :size="15" weight="duotone" />
            Nilai pengumpulan
          </RouterLink>
        </div>
      </div>
    </header>

    <section class="px-5 py-5 sm:px-6 lg:px-8">
      <section
        v-if="isLoading"
        class="grid gap-5 lg:grid-cols-[minmax(0,1fr)_300px]"
      >
        <div
          class="h-80 animate-pulse rounded-xl border border-border bg-white"
        />
        <div
          class="h-64 animate-pulse rounded-xl border border-border bg-white"
        />
      </section>

      <section
        v-else-if="errorMessage"
        class="mx-auto max-w-xl rounded-xl border border-[#fecaca] bg-[#fef2f2] px-5 py-8 text-center"
      >
        <PhWarningCircle
          :size="30"
          class="mx-auto text-[#d97757]"
          weight="duotone"
        />
        <h2 class="mt-3 text-lg font-semibold text-foreground">
          Tugas belum bisa dimuat
        </h2>
        <p class="mt-2 text-sm leading-6 text-[#6b6475]">
          {{ errorMessage }}
        </p>
        <button
          class="mt-5 rounded-lg bg-foreground px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#2f2b3a]"
          type="button"
          @click="loadAssignment"
        >
          Coba lagi
        </button>
      </section>

      <section
        v-else-if="didLoad && !assignment"
        class="mx-auto max-w-xl rounded-xl border border-border bg-white px-5 py-10 text-center"
      >
        <div
          class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-[#eef2ff] text-brand"
        >
          <PhClipboardText :size="24" weight="duotone" />
        </div>
        <h2 class="mt-3 text-lg font-semibold text-foreground">
          Tugas tidak ditemukan
        </h2>
        <p class="mt-2 text-sm leading-6 text-[#6b6475]">
          Tugas ini tidak tersedia atau bukan bagian dari mata pelajaran yang
          Anda ajar.
        </p>
        <RouterLink
          :to="`/teacher/subjects/${subjectClassId}`"
          class="mt-5 inline-flex items-center gap-2 rounded-lg border border-border bg-white px-4 py-2.5 text-sm font-medium text-brand transition hover:border-brand hover:bg-[#eef2ff]"
        >
          <PhArrowLeft :size="16" />
          Kembali ke ruang mengajar
        </RouterLink>
      </section>

      <section
        v-else-if="assignment"
        class="grid min-w-0 gap-5 lg:grid-cols-[minmax(0,1fr)_300px]"
      >
        <div class="min-w-0 space-y-5">
          <article class="rounded-xl border border-border bg-white shadow-sm p-5 sm:p-6">
            <div class="flex flex-wrap items-center gap-2">
              <span
                v-if="assignment.categoryName"
                class="rounded-full bg-[#eef2ff] px-2.5 py-1 text-[11px] font-medium text-brand"
              >
                {{ assignment.categoryName }}
              </span>
              <span
                v-if="
                  assignment.subjectName ||
                  assignment.subjectCode ||
                  subjectClass?.subjectName ||
                  subjectClass?.subjectCode
                "
                class="rounded-full bg-[#f8f7f4] px-2.5 py-1 text-[11px] text-muted"
              >
                {{
                  assignment.subjectName ||
                  assignment.subjectCode ||
                  subjectClass?.subjectName ||
                  subjectClass?.subjectCode
                }}
              </span>
            </div>

            <div class="mt-5 border-t border-[#f0ede8] pt-5">
              <h2 class="text-sm font-semibold text-foreground">
                Instruksi tugas
              </h2>
              <p
                v-if="assignment.assignmentDescription"
                class="mt-3 whitespace-pre-line wrap-break-word text-sm leading-7 text-[#4a4356]"
              >
                {{ assignment.assignmentDescription }}
              </p>
              <p v-else class="mt-3 text-sm leading-6 text-[#7a7385]">
                Instruksi tugas belum tersedia.
              </p>
            </div>
          </article>

          <article class="rounded-xl border border-border bg-white shadow-sm p-5 sm:p-6">
            <div class="flex items-center gap-2">
              <PhPaperclip :size="18" class="text-brand" />
              <h2 class="text-sm font-semibold text-foreground">
                Lampiran tugas
              </h2>
            </div>
            <p class="mt-1 text-xs leading-5 text-[#7a7385]">
              File pendukung yang bisa dibuka oleh peserta kelas.
            </p>
            <AttachmentPreviewList
              class="mt-4"
              :attachments="assignment.attachments"
              empty-text="Tugas ini tidak memiliki lampiran."
            />
          </article>

          <CommentThread
            source-type="assignment"
            :source-id="assignment.assignmentId"
            title="Diskusi tugas"
            placeholder="Tulis tanggapan untuk diskusi tugas..."
            empty-text="Belum ada diskusi untuk tugas ini."
          />
        </div>

        <aside class="min-w-0 lg:sticky lg:top-6">
          <article class="rounded-xl border border-border bg-white shadow-sm p-5">
            <p class="text-sm font-semibold text-foreground">Ringkasan tugas</p>
            <dl
              class="mt-4 divide-y divide-[#f0ede8] rounded-lg bg-[#fbfaf8] px-3"
            >
              <div class="flex items-start justify-between gap-4 py-3">
                <dt class="inline-flex items-center gap-1.5 text-xs text-[#7a7385]">
                  <PhCalendarBlank :size="14" />
                  Tenggat
                </dt>
                <dd class="text-right text-xs font-medium text-foreground">
                  {{
                    assignment.deadline
                      ? formatDateTime(assignment.deadline)
                      : "Tidak ada tenggat"
                  }}
                </dd>
              </div>
              <div class="flex items-start justify-between gap-4 py-3">
                <dt class="text-xs text-[#7a7385]">Pengumpulan terlambat</dt>
                <dd class="text-right text-xs font-medium text-foreground">
                  {{
                    assignment.allowLateSubmission
                      ? "Diizinkan"
                      : "Tidak diizinkan"
                  }}
                </dd>
              </div>
              <div
                v-if="assignment.createdAt"
                class="flex items-start justify-between gap-4 py-3"
              >
                <dt class="text-xs text-[#7a7385]">Dibuat</dt>
                <dd class="text-right text-xs font-medium text-foreground">
                  {{ formatDateTime(assignment.createdAt) }}
                </dd>
              </div>
            </dl>
          </article>
        </aside>
      </section>
    </section>
  </main>
</template>
