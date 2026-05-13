import { NextResponse } from "next/server"
import { getSession } from "@/lib/session"
import { idpSloUrl, spEntityId } from "@/lib/saml"

export async function POST(req: Request) {
  const session = await getSession()
  session.destroy()

  const sloUrl = idpSloUrl()
  if (sloUrl) {
    const base = new URL("/", req.url).href
    const target = new URL(sloUrl)
    target.searchParams.set("RelayState", base)
    target.searchParams.set("entity_id", spEntityId())
    return NextResponse.redirect(target.toString())
  }

  return NextResponse.redirect(new URL("/", req.url))
}
