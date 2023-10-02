import { NextRequest } from "next/server"
import { prisma } from "@/app/prisma/prisma"
import { ChatServiceClientFactory } from "@/grpc/chat-service-client"

export async function GET(
  request: NextRequest,
  { params }: { params: { messageId: string } },
) {
  const message = await prisma.message.findUniqueOrThrow({
    where: { id: params.messageId },
    include: { chat: true },
  })

  const transformStream = new TransformStream()
  const writer = transformStream.writable.getWriter()
  if (message.has_answered) {
    writeStream(writer, "error", "Message already answered")
    writer.close()
    return response(transformStream, 403)
  }
  if (message.is_from_bot) {
    writeStream(writer, "error", "Message from bot")
    writer.close()
    return response(transformStream, 403)
  }
  const chatServiceClient = ChatServiceClientFactory.create()

  //grpc call
  const grpcStream = chatServiceClient.chatStream({
    message: message.content,
    chat_id: message.chat.remote_chat_id,
    user_id: "1", //this will come from auth
  })
  let messageReceived: MessageReceived = null
  grpcStream.on("data", (data) => {
    messageReceived = data
    writeStream(writer, "message", data)
  })
  grpcStream.on("error", async (error) => {
    console.error(error)
    writeStream(writer, "error", error)
    await writer.close()
  })
  grpcStream.on("end", async () => {
    if (!messageReceived) {
      //end flow without touch in data
      writeStream(writer, "error", "Message not received")
      await writer.close()
      return
    }
    const [newMessage] = await prisma.$transaction([
      prisma.message.create({
        data: {
          content: messageReceived.content,
          chat_id: message.chat.id,
          has_answered: true,
          is_from_bot: true,
        },
      }),
      prisma.chat.update({
        where: { id: message.chat.id },
        data: { remote_chat_id: messageReceived.chatId },
      }),
      prisma.message.update({
        where: { id: message.id },
        data: { has_answered: true },
      }),
    ])
    writeStream(writer, "end", newMessage)
    await writer.close()
  })
  return response(transformStream)
}

function response(responseStream: TransformStream, status: number = 200) {
  return new Response(responseStream.readable, {
    headers: {
      "content-type": "text/event-stream",
      "cache-control": "no-cache",
      connection: "keep-alive",
    },
    status,
  })
}

type MessageReceived = {
  content: string
  chatId: string
} | null

type Event = "message" | "error" | "end"

function writeStream(
  writer: WritableStreamDefaultWriter,
  event: Event,
  data: any,
) {
  const encoder = new TextEncoder()
  writer.write(encoder.encode(`event: ${event}\n`))
  writer.write(encoder.encode(`id: ${new Date().getTime()}\n`))
  const streamData = JSON.stringify(data)
  writer.write(encoder.encode(`data: ${streamData}\n\n`))
}
