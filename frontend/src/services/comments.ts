import { api } from "./api";
import type {
  CommentItem,
  CreateCommentPayload,
  GetCommentsParams,
} from "../types/comment";

export async function getComments({
  sourceType,
  sourceId,
}: GetCommentsParams) {
  const { data } = await api.get<CommentItem[]>("/comments", {
    params: { type: sourceType, id: sourceId },
  });
  return data ?? [];
}

export async function createComment(payload: CreateCommentPayload) {
  const { data } = await api.post<{ message: string }>("/comments", payload);
  return data;
}

export async function deleteComment(commentId: string) {
  const { data } = await api.delete<{ message: string }>(
    `/comments/${commentId}`,
  );
  return data;
}
