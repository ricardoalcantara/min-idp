import { NextResponse } from "next/server"
import { createSAML } from "@/lib/saml"

export async function GET(req: Request) {
  const saml = createSAML()
  const { searchParams } = new URL(req.url)
  const relayState = searchParams.get("relay") ?? "/"

  const redirectUrl = await saml.getAuthorizeUrlAsync(relayState, req.headers.get("host") ?? "", {})
  return NextResponse.redirect(redirectUrl)
}
