<script setup lang="ts">
import { onMounted, onUnmounted, ref } from "vue";
import {
  PhCheckCircle,
  PhKey,
  PhShieldWarning,
  PhWarningCircle,
} from "@phosphor-icons/vue";
import { getSuperAdminSecurityDashboard } from "../../services/securityDashboard";
import {
  connectAuditSocket,
  type AuditSocketStatus,
} from "../../services/auditLogSocket";
import type { AuditLogEvent } from "../../types/auditLog";
import type { SecurityDashboardSummary } from "../../types/securityDashboard";
import { formatDateTime } from "../../utils/date";

// super_admin only (Phase 11.5.2) — the school-admin variant of this page
// and its endpoint were removed entirely, not just hidden, so there is no
// mode prop here to branch on anymore. Live updates use the "platform"
// channel, same one the existing audit log live feed already uses — every
// action this page reads is scope=platform (see docs/api/dashboard.md §5),
// which is also why a school-admin variant could never get a true live
// push in the first place (school admins are forbidden from that channel).

const loading = ref(true);
const errorMessage = ref("");
const dashboard = ref<SecurityDashboardSummary | null>(null);

const liveStatus = ref<AuditSocketStatus>("disconnected");
let socket: ReturnType<typeof connectAuditSocket> | null = null;
let refreshDebounceTimer: number | undefined;

const relevantActions = new Set([
  "auth.login.failed",
  "auth.password.reset.requested",
  "auth.password.reset.completed",
  "auth.token.reuse_detected",
  "auth.mfa.recovery_code.used",
  "auth.mfa.verify.failed",
]);

const actionLabels: Record<string, string> = {
  "auth.token.reuse_detected":
    "Refresh token dipakai ulang — kemungkinan token dicuri",
  "auth.mfa.recovery_code.used": "Recovery code MFA dipakai untuk masuk",
  "auth.mfa.verify.failed": "Percobaan verifikasi MFA gagal",
};

function actionLabel(action: string) {
  return actionLabels[action] ?? action;
}

function severityBadgeClass(value?: string) {
  if (value === "HIGH") return "bg-danger-soft text-danger";
  if (value === "MEDIUM") return "bg-warning-soft text-warning-hover";
  return "bg-info-soft text-info-hover";
}

async function loadDashboard() {
  errorMessage.value = "";
  try {
    dashboard.value = await getSuperAdminSecurityDashboard();
  } catch {
    errorMessage.value =
      "Dashboard keamanan belum bisa dimuat. Periksa koneksi atau coba lagi nanti.";
  } finally {
    loading.value = false;
  }
}

// A single incoming event can't safely be turned into an updated brute-force
// incident list client-side (that needs the same 15-minute density check
// the backend already runs across every attempt for that email, not just
// the one new event) — so a live event just schedules a debounced re-fetch
// of the whole summary instead of re-implementing that logic in the
// frontend. Bursts of events (e.g. an active brute-force attempt) coalesce
// into one refetch a few seconds after the burst quiets down.
function scheduleRefresh() {
  if (refreshDebounceTimer) window.clearTimeout(refreshDebounceTimer);
  refreshDebounceTimer = window.setTimeout(loadDashboard, 3000);
}

function handleAuditEvent(event: AuditLogEvent) {
  if (!relevantActions.has(event.payload.action)) return;
  scheduleRefresh();
}

function connectLiveFeed() {
  socket = connectAuditSocket({
    channel: "platform",
    onEvent: handleAuditEvent,
    onStatusChange: (status) => (liveStatus.value = status),
  });
}

function disconnectLiveFeed() {
  socket?.close();
  socket = null;
  if (refreshDebounceTimer) window.clearTimeout(refreshDebounceTimer);
}

onMounted(async () => {
  await loadDashboard();
  connectLiveFeed();
});

