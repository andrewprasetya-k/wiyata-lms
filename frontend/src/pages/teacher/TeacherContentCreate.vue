<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  PhArrowLeft,
  PhClipboardText,
  PhFileText,
  PhPaperPlaneTilt,
  PhInfo,
  PhCalendarBlank,
  PhClock,
} from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import { useToastStore } from "../../stores/toast";
import { getMyTeachingSubjectClassById } from "../../services/teacherSubjects";
import {
  getAssignmentCategories,
  createAssignment,
  updateAssignment,
} from "../../services/teacherAssignment";
import { createMaterial, updateMaterial } from "../../services/teacherMaterial";
import { getSubjectAssignmentDetail } from "../../services/assignment";
import { getMaterialById } from "../../services/classWorkspace";
import { deleteMedia } from "../../services/media";
import MediaUploader from "../../components/common/MediaUploader.vue";
import type { TeacherSubjectClass } from "../../types/teacherSubjects";
import type { AssignmentCategory } from "../../types/teacherAssignment";
import {
  formatDateInputValue,
  formatTimeInputValue,
} from "../../utils/date";

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();
const toast = useToastStore();

const subjectClassId = computed(() =>
  String(route.params.subjectClassId ?? ""),
);
const subject = ref<TeacherSubjectClass | null>(null);
const categories = ref<AssignmentCategory[]>([]);
const materialId = computed(() => String(route.params.matId ?? ""));
const assignmentId = computed(() => String(route.params.asgId ?? ""));
const isEditMode = computed(() => !!materialId.value || !!assignmentId.value);
const activeTab = ref<"material" | "assignment">("material");
const existingAttachments = ref<any[]>([]);
const loading = ref(false);
const submitting = ref(false);
const errorMessage = ref("");
const categoryErrorMessage = ref("");
const uploaderKey = ref(0);
const isUploadingMedia = ref(false);
const hasMediaUploadError = ref(false);
const activeSchoolId = computed(
  () => auth.activeSchoolId ?? auth.defaultContext?.schoolId ?? "",
);
const activeSchoolCode = computed(() => {
  const activeMembership = auth.memberships.find(
    (membership) => membership.school.id === activeSchoolId.value,
  );
  return activeMembership?.school.code ?? "";
});
const hasRequiredContext = computed(() =>
  Boolean(activeSchoolId.value && subjectClassId.value && subject.value),
);
const isSubmitDisabled = computed(
  () =>
    submitting.value ||
    !hasRequiredContext.value ||
    isUploadingMedia.value ||
    hasMediaUploadError.value ||
    (activeTab.value === "assignment" && categories.value.length === 0),
);

function buildJakartaDeadlineISOString(date: string, time: string) {
  if (!date) return undefined;
  return `${date}T${time || "23:59"}:00+07:00`;
}

// Form State
const form = ref({
  title: "",
  description: "",
  materialType: "pdf" as "pdf" | "video" | "ppt" | "other",
  categoryId: "",
  deadlineDate: "",
  deadlineTime: "23:59",
  allowLate: false,
  mediaIds: [] as string[],
});

