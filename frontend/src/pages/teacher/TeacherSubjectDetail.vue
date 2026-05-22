<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink, useRoute } from "vue-router";
import {
  PhArrowLeft,
  PhBookOpen,
  PhChatCircle,
  PhClipboardText,
  PhFileText,
  PhUsersThree,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { getMyTeachingSubjectClassById } from "../../services/teacherSubjects";
import type { TeacherSubjectClass } from "../../types/teacherSubjects";
import { getSubjectColor } from "../../utils/color";

const route = useRoute();
const subjectClassId = computed(() =>
  String(route.params.subjectClassId ?? ""),
);
const subject = ref<TeacherSubjectClass | null>(null);
const loading = ref(false);
const errorMessage = ref("");

async function loadSubject() {
  loading.value = true;
  errorMessage.value = "";
  try {
    subject.value = await getMyTeachingSubjectClassById(subjectClassId.value);
  } catch {
    errorMessage.value =
      "Subject workspace belum bisa dimuat. Coba lagi beberapa saat.";
  } finally {
    loading.value = false;
  }
}

onMounted(loadSubject);
</script>

<template>
  <main class="min-h-screen flex-1 px-5 py-8 md:px-8 lg:px-10">
    <section class="mx-auto flex max-w-6xl flex-col gap-6">
      <RouterLink
        to="/teacher/subjects"
        class="inline-flex items-center gap-2 self-start text-sm font-medium text-[#6b6475] transition hover:text-[#171322]"
      >
        <PhArrowLeft :size="18" />
        Kembali ke subjects
      </RouterLink>

      <header
        class="rounded-4xl bg-white p-6 shadow-sm ring-1 ring-black/5 md:p-8"
      >
        <div
          class="mb-5 flex h-14 w-14 items-center justify-center rounded-2xl text-white"
          :style="{ backgroundColor: getSubjectColor(subjectClassId) }"
        >
          <PhBookOpen :size="28" weight="duotone" />
        </div>
        <p class="text-sm font-medium text-[#7b61a8]">Subject workspace</p>
        <h1 class="mt-3 text-3xl font-medium text-[#171322]">
          {{
            subject?.subjectName ??
            (loading ? "Memuat subject..." : "Workspace subject")
          }}
        </h1>
        <p class="mt-3 max-w-2xl text-sm leading-6 text-[#6b6475]">
          <span v-if="subject">
            {{ subject.className }} menjadi konteks class untuk subject ini.
            Material dan tugas berikutnya akan dibuat di level subject class.
          </span>
          <span v-else>
            Detail subject class mengambil data dari endpoint current teacher
            agar guru hanya melihat subject yang dia ampu.
          </span>
        </p>
      </header>

      <section
        v-if="errorMessage"
        class="rounded-[28px] bg-white p-6 shadow-sm ring-1 ring-black/5"
      >
        <div class="flex items-start gap-3">
          <PhWarningCircle
            :size="24"
            class="mt-0.5 text-[#e58f86]"
            weight="duotone"
          />
          <div>
            <h2 class="text-lg font-medium text-[#171322]">
              Gagal memuat workspace
            </h2>
            <p class="mt-2 text-sm leading-6 text-[#6b6475]">
              {{ errorMessage }}
            </p>
          </div>
        </div>
      </section>

      <section
        v-else-if="!loading && !subject"
        class="rounded-[28px] bg-white p-6 shadow-sm ring-1 ring-black/5"
      >
        <h2 class="text-lg font-medium text-[#171322]">
          Subject tidak ditemukan
        </h2>
        <p class="mt-2 text-sm leading-6 text-[#6b6475]">
          Subject class ini tidak tersedia untuk akun guru pada school aktif.
        </p>
      </section>

      <section v-if="subject" class="grid gap-4 md:grid-cols-4">
        <article
          class="rounded-3xl bg-white p-5 shadow-sm ring-1 ring-black/5"
        >
          <PhUsersThree :size="24" class="text-[#74bfa5]" weight="duotone" />
          <p class="mt-4 text-sm text-[#8a8494]">Siswa</p>
          <p class="mt-1 text-2xl font-medium text-[#171322]">
            {{ subject.studentCount }}
          </p>
        </article>
        <article
          class="rounded-3xl bg-white p-5 shadow-sm ring-1 ring-black/5"
        >
          <PhFileText :size="24" class="text-[#7aa7d9]" weight="duotone" />
          <p class="mt-4 text-sm text-[#8a8494]">Materi</p>
          <p class="mt-1 text-2xl font-medium text-[#171322]">
            {{ subject.materialCount }}
          </p>
        </article>
        <article
          class="rounded-3xl bg-white p-5 shadow-sm ring-1 ring-black/5"
        >
          <PhClipboardText :size="24" class="text-[#e58f86]" weight="duotone" />
          <p class="mt-4 text-sm text-[#8a8494]">Tugas</p>
          <p class="mt-1 text-2xl font-medium text-[#171322]">
            {{ subject.assignmentCount }}
          </p>
        </article>
        <article
          class="rounded-3xl bg-white p-5 shadow-sm ring-1 ring-black/5"
        >
          <PhWarningCircle :size="24" class="text-[#b889c9]" weight="duotone" />
          <p class="mt-4 text-sm text-[#8a8494]">Perlu review</p>
          <p class="mt-1 text-2xl font-medium text-[#171322]">
            {{ subject.pendingSubmissions }}
          </p>
        </article>
      </section>

      <section class="grid gap-4 md:grid-cols-3">
        <article
          class="rounded-[28px] bg-white p-5 shadow-sm ring-1 ring-black/5"
        >
          <PhFileText :size="26" class="text-[#7aa7d9]" weight="duotone" />
          <h2 class="mt-4 text-lg font-medium text-[#171322]">Materi</h2>
          <p class="mt-2 text-sm leading-6 text-[#6b6475]">
            Daftar dan pembuatan materi akan memakai subjectClassId saat create
            flow siap.
          </p>
        </article>
        <article
          class="rounded-[28px] bg-white p-5 shadow-sm ring-1 ring-black/5"
        >
          <PhClipboardText :size="26" class="text-[#e58f86]" weight="duotone" />
          <h2 class="mt-4 text-lg font-medium text-[#171322]">Tugas</h2>
          <p class="mt-2 text-sm leading-6 text-[#6b6475]">
            Assignment dan review submission akan ditampilkan di sini tanpa data
            palsu.
          </p>
        </article>
        <article
          class="rounded-[28px] bg-white p-5 shadow-sm ring-1 ring-black/5"
        >
          <PhChatCircle :size="26" class="text-[#74bfa5]" weight="duotone" />
          <h2 class="mt-4 text-lg font-medium text-[#171322]">Chat subject</h2>
          <p class="mt-2 text-sm leading-6 text-[#6b6475]">
            Chat realtime masih fitur masa depan dan tetap placeholder pada
            tahap ini.
          </p>
        </article>
      </section>
    </section>
  </main>
</template>
