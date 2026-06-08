<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { 
  PhArrowLeft, 
  PhClipboardText, 
  PhFileText, 
  PhPaperPlaneTilt, 
  PhInfo,
  PhCalendarBlank,
  PhClock
} from '@phosphor-icons/vue'
import { useAuthStore } from '../../stores/auth'
import { getMyTeachingSubjectClassById } from '../../services/teacherSubjects'
import { getAssignmentCategories, createAssignment } from '../../services/teacherAssignment'
import { createMaterial } from '../../services/teacherMaterial'
import MediaUploader from '../../components/common/MediaUploader.vue'
import type { TeacherSubjectClass } from '../../types/teacherSubjects'
import type { AssignmentCategory } from '../../types/teacherAssignment'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const subjectClassId = computed(() => String(route.params.subjectClassId ?? ''))
const subject = ref<TeacherSubjectClass | null>(null)
const categories = ref<AssignmentCategory[]>([])
const activeTab = ref<'material' | 'assignment'>('material')
const loading = ref(false)
const submitting = ref(false)
const errorMessage = ref('')
const activeSchoolCode = computed(() => {
  const activeMembership = auth.memberships.find(
    (membership) => membership.school.id === auth.activeSchoolId,
  )
  return activeMembership?.school.code ?? auth.memberships[0]?.school.code ?? ''
})

// Form State
const form = ref({
  title: '',
  description: '',
  materialType: 'pdf' as 'pdf' | 'video' | 'ppt' | 'other',
  categoryId: '',
  deadlineDate: '',
  deadlineTime: '23:59',
  allowLate: false,
  mediaIds: [] as string[]
})

