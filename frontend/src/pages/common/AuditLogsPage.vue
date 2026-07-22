<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import {
  PhClockCounterClockwise,
  PhMagnifyingGlass,
  PhWarningCircle,
  PhX,
} from "@phosphor-icons/vue";
import { useAuthStore } from "../../stores/auth";
import {
  getPlatformAuditLogDetail,
  getPlatformAuditLogs,
  getSchoolAuditLogDetail,
  getSchoolAuditLogs,
} from "../../services/auditLog";
import {
  connectAuditSocket,
  type AuditSocketStatus,
} from "../../services/auditLogSocket";
import { getSuperAdminSchools } from "../../services/superAdminSchool";
import type {
  AuditLogDetail,
  AuditLogEvent,
  AuditLogListItem,
} from "../../types/auditLog";
import { formatDateTime } from "../../utils/date";
import { getApiError } from "../../utils/error";
import PaginationBar from "../../components/common/PaginationBar.vue";
import JsonViewer from "../../components/common/JsonViewer.vue";

const props = defineProps<{
  mode: "admin" | "superadmin";
}>();

const auth = useAuthStore();
const LIMIT = 20;

// A super admin viewing this page while "acting as" a school (active school
// context selected) still gets the unrestricted platform search — the
// schoolId filter below just narrows it. Only genuine school admins are
// forced onto the pinned /logs/school/:schoolId/search route.
const isSuperAdmin = computed(() => props.mode === "superadmin");
const ownSchoolId = computed(() => auth.activeSchoolId ?? "");

const items = ref<AuditLogListItem[]>([]);
const loading = ref(false);
const errorMessage = ref("");
const page = ref(1);
const totalPages = ref(1);
const totalItems = ref(0);

const search = ref("");
const severity = ref("");
const scope = ref("");
const entityType = ref("");
const actorUserId = ref("");
const correlationId = ref("");
const dateFrom = ref("");
const dateTo = ref("");
const schoolFilter = ref("");

const schoolOptions = ref<{ id: string; name: string; code: string }[]>([]);
const schoolOptionsLoading = ref(false);

const severityOptions = [
  { value: "", label: "Semua tingkat" },
  { value: "LOW", label: "LOW" },
  { value: "MEDIUM", label: "MEDIUM" },
  { value: "HIGH", label: "HIGH" },
];

const scopeOptions = [
  { value: "", label: "Semua cakupan" },
  { value: "platform", label: "Platform" },
  { value: "school", label: "Sekolah" },
];

function severityBadgeClass(value?: string) {
  if (value === "HIGH") return "bg-danger-soft text-danger";
  if (value === "MEDIUM") return "bg-warning-soft text-warning-hover";
  if (value === "LOW") return "bg-info-soft text-info-hover";
  return "bg-surface-subtle text-muted";
}

function scopeBadgeClass(value?: string) {
  if (value === "platform") return "bg-brand-soft text-brand";
  if (value === "school") return "bg-surface-strong text-foreground-secondary";
  return "bg-surface-subtle text-muted";
}

function scopeLabel(value?: string) {
  if (value === "platform") return "Platform";
  if (value === "school") return "Sekolah";
  return value || "—";
}

async function loadSchoolOptions() {
  if (!isSuperAdmin.value) return;
  schoolOptionsLoading.value = true;
  try {
    const data = await getSuperAdminSchools({ page: 1, limit: 200 });
    schoolOptions.value = (data.data ?? []).map((school) => ({
      id: school.schoolId,
      name: school.schoolName,
      code: school.schoolCode,
    }));
  } catch {
    // Non-critical: the school filter just stays empty if this fails.
  } finally {
    schoolOptionsLoading.value = false;
  }
}

