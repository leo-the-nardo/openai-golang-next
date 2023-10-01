import nextJest from "next/jest"

const createJestConfig = nextJest({
  dir: "./",
})

/** @type {import('jest').Config} */
const config = {
  testEnvironment: "node",
}

export default createJestConfig(config)
