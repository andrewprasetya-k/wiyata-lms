<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  PhArrowLeft,
  PhBookOpen,
  PhCalendarBlank,
  PhClipboardText,
  PhFileText,
  PhNotebook,
  PhArrowRight,
  PhPaperclip,
  PhUserCircle,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { getSubjectMaterials } from "../../services/classWorkspace";
import { getSubjectAssignments } from "../../services/assignment";
import { getStudentSubjectClassNotes } from "../../services/studentNotes";
import type { MaterialItem } from "../../types/classWorkspace";
import type { AssignmentItem } from "../../types/assignment";
import type { StudentSubjectMaterialNote } from "../../types/studentNotes";
import { formatDateTime } from "../../utils/date";

const route = useRoute();
const router = useRouter();
const subjectClassId = computed(() => String(route.params.sclId ?? ""));
const activeTab = ref("materials");
const materials = ref<MaterialItem[]>([]);
const assignments = ref<AssignmentItem[]>([]);
const notes = ref<StudentSubjectMaterialNote[]>([]);
const notesLoaded = ref(false);
const notesLoading = ref(false);
const notesError = ref("");
const subjectTitle = ref(String(route.query.title ?? "Detail Mata Pelajaran"));
const teacherName = ref("");
const isLoading = ref(true);
const errorMessage = ref("");

const tabs = [
  {
    key: "materials",
    label: "Materi",
    icon: PhBookOpen,
  },
  {
    key: "assignments",
    label: "Tugas",
    icon: PhClipboardText,
  },
  {
    key: "notes",
    label: "Catatan",
    icon: PhNotebook,
  },
];

const currentTab = computed(
  () => tabs.find((tab) => tab.key === activeTab.value) ?? tabs[0],
);

async function loadSubject() {
  if (!subjectClassId.value) {
    isLoading.value = false;
    errorMessage.value = "Subject class ID tidak tersedia.";
    return;
  }

  if (activeTab.value === "notes") {
    await loadNotes();
    return;
  }

  isLoading.value = true;
  errorMessage.value = "";
  try {
    if (activeTab.value === "materials") {
      const data = await getSubjectMaterials(subjectClassId.value);
      materials.value = data.materials;
      subjectTitle.value =
        data.subjectClass.subjectName ||
        data.subjectClass.subjectCode ||
        subjectTitle.value;
      teacherName.value = data.subjectClass.teacherName || "";
    } else if (activeTab.value === "assignments") {
      const data = await getSubjectAssignments(subjectClassId.value);
      assignments.value = data.data.data;
      subjectTitle.value =
        data.subjectClass.subjectName ||
        data.subjectClass.subjectCode ||
        subjectTitle.value;
      teacherName.value = data.subjectClass.teacherName || "";
    }
  } catch {
    errorMessage.value =
      "Detail mata pelajaran belum bisa dimuat. Periksa koneksi atau coba lagi nanti.";
  } finally {
    isLoading.value = false;
  }
}

async function loadNotes(force = false) {
  if (!subjectClassId.value || (notesLoaded.value && !force)) {
    return;
  }

  notesLoading.value = true;
  notesError.value = "";

  try {
    const data = await getStudentSubjectClassNotes(subjectClassId.value);
    notes.value = data.notes;
    notesLoaded.value = true;
  } catch {
    notesError.value =
      "Catatan belum bisa dimuat. Periksa koneksi atau coba lagi nanti.";
  } finally {
    notesLoading.value = false;
  }
}

watch(activeTab, () => {
  loadSubject();
});