async function loadLogs(targetPage = page.value) {
  loading.value = true;
  errorMessage.value = "";
  try {
    const filters = {
      search: search.value.trim(),
      severity: severity.value,
      scope: isSuperAdmin.value ? scope.value : undefined,
      entityType: entityType.value.trim(),
      actorUserId: actorUserId.value.trim(),
      correlationId: correlationId.value.trim(),
      dateFrom: dateFrom.value,
      dateTo: dateTo.value,
      page: targetPage,
      limit: LIMIT,
    };

    const data = isSuperAdmin.value
      ? await getPlatformAuditLogs({
          ...filters,
          schoolId: schoolFilter.value || undefined,
        })
      : await getSchoolAuditLogs(ownSchoolId.value, filters);

    items.value = data.data ?? [];
    page.value = data.page ?? targetPage;
    totalPages.value = data.totalPages ?? 1;
    totalItems.value = Number(data.totalItems ?? 0);
  } catch (error) {
    errorMessage.value = getApiError(error) || "Log audit belum bisa dimuat.";
  } finally {
    loading.value = false;
  }
}

let searchTimer: ReturnType<typeof setTimeout> | null = null;
function scheduleReload() {
  if (searchTimer) clearTimeout(searchTimer);
  searchTimer = setTimeout(() => {
    page.value = 1;
    loadLogs(1);
  }, 300);
}

watch(
  [
    search,
    severity,
    scope,
    entityType,
    actorUserId,
    correlationId,
    dateFrom,
    dateTo,
    schoolFilter,
  ],
  scheduleReload,
);

function resetFilters() {
  search.value = "";
  severity.value = "";
  scope.value = "";
  entityType.value = "";
  actorUserId.value = "";
  correlationId.value = "";
  dateFrom.value = "";
  dateTo.value = "";
  schoolFilter.value = "";
}

// Detail drawer
const detailOpen = ref(false);
const detailLoading = ref(false);
const detailError = ref("");
const detail = ref<AuditLogDetail | null>(null);
const parsedMetadata = computed<unknown>(() => {
  if (!detail.value?.metadata) return null;
  try {
    return JSON.parse(detail.value.metadata);
  } catch {
    return detail.value.metadata;
  }
});

async function openDetail(row: AuditLogListItem) {
  detailOpen.value = true;
  detailLoading.value = true;
  detailError.value = "";
  detail.value = null;
  try {
    detail.value = isSuperAdmin.value
      ? await getPlatformAuditLogDetail(row.logId)
      : await getSchoolAuditLogDetail(ownSchoolId.value, row.logId);
  } catch (error) {
    detailError.value = getApiError(error) || "Detail log belum bisa dimuat.";
  } finally {
    detailLoading.value = false;
  }
}

function closeDetail() {
  detailOpen.value = false;
}

// Live feed (Phase 10.10). REST stays the source of truth for the actual
// page contents — the socket only prepends/updates counts for what the
// user is already looking at on page 1, and never forces navigation.
const liveStatus = ref<AuditSocketStatus>("disconnected");
const highlightedIds = ref<Set<string>>(new Set());
const highlightTimers = new Map<string, ReturnType<typeof setTimeout>>();

function eventMatchesActiveFilters(item: AuditLogListItem) {
  // search and date-range can't be verified reliably against the live
  // payload (no actor name/email; date-format parsing is fragile) —
  // safest to skip live updates entirely while either is active rather
  // than risk a wrong count/row. REST stays correct regardless.
  if (search.value.trim() || dateFrom.value || dateTo.value) return false;
  if (severity.value && item.severity !== severity.value) return false;
  if (isSuperAdmin.value && scope.value && item.scope !== scope.value)
    return false;
  if (entityType.value && item.entityType !== entityType.value) return false;
  if (actorUserId.value && item.actorUserId !== actorUserId.value) return false;
  if (
    isSuperAdmin.value &&
    schoolFilter.value &&
    item.schoolId !== schoolFilter.value
  ) {
    return false;
  }
  if (correlationId.value && item.correlationId !== correlationId.value) {
    return false;
  }
  return true;
}

function handleAuditEvent(event: AuditLogEvent) {
  const item = event.payload;
  if (!item?.logId || !eventMatchesActiveFilters(item)) return;

  totalItems.value += 1;
  totalPages.value = Math.max(1, Math.ceil(totalItems.value / LIMIT));

  if (page.value !== 1) return; // don't disturb whatever page the user is on
  if (items.value[0]?.logId === item.logId) return; // dedupe safety net

  items.value = [item, ...items.value].slice(0, LIMIT);
  highlightedIds.value.add(item.logId);
  const timer = setTimeout(() => {
    highlightedIds.value.delete(item.logId);
    highlightTimers.delete(item.logId);
  }, 4000);
  highlightTimers.set(item.logId, timer);
}