onUnmounted(disconnectLiveFeed);
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-background">
    <header class="border-b border-border bg-surface">
      <div class="flex min-w-0 flex-col gap-3 px-5 py-5 sm:px-6 lg:px-8">
        <div class="flex items-center justify-between gap-3">
          <div class="flex items-center gap-3">
            <div
              class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-brand-soft text-brand"
            >
              <PhShieldWarning :size="20" weight="duotone" />
            </div>
            <div class="min-w-0">
              <h1 class="text-2xl font-semibold text-foreground sm:text-3xl">
                Keamanan
              </h1>
              <p class="mt-1 text-sm leading-6 text-muted">
                Aktivitas keamanan di seluruh platform, 24 jam terakhir.
              </p>
            </div>
          </div>

          <span
            class="inline-flex shrink-0 items-center gap-1.5 rounded-full px-2.5 py-1 text-[11px] font-medium"
            :class="
              liveStatus === 'connected'
                ? 'bg-success-soft text-success'
                : liveStatus === 'failed'
                  ? 'bg-danger-soft text-danger'
                  : 'bg-surface-subtle text-muted'
            "
          >
            <span
              class="h-1.5 w-1.5 rounded-full"
              :class="
                liveStatus === 'connected'
                  ? 'bg-success'
                  : liveStatus === 'failed'
                    ? 'bg-danger'
                    : 'bg-muted'
              "
            />
            {{
              liveStatus === "connected"
                ? "Live"
                : liveStatus === "failed"
                  ? "Terputus"
                  : "Menghubungkan..."
            }}
          </span>
        </div>
      </div>
    </header>

    <div class="space-y-6 px-5 py-6 sm:px-6 lg:px-8">
      <!-- Error state -->
      <section
        v-if="errorMessage && !loading"
        class="rounded-xl border border-danger-line bg-danger-soft p-5"
      >
        <div class="flex items-start gap-3">
          <PhWarningCircle
            class="mt-0.5 h-5 w-5 shrink-0 text-danger"
            weight="duotone"
          />
          <div class="min-w-0">
            <p class="text-sm font-medium text-foreground">
              Dashboard tidak dapat dimuat
            </p>
            <p class="mt-1 text-sm leading-6 text-muted">{{ errorMessage }}</p>
            <button
              class="mt-3 rounded-lg bg-brand px-4 py-2 text-sm font-medium text-white transition hover:bg-brand-hover"
              type="button"
              @click="loadDashboard"
            >
              Coba lagi
            </button>
          </div>
        </div>
      </section>

      <!-- Stat cards: failed login + password reset -->
      <section class="grid grid-cols-1 gap-3 sm:grid-cols-3">
        <template v-if="loading">
          <div
            v-for="i in 3"
            :key="i"
            class="h-28 animate-pulse rounded-xl border border-border bg-surface shadow-sm"
          />
        </template>
        <template v-else-if="dashboard">
          <div class="rounded-xl border border-border bg-surface shadow-sm p-4">
            <div
              class="flex h-10 w-10 items-center justify-center rounded-xl bg-danger-soft text-danger"
            >
              <PhKey :size="20" weight="duotone" />
            </div>
            <p class="mt-4 text-2xl font-semibold text-foreground">
              {{ dashboard.failedLoginCount }}
            </p>
            <p class="mt-1 text-xs text-muted">
              Percobaan login gagal ({{ dashboard.windowHours }} jam terakhir)
            </p>
          </div>

          <div class="rounded-xl border border-border bg-surface shadow-sm p-4">
            <div
              class="flex h-10 w-10 items-center justify-center rounded-xl bg-warning-soft text-[#ea580c]"
            >
              <PhKey :size="20" weight="duotone" />
            </div>
            <p class="mt-4 text-2xl font-semibold text-foreground">
              {{ dashboard.passwordResetRequestedCount }}
            </p>
            <p class="mt-1 text-xs text-muted">Reset kata sandi diminta</p>
          </div>

          <div class="rounded-xl border border-border bg-surface shadow-sm p-4">
            <div
              class="flex h-10 w-10 items-center justify-center rounded-xl bg-success-soft text-success"
            >
              <PhCheckCircle :size="20" weight="duotone" />
            </div>
            <p class="mt-4 text-2xl font-semibold text-foreground">
              {{ dashboard.passwordResetCompletedCount }}
            </p>
            <p class="mt-1 text-xs text-muted">Reset kata sandi selesai</p>
          </div>
        </template>
      </section>

      <p v-if="!loading && dashboard" class="-mt-3 text-xs leading-5 text-muted">
        Jumlah "diminta" dan "selesai" dihitung independen dalam rentang
        waktu yang sama — bukan pasangan satu-ke-satu, karena permintaan
        reset tidak selalu diselesaikan dalam rentang yang sama pula.
      </p>

      <!-- Brute force incidents -->
      <section class="rounded-xl border border-border bg-surface shadow-sm p-5">
        <div class="mb-4">
          <h2 class="text-sm font-semibold text-foreground">
            Dugaan Serangan Brute Force
          </h2>
          <p class="mt-0.5 text-xs text-muted">
            Akun atau alamat IP dengan minimal 5 percobaan login gagal
            dalam rentang 15 menit mana pun.
          </p>
        </div>

        <div v-if="loading" class="space-y-2">
          <div
            v-for="i in 3"
            :key="i"
            class="h-12 animate-pulse rounded-lg bg-surface-strong"
          />
        </div>

        <div
          v-else-if="!dashboard?.bruteForceIncidents?.length"
          class="rounded-lg bg-surface-subtle px-4 py-8 text-center"
        >
          <PhCheckCircle
            class="mx-auto h-7 w-7 text-border-strong"
            weight="duotone"
          />
          <p class="mt-3 text-sm font-medium text-foreground">
            Tidak ada dugaan brute force
          </p>
          <p class="mt-1 text-xs leading-5 text-muted">
            Belum ada akun dengan pola login gagal yang mencurigakan.
          </p>
        </div>

        <div v-else class="divide-y divide-[#f3f4f6]">
          <div
            v-for="(incident, index) in dashboard.bruteForceIncidents"
            :key="`${incident.targetType}-${incident.target}-${index}`"
            class="flex items-center justify-between gap-3 py-2.5 first:pt-0 last:pb-0"
          >
            <div class="min-w-0 flex-1">
              <p class="flex items-center gap-2 truncate text-sm font-medium text-foreground">
                <span
                  class="shrink-0 rounded-full bg-surface-strong px-1.5 py-0.5 text-[10px] font-semibold uppercase text-muted"
                >
                  {{ incident.targetType === "ip" ? "IP" : "Akun" }}
                </span>
                <span class="truncate">{{ incident.target }}</span>
              </p>
              <p class="mt-0.5 text-xs text-muted">
                Percobaan terakhir {{ formatDateTime(incident.lastAttemptAt) }}
              </p>
            </div>
            <span
              class="shrink-0 rounded-full bg-danger-soft px-2.5 py-1 text-xs font-medium text-danger"
            >
              {{ incident.failureCount }}x gagal
            </span>
          </div>
        </div>
      </section>

      <!-- Suspicious activity -->
      <section class="rounded-xl border border-border bg-surface shadow-sm p-5">
        <div class="mb-4">
          <h2 class="text-sm font-semibold text-foreground">
            Aktivitas Mencurigakan
          </h2>
          <p class="mt-0.5 text-xs text-muted">
            Token yang dipakai ulang, recovery code MFA yang dipakai, dan
            verifikasi MFA yang gagal.
          </p>
        </div>

        <div v-if="loading" class="space-y-2">
          <div
            v-for="i in 4"
            :key="i"
            class="h-12 animate-pulse rounded-lg bg-surface-strong"
          />
        </div>

        <div
          v-else-if="!dashboard?.suspiciousActivities?.length"
          class="rounded-lg bg-surface-subtle px-4 py-8 text-center"
        >
          <PhShieldWarning
            class="mx-auto h-7 w-7 text-border-strong"
            weight="duotone"
          />
          <p class="mt-3 text-sm font-medium text-foreground">
            Belum ada aktivitas mencurigakan
          </p>
        </div>

        <div v-else class="divide-y divide-[#f3f4f6]">
          <div
            v-for="activity in dashboard.suspiciousActivities"
            :key="activity.logId"
            class="flex items-start justify-between gap-3 py-3 first:pt-0 last:pb-0"
          >
            <div class="min-w-0 flex-1">
              <p class="text-sm leading-5 text-foreground">
                {{ actionLabel(activity.action) }}
              </p>
              <p class="mt-0.5 truncate text-xs text-muted">
                {{ activity.userName || activity.userEmail || "Tidak diketahui" }}
                · {{ formatDateTime(activity.createdAt) }}
              </p>
            </div>
            <span
              class="shrink-0 rounded-full px-2 py-0.5 text-[11px] font-medium"
              :class="severityBadgeClass(activity.severity)"
            >
              {{ activity.severity }}
            </span>
          </div>
        </div>
      </section>
    </div>
  </main>
</template>
