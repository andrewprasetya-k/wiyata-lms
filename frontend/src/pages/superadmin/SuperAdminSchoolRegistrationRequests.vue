<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import {
  PhArrowClockwise,
  PhCheckCircle,
  PhClipboardText,
  PhCopy,
  PhXCircle,
} from '@phosphor-icons/vue'
import { useToastStore } from '../../stores/toast'
import {
  approveSchoolRegistrationRequest,
  getSchoolRegistrationRequestDetail,
  getSchoolRegistrationRequests,
  rejectSchoolRegistrationRequest,
  type ApproveSchoolRegistrationResponse,
  type SchoolRegistrationRequestDetail,
  type SchoolRegistrationStatus,
} from '../../services/onboarding'

const toast = useToastStore()

const statusTabs: Array<{ label: string; value: SchoolRegistrationStatus }> = [
  { label: 'Pending', value: 'pending' },
  { label: 'Approved', value: 'approved' },
  { label: 'Rejected', value: 'rejected' },
]

const requests = ref<SchoolRegistrationRequestDetail[]>([])
const selectedRequest = ref<SchoolRegistrationRequestDetail | null>(null)
const activeStatus = ref<SchoolRegistrationStatus>('pending')
const page = ref(1)
const limit = 10
const totalPages = ref(1)
const totalItems = ref(0)
const isLoading = ref(false)
const detailLoading = ref(false)
const loadingDetailRequestId = ref('')
const actionLoading = ref(false)
const errorMessage = ref('')
const actionMode = ref<'approve' | 'reject' | null>(null)
const approveResult = ref<ApproveSchoolRegistrationResponse | null>(null)
let detailRequestVersion = 0

const approveForm = reactive({
  schoolCode: '',
  schoolName: '',
  adminName: '',
  adminEmail: '',
  note: '',
})

const rejectForm = reactive({
  reason: '',
})

const selectedIsPending = computed(
  () => selectedRequest.value?.status === 'pending',
)

const invitationLink = computed(() => {
  const acceptUrl = approveResult.value?.invitation.acceptUrl
  if (!acceptUrl) return ''
  return `${window.location.origin}${acceptUrl}`
})

function formatDate(value?: string) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return new Intl.DateTimeFormat('id-ID', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date)
}

function getApiErrorMessage(error: unknown, fallback: string) {
  if (typeof error === 'object' && error !== null && 'response' in error) {
    const response = (
      error as {
        response?: { data?: { error?: unknown; message?: unknown } | string }
      }
    ).response
    if (typeof response?.data === 'string') return response.data
    if (typeof response?.data?.error === 'string') return response.data.error
    if (typeof response?.data?.message === 'string') return response.data.message
  }
  return fallback
}

function statusLabel(status: SchoolRegistrationStatus) {
  if (status === 'approved') return 'Approved'
  if (status === 'rejected') return 'Rejected'
  return 'Pending'
}

function statusClass(status: SchoolRegistrationStatus) {
  if (status === 'approved') return 'bg-[#ecfdf3] text-[#027a48]'
  if (status === 'rejected') return 'bg-[#fff1f0] text-[#b42318]'
  return 'bg-[#fff7ed] text-[#b45309]'
}

function resetForms() {
  actionMode.value = null
  approveResult.value = null
  rejectForm.reason = ''
  if (selectedRequest.value) {
    approveForm.schoolCode = ''
    approveForm.schoolName = selectedRequest.value.schoolName
    approveForm.adminName = selectedRequest.value.picName
    approveForm.adminEmail = selectedRequest.value.picEmail
    approveForm.note = ''
  }
}

async function loadRequests(targetPage = page.value) {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const response = await getSchoolRegistrationRequests({
      status: activeStatus.value,
      page: targetPage,
      limit,
    })
    requests.value = response.data ?? []
    page.value = response.page
    totalPages.value = response.totalPages || 1
    totalItems.value = response.totalItems
  } catch (error) {
    requests.value = []
    errorMessage.value = getApiErrorMessage(
      error,
      'Permintaan pendaftaran belum bisa dimuat.',
    )
  } finally {
    isLoading.value = false
  }
}

