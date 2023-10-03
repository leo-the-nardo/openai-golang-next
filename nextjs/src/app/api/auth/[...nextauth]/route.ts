import NextAuth, { NextAuthOptions } from "next-auth"
import KeycloakProvider from "next-auth/providers/keycloak"

export const authConfig: NextAuthOptions = {
  providers: [
    KeycloakProvider({
      clientId: process.env.KEYCLOAK_CLIENT_ID!,
      clientSecret: process.env.KEYCLOAK_CLIENT_SECRET!,
      issuer: process.env.KEYCLOAK_ISSUER!,
    }),
  ],
}
const handler = NextAuth(authConfig)

export { handler as GET, handler as POST }
