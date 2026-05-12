import { NextResponse } from "next/server"
import { createSAML } from "@/lib/saml"
import { getSession } from "@/lib/session"

export async function POST(req: Request) {
  const saml = createSAML()

  const formData = await req.formData()
  const body: Record<string, string> = {}
  formData.forEach((v, k) => { body[k] = v.toString() })

  try {
    const { profile } = await saml.validatePostResponseAsync(body)

    const session = await getSession()
    session.nameId      = profile?.nameID ?? ""
    session.nameIdFormat = profile?.nameIDFormat ?? ""
    session.attributes  = (profile?.attributes ?? {}) as Record<string, string | string[]>
    await session.save()

    const base = process.env.SAML_SP_ENTITY_ID ?? "http://localhost:3002"
    const relayState = body["RelayState"] ?? "/"
    return NextResponse.redirect(new URL(relayState, base), { status: 303 })
  } catch (err) {
    console.error("SAML callback error:", err)
    return NextResponse.redirect(new URL("/?error=saml_failed", process.env.SAML_SP_ENTITY_ID ?? "http://localhost:3002"), { status: 303 })
  }
}
