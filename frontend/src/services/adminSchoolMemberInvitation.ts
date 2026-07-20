import { api } from "./api";
import type {
  CreateSchoolMemberInvitationPayload,
  CreateSchoolMemberInvitationResponse,
  SchoolMemberInvitationListResponse,
} from "../types/adminSchoolMemberInvitation";

export async function createSchoolMemberInvitation(
  payload: CreateSchoolMemberInvitationPayload,
) {
  const { data } = await api.post<CreateSchoolMemberInvitationResponse>(
    "/admin/school-member-invitations",
    payload,
  );
  return data;
}

export async function listSchoolMemberInvitations(params: {
  status?: "pending" | "accepted" | "revoked" | "expired";
  page?: number;
  limit?: number;
}) {
  const { data } = await api.get<SchoolMemberInvitationListResponse>(
    "/admin/school-member-invitations",
    {
      params: {
        status: params.status || undefined,
        page: params.page ?? 1,
        limit: params.limit ?? 20,
      },
    },
  );
  return data;
}
