// Mirrors the backend's policy (dto.ValidatePasswordComplexity: min 8 chars
// + upper/lower/number) so the user gets fast feedback client-side — the
// server re-validates regardless, this is UX only, not the source of truth.
export function passwordPolicyError(password: string): string {
  if (password.length < 8) return "Password baru minimal 8 karakter.";
  if (!/[A-Z]/.test(password))
    return "Password baru harus mengandung minimal satu huruf besar.";
  if (!/[a-z]/.test(password))
    return "Password baru harus mengandung minimal satu huruf kecil.";
  if (!/[0-9]/.test(password))
    return "Password baru harus mengandung minimal satu angka.";
  return "";
}
