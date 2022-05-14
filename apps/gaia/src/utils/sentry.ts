import * as Sentry from "@sentry/node"
import { RewriteFrames } from "@sentry/integrations"

export const initSentry = () => {
  if (process.env.NEXT_PUBLIC_SENTRY_DSN) {
    const integrations = []
    if (
      process.env.NEXT_IS_SERVER === "true" &&
      process.env.NEXT_PUBLIC_SENTRY_SERVER_ROOT_DIR
    ) {
      // For Node.js, rewrite Error.stack to use relative paths, so that source
      // maps starting with ~/_next map to files in Error.stack with path
      // app:///_next
      integrations.push(
        new RewriteFrames({
          iteratee: (frame) => {
            // eslint-disable-next-line no-param-reassign
            frame.filename = frame.filename?.replace(
              process.env.NEXT_PUBLIC_SENTRY_SERVER_ROOT_DIR ?? "",
              "app:///"
            )
            // eslint-disable-next-line no-param-reassign
            frame.filename = frame.filename?.replace(".next", "_next")
            return frame
          },
        })
      )
    }

    Sentry.init({
      enabled: process.env.NODE_ENV === "production",
      dsn: "https://cb901298e868441898d8717f07a20188@o330610.ingest.sentry.io/5447244",
      release: process.env.NEXT_PUBLIC_RELEASE,
      integrations,
    })
  }
}
