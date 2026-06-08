export interface CreateMaterialPayload {
  schoolId: string
  subjectClassId: string
  materialTitle: string
  materialDesc?: string
  materialType: 'video' | 'pdf' | 'ppt' | 'other'
  mediaIds: string[]
}

export interface MaterialItem {
  materialId: string
  subjectClassId?: string
  subjectName?: string
  materialTitle: string
  materialDesc?: string
  materialType: string
  creatorName?: string
  createdAt: string
  attachments?: {
    mediaId: string
    mediaName: string
    fileSize?: number
    mimeType?: string
    fileUrl?: string
    thumbnailUrl?: string
    ownerType?: string
    createdAt?: string
  }[]
}

export interface MaterialListResponse {
  subjectClass: {
    subjectClassId: string
    subjectName: string
  }
  data: {
    data: MaterialItem[]
    totalItems?: number
    page?: number
    limit?: number
    totalPages?: number
  }
}
