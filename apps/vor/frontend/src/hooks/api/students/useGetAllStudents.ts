import { useQuery } from "react-query"
import { ApiError, BASE_URL } from "../useApi"
import { getSchoolId } from "../../schoolIdState"
import { navigate } from "../../../components/Link/Link"

export interface Student {
  id: string
  name: string
  active: boolean
  profileImageUrl?: string
  classes: {
    classId: string
    className: string
  }[]
}

export const useGetAllStudents = (classId = "", active?: boolean) => {
  const schoolId = getSchoolId()
  const fetchStudents = async (): Promise<Student[]> => {
    const url = `/schools/${schoolId}/students?classId=${classId}&active=${
      active ?? ""
    }`
    const result = await fetch(`${BASE_URL}${url}`, {
      credentials: "same-origin",
    })

    // Throw user to login when something gets 401
    if (result.status === 401) {
      await navigate("/login")
    }

    if (result.status === 404) {
      await navigate("/choose-school")
    }

    if (result.status !== 200) {
      const response: ApiError = await result.json()
      throw Error(response.error?.message)
    }

    return result.json()
  }

  return useQuery(["students", schoolId, classId, active], fetchStudents)
}