async function loadInitialData() {
  loading.value = true;
  errorMessage.value = "";
  categoryErrorMessage.value = "";

  if (!isEditMode.value) {
    activeTab.value =
      route.query.type === "assignment" ? "assignment" : "material";
  }

  try {
    if (!subjectClassId.value) {
      errorMessage.value =
        "Konteks subject belum tersedia. Pilih subject terlebih dahulu.";
      return;
    }
    if (!activeSchoolId.value) {
      errorMessage.value =
        "Konteks sekolah belum tersedia. Silakan login ulang.";
      return;
    }

    const subjectData = await getMyTeachingSubjectClassById(
      subjectClassId.value,
    );
    if (!subjectData) {
      errorMessage.value =
        "Subject ini tidak tersedia untuk akun guru pada school aktif.";
      return;
    }
    subject.value = subjectData;

    if (!activeSchoolCode.value) {
      categoryErrorMessage.value =
        "Kode sekolah aktif belum tersedia. Kategori tugas tidak bisa dimuat.";
      return;
    }

    try {
      const categoriesData = await getAssignmentCategories(
        activeSchoolCode.value,
      );
      categories.value = categoriesData.categories;
      if (categories.value.length > 0 && !isEditMode.value) {
        form.value.categoryId = categories.value[0].categoryId;
      }
    } catch {
      categories.value = [];
      categoryErrorMessage.value = "Kategori tugas belum bisa dimuat.";
    }

    if (isEditMode.value) {
      if (materialId.value) {
        activeTab.value = "material";
        const mat = await getMaterialById(materialId.value);
        if (mat) {
          form.value.title = mat.materialTitle;
          form.value.description = mat.materialDesc || "";
          form.value.materialType = (mat.materialType as any) || "pdf";
          if (mat.attachments) {
            existingAttachments.value = mat.attachments.map((a: any) => ({
              mediaId: a.mediaId,
              mediaName: a.mediaName,
              fileSize: a.fileSize,
              fileUrl: a.fileUrl,
            }));
            form.value.mediaIds = mat.attachments.map((a: any) => a.mediaId);
          }
        }
      } else if (assignmentId.value) {
        activeTab.value = "assignment";
        const asgData = await getSubjectAssignmentDetail(
          subjectClassId.value,
          assignmentId.value,
        );
        if (asgData && asgData.assignment) {
          const asg = asgData.assignment;
          form.value.title = asg.assignmentTitle;
          form.value.description = asg.assignmentDescription || "";
          form.value.allowLate = asg.allowLateSubmission ?? false;
          if (asg.deadline) {
            form.value.deadlineDate = formatDateInputValue(asg.deadline);
            form.value.deadlineTime = formatTimeInputValue(asg.deadline);
          }
          if (asg.attachments) {
            existingAttachments.value = asg.attachments.map((a: any) => ({
              mediaId: a.mediaId,
              mediaName: a.mediaName,
              fileSize: a.fileSize,
              fileUrl: a.fileUrl,
            }));
            form.value.mediaIds = asg.attachments.map((a: any) => a.mediaId);
          }
          // find category ID by name
          const cat = categories.value.find(
            (c) => c.categoryName === asg.categoryName,
          );
          if (cat) {
            form.value.categoryId = cat.categoryId;
          }
        }
      }
    }
  } catch (err) {
    errorMessage.value = "Gagal memuat data pendukung. Coba refresh halaman.";
  } finally {
    loading.value = false;
  }
}