onMounted(loadSubject);
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 bg-[#f8f7f4]">
    <header class="border-b border-border bg-white">
      <div
        class="flex min-w-0 items-center gap-2 px-5 py-5 text-xs text-muted sm:px-6 lg:px-8"
      >
        <button
          class="inline-flex shrink-0 items-center gap-1.5 transition hover:text-brand"
          type="button"
          @click="router.push('/student/subjects')"
        >
          <PhArrowLeft :size="15" />
          Mata pelajaran
        </button>
        <span class="text-[#d1d5db]">/</span>
        <span class="min-w-0 truncate font-medium text-foreground">
          {{ subjectTitle }}
        </span>
      </div>

      <div class="px-5 pt-2 sm:px-6 lg:px-8">
        <div class="flex min-w-0 items-start gap-3 pb-4">
          <span
            class="mt-1.5 h-3 w-3 shrink-0 rounded-full bg-[#4f8ef7]"
            aria-hidden="true"
          />
          <div class="min-w-0">
            <h1 class="truncate text-xl font-semibold text-foreground sm:text-2xl">
              {{ subjectTitle }}
            </h1>
            <p
              class="mt-1 flex items-center gap-1.5 text-xs text-muted sm:text-sm"
            >
              <PhUserCircle :size="16" class="shrink-0" />
              <span class="truncate">
                {{ teacherName || "Guru belum tersedia" }}
              </span>
            </p>
          </div>
        </div>

        <nav
          class="-mb-px flex min-w-0 gap-1 overflow-x-auto"
          aria-label="Konten mata pelajaran"
        >
          <button
            v-for="tab in tabs"
            :key="tab.key"
            class="inline-flex h-11 shrink-0 items-center gap-2 border-b-2 px-3 text-sm transition sm:px-4"
            :class="
              activeTab === tab.key
                ? 'border-brand font-medium text-brand'
                : 'border-transparent text-muted hover:border-[#d8d5df] hover:text-[#374151]'
            "
            type="button"
            @click="activeTab = tab.key"
          >
            <component :is="tab.icon" :size="17" />
            {{ tab.label }}
          </button>
        </nav>
      </div>
    </header>

    <section class="px-5 py-5 sm:px-6 lg:px-8 lg:py-6">
      <div class="mx-auto max-w-7xl">
        <div class="mb-4 flex items-center gap-3">
          <div
            class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-[#eef2ff] text-brand"
          >
            <component :is="currentTab.icon" :size="19" weight="duotone" />
          </div>
          <div>
            <h2 class="text-base font-semibold text-foreground">
              {{
                activeTab === "materials"
                  ? "Materi pembelajaran"
                  : activeTab === "assignments"
                    ? "Tugas mata pelajaran"
                    : "Catatan materi saya"
              }}
            </h2>
            <p class="mt-0.5 text-xs text-[#7a7385]">
              {{
                activeTab === "materials"
                  ? "Pelajari materi dan buka lampiran yang dibagikan guru."
                  : activeTab === "assignments"
                    ? "Lihat instruksi, tenggat, dan pengumpulan dari detail tugas."
                    : "Buka kembali catatan pribadi yang tersimpan dari setiap materi."
              }}
            </p>
          </div>
        </div>

        <template v-if="activeTab === 'materials'">
          <div
            v-if="isLoading"
            class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3"
          >
            <div
              v-for="item in 6"
              :key="item"
              class="h-48 animate-pulse rounded-xl border border-border bg-white"
            />
          </div>

          <article
            v-else-if="errorMessage"
            class="rounded-xl border border-[#fecaca] bg-[#fef2f2] p-5"
          >
            <div class="flex items-start gap-3">
              <PhWarningCircle
                :size="22"
                class="mt-0.5 shrink-0 text-[#dc2626]"
                weight="duotone"
              />
              <div>
                <p class="text-sm font-semibold text-foreground">
                  Tidak bisa memuat materi
                </p>
                <p class="mt-1 text-sm leading-6 text-[#7a7385]">
                  {{ errorMessage }}
                </p>
                <button
                  class="mt-4 rounded-lg bg-brand px-4 py-2 text-sm font-medium text-white transition hover:bg-[#4338ca]"
                  type="button"
                  @click="loadSubject"
                >
                  Coba lagi
                </button>
              </div>
            </div>
          </article>

          <article
            v-else-if="materials.length === 0"
            class="rounded-xl border border-border bg-white p-6 text-center"
          >
            <div
              class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-[#eef2ff] text-brand"
            >
              <PhBookOpen class="h-6 w-6" weight="duotone" />
            </div>
            <p class="mt-3 text-base font-semibold text-foreground">
              Belum ada materi
            </p>
            <p class="mt-2 text-sm leading-6 text-muted">
              Materi akan tampil setelah guru menambahkan konten pada mata
              pelajaran ini.
            </p>
          </article>

          <div v-else class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
            <article
              v-for="material in materials"
              :key="material.materialId"
              class="group flex min-w-0 cursor-pointer flex-col rounded-xl border border-border bg-white p-4 transition hover:-translate-y-0.5 hover:border-[#c7c3ef] hover:shadow-[0_14px_30px_rgba(66,55,40,0.07)]"
              tabindex="0"
              @click="
                router.push(
                  `/student/subjects/${subjectClassId}/materials/${material.materialId}`,
                )
              "
              @keydown.enter="
                router.push(
                  `/student/subjects/${subjectClassId}/materials/${material.materialId}`,
                )
              "
            >
              <div class="flex items-start justify-between gap-3">
                <div
                  class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-[#eef2ff] text-brand"
                >
                  <PhFileText :size="19" weight="duotone" />
                </div>
                <span
                  class="max-w-[55%] truncate rounded-full bg-[#f3ecff] px-2 py-1 text-[10px] font-medium uppercase text-[#8b5bb3]"
                >
                  {{ material.materialType || "Materi" }}
                </span>
              </div>

              <div class="mt-4 min-w-0 flex-1">
                <h3
                  class="line-clamp-2 text-sm font-semibold leading-5 text-foreground"
                >
                  {{ material.materialTitle }}
                </h3>
                <p
                  v-if="material.materialDesc"
                  class="mt-2 line-clamp-3 text-xs leading-5 text-[#6b6475]"
                >
                  {{ material.materialDesc }}
                </p>
                <p v-else class="mt-2 text-xs leading-5 text-[#9ca3af]">
                  Buka materi untuk melihat konten pembelajaran.
                </p>
              </div>

              <div
                class="mt-4 flex min-w-0 items-center justify-between gap-3 border-t border-[#f0ede8] pt-3"
              >
                <div class="min-w-0 text-[11px] text-[#9ca3af]">
                  <p class="truncate">
                    {{ material.creatorName || "Guru" }}
                  </p>
                  <p class="mt-0.5 truncate">
                    {{ formatDateTime(material.createdAt) }}
                  </p>
                </div>
                <div class="flex shrink-0 items-center gap-3">
                  <span
                    v-if="material.attachments?.length"
                    class="inline-flex items-center gap-1 text-[11px] text-muted"
                    :title="`${material.attachments.length} lampiran`"
                  >
                    <PhPaperclip :size="14" />
                    {{ material.attachments.length }}
                  </span>
                  <span
                    class="inline-flex items-center gap-1 text-xs font-medium text-brand"
                  >
                    Buka
                    <PhArrowRight
                      :size="14"
                      class="transition group-hover:translate-x-0.5"
                    />
                  </span>
                </div>
              </div>
            </article>
          </div>
        </template>

        <template v-else-if="activeTab === 'assignments'">
          <div v-if="isLoading" class="space-y-3">
            <div
              v-for="item in 4"
              :key="item"
              class="h-24 animate-pulse rounded-xl border border-border bg-white"
            />
          </div>

          <article
            v-else-if="errorMessage"
            class="rounded-xl border border-[#fecaca] bg-[#fef2f2] p-5"
          >
            <div class="flex items-start gap-3">
              <PhWarningCircle
                :size="22"
                class="mt-0.5 shrink-0 text-[#dc2626]"
                weight="duotone"
              />
              <div>
                <p class="text-sm font-semibold text-foreground">
                  Tidak bisa memuat tugas
                </p>
                <p class="mt-1 text-sm leading-6 text-[#7a7385]">
                  {{ errorMessage }}
                </p>
                <button
                  class="mt-4 rounded-lg bg-brand px-4 py-2 text-sm font-medium text-white transition hover:bg-[#4338ca]"
                  type="button"
                  @click="loadSubject"
                >
                  Coba lagi
                </button>
              </div>
            </div>
          </article>

          <article
            v-else-if="assignments.length === 0"
            class="rounded-xl border border-border bg-white p-6 text-center"
          >
            <div
              class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-[#eef2ff] text-brand"
            >
              <PhClipboardText class="h-6 w-6" weight="duotone" />
            </div>
            <p class="mt-3 text-base font-semibold text-foreground">
              Belum ada tugas
            </p>
            <p class="mt-2 text-sm leading-6 text-muted">
              Tugas akan tampil setelah guru menambahkannya pada mata pelajaran
              ini.
            </p>
          </article>

          <div v-else class="space-y-3">
            <article
              v-for="assignment in assignments"
              :key="assignment.assignmentId"
              class="group cursor-pointer rounded-xl border border-border bg-white p-4 transition hover:border-[#c7c3ef] hover:shadow-[0_10px_24px_rgba(66,55,40,0.06)]"
              tabindex="0"
              @click="
                router.push(
                  `/student/subjects/${subjectClassId}/assignments/${assignment.assignmentId}`,
                )
              "
              @keydown.enter="
                router.push(
                  `/student/subjects/${subjectClassId}/assignments/${assignment.assignmentId}`,
                )
              "
            >
              <div
                class="flex min-w-0 flex-col gap-4 sm:flex-row sm:items-center"
              >
                <div
                  class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-[#eef2ff] text-brand"
                >
                  <PhClipboardText :size="20" weight="duotone" />
                </div>

                <div class="min-w-0 flex-1">
                  <div class="flex flex-wrap items-center gap-2">
                    <h3 class="text-sm font-semibold text-foreground">
                      {{ assignment.assignmentTitle }}
                    </h3>
                    <span
                      v-if="assignment.categoryName"
                      class="rounded-full bg-[#eff6ff] px-2 py-1 text-[10px] font-medium text-[#2563eb]"
                    >
                      {{ assignment.categoryName }}
                    </span>
                  </div>
                  <p
                    v-if="assignment.assignmentDescription"
                    class="mt-1 line-clamp-2 text-xs leading-5 text-[#6b6475]"
                  >
                    {{ assignment.assignmentDescription }}
                  </p>
                  <div
                    class="mt-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-[11px] text-[#7a7385]"
                  >
                    <span class="inline-flex items-center gap-1.5">
                      <PhCalendarBlank :size="14" />
                      {{
                        assignment.deadline
                          ? `Tenggat ${formatDateTime(assignment.deadline)}`
                          : "Tanpa tenggat"
                      }}
                    </span>
                    <span
                      v-if="assignment.attachments?.length"
                      class="inline-flex items-center gap-1.5"
                    >
                      <PhPaperclip :size="14" />
                      {{ assignment.attachments.length }} lampiran
                    </span>
                  </div>
                </div>

                <div
                  class="flex shrink-0 items-center justify-between gap-3 sm:justify-end"
                >
                  <span class="text-xs text-[#9ca3af]">
                    Status tersedia di detail
                  </span>
                  <span
                    class="inline-flex h-8 w-8 items-center justify-center rounded-lg border border-[#ddd8e4] text-brand transition group-hover:border-brand group-hover:bg-[#eef2ff]"
                    title="Buka tugas"
                  >
                    <PhArrowRight :size="16" />
                  </span>
                </div>
              </div>
            </article>
          </div>
        </template>

        <template v-else>
          <div v-if="notesLoading" class="grid gap-3 md:grid-cols-2">
            <div
              v-for="item in 4"
              :key="item"
              class="h-44 animate-pulse rounded-xl border border-border bg-white"
            />
          </div>

          <article
            v-else-if="notesError"
            class="rounded-xl border border-[#fecaca] bg-[#fef2f2] p-5"
          >
            <div class="flex items-start gap-3">
              <PhWarningCircle
                :size="22"
                class="mt-0.5 shrink-0 text-[#dc2626]"
                weight="duotone"
              />
              <div>
                <p class="text-sm font-semibold text-foreground">
                  Tidak bisa memuat catatan
                </p>
                <p class="mt-1 text-sm leading-6 text-[#7a7385]">
                  {{ notesError }}
                </p>
                <button
                  class="mt-4 rounded-lg bg-brand px-4 py-2 text-sm font-medium text-white transition hover:bg-[#4338ca]"
                  type="button"
                  @click="loadNotes(true)"
                >
                  Coba lagi
                </button>
              </div>
            </div>
          </article>

          <article
            v-else-if="notes.length === 0"
            class="rounded-xl border border-border bg-white p-6 text-center"
          >
            <div
              class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-[#eef2ff] text-brand"
            >
              <PhNotebook class="h-6 w-6" weight="duotone" />
            </div>
            <p class="mt-3 text-base font-semibold text-foreground">
              Belum ada catatan untuk materi di mata pelajaran ini
            </p>
            <p class="mt-2 text-sm leading-6 text-muted">
              Catatan yang kamu simpan dari halaman materi akan tampil di sini.
            </p>
          </article>

          <div v-else class="grid gap-3 md:grid-cols-2">
            <article
              v-for="note in notes"
              :key="note.noteId"
              class="flex min-w-0 flex-col rounded-xl border border-border bg-white shadow-sm p-4"
            >
              <div class="flex min-w-0 items-start gap-3">
                <div
                  class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-[#f3ecff] text-[#8b5bb3]"
                >
                  <PhNotebook :size="18" weight="duotone" />
                </div>
                <div class="min-w-0">
                  <h3 class="truncate text-sm font-semibold text-foreground">
                    {{ note.materialTitle }}
                  </h3>
                  <p class="mt-0.5 text-[11px] text-[#9ca3af]">
                    Diperbarui {{ formatDateTime(note.updatedAt) }}
                  </p>
                </div>
              </div>

              <p
                class="mt-4 line-clamp-5 flex-1 whitespace-pre-line wrap-break-word text-sm leading-6 text-[#6b6475]"
              >
                {{ note.content }}
              </p>

              <div
                class="mt-4 flex flex-wrap items-center gap-2 border-t border-[#f0ede8] pt-3"
              >
                <button
                  class="rounded-lg border border-[#ddd8e4] px-3 py-2 text-xs font-medium text-[#6b6475] transition hover:bg-[#f8f7f4]"
                  type="button"
                  @click="
                    router.push(
                      `/student/subjects/${subjectClassId}/materials/${note.materialId}`,
                    )
                  "
                >
                  Lihat materi
                </button>
                <button
                  class="inline-flex items-center gap-2 rounded-lg bg-brand px-3 py-2 text-xs font-medium text-white transition hover:bg-[#4338ca]"
                  type="button"
                  @click="
                    router.push(
                      `/student/subjects/${subjectClassId}/materials/${note.materialId}/note`,
                    )
                  "
                >
                  Buka catatan
                  <PhArrowRight :size="14" />
                </button>
              </div>
            </article>
          </div>
        </template>
      </div>
    </section>
  </main>
</template>
