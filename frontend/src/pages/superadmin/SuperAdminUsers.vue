<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import {
  PhArrowClockwise,
  PhEnvelopeSimple,
  PhIdentificationBadge,
  PhMagnifyingGlass,
  PhShieldCheck,
  PhUsers,
} from "@phosphor-icons/vue";
import { getAdminUsers } from "../../services/adminUser";
import type { AdminUserItem } from "../../types/adminUser";

const users = ref<AdminUserItem[]>([]);
const isLoading = ref(false);
const errorMessage = ref("");
const searchQuery = ref("");

const filteredUsers = computed(() => {
  const query = searchQuery.value.trim().toLowerCase();
  if (!query) return users.value;

  return users.value.filter((user) =>
    [user.fullName, user.email]
      .filter(Boolean)
      .some((value) => value.toLowerCase().includes(query)),
  );
});

const activeCount = computed(
  () => users.value.filter((user) => user.isActive).length,
);

const inactiveCount = computed(() => users.value.length - activeCount.value);

function getApiErrorMessage(error: unknown, fallback: string) {
  if (typeof error === "object" && error !== null && "response" in error) {
    const response = (
      error as {
        response?: { data?: { error?: unknown; message?: unknown } | string };
      }
    ).response;
    if (typeof response?.data === "string") return response.data;
    if (typeof response?.data?.error === "string") return response.data.error;
    if (typeof response?.data?.message === "string")
      return response.data.message;
  }

  return fallback;
}

async function loadUsers() {
  isLoading.value = true;
  errorMessage.value = "";

  try {
    const response = await getAdminUsers({ page: 1, limit: 100 });
    users.value = response.data ?? [];
  } catch (error) {
    users.value = [];
    errorMessage.value = getApiErrorMessage(
      error,
      "Daftar akun global belum bisa dimuat.",
    );
  } finally {
    isLoading.value = false;
  }
}

