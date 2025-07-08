import { BACKEND_URL, handleStreamResponse, type StreamResponse } from "./controllers";

const sendPromptAIUrl = `${BACKEND_URL}/ai/request`;

const sendAIPrompt = (
  prompt: string,
  onReceive: (streamResponse: StreamResponse) => void,
  onError: (error: string) => void
): AbortController => {
  const controller = new AbortController();

  const run = async () => {
    try {
      const response = await fetch(sendPromptAIUrl, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ prompt }),
        signal: controller.signal,
      });
      await handleStreamResponse(response, onReceive, onError);
    } catch (error) {
      if (error instanceof Error && error.name !== 'AbortError') {
        onError(error.message);
      }
    }
  };

  run();

  return controller;
};

export { sendAIPrompt };