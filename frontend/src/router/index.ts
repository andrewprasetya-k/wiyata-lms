import { createRouter, createWebHistory } from "vue-router";
import { useAuthStore } from "../stores/auth";
import type { RoleName } from "../types/auth";
import AuthLayout from "../layouts/AuthLayout.vue";
import StudentLayout from "../layouts/StudentLayout.vue";
import TeacherLayout from "../layouts/TeacherLayout.vue";
import AdminLayout from "../layouts/AdminLayout.vue";
import SuperAdminLayout from "../layouts/SuperAdminLayout.vue";
import LoginPage from "../pages/auth/LoginPage.vue";
import RegisterPage from "../pages/auth/RegisterPage.vue";
import UnauthorizedPage from "../pages/auth/UnauthorizedPage.vue";
import OnboardingPage from "../pages/onboarding/OnboardingPage.vue";
import StudentDashboard from "../pages/student/StudentDashboard.vue";
import StudentFeed from "../pages/student/StudentFeed.vue";
import StudentSubjectDetail from "../pages/student/StudentSubjectDetail.vue";
import StudentAssignmentDetail from "../pages/student/StudentAssignmentDetail.vue";
import StudentAssignments from "../pages/student/StudentAssignments.vue";
import StudentMaterialDetail from "../pages/student/StudentMaterialDetail.vue";
import StudentMaterialNoteEditor from "../pages/student/StudentMaterialNoteEditor.vue";
import StudentNotes from "../pages/student/StudentNotes.vue";
import StudentSubjects from "../pages/student/StudentSubjects.vue";
import StudentGrades from "../pages/student/StudentGrades.vue";
import StudentChat from "../pages/student/StudentChat.vue";
import StudentActivity from "../pages/student/StudentActivity.vue";
import TeacherDashboard from "../pages/teacher/TeacherDashboard.vue";
import TeacherSubjectDetail from "../pages/teacher/TeacherSubjectDetail.vue";
import TeacherClassGradeReport from "../pages/teacher/TeacherClassGradeReport.vue";
import TeacherStudentGradeDetail from "../pages/teacher/TeacherStudentGradeDetail.vue";
import TeacherStudentReport from "../pages/teacher/TeacherStudentReport.vue";
import TeacherSubjects from "../pages/teacher/TeacherSubjects.vue";
import TeacherActivity from "../pages/teacher/TeacherActivity.vue";
import AdminDashboard from "../pages/admin/AdminDashboard.vue";
import AdminAcademicYears from "../pages/admin/AdminAcademicYears.vue";
import AdminClasses from "../pages/admin/AdminClasses.vue";
import AdminClassDetail from "../pages/admin/AdminClassDetail.vue";
import AdminEnrollments from "../pages/admin/AdminEnrollments.vue";
import AdminSubjectClasses from "../pages/admin/AdminSubjectClasses.vue";
import AdminUsers from "../pages/admin/AdminUsers.vue";
import AdminChat from "../pages/admin/AdminChat.vue";
import SuperAdminDashboard from "../pages/superadmin/SuperAdminDashboard.vue";
import SuperAdminSchools from "../pages/superadmin/SuperAdminSchools.vue";
import SuperAdminUsers from "../pages/superadmin/SuperAdminUsers.vue";
import ReadProfile from "../pages/profile/ReadProfile.vue";
import TeacherCreate from "../pages/teacher/TeacherCreate.vue";
import TeacherContentCreate from "../pages/teacher/TeacherContentCreate.vue";
import TeacherAssignmentReview from "../pages/teacher/TeacherAssignmentReview.vue";
import TeacherAssignmentDetail from "../pages/teacher/TeacherAssignmentDetail.vue";
import TeacherMaterialDetail from "../pages/teacher/TeacherMaterialDetail.vue";
import TeacherSubmissions from "../pages/teacher/TeacherSubmissions.vue";
import TeacherAssignments from "../pages/teacher/TeacherAssignments.vue";
import TeacherFeed from "../pages/teacher/TeacherFeed.vue";
import TeacherChat from "../pages/teacher/TeacherChat.vue";
import NotificationCenter from "../components/notifications/NotificationCenter.vue";
import HomePage from "../pages/preview/HomePage.vue";
import CreateSchool from "../pages/onboarding/CreateSchool.vue";
import AcceptInvitation from "../pages/public/AcceptInvitation.vue";
import VerifyEmail from "../pages/public/VerifyEmail.vue";
import NotFoundPage from "../pages/common/NotFoundPage.vue";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/",
      redirect: "/home",
    },
    {
      path: "/home",
      name: "home",
      component: HomePage,
      meta: { title: "Satu workspace untuk aktivitas akademik sekolah." },
    },
    {
      path: "/create-school",
      name: "create-school",
      component: CreateSchool,
      meta: { requiresAuth: true, title: "Buat Sekolah" },
    },
    {
      path: "/invite/:token",
      name: "accept-invitation",
      component: AcceptInvitation,
      meta: { title: "Terima Undangan" },
    },
    {
      path: "/verify-email",
      name: "verify-email",
      component: VerifyEmail,
      meta: { title: "Verifikasi Email" },
    },
    {
      path: "/onboarding",
      name: "onboarding",
      component: OnboardingPage,
      meta: { requiresAuth: true, title: "Selamat Datang" },
    },
    {
      path: "/",
      component: AuthLayout,
      children: [
        {
          path: "login",
          name: "login",
          component: LoginPage,
          meta: { title: "Login" },
        },
        {
          path: "register",
          name: "register",
          component: RegisterPage,
          meta: { title: "Daftar" },
        },
        {
          path: "unauthorized",
          name: "unauthorized",
          component: UnauthorizedPage,
          meta: { title: "Akses Ditolak" },
        },
      ],
    },
    {
      path: "/student",
      component: StudentLayout,
      meta: { requiresAuth: true, roles: ["student"] },
      children: [
        {
          path: "dashboard",
          name: "student-dashboard",
          component: StudentDashboard,
          meta: { title: "Dashboard Siswa" },
        },
        {
          path: "classes",
          redirect: "/student/subjects",
        },
        {
          path: "classes/:classId",
          redirect: "/student/subjects",
        },
        {
          path: "subjects",
          name: "student-subjects",
          component: StudentSubjects,
          meta: { title: "Mata Pelajaran" },
        },
        {
          path: "subjects/:sclId",
          name: "student-subject-detail",
          component: StudentSubjectDetail,
          meta: { title: "Detail Mata Pelajaran" },
        },
        {
          path: "subjects/:sclId/materials/:matId",
          name: "student-material-detail",
          component: StudentMaterialDetail,
          meta: { title: "Materi" },
        },
        {
          path: "subjects/:sclId/materials/:matId/note",
          name: "student-material-note",
          component: StudentMaterialNoteEditor,
          meta: { title: "Catatan Materi" },
        },
        {
          path: "subjects/:sclId/assignments/:asgId",
          name: "student-assignment-detail",
          component: StudentAssignmentDetail,
          meta: { title: "Detail Tugas" },
        },
        {
          path: "feed",
          name: "student-feed",
          component: StudentFeed,
          meta: { title: "Feed Kelas" },
        },
        {
          path: "activity",
          name: "student-activity",
          component: StudentActivity,
          meta: { title: "Aktivitas Akademik" },
        },
        {
          path: "notifications",
          name: "student-notifications",
          component: NotificationCenter,
          meta: { title: "Notifikasi" },
        },
        {
          path: "assignments",
          name: "student-assignments",
          component: StudentAssignments,
          meta: { title: "Tugas Saya" },
        },
        {
          path: "grades",
          name: "student-grades",
          component: StudentGrades,
          meta: { title: "Nilai" },
        },
        {
          path: "chat",
          name: "student-chat",
          component: StudentChat,
          meta: { title: "Chat" },
        },
        {
          path: "notes",
          name: "student-notes",
          component: StudentNotes,
          meta: { title: "Catatan Saya" },
        },
        {
          path: "profile",
          name: "student-profile",
          component: ReadProfile,
          meta: { title: "Profil Siswa" },
          props: {
            eyebrow: "Student profile",
            title: "Profil Student",
            helper:
              "Lihat informasi akun, role, dan konteks sekolah aktif. Profil masih read-only untuk MVP.",
          },
        },
      ],
    },
    {
      path: "/teacher",
      component: TeacherLayout,
      meta: { requiresAuth: true, roles: ["teacher"] },
      children: [
        {
          path: "dashboard",
          name: "teacher-dashboard",
          component: TeacherDashboard,
          meta: { title: "Dashboard Guru" },
        },
        {
          path: "classes",
          redirect: "/teacher/subjects",
        },
        {
          path: "subjects",
          name: "teacher-subjects",
          component: TeacherSubjects,
          meta: { title: "Mata Pelajaran" },
        },
        {
          path: "subjects/:subjectClassId",
          name: "teacher-subject-detail",
          component: TeacherSubjectDetail,
          meta: { title: "Ruang Mengajar" },
        },
        {
          path: "subjects/:subjectClassId/materials/:matId",
          name: "teacher-material-detail",
          component: TeacherMaterialDetail,
          meta: { title: "Detail Materi" },
        },
        {
          path: "subjects/:subjectClassId/materials/:matId/edit",
          name: "teacher-material-edit",
          component: TeacherContentCreate,
          meta: { title: "Edit Materi" },
        },
        {
          path: "subjects/:subjectClassId/assignments/:assignmentId",
          name: "teacher-assignment-detail",
          component: TeacherAssignmentDetail,
          meta: { title: "Detail Tugas" },
        },
        {
          path: "subjects/:subjectClassId/assignments/:asgId/edit",
          name: "teacher-assignment-edit",
          component: TeacherContentCreate,
          meta: { title: "Edit Tugas" },
        },
        {
          path: "subjects/:subjectClassId/create",
          name: "teacher-content-create",
          component: TeacherContentCreate,
          meta: { title: "Buat Konten" },
        },
        {
          path: "grades/class/:classId/subject/:subjectId",
          name: "teacher-class-grade-report",
          component: TeacherClassGradeReport,
          meta: { title: "Nilai Kelas" },
        },
        {
          path: "grades/class/:classId/subject/:subjectId/student/:studentId",
          name: "teacher-student-grade-detail",
          component: TeacherStudentGradeDetail,
          meta: { title: "Detail Nilai Siswa" },
        },
        {
          path: "grades/class/:classId/student/:studentId/report",
          name: "teacher-student-report",
          component: TeacherStudentReport,
          meta: { title: "Rapor Siswa" },
        },
        {
          path: "assignments/:assignmentId/review",
          name: "teacher-assignment-review",
          component: TeacherAssignmentReview,
          meta: { title: "Review Tugas" },
        },
        {
          path: "assignments",
          name: "teacher-assignments",
          component: TeacherAssignments,
          meta: { title: "Tugas" },
        },
        {
          path: "submissions",
          name: "teacher-submissions",
          component: TeacherSubmissions,
          meta: { title: "Pengumpulan" },
        },
        {
          path: "create",
          name: "teacher-create",
          component: TeacherCreate,
          meta: { title: "Pilih Mata Pelajaran" },
        },
        {
          path: "feed",
          name: "teacher-feed",
          component: TeacherFeed,
          meta: { title: "Feed Kelas" },
        },
        {
          path: "activity",
          name: "teacher-activity",
          component: TeacherActivity,
          meta: { title: "Aktivitas Akademik" },
        },
        {
          path: "notifications",
          name: "teacher-notifications",
          component: NotificationCenter,
          meta: { title: "Notifikasi" },
        },
        {
          path: "chat",
          name: "teacher-chat",
          component: TeacherChat,
          meta: { title: "Chat" },
        },
        {
          path: "profile",
          name: "teacher-profile",
          component: ReadProfile,
          meta: { title: "Profil Guru" },
          props: {
            eyebrow: "Teacher profile",
            title: "Profil Teacher",
            helper:
              "Lihat informasi akun guru, role aktif, dan membership sekolah dari sesi login saat ini.",
          },
        },
      ],
    },
    {
      path: "/admin",
      component: AdminLayout,
      meta: { requiresAuth: true, roles: ["admin"] },
      children: [
        {
          path: "dashboard",
          name: "admin-dashboard",
          component: AdminDashboard,
          meta: { title: "Dashboard Admin" },
        },
        {
          path: "classes",
          name: "admin-classes",
          component: AdminClasses,
          meta: { title: "Kelas" },
        },
        {
          path: "classes/:classId",
          name: "admin-class-detail",
          component: AdminClassDetail,
          meta: { title: "Detail Kelas" },
        },
        {
          path: "users",
          name: "admin-users",
          component: AdminUsers,
          meta: { title: "Warga Sekolah" },
        },
        {
          path: "enrollments",
          name: "admin-enrollments",
          component: AdminEnrollments,
          meta: { title: "Penempatan Kelas" },
        },
        {
          path: "subject-classes",
          name: "admin-subject-classes",
          component: AdminSubjectClasses,
          meta: { title: "Penugasan Mengajar" },
        },
        {
          path: "academic-years",
          name: "admin-academic-years",
          component: AdminAcademicYears,
          meta: { title: "Struktur Akademik" },
        },
        {
          path: "chat",
          name: "admin-chat",
          component: AdminChat,
          meta: { title: "Chat Sekolah" },
        },
        {
          path: "profile",
          name: "admin-profile",
          component: ReadProfile,
          meta: { title: "Profil Admin" },
          props: {
            eyebrow: "Admin profile",
            title: "Profil Admin",
            helper:
              "Lihat informasi akun admin sekolah dan active school context. Perubahan profil belum tersedia.",
          },
        },
      ],
    },
    {
      path: "/superadmin",
      component: SuperAdminLayout,
      meta: { requiresAuth: true, roles: ["super_admin"] },
      children: [
        {
          path: "dashboard",
          name: "superadmin-dashboard",
          component: SuperAdminDashboard,
          meta: { title: "Pusat Platform" },
        },
        {
          path: "schools",
          name: "superadmin-schools",
          component: SuperAdminSchools,
          meta: { title: "Sekolah" },
        },
        {
          path: "users",
          name: "superadmin-users",
          component: SuperAdminUsers,
          meta: { title: "Akun Global" },
        },
        {
          path: "profile",
          name: "superadmin-profile",
          component: ReadProfile,
          meta: { title: "Profil Super Admin" },
          props: {
            eyebrow: "Profil Super Admin",
            title: "Profil Super Admin",
            helper:
              "Lihat informasi akun platform dan peran Super Admin dari sesi login.",
          },
        },
      ],
    },
    {
      path: "/:pathMatch(.*)*",
      name: "not-found",
      component: NotFoundPage,
      meta: { title: "404 - Halaman Tidak Ditemukan" },
    },
  ],
});

