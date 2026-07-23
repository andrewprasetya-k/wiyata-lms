<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import {
  PhBuildings,
  PhDeviceMobile,
  PhEye,
  PhEyeSlash,
  PhIdentificationCard,
  PhLockKey,
  PhUserCircle,
} from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import { useToastStore } from "../../stores/toast";
import { useConfirmStore } from "../../stores/confirm";
import { usePasswordVisibility } from "../../composables/usePasswordVisibility";
import { changePassword } from "../../services/changePassword";
import {
  listSessions,
  revokeSession,
  type Session,
} from "../../services/sessions";
import { passwordPolicyError } from "../../utils/passwordPolicy";
import { summarizeUserAgent } from "../../utils/userAgent";
import { formatDateTime } from "../../utils/date";
import { getApiError } from "../../utils/error";
import type { RoleName } from "../../types/auth";

const props = defineProps<{
  eyebrow: string;
  title: string;
  helper: string;
}>();

const auth = useAuthStore();
const toast = useToastStore();
const confirm = useConfirmStore();

const activeMembership = computed(() => auth.activeMembership);
const roleLabels: Record<RoleName, string> = {
  student: "Siswa",
  teacher: "Guru",
  admin: "Admin",
  super_admin: "Superadmin",
};

const titleLabels: Record<string, string> = {
  "Profil Student": "Profil Siswa",
  "Profil Teacher": "Profil Guru",
  "Profil Admin": "Profil Admin",
  "Profil Superadmin": "Profil Superadmin",
};

const eyebrowLabels: Record<string, string> = {
  "Student profile": "Akun siswa",
  "Teacher profile": "Akun guru",
  "Admin profile": "Akun admin",
  "Super admin profile": "Akun superadmin",
};

const helperLabels: Record<string, string> = {
  "Student profile":
    "Lihat informasi akun, peran, dan sekolah yang sedang aktif.",
  "Teacher profile":
    "Lihat informasi akun guru, peran aktif, dan akses sekolah.",
  "Admin profile":
    "Lihat informasi akun admin dan sekolah yang sedang dikelola.",
  "Super admin profile": "Lihat informasi akun dan peran pengelola platform.",
};

const pageTitle = computed(() => titleLabels[props.title] ?? props.title);
const pageEyebrow = computed(
  () => eyebrowLabels[props.eyebrow] ?? "Informasi akun",
);
const pageHelper = computed(
  () =>
    helperLabels[props.eyebrow] ||
    props.helper ||
    "Lihat informasi akun dan konteks aktif.",
);

const currentRoles = computed(() => {
  const roles = auth.activeRoles.length > 0 ? auth.activeRoles : auth.allRoles;
  return roles.map((role) => roleLabels[role] ?? role).join(", ") || "-";
});

const initials = computed(() => {
  const name = auth.user?.fullName ?? "";
  const value = name
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0])
    .join("")
    .toUpperCase();
  return value || "EV";
});

const profileRows = computed(() => [
  { label: "Nama lengkap", value: auth.user?.fullName || "-" },
  { label: "Email", value: auth.user?.email || "-" },
  { label: "Peran aktif", value: currentRoles.value },
]);

const schoolRows = computed(() => [
  { label: "Sekolah aktif", value: activeMembership.value?.school.name || "-" },
  { label: "Kode sekolah", value: activeMembership.value?.school.code || "-" },
  {
    label: "Sekolah utama",
    value: activeMembership.value?.isDefault ? "Ya" : "Bukan",
  },
]);

// --- Change password (self-service) ---
const currentPassword = ref("");
const newPassword = ref("");
const confirmNewPassword = ref("");
const isChangingPassword = ref(false);
const passwordErrorMessage = ref("");

const {
  visible: currentPasswordVisible,
  inputType: currentPasswordInputType,
  toggle: toggleCurrentPasswordVisibility,
} = usePasswordVisibility();
const {
  visible: newPasswordVisible,
  inputType: newPasswordInputType,
  toggle: toggleNewPasswordVisibility,
} = usePasswordVisibility();
const {
  visible: confirmNewPasswordVisible,
  inputType: confirmNewPasswordInputType,
  toggle: toggleConfirmNewPasswordVisibility,
} = usePasswordVisibility();

