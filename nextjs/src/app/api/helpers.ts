import { NextRequest, NextResponse } from "next/server"
import { getToken, JWT } from "next-auth/jwt"

type Config = { params: any }

type RouteHandler = (
  req: NextRequest,
  token: JWT,
  config: Config,
) => Promise<NextResponse | Response> | NextResponse

export function withAuth(routeHandler: RouteHandler) {
  return async function (req: NextRequest, config: Config) {
    const token = await getToken({ req })
    if (!token)
      return new NextResponse(JSON.stringify({ error: "Unauthenticated" }), {
        status: 401,
      })
    return routeHandler(req, token, config)
  }
}
