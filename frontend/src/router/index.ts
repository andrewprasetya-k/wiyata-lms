import { createRouter, createWebHistory } from "vue-router";
import { useAuthStore } from "../stores/auth";
import type { RoleName } from "../types/auth";
import AuthLayout from "../layouts/AuthLayout.vue";
import StudentLayout from "../layouts/StudentLayout.vue";
import TeacherLayout from "../layouts/TeacherLayout.vue";
import AdminLayout from "../layouts/AdminLayout.vue";
import SuperAdminLayout from "../layouts/SuperAdminLayout.vue";
import LoginPage from "../pages/auth/LoginPage.vue";
import UnauthorizedPage from "../pages/auth/UnauthorizedPage.vue";
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
import TeacherDashboard from "../pages/teacher/TeacherDashboard.vue";
import TeacherSubjectDetail from "../pages/teacher/TeacherSubjectDetail.vue";
import TeacherSubjects from "../pages/teacher/TeacherSubjects.vue";
import AdminDashboard from "../pages/admin/AdminDashboard.vue";
import AdminAcademicYears from "../pages/admin/AdminAcademicYears.vue";
import AdminClasses from "../pages/admin/AdminClasses.vue";
import AdminEnrollments from "../pages/admin/AdminEnrollments.vue";
import AdminSubjectClasses from "../pages/admin/AdminSubjectClasses.vue";
import AdminUsers from "../pages/admin/AdminUsers.vue";
import SuperAdminDashboard from "../pages/superadmin/SuperAdminDashboard.vue";
import SuperAdminSchools from "../pages/superadmin/SuperAdminSchools.vue";
import SuperAdminUsers from "../pages/superadmin/SuperAdminUsers.vue";
import FeaturePlaceholder from "../components/common/FeaturePlaceholder.vue";
import ReadProfile from "../pages/profile/ReadProfile.vue";
import TeacherCreate from "../pages/teacher/TeacherCreate.vue";
import TeacherContentCreate from "../pages/teacher/TeacherContentCreate.vue";
import TeacherAssignmentReview from "../pages/teacher/TeacherAssignmentReview.vue";
import TeacherMaterialDetail from "../pages/teacher/TeacherMaterialDetail.vue";
import TeacherSubmissions from "../pages/teacher/TeacherSubmissions.vue";
import TeacherAssignments from "../pages/teacher/TeacherAssignments.vue";
import TeacherFeed from "../pages/teacher/TeacherFeed.vue";
import HomePage from "../pages/preview/HomePage.vue";

