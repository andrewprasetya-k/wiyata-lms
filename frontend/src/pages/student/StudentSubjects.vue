<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import {
  PhArrowRight,
  PhBooks,
  PhMegaphone,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { getSubjectClassesByClass } from "../../services/classWorkspace";
import { useActiveClassStore } from "../../stores/activeClass";
import { useAuthStore } from "../../stores/auth";
import type { SubjectClassItem } from "../../types/classWorkspace";
import { resolveSubjectColor } from "../../utils/color";

const auth = useAuthStore();
const activeClassStore = useActiveClassStore();
const router = useRouter();

const subjects = ref<SubjectClassItem[]>([]);
const isLoading = ref(true);
const errorMessage = ref("");

const schoolUserId = computed(() => auth.activeSchoolUserId);
const activeClass = computed(() => activeClassStore.activeClass);

async function loadSubjects() {
  if (!schoolUserId.value) {
    isLoading.value = false;
    errorMessage.value =
      "Konteks sekolah belum tersedia. Silakan login ulang atau pilih sekolah aktif terlebih dahulu.";
    return;
  }

  isLoading.value = true;
  errorMessage.value = "";

  try {
    await activeClassStore.loadClasses(schoolUserId.value);
    if (activeClassStore.errorMessage) {
      errorMessage.value = activeClassStore.errorMessage;
      subjects.value = [];
      return;
    }

    if (!activeClassStore.activeClassId) {
      subjects.value = [];
      return;
    }

    const data = await getSubjectClassesByClass(activeClassStore.activeClassId);
    subjects.value = data.subjects;
  } catch {
    errorMessage.value =
      activeClassStore.errorMessage ||
      "Mata pelajaran belum bisa dimuat. Periksa koneksi atau coba lagi nanti.";
  } finally {
    isLoading.value = false;
  }
}

function openSubject(subject: SubjectClassItem) {
  router.push({
    path: `/student/subjects/${subject.subjectClassId}`,
    query: subject.subjectName ? { title: subject.subjectName } : undefined,
  });
}

onMounted(loadSubjects);
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-background">
    <header class="border-b border-border bg-surface">
      <div class="flex min-w-0 flex-col gap-1 px-5 py-5 sm:px-6 lg:px-8">
        <h1 class="text-2xl font-semibold text-foreground sm:text-3xl">
          Mata pelajaran
        </h1>
        <p class="mt-2 max-w-3xl text-sm leading-6 text-muted">
          Buka materi, tugas, dan catatan dari kelas aktifmu.
        </p>
        <RouterLink
          class="inline-flex w-full shrink-0 items-center justify-center gap-2 rounded-lg border border-border bg-surface px-3 py-2 text-xs font-medium text-foreground-secondary transition hover:border-brand hover:text-brand sm:w-auto"
          to="/student/feed"
        >
          <PhMegaphone :size="16" />
          Feed kelas
        </RouterLink>
      </div>

      <!-- <div class="px-5 py-3 sm:px-6 lg:px-8">
        <div
          class="flex min-w-0 flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
        >
          <div class="flex min-w-0 flex-col gap-2 sm:flex-row sm:items-center">
            <span class="shrink-0 text-[11px] text-muted">
              Kelas aktif
            </span>
            <div class="flex min-w-0 max-w-full items-center gap-2">
              <div
                class="flex min-w-0 max-w-full items-center gap-2 rounded-lg border border-border bg-surface-subtle px-3 py-2"
              >
                <div
                  class="flex h-6 w-6 shrink-0 items-center justify-center rounded-md bg-brand text-[10px] font-medium text-white"
                >
                  {{
                    activeClass?.classTitle?.slice(0, 2).toUpperCase() || "EV"
                  }}
                </div>
                <div class="min-w-0">
                  <p
                    class="max-w-[12rem] truncate text-xs font-medium text-foreground sm:max-w-[16rem]"
                  >
                    {{ activeClass?.classTitle || "Belum ada kelas aktif" }}
                  </p>
                  <p
                    class="max-w-[12rem] truncate text-[10px] text-muted sm:max-w-[16rem]"
                  >
                    {{ schoolName }}
                  </p>
                </div>
                <PhCaretDown :size="13" class="shrink-0 text-muted" />
              </div>
              <select
                v-if="classes.length > 1"
                class="min-w-0 max-w-full rounded-lg border border-border bg-surface px-3 py-2 text-xs text-foreground outline-none transition focus:border-brand"
                :value="activeClassStore.activeClassId ?? ''"
                aria-label="Pilih kelas aktif"
                @change="
                  changeActiveClass(($event.target as HTMLSelectElement).value)
                "
              >
                <option
                  v-for="item in classes"
                  :key="item.enrollmentId"
                  :value="item.classId"
                >
                  {{ item.classTitle || "Kelas" }}
                </option>
              </select>
            </div>
          </div>
        </div>
      </div> -->
    </header>

    <section
      class="mx-auto max-w-screen min-w-0 px-5 py-5 sm:px-6 lg:px-8 lg:py-6"
    >
      <section
        v-if="isLoading || activeClassStore.isLoading"
        class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3"
      >
        <div
          v-for="item in 6"
          :key="item"
          class="h-44 animate-pulse rounded-xl border border-border bg-surface"
        />
      </section>

      <section
        v-else-if="errorMessage"
        class="flex min-h-[55vh] items-center justify-center"
      >
        <article
          class="w-full max-w-xl rounded-xl border border-danger-line bg-danger-soft p-6"
        >
          <div class="flex items-start gap-3">
            <div
              class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-danger-soft text-danger"
            >
              <PhWarningCircle :size="22" weight="duotone" />
            </div>
            <div class="min-w-0">
              <h2 class="text-base font-semibold text-foreground">
                Mata pelajaran tidak dapat dimuat
              </h2>
              <p class="mt-1 text-sm leading-6 text-muted">
                {{ errorMessage }}
              </p>
              <button
                class="mt-4 rounded-lg bg-brand px-4 py-2 text-sm font-medium text-white transition hover:bg-brand-hover"
                type="button"
                @click="loadSubjects"
              >
                Coba lagi
              </button>
            </div>
          </div>
        </article>
      </section>

      <section
        v-else-if="!activeClass"
        class="flex min-h-[55vh] items-center justify-center"
      >
        <article
          class="w-full max-w-xl rounded-xl border border-border bg-surface p-8 text-center"
        >
          <div
            class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-brand-soft text-brand"
          >
            <PhBooks class="h-6 w-6" weight="duotone" />
          </div>
          <h2 class="mt-3 text-base font-semibold text-foreground">
            Belum ada kelas aktif
          </h2>
          <p class="mx-auto mt-2 max-w-md text-sm leading-6 text-muted">
            Mata pelajaran akan tampil setelah akunmu terdaftar pada kelas di
            sekolah aktif.
          </p>
        </article>
      </section>

      <section
        v-else-if="subjects.length === 0"
        class="flex min-h-[55vh] items-center justify-center"
      >
        <article
          class="w-full max-w-xl rounded-xl border border-border bg-surface p-8 text-center"
        >
          <div
            class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-brand-soft text-brand"
          >
            <PhBooks class="h-6 w-6" weight="duotone" />
          </div>
          <h2 class="mt-3 text-base font-semibold text-foreground">
            Mata pelajaran belum tersedia
          </h2>
          <p class="mx-auto mt-2 max-w-md text-sm leading-6 text-muted">
            Kelas aktif belum memiliki mata pelajaran yang bisa ditampilkan.
          </p>
        </article>
      </section>

      <section v-else>
        <div class="mb-4 flex min-w-0 items-end justify-between gap-3">
          <div class="min-w-0">
            <h2 class="text-sm font-semibold text-foreground">
              Daftar mata pelajaran
            </h2>
            <p class="mt-1 truncate text-xs text-muted sm:text-sm">
              {{ subjects.length }} mata pelajaran tersedia
              <span v-if="activeClass?.classTitle">
                · {{ activeClass.classTitle }}</span
              >
            </p>
          </div>
        </div>

        <div class="grid min-w-0 gap-3 sm:grid-cols-2 xl:grid-cols-3">
          <article
            v-for="subject in subjects"
            :key="subject.subjectClassId"
            class="group min-w-0 overflow-hidden rounded-xl border border-border bg-surface transition hover:-translate-y-0.5 hover:shadow-[0_18px_40px_rgba(66,55,40,0.08)]"
          >
            <button
              class="block w-full min-w-0 text-left"
              type="button"
              @click="openSubject(subject)"
            >
              <div
                class="relative flex h-24 flex-col justify-end px-4 pb-4 text-white"
                :style="{
                  backgroundColor: resolveSubjectColor(subject),
                }"
              >
                <span
                  class="line-clamp-2 wrap-break-word text-base font-medium"
                >
                  {{
                    subject.subjectName ||
                    subject.subjectCode ||
                    "Mata pelajaran"
                  }}
                </span>
                <p class="mt-0.5 truncate text-xs text-white/80">
                  {{ subject.teacherName || "Guru belum tersedia" }}
                </p>
              </div>
              <div class="space-y-3 px-4 py-4">
                <div class="flex items-center justify-between gap-3">
                  <div class="min-w-0">
                    <p class="text-xs text-[#9a95a3]">Kode Mapel</p>
                    <p class="mt-1 truncate text-sm font-medium text-foreground">
                      {{ subject.subjectCode || "Kode belum tersedia" }}
                    </p>
                  </div>
                  <PhArrowRight
                    :size="18"
                    class="text-muted transition group-hover:translate-x-0.5 group-hover:text-brand"
                  />
                </div>
                <div
                  class="flex flex-wrap gap-2 border-t border-[#f3f1ec] pt-3"
                >
                  <span
                    class="rounded-full bg-surface-strong px-2 py-1 text-[11px] text-muted"
                  >
                    Materi
                  </span>
                  <span
                    class="rounded-full bg-warning-soft px-2 py-1 text-[11px] text-warning"
                  >
                    Tugas
                  </span>
                  <span
                    class="rounded-full bg-brand-soft px-2 py-1 text-[11px] text-brand"
                  >
                    Catatan
                  </span>
                </div>
              </div>
            </button>
          </article>
        </div>
      </section>

      <!-- <section v-if="classes.length > 1" class="mt-6 rounded-2xl border border-border bg-surface p-4">
        <p class="text-xs font-medium uppercase tracking-[0.08em] text-[#9a95a3]">Class context</p>
        <div class="mt-3 grid gap-2 sm:grid-cols-2 xl:grid-cols-3">
          <div
            v-for="item in classes"
            :key="item.enrollmentId"
            class="flex items-center gap-3 rounded-xl px-3 py-2"
            :class="item.classId === activeClass?.classId ? 'bg-brand-soft' : 'bg-surface-subtle'"
          >
            <div class="h-2 w-2 rounded-full bg-brand" />
            <div class="min-w-0 flex-1">
              <p class="truncate text-xs font-medium text-foreground">{{ item.classTitle || 'Kelas' }}</p>
              <p class="text-[11px] text-muted">Dipakai sebagai konteks akademik aktif</p>
            </div>
            <PhCheck v-if="item.classId === activeClass?.classId" :size="15" class="text-brand" />
          </div>
        </div>
      </section> -->
    </section>
  </main>
</template>
