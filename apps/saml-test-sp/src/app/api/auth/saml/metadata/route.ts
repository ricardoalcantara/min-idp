import { NextResponse } from "next/server"
import { createSAML, spCert } from "@/lib/saml"

export async function GET() {
  const saml = createSAML()
  const cert = spCert()
  const xml  = saml.generateServiceProviderMetadata(cert, cert)
  return new NextResponse(xml, {
    headers: { "Content-Type": "application/xml; charset=utf-8" },
  })
}
