export interface SchoolHeader {
  schoolId: string
  schoolName: string
  schoolCode: string
  schoolLogo?: string
}

export interface AcademicYearItem {
  academicYearId: string
  schoolId: string
  schoolName?: string
  schoolCode?: string
  academicYearName: string
  isActive: boolean
  createdAt: string
}

export interface AcademicYearsBySchoolResponse {
  school: SchoolHeader
  data: AcademicYearItem[]
}

export interface CreateAcademicYearPayload {
  schoolId: string
  academicYearName: string
}

export interface TermItem {
  termId: string
  academicYearId: string
  academicYearName?: string
  schoolName?: string
  termName: string
  isActive: boolean
  createdAt: string
}

export interface CreateTermPayload {
  academicYearId: string
  termName: string
}

export interface SubjectItem {
  subjectId: string
  schoolId: string
  schoolName?: string
  schoolCode?: string
  subjectName: string
  subjectCode: string
  createdAt: string
}

export interface SchoolSubjectsResponse {
  school: SchoolHeader
  subjects: SubjectItem[]
}

export interface CreateSubjectPayload {
  schoolId: string
  subjectName: string
  subjectCode: string
}

export interface AssignmentCategoryItem {
  categoryId: string
  schoolId: string
  categoryName: string
  createdAt: string
}

export interface SchoolAssignmentCategoriesResponse {
  school: SchoolHeader
  categories: AssignmentCategoryItem[]
}

export interface CreateAssignmentCategoryPayload {
  schoolId: string
  categoryName: string
}

export interface AssessmentWeightItem {
  weightId: string
  categoryId: string
  categoryName: string
  weight: number
}

export interface AssessmentWeightsResponse {
  subjectId: string
  subjectName: string
  subjectCode: string
  weights: AssessmentWeightItem[]
  totalWeight: number
}

export interface SaveAssessmentWeightItem {
  categoryId: string
  weight: number
}

export interface SaveAssessmentWeightsPayload {
  subjectId: string
  weights: SaveAssessmentWeightItem[]
}
