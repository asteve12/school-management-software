import { useQuery, UseQueryOptions } from "react-query"
import { getApi } from "./fetchApi"
import { Subject } from "./useGetAreaSubjects"

export function useGetSubject(
  subjectId: string,
  option?: UseQueryOptions<Subject>
) {
  const fetchSubjectMaterials = getApi<Subject>(
    `/curriculums/subjects/${subjectId}`
  )
  return useQuery(["subject", subjectId], fetchSubjectMaterials, option)
}