async function selectStatus(status: SchoolRegistrationStatus) {
  if (activeStatus.value === status) return
  detailRequestVersion += 1
  loadingDetailRequestId.value = ''
  detailLoading.value = false
  activeStatus.value = status
  page.value = 1
  selectedRequest.value = null
  resetForms()
  await loadRequests(1)
}

async function openDetail(id: string) {
  const requestVersion = ++detailRequestVersion
  loadingDetailRequestId.value = id
  detailLoading.value = true
  approveResult.value = null
  try {
    const detail = await getSchoolRegistrationRequestDetail(id)
    if (
      requestVersion !== detailRequestVersion ||
      loadingDetailRequestId.value !== id
    ) {
      return
    }
    selectedRequest.value = detail
    resetForms()
  } catch (error) {
    if (
      requestVersion === detailRequestVersion &&
      loadingDetailRequestId.value === id
    ) {
      toast.error(
        getApiErrorMessage(error, 'Detail request belum bisa dimuat.'),
      )
    }
  } finally {
    if (
      requestVersion === detailRequestVersion &&
      loadingDetailRequestId.value === id
    ) {
      detailLoading.value = false
      loadingDetailRequestId.value = ''
    }
  }
}

function replaceRequest(updated: SchoolRegistrationRequestDetail) {
  requests.value = requests.value.filter((item) => item.requestId !== updated.requestId)
  if (updated.status === activeStatus.value) {
    requests.value = [updated, ...requests.value]
  }
  selectedRequest.value = updated
}

async function submitApprove() {
  if (!selectedRequest.value || actionLoading.value) return
  if (!approveForm.schoolCode.trim()) {
    toast.error('Kode sekolah wajib diisi.')
    return
  }

  actionLoading.value = true
  approveResult.value = null
  try {
    const response = await approveSchoolRegistrationRequest(
      selectedRequest.value.requestId,
      {
        schoolCode: approveForm.schoolCode.trim(),
        schoolName: approveForm.schoolName.trim() || undefined,
        adminName: approveForm.adminName.trim() || undefined,
        adminEmail: approveForm.adminEmail.trim() || undefined,
        note: approveForm.note.trim() || undefined,
      },
    )
    approveResult.value = response
    replaceRequest(response.request)
    actionMode.value = null
    toast.success('Request pendaftaran sekolah disetujui.')
  } catch (error) {
    toast.error(
      getApiErrorMessage(error, 'Request belum bisa disetujui.'),
    )
  } finally {
    actionLoading.value = false
  }
}

async function submitReject() {
  if (!selectedRequest.value || actionLoading.value) return

  actionLoading.value = true
  approveResult.value = null
  try {
    const response = await rejectSchoolRegistrationRequest(
      selectedRequest.value.requestId,
      { reason: rejectForm.reason.trim() || undefined },
    )
    replaceRequest(response.request)
    actionMode.value = null
    toast.success('Request pendaftaran sekolah ditolak.')
  } catch (error) {
    toast.error(
      getApiErrorMessage(error, 'Request belum bisa ditolak.'),
    )
  } finally {
    actionLoading.value = false
  }
}

async function copyInvitationLink() {
  if (!invitationLink.value) return
  try {
    await navigator.clipboard.writeText(invitationLink.value)
    toast.success('Link undangan disalin.')
  } catch {
    toast.error('Link belum bisa disalin otomatis.')
  }
}

onMounted(() => {
  loadRequests()
})
</script>

