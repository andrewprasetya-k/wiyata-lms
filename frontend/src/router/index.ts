import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import type { RoleName } from '../types/auth'
import AuthLayout from '../layouts/AuthLayout.vue'
import StudentLayout from '../layouts/StudentLayout.vue'
import TeacherLayout from '../layouts/TeacherLayout.vue'
import AdminLayout from '../layouts/AdminLayout.vue'
import SuperAdminLayout from '../layouts/SuperAdminLayout.vue'
import LoginPage from '../pages/auth/LoginPage.vue'
import UnauthorizedPage from '../pages/auth/UnauthorizedPage.vue'
import StudentDashboard from '../pages/student/StudentDashboard.vue'
import StudentFeed from '../pages/student/StudentFeed.vue'
import StudentSubjectDetail from '../pages/student/StudentSubjectDetail.vue'
import StudentAssignmentDetail from '../pages/student/StudentAssignmentDetail.vue'
import StudentSubjects from '../pages/student/StudentSubjects.vue'
import TeacherDashboard from '../pages/teacher/TeacherDashboard.vue'
import AdminDashboard from '../pages/admin/AdminDashboard.vue'
import SuperAdminDashboard from '../pages/superadmin/SuperAdminDashboard.vue'
import FeaturePlaceholder from '../components/common/FeaturePlaceholder.vue'

