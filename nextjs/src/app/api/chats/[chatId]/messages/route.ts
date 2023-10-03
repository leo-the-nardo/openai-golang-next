import { NextRequest, NextResponse } from "next/server"
import { prisma } from "@/app/prisma/prisma"
import { withAuth } from "@/app/api/helpers"

export const GET = withAuth(async function (
  _request: NextRequest,
  token,
  { params }: { params: { chatId: string } },
) {
  const chat = await prisma.chat.findUniqueOrThrow({
    where: { id: params.chatId },
  })
  if (chat.user_id !== token.sub) {
    return new NextResponse(JSON.stringify({ error: "Not Found" }), {
      status: 404,
    })
  }
  const messages = await prisma.message.findMany({
    where: { chat_id: params.chatId },
    orderBy: { created_at: "asc" },
  })
  return NextResponse.json(messages)
})

export const POST = withAuth(async function (
  request: NextRequest,
  token,
  { params }: { params: { chatId: string } },
) {
  const chat = await prisma.chat.findUniqueOrThrow({
    where: { id: params.chatId },
  })
  if (chat.user_id !== token.sub) {
    return new NextResponse(JSON.stringify({ error: "Not Found" }), {
      status: 404,
    })
  }
  if (!chat) {
    throw new Error("Chat not found")
  }
  const body = await request.json()
  const message = await prisma.message.create({
    data: {
      content: body.message,
      chat_id: chat.id,
    },
  })
  return NextResponse.json(message)
})