<template>
  <main class="min-h-screen min-w-0 flex-1 overflow-x-hidden bg-[#f8f7f4]">
    <header class="border-b border-[#ebe7df] bg-white">
      <div
        class="flex min-w-0 flex-col gap-3 px-5 py-5 sm:px-6 lg:flex-row lg:items-end lg:justify-between lg:px-8"
      >
        <div class="min-w-0">
          <p class="text-xs font-semibold uppercase tracking-[0.18em] text-[#ea580c]">
            Super Admin
          </p>
          <h1 class="mt-2 text-2xl font-semibold text-[#171322] sm:text-3xl">
            Permintaan Pendaftaran Sekolah
          </h1>
          <p class="mt-2 max-w-3xl text-sm leading-6 text-[#6b7280]">
            Review request dari landing page, lalu approve untuk membuat sekolah
            dan token undangan admin.
          </p>
        </div>
        <button
          type="button"
          class="inline-flex w-full items-center justify-center gap-2 rounded-lg border border-[#ebe7df] bg-white px-4 py-2.5 text-sm font-medium text-[#374151] transition hover:border-[#4f46e5] hover:text-[#4f46e5] disabled:cursor-not-allowed disabled:opacity-60 sm:w-auto"
          :disabled="isLoading"
          @click="loadRequests()"
        >
          <PhArrowClockwise :size="16" weight="bold" />
          Muat ulang
        </button>
      </div>
    </header>

    <section
      class="grid w-full max-w-none gap-6 px-5 py-6 sm:px-6 lg:px-8 xl:grid-cols-[minmax(0,1fr)_420px]"
    >
      <section class="min-w-0 rounded-xl border border-[#ebe7df] bg-white p-5 shadow-sm">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <p class="text-xs font-semibold uppercase tracking-[0.16em] text-[#ea580c]">
              Antrian request
            </p>
            <h2 class="mt-2 text-xl font-semibold text-[#171322]">
              Request sekolah
            </h2>
            <p class="mt-1 text-sm text-[#6b7280]">
              {{ totalItems }} request pada status {{ statusLabel(activeStatus) }}.
            </p>
          </div>
          <div class="flex flex-wrap gap-2">
            <button
              v-for="tab in statusTabs"
              :key="tab.value"
              type="button"
              class="rounded-lg border px-3 py-2 text-sm font-semibold transition"
              :class="
                activeStatus === tab.value
                  ? 'border-[#ea580c] bg-[#fff7ed] text-[#c2410c]'
                  : 'border-[#e5e7eb] bg-white text-[#6b7280] hover:text-[#171322]'
              "
              @click="selectStatus(tab.value)"
            >
              {{ tab.label }}
            </button>
          </div>
        </div>

        <div class="mt-5 space-y-3">
          <div
            v-if="isLoading"
            class="rounded-lg border border-[#e5e7eb] bg-[#fafafa] px-4 py-5 text-sm text-[#6b7280]"
          >
            Memuat permintaan pendaftaran...
          </div>

          <div
            v-else-if="errorMessage"
            class="rounded-lg border border-[#fecaca] bg-[#fff8f6] px-4 py-4"
          >
            <p class="text-sm leading-6 text-[#a8665d]">{{ errorMessage }}</p>
            <button
              type="button"
              class="mt-3 inline-flex items-center justify-center gap-2 rounded-lg border border-[#ebe7df] bg-white px-3 py-2 text-sm font-medium text-[#374151] transition hover:border-[#4f46e5] hover:text-[#4f46e5] disabled:cursor-not-allowed disabled:opacity-60"
              @click="loadRequests()"
            >
              Coba lagi
            </button>
          </div>

          <div
            v-else-if="requests.length === 0"
            class="rounded-lg bg-[#fbfaf8] px-5 py-8 text-center"
          >
            <PhClipboardText
              class="mx-auto h-7 w-7 text-[#9ca3af]"
              weight="duotone"
            />
            <h3 class="mt-3 text-sm font-semibold text-[#171322]">
              Belum ada request
            </h3>
            <p class="mt-2 text-sm leading-6 text-[#6b7280]">
              Belum ada request dengan status {{ statusLabel(activeStatus) }}.
            </p>
          </div>

          <article
            v-for="request in requests"
            v-else
            :key="request.requestId"
            class="rounded-xl border border-[#ebe7df] bg-[#fcfbf8] p-4"
          >
            <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
              <div class="min-w-0">
                <div class="flex min-w-0 flex-wrap items-center gap-2">
                  <h3 class="truncate text-base font-semibold text-[#171322]">
                    {{ request.schoolName }}
                  </h3>
                  <span
                    class="rounded-full px-2.5 py-1 text-xs font-semibold"
                    :class="statusClass(request.status)"
                  >
                    {{ statusLabel(request.status) }}
                  </span>
                </div>
                <div class="mt-3 grid gap-1 text-sm leading-6 text-[#6b7280] md:grid-cols-2">
                  <p class="truncate">PIC: {{ request.picName }}</p>
                  <p class="truncate">{{ request.picEmail }}</p>
                  <p class="text-xs text-[#9ca3af] md:col-span-2">
                    Masuk {{ formatDate(request.createdAt) }}
                  </p>
                </div>
              </div>
              <button
                type="button"
                class="inline-flex shrink-0 items-center justify-center gap-2 rounded-lg border border-[#ebe7df] bg-white px-3 py-2 text-sm font-medium text-[#374151] transition hover:border-[#4f46e5] hover:text-[#4f46e5] disabled:cursor-not-allowed disabled:opacity-60"
                :disabled="detailLoading"
                @click="openDetail(request.requestId)"
              >
                {{ loadingDetailRequestId === request.requestId ? 'Memuat...' : 'Detail' }}
              </button>
            </div>
          </article>
        </div>

        <div
          v-if="!isLoading && !errorMessage && totalPages > 1"
          class="mt-5 flex items-center justify-between border-t border-[#ebe7df] pt-4 text-sm"
        >
          <button
            type="button"
            class="inline-flex items-center justify-center gap-2 rounded-lg border border-[#ebe7df] bg-white px-3 py-2 text-sm font-medium text-[#374151] transition hover:border-[#4f46e5] hover:text-[#4f46e5] disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="page <= 1"
            @click="loadRequests(page - 1)"
          >
            Sebelumnya
          </button>
          <span class="text-[#6b7280]">Halaman {{ page }} dari {{ totalPages }}</span>
          <button
            type="button"
            class="inline-flex items-center justify-center gap-2 rounded-lg border border-[#ebe7df] bg-white px-3 py-2 text-sm font-medium text-[#374151] transition hover:border-[#4f46e5] hover:text-[#4f46e5] disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="page >= totalPages"
            @click="loadRequests(page + 1)"
          >
            Berikutnya
          </button>
        </div>
      </section>

      <aside class="min-w-0">
        <section class="rounded-xl border border-[#ebe7df] bg-white p-5 shadow-sm xl:sticky xl:top-6">
          <div class="flex items-start justify-between gap-4">
            <div>
              <p class="text-xs font-semibold uppercase tracking-[0.16em] text-[#ea580c]">
                Detail request
              </p>
              <h2 class="mt-2 text-xl font-semibold text-[#171322]">
                Review pendaftaran
              </h2>
            </div>
            <span class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-[#fff4ee] text-[#ea580c]">
              <PhClipboardText :size="22" weight="duotone" />
            </span>
          </div>

          <div
            v-if="detailLoading"
            class="mt-5 rounded-lg border border-[#e5e7eb] bg-[#fafafa] px-4 py-5 text-sm text-[#6b7280]"
          >
            Memuat detail request...
          </div>

          <div
            v-else-if="!selectedRequest"
            class="mt-5 rounded-lg border border-dashed border-[#d1d5db] bg-[#fafafa] px-4 py-8 text-sm leading-6 text-[#6b7280]"
          >
            Pilih salah satu request untuk melihat detail dan action.
          </div>

          <div v-else class="mt-5 space-y-5">
            <div class="space-y-3 rounded-xl border border-[#ebe7df] bg-[#fcfbf8] p-4">
              <div class="flex flex-wrap items-center gap-2">
                <p class="text-base font-semibold text-[#171322]">
                  {{ selectedRequest.schoolName }}
                </p>
                <span
                  class="rounded-full px-2.5 py-1 text-xs font-semibold"
                  :class="statusClass(selectedRequest.status)"
                >
                  {{ statusLabel(selectedRequest.status) }}
                </span>
              </div>
              <dl class="grid gap-3 text-sm">
                <div>
                  <dt class="text-xs text-[#9ca3af]">NPSN</dt>
                  <dd class="mt-1 text-[#374151]">{{ selectedRequest.npsn || '-' }}</dd>
                </div>
                <div>
                  <dt class="text-xs text-[#9ca3af]">PIC</dt>
                  <dd class="mt-1 text-[#374151]">
                    {{ selectedRequest.picName }} · {{ selectedRequest.picEmail }}
                  </dd>
                </div>
                <div>
                  <dt class="text-xs text-[#9ca3af]">Kontak</dt>
                  <dd class="mt-1 text-[#374151]">
                    {{ selectedRequest.picPhone || '-' }} · {{ selectedRequest.picRole || '-' }}
                  </dd>
                </div>
                <div>
                  <dt class="text-xs text-[#9ca3af]">Pesan</dt>
                  <dd class="mt-1 whitespace-pre-wrap text-[#374151]">
                    {{ selectedRequest.message || '-' }}
                  </dd>
                </div>
                <div>
                  <dt class="text-xs text-[#9ca3af]">Dibuat</dt>
                  <dd class="mt-1 text-[#374151]">{{ formatDate(selectedRequest.createdAt) }}</dd>
                </div>
                <div v-if="selectedRequest.reviewedAt || selectedRequest.reviewNote">
                  <dt class="text-xs text-[#9ca3af]">Review</dt>
                  <dd class="mt-1 text-[#374151]">
                    {{ formatDate(selectedRequest.reviewedAt) }}
                    <span v-if="selectedRequest.reviewNote">
                      · {{ selectedRequest.reviewNote }}
                    </span>
                  </dd>
                </div>
              </dl>
            </div>

            <div
              v-if="approveResult"
              class="rounded-xl border border-[#bbf7d0] bg-[#f0fdf4] p-4"
            >
              <div class="flex items-start gap-3">
                <PhCheckCircle :size="22" class="mt-0.5 shrink-0 text-[#027a48]" weight="duotone" />
                <div class="min-w-0">
                  <p class="text-sm font-semibold text-[#166534]">
                    Request disetujui
                  </p>
                  <p class="mt-1 text-xs leading-5 text-[#166534]">
                    Bagikan link undangan ini secara manual ke PIC. Email otomatis belum aktif.
                  </p>
                </div>
              </div>
              <div class="mt-4 rounded-lg border border-[#bbf7d0] bg-white p-3">
                <p class="break-all text-xs leading-5 text-[#166534]">
                  {{ invitationLink }}
                </p>
              </div>
              <button
                type="button"
                class="mt-3 inline-flex items-center justify-center gap-2 rounded-lg border border-[#ebe7df] bg-white px-3 py-2 text-sm font-medium text-[#374151] transition hover:border-[#4f46e5] hover:text-[#4f46e5] disabled:cursor-not-allowed disabled:opacity-60"
                @click="copyInvitationLink"
              >
                <PhCopy :size="16" weight="bold" />
                Copy Link
              </button>
            </div>

            <div v-if="selectedIsPending" class="flex flex-wrap gap-2">
              <button
                type="button"
                class="inline-flex items-center justify-center gap-2 rounded-lg bg-[#16a34a] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#15803d] disabled:cursor-not-allowed disabled:opacity-60"
                @click="actionMode = actionMode === 'approve' ? null : 'approve'"
              >
                <PhCheckCircle :size="16" weight="bold" />
                Approve
              </button>
              <button
                type="button"
                class="inline-flex items-center justify-center gap-2 rounded-lg border border-[#fecaca] bg-white px-3 py-2 text-sm font-medium text-[#dc2626] transition hover:bg-[#fef2f2] disabled:cursor-not-allowed disabled:opacity-60"
                @click="actionMode = actionMode === 'reject' ? null : 'reject'"
              >
                <PhXCircle :size="16" weight="bold" />
                Reject
              </button>
            </div>

            <form
              v-if="selectedIsPending && actionMode === 'approve'"
              class="space-y-4 rounded-xl border border-[#bbf7d0] bg-[#f6fef9] p-4"
              @submit.prevent="submitApprove"
            >
              <p class="text-sm font-semibold text-[#171322]">Approve request</p>
              <label class="block text-sm font-medium text-[#374151]">
                Kode sekolah
                <input
                  v-model="approveForm.schoolCode"
                  class="mt-2 w-full rounded-lg border border-[#d1d5db] bg-white px-3 py-2.5 text-sm outline-none focus:border-[#027a48]"
                  placeholder="SMWM"
                />
              </label>
              <label class="block text-sm font-medium text-[#374151]">
                Nama sekolah
                <input
                  v-model="approveForm.schoolName"
                  class="mt-2 w-full rounded-lg border border-[#d1d5db] bg-white px-3 py-2.5 text-sm outline-none focus:border-[#027a48]"
                />
              </label>
              <label class="block text-sm font-medium text-[#374151]">
                Nama admin
                <input
                  v-model="approveForm.adminName"
                  class="mt-2 w-full rounded-lg border border-[#d1d5db] bg-white px-3 py-2.5 text-sm outline-none focus:border-[#027a48]"
                />
              </label>
              <label class="block text-sm font-medium text-[#374151]">
                Email admin
                <input
                  v-model="approveForm.adminEmail"
                  type="email"
                  class="mt-2 w-full rounded-lg border border-[#d1d5db] bg-white px-3 py-2.5 text-sm outline-none focus:border-[#027a48]"
                />
              </label>
              <label class="block text-sm font-medium text-[#374151]">
                Catatan
                <textarea
                  v-model="approveForm.note"
                  rows="3"
                  class="mt-2 w-full resize-none rounded-lg border border-[#d1d5db] bg-white px-3 py-2.5 text-sm outline-none focus:border-[#027a48]"
                  placeholder="Opsional"
                />
              </label>
              <button
                type="submit"
                class="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-[#16a34a] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#15803d] disabled:cursor-not-allowed disabled:opacity-60"
                :disabled="actionLoading"
              >
                {{ actionLoading ? 'Memproses...' : 'Approve dan buat undangan' }}
              </button>
            </form>

            <form
              v-if="selectedIsPending && actionMode === 'reject'"
              class="space-y-4 rounded-xl border border-[#fecaca] bg-[#fff8f6] p-4"
              @submit.prevent="submitReject"
            >
              <p class="text-sm font-semibold text-[#171322]">Reject request</p>
              <label class="block text-sm font-medium text-[#374151]">
                Alasan
                <textarea
                  v-model="rejectForm.reason"
                  rows="4"
                  class="mt-2 w-full resize-none rounded-lg border border-[#fecaca] bg-white px-3 py-2.5 text-sm outline-none focus:border-[#b42318]"
                  placeholder="Opsional"
                />
              </label>
              <button
                type="submit"
                class="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-[#dc2626] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[#b91c1c] disabled:cursor-not-allowed disabled:opacity-60"
                :disabled="actionLoading"
              >
                {{ actionLoading ? 'Memproses...' : 'Reject request' }}
              </button>
            </form>
          </div>
        </section>
      </aside>
    </section>
  </main>
</template>
