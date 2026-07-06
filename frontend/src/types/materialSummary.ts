export interface MaterialDocumentSummaryResponse {
  status: "generated" | "cached";
  summary: string;
  source: {
    materialId: string;
    mediaId: string;
    mediaName: string;
    mimeType: string;
  };
}
