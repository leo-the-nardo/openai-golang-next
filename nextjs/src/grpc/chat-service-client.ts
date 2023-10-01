import { ChatServiceClient as GrpcChatServiceClient } from "./rpc/pb/ChatService"
import { chatClient } from "@/grpc/client"
import * as grpc from "@grpc/grpc-js"

export class ChatServiceClient {
  private static AUTH_KEY = "123456"

  constructor(private chatClient: GrpcChatServiceClient) {}

  chatStream(data: {
    chat_id: string | null
    user_id: string
    message: string
  }) {
    const metadata = new grpc.Metadata()
    metadata.set("authorization", ChatServiceClient.AUTH_KEY)

    const stream = this.chatClient.chatStream(
      {
        chatId: data.chat_id!,
        userId: data.user_id,
        userMessage: data.message,
      },
      metadata,
    )
    // stream.on("data", (response) => {
    //   console.log(response)
    // })
    // stream.on("error", (error) => {
    //   console.log(error)
    // })
    // stream.on("end", () => {
    //   console.log("end")
    // })
    return stream
  }
}

export class ChatServiceClientFactory {
  static create() {
    return new ChatServiceClient(chatClient)
  }
}