export const dashboardByRole: Record<RoleName, string> = {
  super_admin: "/superadmin/dashboard",
  admin: "/admin/dashboard",
  teacher: "/teacher/dashboard",
  student: "/student/dashboard",
};

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/",
      redirect: "/home",
    },
    {
      path: "/",
      children: [
        {
          path: "home",
          name: "home",
          component: HomePage,
        },
      ],
    },
    {
      path: "/auth",
      component: AuthLayout,
      children: [
        {
          path: "login",
          name: "login",
          component: LoginPage,
        },
        {
          path: "unauthorized",
          name: "unauthorized",
          component: UnauthorizedPage,
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
        },
        {
          path: "subjects/:sclId",
          name: "student-subject-detail",
          component: StudentSubjectDetail,
        },
        {
          path: "subjects/:sclId/materials/:matId",
          name: "student-material-detail",
          component: StudentMaterialDetail,
        },
        {
          path: "subjects/:sclId/materials/:matId/note",
          name: "student-material-note",
          component: StudentMaterialNoteEditor,
        },
        {
          path: "subjects/:sclId/assignments/:asgId",
          name: "student-assignment-detail",
          component: StudentAssignmentDetail,
        },
        {
          path: "feed",
          name: "student-feed",
          component: StudentFeed,
        },
        {
          path: "assignments",
          name: "student-assignments",
          component: StudentAssignments,
        },
        {
          path: "grades",
          name: "student-grades",
          component: StudentGrades,
        },
        {
          path: "chat",
          name: "student-chat",
          component: FeaturePlaceholder,
          props: {
            eyebrow: "Realtime chat",
            title: "Chat student",
            description:
              "Chat student direncanakan untuk komunikasi kelas dan subject setelah MVP sekolah. Untuk saat ini, komunikasi kelas berjalan melalui Feed kelas.",
          },
        },
        {
          path: "notes",
          name: "student-notes",
          component: StudentNotes,
        },
        {
          path: "profile",
          name: "student-profile",
          component: ReadProfile,
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
        },
        {
          path: "classes",
          redirect: "/teacher/subjects",
        },
        {
          path: "subjects",
          name: "teacher-subjects",
          component: TeacherSubjects,
        },
        {
          path: "subjects/:subjectClassId",
          name: "teacher-subject-detail",
          component: TeacherSubjectDetail,
        },
        {
          path: "subjects/:subjectClassId/materials/:matId",
          name: "teacher-material-detail",
          component: TeacherMaterialDetail,
        },
        {
          path: "subjects/:subjectClassId/materials/:matId/edit",
          name: "teacher-material-edit",
          component: TeacherContentCreate,
        },
        {
          path: "subjects/:subjectClassId/assignments/:asgId/edit",
          name: "teacher-assignment-edit",
          component: TeacherContentCreate,
        },
        {
          path: "subjects/:subjectClassId/create",
          name: "teacher-content-create",
          component: TeacherContentCreate,
        },
        {
          path: "assignments/:assignmentId/review",
          name: "teacher-assignment-review",
          component: TeacherAssignmentReview,
        },
        {
          path: "assignments",
          name: "teacher-assignments",
          component: TeacherAssignments,
        },
        {
          path: "submissions",
          name: "teacher-submissions",
          component: TeacherSubmissions,
        },
        {
          path: "create",
          name: "teacher-create",
          component: TeacherCreate,
        },
        {
          path: "feed",
          name: "teacher-feed",
          component: TeacherFeed,
        },
        {
          path: "chat",
          name: "teacher-chat",
          component: FeaturePlaceholder,
          props: {
            eyebrow: "Realtime chat",
            title: "Chat kelas",
            description:
              "Chat kelas direncanakan untuk komunikasi cepat guru dan siswa setelah MVP sekolah. Untuk saat ini, pengumuman kelas berjalan melalui Feed.",
          },
        },
        {
          path: "profile",
          name: "teacher-profile",
          component: ReadProfile,
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
        },
        {
          path: "classes",
          name: "admin-classes",
          component: AdminClasses,
        },
        {
          path: "users",
          name: "admin-users",
          component: AdminUsers,
        },
        {
          path: "enrollments",
          name: "admin-enrollments",
          component: AdminEnrollments,
        },
        {
          path: "subject-classes",
          name: "admin-subject-classes",
          component: AdminSubjectClasses,
        },
        {
          path: "academic-years",
          name: "admin-academic-years",
          component: AdminAcademicYears,
        },
        {
          path: "profile",
          name: "admin-profile",
          component: ReadProfile,
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
        },
        {
          path: "schools",
          name: "superadmin-schools",
          component: SuperAdminSchools,
        },
        {
          path: "users",
          name: "superadmin-users",
          component: SuperAdminUsers,
        },
        {
          path: "profile",
          name: "superadmin-profile",
          component: ReadProfile,
          props: {
            eyebrow: "Profil Super Admin",
            title: "Profil Super Admin",
            helper:
              "Lihat informasi akun platform dan peran Super Admin dari sesi login.",
          },
        },
      ],
    },
  ],
});

router.beforeEach((to) => {
  const auth = useAuthStore();
  auth.restoreSession();

  if (to.name === "login" && auth.isAuthenticated) {
    const role = auth.primaryRole();
    return role ? dashboardByRole[role] : "/auth/unauthorized";
  }

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: "login", query: { redirect: to.fullPath } };
  }

  const requiredRoles = to.matched.flatMap((record) => record.meta.roles ?? []);
  if (requiredRoles.length > 0 && !auth.hasAnyRole(requiredRoles)) {
    return { name: "unauthorized" };
  }

  return true;
});

export default router;
