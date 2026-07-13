import { useRouter } from "vue-router";
import { useAuthStore } from "../stores/auth";
import { useConfirmStore } from "../stores/confirm";

export function useLogoutConfirm(options: { redirectTo?: string | false } = {}) {
  const { redirectTo = "/login" } = options;
  const auth = useAuthStore();
  const router = useRouter();
  const confirm = useConfirmStore();

  async function confirmLogout() {
    const ok = await confirm.confirm({
      title: "Keluar?",
      description: "Apakah Anda yakin ingin keluar dari akun ini?",
      confirmLabel: "Keluar",
      variant: "danger",
    });
    if (!ok) return;

    auth.logout();
    if (redirectTo) {
      router.push(redirectTo);
    }
  }

  return { confirmLogout };
}
