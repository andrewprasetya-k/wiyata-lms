<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import {
  PhArrowRight,
  PhBookOpen,
  PhChalkboardTeacher,
  PhClipboardText,
  PhMegaphone,
  PhPlusCircle,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import { getTeacherDashboard } from "../../services/teacherDashboard";
import type { MembershipInfo } from "../../types/auth";
import type { TeacherDashboardSummary } from "../../types/teacherDashboard";
import { getSubjectColor } from "../../utils/color";

const auth = useAuthStore();

const loading = ref(false);
const errorMessage = ref("");
const summary = ref<TeacherDashboardSummary | null>(null);

const activeMembership = computed<MembershipInfo | undefined>(() => {
  const activeSchoolMembership = auth.memberships.find(
    (membership) => membership.school.id === auth.activeSchoolId,
  );
  const defaultMembership = auth.memberships.find(
    (membership) => membership.isDefault,
  );
  return activeSchoolMembership ?? defaultMembership ?? auth.memberships[0];
});

const schoolUserId = computed(() => {
  return (
    activeMembership.value?.schoolUserId ??
    auth.defaultContext?.schoolUserId ??
    ""
  );
});

const activeSchoolName = computed(
  () => activeMembership.value?.school.name ?? "Sekolah aktif",
);
const teacherName = computed(() => auth.user?.fullName ?? "Guru");

const stats = computed(() => [
  {
    label: "Review menunggu",
    value: summary.value?.pendingReviews,
    helper: "Submission yang belum dinilai",
    tone: "#e58f86",
  },
  {
    label: "Total siswa",
    value: summary.value?.totalStudents,
    helper: "Siswa unik dari kelas yang diajar",
    tone: "#74bfa5",
  },
  {
    label: "Submission rate",
    value:
      typeof summary.value?.submissionRate === "number"
        ? `${summary.value.submissionRate.toFixed(1)}%`
        : undefined,
    helper: "Rata-rata pengumpulan tugas",
    tone: "#7aa7d9",
  },
]);

async function loadDashboard() {
  if (!schoolUserId.value) {
    summary.value = null;
    errorMessage.value = "";
    return;
  }

  loading.value = true;
  errorMessage.value = "";
  try {
    summary.value = await getTeacherDashboard(schoolUserId.value);
  } catch {
    errorMessage.value =
      "Dashboard guru belum bisa dimuat. Coba lagi beberapa saat.";
  } finally {
    loading.value = false;
  }
}

onMounted(loadDashboard);
</script>

<template>
  <main class="min-h-screen flex-1 px-5 py-5 sm:px-6 lg:px-8">
    <section class="flex w-full max-w-none flex-col gap-5">
      <div
        class="rounded-[22px] bg-[#f0e9dd] px-5 py-5 shadow-sm ring-1 ring-black/5 md:px-6"
      >
        <div
          class="flex flex-col gap-5 lg:flex-row lg:items-end lg:justify-between"
        >
          <div>
            <p class="text-sm font-medium text-[#8a6d3b]">
              {{ activeSchoolName }}
            </p>
            <h1 class="mt-3 text-3xl font-medium text-[#171322] md:text-4xl">
              Selamat mengajar, {{ teacherName }}
            </h1>
            <p class="mt-3 max-w-2xl text-sm leading-6 text-[#6b6475]">
              Dashboard Guru
            </p>
          </div>

          <RouterLink
            to="/teacher/subjects"
            class="inline-flex items-center gap-2 self-start rounded-2xl bg-[#171322] px-4 py-3 text-sm font-medium text-white transition hover:bg-[#2f2b3a] lg:self-auto"
          >
            <PhPlusCircle :size="18" weight="duotone" />
            Pilih subject
          </RouterLink>
        </div>
      </div>

      <div
        v-if="!schoolUserId"
        class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5"
      >
        <div class="flex items-start gap-3">
          <PhWarningCircle
            :size="24"
            class="mt-0.5 text-[#e58f86]"
            weight="duotone"
          />
          <div>
            <h2 class="text-lg font-medium text-[#171322]">
              Konteks guru belum tersedia
            </h2>
            <p class="mt-2 text-sm leading-6 text-[#6b6475]">
              Dashboard guru membutuhkan schoolUserId dari membership login.
              Pastikan akun memiliki role teacher pada school aktif.
            </p>
          </div>
        </div>
      </div>

      <template v-else>
        <div class="grid gap-4 md:grid-cols-3">
          <article
            v-for="item in stats"
            :key="item.label"
            class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5"
          >
            <div
              class="mb-5 flex h-11 w-11 items-center justify-center rounded-2xl text-white"
              :style="{ backgroundColor: item.tone }"
            >
              <PhClipboardText :size="22" weight="duotone" />
            </div>
            <p class="text-sm text-[#7b7486]">{{ item.label }}</p>
            <p class="mt-2 text-3xl font-medium text-[#171322]">
              {{ item.value ?? (loading ? "..." : "-") }}
            </p>
            <p class="mt-2 text-sm leading-5 text-[#8a8494]">
              {{ item.helper }}
            </p>
          </article>
        </div>

        <div
          v-if="loading"
          class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5"
        >
          <p class="text-sm text-[#6b6475]">Memuat data dashboard guru...</p>
        </div>

        <div
          v-else-if="errorMessage"
          class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5"
        >
          <div
            class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"
          >
            <div>
              <h2 class="text-lg font-medium text-[#171322]">
                Dashboard gagal dimuat
              </h2>
              <p class="mt-2 text-sm leading-6 text-[#6b6475]">
                {{ errorMessage }}
              </p>
            </div>
            <button
              class="rounded-2xl bg-[#171322] px-4 py-3 text-sm font-medium text-white"
              type="button"
              @click="loadDashboard"
            >
              Coba lagi
            </button>
          </div>
        </div>

        <div class="grid gap-5 xl:grid-cols-[1.35fr_0.75fr]">
          <section
            class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5"
          >
            <div
              class="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between"
            >
              <div>
                <p class="text-sm font-medium text-[#7b61a8]">
                  Subject yang diajar
                </p>
                <h2 class="mt-2 text-2xl font-medium text-[#171322]">
                  Ringkasan kelas dan subject
                </h2>
              </div>
              <RouterLink
                to="/teacher/subjects"
                class="inline-flex items-center gap-2 text-sm font-medium text-[#4f46e5]"
              >
                Buka teaching workspace
                <PhArrowRight :size="16" />
              </RouterLink>
            </div>

            <div
              v-if="summary?.classPerformance?.length"
              class="mt-6 grid gap-4 md:grid-cols-2"
            >
              <article
                v-for="item in summary.classPerformance"
                :key="`${item.classId}-${item.subjectName}`"
                class="rounded-[18px] bg-[#faf8f4] p-5 ring-1 ring-black/5"
              >
                <div
                  class="mb-5 flex h-12 w-12 items-center justify-center rounded-2xl text-white"
                  :style="{
                    backgroundColor: getSubjectColor(
                      `${item.classId}-${item.subjectName}`,
                    ),
                  }"
                >
                  <PhBookOpen :size="24" weight="duotone" />
                </div>
                <p class="text-sm font-medium text-[#171322]">
                  {{ item.subjectName }}
                </p>
                <h3 class="mt-1 text-xl font-medium text-[#2f2b3a]">
                  {{ item.className }}
                </h3>
                <dl class="mt-5 grid grid-cols-3 gap-3 text-sm">
                  <div>
                    <dt class="text-[#8a8494]">Siswa</dt>
                    <dd class="mt-1 font-medium text-[#171322]">
                      {{ item.totalStudents }}
                    </dd>
                  </div>
                  <div>
                    <dt class="text-[#8a8494]">Submit</dt>
                    <dd class="mt-1 font-medium text-[#171322]">
                      {{ item.submissionRate.toFixed(1) }}%
                    </dd>
                  </div>
                  <div>
                    <dt class="text-[#8a8494]">Rata-rata</dt>
                    <dd class="mt-1 font-medium text-[#171322]">
                      {{ item.averageScore.toFixed(1) }}
                    </dd>
                  </div>
                </dl>
              </article>
            </div>

            <div v-else class="mt-5 rounded-[18px] bg-[#faf8f4] p-5">
              <h3 class="text-lg font-medium text-[#171322]">
                Belum ada ringkasan subject
              </h3>
              <p class="mt-2 text-sm leading-6 text-[#6b6475]">
                Ringkasan dashboard belum memiliki data untuk school aktif. Buka
                teaching workspace untuk melihat daftar subject class yang
                diajar dari endpoint current teacher.
              </p>
            </div>
          </section>

          <aside class="flex flex-col gap-4">
            <article
              class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5"
            >
              <div class="flex items-center gap-3">
                <div
                  class="flex h-11 w-11 items-center justify-center rounded-2xl bg-[#eef2ff] text-[#4f46e5]"
                >
                  <PhChalkboardTeacher :size="22" weight="duotone" />
                </div>
                <div>
                  <p class="text-sm font-medium text-[#171322]">
                    Pending submissions
                  </p>
                  <p class="text-sm text-[#7b7486]">
                    Review detail akan tersedia di halaman submissions.
                  </p>
                </div>
              </div>
            </article>

            <article
              class="rounded-[22px] bg-white p-5 shadow-sm ring-1 ring-black/5"
            >
              <div class="flex items-center gap-3">
                <div
                  class="flex h-11 w-11 items-center justify-center rounded-2xl bg-[#f7eadf] text-[#b86845]"
                >
                  <PhMegaphone :size="22" weight="duotone" />
                </div>
                <div>
                  <p class="text-sm font-medium text-[#171322]">Feed kelas</p>
                  <p class="text-sm text-[#7b7486]">
                    Pengumuman class-level akan dibuat dari konteks class yang
                    valid.
                  </p>
                </div>
              </div>
            </article>
          </aside>
        </div>
      </template>
    </section>
  </main>
</template>