onMounted(() => {
  loadUsers();
});
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-[#f8f7f4]">
    <header class="border-b border-[#ebe7df] bg-white">
      <div
        class="flex min-w-0 flex-col gap-3 px-5 py-5 sm:px-6 lg:flex-row lg:items-end lg:justify-between lg:px-8"
      >
        <div class="min-w-0">
          <p
            class="text-xs font-semibold uppercase tracking-[0.18em] text-[#ea580c]"
          >
            Super Admin
          </p>
          <h1 class="mt-2 text-2xl font-semibold text-[#171322] sm:text-3xl">
            Akun Global
          </h1>
          <p class="mt-2 max-w-3xl text-sm leading-6 text-[#6b7280]">
            Lihat akun login EduVerse lintas sekolah. Akses sekolah dan peran
            tidak dikelola dari halaman ini.
          </p>
        </div>
        <button
          type="button"
          class="inline-flex w-full items-center justify-center gap-2 rounded-lg border border-[#e5e7eb] bg-white px-4 py-2.5 text-sm font-semibold text-[#171322] transition hover:bg-[#fafafa] disabled:cursor-not-allowed disabled:opacity-60 sm:w-auto"
          :disabled="isLoading"
          @click="loadUsers"
        >
          <PhArrowClockwise :size="16" weight="bold" />
          Muat ulang
        </button>
      </div>
    </header>

    <section
      class="grid w-full max-w-none gap-6 px-5 py-6 sm:px-6 lg:px-8 xl:grid-cols-[minmax(0,1fr)_340px]"
    >
      <div class="flex min-w-0 flex-col gap-6">
        <section class="grid gap-3 sm:grid-cols-3">
          <article class="rounded-xl border border-[#ebe7df] bg-white p-4 shadow-sm">
            <p class="text-xs font-medium text-[#6b7280]">Total akun</p>
            <p class="mt-2 text-2xl font-semibold text-[#171322]">
              {{ users.length }}
            </p>
          </article>
          <article class="rounded-xl border border-[#ebe7df] bg-white p-4 shadow-sm">
            <p class="text-xs font-medium text-[#6b7280]">Aktif</p>
            <p class="mt-2 text-2xl font-semibold text-[#027a48]">
              {{ activeCount }}
            </p>
          </article>
          <article class="rounded-xl border border-[#ebe7df] bg-white p-4 shadow-sm">
            <p class="text-xs font-medium text-[#6b7280]">Nonaktif</p>
            <p class="mt-2 text-2xl font-semibold text-[#b45309]">
              {{ inactiveCount }}
            </p>
          </article>
        </section>

        <section
          class="rounded-xl border border-[#ebe7df] bg-white p-5 shadow-sm"
        >
          <div
            class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between"
          >
            <div class="min-w-0">
              <p
                class="text-xs font-semibold uppercase tracking-[0.16em] text-[#ea580c]"
              >
                Daftar akun login
              </p>
              <h2 class="mt-2 text-xl font-semibold text-[#171322]">
                Pengguna EduVerse
              </h2>
              <p class="mt-2 text-sm leading-6 text-[#6b7280]">
                Akun Global adalah identitas login. Keanggotaan sekolah dan
                peran pengguna akan dikelola melalui alur terpisah.
              </p>
            </div>
            <label class="relative block w-full lg:max-w-xs">
              <PhMagnifyingGlass
                :size="17"
                class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-[#9ca3af]"
              />
              <input
                v-model="searchQuery"
                type="search"
                placeholder="Cari nama atau email..."
                class="w-full rounded-lg border border-[#e5e7eb] bg-white py-2.5 pl-10 pr-3 text-sm text-[#171322] outline-none transition placeholder:text-[#9ca3af] focus:border-[#ea580c] focus:ring-2 focus:ring-[#fed7aa]"
              />
            </label>
          </div>

          <div class="mt-5 space-y-3">
            <div
              v-if="isLoading"
              class="rounded-lg border border-[#e5e7eb] bg-[#fafafa] px-4 py-5 text-sm text-[#6b7280]"
            >
              Memuat akun global...
            </div>

            <div
              v-else-if="errorMessage"
              class="rounded-lg border border-[#fecaca] bg-[#fff8f6] px-4 py-4"
            >
              <p class="text-sm leading-6 text-[#a8665d]">{{ errorMessage }}</p>
              <button
                type="button"
                class="mt-3 inline-flex items-center gap-2 rounded-lg border border-[#fecaca] bg-white px-3 py-2 text-sm font-semibold text-[#a8665d] transition hover:bg-[#fff8f6]"
                @click="loadUsers"
              >
                Coba lagi
              </button>
            </div>

            <div
              v-else-if="users.length === 0"
              class="rounded-lg border border-dashed border-[#d1d5db] bg-[#fafafa] px-4 py-8 text-sm leading-6 text-[#6b7280]"
            >
              Belum ada akun global yang bisa ditampilkan.
            </div>

            <div
              v-else-if="filteredUsers.length === 0"
              class="rounded-lg border border-dashed border-[#d1d5db] bg-[#fafafa] px-4 py-8 text-sm leading-6 text-[#6b7280]"
            >
              Tidak ada akun yang cocok dengan pencarian.
            </div>

            <template v-else>
              <article
                v-for="user in filteredUsers"
                :key="user.userId"
                class="rounded-xl border border-[#ebe7df] bg-[#fcfbf8] p-4"
              >
                <div
                  class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between"
                >
                  <div class="min-w-0">
                    <div class="flex min-w-0 flex-wrap items-center gap-2">
                      <h3 class="truncate text-base font-semibold text-[#171322]">
                        {{ user.fullName }}
                      </h3>
                      <span
                        class="rounded-full px-2.5 py-1 text-xs font-semibold"
                        :class="
                          user.isActive
                            ? 'bg-[#ecfdf3] text-[#027a48]'
                            : 'bg-[#fff7ed] text-[#b45309]'
                        "
                      >
                        {{ user.isActive ? "Aktif" : "Nonaktif" }}
                      </span>
                    </div>
                    <p
                      class="mt-3 flex min-w-0 items-center gap-2 text-sm text-[#6b7280]"
                    >
                      <PhEnvelopeSimple :size="16" class="shrink-0" />
                      <span class="truncate">{{ user.email }}</span>
                    </p>
                  </div>

                  <p class="shrink-0 text-xs leading-5 text-[#9ca3af]">
                    Dibuat {{ user.createdAt }}
                  </p>
                </div>
              </article>
            </template>
          </div>
        </section>
      </div>

      <aside class="min-w-0">
        <section
          class="rounded-xl border border-[#ebe7df] bg-white p-5 shadow-sm xl:sticky xl:top-6"
        >
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <p
                class="text-xs font-semibold uppercase tracking-[0.16em] text-[#ea580c]"
              >
                Informasi akses
              </p>
              <h2 class="mt-2 text-xl font-semibold text-[#171322]">
                Identitas global saja
              </h2>
              <p class="mt-2 text-sm leading-6 text-[#6b7280]">
                Halaman ini bersifat read-only dan hanya menampilkan akun login
                EduVerse.
              </p>
            </div>
            <span
              class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-[#fff4ee] text-[#ea580c]"
            >
              <PhUsers :size="22" weight="duotone" />
            </span>
          </div>

          <div class="mt-5 space-y-3">
            <article class="rounded-lg border border-[#ebe7df] bg-[#fcfbf8] p-4">
              <div class="flex gap-3">
                <PhIdentificationBadge
                  :size="20"
                  class="mt-0.5 shrink-0 text-[#ea580c]"
                  weight="duotone"
                />
                <div>
                  <p class="text-sm font-semibold text-[#171322]">
                    Akun Global
                  </p>
                  <p class="mt-1 text-xs leading-5 text-[#6b7280]">
                    Identitas login pengguna EduVerse lintas sekolah.
                  </p>
                </div>
              </div>
            </article>

            <article class="rounded-lg border border-[#ebe7df] bg-[#fcfbf8] p-4">
              <div class="flex gap-3">
                <PhShieldCheck
                  :size="20"
                  class="mt-0.5 shrink-0 text-[#4f46e5]"
                  weight="duotone"
                />
                <div>
                  <p class="text-sm font-semibold text-[#171322]">
                    Akses sekolah dan peran
                  </p>
                  <p class="mt-1 text-xs leading-5 text-[#6b7280]">
                    Keanggotaan sekolah dan peran pengguna tidak diubah dari
                    halaman ini.
                  </p>
                </div>
              </div>
            </article>

            <article class="rounded-lg border border-[#fed7aa] bg-[#fff7ed] p-4">
              <p class="text-xs leading-5 text-[#9a3412]">
                Operasional akademik seperti kelas, materi, tugas, dan nilai
                tetap dikelola oleh Admin Sekolah.
              </p>
            </article>
          </div>
        </section>
      </aside>
    </section>
  </main>
</template>
