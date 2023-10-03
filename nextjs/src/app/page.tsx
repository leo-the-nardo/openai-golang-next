"use client"

import useSWR from "swr"
import useSWRSubscription from "swr/subscription"
import ClientHttp, { fetcher } from "@/http/http"
import { Chat as PrismaChat, Message } from "@prisma/client"
import { useRouter, useSearchParams } from "next/navigation"
import { FormEvent, useEffect, useLayoutEffect, useState } from "react"
import { PlusIcon } from "@/app/components/PlusIcon"
import { MessageIcon } from "@/app/components/MessageIcon"
import { ArrowRightIcon } from "@/app/components/ArrowRightIcon"
import Image from "next/image"
import { UserIcon } from "@/app/components/UserIcon"
import { marked } from "marked"
import hljs from "highlight.js"
import DOMPurify from "dompurify"
import { LogoutIcon } from "@/app/components/LogoutIcon"
import { signOut } from "next-auth/react"

const renderer = new marked.Renderer()

renderer.code = function (code, language) {
  const validLanguage = hljs.getLanguage(language!) ? language : "plaintext"
  const highlightedCode = hljs.highlight(code, {
    language: validLanguage!,
  }).value
  return `<pre><code class="hljs ${validLanguage}">${highlightedCode}</code></pre>`
}
marked.use({ renderer })
type Chat = PrismaChat & {
  messages: [Message] // only the first message
}

function ChatItemError({ children }: { children: any }) {
  return (
    <li className="w-full bg-gray-800 text-gray-100">
      <div className="m-auto flex flex-row items-start space-x-4 py-6 md:max-w-2xl lg:max-w-xl xl:max-w-3xl">
        <Image src="/vercel.svg" width={30} height={30} alt="" />
        <div className="relative flex w-[calc(100%-115px)] flex-col gap-1">
          <span className="text-red-500">Ops! Ocorreu um erro: {children}</span>
        </div>
      </div>
    </li>
  )
}

const keyDownHandler = (event: KeyboardEvent) => {
  if (event.key === "Enter" && !event.shiftKey) {
    event.preventDefault()
  }
}
const Loading = () => (
  <span className="h-6 w-[5px] animate-spin rounded bg-white"></span>
)

function ChatItem({
  content,
  is_from_bot,
  loading = false,
}: {
  content: string
  is_from_bot: boolean
  loading?: boolean
}) {
  const background = is_from_bot ? "bg-gray-800" : "bg-gray-600"

  return (
    <li className={`w-full text-gray-100 ${background}`}>
      <div className="flex-col">
        <div className="m-auto flex flex-row items-start space-x-4 py-6 md:max-w-2xl lg:max-w-xl xl:max-w-3xl">
          {is_from_bot ? (
            <Image
              src="https://github.com/leo-the-nardo.png"
              width={30}
              height={30}
              alt=""
            />
          ) : (
            <UserIcon className="start relative flex w-[30px] flex-col" />
          )}

          <div
            className="relative flex w-[calc(100%-115px)] flex-col gap-1 break-words transition duration-100 ease-linear"
            dangerouslySetInnerHTML={{
              __html: DOMPurify.sanitize(
                marked(content, { breaks: true }) as string,
              ), //sanitized
            }}
          />
        </div>
        {loading && (
          <div className="flex items-center justify-center pb-2">
            <Loading />
          </div>
        )}
      </div>
    </li>
  )
}

