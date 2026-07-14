<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { PhUsersThree, PhWarningCircle } from "@phosphor-icons/vue";
import { getClassGradeReport } from "../../services/teacherGrades";
import { useRouter } from "vue-router";
import type { ClassGradeReportResponse } from "../../types/teacherGrades";

const props = defineProps<{
  classId: string;
  subjectId: string;
}>();
const router = useRouter();

const emit = defineEmits<{
  loaded: [report: ClassGradeReportResponse];
}>();

const goToStudent = (studentId: string) => {
  router.push({
    name: "teacher-student-grade-detail",
    params: {
      classId: props.classId,
      subjectId: props.subjectId,
      studentId,
    },
  });
};

const report = ref<ClassGradeReportResponse | null>(null);
const loading = ref(false);
const errorMessage = ref("");

const students = computed(() => report.value?.students ?? []);

async function loadReport() {
  if (!props.classId || !props.subjectId) {
    errorMessage.value = "Kelas atau mata pelajaran tidak valid.";
    return;
  }

  loading.value = true;
  errorMessage.value = "";
  report.value = null;

  try {
    report.value = await getClassGradeReport(props.classId, props.subjectId);
    if (report.value) emit("loaded", report.value);
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

watch(
  () => [props.classId, props.subjectId],
  () => {
    loadReport();
  },
);

onMounted(loadReport);
</script>

<template>
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

  <section v-else class="overflow-x-auto bg-surface">
    <table class="w-full min-w-140 text-left text-sm">
      <thead
        class="bg-surface-subtle text-xs uppercase tracking-wide text-muted"
      >
        <tr>
          <th class="px-4 py-3 font-medium">Nama</th>
          <th class="px-4 py-3 font-medium">Email</th>
          <th class="px-4 py-3 font-medium">Final Grade</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-border">
        <tr
          v-for="student in students"
          :key="student.studentId"
          @click="goToStudent(student.studentId)"
          class="cursor-pointer hover:bg-surface-hover"
        >
          <td class="px-4 py-3 font-medium text-foreground">
            {{ student.studentName }}
          </td>
          <td class="px-4 py-3 text-muted">{{ student.studentEmail }}</td>
          <td class="px-4 py-3 text-foreground">
            {{ formatScore(student.finalGrade) }}
          </td>
        </tr>
      </tbody>
    </table>
  </section>
</template>
