import { api } from "./api";
import type { MaterialDocumentSummaryResponse } from "../types/materialSummary";

export async function summarizeMaterialDocument(
  materialId: string,
  mediaId: string,
) {
  const { data } = await api.post<MaterialDocumentSummaryResponse>(
    `/materials/${materialId}/media/${mediaId}/summary`,
  );
  return data;
}