export default function Home() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const chatIdParam = searchParams.get("id")
  const [chatId, setChatId] = useState<string | null>(chatIdParam)
  const [questionId, setQuestionId] = useState<string | null>(null)
  const { data: chats, mutate: mutateChats } = useSWR<Chat[]>( // GET/api/chats
    "chats",
    fetcher,
    { fallbackData: [], revalidateOnFocus: false },
  )
  const { data: messages, mutate: mutateMessages } = useSWR<Message[]>(
    chatId ? `chats/${chatId}/messages` : null, // GET/api/chats/:id/messages only fetch if chatId is not null
    fetcher,
    { fallbackData: [], revalidateOnFocus: false },
  )
  const { data: messageLoading, error: errorMessageLoad } = useSWRSubscription(
    questionId ? `/api/messages/${questionId}/sse` : null, // GET /api/messages/:id/sse only connect if questionId is not null
    (path, { next }) => {
      console.log(`useSWRSubscription -> init event source`, path)
      const eventSource = new EventSource(path)
      eventSource.onmessage = (event) => {
        console.log(`useSWRSubscription -> onmessage`, event)
        const message = JSON.parse(event.data)
        next(null, message.content)
        // mutateMessages((messages) => [...messages!, message], false)
      }
      eventSource.onerror = (event) => {
        console.log(`useSWRSubscription -> onerror`, event)
        eventSource.close()
        // @ts-ignore
        next(event.data, null)
      }
      eventSource.addEventListener("end", (event) => {
        console.log(`useSWRSubscription -> on end`, event)
        const newMessage = JSON.parse(event.data)
        mutateMessages((messages) => [...messages!, newMessage], false)
        next(null, null)
        eventSource.close()
      })
      return () => {
        console.log(`useSWRSubscription -> close event source`)
        eventSource.close()
      }
    },
    {
      revalidateOnFocus: false,
      revalidateOnReconnect: false,
      fallbackData: null,
    },
  )

  useEffect(() => {
    setChatId(chatIdParam)
    console.log(`useEffect -> chatIdParam ${chatIdParam}`)
  }, [chatIdParam])

  // Add event listeners
  useEffect(() => {
    const textarea = document.querySelector("#message") as HTMLTextAreaElement
    textarea.addEventListener("keydown", keyDownHandler)
    textarea.addEventListener("keyup", keyUpHandler)
    return () => {
      textarea.removeEventListener("keydown", keyDownHandler)
      textarea.removeEventListener("keyup", keyUpHandler)
    }
  }, [])
  useLayoutEffect(() => {
    console.log("messageLoading", messageLoading)
    if (!messageLoading) return
    const chatting = document.querySelector("#chatting") as HTMLUListElement
    chatting.scrollTop = chatting.scrollHeight
  }, [messageLoading])

  const keyUpHandler = (event: KeyboardEvent) => {
    const textArea = event.target as HTMLTextAreaElement
    if (event.key === "Enter" && !event.shiftKey) {
      event.preventDefault()
      const form = document.querySelector("form") as HTMLFormElement
      const buttonSubmit = form.querySelector(
        "button[type=submit]",
      ) as HTMLButtonElement
      form.requestSubmit(buttonSubmit)
    }
    if (textArea.scrollHeight >= 200) {
      textArea.style.overflowY = "scroll"
    } else {
      textArea.style.overflowY = "hidden"
      textArea.style.height = "auto"
      textArea.style.height = textArea.scrollHeight + "px"
    }
  }

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (messageLoading) return
    const textarea = event.currentTarget.querySelector(
      "textarea",
    ) as HTMLTextAreaElement
    const textareaContent = textarea.value
    if (textareaContent.trim() === "") return
    const isNewChat = !chatId
    if (isNewChat) {
      console.log("on submit -> new chat")
      // POST /api/chats
      const createdChat: Chat = await ClientHttp.post("chats", {
        message: textareaContent,
      })
      console.log(JSON.stringify(createdChat, null, 2))
      const firstQuestion = createdChat.messages[0]
      mutateChats([createdChat, ...chats!], false)
      setChatId(createdChat.id)
      setQuestionId(firstQuestion.id)
      textarea.value = ""
      return
    }
    // POST /api/chats/:id/messages
    const question: Message = await ClientHttp.post(
      `chats/${chatId}/messages`,
      {
        message: textareaContent,
      },
    )
    mutateMessages([...messages!, question], false)
    setQuestionId(question.id)
    textarea.value = ""
  }

  function newChatHandler() {
    router.push("/")
    setChatId(null)
    setQuestionId(null)
  }

  async function logout() {
    await signOut({ redirect: false })
    const searchParams = new URLSearchParams({
      redirect: window.location.origin,
    })
    const { url: logoutUrl } = await ClientHttp.get(
      `logout-url?${searchParams}`,
    )
    window.location.href = logoutUrl
  }

  return (
    <div className="relative flex h-full w-full overflow-hidden">
      {/* -- sidebar -- */}
      <div className="flex h-screen w-[300px] flex-col bg-gray-900 p-2">
        {/* -- button new chat -- */}
        <button
          className="mb-1 flex cursor-pointer gap-3 rounded border border-white/20 p-3 text-sm text-white transition-colors duration-200 hover:bg-gray-500/10"
          onClick={newChatHandler}
        >
          <PlusIcon className="h-5 w-5" />
          New chat
        </button>
        {/* -- end button new chat -- */}
        {/* -- chats -- */}
        <div className="-mr-2 flex-grow overflow-hidden overflow-y-auto">
          {chats?.map((chat, key) => (
            <div className="mr-2 pb-2 text-sm text-gray-100" key={key}>
              <button
                className="group flex w-full cursor-pointer gap-3 rounded p-3 hover:bg-[#3f4679] hover:pr-4"
                onClick={() => router.push(`/?id=${chat.id}`)}
              >
                <MessageIcon className="h-5 w-5" />
                <div className="relative max-h-5 w-full overflow-hidden break-all text-left">
                  {chat.messages[0].content}
                  <div className="absolute inset-y-0 right-0 z-10 w-8 bg-gradient-to-l from-gray-900 group-hover:from-[#3f4679]"></div>
                </div>
              </button>
            </div>
          ))}
        </div>
        <button
          className="mt-1 flex gap-3 rounded p-3 text-sm text-white hover:bg-gray-500/10"
          onClick={() => logout()}
        >
          <LogoutIcon className="h-5 w-5" />
          Log out
        </button>
      </div>
      {/* -- end sidebar -- */}

      {/* -- main content */}
      <div className="relative flex-1 flex-col">
        <ul id="chatting" className="h-screen overflow-y-auto bg-gray-800">
          {messages?.map((message, key) => (
            <ChatItem
              key={key}
              content={message.content}
              is_from_bot={message.is_from_bot}
            />
          ))}
          {messageLoading && (
            <ChatItem
              content={messageLoading}
              is_from_bot={true}
              loading={true}
            />
          )}
          {errorMessageLoad && (
            <ChatItemError>{errorMessageLoad}</ChatItemError>
          )}

          <li className="h-36 bg-gray-800"></li>
        </ul>

        <div className="absolute bottom-0 w-full !bg-transparent bg-gradient-to-b from-gray-800 to-gray-950">
          <div className="mx-auto mb-6 max-w-3xl">
            <form id="form" onSubmit={onSubmit}>
              <div className="relative flex flex-col rounded bg-gray-700 py-3 pl-4 text-white">
                <textarea
                  id="message"
                  tabIndex={0}
                  rows={1}
                  placeholder="Digite sua pergunta"
                  className="resize-none bg-transparent pl-0 pr-14 outline-none"
                  defaultValue="Gere uma classe de produto em JavaScript"
                ></textarea>
                <button
                  type="submit"
                  className="absolute bottom-2.5 top-1 rounded text-gray-400 hover:bg-gray-900 hover:text-gray-400 md:right-4"
                  disabled={messageLoading}
                >
                  <ArrowRightIcon className="text-white-500 w-8" />
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>
      {/* -- main content */}
    </div>
  )
}
