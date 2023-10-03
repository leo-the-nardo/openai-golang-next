"use client"
import { SessionProvider as NextSessionProvider } from "next-auth/react"
import { Session } from "next-auth"
import { PropsWithChildren } from "react"

type Props = PropsWithChildren<{
  session: Session | null //comes from server
}>

export function SessionProvider(props: Props) {
  return (
    <NextSessionProvider session={props.session}>
      {props.children}
    </NextSessionProvider>
  )
}