const canSubmitPasswordChange = computed(
  () =>
    currentPassword.value !== "" &&
    newPassword.value !== "" &&
    confirmNewPassword.value !== "" &&
    !isChangingPassword.value,
);

function passwordErrorFromResponse(error: unknown) {
  const maybeError = error as { response?: { data?: { error?: string } } };
  return (
    maybeError.response?.data?.error ??
    "Password belum bisa diubah. Coba lagi sebentar lagi."
  );
}

async function submitPasswordChange() {
  if (isChangingPassword.value) return;
  passwordErrorMessage.value = "";

  if (newPassword.value !== confirmNewPassword.value) {
    passwordErrorMessage.value = "Konfirmasi password baru belum sama.";
    return;
  }
  const policyError = passwordPolicyError(newPassword.value);
  if (policyError) {
    passwordErrorMessage.value = policyError;
    return;
  }

  isChangingPassword.value = true;
  try {
    await changePassword({
      currentPassword: currentPassword.value,
      newPassword: newPassword.value,
    });
    currentPassword.value = "";
    newPassword.value = "";
    confirmNewPassword.value = "";
    toast.success("Password berhasil diubah.");
  } catch (error) {
    passwordErrorMessage.value = passwordErrorFromResponse(error);
  } finally {
    isChangingPassword.value = false;
  }
}

// --- Login sessions ---
const sessions = ref<Session[]>([]);
const sessionsLoading = ref(true);
const sessionsErrorMessage = ref("");
const revokingSessionId = ref<string | null>(null);

async function loadSessions() {
  sessionsLoading.value = true;
  sessionsErrorMessage.value = "";
  try {
    sessions.value = await listSessions();
  } catch (error) {
    sessionsErrorMessage.value = getApiError(error);
  } finally {
    sessionsLoading.value = false;
  }
}

async function confirmRevokeSession(session: Session) {
  const ok = await confirm.confirm({
    title: "Keluar dari sesi ini?",
    description: `Sesi pada ${summarizeUserAgent(session.userAgent)} (${session.ipAddress || "IP tidak diketahui"}) akan diakhiri. Jika ini adalah sesi yang sedang Anda gunakan, Anda akan diminta masuk kembali.`,
    confirmLabel: "Keluar",
    variant: "danger",
  });
  if (!ok) return;

  revokingSessionId.value = session.id;
  try {
    await revokeSession(session.id);
    sessions.value = sessions.value.filter((item) => item.id !== session.id);
    toast.success("Sesi berhasil diakhiri.");
  } catch (error) {
    toast.error(getApiError(error));
  } finally {
    revokingSessionId.value = null;
  }
}

