import { useMutation } from "react-query"
import { track } from "../../analytics"

const usePostResetPasswordEmail = () => {
  const postMailReset = async (email: string) => {
    const result = await fetch(`/auth/mailPasswordReset`, {
      credentials: "same-origin",
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email }),
    })

    if (!result.ok) {
      throw new Error("fail to send reset password email")
    }

    return result
  }
  return useMutation(postMailReset, {
    onSuccess: () => {
      track("Password Reset")
    },
  })
}

export default usePostResetPasswordEmail
