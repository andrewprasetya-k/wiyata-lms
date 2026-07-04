export interface AdminSchoolMemberItem {
  schoolUserId: string;
  userId: string;
  fullName: string;
  email: string;
  roles: string[];
  classCodes?: string[];
  createdAt: string;
  deletedAt?: string | null;
  userCreated?: boolean;
  membershipAction?: string;
  emailNotification?: string;
}

export interface AdminSchoolMemberListResponse {
  data: AdminSchoolMemberItem[];
  totalItems: number;
  page: number;
  limit: number;
  totalPages: number;
}

export interface AdminSchoolMemberCreatePayload {
  fullName: string;
  email: string;
  password: string;
  role: "student" | "teacher" | "admin";
  classCode?: string;
}
