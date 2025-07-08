import { BACKEND_URL, handleStreamResponse, type StreamResponse } from "./controllers";

const sendPromptAIUrl = `${BACKEND_URL}/ai/request`;

const sendAIPrompt = async (
  prompt: string,
  onReceive: (streamResponse: StreamResponse) => void,
  onError: (error: string) => void
): Promise<AbortController> => {
  const response = await fetch(sendPromptAIUrl, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ prompt }),
  });

  const controller = new AbortController();
  handleStreamResponse(response, onReceive, onError);
  return controller;
};

export { sendAIPrompt };