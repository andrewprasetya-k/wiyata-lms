<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import {
  PhArrowRight,
  PhBookOpen,
  PhClock,
  PhFileText,
  PhNotebook,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { getStudentNotes } from "../../services/studentNotes";
import type { StudentGlobalMaterialNote } from "../../types/studentNotes";
import { formatDateTime, parseBackendTimestamp } from "../../utils/date";
import { getSubjectColor } from "../../utils/color";

interface NoteGroup {
  key: string;
  className: string;
  classCode: string;
  subjectName: string;
  subjectCode: string;
  latestUpdatedAt: string;
  notes: StudentGlobalMaterialNote[];
}

const notes = ref<StudentGlobalMaterialNote[]>([]);
const selectedNoteId = ref("");
const isLoading = ref(true);
const errorMessage = ref("");

const sortedNotes = computed(() =>
  [...notes.value].sort((a, b) => getTime(b.updatedAt) - getTime(a.updatedAt)),
);

const selectedNote = computed(
  () =>
    notes.value.find((note) => note.noteId === selectedNoteId.value) ??
    sortedNotes.value[0] ??
    null,
);

const groupedNotes = computed<NoteGroup[]>(() => {
  const groups = new Map<string, NoteGroup>();

  for (const note of sortedNotes.value) {
    const key = `${note.classId}:${note.subjectId}`;
    const existing = groups.get(key);

    if (existing) {
      existing.notes.push(note);
      continue;
    }

    groups.set(key, {
      key,
      className: note.className,
      classCode: note.classCode,
      subjectName: note.subjectName,
      subjectCode: note.subjectCode,
      latestUpdatedAt: note.updatedAt,
      notes: [note],
    });
  }

  return [...groups.values()].sort(
    (a, b) => getTime(b.latestUpdatedAt) - getTime(a.latestUpdatedAt),
  );
});

async function loadNotes() {
  isLoading.value = true;
  errorMessage.value = "";

  try {
    const response = await getStudentNotes();
    notes.value = response.notes;

    const selectedStillExists = response.notes.some(
      (note) => note.noteId === selectedNoteId.value,
    );
    if (!selectedStillExists) {
      selectedNoteId.value =
        [...response.notes].sort(
          (a, b) => getTime(b.updatedAt) - getTime(a.updatedAt),
        )[0]?.noteId ?? "";
    }
  } catch {
    errorMessage.value =
      "Catatan belum bisa dimuat. Periksa koneksi atau coba lagi nanti.";
  } finally {
    isLoading.value = false;
  }
}

function selectNote(noteId: string) {
  selectedNoteId.value = noteId;
}

function getTime(value?: string | null) {
  if (!value) return 0;
  const time = parseBackendTimestamp(value)?.getTime() ?? Number.NaN;
  return Number.isNaN(time) ? 0 : time;
}

function groupLabel(group: NoteGroup) {
  const subject = group.subjectCode
    ? `${group.subjectName} · ${group.subjectCode}`
    : group.subjectName;
  const className = group.classCode
    ? `${group.className} · ${group.classCode}`
    : group.className;
  return `${subject || "Mata pelajaran"} — ${className || "Kelas"}`;
}

onMounted(loadNotes);
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 bg-background">
    <header class="border-b border-border bg-surface px-5 py-4 sm:px-6 lg:px-8">
      <div
        class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between"
      >
        <div>
          <h1 class="mt-1 text-2xl font-medium text-foreground">
            Catatan Saya
          </h1>
          <p class="mt-1 text-sm leading-6 text-muted">
            Baca kembali catatan materi yang masih tersedia di kelas aktifmu.
          </p>
        </div>
        <p v-if="!isLoading && !errorMessage" class="text-xs text-muted">
          {{ notes.length }} catatan tersimpan
        </p>
      </div>
    </header>

    <section
      v-if="isLoading"
      class="grid min-h-[calc(100vh-116px)] lg:grid-cols-[300px_minmax(0,1fr)]"
    >
      <div class="border-r border-border bg-surface p-4">
        <div class="h-5 w-28 animate-pulse rounded bg-[#f1efeb]" />
        <div class="mt-5 space-y-3">
          <div
            v-for="item in 5"
            :key="item"
            class="h-24 animate-pulse rounded-xl bg-background"
          />
        </div>
      </div>
      <div class="p-5 sm:p-6 lg:p-8">
        <div
          class="mx-auto h-full min-h-120 max-w-4xl animate-pulse rounded-[22px] border border-border bg-surface"
        />
      </div>
    </section>

    <section
      v-else-if="errorMessage"
      class="flex min-h-[calc(100vh-116px)] items-center justify-center px-5 py-10"
    >
      <article
        class="w-full max-w-xl rounded-[22px] border border-danger-line bg-danger-soft p-6"
      >
        <div class="flex items-start gap-3">
          <PhWarningCircle
            :size="24"
            class="mt-0.5 shrink-0 text-danger"
            weight="duotone"
          />
          <div>
            <h2 class="text-base font-medium text-foreground">
              Catatan tidak dapat dimuat
            </h2>
            <p class="mt-1 text-sm leading-6 text-muted">
              {{ errorMessage }}
            </p>
            <button
              class="mt-4 rounded-lg bg-brand px-4 py-2 text-sm font-medium text-white transition hover:bg-brand-hover"
              type="button"
              @click="loadNotes"
            >
              Coba lagi
            </button>
          </div>
        </div>
      </article>
    </section>

    <section
      v-else-if="notes.length === 0"
      class="flex min-h-[calc(100vh-116px)] items-center justify-center px-5 py-10"
    >
      <article
        class="w-full max-w-xl rounded-[22px] border border-border bg-surface p-8 text-center"
      >
        <div
          class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-brand-soft text-brand"
        >
          <PhNotebook :size="24" weight="duotone" />
        </div>
        <h2 class="mt-4 text-base font-medium text-foreground">
          Belum ada catatan
        </h2>
        <p class="mx-auto mt-2 max-w-md text-sm leading-6 text-muted">
          Belum ada catatan. Buka materi dan tulis catatan untuk mulai membangun
          ruang belajarmu.
        </p>
        <RouterLink
          class="mt-5 inline-flex items-center gap-2 rounded-lg bg-brand px-4 py-2 text-sm font-medium text-white transition hover:bg-brand-hover"
          to="/student/subjects"
        >
          Buka mata pelajaran
          <PhArrowRight :size="16" />
        </RouterLink>
      </article>
    </section>

    <section
      v-else
      class="grid min-h-[calc(100vh-116px)] lg:h-[calc(100vh-116px)] lg:grid-cols-[300px_minmax(0,1fr)] lg:overflow-hidden"
    >
      <aside
        class="min-w-0 border-b border-border bg-surface lg:border-b-0 lg:border-r"
      >
        <div class="flex items-center justify-between px-4 py-3">
          <p class="text-sm font-medium text-foreground">Semua catatan</p>
          <span class="text-xs text-muted">{{ notes.length }}</span>
        </div>

        <div
          class="max-h-[46vh] overflow-y-auto p-3 lg:h-[calc(100%-49px)] lg:max-h-none"
        >
          <section
            v-for="group in groupedNotes"
            :key="group.key"
            class="mb-4 last:mb-0"
          >
            <p
              class="px-2 pb-2 text-[10px] font-medium uppercase tracking-wide text-muted"
              :title="groupLabel(group)"
            >
              {{ group.subjectName || "Mata pelajaran" }}
              <span v-if="group.className"> · {{ group.className }}</span>
            </p>

            <div class="space-y-1">
              <button
                v-for="note in group.notes"
                :key="note.noteId"
                class="w-full rounded-xl p-3 text-left transition"
                :class="
                  selectedNote?.noteId === note.noteId
                    ? 'bg-brand-soft'
                    : 'hover:bg-background'
                "
                type="button"
                @click="selectNote(note.noteId)"
              >
                <div class="flex items-center gap-2">
                  <span
                    class="h-2 w-2 shrink-0 rounded-full"
                    :style="{
                      backgroundColor: getSubjectColor(
                        note.subjectName || note.subjectCode,
                      ),
                    }"
                  />
                  <span class="truncate text-xs text-muted">
                    {{ note.subjectName || "Mata pelajaran" }}
                  </span>
                </div>
                <p class="mt-2 truncate text-sm font-medium text-foreground">
                  {{ note.materialTitle }}
                </p>
                <p class="mt-1 truncate whitespace-pre-line text-xs text-muted">
                  {{ note.content }}
                </p>
                <p class="mt-2 text-[10px] text-[#b0aab7]">
                  {{ formatDateTime(note.updatedAt) }}
                </p>
              </button>
            </div>
          </section>
        </div>
      </aside>

      <article
        v-if="selectedNote"
        class="min-w-0 bg-surface-subtle p-4 sm:p-6 lg:overflow-y-auto lg:p-8"
      >
        <div class="mx-auto flex min-h-full w-full max-w-4xl flex-col">
          <header
            class="flex flex-col gap-4 border-b border-border pb-5 sm:flex-row sm:items-start sm:justify-between"
          >
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2 text-xs text-muted">
                <span>{{ selectedNote.subjectName || "Mata pelajaran" }}</span>
                <span class="text-[#d1ccd5]">/</span>
                <span>{{ selectedNote.className || "Kelas" }}</span>
              </div>
              <h2
                class="mt-3 wrap-break-word text-2xl font-medium text-foreground"
              >
                {{ selectedNote.materialTitle }}
              </h2>
              <p class="mt-2 inline-flex items-center gap-2 text-xs text-muted">
                <PhClock :size="14" />
                Diperbarui {{ formatDateTime(selectedNote.updatedAt) }}
              </p>
            </div>

            <RouterLink
              class="inline-flex shrink-0 items-center justify-center gap-2 rounded-xl bg-brand px-4 py-2.5 text-sm font-medium text-white transition hover:bg-brand-hover"
              :to="`/student/subjects/${selectedNote.subjectClassId}/materials/${selectedNote.materialId}/note`"
            >
              Buka catatan
              <PhArrowRight :size="16" />
            </RouterLink>
          </header>

          <RouterLink
            class="mt-5 flex items-center gap-3 rounded-xl border border-[#dfe3ff] bg-brand-soft p-4 transition hover:border-[#aeb8ff]"
            :to="`/student/subjects/${selectedNote.subjectClassId}/materials/${selectedNote.materialId}`"
          >
            <span
              class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-surface text-brand"
            >
              <PhFileText :size="20" weight="duotone" />
            </span>
            <span class="min-w-0 flex-1">
              <span class="block text-xs font-medium text-brand">
                Materi terkait
              </span>
              <span class="mt-1 block truncate text-sm text-foreground">
                {{ selectedNote.materialTitle }}
                <template v-if="selectedNote.materialType">
                  · {{ selectedNote.materialType.toUpperCase() }}
                </template>
              </span>
            </span>
            <PhArrowRight :size="16" class="shrink-0 text-brand" />
          </RouterLink>

          <div
            class="mt-5 min-h-64 flex-1 rounded-[18px] border border-border bg-surface p-5 sm:p-6"
          >
            <p
              class="whitespace-pre-wrap wrap-break-word text-sm leading-7 text-foreground"
            >
              {{ selectedNote.content }}
            </p>
          </div>

          <footer
            class="mt-4 flex flex-col gap-3 text-xs text-muted sm:flex-row sm:items-center sm:justify-between"
          >
            <span>
              {{ selectedNote.subjectCode || selectedNote.subjectName }}
              <template v-if="selectedNote.classCode">
                · {{ selectedNote.classCode }}
              </template>
            </span>
            <RouterLink
              class="inline-flex items-center gap-2 font-medium text-brand transition hover:text-brand-hover"
              :to="`/student/subjects/${selectedNote.subjectClassId}/materials/${selectedNote.materialId}`"
            >
              <PhBookOpen :size="15" />
              Lihat materi
            </RouterLink>
          </footer>
        </div>
      </article>
    </section>
  </main>
</template>
