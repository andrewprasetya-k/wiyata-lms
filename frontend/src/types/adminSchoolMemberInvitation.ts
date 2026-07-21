export type SchoolMemberInvitationRole = "student" | "teacher";

export interface SchoolMemberInvitationClass {
  classId: string;
  classCode: string;
  classTitle: string;
}

export interface SchoolMemberInvitationItem {
  invitationId: string;
  fullName: string;
  email: string;
  role: SchoolMemberInvitationRole;
  class?: SchoolMemberInvitationClass;
  status: "pending" | "accepted" | "revoked" | "expired";
  expiresAt: string;
  acceptedAt?: string | null;
  revokedAt?: string | null;
  createdAt: string;
}

export interface CreateSchoolMemberInvitationPayload {
  email: string;
  role: SchoolMemberInvitationRole;
  classCode?: string;
}

export interface CreateSchoolMemberInvitationResponse {
  message: string;
  invitation: SchoolMemberInvitationItem;
  acceptUrl: string;
  token: string;
}

export interface SchoolMemberInvitationListResponse {
  data: SchoolMemberInvitationItem[];
  totalItems: number;
  page: number;
  limit: number;
  totalPages: number;
}
