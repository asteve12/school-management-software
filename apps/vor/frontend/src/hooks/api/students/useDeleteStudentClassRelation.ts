import { useMutation, useQueryClient } from "react-query"
import { track } from "../../../analytics"
import { navigate } from "../../../components/Link/Link"
import { ApiError, BASE_URL } from "../useApi"

export const useDeleteStudentClassRelation = (studentId: string) => {
  const queryCache = useQueryClient()
  const deleteStudentClassRelation = async (classId: string) => {
    const result = await fetch(`${BASE_URL}/students/${studentId}/classes`, {
      credentials: "same-origin",
      method: "DELETE",
      body: JSON.stringify({ classId }),
    })

    // Throw user to login when something gets 401
    if (result.status === 401) {
      await navigate("/login")
      return result
    }
    if (!result.ok) {
      const body: ApiError = await result.json()
      throw Error(body?.error?.message ?? "")
    }

    return result
  }

  return useMutation(deleteStudentClassRelation, {
    onSuccess: async () => {
      track("Student Class Relation Deleted")
      await queryCache.invalidateQueries(["student", studentId])
    },
  })
}

export default useDeleteStudentClassRelation
