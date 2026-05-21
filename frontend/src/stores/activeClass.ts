import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { getStudentClasses } from '../services/studentClasses'
import type { StudentClassEnrollment } from '../types/studentClasses'

const ACTIVE_CLASS_KEY = 'edv_active_class_id'

export const useActiveClassStore = defineStore('activeClass', () => {
  const classes = ref<StudentClassEnrollment[]>([])
  const activeClassId = ref<string | null>(localStorage.getItem(ACTIVE_CLASS_KEY))
  const isLoading = ref(false)
  const errorMessage = ref('')
  const loadedForSchoolUserId = ref<string | null>(null)

  const activeClass = computed(
    () => classes.value.find((item) => item.classId === activeClassId.value) ?? null,
  )
  const activeClassTitle = computed(() => activeClass.value?.classTitle ?? '')
  const hasClasses = computed(() => classes.value.length > 0)

  function persistActiveClass(classId: string | null) {
    activeClassId.value = classId
    if (classId) {
      localStorage.setItem(ACTIVE_CLASS_KEY, classId)
    } else {
      localStorage.removeItem(ACTIVE_CLASS_KEY)
    }
  }

  function setActiveClass(classId: string) {
    const exists = classes.value.some((item) => item.classId === classId)
    if (!exists) return
    persistActiveClass(classId)
  }

  async function loadClasses(schoolUserId: string, options: { force?: boolean } = {}) {
    if (!schoolUserId) {
      classes.value = []
      persistActiveClass(null)
      loadedForSchoolUserId.value = null
      errorMessage.value = 'Konteks sekolah belum tersedia.'
      return
    }

    if (
      !options.force &&
      loadedForSchoolUserId.value === schoolUserId &&
      (classes.value.length > 0 || errorMessage.value)
    ) {
      return
    }

    isLoading.value = true
    errorMessage.value = ''

    try {
      const nextClasses = await getStudentClasses(schoolUserId)
      classes.value = nextClasses
      loadedForSchoolUserId.value = schoolUserId

      const storedClassId = localStorage.getItem(ACTIVE_CLASS_KEY)
      const validStoredClass = nextClasses.find((item) => item.classId === storedClassId)
      const fallbackClass = nextClasses[0]
      persistActiveClass(validStoredClass?.classId ?? fallbackClass?.classId ?? null)
    } catch {
      errorMessage.value = 'Daftar kelas belum bisa dimuat. Periksa koneksi atau coba lagi nanti.'
    } finally {
      isLoading.value = false
    }
  }

  function reset() {
    classes.value = []
    persistActiveClass(null)
    isLoading.value = false
    errorMessage.value = ''
    loadedForSchoolUserId.value = null
  }

  return {
    classes,
    activeClassId,
    activeClass,
    activeClassTitle,
    hasClasses,
    isLoading,
    errorMessage,
    loadClasses,
    setActiveClass,
    reset,
  }
})
