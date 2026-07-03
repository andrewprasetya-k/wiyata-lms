import { api } from "./api";
import type {
  CreateSchoolMemberInvitationPayload,
  CreateSchoolMemberInvitationResponse,
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
