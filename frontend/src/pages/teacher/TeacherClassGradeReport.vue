<script setup lang="ts">
import { computed, ref } from "vue";
import { RouterLink, useRoute } from "vue-router";
import { PhArrowLeft, PhChartBar } from "@phosphor-icons/vue";
import TeacherClassGradeReportTab from "../../components/teacher/TeacherClassGradeReportTab.vue";
import type { ClassGradeReportResponse } from "../../types/teacherGrades";
import { resolveSubjectColor } from "../../utils/color";

const route = useRoute();
const classId = computed(() => String(route.params.classId ?? ""));
const subjectId = computed(() => String(route.params.subjectId ?? ""));

// Populated from the tab component's "loaded" event — reused here only for
// the page header (class/subject title), so there is a single fetch shared
// between header and table.
const report = ref<ClassGradeReportResponse | null>(null);
const subjectAccentColor = computed(() =>
  resolveSubjectColor({
    subjectId: report.value?.subject.subjectId,
    subjectName: report.value?.subject.subjectName,
    subjectCode: report.value?.subject.subjectCode,
  }),
);
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
              {{ report?.subject.subjectName ?? "Nilai Kelas" }}
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

    <section class="mx-auto max-w-screen px-5 py-5 sm:px-6 lg:px-8 lg:py-6">
      <TeacherClassGradeReportTab
        :class-id="classId"
        :subject-id="subjectId"
        @loaded="report = $event"
      />
    </section>
  </main>
</template>