async function handleSubmit() {
  errorMessage.value = "";

  if (!activeSchoolId.value) {
    errorMessage.value = "Konteks sekolah belum tersedia. Silakan login ulang.";
    return;
  }
  if (!subjectClassId.value || !subject.value) {
    errorMessage.value =
      "Konteks subject belum tersedia. Pilih subject terlebih dahulu.";
    return;
  }
  if (!form.value.title.trim()) {
    errorMessage.value = "Judul wajib diisi.";
    return;
  }
  if (isUploadingMedia.value) {
    errorMessage.value = "Tunggu sampai upload selesai sebelum menerbitkan.";
    return;
  }
  if (hasMediaUploadError.value) {
    errorMessage.value =
      "Ada lampiran yang gagal diunggah. Hapus atau unggah ulang file tersebut.";
    return;
  }
  if (activeTab.value === "assignment" && !form.value.categoryId) {
    errorMessage.value =
      "Kategori tugas belum tersedia. Tambahkan kategori terlebih dahulu sebelum membuat tugas.";
    return;
  }

  submitting.value = true;
  try {
    if (activeTab.value === "material") {
      if (isEditMode.value && materialId.value) {
        await updateMaterial(materialId.value, {
          materialTitle: form.value.title,
          materialDesc: form.value.description,
          materialType: form.value.materialType,
          mediaIds: form.value.mediaIds,
        });
        toast.success("Materi berhasil diperbarui.");
      } else {
        await createMaterial({
          schoolId: activeSchoolId.value,
          subjectClassId: subjectClassId.value,
          materialTitle: form.value.title,
          materialDesc: form.value.description,
          materialType: form.value.materialType,
          mediaIds: form.value.mediaIds,
        });
        toast.success("Materi berhasil dibuat.");
      }
    } else {
      const deadline = buildJakartaDeadlineISOString(
        form.value.deadlineDate,
        form.value.deadlineTime,
      );

      if (isEditMode.value && assignmentId.value) {
        await updateAssignment(assignmentId.value, {
          categoryId: form.value.categoryId,
          assignmentTitle: form.value.title,
          assignmentDescription: form.value.description,
          deadline,
          allowLateSubmission: form.value.allowLate,
          mediaIds: form.value.mediaIds,
        });
        toast.success("Tugas berhasil diperbarui.");
      } else {
        await createAssignment({
          schoolId: activeSchoolId.value,
          subjectClassId: subjectClassId.value,
          categoryId: form.value.categoryId,
          assignmentTitle: form.value.title,
          assignmentDescription: form.value.description,
          deadline,
          allowLateSubmission: form.value.allowLate,
          mediaIds: form.value.mediaIds,
        });
        toast.success("Tugas berhasil dibuat.");
      }
    }

    router.push(`/teacher/subjects/${subjectClassId.value}`);
  } catch (err) {
    const uploadedMediaIds = [...form.value.mediaIds];
    if (uploadedMediaIds.length > 0) {
      await Promise.allSettled(
        uploadedMediaIds.map(async (mediaId) => {
          try {
            await deleteMedia(mediaId);
          } catch (cleanupError) {
            console.warn(
              "Failed to cleanup uploaded teacher content media",
              mediaId,
              cleanupError,
            );
          }
        }),
      );
      form.value.mediaIds = [];
      isUploadingMedia.value = false;
      hasMediaUploadError.value = false;
      uploaderKey.value += 1;
    }

    errorMessage.value =
      getErrorMessage(err) ??
      "Gagal menyimpan konten. Jika lampiran sudah terunggah, lampiran perlu dipilih ulang.";
  } finally {
    submitting.value = false;
  }
}

function getErrorMessage(error: unknown) {
  if (typeof error === "object" && error !== null && "response" in error) {
    const response = (
      error as { response?: { data?: { error?: string; message?: string } } }
    ).response;
    return response?.data?.error ?? response?.data?.message;
  }
  return undefined;
}

