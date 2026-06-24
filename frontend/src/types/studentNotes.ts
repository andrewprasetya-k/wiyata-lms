export interface StudentMaterialNote {
  noteId: string
  materialId: string
  content: string
  createdAt: string
  updatedAt: string
}

export interface StudentMaterialNoteResponse {
  note: StudentMaterialNote | null
}

export interface SaveStudentMaterialNotePayload {
  content: string
}