declare module "vue-router" {
  interface RouteMeta {
    title?: string;
    requiresAuth?: boolean;
    roles?: RoleName[];
  }
}

const APP_NAME = "Wiyata";

router.beforeEach(async (to) => {
  const auth = useAuthStore();
  auth.restoreSession();

  if ((to.name === "login" || to.name === "register") && auth.isAuthenticated) {
    await auth.ensureUserContext();
    return auth.landingRoute();
  }

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: "login", query: { redirect: to.fullPath } };
  }

  if (to.meta.requiresAuth) {
    await auth.ensureUserContext();
    const hasActiveContext = Boolean(auth.activeContext);

    // Onboarding butuh auth tapi TIDAK butuh activeContext, karena justru
    // itu yang belum dimiliki user pada tahap ini.
    if (to.name === "onboarding") {
      if (hasActiveContext) {
        return auth.landingRoute();
      }
      return true;
    }

    // Create School hanya butuh login, bukan activeContext — ini memang
    // tujuan utama user yang belum punya sekolah sama sekali.
    if (to.name === "create-school") {
      return true;
    }

    if (!hasActiveContext) {
      return { name: "onboarding" };
    }
  }

  const requiredRoles = to.matched.flatMap((record) => record.meta.roles ?? []);
  if (requiredRoles.length > 0 && !auth.hasAnyRole(requiredRoles)) {
    return { name: "unauthorized" };
  }

  return true;
});

router.afterEach((to) => {
  const nearestTitle = [...to.matched]
    .reverse()
    .find((record) => record.meta.title)?.meta.title;

  document.title = nearestTitle ? `${nearestTitle} | ${APP_NAME}` : APP_NAME;
});

export default router;