onMounted(loadInitialData);
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-[#f8f7f4]">
    <header class="border-b border-border bg-white">
      <div class="px-5 py-5 sm:px-6 lg:px-8">
        <div class="flex min-w-0 items-center gap-2 text-xs text-muted">
          <button
            type="button"
            class="inline-flex shrink-0 items-center gap-1.5 transition hover:text-brand"
            @click="router.back()"
          >
            <PhArrowLeft :size="15" />
            Mata pelajaran
          </button>
          <span class="text-[#d1d5db]">/</span>
          <span class="min-w-0 truncate font-medium text-foreground">
            {{ isEditMode ? "Edit" : "Buat" }}
            {{ activeTab === "material" ? "Materi" : "Tugas" }}
          </span>
        </div>

        <div
          class="mt-4 flex min-w-0 flex-col gap-3 lg:flex-row lg:items-center lg:justify-between"
        >
          <div class="min-w-0">
            <p class="text-xs font-medium uppercase tracking-wide text-[#7b61a8]">
              {{ subject?.subjectName || "Konten mata pelajaran" }}
            </p>
            <h1 class="mt-1 text-xl font-semibold text-foreground sm:text-2xl">
              {{ isEditMode ? "Edit" : "Buat" }}
              {{ activeTab === "material" ? "Materi" : "Tugas" }}
            </h1>
            <p class="mt-1 text-sm text-muted">
              <template v-if="subject">
                {{ subject.className }}
                <span v-if="subject.subjectCode">
                  · {{ subject.subjectCode }}
                </span>
              </template>
              <template v-else> Lengkapi informasi konten untuk siswa. </template>
            </p>
          </div>
          <span
            class="inline-flex self-start rounded-lg px-2.5 py-1.5 text-xs font-medium lg:self-auto"
            :class="
              activeTab === 'material'
                ? 'bg-brand-soft text-brand'
                : 'bg-warning-soft text-[#ea580c]'
            "
          >
            {{ activeTab === "material" ? "Materi" : "Tugas" }}
          </span>
        </div>
      </div>
    </header>

    <section class="px-5 py-5 sm:px-6 lg:px-8 lg:py-6">
      <section
        v-if="errorMessage"
        class="mb-5 flex items-start gap-3 rounded-xl border border-danger-line bg-danger-soft p-4 text-sm leading-6 text-[#b42318]"
      >
        <PhInfo :size="19" class="mt-0.5 shrink-0" weight="duotone" />
        <p>{{ errorMessage }}</p>
      </section>

      <template v-if="loading">
        <div class="grid gap-5 lg:grid-cols-[minmax(0,1fr)_320px]">
          <div class="space-y-4">
            <div
              class="h-56 animate-pulse rounded-xl border border-border bg-white"
            />
            <div
              class="h-48 animate-pulse rounded-xl border border-border bg-white"
            />
          </div>
          <div
            class="h-80 animate-pulse rounded-xl border border-border bg-white"
          />
        </div>
      </template>

      <template v-else>
        <div
          class="mb-5 flex max-w-full gap-2 overflow-x-auto rounded-xl border border-border bg-white p-1.5 sm:w-fit"
        >
          <button
            type="button"
            class="inline-flex shrink-0 items-center justify-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition sm:min-w-28"
            :class="[
              activeTab === 'material'
                ? 'bg-brand-soft text-brand'
                : 'text-muted',
              !isEditMode && activeTab !== 'material'
                ? 'cursor-pointer hover:bg-[#f3f1ec] hover:text-foreground'
                : isEditMode && activeTab !== 'material'
                  ? 'cursor-not-allowed opacity-50'
                  : '',
            ]"
            :disabled="isEditMode"
            @click="!isEditMode && (activeTab = 'material')"
          >
            <PhFileText :size="17" weight="duotone" />
            Materi
          </button>
          <button
            type="button"
            class="inline-flex shrink-0 items-center justify-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition sm:min-w-28"
            :class="[
              activeTab === 'assignment'
                ? 'bg-brand-soft text-brand'
                : 'text-muted',
              !isEditMode && activeTab !== 'assignment'
                ? 'cursor-pointer hover:bg-[#f3f1ec] hover:text-foreground'
                : isEditMode && activeTab !== 'assignment'
                  ? 'cursor-not-allowed opacity-50'
                  : '',
            ]"
            :disabled="isEditMode"
            @click="!isEditMode && (activeTab = 'assignment')"
          >
            <PhClipboardText :size="17" weight="duotone" />
            Tugas
          </button>
        </div>

        <div class="grid min-w-0 gap-5 lg:grid-cols-[minmax(0,1fr)_320px]">
          <div class="min-w-0 space-y-5">
            <section
              class="rounded-xl border border-border bg-white p-5 sm:p-6"
            >
              <div class="flex items-start gap-3">
                <div
                  class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-brand-soft text-brand"
                >
                  <PhInfo :size="20" weight="duotone" />
                </div>
                <div>
                  <h2 class="text-base font-semibold text-foreground">
                    Informasi utama
                  </h2>
                  <p class="mt-1 text-xs leading-5 text-[#8a8494]">
                    Isi judul dan
                    {{
                      activeTab === "material"
                        ? "deskripsi materi"
                        : "instruksi tugas"
                    }}
                    untuk siswa.
                  </p>
                </div>
              </div>

              <div class="mt-5 space-y-5">
                <div>
                  <label
                    class="block text-sm font-medium text-[#374151]"
                    for="content-title"
                  >
                    Judul
                    {{ activeTab === "material" ? "materi" : "tugas" }}
                  </label>
                  <input
                    id="content-title"
                    v-model="form.title"
                    type="text"
                    class="mt-2 w-full rounded-lg border border-border bg-[#fbfaf8] px-4 py-3 text-sm text-foreground outline-none transition placeholder:text-[#a09aa8] focus:border-brand focus:bg-white"
                    placeholder="Contoh: Pengenalan Aljabar Linear"
                  />
                </div>

                <div>
                  <label
                    class="block text-sm font-medium text-[#374151]"
                    for="content-description"
                  >
                    {{
                      activeTab === "material"
                        ? "Deskripsi materi"
                        : "Deskripsi atau instruksi"
                    }}
                    <span class="font-normal text-[#9ca3af]">(opsional)</span>
                  </label>
                  <textarea
                    id="content-description"
                    v-model="form.description"
                    rows="7"
                    class="mt-2 w-full resize-none rounded-lg border border-border bg-[#fbfaf8] px-4 py-3 text-sm leading-6 text-[#374151] outline-none transition placeholder:text-[#a09aa8] focus:border-brand focus:bg-white"
                    placeholder="Berikan instruksi atau detail tambahan..."
                  />
                </div>
              </div>
            </section>

            <section
              class="rounded-xl border border-border bg-white p-5 sm:p-6"
            >
              <div class="flex items-start gap-3">
                <div
                  class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-[#f3f1ec] text-muted"
                >
                  <PhFileText :size="20" weight="duotone" />
                </div>
                <div>
                  <h2 class="text-base font-semibold text-foreground">
                    Lampiran dan media
                  </h2>
                  <p class="mt-1 text-xs leading-5 text-[#8a8494]">
                    Tambahkan file pendukung yang dibutuhkan siswa.
                  </p>
                </div>
              </div>

              <div class="mt-5">
                <MediaUploader
                  v-if="hasRequiredContext"
                  :key="uploaderKey"
                  :school-id="activeSchoolId"
                  :owner-type="activeTab"
                  :initial-media="existingAttachments"
                  v-model:is-uploading="isUploadingMedia"
                  v-model:has-upload-error="hasMediaUploadError"
                  cleanup-on-remove
                  @update:media-ids="form.mediaIds = $event"
                />
                <p
                  v-else
                  class="rounded-lg border border-danger-line bg-danger-soft p-4 text-sm leading-6 text-[#b42318]"
                >
                  Lampiran belum bisa diunggah sampai konteks sekolah dan mata
                  pelajaran tersedia.
                </p>
                <p
                  v-if="isUploadingMedia"
                  class="mt-3 rounded-lg bg-brand-soft p-4 text-sm leading-6 text-brand-hover"
                >
                  Tunggu sampai upload selesai sebelum menerbitkan.
                </p>
                <p
                  v-if="hasMediaUploadError"
                  class="mt-3 rounded-lg bg-danger-soft p-4 text-sm leading-6 text-[#b42318]"
                >
                  Ada lampiran yang gagal diunggah. Hapus atau unggah ulang file
                  tersebut.
                </p>
              </div>
            </section>
          </div>

          <aside class="min-w-0">
            <div class="space-y-4 lg:sticky lg:top-6">
              <section class="rounded-xl border border-border bg-white p-5">
                <p
                  class="text-[10px] font-medium uppercase tracking-[0.08em] text-[#9ca3af]"
                >
                  Metadata
                </p>
                <h2 class="mt-1 text-base font-semibold text-foreground">
                  Pengaturan konten
                </h2>

                <div v-if="activeTab === 'material'" class="mt-5">
                  <label
                    class="block text-xs font-medium text-muted"
                    for="material-type"
                  >
                    Tipe materi
                  </label>
                  <select
                    id="material-type"
                    v-model="form.materialType"
                    class="mt-2 w-full rounded-lg border border-border bg-[#fbfaf8] px-3.5 py-2.5 text-sm text-[#374151] outline-none transition focus:border-brand focus:bg-white"
                  >
                    <option value="pdf">PDF</option>
                    <option value="video">Video</option>
                    <option value="ppt">PPT / Slide</option>
                    <option value="other">Lainnya</option>
                  </select>
                </div>

                <div v-else class="mt-5 space-y-5">
                  <div>
                    <label
                      class="block text-xs font-medium text-muted"
                      for="assignment-category"
                    >
                      Kategori tugas
                    </label>
                    <select
                      id="assignment-category"
                      v-model="form.categoryId"
                      :disabled="categories.length === 0"
                      class="mt-2 w-full rounded-lg border border-border bg-[#fbfaf8] px-3.5 py-2.5 text-sm text-[#374151] outline-none transition focus:border-brand focus:bg-white disabled:cursor-not-allowed disabled:opacity-60"
                    >
                      <option
                        v-for="cat in categories"
                        :key="cat.categoryId"
                        :value="cat.categoryId"
                      >
                        {{ cat.categoryName }}
                      </option>
                    </select>
                    <p
                      v-if="categoryErrorMessage || categories.length === 0"
                      class="mt-2 text-xs leading-5 text-[#b42318]"
                    >
                      {{
                        categoryErrorMessage || "Kategori tugas belum tersedia."
                      }}
                    </p>
                  </div>

                  <div>
                    <p class="text-xs font-medium text-muted">Tenggat</p>
                    <div class="mt-2 grid gap-2">
                      <div class="relative">
                        <PhCalendarBlank
                          :size="16"
                          class="absolute left-3 top-1/2 -translate-y-1/2 text-[#9ca3af]"
                        />
                        <input
                          v-model="form.deadlineDate"
                          type="date"
                          class="w-full rounded-lg border border-border bg-[#fbfaf8] py-2.5 pl-10 pr-3 text-sm text-[#374151] outline-none transition focus:border-brand focus:bg-white"
                        />
                      </div>
                      <div class="relative">
                        <PhClock
                          :size="16"
                          class="absolute left-3 top-1/2 -translate-y-1/2 text-[#9ca3af]"
                        />
                        <input
                          v-model="form.deadlineTime"
                          type="time"
                          class="w-full rounded-lg border border-border bg-[#fbfaf8] py-2.5 pl-10 pr-3 text-sm text-[#374151] outline-none transition focus:border-brand focus:bg-white"
                        />
                      </div>
                    </div>
                  </div>

                  <div class="border-t border-[#f3f4f6] pt-4">
                    <div class="flex items-center justify-between gap-3">
                      <div class="min-w-0">
                        <p class="text-xs font-medium text-[#374151]">
                          Izinkan terlambat
                        </p>
                        <p class="mt-1 text-[11px] leading-4 text-[#9ca3af]">
                          Siswa tetap dapat mengumpulkan setelah tenggat.
                        </p>
                      </div>
                      <button
                        type="button"
                        class="relative h-5 w-10 shrink-0 rounded-full transition"
                        :class="
                          form.allowLate ? 'bg-brand' : 'bg-[#e5e7eb]'
                        "
                        :aria-pressed="form.allowLate"
                        @click="form.allowLate = !form.allowLate"
                      >
                        <span
                          class="absolute left-0.5 top-0.5 h-4 w-4 rounded-full bg-white shadow-sm transition-transform"
                          :class="
                            form.allowLate ? 'translate-x-5' : 'translate-x-0'
                          "
                        />
                      </button>
                    </div>
                  </div>
                </div>
              </section>

              <section
                class="rounded-xl border border-[#fed7aa] bg-warning-soft p-4"
              >
                <p
                  class="text-[10px] font-medium uppercase tracking-[0.08em] text-[#ea580c]"
                >
                  Publikasi
                </p>
                <p class="mt-2 text-xs leading-5 text-[#9a3412]">
                  Konten akan langsung tersedia bagi siswa di
                  <strong>{{ subject?.className }}</strong> setelah disimpan.
                </p>
              </section>

              <section class="rounded-xl border border-border bg-white p-4">
                <div class="grid gap-2">
                  <button
                    type="button"
                    class="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-brand px-4 py-2.5 text-sm font-medium text-white transition hover:bg-brand-hover disabled:cursor-not-allowed disabled:opacity-50"
                    :disabled="isSubmitDisabled"
                    @click="handleSubmit"
                  >
                    <PhPaperPlaneTilt
                      v-if="!submitting"
                      :size="17"
                      weight="bold"
                    />
                    {{
                      submitting
                        ? "Menyimpan..."
                        : isEditMode
                          ? "Simpan perubahan"
                          : "Publish"
                    }}
                  </button>
                  <button
                    type="button"
                    class="inline-flex w-full items-center justify-center rounded-lg border border-border bg-white px-4 py-2.5 text-sm font-medium text-[#374151] transition hover:bg-[#f3f1ec]"
                    @click="router.back()"
                  >
                    Batal
                  </button>
                </div>
              </section>
            </div>
          </aside>
        </div>
      </template>
    </section>
  </main>
</template>
