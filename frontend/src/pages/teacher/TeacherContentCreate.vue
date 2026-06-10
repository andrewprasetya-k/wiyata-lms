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
import { getMyTeachingSubjectClassById } from "../../services/teacherSubjects";
import {
  getAssignmentCategories,
  createAssignment,
} from "../../services/teacherAssignment";
import { createMaterial } from "../../services/teacherMaterial";
import { deleteMedia } from "../../services/media";
import MediaUploader from "../../components/common/MediaUploader.vue";
import type { TeacherSubjectClass } from "../../types/teacherSubjects";
import type { AssignmentCategory } from "../../types/teacherAssignment";

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();

const subjectClassId = computed(() =>
  String(route.params.subjectClassId ?? ""),
);
const subject = ref<TeacherSubjectClass | null>(null);
const categories = ref<AssignmentCategory[]>([]);
const activeTab = ref<"material" | "assignment">("material");
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
      if (categories.value.length > 0) {
        form.value.categoryId = categories.value[0].categoryId;
      }
    } catch {
      categories.value = [];
      categoryErrorMessage.value = "Kategori tugas belum bisa dimuat.";
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
      await createMaterial({
        schoolId: activeSchoolId.value,
        subjectClassId: subjectClassId.value,
        materialTitle: form.value.title,
        materialDesc: form.value.description,
        materialType: form.value.materialType,
        mediaIds: form.value.mediaIds,
      });
    } else {
      let deadline = undefined;
      if (form.value.deadlineDate) {
        deadline = `${form.value.deadlineDate}T${form.value.deadlineTime}:00Z`;
      }

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
  <main class="min-h-screen flex-1 px-4 py-5 sm:px-6 lg:px-8">
    <div class="mx-auto max-w-6xl">
      <!-- Topbar / Breadcrumb -->
      <div class="mb-5 flex items-center justify-between">
        <div class="flex items-center gap-4">
          <button
            @click="router.back()"
            class="flex items-center gap-2 text-sm font-medium text-[#6B7280] hover:text-[#111827] transition"
          >
            <PhArrowLeft :size="18" />
            <span class="hidden sm:inline">{{
              subject?.subjectName || "Kembali"
            }}</span>
          </button>
          <span class="text-[#D1D5DB]">/</span>
          <h1 class="text-sm font-semibold text-[#111827]">Buat Konten Baru</h1>
        </div>

        <div class="flex items-center gap-3">
          <button
            @click="router.back()"
            class="px-4 py-2 text-sm font-medium text-[#374151] bg-white border border-[#EBEBEB] rounded-xl hover:bg-[#F9FAFB] transition"
          >
            Batal
          </button>
          <button
            @click="handleSubmit"
            :disabled="isSubmitDisabled"
            class="flex items-center gap-2 px-5 py-2 text-sm font-medium text-white bg-[#4F46E5] rounded-xl hover:bg-[#4338CA] transition disabled:opacity-50"
          >
            <PhPaperPlaneTilt v-if="!submitting" :size="18" weight="bold" />
            {{ submitting ? "Menyimpan..." : "Terbitkan" }}
          </button>
        </div>
      </div>

      <div
        v-if="errorMessage"
        class="mb-5 rounded-[18px] border border-[#FECACA] bg-[#FEF2F2] p-4 text-sm leading-6 text-[#B42318]"
      >
        {{ errorMessage }}
      </div>

      <div
        v-if="loading"
        class="rounded-[18px] bg-white p-5 text-sm text-[#6B7280] shadow-sm ring-1 ring-black/5"
      >
        Memuat data pendukung...
      </div>

      <!-- Type Switcher -->
      <div
        v-if="!loading"
        class="mb-5 flex w-fit gap-2 rounded-2xl bg-[#F3F4F6] p-1.5"
      >
        <button
          @click="activeTab = 'material'"
          :class="[
            'flex items-center gap-2 px-6 py-2.5 text-sm font-medium rounded-xl transition',
            activeTab === 'material'
              ? 'bg-white text-[#4F46E5] shadow-sm'
              : 'text-[#6B7280] hover:text-[#111827]',
          ]"
        >
          <PhFileText :size="18" weight="duotone" />
          Materi
        </button>
        <button
          @click="activeTab = 'assignment'"
          :class="[
            'flex items-center gap-2 px-6 py-2.5 text-sm font-medium rounded-xl transition',
            activeTab === 'assignment'
              ? 'bg-white text-[#4F46E5] shadow-sm'
              : 'text-[#6B7280] hover:text-[#111827]',
          ]"
        >
          <PhClipboardText :size="18" weight="duotone" />
          Tugas
        </button>
      </div>

      <div v-if="!loading" class="grid gap-5 lg:grid-cols-[1fr_320px]">
        <!-- Main Form -->
        <div class="space-y-5">
          <section
            class="rounded-[18px] border border-[#EBEBEB] bg-white p-5 shadow-sm"
          >
            <h2
              class="mb-5 flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-[#374151]"
            >
              <PhInfo :size="16" weight="bold" />
              Informasi Utama
            </h2>

            <div class="space-y-5">
              <div>
                <label class="block text-sm font-medium text-[#6B7280] mb-2"
                  >Judul
                  {{ activeTab === "material" ? "Materi" : "Tugas" }}</label
                >
                <input
                  v-model="form.title"
                  type="text"
                  class="w-full px-4 py-3 bg-[#F9FAFB] border border-[#EBEBEB] rounded-2xl outline-none focus:border-[#4F46E5] transition"
                  placeholder="Contoh: Pengenalan Aljabar Linear"
                />
              </div>

              <div>
                <label class="block text-sm font-medium text-[#6B7280] mb-2"
                  >Deskripsi (Opsional)</label
                >
                <textarea
                  v-model="form.description"
                  rows="5"
                  class="w-full px-4 py-3 bg-[#F9FAFB] border border-[#EBEBEB] rounded-2xl outline-none focus:border-[#4F46E5] transition resize-none"
                  placeholder="Berikan instruksi atau detail tambahan..."
                ></textarea>
              </div>
            </div>
          </section>

          <section
            class="rounded-[18px] border border-[#EBEBEB] bg-white p-5 shadow-sm"
          >
            <h2
              class="mb-5 flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-[#374151]"
            >
              <PhFileText :size="16" weight="bold" />
              Lampiran & Media
            </h2>

            <MediaUploader
              v-if="hasRequiredContext"
              :key="uploaderKey"
              :school-id="activeSchoolId"
              :owner-type="activeTab"
              v-model:is-uploading="isUploadingMedia"
              v-model:has-upload-error="hasMediaUploadError"
              cleanup-on-remove
              @update:media-ids="form.mediaIds = $event"
            />
            <p
              v-else
              class="rounded-2xl bg-[#FEF2F2] p-4 text-sm leading-6 text-[#B42318]"
            >
              Lampiran belum bisa diunggah sampai konteks school dan subject
              tersedia.
            </p>
            <p
              v-if="isUploadingMedia"
              class="mt-3 rounded-2xl bg-[#EEF2FF] p-4 text-sm leading-6 text-[#4338CA]"
            >
              Tunggu sampai upload selesai sebelum menerbitkan.
            </p>
            <p
              v-if="hasMediaUploadError"
              class="mt-3 rounded-2xl bg-[#FEF2F2] p-4 text-sm leading-6 text-[#B42318]"
            >
              Ada lampiran yang gagal diunggah. Hapus atau unggah ulang file
              tersebut.
            </p>
          </section>
        </div>

        <!-- Sidebar Settings -->
        <aside class="space-y-5">
          <section
            class="rounded-[18px] border border-[#EBEBEB] bg-white p-5 shadow-sm"
          >
            <h2
              class="mb-5 text-xs font-bold uppercase tracking-wider text-[#374151]"
            >
              Pengaturan
            </h2>

            <div v-if="activeTab === 'material'" class="space-y-4">
              <div>
                <label class="block text-xs font-medium text-[#6B7280] mb-2"
                  >Tipe Materi</label
                >
                <select
                  v-model="form.materialType"
                  class="w-full px-3 py-2.5 bg-[#F9FAFB] border border-[#EBEBEB] rounded-xl outline-none text-sm"
                >
                  <option value="pdf">PDF</option>
                  <option value="video">Video</option>
                  <option value="ppt">PPT / Slide</option>
                  <option value="other">Lainnya</option>
                </select>
              </div>
            </div>

            <div v-else class="space-y-5">
              <div>
                <label class="block text-xs font-medium text-[#6B7280] mb-2"
                  >Kategori Tugas</label
                >
                <select
                  v-model="form.categoryId"
                  :disabled="categories.length === 0"
                  class="w-full px-3 py-2.5 bg-[#F9FAFB] border border-[#EBEBEB] rounded-xl outline-none text-sm"
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
                  class="mt-2 text-xs leading-5 text-[#B42318]"
                >
                  {{ categoryErrorMessage || "Kategori tugas belum tersedia." }}
                </p>
              </div>

              <div>
                <label class="block text-xs font-medium text-[#6B7280] mb-2"
                  >Deadline</label
                >
                <div class="space-y-2">
                  <div class="relative">
                    <PhCalendarBlank
                      :size="16"
                      class="absolute left-3 top-1/2 -translate-y-1/2 text-[#9CA3AF]"
                    />
                    <input
                      v-model="form.deadlineDate"
                      type="date"
                      class="w-full pl-10 pr-3 py-2 bg-[#F9FAFB] border border-[#EBEBEB] rounded-xl outline-none text-sm"
                    />
                  </div>
                  <div class="relative">
                    <PhClock
                      :size="16"
                      class="absolute left-3 top-1/2 -translate-y-1/2 text-[#9CA3AF]"
                    />
                    <input
                      v-model="form.deadlineTime"
                      type="time"
                      class="w-full pl-10 pr-3 py-2 bg-[#F9FAFB] border border-[#EBEBEB] rounded-xl outline-none text-sm"
                    />
                  </div>
                </div>
              </div>

              <div class="pt-2 border-t border-[#F3F4F6]">
                <label
                  class="flex items-center justify-between cursor-pointer group"
                >
                  <div class="space-y-0.5">
                    <p class="text-xs font-medium text-[#374151]">
                      Izinkan Terlambat
                    </p>
                    <p class="text-[10px] text-[#9CA3AF]">
                      Siswa tetap bisa submit
                    </p>
                  </div>
                  <div
                    @click="form.allowLate = !form.allowLate"
                    :class="[
                      'w-10 h-5 rounded-full relative transition duration-200',
                      form.allowLate ? 'bg-[#4F46E5]' : 'bg-[#E5E7EB]',
                    ]"
                  >
                    <div
                      :class="[
                        'absolute top-0.5 left-0.5 w-4 h-4 bg-white rounded-full shadow-sm transition transform duration-200',
                        form.allowLate ? 'translate-x-5' : 'translate-x-0',
                      ]"
                    ></div>
                  </div>
                </label>
              </div>
            </div>
          </section>

          <!-- Status Card -->
          <div class="rounded-[18px] border border-[#FED7AA] bg-[#FFF7ED] p-5">
            <h3
              class="mb-3 text-xs font-bold uppercase tracking-wider text-[#EA580C]"
            >
              Status Publikasi
            </h3>
            <p class="text-[11px] leading-relaxed text-[#9A3412]">
              Konten ini akan langsung tersedia bagi siswa yang terdaftar di
              kelas
              <strong>{{ subject?.className }}</strong> segera setelah
              diterbitkan.
            </p>
          </div>
        </aside>
      </div>
    </div>
  </main>
</template>