let platformSocket: ReturnType<typeof connectAuditSocket> | null = null;
let schoolSocket: ReturnType<typeof connectAuditSocket> | null = null;
let watchedSchoolChannel = "";

function connectLiveFeed() {
  if (isSuperAdmin.value) {
    platformSocket = connectAuditSocket({
      channel: "platform",
      onEvent: handleAuditEvent,
      onStatusChange: (status) => (liveStatus.value = status),
    });
  } else if (ownSchoolId.value) {
    schoolSocket = connectAuditSocket({
      channel: ownSchoolId.value,
      onEvent: handleAuditEvent,
      onStatusChange: (status) => (liveStatus.value = status),
    });
    watchedSchoolChannel = ownSchoolId.value;
  }
}

function reloadPage() {
  window.location.reload();
}

function disconnectLiveFeed() {
  platformSocket?.close();
  platformSocket = null;
  schoolSocket?.close();
  schoolSocket = null;
  watchedSchoolChannel = "";
  for (const timer of highlightTimers.values()) clearTimeout(timer);
  highlightTimers.clear();
  highlightedIds.value = new Set();
}

// Super admin: also watch the specifically-filtered school's channel (in
// addition to the always-on platform channel) so narrowing to one school
// gets live updates for that school too — "platform + school sesuai
// kebutuhan" per the permission brief.
watch(schoolFilter, (nextSchoolId) => {
  if (!isSuperAdmin.value) return;
  if (nextSchoolId === watchedSchoolChannel) return;

  schoolSocket?.close();
  schoolSocket = null;
  watchedSchoolChannel = nextSchoolId;
  if (nextSchoolId) {
    schoolSocket = connectAuditSocket({
      channel: nextSchoolId,
      onEvent: handleAuditEvent,
    });
  }
});

onMounted(async () => {
  await loadSchoolOptions();
  await loadLogs();
  connectLiveFeed();
});