onMounted(loadSessions);
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-background">
    <header class="border-b border-border bg-surface">
      <div class="px-5 py-4 sm:px-6 lg:px-8">
        <p class="text-[11px] font-medium text-brand">
          {{ pageEyebrow }}
        </p>
        <h1 class="mt-1 text-xl font-medium text-foreground sm:text-2xl">
          {{ pageTitle }}
        </h1>
        <p class="mt-1 max-w-2xl text-xs leading-5 text-muted sm:text-sm">
          {{ pageHelper }}
        </p>
      </div>
    </header>

    <section
      v-if="!auth.user"
      class="flex min-h-[65vh] items-center justify-center px-5 py-10 sm:px-6 lg:px-8"
    >
      <article
        class="w-full max-w-xl rounded-xl border border-border bg-surface p-8 text-center"
      >
        <div
          class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-brand-soft text-brand"
        >
          <PhUserCircle :size="25" weight="duotone" />
        </div>
        <h2 class="mt-3 text-lg font-medium text-foreground">
          Data profil belum tersedia
        </h2>
        <p class="mx-auto mt-2 max-w-md text-sm leading-6 text-muted">
          Informasi akun belum ditemukan pada sesi aktif. Silakan masuk kembali
          untuk memuat profil.
        </p>
      </article>
    </section>

    <section
      v-else
      class="mx-auto grid max-w-screen min-w-0 gap-5 px-5 py-5 sm:px-6 lg:grid-cols-[320px_minmax(0,1fr)] lg:px-8 lg:py-6"
    >
      <aside class="min-w-0">
        <article
          class="rounded-xl border border-border bg-surface p-5 lg:sticky lg:top-6"
        >
          <div
            class="flex h-14 w-14 items-center justify-center rounded-xl bg-brand text-base font-medium text-white"
          >
            {{ initials }}
          </div>
          <p class="mt-4 text-xs text-muted">Akun Wiyata</p>
          <h2 class="mt-1 wrap-break-word text-xl font-medium text-foreground">
            {{ auth.user.fullName || "Nama tidak tersedia" }}
          </h2>
          <p class="mt-1 break-all text-sm leading-6 text-muted">
            {{ auth.user.email || "Email tidak tersedia" }}
          </p>
        </article>
      </aside>

      <div class="min-w-0 space-y-5">
        <section class="grid min-w-0 gap-5 xl:grid-cols-2">
          <article
            class="min-w-0 rounded-xl border border-border bg-surface p-5"
          >
            <div class="mb-4 flex min-w-0 items-center gap-3">
              <div
                class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-brand-soft text-brand"
              >
                <PhIdentificationCard :size="21" weight="duotone" />
              </div>
              <div class="min-w-0">
                <h2 class="text-sm font-medium text-foreground">
                  Identitas akun
                </h2>
                <p class="mt-1 text-xs text-muted">
                  Informasi dari sesi yang sedang digunakan.
                </p>
              </div>
            </div>

            <dl class="divide-y divide-border">
              <div
                v-for="row in profileRows"
                :key="row.label"
                class="grid min-w-0 gap-1 py-3 first:pt-0 last:pb-0 sm:grid-cols-[120px_minmax(0,1fr)] sm:gap-4"
              >
                <dt class="text-xs text-muted">{{ row.label }}</dt>
                <dd
                  class="wrap-break-word text-sm font-medium text-foreground sm:text-right"
                >
                  {{ row.value }}
                </dd>
              </div>
            </dl>
          </article>

          <article
            class="min-w-0 rounded-xl border border-border bg-surface p-5"
          >
            <div class="mb-4 flex min-w-0 items-center gap-3">
              <div
                class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-[#f0fdf4] text-[#059669]"
              >
                <PhBuildings :size="21" weight="duotone" />
              </div>
              <div class="min-w-0">
                <h2 class="text-sm font-medium text-foreground">
                  Konteks sekolah
                </h2>
                <p class="mt-1 text-xs text-muted">
                  Sekolah yang sedang digunakan pada sesi ini.
                </p>
              </div>
            </div>

            <dl class="divide-y divide-border">
              <div
                v-for="row in schoolRows"
                :key="row.label"
                class="grid min-w-0 gap-1 py-3 first:pt-0 last:pb-0 sm:grid-cols-[120px_minmax(0,1fr)] sm:gap-4"
              >
                <dt class="text-xs text-muted">{{ row.label }}</dt>
                <dd
                  class="wrap-break-word text-sm font-medium text-foreground sm:text-right"
                >
                  {{ row.value }}
                </dd>
              </div>
            </dl>
          </article>
        </section>

        <section class="min-w-0 rounded-xl border border-border bg-surface p-5">
          <div class="mb-4 flex min-w-0 items-center gap-3">
            <div
              class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-brand-soft text-brand"
            >
              <PhLockKey :size="21" weight="duotone" />
            </div>
            <div class="min-w-0">
              <h2 class="text-sm font-medium text-foreground">
                Ubah kata sandi
              </h2>
              <p class="mt-1 text-xs text-muted">
                Kata sandi baru minimal 8 karakter, kombinasi huruf besar, huruf
                kecil, dan angka.
              </p>
            </div>
          </div>

          <form
            class="grid min-w-0 gap-4 md:grid-cols-2"
            @submit.prevent="submitPasswordChange"
          >
            <label class="block md:col-span-2">
              <span
                class="mb-1.5 block text-xs font-medium text-foreground-secondary"
              >
                Password saat ini
              </span>
              <div class="relative">
                <input
                  v-model="currentPassword"
                  class="h-11 w-full rounded-lg border border-border bg-surface-subtle px-3 pr-11 text-sm outline-none transition focus:border-brand focus:bg-surface"
                  :type="currentPasswordInputType"
                  autocomplete="current-password"
                  placeholder="Password saat ini"
                />
                <button
                  type="button"
                  class="absolute right-2.5 top-1/2 -translate-y-1/2 rounded-lg p-1.5 text-muted transition hover:text-foreground"
                  :aria-label="
                    currentPasswordVisible
                      ? 'Sembunyikan password'
                      : 'Tampilkan password'
                  "
                  :aria-pressed="currentPasswordVisible"
                  @click="toggleCurrentPasswordVisibility"
                >
                  <PhEyeSlash v-if="currentPasswordVisible" :size="16" />
                  <PhEye v-else :size="16" />
                </button>
              </div>
            </label>

            <label class="block">
              <span
                class="mb-1.5 block text-xs font-medium text-foreground-secondary"
              >
                Password baru
              </span>
              <div class="relative">
                <input
                  v-model="newPassword"
                  class="h-11 w-full rounded-lg border border-border bg-surface-subtle px-3 pr-11 text-sm outline-none transition focus:border-brand focus:bg-surface"
                  :type="newPasswordInputType"
                  autocomplete="new-password"
                  placeholder="Minimal 8 karakter"
                />
                <button
                  type="button"
                  class="absolute right-2.5 top-1/2 -translate-y-1/2 rounded-lg p-1.5 text-muted transition hover:text-foreground"
                  :aria-label="
                    newPasswordVisible
                      ? 'Sembunyikan password'
                      : 'Tampilkan password'
                  "
                  :aria-pressed="newPasswordVisible"
                  @click="toggleNewPasswordVisibility"
                >
                  <PhEyeSlash v-if="newPasswordVisible" :size="16" />
                  <PhEye v-else :size="16" />
                </button>
              </div>
            </label>

            <label class="block">
              <span
                class="mb-1.5 block text-xs font-medium text-foreground-secondary"
              >
                Konfirmasi password baru
              </span>
              <div class="relative">
                <input
                  v-model="confirmNewPassword"
                  class="h-11 w-full rounded-lg border border-border bg-surface-subtle px-3 pr-11 text-sm outline-none transition focus:border-brand focus:bg-surface"
                  :type="confirmNewPasswordInputType"
                  autocomplete="new-password"
                  placeholder="Ulangi password baru"
                />
                <button
                  type="button"
                  class="absolute right-2.5 top-1/2 -translate-y-1/2 rounded-lg p-1.5 text-muted transition hover:text-foreground"
                  :aria-label="
                    confirmNewPasswordVisible
                      ? 'Sembunyikan password'
                      : 'Tampilkan password'
                  "
                  :aria-pressed="confirmNewPasswordVisible"
                  @click="toggleConfirmNewPasswordVisibility"
                >
                  <PhEyeSlash v-if="confirmNewPasswordVisible" :size="16" />
                  <PhEye v-else :size="16" />
                </button>
              </div>
            </label>

            <p
              v-if="passwordErrorMessage"
              class="rounded-lg bg-danger-soft px-3 py-2.5 text-sm text-danger md:col-span-2"
            >
              {{ passwordErrorMessage }}
            </p>

            <div class="md:col-span-2">
              <button
                class="flex h-11 items-center justify-center gap-2 rounded-lg bg-brand px-5 text-sm font-medium text-white transition hover:bg-brand-hover disabled:cursor-not-allowed disabled:bg-[#bab7d8]"
                type="submit"
                :disabled="!canSubmitPasswordChange"
              >
                {{
                  isChangingPassword ? "Menyimpan..." : "Simpan password baru"
                }}
              </button>
            </div>
          </form>
        </section>

        <section class="min-w-0 rounded-xl border border-border bg-surface p-5">
          <div class="mb-4 flex min-w-0 items-center gap-3">
            <div
              class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-brand-soft text-brand"
            >
              <PhDeviceMobile :size="21" weight="duotone" />
            </div>
            <div class="min-w-0">
              <h2 class="text-sm font-medium text-foreground">Sesi login</h2>
              <p class="mt-1 text-xs text-muted">
                Perangkat yang sedang masuk ke akun ini.
              </p>
            </div>
          </div>

          <div v-if="sessionsLoading" class="space-y-3">
            <div class="h-16 w-full animate-pulse rounded-lg bg-[#f0ece5]" />
            <div class="h-16 w-full animate-pulse rounded-lg bg-[#f0ece5]" />
          </div>

          <p
            v-else-if="sessionsErrorMessage"
            class="rounded-lg bg-danger-soft px-3 py-2.5 text-sm text-danger"
          >
            {{ sessionsErrorMessage }}
          </p>

          <p v-else-if="sessions.length === 0" class="text-sm text-muted">
            Tidak ada sesi aktif yang tercatat.
          </p>

          <ul v-else class="space-y-3">
            <li
              v-for="session in sessions"
              :key="session.id"
              class="flex min-w-0 flex-col gap-3 rounded-lg border border-border bg-surface-subtle p-4 sm:flex-row sm:items-center sm:justify-between"
            >
              <div class="min-w-0">
                <p class="truncate text-sm font-medium text-foreground">
                  {{ summarizeUserAgent(session.userAgent) }}
                </p>
                <p class="mt-1 text-xs text-muted">
                  {{ session.ipAddress || "IP tidak diketahui" }} · Masuk sejak
                  {{ formatDateTime(session.loggedInAt) }}
                </p>
              </div>
              <button
                type="button"
                class="inline-flex h-9 shrink-0 items-center justify-center rounded-lg border border-border px-4 text-sm font-medium text-danger transition hover:bg-danger-soft disabled:cursor-not-allowed disabled:opacity-60"
                :disabled="revokingSessionId === session.id"
                @click="confirmRevokeSession(session)"
              >
                {{
                  revokingSessionId === session.id ? "Memproses..." : "Keluar"
                }}
              </button>
            </li>
          </ul>
        </section>

        <section
          v-if="auth.memberships.length > 0"
          class="min-w-0 rounded-xl border border-border bg-surface p-5"
        >
          <div class="flex min-w-0 items-center gap-3">
            <div
              class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-warning-soft text-[#ea580c]"
            >
              <PhUserCircle :size="21" weight="duotone" />
            </div>
            <div class="min-w-0">
              <h2 class="text-sm font-medium text-foreground">Akses sekolah</h2>
              <p class="mt-1 text-xs text-muted">
                Sekolah dan peran yang tersedia untuk akun ini.
              </p>
            </div>
          </div>

          <div class="mt-4 grid min-w-0 gap-3 md:grid-cols-2 xl:grid-cols-3">
            <article
              v-for="membership in auth.memberships"
              :key="membership.schoolUserId"
              class="min-w-0 rounded-lg border border-border bg-surface-subtle p-4"
            >
              <p class="truncate text-sm font-medium text-foreground">
                {{ membership.school.name || "Sekolah tidak tersedia" }}
              </p>
              <p class="mt-1 truncate text-xs text-muted">
                {{ membership.school.code || "Kode tidak tersedia" }}
              </p>
              <div class="mt-3 flex flex-wrap gap-2">
                <span
                  v-for="role in membership.roles"
                  :key="role"
                  class="rounded-full bg-brand-soft px-2.5 py-1 text-xs font-medium text-brand"
                >
                  {{ roleLabels[role] ?? role }}
                </span>
                <span
                  v-if="membership.isDefault"
                  class="rounded-full bg-success-soft px-2.5 py-1 text-xs font-medium text-success"
                >
                  Utama
                </span>
              </div>
            </article>
          </div>
        </section>

        <section
          v-else
          class="rounded-xl border border-border bg-surface p-6 text-center"
        >
          <h2 class="text-sm font-medium text-foreground">
            Akses sekolah belum tersedia
          </h2>
          <p class="mt-2 text-sm leading-6 text-muted">
            Akun ini belum memiliki akses sekolah pada sesi aktif.
          </p>
        </section>
      </div>
    </section>
  </main>
</template>
