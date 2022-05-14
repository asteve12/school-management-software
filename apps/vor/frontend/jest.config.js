module.exports = {
  transform: {
    "^.+\\.(js|ts|tsx)?$": `<rootDir>/jest-preprocess.js`,
    "^.+\\.(esm.js)?$": `<rootDir>/jest-preprocess.js`,
  },
  moduleNameMapper: {
    // ".+\\.(css|styl|less|sass|scss)$": `identity-obj-proxy`,
    "\\.svg$": "<rootDir>/src/__mocks__/svgrMock.js",
    ".+\\.(jpg|jpeg|png|gif|eot|otf|webp|ttf|woff|woff2|mp4|webm|wav|mp3|m4a|aac|oga)$": `<rootDir>/src/__mocks__/file-mock.js`,
  },
  testPathIgnorePatterns: [`node_modules`, `.cache`],
  transformIgnorePatterns: [
    `node_modules/(?!(gatsby|gatsby-theme-i18n|react-adaptive-hooks)/)`,
  ],
  testMatch: ["**/?(*.)+(test).[jt]s?(x)"],
  globals: {
    __PATH_PREFIX__: ``,
    __BASE_PATH__: ``,
  },
  testURL: `http://localhost`,
  setupFiles: [
    `<rootDir>/loadershim.ts`,
    `jest-date-mock`,
    "fake-indexeddb/auto",
  ],
  snapshotSerializers: ["@emotion/jest/serializer"],
  testEnvironment: "jsdom",
}
