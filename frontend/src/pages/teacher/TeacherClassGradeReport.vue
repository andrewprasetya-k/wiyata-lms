<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink, useRoute } from "vue-router";
import {
  PhArrowLeft,
  PhChartBar,
  PhUsersThree,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { getClassGradeReport } from "../../services/teacherGrades";
import type { ClassGradeReportResponse } from "../../types/teacherGrades";
import { resolveSubjectColor } from "../../utils/color";

const route = useRoute();
const classId = computed(() => String(route.params.classId ?? ""));
const subjectId = computed(() => String(route.params.subjectId ?? ""));

const report = ref<ClassGradeReportResponse | null>(null);
const loading = ref(false);
const errorMessage = ref("");

const students = computed(() => report.value?.students ?? []);
const subjectAccentColor = computed(() =>
  resolveSubjectColor({
    subjectId: report.value?.subject.subjectId,
    subjectName: report.value?.subject.subjectName,
    subjectCode: report.value?.subject.subjectCode,
  }),
);

async function loadReport() {
  if (!classId.value || !subjectId.value) {
    errorMessage.value = "Kelas atau mata pelajaran tidak valid.";
    return;
  }

  loading.value = true;
  errorMessage.value = "";
  report.value = null;

  try {
    report.value = await getClassGradeReport(classId.value, subjectId.value);
  } catch {
    errorMessage.value =
      "Nilai kelas belum bisa dimuat. Coba lagi beberapa saat.";
  } finally {
    loading.value = false;
  }
}

function formatScore(value?: number | null) {
  if (value === null || value === undefined) return "Belum tersedia";
  return new Intl.NumberFormat("id-ID", { maximumFractionDigits: 2 }).format(
    value,
  );
}

onMounted(loadReport);
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 bg-background">
    <header class="border-b border-border bg-surface">
      <div class="px-5 py-5 sm:px-6 lg:px-8">
        <div class="flex min-w-0 items-center gap-2 text-xs text-muted">
          <RouterLink
            to="/teacher/subjects"
            class="inline-flex shrink-0 items-center gap-1.5 transition hover:text-brand"
          >
            <PhArrowLeft :size="15" />
            Mata pelajaran
          </RouterLink>
          <span class="text-border-strong">/</span>
          <span class="min-w-0 truncate font-medium text-foreground">
            Nilai Kelas
          </span>
        </div>

        <div class="mt-4 flex min-w-0 items-start gap-3">
          <div
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl text-white"
            :style="{ backgroundColor: subjectAccentColor }"
          >
            <PhChartBar :size="21" weight="duotone" />
          </div>
          <div class="min-w-0">
            <h1
              class="truncate text-xl font-semibold text-foreground sm:text-2xl"
            >
              {{
                report?.subject.subjectName ??
                (loading ? "Memuat nilai kelas..." : "Nilai Kelas")
              }}
            </h1>
            <p class="mt-1 truncate text-xs text-muted sm:text-sm">
              {{
                report
                  ? [report.class.classTitle, report.subject.subjectCode]
                      .filter(Boolean)
                      .join(" · ")
                  : "Rekap nilai berbobot seluruh siswa pada mata pelajaran ini."
              }}
            </p>
          </div>
        </div>
      </div>
    </header>

    <section
      class="mx-auto max-w-screen px-5 py-5 sm:px-6 lg:px-8 lg:py-6"
    >
      <section v-if="loading" class="space-y-3">
        <div
          v-for="item in 4"
          :key="item"
          class="h-16 animate-pulse rounded-xl border border-border bg-surface"
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
            <div>
              <h2 class="text-base font-semibold text-foreground">
                Nilai kelas tidak dapat dimuat
              </h2>
              <p class="mt-1 text-sm leading-6 text-muted">
                {{ errorMessage }}
              </p>
              <button
                class="mt-4 rounded-lg bg-brand px-4 py-2 text-sm font-medium text-white transition hover:bg-brand-hover"
                type="button"
                @click="loadReport"
              >
                Coba lagi
              </button>
            </div>
          </div>
        </article>
      </section>

      <section
        v-else-if="students.length === 0"
        class="flex min-h-[55vh] items-center justify-center"
      >
        <article
          class="w-full max-w-xl rounded-xl border border-border bg-surface p-8 text-center"
        >
          <div
            class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-brand-soft text-brand"
          >
            <PhUsersThree class="h-6 w-6" weight="duotone" />
          </div>
          <h2 class="mt-3 text-base font-semibold text-foreground">
            Belum ada siswa
          </h2>
          <p class="mx-auto mt-2 max-w-md text-sm leading-6 text-muted">
            Belum ada siswa yang terdaftar pada kelas ini untuk mata pelajaran
            tersebut.
          </p>
        </article>
      </section>

      <section
        v-else
        class="overflow-x-auto rounded-xl border border-border bg-surface shadow-sm"
      >
        <table class="w-full min-w-140 text-left text-sm">
          <thead
            class="border-b border-border bg-surface-subtle text-xs uppercase tracking-wide text-muted"
          >
            <tr>
              <th class="px-4 py-3 font-medium">Nama</th>
              <th class="px-4 py-3 font-medium">Email</th>
              <th class="px-4 py-3 font-medium">Final Grade</th>
              <th class="px-4 py-3 font-medium">Letter Grade</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-border">
            <tr v-for="student in students" :key="student.studentId">
              <td class="px-4 py-3 font-medium text-foreground">
                <RouterLink
                  :to="{
                    name: 'teacher-student-grade-detail',
                    params: {
                      classId: classId,
                      subjectId: subjectId,
                      studentId: student.studentId,
                    },
                  }"
                  class="transition hover:text-brand"
                >
                  {{ student.studentName }}
                </RouterLink>
              </td>
              <td class="px-4 py-3 text-muted">{{ student.studentEmail }}</td>
              <td class="px-4 py-3 text-foreground">
                {{ formatScore(student.finalGrade) }}
              </td>
              <td class="px-4 py-3">
                <span
                  class="rounded-full bg-brand-soft px-3 py-1.5 text-xs font-medium text-brand"
                >
                  {{ student.letterGrade }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </section>
    </section>
  </main>
</template>