onUnmounted(() => {
  disconnectLiveFeed();
});
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
              <PhClockCounterClockwise :size="20" weight="duotone" />
            </div>
            <div class="min-w-0">
              <h1 class="text-2xl font-semibold text-foreground sm:text-3xl">
                Log Audit
              </h1>
              <p class="mt-1 text-sm leading-6 text-muted">
                {{
                  isSuperAdmin
                    ? "Aktivitas di seluruh sekolah dan platform."
                    : "Aktivitas admin dan anggota pada sekolah aktif."
                }}
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
        <div
          v-if="liveStatus === 'failed'"
          class="mt-3 flex flex-wrap items-center justify-between gap-3 rounded-lg border border-danger-line bg-danger-soft p-3 text-sm text-danger"
        >
          <span>Live feed terputus — muat ulang halaman untuk menyambung kembali.</span>
          <button
            type="button"
            class="shrink-0 rounded-lg border border-danger-line bg-surface px-3 py-1.5 text-xs font-semibold text-danger transition hover:brightness-95"
            @click="reloadPage"
          >
            Muat Ulang
          </button>
        </div>
      </div>
    </header>

    <section class="px-5 py-5 sm:px-6 lg:px-8 lg:py-6">
      <div class="rounded-xl border border-border bg-surface">
        <div class="flex flex-col gap-4 p-5">
          <div class="relative min-w-0">
            <PhMagnifyingGlass
              :size="17"
              class="pointer-events-none absolute left-3.5 top-1/2 -translate-y-1/2 text-muted"
            />
            <input
              v-model="search"
              type="search"
              placeholder="Cari action, entity, atau nama/email actor"
              class="w-full rounded-lg border border-border bg-surface-subtle py-2.5 pl-10 pr-3.5 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-brand focus:bg-surface"
            />
          </div>

          <div class="flex flex-wrap items-end gap-3">
            <label class="block text-xs font-medium text-muted">
              Tingkat
              <select
                v-model="severity"
                class="mt-1.5 rounded-lg border border-border bg-surface-subtle px-3 py-2 text-sm text-foreground outline-none transition focus:border-brand focus:bg-surface"
              >
                <option
                  v-for="option in severityOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ option.label }}
                </option>
              </select>
            </label>

            <label
              v-if="isSuperAdmin"
              class="block text-xs font-medium text-muted"
            >
              Cakupan
              <select
                v-model="scope"
                class="mt-1.5 rounded-lg border border-border bg-surface-subtle px-3 py-2 text-sm text-foreground outline-none transition focus:border-brand focus:bg-surface"
              >
                <option
                  v-for="option in scopeOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ option.label }}
                </option>
              </select>
            </label>

            <label
              v-if="isSuperAdmin"
              class="block text-xs font-medium text-muted"
            >
              Sekolah
              <select
                v-model="schoolFilter"
                class="mt-1.5 max-w-52 rounded-lg border border-border bg-surface-subtle px-3 py-2 text-sm text-foreground outline-none transition focus:border-brand focus:bg-surface"
                :disabled="schoolOptionsLoading"
              >
                <option value="">Semua sekolah</option>
                <option
                  v-for="school in schoolOptions"
                  :key="school.id"
                  :value="school.id"
                >
                  {{ school.name }} ({{ school.code }})
                </option>
              </select>
            </label>

            <label class="block text-xs font-medium text-muted">
              Entity type
              <input
                v-model="entityType"
                type="text"
                placeholder="mis. school_user"
                class="mt-1.5 w-36 rounded-lg border border-border bg-surface-subtle px-3 py-2 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-brand focus:bg-surface"
              />
            </label>

            <label class="block text-xs font-medium text-muted">
              Dari tanggal
              <input
                v-model="dateFrom"
                type="date"
                class="mt-1.5 rounded-lg border border-border bg-surface-subtle px-3 py-2 text-sm text-foreground outline-none transition focus:border-brand focus:bg-surface"
              />
            </label>

            <label class="block text-xs font-medium text-muted">
              Sampai tanggal
              <input
                v-model="dateTo"
                type="date"
                class="mt-1.5 rounded-lg border border-border bg-surface-subtle px-3 py-2 text-sm text-foreground outline-none transition focus:border-brand focus:bg-surface"
              />
            </label>

            <label class="block text-xs font-medium text-muted">
              Actor user ID
              <input
                v-model="actorUserId"
                type="text"
                placeholder="uuid"
                class="mt-1.5 w-40 rounded-lg border border-border bg-surface-subtle px-3 py-2 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-brand focus:bg-surface"
              />
            </label>

            <label class="block text-xs font-medium text-muted">
              Correlation ID
              <input
                v-model="correlationId"
                type="text"
                placeholder="uuid"
                class="mt-1.5 w-40 rounded-lg border border-border bg-surface-subtle px-3 py-2 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-brand focus:bg-surface"
              />
            </label>

            <button
              type="button"
              class="rounded-lg border border-border bg-surface px-3.5 py-2 text-xs font-medium text-foreground-secondary transition hover:border-brand hover:text-brand"
              @click="resetFilters"
            >
              Reset filter
            </button>
          </div>
        </div>

        <div v-if="loading" class="space-y-3 p-5">
          <div
            v-for="item in 5"
            :key="item"
            class="h-12 animate-pulse rounded-lg bg-surface-subtle"
          />
        </div>

        <div v-else-if="errorMessage" class="p-8 text-center">
          <PhWarningCircle
            :size="26"
            class="mx-auto text-danger"
            weight="duotone"
          />
          <h3 class="mt-3 text-sm font-semibold text-foreground">
            Log audit belum bisa dimuat
          </h3>
          <p class="mt-2 text-sm leading-6 text-muted">{{ errorMessage }}</p>
          <button
            type="button"
            class="mt-4 inline-flex items-center justify-center gap-2 rounded-lg border border-border bg-surface px-4 py-2.5 text-sm font-medium text-foreground-secondary transition hover:border-brand hover:text-brand"
            @click="loadLogs()"
          >
            Coba lagi
          </button>
        </div>

        <div v-else-if="items.length === 0" class="p-8 text-center">
          <PhClockCounterClockwise
            class="mx-auto h-7 w-7 text-muted"
            weight="duotone"
          />
          <h3 class="mt-3 text-sm font-semibold text-foreground">
            Belum ada log yang cocok
          </h3>
          <p class="mt-2 text-sm leading-6 text-muted">
            Coba ubah kata kunci atau filter di atas.
          </p>
        </div>

        <div v-else class="overflow-x-auto">
          <table class="w-full min-w-240 text-left text-sm">
            <thead
              class="bg-surface-subtle text-xs uppercase tracking-wide text-muted"
            >
              <tr>
                <th class="px-4 py-3 font-medium">Waktu</th>
                <th class="px-4 py-3 font-medium">Tingkat</th>
                <th class="px-4 py-3 font-medium">Cakupan</th>
                <th class="px-4 py-3 font-medium">Action</th>
                <th class="px-4 py-3 font-medium">Entity</th>
                <th class="px-4 py-3 font-medium">Actor</th>
                <th v-if="isSuperAdmin" class="px-4 py-3 font-medium">
                  Sekolah
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border">
              <tr
                v-for="row in items"
                :key="row.logId"
                class="cursor-pointer transition-colors duration-500 hover:bg-surface-hover"
                :class="highlightedIds.has(row.logId) ? 'bg-brand-soft/60' : ''"
                @click="openDetail(row)"
              >
                <td class="whitespace-nowrap px-4 py-3 text-muted">
                  <span class="inline-flex items-center gap-1.5">
                    {{ formatDateTime(row.createdAt) }}
                    <span
                      v-if="highlightedIds.has(row.logId)"
                      class="rounded-full bg-brand-soft px-1.5 py-0.5 text-[9px] font-semibold uppercase tracking-wide text-brand"
                    >
                      Baru
                    </span>
                  </span>
                </td>
                <td class="px-4 py-3">
                  <span
                    class="rounded-full px-2 py-1 text-[10px] font-semibold"
                    :class="severityBadgeClass(row.severity)"
                  >
                    {{ row.severity || "—" }}
                  </span>
                </td>
                <td class="px-4 py-3">
                  <span
                    class="rounded-full px-2 py-1 text-[10px] font-semibold"
                    :class="scopeBadgeClass(row.scope)"
                  >
                    {{ scopeLabel(row.scope) }}
                  </span>
                </td>
                <td
                  class="px-4 py-3 font-mono text-xs font-medium text-foreground"
                >
                  {{ row.action }}
                </td>
                <td class="px-4 py-3 text-xs text-muted">
                  <span v-if="row.entityType">{{ row.entityType }}</span>
                  <span v-else>—</span>
                </td>
                <td class="min-w-0 px-4 py-3">
                  <p class="truncate text-xs font-medium text-foreground">
                    {{ row.actorName || "Tidak diketahui" }}
                  </p>
                  <p class="truncate text-[11px] text-muted">
                    {{ row.actorEmail }}
                  </p>
                </td>
                <td
                  v-if="isSuperAdmin"
                  class="min-w-0 px-4 py-3 text-xs text-muted"
                >
                  <span v-if="row.schoolName">{{ row.schoolName }}</span>
                  <span v-else>Platform</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div
          v-if="!loading && !errorMessage && items.length > 0"
          class="p-5 pt-0"
        >
          <PaginationBar
            :page="page"
            :total-pages="totalPages"
            :total-items="totalItems"
            :limit="LIMIT"
            @change="(p) => loadLogs(p)"
          />
        </div>
      </div>
    </section>

    <Teleport to="body">
      <Transition
        enter-active-class="transition duration-200 ease-out"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition duration-150 ease-in"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
        <div
          v-if="detailOpen"
          class="fixed inset-0 z-60 flex justify-end"
          role="dialog"
          aria-modal="true"
        >
          <div
            class="absolute inset-0 bg-black/40 backdrop-blur-[2px]"
            aria-hidden="true"
            @click="closeDetail"
          />
          <Transition
            appear
            enter-active-class="transition duration-200 ease-out"
            enter-from-class="translate-x-full"
            enter-to-class="translate-x-0"
            leave-active-class="transition duration-150 ease-in"
            leave-from-class="translate-x-0"
            leave-to-class="translate-x-full"
          >
            <aside
              v-if="detailOpen"
              class="relative flex h-full w-full max-w-lg flex-col bg-surface shadow-2xl shadow-black/20"
            >
              <div class="flex items-start justify-between gap-3 p-5">
                <div class="min-w-0">
                  <p class="eyebrow-muted">Detail log audit</p>
                  <h2
                    class="mt-1 break-all font-mono text-sm font-semibold text-foreground"
                  >
                    {{ detail?.action || "Memuat..." }}
                  </h2>
                </div>
                <button
                  type="button"
                  class="shrink-0 rounded-lg p-1.5 text-muted transition hover:text-foreground"
                  aria-label="Tutup"
                  @click="closeDetail"
                >
                  <PhX :size="18" />
                </button>
              </div>

              <div class="flex-1 overflow-y-auto p-5">
                <div v-if="detailLoading" class="space-y-3">
                  <div
                    v-for="item in 4"
                    :key="item"
                    class="h-8 animate-pulse rounded-lg bg-surface-subtle"
                  />
                </div>

                <div
                  v-else-if="detailError"
                  class="rounded-lg border border-danger-line bg-danger-soft p-4 text-sm text-danger"
                >
                  {{ detailError }}
                </div>

                <div v-else-if="detail" class="space-y-5">
                  <div class="flex flex-wrap gap-2">
                    <span
                      class="rounded-full px-2.5 py-1 text-[11px] font-semibold"
                      :class="severityBadgeClass(detail.severity)"
                    >
                      {{ detail.severity || "—" }}
                    </span>
                    <span
                      class="rounded-full px-2.5 py-1 text-[11px] font-semibold"
                      :class="scopeBadgeClass(detail.scope)"
                    >
                      {{ scopeLabel(detail.scope) }}
                    </span>
                  </div>

                  <dl class="grid grid-cols-2 gap-x-4 gap-y-3 text-xs">
                    <div>
                      <dt class="text-muted">Waktu</dt>
                      <dd class="mt-0.5 font-medium text-foreground">
                        {{ formatDateTime(detail.createdAt) }}
                      </dd>
                    </div>
                    <div>
                      <dt class="text-muted">Entity</dt>
                      <dd class="mt-0.5 break-all font-medium text-foreground">
                        {{ detail.entityType || "—" }}
                        <span
                          v-if="detail.entityId"
                          class="block text-[11px] font-normal text-muted"
                        >
                          {{ detail.entityId }}
                        </span>
                      </dd>
                    </div>
                    <div>
                      <dt class="text-muted">Actor</dt>
                      <dd class="mt-0.5 font-medium text-foreground">
                        {{ detail.actorName || "Tidak diketahui" }}
                        <span class="block text-[11px] font-normal text-muted">
                          {{ detail.actorEmail }}
                        </span>
                      </dd>
                    </div>
                    <div>
                      <dt class="text-muted">Sekolah</dt>
                      <dd class="mt-0.5 font-medium text-foreground">
                        {{ detail.schoolName || "Platform" }}
                      </dd>
                    </div>
                    <div v-if="detail.correlationId" class="col-span-2">
                      <dt class="text-muted">Correlation ID</dt>
                      <dd
                        class="mt-0.5 break-all font-mono text-[11px] text-foreground"
                      >
                        {{ detail.correlationId }}
                      </dd>
                    </div>
                  </dl>

                  <div>
                    <p class="eyebrow-muted">Metadata</p>
                    <div class="mt-2 rounded-lg bg-surface-subtle">
                      <JsonViewer :value="parsedMetadata" />
                    </div>
                  </div>
                </div>
              </div>
            </aside>
          </Transition>
        </div>
      </Transition>
    </Teleport>
  </main>
</template>
