import { api } from "./api"
import type {
  AcademicActivityParams,
  AcademicActivityResponse,
} from "../types/activity"

export async function getAcademicActivities(
  params: AcademicActivityParams = {},
) {
  const { data } = await api.get<AcademicActivityResponse>(
    "/academic-activity",
    {
      params: {
        from: params.from || undefined,
        to: params.to || undefined,
      },
    },
  )

  return {
    ...data,
    items: data.items ?? [],
  }
}
