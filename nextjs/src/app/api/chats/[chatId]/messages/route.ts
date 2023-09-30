import { NextRequest, NextResponse } from "next/server"
import { prisma } from "@/app/prisma/prisma"

export async function GET(
  _request: NextRequest,
  { params }: { params: { chatId: string } },
) {
  const messages = await prisma.message.findMany({
    where: { chat_id: params.chatId },
    orderBy: { created_at: "asc" },
  })
  return NextResponse.json(messages)
}

export async function POST(
  request: NextRequest,
  { params }: { params: { chatId: string } },
) {
  const chat = await prisma.chat.findUnique({
    where: { id: params.chatId },
    select: { id: true },
  })
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
}