async function loadInitialData() {
  loading.value = true
  try {
    const [subjectData, categoriesData] = await Promise.all([
      getMyTeachingSubjectClassById(subjectClassId.value),
      getAssignmentCategories(activeSchoolCode.value)
    ])
    subject.value = subjectData
    categories.value = categoriesData.data
    
    if (categories.value.length > 0) {
      form.value.categoryId = categories.value[0].asc_id
    }
  } catch (err) {
    errorMessage.value = 'Gagal memuat data pendukung. Coba refresh halaman.'
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  if (!form.value.title.trim()) {
    alert('Judul wajib diisi')
    return
  }

  submitting.value = true
  try {
    if (activeTab.value === 'material') {
      await createMaterial({
        schoolId: auth.activeSchoolId!,
        subjectClassId: subjectClassId.value,
        materialTitle: form.value.title,
        materialDesc: form.value.description,
        materialType: form.value.materialType,
        mediaIds: form.value.mediaIds
      })
    } else {
      let deadline = undefined
      if (form.value.deadlineDate) {
        deadline = `${form.value.deadlineDate}T${form.value.deadlineTime}:00Z`
      }

      await createAssignment({
        schoolId: auth.activeSchoolId!,
        subjectClassId: subjectClassId.value,
        categoryId: form.value.categoryId,
        assignmentTitle: form.value.title,
        assignmentDescription: form.value.description,
        deadline,
        allowLateSubmission: form.value.allowLate,
        mediaIds: form.value.mediaIds
      })
    }
    
    router.push(`/teacher/subjects/${subjectClassId.value}`)
  } catch (err: any) {
    alert(err.response?.data?.error || 'Gagal menyimpan konten')
  } finally {
    submitting.value = false
  }
}

onMounted(loadInitialData)
</script>

<template>
  <main class="min-h-screen flex-1 px-5 py-8 md:px-8 lg:px-10">
    <div class="mx-auto max-w-5xl">
      <!-- Topbar / Breadcrumb -->
      <div class="flex items-center justify-between mb-8">
        <div class="flex items-center gap-4">
          <button 
            @click="router.back()" 
            class="flex items-center gap-2 text-sm font-medium text-[#6B7280] hover:text-[#111827] transition"
          >
            <PhArrowLeft :size="18" />
            <span class="hidden sm:inline">{{ subject?.subjectName || 'Kembali' }}</span>
          </button>
          <span class="text-[#D1D5DB]">/</span>
          <h1 class="text-sm font-semibold text-[#111827]">Buat Konten Baru</h1>
        </div>
        
        <div class="flex items-center gap-3">
          <button 
            @click="router.back()"
            class="px-4 py-2 text-sm font-medium text-[#374151] bg-white border border-[#EBEBEB] rounded-xl hover:bg-[#F9FAFB] transition"
          >
            Batal
          </button>
          <button 
            @click="handleSubmit"
            :disabled="submitting"
            class="flex items-center gap-2 px-5 py-2 text-sm font-medium text-white bg-[#4F46E5] rounded-xl hover:bg-[#4338CA] transition disabled:opacity-50"
          >
            <PhPaperPlaneTilt v-if="!submitting" :size="18" weight="bold" />
            {{ submitting ? 'Menyimpan...' : 'Terbitkan' }}
          </button>
        </div>
      </div>

      <!-- Type Switcher -->
      <div class="flex gap-2 p-1.5 bg-[#F3F4F6] rounded-2xl mb-8 w-fit">
        <button 
          @click="activeTab = 'material'"
          :class="[
            'flex items-center gap-2 px-6 py-2.5 text-sm font-medium rounded-xl transition',
            activeTab === 'material' ? 'bg-white text-[#4F46E5] shadow-sm' : 'text-[#6B7280] hover:text-[#111827]'
          ]"
        >
          <PhFileText :size="18" weight="duotone" />
          Materi
        </button>
        <button 
          @click="activeTab = 'assignment'"
          :class="[
            'flex items-center gap-2 px-6 py-2.5 text-sm font-medium rounded-xl transition',
            activeTab === 'assignment' ? 'bg-white text-[#4F46E5] shadow-sm' : 'text-[#6B7280] hover:text-[#111827]'
          ]"
        >
          <PhClipboardText :size="18" weight="duotone" />
          Tugas
        </button>
      </div>

      <div class="grid gap-8 lg:grid-cols-[1fr_320px]">
        <!-- Main Form -->
        <div class="space-y-6">
          <section class="bg-white rounded-3xl p-6 border border-[#EBEBEB] shadow-sm">
            <h2 class="text-xs font-bold text-[#374151] uppercase tracking-wider mb-6 flex items-center gap-2">
              <PhInfo :size="16" weight="bold" />
              Informasi Utama
            </h2>
            
            <div class="space-y-5">
              <div>
                <label class="block text-sm font-medium text-[#6B7280] mb-2">Judul {{ activeTab === 'material' ? 'Materi' : 'Tugas' }}</label>
                <input 
                  v-model="form.title"
                  type="text" 
                  class="w-full px-4 py-3 bg-[#F9FAFB] border border-[#EBEBEB] rounded-2xl outline-none focus:border-[#4F46E5] transition"
                  placeholder="Contoh: Pengenalan Aljabar Linear"
                />
              </div>

              <div>
                <label class="block text-sm font-medium text-[#6B7280] mb-2">Deskripsi (Opsional)</label>
                <textarea 
                  v-model="form.description"
                  rows="5"
                  class="w-full px-4 py-3 bg-[#F9FAFB] border border-[#EBEBEB] rounded-2xl outline-none focus:border-[#4F46E5] transition resize-none"
                  placeholder="Berikan instruksi atau detail tambahan..."
                ></textarea>
              </div>
            </div>
          </section>

          <section class="bg-white rounded-3xl p-6 border border-[#EBEBEB] shadow-sm">
            <h2 class="text-xs font-bold text-[#374151] uppercase tracking-wider mb-6 flex items-center gap-2">
              <PhFileText :size="16" weight="bold" />
              Lampiran & Media
            </h2>
            
            <MediaUploader 
              :school-id="auth.activeSchoolId!" 
              :owner-type="activeTab"
              @update:media-ids="form.mediaIds = $event"
            />
          </section>
        </div>

        <!-- Sidebar Settings -->
        <aside class="space-y-6">
          <section class="bg-white rounded-3xl p-6 border border-[#EBEBEB] shadow-sm">
            <h2 class="text-xs font-bold text-[#374151] uppercase tracking-wider mb-6">Pengaturan</h2>
            
            <div v-if="activeTab === 'material'" class="space-y-4">
              <div>
                <label class="block text-xs font-medium text-[#6B7280] mb-2">Tipe Materi</label>
                <select 
                  v-model="form.materialType"
                  class="w-full px-3 py-2.5 bg-[#F9FAFB] border border-[#EBEBEB] rounded-xl outline-none text-sm"
                >
                  <option value="pdf">PDF</option>
                  <option value="video">Video</option>
                  <option value="ppt">PPT / Slide</option>
                  <option value="other">Lainnya</option>
                </select>
              </div>
            </div>

            <div v-else class="space-y-5">
              <div>
                <label class="block text-xs font-medium text-[#6B7280] mb-2">Kategori Tugas</label>
                <select 
                  v-model="form.categoryId"
                  class="w-full px-3 py-2.5 bg-[#F9FAFB] border border-[#EBEBEB] rounded-xl outline-none text-sm"
                >
                  <option v-for="cat in categories" :key="cat.asc_id" :value="cat.asc_id">
                    {{ cat.asc_name }}
                  </option>
                </select>
              </div>

              <div>
                <label class="block text-xs font-medium text-[#6B7280] mb-2">Deadline</label>
                <div class="space-y-2">
                  <div class="relative">
                    <PhCalendarBlank :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-[#9CA3AF]" />
                    <input 
                      v-model="form.deadlineDate"
                      type="date" 
                      class="w-full pl-10 pr-3 py-2 bg-[#F9FAFB] border border-[#EBEBEB] rounded-xl outline-none text-sm"
                    />
                  </div>
                  <div class="relative">
                    <PhClock :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-[#9CA3AF]" />
                    <input 
                      v-model="form.deadlineTime"
                      type="time" 
                      class="w-full pl-10 pr-3 py-2 bg-[#F9FAFB] border border-[#EBEBEB] rounded-xl outline-none text-sm"
                    />
                  </div>
                </div>
              </div>

              <div class="pt-2 border-t border-[#F3F4F6]">
                <label class="flex items-center justify-between cursor-pointer group">
                  <div class="space-y-0.5">
                    <p class="text-xs font-medium text-[#374151]">Izinkan Terlambat</p>
                    <p class="text-[10px] text-[#9CA3AF]">Siswa tetap bisa submit</p>
                  </div>
                  <div 
                    @click="form.allowLate = !form.allowLate"
                    :class="[
                      'w-10 h-5 rounded-full relative transition duration-200',
                      form.allowLate ? 'bg-[#4F46E5]' : 'bg-[#E5E7EB]'
                    ]"
                  >
                    <div 
                      :class="[
                        'absolute top-0.5 left-0.5 w-4 h-4 bg-white rounded-full shadow-sm transition transform duration-200',
                        form.allowLate ? 'translate-x-5' : 'translate-x-0'
                      ]"
                    ></div>
                  </div>
                </label>
              </div>
            </div>
          </section>

          <!-- Status Card -->
          <div class="bg-[#FFF7ED] border border-[#FED7AA] rounded-3xl p-5">
            <h3 class="text-xs font-bold text-[#EA580C] uppercase tracking-wider mb-3">Status Publikasi</h3>
            <p class="text-[11px] leading-relaxed text-[#9A3412]">
              Konten ini akan langsung tersedia bagi siswa yang terdaftar di kelas 
              <strong>{{ subject?.className }}</strong> segera setelah diterbitkan.
            </p>
          </div>
        </aside>
      </div>
    </div>
  </main>
</template>
