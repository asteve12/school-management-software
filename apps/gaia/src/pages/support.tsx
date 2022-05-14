import { withPageAuthRequired } from "@auth0/nextjs-auth0"
import { useEffect } from "react"
import Head from "next/head"
import { useQueryString } from "../hooks/useQueryString"
import loadCanny from "../utils/canny"
import useGetUser from "../hooks/api/useGetUser"
import useGetChild from "../hooks/api/useGetChild"
import Icon from "../components/Icon/Icon"

const SupportPage = () => {
  const childId = useQueryString("childId")
  const child = useGetChild(childId)
  const user = useGetUser()

  useEffect(() => {
    if (user.isSuccess && child.isSuccess) {
      loadCanny()
      Canny("identify", {
        appID: "5f0d32f03899af5d46779764",
        user: {
          email: user.data?.name,
          name: user.data?.name,
          id: user.data?.sub,
          companies: [
            {
              name: child.data?.schoolName,
              id: child.data?.schoolId,
            },
          ],
        },
      })
    }
  }, [user.data, child.data])

  return (
    <>
      <Head>
        <title>Support | Obserfy for Parents</title>
      </Head>
      <div className="max-w-3xl mx-auto mt-2 flex flex-col md:flex-row items-stretch">
        <a
          href="https://feedback.obserfy.com/parent-dashboard"
          className="block w-full md:rounded p-3 bg-white border border-gray-300 mb-2 md:mb-0 flex items-center md:mr-2 hover:border-primary"
        >
          <div className="mr-3">
            <h6 className="text-gray-900 font-bold text-sm mb-2">
              Send feedback
            </h6>
            <p className="text-gray-700 text-sm">
              Send us suggestions on how to improve obserfy.
            </p>
          </div>
          <Icon
            className="my-auto h-full ml-auto"
            alt="logout icon"
            src="/icons/external-link.svg"
            size={20}
          />
        </a>
        <a
          href="mailto:support@obserfy.com"
          className="w-full md:rounded p-3 bg-white border border-gray-300 flex items-center hover:border-primary"
        >
          <div className="mr-3">
            <h6 className="text-gray-900 font-bold text-sm mb-2">Email Us</h6>
            <p className="text-gray-700 text-sm">
              Have a question? Shoot us an email at support@obserfy.com
            </p>
          </div>
          <Icon
            className="my-auto h-full ml-auto text-gray-900 flex-shrink-0"
            src="/icons/mail.svg"
            size={24}
          />
        </a>
      </div>
    </>
  )
}

export default withPageAuthRequired(SupportPage)