export const dashboardByRole: Record<RoleName, string> = {
  super_admin: '/superadmin/dashboard',
  admin: '/admin/dashboard',
  teacher: '/teacher/dashboard',
  student: '/student/dashboard',
}

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/login',
    },
    {
      path: '/',
      component: AuthLayout,
      children: [
        {
          path: 'login',
          name: 'login',
          component: LoginPage,
        },
        {
          path: 'unauthorized',
          name: 'unauthorized',
          component: UnauthorizedPage,
        },
      ],
    },
    {
      path: '/student',
      component: StudentLayout,
      meta: { requiresAuth: true, roles: ['student'] },
      children: [
        {
          path: 'dashboard',
          name: 'student-dashboard',
          component: StudentDashboard,
        },
        {
          path: 'classes',
          redirect: '/student/subjects',
        },
        {
          path: 'classes/:classId',
          redirect: '/student/subjects',
        },
        {
          path: 'subjects',
          name: 'student-subjects',
          component: StudentSubjects,
        },
        {
          path: 'subjects/:sclId',
          name: 'student-subject-detail',
          component: StudentSubjectDetail,
        },
        {
          path: 'subjects/:sclId/assignments/:asgId',
          name: 'student-assignment-detail',
          component: StudentAssignmentDetail,
        },
        {
          path: 'feed',
          name: 'student-feed',
          component: StudentFeed,
        },
        {
          path: 'assignments',
          name: 'student-assignments',
          component: FeaturePlaceholder,
          props: {
            eyebrow: 'Student assignments',
            title: 'Daftar tugas siswa',
            description:
              'Halaman ini nantinya membantu siswa melihat tugas aktif, deadline, status submit, dan hasil penilaian.',
          },
        },
        {
          path: 'grades',
          name: 'student-grades',
          component: FeaturePlaceholder,
          props: {
            eyebrow: 'Student grades',
            title: 'Nilai dan progress akademik',
            description:
              'Halaman ini nantinya merangkum nilai per subject, rata-rata, dan progress belajar dari grade book backend.',
          },
        },
        {
          path: 'chat',
          name: 'student-chat',
          component: FeaturePlaceholder,
          props: {
            eyebrow: 'Realtime chat',
            title: 'Chat sedang dikembangkan',
            description:
              'Chat akan menjadi fitur DM, group, atau subject realtime. WebSocket belum diimplementasikan pada tahap ini.',
          },
        },
        {
          path: 'notes',
          name: 'student-notes',
          component: FeaturePlaceholder,
          props: {
            eyebrow: 'Learning notes',
            title: 'Catatan belajar',
            description:
              'Notes akan tersedia dari halaman detail material, termasuk catatan per materi. Autosave belum diimplementasikan pada tahap ini.',
          },
        },
        {
          path: 'profile',
          name: 'student-profile',
          component: FeaturePlaceholder,
          props: {
            eyebrow: 'Student profile',
            title: 'Profil siswa',
            description:
              'Halaman ini nantinya menampilkan informasi akun, konteks sekolah aktif, dan preferensi dasar pengguna.',
          },
        },
      ],
    },
    {
      path: '/teacher',
      component: TeacherLayout,
      meta: { requiresAuth: true, roles: ['teacher'] },
      children: [
        {
          path: 'dashboard',
          name: 'teacher-dashboard',
          component: TeacherDashboard,
        },
        {
          path: 'classes',
          name: 'teacher-classes',
          component: FeaturePlaceholder,
          props: {
            eyebrow: 'Teacher classes',
            title: 'Kelas yang diajar',
            description:
              'Halaman ini nantinya menampilkan kelas guru, subject class, material, dan aktivitas siswa.',
          },
        },
        {
          path: 'assignments',
          name: 'teacher-assignments',
          component: FeaturePlaceholder,
          props: {
            eyebrow: 'Teacher assignments',
            title: 'Manajemen tugas',
            description:
              'Halaman ini nantinya dipakai guru untuk membuat, mengubah, dan memantau tugas kelas.',
          },
        },
        {
          path: 'submissions',
          name: 'teacher-submissions',
          component: FeaturePlaceholder,
          props: {
            eyebrow: 'Submission review',
            title: 'Review submission',
            description:
              'Halaman ini nantinya menampilkan submission siswa yang perlu dinilai dan feedback yang sudah diberikan.',
          },
        },
        {
          path: 'chat',
          name: 'teacher-chat',
          component: FeaturePlaceholder,
          props: {
            eyebrow: 'Realtime chat',
            title: 'Chat kelas',
            description:
              'Fitur chat guru akan memakai WebSocket untuk komunikasi realtime dengan kelas. WebSocket belum diimplementasikan pada tahap ini.',
          },
        },
        {
          path: 'profile',
          name: 'teacher-profile',
          component: FeaturePlaceholder,
          props: {
            eyebrow: 'Teacher profile',
            title: 'Profil guru',
            description:
              'Halaman ini nantinya menampilkan informasi akun guru, role aktif, dan konteks sekolah.',
          },
        },
      ],
    },
    {
      path: '/admin',
      component: AdminLayout,
      meta: { requiresAuth: true, roles: ['admin'] },
      children: [
        {
          path: 'dashboard',
          name: 'admin-dashboard',
          component: AdminDashboard,
        },
        {
          path: 'classes',
          name: 'admin-classes',
          component: FeaturePlaceholder,
          props: {
            eyebrow: 'School classes',
            title: 'Manajemen kelas',
            description:
              'Halaman ini nantinya dipakai admin sekolah untuk mengelola kelas, subject class, dan struktur kelas.',
          },
        },
        {
          path: 'users',
          name: 'admin-users',
          component: FeaturePlaceholder,
          props: {
            eyebrow: 'School users',
            title: 'Manajemen user sekolah',
            description:
              'Halaman ini nantinya dipakai admin untuk mengelola user, membership sekolah, dan role.',
          },
        },
        {
          path: 'enrollments',
          name: 'admin-enrollments',
          component: FeaturePlaceholder,
          props: {
            eyebrow: 'Enrollments',
            title: 'Enrollment kelas',
            description:
              'Halaman ini nantinya mengatur siswa dan guru yang tergabung dalam kelas tertentu.',
          },
        },
        {
          path: 'academic-years',
          name: 'admin-academic-years',
          component: FeaturePlaceholder,
          props: {
            eyebrow: 'Academic setup',
            title: 'Tahun ajaran dan semester',
            description:
              'Halaman ini nantinya mengelola academic year, term, dan status aktif periode akademik.',
          },
        },
        {
          path: 'profile',
          name: 'admin-profile',
          component: FeaturePlaceholder,
          props: {
            eyebrow: 'Admin profile',
            title: 'Profil admin sekolah',
            description:
              'Halaman ini nantinya menampilkan informasi akun admin dan konteks sekolah aktif.',
          },
        },
      ],
    },
    {
      path: '/superadmin',
      component: SuperAdminLayout,
      meta: { requiresAuth: true, roles: ['super_admin'] },
      children: [
        {
          path: 'dashboard',
          name: 'superadmin-dashboard',
          component: SuperAdminDashboard,
        },
        {
          path: 'schools',
          name: 'superadmin-schools',
          component: FeaturePlaceholder,
          props: {
            eyebrow: 'Platform schools',
            title: 'Manajemen sekolah',
            description:
              'Halaman ini nantinya mengelola tenant sekolah, status sekolah, dan konfigurasi dasar platform.',
          },
        },
        {
          path: 'users',
          name: 'superadmin-users',
          component: FeaturePlaceholder,
          props: {
            eyebrow: 'Platform users',
            title: 'User platform',
            description:
              'Halaman ini nantinya membantu super admin melihat user lintas sekolah dan setup role platform.',
          },
        },
        {
          path: 'profile',
          name: 'superadmin-profile',
          component: FeaturePlaceholder,
          props: {
            eyebrow: 'Super admin profile',
            title: 'Profil super admin',
            description:
              'Halaman ini nantinya menampilkan informasi akun dan akses platform yang sedang aktif.',
          },
        },
      ],
    },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  auth.restoreSession()

  if (to.name === 'login' && auth.isAuthenticated) {
    const role = auth.primaryRole()
    return role ? dashboardByRole[role] : '/unauthorized'
  }

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  const requiredRoles = to.matched.flatMap((record) => record.meta.roles ?? [])
  if (requiredRoles.length > 0 && !auth.hasAnyRole(requiredRoles)) {
    return { name: 'unauthorized' }
  }

  return true
})

export default router
