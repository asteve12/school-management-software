import { useMutation } from "react-query"
import { track } from "../../analytics"
import { getSchoolId } from "../schoolIdState"
import { postApi } from "./fetchApi"

interface Payload {
  email: Array<string>
}
export const usePostUserInvite = () => {
  const sendInvite = postApi<Payload>(`/schools/${getSchoolId()}/invite-user`)
  return useMutation(sendInvite, {
    onSuccess: (data, variables) => {
      track("Users Invited", {
        count: variables.email.length,
      })
    },
  })
}
