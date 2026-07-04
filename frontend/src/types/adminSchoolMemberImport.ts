export interface AdminSchoolMemberImportRow {
  rowNumber: number;
  fullName: string;
  email: string;
  role: string;
  classCode?: string;
  status: "valid" | "invalid";
  errors: string[];
}

export interface AdminSchoolMemberImportPreviewResponse {
  rows: AdminSchoolMemberImportRow[];
  validCount: number;
  invalidCount: number;
}

export interface AdminSchoolMemberImportCommitPayload {
  defaultPassword: string;
  rows: AdminSchoolMemberImportRow[];
}

export interface AdminSchoolMemberImportResult {
  rowNumber: number;
  fullName: string;
  email: string;
  role: string;
  classCode?: string;
  status: "imported" | "skipped" | "failed";
  reason?: string;
  userCreated?: boolean;
  membershipAction?: string;
  emailNotification?: string;
}

export interface AdminSchoolMemberImportCommitResponse {
  importedCount: number;
  skippedCount: number;
  failedCount: number;
  results: AdminSchoolMemberImportResult[];
}
