export const MESSAGE_COLORS = {
  system: {
    bg: 'bg-system-50',
    border: 'border-system-200',
    text: 'text-system-900',
    icon: '🤖'
  },
  user: {
    bg: 'bg-user-50',
    border: 'border-user-200',
    text: 'text-user-900',
    icon: '👤'
  },
  assistant: {
    bg: 'bg-assistant-50',
    border: 'border-assistant-200',
    text: 'text-assistant-900',
    icon: '🤖'
  },
  tool_call: {
    bg: 'bg-tool-50',
    border: 'border-tool-200',
    text: 'text-tool-900',
    icon: '🛠️'
  },
  tool_response: {
    bg: 'bg-tool-response-50',
    border: 'border-tool-response-200',
    text: 'text-tool-response-900',
    icon: '⚡'
  },
  error: {
    bg: 'bg-red-50',
    border: 'border-red-200',
    text: 'text-red-900',
    icon: '❌'
  }
} as const

export type MessageType = keyof typeof MESSAGE_COLORS