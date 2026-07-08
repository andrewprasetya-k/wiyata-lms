<script setup lang="ts">
import { onMounted, ref } from "vue";
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
import { getApiError } from "../../utils/error";
import PaginationBar from "../../components/common/PaginationBar.vue";

const LIMIT = 20;

const users = ref<AdminUserItem[]>([]);
const isLoading = ref(false);
const errorMessage = ref("");
const searchQuery = ref("");
const page = ref(1);
const totalPages = ref(1);
const totalItems = ref(0);

let searchVersion = 0;
let searchTimer: ReturnType<typeof setTimeout> | null = null;

async function loadUsers(targetPage = page.value) {
  isLoading.value = true;
  errorMessage.value = "";

  try {
    const response = await getAdminUsers({
      page: targetPage,
      limit: LIMIT,
      search: searchQuery.value.trim() || undefined,
    });
    users.value = response.data ?? [];
    page.value = response.page ?? targetPage;
    totalPages.value = response.totalPages ?? 1;
    totalItems.value = Number(response.totalItems ?? 0);
  } catch (error) {
    users.value = [];
    errorMessage.value = getApiError(error);
  } finally {
    isLoading.value = false;
  }
}

function onSearchInput() {
  if (searchTimer) clearTimeout(searchTimer);
  searchTimer = setTimeout(async () => {
    const version = ++searchVersion;
    page.value = 1;
    isLoading.value = true;
    errorMessage.value = "";
    try {
      const response = await getAdminUsers({
        page: 1,
        limit: LIMIT,
        search: searchQuery.value.trim() || undefined,
      });
      if (version !== searchVersion) return;
      users.value = response.data ?? [];
      page.value = response.page ?? 1;
      totalPages.value = response.totalPages ?? 1;
      totalItems.value = Number(response.totalItems ?? 0);
    } catch (error) {
      if (version !== searchVersion) return;
      errorMessage.value = getApiError(error);
    } finally {
      if (version === searchVersion) isLoading.value = false;
    }
  }, 300);
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
          <p class="eyebrow">Super Admin</p>
          <h1 class="mt-2 text-2xl font-semibold text-[#171322] sm:text-3xl">
            Akun Global
          </h1>
          <p class="mt-2 max-w-3xl text-sm leading-6 text-[#6b7280]">
            Lihat akun login Wiyata lintas sekolah. Akses sekolah dan peran
            tidak dikelola dari halaman ini.
          </p>
        </div>
        <button
          type="button"
          class="inline-flex w-full items-center justify-center gap-2 rounded-lg border border-[#ebe7df] bg-white px-4 py-2.5 text-sm font-medium text-[#374151] transition hover:border-[#4f46e5] hover:text-[#4f46e5] disabled:cursor-not-allowed disabled:opacity-60 sm:w-auto"
          :disabled="isLoading"
          @click="loadUsers(1)"
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
          <article
            class="rounded-xl border border-[#ebe7df] bg-white p-4 shadow-sm"
          >
            <p class="text-xs font-medium text-[#6b7280]">Total akun</p>
            <p class="mt-2 text-2xl font-semibold text-[#171322]">
              {{ totalItems || "–" }}
            </p>
          </article>
          <article
            class="col-span-2 rounded-xl border border-[#ebe7df] bg-white p-4 shadow-sm"
          >
            <p class="text-xs font-medium text-[#6b7280]">Halaman ini</p>
            <p class="mt-2 text-sm leading-6 text-[#374151]">
              Menampilkan {{ users.length }} dari {{ totalItems }} akun.
              Gunakan pencarian untuk menyaring hasil.
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
              <p class="eyebrow">Daftar akun login</p>
              <h2 class="mt-2 text-xl font-semibold text-[#171322]">
                Pengguna Wiyata
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
                @input="onSearchInput"
              />
            </label>
          </div>

          <div class="mt-5 flex flex-col gap-4">
            <div v-if="isLoading" class="space-y-3">
              <div
                v-for="item in 3"
                :key="item"
                class="h-20 animate-pulse rounded-lg bg-[#fbfaf8]"
              />
            </div>

            <div
              v-else-if="errorMessage"
              class="rounded-lg border border-[#fecaca] bg-[#fff8f6] px-4 py-4"
            >
              <p class="text-sm leading-6 text-[#a8665d]">{{ errorMessage }}</p>
              <button
                type="button"
                class="mt-3 inline-flex items-center justify-center gap-2 rounded-lg border border-[#ebe7df] bg-white px-3 py-2 text-sm font-medium text-[#374151] transition hover:border-[#4f46e5] hover:text-[#4f46e5] disabled:cursor-not-allowed disabled:opacity-60"
                @click="loadUsers(1)"
              >
                Coba lagi
              </button>
            </div>

            <div
              v-else-if="users.length === 0 && !searchQuery"
              class="rounded-lg bg-[#fbfaf8] px-5 py-8 text-center"
            >
              <PhUsers class="mx-auto h-7 w-7 text-[#9ca3af]" weight="duotone" />
              <h3 class="mt-3 text-sm font-semibold text-[#171322]">
                Belum ada akun global
              </h3>
              <p class="mt-2 text-sm leading-6 text-[#6b7280]">
                Akun global akan muncul setelah pengguna terdaftar.
              </p>
            </div>

            <div
              v-else-if="users.length === 0 && searchQuery"
              class="rounded-lg bg-[#fbfaf8] px-5 py-8 text-center"
            >
              <PhMagnifyingGlass
                class="mx-auto h-7 w-7 text-[#9ca3af]"
                weight="duotone"
              />
              <h3 class="mt-3 text-sm font-semibold text-[#171322]">
                Tidak ada akun yang cocok
              </h3>
              <p class="mt-2 text-sm leading-6 text-[#6b7280]">
                Ubah kata kunci pencarian untuk melihat akun lain.
              </p>
            </div>

            <template v-else>
              <div class="space-y-3">
                <article
                  v-for="user in users"
                  :key="user.userId"
                  class="rounded-xl border border-[#ebe7df] bg-[#fcfbf8] p-4"
                >
                  <div
                    class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between"
                  >
                    <div class="min-w-0">
                      <div class="flex min-w-0 flex-wrap items-center gap-2">
                        <h3
                          class="truncate text-base font-semibold text-[#171322]"
                        >
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
              </div>

              <PaginationBar
                :page="page"
                :total-pages="totalPages"
                :total-items="totalItems"
                :limit="LIMIT"
                @change="(p) => loadUsers(p)"
              />
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
              <p class="eyebrow">Informasi akses</p>
              <h2 class="mt-2 text-xl font-semibold text-[#171322]">
                Identitas global saja
              </h2>
              <p class="mt-2 text-sm leading-6 text-[#6b7280]">
                Halaman ini hanya menampilkan akun login Wiyata.
              </p>
            </div>
            <span
              class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-[#fff4ee] text-[#ea580c]"
            >
              <PhUsers :size="22" weight="duotone" />
            </span>
          </div>

          <div class="mt-5 space-y-3">
            <article
              class="rounded-lg border border-[#ebe7df] bg-[#fcfbf8] p-4"
            >
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
                    Identitas login pengguna Wiyata lintas sekolah.
                  </p>
                </div>
              </div>
            </article>

            <article
              class="rounded-lg border border-[#ebe7df] bg-[#fcfbf8] p-4"
            >
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

            <article
              class="rounded-lg border border-[#fed7aa] bg-[#fff7ed] p-4"
            >
              <p class="text-xs leading-5 text-[#9a3412]">
                Operasional akademik setiap sekolah tetap dikelola oleh Admin
                Sekolah.
              </p>
            </article>
          </div>
        </section>
      </aside>
    </section>
  </main>
</template>
