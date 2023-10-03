import { NextRequest, NextResponse } from "next/server"

// maybe this should be in frontend
export async function GET(req: NextRequest) {
  const redirect = req.nextUrl.searchParams.get("redirect")
  if (!redirect)
    return new NextResponse(
      JSON.stringify({ error: "Missing redirect param" }),
      { status: 400 },
    )
  const url = `${
    process.env.KEYCLOAK_ISSUER
  }/protocol/openid-connect/logout?post_logout_redirect_uri=${encodeURIComponent(
    redirect,
  )}&client_id=${process.env.KEYCLOAK_CLIENT_ID}`
  return NextResponse.json({ url })
}
